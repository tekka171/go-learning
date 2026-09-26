package training

import (
	"sync"
)

// Rules:
// Get(key): return the value if the key exists, otherwise return -1. A successful Get marks the key as most recently used.
// Put(key, value): insert the key, or update its value if it already exists. Either way, mark it as most recently used.
// If a Put makes the cache go over capacity, evict the least recently used key.
// Both Get and Put must run in O(1) time.

type LRUCacheSentinel struct {
	mu       sync.Mutex
	items    map[int]*nodeV2
	capacity int
	head     *nodeV2
	tail     *nodeV2
}

type nodeV2 struct {
	key   int
	value int
	prev  *nodeV2
	next  *nodeV2
}

func NewLRUCacheSentinel(capacity int) *LRUCacheSentinel {
	head, tail := &nodeV2{}, &nodeV2{}
	head.next = tail
	tail.prev = head

	return &LRUCacheSentinel{
		items:    make(map[int]*nodeV2, capacity),
		capacity: capacity,
		head:     head,
		tail:     tail,
	}
}

func (c *LRUCacheSentinel) remove(n *nodeV2) {
	n.prev.next = n.next // re-link left side
	n.next.prev = n.prev // re-link right side
	n.prev, n.next = nil, nil
}

func (c *LRUCacheSentinel) addToFront(n *nodeV2) {
	n.prev = c.head      // link node's left side
	n.next = c.head.next // link node's right side
	c.head.next.prev = n // re-link right node
	c.head.next = n      // re-link left node
}

func (c *LRUCacheSentinel) moveToFront(n *nodeV2) {
	if n.prev == c.head {
		return
	}

	c.remove(n)
	c.addToFront(n)
}

func (c *LRUCacheSentinel) Get(key int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.items[key]
	if !ok {
		return -1
	}

	c.moveToFront(n)
	return n.value
}

func (c *LRUCacheSentinel) Put(key int, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.items[key]; ok { // if exist, do update
		n.value = value
		c.moveToFront(n)
		return
	}

	if c.capacity == len(c.items) { // if at capacity, do evict
		lru := c.tail.prev
		c.remove(lru)
		delete(c.items, lru.key)
	}

	n := &nodeV2{
		key:   key,
		value: value,
	}
	c.addToFront(n)
	c.items[key] = n
}
