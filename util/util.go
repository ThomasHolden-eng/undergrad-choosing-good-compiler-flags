package util

import (
	"slices"
	"time"
)

// Number is an interface that allows generic usage with any float or int.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 |
		~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

// Range stores a pair of numbers, and Within(n) returns true if start <= n <= end.
type Range[T Number] struct {
	start T
	end   T
}

const (
	MAX_DURATION time.Duration = time.Hour * 292     // The maximum possible duration value.
	MAX_INT64    int64         = 9223372036854775807 // The maximum possible int64 value.
	MAX_FLOAT32  float32       = 3.402823e+38        // The maximum possible float32 value.
)

// ArrayFromChannel takes a channel and returns an array containing all the elements queued.
func ArrayFromChannel[T any](channel chan T) []T {
	sliceT := make([]T, 0, len(channel))
	for elem := range channel {
		sliceT = append(sliceT, elem)
	}

	return sliceT
}

// Median returns the median of any slice of ints or floats.
func Median[T Number](list []T) *T {
	slices.Sort(list)

	length := len(list)
	if length == 0 {
		return nil
	} else if length%2 == 0 {
		val := list[length/2] + list[-1+length/2]/2
		return &val
	} else {
		val := list[length/2]
		return &val
	}
}

// ForEach applies a function to every element of a slice.
func ForEach[T, S any](slice []T, f func(T) S) []S {
	newSlice := make([]S, len(slice))

	for i, v := range slice {
		newSlice[i] = f(v)
	}

	return newSlice
}

// NewRange returns an initialised range from numbers.
func NewRange[T Number](a, b T) *Range[T] {
	newRange := Range[T]{a, b}

	return &newRange
}

// Within tests whether a value falls between the two values of a range.
func (r *Range[T]) Within(n T) bool {
	return r.start <= n && n <= r.end
}
