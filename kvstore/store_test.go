package kvstore

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

// fake clock: tests in the same package can replace s.now
func newTestStore() (*Store, *time.Time) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewStore(0)
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
	s := NewStore(0)
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
