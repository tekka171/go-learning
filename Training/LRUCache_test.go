package training

import "testing"

type lruOp struct {
	op     string // "put" or "get"
	key    int
	value  int // used for "put"
	expect int // used for "get"
}

func TestLRUCache(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		ops      []lruOp
	}{
		{
			name:     "leetcode example",
			capacity: 2,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "put", key: 2, value: 2},
				{op: "get", key: 1, expect: 1},
				{op: "put", key: 3, value: 3}, // evicts key 2
				{op: "get", key: 2, expect: -1},
				{op: "put", key: 4, value: 4}, // evicts key 1
				{op: "get", key: 1, expect: -1},
				{op: "get", key: 3, expect: 3},
				{op: "get", key: 4, expect: 4},
			},
		},
		{
			name:     "get on empty cache",
			capacity: 2,
			ops: []lruOp{
				{op: "get", key: 1, expect: -1},
			},
		},
		{
			name:     "get on missing key",
			capacity: 2,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "get", key: 2, expect: -1},
			},
		},
		{
			name:     "update existing key does not evict",
			capacity: 2,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "put", key: 2, value: 2},
				{op: "put", key: 1, value: 10},
				{op: "get", key: 1, expect: 10},
				{op: "get", key: 2, expect: 2},
			},
		},
		{
			name:     "capacity of one",
			capacity: 1,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "put", key: 2, value: 2}, // evicts key 1
				{op: "get", key: 1, expect: -1},
				{op: "get", key: 2, expect: 2},
			},
		},
		{
			name:     "get refreshes recency to avoid eviction",
			capacity: 2,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "put", key: 2, value: 2},
				{op: "get", key: 1, expect: 1}, // key 1 now most recently used
				{op: "put", key: 3, value: 3},  // evicts key 2, not key 1
				{op: "get", key: 1, expect: 1},
				{op: "get", key: 2, expect: -1},
				{op: "get", key: 3, expect: 3},
			},
		},
		{
			name:     "update same key when only one node exists",
			capacity: 2,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "put", key: 1, value: 100},
				{op: "get", key: 1, expect: 100},
			},
		},
		{
			name:     "repeated eviction across many inserts",
			capacity: 2,
			ops: []lruOp{
				{op: "put", key: 1, value: 1},
				{op: "put", key: 2, value: 2},
				{op: "put", key: 3, value: 3}, // evicts 1
				{op: "put", key: 4, value: 4}, // evicts 2
				{op: "get", key: 1, expect: -1},
				{op: "get", key: 2, expect: -1},
				{op: "get", key: 3, expect: 3},
				{op: "get", key: 4, expect: 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := Constructor(tt.capacity)

			for i, op := range tt.ops {
				switch op.op {
				case "put":
					cache.Put(op.key, op.value)
				case "get":
					got := cache.Get(op.key)
					if got != op.expect {
						t.Fatalf("op %d: Get(%d) = %d, want %d", i, op.key, got, op.expect)
					}
				default:
					t.Fatalf("op %d: unknown op %q", i, op.op)
				}
			}
		})
	}
}
