package testharness

import (
	"fmt"
	"os"
	"regexp"
	"slices"
)

type Harness struct {
	flagIntMap BidirectionalMap[string, int]
}

var cFileRegex *regexp.Regexp = regexp.MustCompile(`\w+?\.c(?:(?:pp)|(?:xx)|(?:c))?`)

func getFileSize(file *os.File) (int, error) {
	stat, err := file.Stat()
	if err != nil {
		return -1, fmt.Errorf("failed to stat file '%s': %w", file.Name(), err)
	}
	return int(stat.Size()), nil
}

func extractInputFiles(cmd []string) []string {
	if len(cmd) == 0 {
		return nil
	}

	temporarySlice := slices.Clone(cmd)

	returnS := slices.DeleteFunc(temporarySlice, func(s string) bool {
		return !cFileRegex.MatchString(s)
	})

	return returnS
}

func extractFlags(cmd []string) []string {
	if len(cmd) == 0 {
		return nil
	}

	temporarySlice := slices.Clone(cmd[1:])

	returnS := slices.DeleteFunc(temporarySlice, func(s string) bool {
		return cFileRegex.MatchString(s)
	})

	return returnS
}

func mapInputFilesToBC(cmd []string, bcFiles []string) []string {
	filePtr := 0
	temporarySlice := slices.Clone(cmd)

	for i, s := range temporarySlice {
		if cFileRegex.MatchString(s) {
			temporarySlice[i] = bcFiles[filePtr]
			filePtr++
		}
	}

	return temporarySlice
}

func cleanFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

func ArrayFromChannel[T any](channel chan T) []T {
	sliceT := make([]T, 0, len(channel))
	for elem := range channel {
		sliceT = append(sliceT, elem)
	}
	return sliceT
}

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 |
		uint32 | uint64 | uintptr | float32 | float64
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
