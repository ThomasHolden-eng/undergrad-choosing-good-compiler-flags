package testharness

import (
	"fmt"
	"os"
	"regexp"
	"slices"
)

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
