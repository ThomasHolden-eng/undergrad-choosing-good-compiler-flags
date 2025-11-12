package testharness

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Harness struct {
	flagIntMap BidirectionalMap[string, int]
}

func extractUserTime(timeOut string) (float32, error) {
	userTimeLine := strings.Split(timeOut, "\n")[1]
	userTimeLine = strings.ReplaceAll(userTimeLine, " ", "")
	userTimeLine = strings.TrimPrefix(userTimeLine, "user")

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
