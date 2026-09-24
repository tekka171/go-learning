package training

import (
	"hash/fnv"
	"sync"
	"time"
)

type RedisShard struct {
	shards []*RedisShardValue
}

type RedisShardValue struct {
	mu       sync.RWMutex
	redisMap map[string]RedisValue
}

func NewRedisShard(shardSize int) *RedisShard {
	//round up odd shardSize
	shardSize = nextPowerOfTwo(shardSize)

	res := &RedisShard{
		shards: make([]*RedisShardValue, shardSize),
	}

	for i := range res.shards {
		res.shards[i] = &RedisShardValue{
			mu:       sync.RWMutex{},
			redisMap: map[string]RedisValue{},
		}
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
	shard := r.shards[idx]

	shard.mu.RLock()

	val, ok := shard.redisMap[key]
	if !ok { //if key not found
		shard.mu.RUnlock()
		return nil, ErrKeyNotFound
	}

	if !val.enableExpire || !val.expireAt.Before(time.Now()) { //if key okay
		shard.mu.RUnlock()
		return val.value, nil
	}
	shard.mu.RUnlock()

	// Key looked expired under the read lock; upgrade to a write lock,
	// re-check (it may have been overwritten by a concurrent Set in the
	// meantime), and evict it if it's still expired.
	shard.mu.Lock()
	defer shard.mu.Unlock()

	val, ok = shard.redisMap[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if val.enableExpire && val.expireAt.Before(time.Now()) {
		delete(shard.redisMap, key)
		return nil, ErrKeyExpired
	}

	return val.value, nil
}

func (r *RedisShard) Set(key string, val any, timeout time.Duration) error {
	idx := shardIndex(key, len(r.shards))
	shard := r.shards[idx]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	res := RedisValue{
		value: val,
	}

	if timeout > 0 {
		res.enableExpire = true
		res.expireAt = time.Now().Add(timeout)
	}

	shard.redisMap[key] = res

	return nil
}
