package kvstore

import (
	"bytes"
	"sync"
	"time"

	"github.com/tekka171/go-learning/periodic"
)

// Requirements
// 1
// - Safe for concurrent use by many goroutines.
// - An expired key must never be returned by Get, even if it hasn't been cleaned up yet (lazy expiration).
// - Set on an existing key overwrites both the value and the TTL.
// - Standard library only.
// 2
// - A goroutine wakes every cleanupInterval and deletes all expired keys.
// - Close stops that goroutine and returns only once it has exited, leaving no goroutine leak.
// - Close is safe to call more than once and safe to call concurrently.
// - Close works when there is no janitor (interval 0).

type Store struct {
	mu      sync.RWMutex
	items   map[string]*item
	now     func() time.Time
	janitor *periodic.Worker
}

type item struct {
	key       string
	value     []byte
	expiresAt time.Time
}

// cleanupInterval == 0 means no janitor (lazy expiration only)
func NewStore(cleanupInterval time.Duration) *Store {
	s := &Store{
		items:   map[string]*item{},
		now:     time.Now,
		janitor: periodic.New(cleanupInterval),
	}

	s.janitor.Run(s.deleteExpired)

	return s
}

func (s *Store) deleteExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for k, v := range s.items {
		if v.expired(now) {
			delete(s.items, k)
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

	it := &item{
		key:   key,
		value: bytes.Clone(value),
	}
	if ttl != 0 {
		it.expiresAt = s.now().Add(ttl)
	}
	s.items[key] = it
}

func (i *item) expired(now time.Time) bool {
	return !i.expiresAt.IsZero() && !now.Before(i.expiresAt)
}

// ok is false if the key is missing or expired
func (s *Store) Get(key string) (value []byte, ok bool) {
	now := s.now()
	s.mu.RLock()
	it, ok := s.items[key]
	if !ok { // if not found
		s.mu.RUnlock()
		return nil, false
	}

	if !it.expired(now) { // if not expired, return true
		s.mu.RUnlock()
		return bytes.Clone(it.value), true
	}
	s.mu.RUnlock()

	//if expired, need to do lazy cleanup
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok = s.items[key]
	if !ok {
		return nil, false
	}

	if it.expired(now) {
		delete(s.items, key)
		return nil, false
	}

	return bytes.Clone(it.value), true
}

// reports whether a key was actually removed
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.items[key]
	if !ok {
		return false
	}

	delete(s.items, key)
	return !it.expired(s.now())
}

// counts only live (non-expired) keys
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	now := s.now()
	for _, v := range s.items {
		if v.expired(now) {
			continue
		}
		total++
	}
	return total
}
