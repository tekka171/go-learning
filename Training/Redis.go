package training

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired  = errors.New("key expired")
)

type IRedis interface {
	Get(key string) (any, error)
	Set(key string, value any, timeout time.Duration) error
}

type Redis struct {
	mu       sync.RWMutex
	redisMap map[string]RedisValue
}

type RedisValue struct {
	value        any
	enableExpire bool
	expireAt     time.Time
}

func NewRedis() *Redis {
	return &Redis{
		mu:       sync.RWMutex{},
		redisMap: map[string]RedisValue{},
	}
}

func (r *Redis) Get(key string) (any, error) {
	r.mu.RLock()

	val, ok := r.redisMap[key]
	if !ok { //if key not found
		r.mu.RUnlock()
		return nil, ErrKeyNotFound
	}

	if !val.enableExpire || !val.expireAt.Before(time.Now()) { //if key okay
		r.mu.RUnlock()
		return val.value, nil
	}
	r.mu.RUnlock()

	// Key looked expired under the read lock; upgrade to a write lock,
	// re-check (it may have been overwritten by a concurrent Set in the
	// meantime), and evict it if it's still expired.
	r.mu.Lock()
	defer r.mu.Unlock()

	val, ok = r.redisMap[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if val.enableExpire && val.expireAt.Before(time.Now()) {
		delete(r.redisMap, key)
		return nil, ErrKeyExpired
	}

	return val.value, nil
}

func (r *Redis) Set(key string, val any, timeout time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	redisValue := RedisValue{
		value: val,
	}

	if timeout > 0 {
		redisValue.enableExpire = true
		redisValue.expireAt = time.Now().Add(timeout)
	}

	r.redisMap[key] = redisValue

	return nil
}
