package util

import (
	"slices"
	"time"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 |
		~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

// Range stores a pair of numbers, and Within(n) returns true if start <= n <= end
type Range[T Number] struct {
	start T
	end   T
}

const (
	MAX_DURATION time.Duration = time.Hour * 292
	MAX_INT64    int64         = 9223372036854775807
	MAX_FLOAT32  float32       = 3.402823e+38
)

func ArrayFromChannel[T any](channel chan T) []T {
	sliceT := make([]T, 0, len(channel))
	for elem := range channel {
		sliceT = append(sliceT, elem)
	}

	return sliceT
}

func Median[T Number](list []T) *float64 {
	slices.Sort(list)

	length := len(list)
	if length == 0 {
		return nil
	} else if length%2 == 0 {
		val := float64(list[length/2]+list[-1+length/2]) / 2
		return &val
	} else {
		val := float64(list[length/2])
		return &val
	}
}

func ForEach[T, S any](slice []T, f func(T) S) []S {
	newSlice := make([]S, len(slice))

	for i, v := range slice {
		newSlice[i] = f(v)
	}

	return newSlice
}

func NewRange[T Number](a, b T) *Range[T] {
	newRange := Range[T]{a, b}

	return &newRange
}

func (r *Range[T]) Within(n T) bool {
	return r.start <= n && n <= r.end
}
