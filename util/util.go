package util

import "slices"

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 |
		uint32 | uint64 | uintptr | float32 | float64
}

func ArrayFromChannel[T any](channel chan T) []T {
	sliceT := make([]T, 0, len(channel))
	for elem := range channel {
		sliceT = append(sliceT, elem)
	}
	return sliceT
}

func Median[T Number](list []T) *T {
	slices.Sort(list)

	length := len(list)
	if length == 0 {
		return nil
	} else if length%2 == 0 {
		val := (list[length/2] + list[-1+length/2]) / 2
		return &val
	} else {
		return &list[length/2]
	}
}

func ForEach[T, S any](slice []T, f func(T) S) []S {
	newSlice := make([]S, len(slice))

	for i, v := range slice {
		newSlice[i] = f(v)
	}

	return newSlice
}
