package testharness

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"syscall"
	"time"
)

// Extracts C-family filenames from a string.
var cFileRegex *regexp.Regexp = regexp.MustCompile(`\w+?\.c(?:(?:pp)|(?:xx)|(?:c))?`)
var trueHash uint64

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

// Takes a file reference
// Returns execution time, memory usage, hash & error
func runFile(file *os.File, maxDuration time.Duration) (time.Duration, int64, uint64, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	ctx := context.Background()
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	file.Chmod(0755)
	cmd := exec.CommandContext(ctx, file.Name())
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return 0, 0, 0, fmt.Errorf("execution failed: %w\nStderr: %s", err, stderr.String())
	}

	duration := cmd.ProcessState.UserTime().Microseconds()
	rawOut := append(stdout.Bytes(), stderr.Bytes()...)
	outHasher := fnv.New64a()
	outHasher.Write(rawOut)
	fnvHash := outHasher.Sum64()

	// Will fail within windows
	memUsage := func() int64 {
		if usage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
			return usage.Maxrss
		}
		return 0
	}()

	totalTime := time.Duration(duration * int64(time.Microsecond))
	return totalTime, memUsage, fnvHash, nil
}

func setTrueHash(hash uint64) {
	trueHash = hash
}

func getExpectedHash(compilationCommand []string, passes string) (uint64, error) {
	_, _, f, err := CompileFileWithPasses(compilationCommand, passes, 0, false)
	defer os.Remove(f.Name())
	defer f.Close() // defer is LIFO

	if err != nil {
		return 0, fmt.Errorf("compilation failure: %w", err)
	}

	_, _, hash, err := runFile(f, time.Hour*24)
	if err != nil {
		return 0, fmt.Errorf("execution failure: %w", err)
	}

	return hash, nil
}

// getExpectedExecutionTime returns only the median execution time of a pass sequence.
func getExpectedExecutionTime(compilationCommand []string, passes string) (time.Duration, error) {
	_, _, f, err := CompileFileWithPasses(compilationCommand, passes, 0, false)
	defer os.Remove(f.Name())
	defer f.Close() // defer is LIFO

	if err != nil {
		return -time.Second, fmt.Errorf("compilation failure: %w", err)
	}

	timeTaken, _, err := RunProgramUntilStable(f, time.Hour*24)
	if err != nil {
		return -time.Second, fmt.Errorf("execution failure: %w", err)
	}

	return time.Duration(timeTaken), nil
}
