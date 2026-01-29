package geneticalgorithm

import (
	"fmt"
	"maps"
	"slices"
)

// BidirectionalMap implements a two way mapping using two unique maps.
type BidirectionalMap[K, V comparable] struct {
	direct  map[K]V // The map from key to value
	inverse map[V]K // The map from value to key
}

// Initialises the map components.
func (b *BidirectionalMap[K, V]) init() {
	if b.direct == nil {
		b.direct = make(map[K]V)
	}
	if b.inverse == nil {
		b.inverse = make(map[V]K)
	}
}

// Put adds a key value pair to the map.
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

// PutInverse adds a value key pair to the map.
//
// The same as calling Put with reversed args.
func (b *BidirectionalMap[K, V]) PutInverse(key V, value K) error {
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

// Get accesses the map by key to obtain a value.
func (b *BidirectionalMap[K, V]) Get(key K) *V {
	if existingValue, ok := b.direct[key]; ok {
		return &existingValue
	} else {
		return nil
	}
}

// GetInverse accesses the map by value to obtain a key.
func (b *BidirectionalMap[K, V]) GetInverse(key V) *K {
	if existingValue, ok := b.inverse[key]; ok {
		return &existingValue
	} else {
		return nil
	}
}

// Delete removes a key mapping from a bidirectional map.
func (b *BidirectionalMap[K, V]) Delete(key K) {
	if existingValue, ok := b.direct[key]; ok {
		delete(b.direct, key)
		delete(b.inverse, existingValue)
	}
}

// DeleteInverse removes a value mapping from a bidirectional map.
func (b *BidirectionalMap[K, V]) DeleteInverse(key V) {
	if existingValue, ok := b.inverse[key]; ok {
		delete(b.direct, existingValue)
		delete(b.inverse, key)
	}
}

// Keys returns all of the keys from the Bidirectional map as a slice.
func (b *BidirectionalMap[K, V]) Keys() []K {
	return slices.Collect(maps.Keys(b.direct))
}

// Values returns all of the values from the Bidirectional map as a slice.
func (b *BidirectionalMap[K, V]) Values() []V {
	return slices.Collect(maps.Keys(b.inverse))
}
