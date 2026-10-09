package kvstore

import (
	"bytes"
	"container/list"
	"sync"
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
	mu      sync.RWMutex
	items   map[string]*list.Element
	lru     *list.List
	now     func() time.Time
	janitor *periodic.Worker
	config  Config
}

type item struct {
	key       string
	value     []byte
	expiresAt time.Time
}

type Config struct {
	CleanupInterval time.Duration // 0 = no janitor

	// MaxEntries is the maximum number of stored entries; 0 means unlimited.
	// Expired entries count until they are removed.
	MaxEntries int
}

func NewStore(config Config) *Store {
	s := &Store{
		items:   map[string]*list.Element{},
		lru:     list.New(),
		now:     time.Now,
		config:  config,
		janitor: periodic.New(config.CleanupInterval),
	}

	s.janitor.Run(s.deleteExpired)

	return s
}

func (s *Store) deleteExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for _, el := range s.items {
		if el.Value.(*item).expired(now) {
			s.removeElement(el)
		}
	}
}

// stops the janitor and waits for it to exit
func (s *Store) Close() {
	s.janitor.Stop()
}

// ttl == 0 means the key never expires
func (s *Store) Set(key string, value []byte, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// do update
	if el, ok := s.items[key]; ok {
		it := el.Value.(*item)
		it.value = bytes.Clone(value)
		it.expiresAt = s.deadline(ttl)

		s.lru.MoveToFront(el)
		return
	}

	// if at capacity, evict lru even though other keys have expired
	if s.config.MaxEntries > 0 && len(s.items) >= s.config.MaxEntries {
		s.removeElement(s.lru.Back())
	}

	//do insert
	it := &item{
		key:       key,
		value:     bytes.Clone(value),
		expiresAt: s.deadline(ttl),
	}

	el := s.lru.PushFront(it)
	s.items[key] = el
}

func (i *item) expired(now time.Time) bool {
	return !i.expiresAt.IsZero() && !now.Before(i.expiresAt)
}

// ok is false if the key is missing or expired
func (s *Store) Get(key string) (value []byte, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	el, ok := s.items[key]
	if !ok { // if not found
		return nil, false
	}

	it := el.Value.(*item)

	//if expired, need to do lazy cleanup
	if it.expired(now) {
		s.removeElement(el)
		return nil, false
	}

	s.lru.MoveToFront(el)
	return bytes.Clone(it.value), true
}

// reports whether a key was actually removed
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	el, ok := s.items[key]
	if !ok {
		return false
	}

	isExpired := el.Value.(*item).expired(s.now())
	s.removeElement(el)

	return !isExpired
}

// counts only live (non-expired) keys
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	now := s.now()
	for _, v := range s.items {
		if v.Value.(*item).expired(now) {
			continue
		}
		total++
	}
	return total
}

func (s *Store) removeElement(el *list.Element) {
	it := el.Value.(*item)
	delete(s.items, it.key)
	s.lru.Remove(el)
}

func (s *Store) deadline(ttl time.Duration) time.Time {
	if ttl != 0 {
		return s.now().Add(ttl)
	}
	return time.Time{}
}
