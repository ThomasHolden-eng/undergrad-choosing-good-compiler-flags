package testharness

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type Harness struct {
	flagIntMap BidirectionalMap[string, int]
}

func extractUserTime(timeOut string) (float32, error) {
	if timeOut == "" {
		return 0.0, fmt.Errorf("received no input")
	}

	r, _ := regexp.Compile("user ([0-9]+.[0-9]+)")
	userTimeLine := r.FindStringSubmatch(timeOut)[1]

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
