package testharness

import "os"

// Takes a file reference and flag list
// Returns compilation time, execution time, memory usage & binary size
func RunFileWithFlags(file *os.File, flags []int) (int, int, int, int) {
	return 0, 0, 0, 0
}

func FlagToInt(s string) int {
	return 0
}

func IntToFlag(n int) string {
	return ""
}
