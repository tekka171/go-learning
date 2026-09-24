package training

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRedisSetAndGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
	}{
		{name: "string value", key: "name", value: "gopher"},
		{name: "int value", key: "count", value: 42},
		{name: "slice value", key: "list", value: []int{1, 2, 3}},
		{name: "zero value", key: "zero", value: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRedis()

			if err := r.Set(tt.key, tt.value, 0); err != nil {
				t.Fatalf("unexpected error on Set: %v", err)
			}

			got, err := r.Get(tt.key)
			if err != nil {
				t.Fatalf("unexpected error on Get: %v", err)
			}
			if !equalAny(got, tt.value) {
				t.Fatalf("expected %v, got %v", tt.value, got)
			}
		})
	}
}

func TestRedisGetNotFound(t *testing.T) {
	r := NewRedis()

	_, err := r.Get("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestRedisSetOverwrite(t *testing.T) {
	r := NewRedis()

	if err := r.Set("key", "first", 0); err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}
	if err := r.Set("key", "second", 0); err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	got, err := r.Get("key")
	if err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}
	if got != "second" {
		t.Fatalf("expected %q, got %q", "second", got)
	}
}

func TestRedisNoExpiration(t *testing.T) {
	r := NewRedis()

	if err := r.Set("key", "value", 0); err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	got, err := r.Get("key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "value" {
		t.Fatalf("expected %q, got %q", "value", got)
	}
}

func TestRedisExpiration(t *testing.T) {
	r := NewRedis()

	if err := r.Set("key", "value", 10*time.Millisecond); err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	got, err := r.Get("key")
	if err != nil {
		t.Fatalf("expected key to be valid immediately after Set, got error %v", err)
	}
	if got != "value" {
		t.Fatalf("expected %q, got %q", "value", got)
	}

	time.Sleep(20 * time.Millisecond)

	_, err = r.Get("key")
	if !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("expected ErrKeyExpired, got %v", err)
	}

	// lazy delete should have evicted the key on the previous Get
	_, err = r.Get("key")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after eviction, got %v", err)
	}
}

func TestRedisConcurrentAccess(t *testing.T) {
	r := NewRedis()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			r.Set("key", i, 0)
		}(i)
		go func() {
			defer wg.Done()
			r.Get("key")
		}()
	}

	wg.Wait()
}

func equalAny(a, b any) bool {
	as, aok := a.([]int)
	bs, bok := b.([]int)
	if aok && bok {
		if len(as) != len(bs) {
			return false
		}
		for i := range as {
			if as[i] != bs[i] {
				return false
			}
		}
		return true
	}
	return a == b
}
