package kvstore

import (
	"bytes"
	"strconv"
	"sync"
	"testing"
	"time"
)

// fake clock: tests in the same package can replace s.now
func newTestStore() (*Store, *time.Time) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewStore(Config{})
	s.now = func() time.Time { return now }
	return s, &now
}

func TestExpiry(t *testing.T) {
	tests := []struct {
		name    string
		ttl     time.Duration
		advance time.Duration
		wantOK  bool
	}{
		{"before deadline", time.Minute, 59 * time.Second, true},
		{"at deadline", time.Minute, time.Minute, false},
		{"after deadline", time.Minute, 2 * time.Minute, false},
		{"no deadline", 0, 100 * time.Second, true},
		{"no deadline, long after", 0, 365 * 24 * time.Hour, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, now := newTestStore()
			s.Set("k", []byte("v"), tc.ttl)
			*now = now.Add(tc.advance)
			_, ok := s.Get("k")
			if ok != tc.wantOK {
				t.Fatalf("Get ok = %v, want %v", ok, tc.wantOK)
			}
		})
	}
}

func TestExpirySetOverwritesTTL(t *testing.T) {
	tests := []struct {
		name      string
		firstTTL  time.Duration
		secondTTL time.Duration
		advance   time.Duration
		wantOK    bool
		wantValue []byte
	}{
		{
			name:      "overwrite removes expiration",
			firstTTL:  5 * time.Second,
			secondTTL: 0,
			advance:   10 * time.Second,
			wantOK:    true,
			wantValue: []byte("second"),
		},
		{
			name:      "overwrite re-introduces expiration",
			firstTTL:  0,
			secondTTL: 5 * time.Second,
			advance:   6 * time.Second,
			wantOK:    false,
			wantValue: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, now := newTestStore()
			s.Set("k", []byte("first"), tc.firstTTL)
			s.Set("k", []byte("second"), tc.secondTTL)
			*now = now.Add(tc.advance)

			got, ok := s.Get("k")
			if ok != tc.wantOK {
				t.Fatalf("Get ok = %v, want %v", ok, tc.wantOK)
			}
			if !bytes.Equal(got, tc.wantValue) {
				t.Fatalf("Get value = %v, want %v", got, tc.wantValue)
			}
		})
	}
}

func TestBasic(t *testing.T) {
	tests := []struct {
		name          string
		setKey        string
		setValue      []byte
		getKey        string
		expectedValue []byte
		setTTL        time.Duration
		wantOK        bool
	}{
		{"normal get after set", "k", []byte("v"), "k", []byte("v"), 10 * time.Second, true},
		{"get on a different missing key", "k", []byte("v"), "missing", nil, 10 * time.Second, false},
		{"empty value round-trips", "k", []byte{}, "k", []byte{}, 0, true},
		{"binary value round-trips", "k", []byte{0x00, 0xFF, 0x10}, "k", []byte{0x00, 0xFF, 0x10}, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newTestStore()
			s.Set(tc.setKey, tc.setValue, tc.setTTL)

			got, ok := s.Get(tc.getKey)
			if ok != tc.wantOK {
				t.Fatalf("Get ok = %v, want %v", ok, tc.wantOK)
			}
			if !bytes.Equal(got, tc.expectedValue) {
				t.Fatalf("Get value = %v, want %v", got, tc.expectedValue)
			}
		})
	}
}

func TestBasicSetOverwritesValue(t *testing.T) {
	s, _ := newTestStore()

	s.Set("k", []byte("first"), 0)
	s.Set("k", []byte("second"), 0)

	got, ok := s.Get("k")
	if !ok {
		t.Fatalf("expected key to be found")
	}
	if !bytes.Equal(got, []byte("second")) {
		t.Fatalf("Get value = %q, want %q", got, "second")
	}
}

func TestSetACopy(t *testing.T) {
	s, _ := newTestStore()

	value := []byte("hello")
	s.Set("k", value, 0)
	value[0] = 'X' //write into the original array

	got, ok := s.Get("k")
	if !ok {
		t.Fatalf("expected key to be found")
	}

	if !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("Set value should not mutating the stored value, got %v", string(got))
	}
}

func TestDeleteExpiredKey(t *testing.T) {
	s, now := newTestStore()

	s.Set("k", []byte{'v'}, 100*time.Second)

	*now = now.Add(200 * time.Second)

	isDeleted := s.Delete("k")
	if isDeleted {
		t.Fatalf("Delete on an expired key should return false")
	}

	if _, ok := s.items["k"]; ok {
		t.Fatalf("Expired key should be removed from map by Delete")
	}
}

func TestExtendAliveTTL(t *testing.T) {
	s, now := newTestStore()

	// set ttl 1 minute
	s.Set("k", []byte("hello"), time.Minute)

	// advance time 30 sec
	*now = now.Add(30 * time.Second)

	// set ttl 1 minute
	s.Set("k", []byte("hello 2"), time.Minute)

	// advance time 40 sec
	*now = now.Add(40 * time.Second)

	got, ok := s.Get("k")
	if !ok {
		t.Fatalf("expected key to be found")
	}

	expectedValue := []byte("hello 2")
	if !bytes.Equal(got, expectedValue) {
		t.Fatalf("Get value = %q, want %q", got, expectedValue)
	}
}

func TestGetReturnsACopy(t *testing.T) {
	s, _ := newTestStore()

	s.Set("k", []byte("hello"), 0)

	got, ok := s.Get("k")
	if !ok {
		t.Fatalf("expected key to be found")
	}

	got[0] = 'X' // mutate the returned slice
	again, _ := s.Get("k")
	if !bytes.Equal(again, []byte("hello")) {
		t.Fatalf("Get must return a copy; mutating the result changed stored value, got %v", again)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(s *Store)
		deleteKey  string
		wantRemove bool
	}{
		{
			name:       "existing key",
			setup:      func(s *Store) { s.Set("k", []byte("v"), 0) },
			deleteKey:  "k",
			wantRemove: true,
		},
		{
			name:       "key that never existed",
			setup:      func(s *Store) {},
			deleteKey:  "missing",
			wantRemove: false,
		},
		{
			name: "already-deleted key",
			setup: func(s *Store) {
				s.Set("k", []byte("v"), 0)
				s.Delete("k")
			},
			deleteKey:  "k",
			wantRemove: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newTestStore()
			tc.setup(s)

			removed := s.Delete(tc.deleteKey)
			if removed != tc.wantRemove {
				t.Fatalf("Delete = %v, want %v", removed, tc.wantRemove)
			}

			if _, ok := s.Get(tc.deleteKey); ok {
				t.Fatalf("expected %q to be gone after Delete", tc.deleteKey)
			}
		})
	}
}

func TestLen(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(s *Store)
		advance time.Duration
		wantLen int
	}{
		{"empty store", func(s *Store) {}, 0, 0},
		{
			name: "several live keys",
			setup: func(s *Store) {
				s.Set("a", []byte("1"), 0)
				s.Set("b", []byte("2"), 0)
				s.Set("c", []byte("3"), 5*time.Second)
			},
			advance: 0,
			wantLen: 3,
		},
		{
			name: "expired keys are not counted",
			setup: func(s *Store) {
				s.Set("a", []byte("1"), 0)
				s.Set("b", []byte("2"), 0)
				s.Set("c", []byte("3"), 5*time.Second)
			},
			advance: 6 * time.Second,
			wantLen: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, now := newTestStore()
			tc.setup(s)
			*now = now.Add(tc.advance)

			if got := s.Len(); got != tc.wantLen {
				t.Fatalf("Len() = %d, want %d", got, tc.wantLen)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	s, _ := newTestStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(4)
		go func(i int) {
			defer wg.Done()
			s.Set("key", []byte{byte(i)}, 0)
		}(i)
		go func() {
			defer wg.Done()
			s.Get("key")
		}()
		go func() {
			defer wg.Done()
			s.Len()
		}()
		go func() {
			defer wg.Done()
			s.Delete("key")
		}()
	}
	wg.Wait()
}

func TestConcurrentAccessWithExpiry(t *testing.T) {
	s := NewStore(Config{})
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(4)
		go func(i int) {
			defer wg.Done()
			s.Set("key", []byte{byte(i)}, time.Nanosecond)
		}(i)
		go func() {
			defer wg.Done()
			s.Get("key")
		}()
		go func() {
			defer wg.Done()
			s.Len()
		}()
		go func() {
			defer wg.Done()
			s.Delete("key")
		}()
	}
	wg.Wait()
}

func TestCloseWithZeroInterval(t *testing.T) {
	s := NewStore(Config{})
	s.Close()
}

func TestCloseMoreThanOnce(t *testing.T) {
	s := NewStore(Config{CleanupInterval: 10 * time.Second})
	s.Close()
	s.Close()
}

func TestJanitorWorks(t *testing.T) {
	s := NewStore(Config{CleanupInterval: 5 * time.Millisecond})
	defer s.Close()

	s.Set("k", []byte("hello"), 1*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	s.mu.RLock()
	if _, ok := s.items["k"]; ok {
		t.Fatalf("item should be cleanup already by janitor")
	}
	s.mu.RUnlock()
}

func TestDeleteExpired(t *testing.T) {
	s, now := newTestStore()
	s.Set("k1", []byte("hello 1"), 10*time.Second)
	s.Set("k2", []byte("hello 2"), 100*time.Second)
	s.Set("k3", []byte("hello 3"), 0)

	*now = now.Add(50 * time.Second)
	s.deleteExpired()

	if got := len(s.items); got != 2 {
		t.Fatalf("len(items) = %d, want 2", got)
	}

	if _, ok := s.items["k1"]; ok {
		t.Fatalf("k1 is expired and should have been removed")
	}

	if _, ok := s.items["k2"]; !ok {
		t.Fatalf("k2 is alive and should have been kept")
	}

	if _, ok := s.items["k3"]; !ok {
		t.Fatalf("k3 has no TTL and should have been kept")
	}
}

func TestEvictOldestAtCapacity(t *testing.T) {
	s := NewStore(Config{MaxEntries: 2})
	s.Set("k1", []byte("hello 1"), 0)
	s.Set("k2", []byte("hello 2"), 0)
	s.Set("k3", []byte("hello 3"), 0)

	if got := len(s.items); got != 2 {
		t.Fatalf("len(items) = %d, want 2", got)
	}

	if _, ok := s.items["k1"]; ok {
		t.Fatalf("k1 is the oldest and should be evicted")
	}

	if _, ok := s.items["k2"]; !ok {
		t.Fatalf("k2 is not the oldest and should exist")
	}

	if _, ok := s.items["k3"]; !ok {
		t.Fatalf("k3 is the newest and should exist")
	}
}

func TestGetRefreshRecency(t *testing.T) {
	s := NewStore(Config{MaxEntries: 3})
	s.Set("k1", []byte("hello 1"), 0)
	s.Set("k2", []byte("hello 2"), 0)
	s.Set("k3", []byte("hello 3"), 0)

	//currently k1 is oldest, do Get() & it should be the most recent
	if _, ok := s.Get("k1"); !ok {
		t.Fatalf("k1 is should exist")
	}

	if mru := s.lru.Front(); mru.Value.(*item).key != "k1" {
		t.Fatalf("k1 should be the most recent instead of %+v", mru.Value)
	}

	if lru := s.lru.Back(); lru.Value.(*item).key != "k2" {
		t.Fatalf("k2 should be the least recent instead of %+v", lru.Value)
	}
}

func TestOverwriteAtCapacity(t *testing.T) {
	s := NewStore(Config{MaxEntries: 3})
	s.Set("k1", []byte("hello 1"), 0)
	s.Set("k2", []byte("hello 2"), 0)
	s.Set("k3", []byte("hello 3"), 0)

	//currently k2 is in the middle, overwrite with Set() & it should be the most recent
	s.Set("k2", []byte("hello 2 new"), 0)

	if got := len(s.items); got != 3 {
		t.Fatalf("len(items) = %d, want 3", got)
	}

	if mru := s.lru.Front(); mru.Value.(*item).key != "k2" {
		t.Fatalf("k2 should be the most recent instead of %+v", mru.Value)
	}

	if lru := s.lru.Back(); lru.Value.(*item).key != "k1" {
		t.Fatalf("k1 should be the least recent instead of %+v", lru.Value)
	}

	got, ok := s.items["k2"]
	if !ok {
		t.Fatalf("k2 should still exist after overwrite")
	}

	if !bytes.Equal(got.Value.(*item).value, []byte("hello 2 new")) {
		t.Fatalf("k2 value = %q, want 'hello 2 new'", got.Value.(*item).value)
	}
}

func TestMaxEntries(t *testing.T) {
	s := NewStore(Config{MaxEntries: 1})
	s.Set("k1", []byte("hello 1"), 0)
	s.Set("k2", []byte("hello 2"), 0)
	s.Set("k3", []byte("hello 3"), 0)

	if got := len(s.items); got != 1 {
		t.Fatalf("len(items) = %d, want 1", got)
	}

	if _, ok := s.items["k1"]; ok {
		t.Fatalf("k1 should be evicted")
	}

	if _, ok := s.items["k2"]; ok {
		t.Fatalf("k2 should be evicted")
	}

	if _, ok := s.items["k3"]; !ok {
		t.Fatalf("k3 is the newest and should be the one kept")
	}
}

func TestMaxEntriesUnlimited(t *testing.T) {
	s := NewStore(Config{MaxEntries: 0})
	for i := range 1000 {
		s.Set("k"+strconv.Itoa(i), []byte("hello "), 0)
	}

	if got := len(s.items); got != 1000 {
		t.Fatalf("total items %d does not match maxEntries %d", got, 1000)
	}
}

func TestDeleteFreeSlot(t *testing.T) {
	s := NewStore(Config{MaxEntries: 2})
	s.Set("k1", []byte("hello 1"), 0)
	s.Set("k2", []byte("hello 2"), 0)
	s.Delete("k2")
	s.Set("k3", []byte("hello 3"), 0)

	if got := len(s.items); got != 2 {
		t.Fatalf("len(items) = %d, want 2", got)
	}

	if _, ok := s.items["k1"]; !ok {
		t.Fatalf("k1 should exist")
	}

	if _, ok := s.items["k3"]; !ok {
		t.Fatalf("k3 should exist since still within capacity")
	}
}

func TestMapAndListInSync(t *testing.T) {
	s := NewStore(Config{MaxEntries: 3})
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	// set 3 entries
	s.Set("k1", []byte("hello 1"), 10*time.Millisecond)
	s.Set("k2", []byte("hello 2"), time.Second)
	s.Set("k3", []byte("hello 3"), time.Minute)
	assertInSync(t, s, "Set 3 entries")
	if got := len(s.items); got != 3 {
		t.Fatalf("len(items) = %d, want 3", got)
	}

	// overwrite 1
	s.Set("k2", []byte("hello 2 new"), 2*time.Second)
	s.Get("k2")
	assertInSync(t, s, "Overwrite & Get k2")

	// delete 1
	s.Delete("k2")
	assertInSync(t, s, "Delete k2")
	if got := len(s.items); got != 2 {
		t.Fatalf("after delete, len(items) = %d, want 2", got)
	}

	// advance time to lazy expire 1
	now = now.Add(100 * time.Millisecond)
	s.Get("k1")
	assertInSync(t, s, "Push time & Get to trigger lazy expire on k1")
	if got := len(s.items); got != 1 {
		t.Fatalf("after lazy expire, len(items) = %d, want 1", got)
	}

	// trigger janitor
	now = now.Add(10 * time.Minute)
	s.deleteExpired()
	assertInSync(t, s, "Push time again & trigger janitor")
	if got := len(s.items); got != 0 {
		t.Fatalf("after janitor cleanup, len(items) = %d, want 0", got)
	}

	// set 4 to trigger eviction
	s.Set("k11", []byte("hello 11"), 10*time.Millisecond)
	s.Set("k12", []byte("hello 12"), time.Second)
	s.Set("k13", []byte("hello 13"), time.Minute)
	s.Set("k14", []byte("hello 14"), time.Minute)
	assertInSync(t, s, "Set 4 & evict")
	if got := len(s.items); got != 3 {
		t.Fatalf("after eviction, len(items) = %d, want 3", got)
	}

}

func assertInSync(t *testing.T, s *Store, step string) {
	t.Helper()
	if s.lru.Len() != len(s.items) {
		t.Fatalf("%s: lru.Len() = %d, len(items) = %d", step, s.lru.Len(), len(s.items))
	}
}

func TestTTL(t *testing.T) {
	tests := []struct {
		name          string
		setKey        string
		ttlKey        string
		ttl           time.Duration
		advance       time.Duration
		wantRemaining time.Duration
		wantExists    bool
		wantHasTTL    bool
	}{
		{
			name:          "missing key",
			setKey:        "k",
			ttlKey:        "k_miss",
			ttl:           100 * time.Second,
			advance:       10 * time.Second,
			wantRemaining: 0,
			wantExists:    false,
			wantHasTTL:    false,
		},
		{
			name:          "no-TTL key",
			setKey:        "k",
			ttlKey:        "k",
			ttl:           0,
			advance:       10 * time.Second,
			wantRemaining: 0,
			wantExists:    true,
			wantHasTTL:    false,
		},
		{
			name:          "valid key with remaining",
			setKey:        "k",
			ttlKey:        "k",
			ttl:           100 * time.Second,
			advance:       90 * time.Second,
			wantRemaining: 10 * time.Second,
			wantExists:    true,
			wantHasTTL:    true,
		},
		{
			name:          "expired key",
			setKey:        "k",
			ttlKey:        "k",
			ttl:           10 * time.Second,
			advance:       100 * time.Second,
			wantRemaining: 0,
			wantExists:    false,
			wantHasTTL:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, now := newTestStore()

			s.Set(tc.setKey, []byte("first"), tc.ttl)

			*now = now.Add(tc.advance)

			remaining, exists, hasTTL := s.TTL(tc.ttlKey)

			if remaining != tc.wantRemaining {
				t.Fatalf("remaining = %v, want %v", remaining, tc.wantRemaining)
			}
			if exists != tc.wantExists {
				t.Fatalf("exists = %v, want %v", exists, tc.wantExists)
			}
			if hasTTL != tc.wantHasTTL {
				t.Fatalf("hasTTL = %v, want %v", hasTTL, tc.wantHasTTL)
			}
		})
	}
}

func TestExpireMissingKey(t *testing.T) {
	s := NewStore(Config{})
	s.Set("k", []byte("hello"), 0)
	if got := s.Expire("k_miss", 100*time.Second); got {
		t.Fatalf("Expire on missing key returned true")
	}

	if len(s.items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(s.items))
	}

	if _, ok := s.items["k"]; !ok {
		t.Fatalf("Key 'k' should exists")
	}

	if _, ok := s.items["k_miss"]; ok {
		t.Fatalf("Key 'k_miss' should not exist since not Set() yet")
	}
}

func TestExpireIsExpired(t *testing.T) {
	s, now := newTestStore()
	s.Set("k", []byte("hello"), 10*time.Second)

	*now = now.Add(100 * time.Second)

	if got := s.Expire("k", 100*time.Second); got {
		t.Fatalf("Key should be expired")
	}

	if _, ok := s.items["k"]; ok {
		t.Fatalf("expired key should be removed by Expire")
	}
}

func TestExpire(t *testing.T) {
	s, now := newTestStore()
	s.Set("k", []byte("hello"), 100*time.Second)

	*now = now.Add(10 * time.Second)

	if got := s.Expire("k", 500*time.Second); !got {
		t.Fatalf("Key expire should be updated since not expired yet")
	}

	val, ok := s.items["k"]
	if !ok {
		t.Fatalf("key not found after Expire")
	}

	if want := now.Add(500 * time.Second); !val.Value.(*item).expiresAt.Equal(want) {
		t.Fatalf("Expires_at = %v, want %v", val.Value.(*item).expiresAt, want)
	}
}

func TestExpireToNoTTL(t *testing.T) {
	s, now := newTestStore()
	s.Set("k", []byte("hello"), 10*time.Second)
	if !s.Expire("k", 0) {
		t.Fatalf("Expire on a live key returned false")
	}

	*now = now.Add(1000 * time.Hour)

	if _, ok := s.Get("k"); !ok {
		t.Fatalf("non expired key returned false")
	}

	remaining, exists, hasTTL := s.TTL("k")
	if remaining != 0 {
		t.Fatalf("remaining = %v, want 0", remaining)
	}

	if !exists {
		t.Fatalf("exists = %v, want true", exists)
	}

	if hasTTL {
		t.Fatalf("Non expired key should have no TTL")
	}
}

func TestRecencyUnchanged(t *testing.T) {
	s := NewStore(Config{MaxEntries: 3})
	s.Set("k1", []byte("hello 1"), 0)
	s.Set("k2", []byte("hello 2"), 0)
	s.Set("k3", []byte("hello 3"), 0)

	//TTL & Expire should not affect lru
	s.TTL("k1")
	if !s.Expire("k1", 100*time.Second) {
		t.Fatalf("Expire should returned true since key still exist")
	}

	s.Set("k4", []byte("hello 4"), 0) //add new, should evict the oldest 'k1'

	if _, ok := s.items["k1"]; ok {
		t.Fatalf("k1 should be evicted since is LRU")
	}

	if _, ok := s.items["k2"]; !ok {
		t.Fatalf("k2 should exist")
	}

	if _, ok := s.items["k3"]; !ok {
		t.Fatalf("k3 should exist")
	}

	if _, ok := s.items["k4"]; !ok {
		t.Fatalf("k4 should exist")
	}
}
