package util

import (
	"math"
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
	MAX_FLOAT64  float64       = 1.7e+308            // The maximum possible float64 value.
)

// ArrayFromChannel takes a channel and returns an array containing all the elements queued.
func ArrayFromChannel[T any](channel chan T) []T {
	sliceT := make([]T, 0, len(channel))
	for elem := range channel {
		sliceT = append(sliceT, elem)
	}

	return sliceT
}

// linearInterpolate returns an interpolated value across the list input.
func linearInterpolate[T Number](list []T, t float64) float64 {
	slices.Sort(list)

	length := len(list) - 1
	lower := int(math.Floor(float64(length) * t))
	upper := int(math.Ceil(float64(length) * t))
	if len(list) == 0 {
		return 0.0
	} else if lower == upper {
		return float64(list[lower])
	} else {
		fraction := (float64(length) * t) - math.Floor(float64(length)*t)
		return float64(list[lower]) + float64(list[upper]-list[lower])*fraction
	}
}

// Median returns the median of any slice of ints or floats.
//
// Sorts as a side effect.
func Median[T Number](list []T) float64 {
	if len(list) == 0 {
		return MAX_FLOAT64
	}
	val := linearInterpolate(list, 0.5)
	return val
}

// IQR returns the IQR of any slice of ints or floats.
//
// Sorts as a side effect.
func IQR[T Number](list []T) float64 {
	if len(list) == 0 {
		return MAX_FLOAT64
	}
	val := linearInterpolate(list, 0.75) - linearInterpolate(list, 0.25)
	return val
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

// stressCoreOnChannel stresses a thread and exits when the exit channel is written to.
//
// Takes a channel on who's closure it will kill the thread, and memory usage in KB.
func StressOnChannel(exit chan byte, memoryExpected int64) {
	memoryExpected = max(1, memoryExpected)
	// *500/707 approximates divsion by sqrt(2) to get a reasonable estimate for avg mem usage
	size := memoryExpected * 1024 / 8 * 500 / 707 // *1024 to get to bytes, /8 for int64 size
	buffer := make([]uint64, size)                // Uses around 128*8=1024 bytes
	for ptr := int64(1); true; ptr++ {
		select {
		case <-exit:
			return
		default:
			buffer[ptr%size] ^= buffer[(ptr-1)%size]*31 + 99
		}
	}
}
