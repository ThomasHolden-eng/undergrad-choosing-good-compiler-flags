package testharness

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
)

type Harness struct {
	flagIntMap BidirectionalMap[string, int]
}

var cFileRegex *regexp.Regexp = regexp.MustCompile(`\w+?\.c(?:(?:pp)|(?:xx)|(?:c))?`)
var userFloatRegex *regexp.Regexp = regexp.MustCompile(`user ([0-9]+\.[0-9]+)`)

func extractUserTime(timeOut string) (float32, error) {
	if timeOut == "" {
		return 0.0, fmt.Errorf("received no input")
	}

	matches := userFloatRegex.FindStringSubmatch(timeOut)

	if len(matches) < 2 {
		return 0.0, fmt.Errorf("no user time found in input")
	}

	userTimeLine := matches[1]

	float, err := strconv.ParseFloat(userTimeLine, 32)
	return float32(float), err
}

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
