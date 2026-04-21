package util

import (
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "0 size cache",
			size: 0,
		},
		{
			name: "1 size cache",
			size: 1,
		},
		{
			name: "100 size cache",
			size: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := Cache[int, int]{}
			cache.Init(test.size)

			if len(cache.queue) != test.size {
				t.Errorf("expected queue length of %d, got %d", test.size, len(cache.queue))
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name string
		hit  bool
	}{
		{name: "cache hit", hit: true},
		{name: "cache miss", hit: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := Cache[int, int]{}
			cache.Init(1)

			if test.hit {
				cache.Set(1, 1)
			}

			value, ok := cache.Get(1)

			if ok != test.hit {
				t.Errorf("expected ok to be %v", test.hit)
			}
			if ok && value != 1 {
				t.Errorf("expected value to be 1, got %d", value)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name      string
		firstKey  int
		firstVal  int
		secondKey int
		secondVal int
		finalLen  int
	}{
		{
			name:      "two unique keys",
			firstKey:  1,
			firstVal:  0,
			secondKey: 2,
			secondVal: 1,
			finalLen:  2,
		},
		{
			name:      "overwrite key",
			firstKey:  1,
			firstVal:  0,
			secondKey: 1,
			secondVal: 1,
			finalLen:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := Cache[int, int]{}
			cache.Init(2)

			cache.Set(test.firstKey, test.firstVal)
			cache.Set(test.secondKey, test.secondVal)

			if len(cache.cache) != test.finalLen {
				t.Errorf(
					"expected queue length of %d, got %d: %v",
					test.finalLen, len(cache.cache), cache.queue,
				)
			}
		})
	}
}

func TestEnqueue(t *testing.T) {
	tests := []struct {
		name         string
		size         int
		writes       int
		containFirst bool
	}{
		{
			name:         "saturated cache",
			size:         10,
			writes:       10,
			containFirst: true,
		},
		{
			name:         "write to 1 size cache",
			size:         1,
			writes:       1,
			containFirst: true,
		},
		{
			name:         "write to saturated cache",
			size:         10,
			writes:       11,
			containFirst: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := Cache[int, int]{}
			cache.Init(test.size)

			for i := range test.writes {
				cache.cache[i] = i
				cache.enqueue(i)
			}

			if test.containFirst && cache.queue[0] != 0 {
				t.Errorf("expected first value to be 0, got %d", cache.queue[0])
			} else if !test.containFirst && cache.queue[0] == 0 {
				t.Errorf("expected first value to not be 0, got %d", cache.queue[0])
			}
		})
	}
}

func TestCache(t *testing.T) {
	type operation struct {
		op  string
		key int
		val int
		ok  bool
	}

	tests := []struct {
		name string
		size int
		ops  []operation
	}{
		{
			name: "basic set and get",
			size: 2,
			ops: []operation{
				{op: "get", key: 1, ok: false},
				{op: "set", key: 1, val: 100},
				{op: "get", key: 1, val: 100, ok: true},
			},
		},
		{
			name: "overwrite existing key",
			size: 2,
			ops: []operation{
				{op: "set", key: 1, val: 10},
				{op: "set", key: 1, val: 99},
				{op: "get", key: 1, val: 99, ok: true},
			},
		},
		{
			name: "eviction of oldest key on saturation",
			size: 2,
			ops: []operation{
				{op: "set", key: 1, val: 10},
				{op: "set", key: 2, val: 20},
				{op: "set", key: 3, val: 30},
				{op: "get", key: 1, ok: false},
				{op: "get", key: 2, val: 20, ok: true},
				{op: "get", key: 3, val: 30, ok: true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := Cache[int, int]{}
			cache.Init(test.size)

			for i, operation := range test.ops {
				switch operation.op {
				case "set":
					cache.Set(operation.key, operation.val)
				case "get":
					val, ok := cache.Get(operation.key)

					if ok != operation.ok {
						t.Errorf("step %d: expected ok to be %v for key %d, got %v", i, operation.ok, operation.key, ok)
					} else if ok && val != operation.val {
						t.Errorf("step %d: expected value to be %d for key %d, got %d", i, operation.val, operation.key, val)
					}
				default:
					t.Fatalf("unknown operation %q", operation.op)
				}
			}
		})
	}
}
