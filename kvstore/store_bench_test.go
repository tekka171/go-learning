package kvstore

import (
	"math/rand/v2"
	"strconv"
	"testing"
)

func BenchmarkSet(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()

	val := []byte("val1")
	for b.Loop() {
		store.Set("key1", val, 0)
	}
}

func BenchmarkSetNewKeys(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()

	keys := make([]string, 100_000)
	for i := range keys {
		keys[i] = "key-" + strconv.Itoa(i)
	}

	i := 0
	val := []byte("val1")
	for b.Loop() {
		store.Set(keys[i%len(keys)], val, 0)
		i++
	}
}

func BenchmarkSetWithEviction(b *testing.B) {
	store := NewStore(Config{MaxEntries: 10_000})
	defer store.Close()

	keys := make([]string, 100_000)
	for i := range keys {
		keys[i] = "key-" + strconv.Itoa(i)
	}

	i := 0
	val := []byte("val1")
	for b.Loop() {
		store.Set(keys[i%len(keys)], val, 0)
		i++
	}
}

func BenchmarkGetMiss(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()

	for b.Loop() {
		store.Get("key1")
	}
}

func BenchmarkGetHit(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()
	store.Set("key1", []byte("val1"), 0)

	for b.Loop() {
		store.Get("key1")
	}
}

func BenchmarkGetManyKeys(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()

	keys := make([]string, 100_000)
	val := []byte("val")
	for i := range keys {
		key := "key_" + strconv.Itoa(i)
		keys[i] = key
		store.Set(key, val, 0)
	}

	for b.Loop() {
		store.Get(keys[rand.IntN(len(keys))])
	}
}

func BenchmarkGetHitParallel(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()
	store.Set("key", []byte("val"), 0)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			store.Get("key")
		}
	})
}

func BenchmarkGetParallelManyKeys(b *testing.B) {
	store := NewStore(Config{})
	defer store.Close()

	keys := make([]string, 100_000)
	val := []byte("val")
	for i := range keys {
		key := "key_" + strconv.Itoa(i)
		keys[i] = key
		store.Set(key, val, 0)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			store.Get(keys[rand.IntN(len(keys))])
		}
	})
}
