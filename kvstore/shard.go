package kvstore

import (
	"bytes"
	"container/list"
	"sync"
	"time"
)

type shard struct {
	mu         sync.Mutex
	items      map[string]*list.Element
	lru        *list.List
	maxEntries int
}

type item struct {
	key       string
	value     []byte
	expiresAt time.Time
}

func (i *item) expired(now time.Time) bool {
	return !i.expiresAt.IsZero() && !now.Before(i.expiresAt)
}

func (sh *shard) deleteExpired(now time.Time) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	for _, el := range sh.items {
		if el.Value.(*item).expired(now) {
			sh.removeElement(el)
		}
	}
}

func (sh *shard) set(key string, value []byte, deadline time.Time) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	// do update
	if el, ok := sh.items[key]; ok {
		it := el.Value.(*item)
		it.value = bytes.Clone(value)
		it.expiresAt = deadline

		sh.lru.MoveToFront(el)
		return
	}

	// if at capacity, evict lru even though other keys have expired
	if sh.maxEntries > 0 && len(sh.items) >= sh.maxEntries {
		sh.removeElement(sh.lru.Back())
	}

	//do insert
	it := &item{
		key:       key,
		value:     bytes.Clone(value),
		expiresAt: deadline,
	}

	el := sh.lru.PushFront(it)
	sh.items[key] = el
}

func (sh *shard) get(key string, now time.Time) (value []byte, ok bool) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	el, ok := sh.items[key]
	if !ok { // if not found
		return nil, false
	}

	it := el.Value.(*item)

	//if expired, need to do lazy cleanup
	if it.expired(now) {
		sh.removeElement(el)
		return nil, false
	}

	sh.lru.MoveToFront(el)
	return bytes.Clone(it.value), true
}

func (sh *shard) delete(key string, now time.Time) bool {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	el, ok := sh.items[key]
	if !ok {
		return false
	}

	isExpired := el.Value.(*item).expired(now)
	sh.removeElement(el)

	return !isExpired
}

func (sh *shard) len(now time.Time) int {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	total := 0
	for _, v := range sh.items {
		if v.Value.(*item).expired(now) {
			continue
		}
		total++
	}
	return total
}

func (sh *shard) removeElement(el *list.Element) {
	it := el.Value.(*item)
	delete(sh.items, it.key)
	sh.lru.Remove(el)
}

func (sh *shard) expire(key string, deadline time.Time, now time.Time) bool {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	el, ok := sh.items[key]
	if !ok {
		return false
	}

	it := el.Value.(*item)
	if it.expired(now) {
		sh.removeElement(el)
		return false
	}

	it.expiresAt = deadline
	return true
}

func (sh *shard) ttl(key string, now time.Time) (remaining time.Duration, exists, hasTTL bool) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	el, ok := sh.items[key]
	if !ok {
		return 0, false, false
	}

	it := el.Value.(*item)
	if it.expired(now) {
		return 0, false, false
	}

	if it.expiresAt.IsZero() {
		return 0, true, false
	}

	return it.expiresAt.Sub(now), true, true
}
