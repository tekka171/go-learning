package training

import "sync"

// Rules:
// Get(key): return the value if the key exists, otherwise return -1. A successful Get marks the key as most recently used.
// Put(key, value): insert the key, or update its value if it already exists. Either way, mark it as most recently used.
// If a Put makes the cache go over capacity, evict the least recently used key.
// Both Get and Put must run in O(1) time.

type LRUCache struct {
	mu       sync.Mutex //to support goroutine, single lock, but heavy & is a bottleneck
	mapCache map[int]*Node
	capacity int
	head     *Node // most recent
	tail     *Node // least recent
}

type Node struct {
	key   int
	value int
	prev  *Node //left side
	next  *Node //right side
}

//Assume capacity always > 0
func Constructor(capacity int) *LRUCache {
	return &LRUCache{
		mapCache: make(map[int]*Node, capacity),
		capacity: capacity,
	}
}

func (c *LRUCache) Get(key int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	currNode, ok := c.mapCache[key]
	if !ok {
		return -1
	}

	if currNode != c.head {
		c.remove(currNode)
		c.addToFront(currNode)
	}

	return currNode.value
}

func (c *LRUCache) addToFront(node *Node) {
	if c.head == nil {
		c.head = node
		c.tail = node
	} else {
		c.head.prev = node
		node.prev = nil
		node.next = c.head
		c.head = node
	}
}

func (c *LRUCache) remove(node *Node) {
	if node.prev == nil && node.next == nil {
		c.head = nil
		c.tail = nil
	} else if c.head == node {
		node.next.prev = nil
		c.head = node.next
	} else if c.tail == node {
		node.prev.next = nil
		c.tail = node.prev
	} else {
		node.next.prev, node.prev.next = node.prev, node.next
	}

	node.prev = nil
	node.next = nil
}

func (c *LRUCache) Put(key int, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	currNode, ok := c.mapCache[key]
	if !ok {
		// if over capacity, evict the least recently used key
		if len(c.mapCache) == c.capacity {
			lru := c.tail
			c.remove(lru)
			delete(c.mapCache, lru.key)
		}

		currNode = &Node{
			key:   key,
			value: value,
		}
		c.addToFront(currNode)
		c.mapCache[key] = currNode
		return
	}

	currNode.value = value
	c.remove(currNode)
	c.addToFront(currNode)
}
