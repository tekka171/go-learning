package kvstore

import (
	"container/list"
	"hash/maphash"
	"time"

	"github.com/tekka171/go-learning/periodic"
)

// Requirements
// 1: Basic + Lazy Expiration
// - Safe for concurrent use by many goroutines.
// - An expired key must never be returned by Get, even if it hasn't been cleaned up yet (lazy expiration).
// - Set on an existing key overwrites both the value and the TTL.
// 2: Enable Background Cleanup
// - A goroutine wakes every cleanupInterval and deletes all expired keys.
// - Close stops that goroutine and returns only once it has exited, leaving no goroutine leak.
// - Close is safe to call more than once and safe to call concurrently.
// - Close works when there is no janitor (interval 0).
// 3: LRU
// - The store never holds more than MaxEntries keys.
// - When a Set of a new key would exceed the limit, the least recently used key is evicted first.
// - Both Get (a hit) and Set count as a use. Len and Delete do not.
// - Overwriting an existing key never evicts anything, but it does make that key the most recently used.
// - Get, Set, and eviction are all O(1), with no scanning for the oldest key.
// - Every path that removes a key (Delete, lazy expiry in Get, the janitor, eviction) keeps the recency structure in sync with the map.

type Store struct {
	shards  []*shard
	seed    maphash.Seed
	now     func() time.Time
	janitor *periodic.Worker
}

type Config struct {
	CleanupInterval time.Duration // 0 = no janitor

	// MaxEntries is the approximate maximum number of stored entries; 0 means
	// unlimited. The limit is divided evenly across shards, rounded up, so the
	// store may hold slightly more than MaxEntries in total (at most
	// Shards-1 extra). Eviction is per shard: when a key's shard is full, that
	// shard's least recently used entry is evicted, even if the store as a whole
	// is below the limit. Expired entries count until they are removed.
	MaxEntries int

	// Shards is a number of independent partitions; 0 means default 16
	// More shards reduce lock contention; LRU order is per shard.
	Shards int
}

func NewStore(config Config) *Store {
	shardCount := config.Shards
	if config.Shards <= 0 {
		shardCount = 16
	}

	s := &Store{
		shards:  make([]*shard, shardCount),
		seed:    maphash.MakeSeed(),
		now:     time.Now,
		janitor: periodic.New(config.CleanupInterval),
	}

	for i := range s.shards {
		s.shards[i] = &shard{
			items:      map[string]*list.Element{},
			lru:        list.New(),
			maxEntries: (config.MaxEntries + shardCount - 1) / shardCount, // round up the capacity
		}
	}

	s.janitor.Run(s.deleteExpired)

	return s
}

func (s *Store) shardFor(key string) *shard {
	h := maphash.String(s.seed, key)
	return s.shards[h%uint64(len(s.shards))]
}

func (s *Store) deleteExpired() {
	now := s.now()
	for i := range s.shards {
		s.shards[i].deleteExpired(now)
	}
}

// stops the janitor and waits for it to exit
func (s *Store) Close() {
	s.janitor.Stop()
}

// ttl == 0 means the key never expires
func (s *Store) Set(key string, value []byte, ttl time.Duration) {
	deadline := s.deadline(ttl)
	s.shardFor(key).set(key, value, deadline)
}

// ok is false if the key is missing or expired
func (s *Store) Get(key string) (value []byte, ok bool) {
	return s.shardFor(key).get(key, s.now())
}

// reports whether a key was actually removed
func (s *Store) Delete(key string) bool {
	return s.shardFor(key).delete(key, s.now())
}

// counts only live (non-expired) keys
func (s *Store) Len() int {
	now := s.now()
	count := 0
	for i := range s.shards {
		count += s.shards[i].len(now)
	}
	return count
}

func (s *Store) deadline(ttl time.Duration) time.Time {
	if ttl != 0 {
		return deadlineFrom(ttl, s.now())
	}
	return time.Time{}
}

func deadlineFrom(ttl time.Duration, now time.Time) time.Time {
	if ttl != 0 {
		return now.Add(ttl)
	}
	return time.Time{}
}

// Expire sets a new TTL on an existing live key and reports whether the key existed.
// A ttl of 0 clears the deadline, so the key never expires.
// A key that has already expired is deleted and reported as missing.
// Expire does not count as a use for LRU eviction.
func (s *Store) Expire(key string, ttl time.Duration) bool {
	now := s.now()
	deadline := deadlineFrom(ttl, now)
	return s.shardFor(key).expire(key, deadline, now)
}

// TTL returns the remaining time for a key.
// exists is false for a missing or expired key; hasTTL is false for a key with no expiry.
// TTL does not count as a use for LRU eviction.
func (s *Store) TTL(key string) (remaining time.Duration, exists, hasTTL bool) {
	return s.shardFor(key).ttl(key, s.now())
}
