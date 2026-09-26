package training

type LRUCacheShard struct {
	shards []*LRUCache
}

// Assume capacity always > 0 & is bigger than shardCount
func ConstructorShard(shardCount int, capacity int) *LRUCacheShard {
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

		s.shards[i] = Constructor(finalCap)
	}

	return s
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
