package testharness

import (
	"fmt"
	"maps"
	"slices"
)

type BidirectionalMap[K, V comparable] struct {
	direct  map[K]V
	inverse map[V]K
}

func (b *BidirectionalMap[K, V]) init() {
	if b.direct == nil {
		b.direct = make(map[K]V)
	}
	if b.inverse == nil {
		b.inverse = make(map[V]K)
	}
}

func (b *BidirectionalMap[K, V]) Put(key K, value V) error {
	if existingValue, ok := b.direct[key]; ok {
		return fmt.Errorf("key %v already bound to value: %v", key, existingValue)
	}
	if existingKey, ok := b.inverse[value]; ok {
		return fmt.Errorf("value %v already bound to key: %v", value, existingKey)
	}

	b.init()

	b.direct[key], b.inverse[value] = value, key
	return nil
}

func (b *BidirectionalMap[K, V]) Putinverse(key V, value K) error {
	if existingValue, ok := b.direct[value]; ok {
		return fmt.Errorf("value %v already bound to key: %v", key, existingValue)
	}
	if existingKey, ok := b.inverse[key]; ok {
		return fmt.Errorf("key %v already bound to value: %v", value, existingKey)
	}

	b.init()

	b.inverse[key], b.direct[value] = value, key
	return nil
}

func (b *BidirectionalMap[K, V]) Get(key K) *V {
	if existingValue, ok := b.direct[key]; ok {
		return &existingValue
	} else {
		return nil
	}
}

func (b *BidirectionalMap[K, V]) GetInverse(key V) *K {
	if existingValue, ok := b.inverse[key]; ok {
		return &existingValue
	} else {
		return nil
	}
}

func (b *BidirectionalMap[K, V]) Delete(key K) {
	if existingValue, ok := b.direct[key]; ok {
		delete(b.direct, key)
		delete(b.inverse, existingValue)
	}
}

func (b *BidirectionalMap[K, V]) DeleteInverse(key V) {
	if existingValue, ok := b.inverse[key]; ok {
		delete(b.direct, existingValue)
		delete(b.inverse, key)
	}
}

func (b *BidirectionalMap[K, V]) Keys() []K {
	return slices.Collect(maps.Keys(b.direct))
}

func (b *BidirectionalMap[K, V]) Values() []V {
	return slices.Collect(maps.Keys(b.inverse))
}
