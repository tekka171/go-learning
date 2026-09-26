package training

import (
	"errors"
	"sync"
)

// Rules:
// Get(key): return the value if the key exists, otherwise return -1. A successful Get marks the key as most recently used.
// Put(key, value): insert the key, or update its value if it already exists. Either way, mark it as most recently used.
// If a Put makes the cache go over capacity, evict the least recently used key.
// Both Get and Put must run in O(1) time.

type LRUCache struct {
	mu       sync.Mutex //to support goroutine, single lock
	items    map[int]*node
	capacity int
	head     *node // most recent
	tail     *node // least recent
}

type node struct {
	key   int
	value int
	prev  *node //left side
	next  *node //right side
}

var (
	ErrInvalidCapacity = errors.New("capacity must be > 0")
)

func NewLRU(capacity int) (*LRUCache, error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &LRUCache{
		items:    make(map[int]*node, capacity),
		capacity: capacity,
	}, nil
}

func (c *LRUCache) Get(key int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	currNode, ok := c.items[key]
	if !ok {
		return -1
	}

	c.moveToFront(currNode)
	return currNode.value
}

func (c *LRUCache) addToFront(node *node) {
	node.prev = nil
	node.next = c.head
	if c.head == nil {
		c.tail = node
	} else {
		c.head.prev = node
	}
	c.head = node
}

func (c *LRUCache) remove(node *node) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}

	node.prev = nil
	node.next = nil
}

func (c *LRUCache) moveToFront(node *node) {
	if c.head == node {
		return
	}

	c.remove(node)
	c.addToFront(node)
}

func (c *LRUCache) Put(key int, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if currNode, ok := c.items[key]; ok { // if exist, then do update
		currNode.value = value
		c.moveToFront(currNode)
		return
	}

	// if at capacity, evict the least recently used key
	if len(c.items) == c.capacity {
		lru := c.tail
		c.remove(lru)
		delete(c.items, lru.key)
	}

	currNode := &node{
		key:   key,
		value: value,
	}
	c.addToFront(currNode)
	c.items[key] = currNode
}
