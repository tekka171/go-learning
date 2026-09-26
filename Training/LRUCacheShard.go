package training

import "errors"

type LRUCacheShard struct {
	shards []*LRUCache
}

var (
	ErrInvalidShardCount    = errors.New("shardCount must be > 0")
	ErrInvalidShardCapacity = errors.New("capacity must be >= shardCount")
)

func NewLRUSharded(shardCount, capacity int) (*LRUCacheShard, error) {
	if shardCount <= 0 {
		return nil, ErrInvalidShardCount
	}

	if capacity < shardCount {
		return nil, ErrInvalidShardCapacity
	}

	s := &LRUCacheShard{
		shards: make([]*LRUCache, shardCount),
	}

	baseCap := capacity / shardCount
	extraCap := capacity % shardCount

	for i := range s.shards {
		finalCap := baseCap
		if i < extraCap {
			finalCap++ //spread extra cap through the shards
		}

		shard, err := NewLRU(finalCap)
		if err != nil {
			return nil, err
		}
		s.shards[i] = shard
	}

	return s, nil
}

func (c *LRUCacheShard) shardFor(key int) *LRUCache {
	return c.shards[uint(key)%uint(len(c.shards))]
}

func (c *LRUCacheShard) Get(key int) int {
	return c.shardFor(key).Get(key)
}

func (c *LRUCacheShard) Put(key int, value int) {
	c.shardFor(key).Put(key, value)
}
