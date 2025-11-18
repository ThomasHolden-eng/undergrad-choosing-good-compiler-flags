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

func mapInputFilesToBC(cmd []string, bcFiles []string) {
	filePtr := 0
	for i, s := range cmd {
		if cFileRegex.MatchString(s) {
			cmd[i] = bcFiles[filePtr]
			filePtr++
		}
	}
}

func cleanFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

func ArrayFromChannel[T any](channel chan T) []T {
	sliceT := make([]T, len(channel))
	for elem := range channel {
		sliceT = append(sliceT, elem)
	}
	return sliceT
}
