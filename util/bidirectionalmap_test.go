package util

import (
	"slices"
	"testing"
)

var m BidirectionalMap[int, int] // m is a bidirectional map for testing purposes

// bidirectionalMapTeardown is used to clean up the bidirectional map post testing.
func bidirectionalMapTeardown() {
	for _, key := range m.Keys() {
		m.Delete(key)
	}
}

// TestPut tests that the Put function correctly handles valid and invalid
// key-value pairs.
func TestPut(t *testing.T) {
	defer bidirectionalMapTeardown()

	err := m.Put(0, 1)
	if err != nil {
		t.Fatalf("Errored on attempted put %v", err)
	}
	err = m.Put(0, 3) // Should error due to key collision
	if err == nil {
		t.Fatalf("Didn't error on invalid put")
	}
}

// TestPutInverse tests that the PutInverse function correctly handles valid and
// invalid key-value pairs.
func TestPutInverse(t *testing.T) {
	defer bidirectionalMapTeardown()

	err := m.PutInverse(0, 1)
	if err != nil {
		t.Fatalf("Errored on attempted inverse put %v", err)
	}
	err = m.PutInverse(0, 3) // Should error due to key collision
	if err == nil {
		t.Fatalf("Didn't error on invalid inverse put")
	}
}

// TestGet verifies that the Get function, returns nil for missing keys and
// the correct value for existing keys.
func TestBDMapGet(t *testing.T) {
	defer bidirectionalMapTeardown()

	val, ok := m.Get(0)
	if ok {
		t.Fatalf("Got %v, but expected nil", val)
	}

	err := m.Put(0, 1)
	if err != nil {
		t.Fatalf("Errored on attempted put %v", err)
	}

	val, ok = m.Get(0)
	if !ok || val != 1 {
		t.Fatalf("Got %v, but expected 1 | ok=%v", val, ok)
	}
}

// TestGet verifies that the GetInverse function, returns nil for missing values
// and the correct key for existing values.
func TestGetInverse(t *testing.T) {
	defer bidirectionalMapTeardown()

	val, ok := m.GetInverse(1)
	if ok {
		t.Fatalf("Got %v, but expected nil", val)
	}

	err := m.Put(0, 1)
	if err != nil {
		t.Fatalf("Errored on attempted put %v", err)
	}

	val, ok = m.GetInverse(1)
	if !ok || val != 0 {
		t.Fatalf("Got %v, but expected 0 | ok=%v", val, ok)
	}
}

// TestDelete tests that the Delete function correctly deletes key-value and
// value-key pairs.
func TestDelete(t *testing.T) {
	defer bidirectionalMapTeardown()

	err := m.Put(0, 1)
	if err != nil {
		t.Fatalf("Errored on attempted put %v", err)
	}

	val, ok := m.Get(0)
	if !ok || val != 1 {
		t.Fatalf("Got %v, but expected 1 | ok=%v", val, ok)
	}

	m.Delete(0)

	val, ok = m.Get(0)
	if ok {
		t.Fatalf("Got %v, but expected nil", val)
	}
}

// TestDeleteInverse tests that the DeleteInverse function correctly deletes
// value-key and key-value pairs.
func TestDeleteInverse(t *testing.T) {
	defer bidirectionalMapTeardown()

	err := m.Put(0, 1)
	if err != nil {
		t.Fatalf("Errored on attempted put %v", err)
	}

	val, ok := m.Get(0)
	if !ok || val != 1 {
		t.Fatalf("Got %v, but expected 1", val)
	}

	m.DeleteInverse(1)

	val, ok = m.Get(0)
	if ok {
		t.Fatalf("Got %v, but expected nil", val)
	}
}

// TestKeys tests that the Keys function returns the expected keys.
func TestKeys(t *testing.T) {
	defer bidirectionalMapTeardown()

	keys := []int{
		0, 1, 2, 3, 4, 5,
	}

	values := []int{
		0, 2, 4, 6, 8, 10,
	}

	for i := range keys {
		err := m.Put(keys[i], values[i])
		if err != nil {
			t.Fatalf("Errored on attempted put %v", err)
		}
	}

	mapKeys := m.Keys()
	slices.Sort(mapKeys)
	if !slices.Equal(mapKeys, keys) {
		t.Fatalf("Expected %v, got %v", keys, m.Keys())
	}
}

// TestValues tests that the Values function returns the expected values.
func TestValues(t *testing.T) {
	defer bidirectionalMapTeardown()

	keys := []int{
		0, 1, 2, 3, 4, 5,
	}

	values := []int{
		0, 2, 4, 6, 8, 10,
	}

	for i := range keys {
		err := m.Put(keys[i], values[i])
		if err != nil {
			t.Fatalf("Errored on attempted put %v", err)
		}
	}

	mapValues := m.Values()
	slices.Sort(mapValues)
	if !slices.Equal(mapValues, values) {
		t.Fatalf("Expected %v, got %v", values, m.Values())
	}
}
