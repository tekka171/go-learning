package training

import (
	"hash/fnv"
	"time"
)

type RedisShard struct {
	shards []*Redis
}

func NewRedisShard(shardCount int) *RedisShard {
	//round up odd shardCount
	shardCount = nextPowerOfTwo(shardCount)

	res := &RedisShard{
		shards: make([]*Redis, shardCount),
	}

	for i := range res.shards {
		res.shards[i] = NewRedis()
	}

	return res
}

func nextPowerOfTwo(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func shardIndex(key string, numShards int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & uint32(numShards-1))
}

func (r *RedisShard) Get(key string) (any, error) {
	idx := shardIndex(key, len(r.shards))
	return r.shards[idx].Get(key)
}

func (r *RedisShard) Set(key string, val any, timeout time.Duration) error {
	idx := shardIndex(key, len(r.shards))
	return r.shards[idx].Set(key, val, timeout)
}
