package testharness

import (
	"fmt"
	"os"
	"regexp"
	"slices"
)

// Extracts C-family filenames from a string.
var cFileRegex *regexp.Regexp = regexp.MustCompile(`\w+?\.c(?:(?:pp)|(?:xx)|(?:c))?`)

// getFileSize returns the size of a file.
func getFileSize(file *os.File) (int, error) {
	stat, err := file.Stat()
	if err != nil {
		return -1, fmt.Errorf("failed to stat file '%s': %w", file.Name(), err)
	}
	return int(stat.Size()), nil
}

// extractInputFiles takes a string slice parsed by schlex and returns only the
// parts determined to be filenames for C-family files.
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

// extractFlags extracts all the portions of a build command that are
// determind not to be filenames for C-family files.
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

// mapInputFilesToBC replaces filenames for C-family files with filenames
// of their LLVM bytecode counterparts.
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
