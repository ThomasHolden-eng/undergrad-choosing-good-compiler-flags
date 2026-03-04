package testharness

import (
	"6CCS3PRJ/thomasholden/util"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"syscall"
	"time"

	"github.com/zeebo/xxh3"
)

// Extracts C-family filenames from a string.
var cFileRegex *regexp.Regexp = regexp.MustCompile(`\w+?\.c(?:(?:pp)|(?:xx)|(?:c))?`)

// Used to validate program execution output.
var trueHash uint64

// getFileSize returns the size of a file in bytes.
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
//
// Defined as a variable function for testing purposes.
var runFile = func(file *os.File, maxDuration time.Duration) (time.Duration, int64, uint64, error) {
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

	var hash uint64 = calculateHash(
		stderr.Bytes(), stdout.Bytes(),
	)

	// Will fail within windows
	memUsage := func() int64 {
		if usage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
			return usage.Maxrss
		}
		return util.MAX_INT64
	}()

	totalTime := time.Duration(duration * int64(time.Microsecond))
	return totalTime, memUsage, hash, nil
}

// calculateHash returns a uint64 hash for any number of byte streams.
func calculateHash(streams ...[]byte) uint64 {
	hasher := xxh3.New()
	for _, stream := range streams {
		hasher.Write(stream)
	}

	return hasher.Sum64()
}

// setTrueHash assigns hash to package variable trueHash.
// Used to calculate whether a programs output has been corrupted.
func setTrueHash(hash uint64) {
	trueHash = hash
}

// getExpectedHash returns the correct hash for the output of the program.
//
// As a byproduct also returns memory usage.
func getExpectedHash(compilationCommand []string, passes string) (uint64, int64, error) {
	_, _, f, err := compileFileWithPasses(compilationCommand, passes, 0, false)
	if err != nil {
		return 0, 0, fmt.Errorf("compilation failure: %w", err)
	}

	defer os.Remove(f.Name())
	defer f.Close() // defer is LIFO

	_, mem, hash, err := runFile(f, time.Hour*24)
	if err != nil {
		return 0, 0, fmt.Errorf("execution failure: %w", err)
	}

	return hash, mem, nil
}

// getExpectedExecutionTime returns only the median execution time of a pass sequence.
func getExpectedExecutionStats(compilationCommand []string, passes string) (time.Duration, error) {
	_, _, f, err := compileFileWithPasses(compilationCommand, passes, 0, false)
	defer os.Remove(f.Name())
	defer f.Close() // defer is LIFO

	if err != nil {
		return -time.Second, fmt.Errorf("compilation failure: %w", err)
	}

	timeTaken, _, err := runProgramUntilStable(f, time.Hour*24)
	if err != nil {
		return -time.Second, fmt.Errorf("execution failure: %w", err)
	}

	return time.Duration(timeTaken), nil
}

// compileFileWithPasses takes a file reference, pass list & unique id.
//
// Returns compilation time, binary size, file reference & error.
//
// Caller must resolve returned executable reference (close and delete).
// Defined as a variable function for testing purposes.
var compileFileWithPasses = func(compilationCommand []string, passes string, id int, outLocal bool) (time.Duration, int, *os.File, error) {
	if len(compilationCommand) == 0 {
		return 0, 0, nil, fmt.Errorf("empty command passed")
	}

	binFile, err := os.CreateTemp("", fmt.Sprintf("bin-%d-*", id))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	binName := binFile.Name()
	binFile.Close()

	// Extract input files as a list of strings
	filesToOptimise := extractInputFiles(compilationCommand)

	var intermediaryFiles []string
	for range filesToOptimise {
		bcFile, err := os.CreateTemp("", fmt.Sprintf("opt-%d-*.bc", id))
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to create temporary file: %w", err)
		}
		intermediaryFiles = append(intermediaryFiles, bcFile.Name())
		defer os.Remove(bcFile.Name()) // Deletes IR when we finish compiling
		bcFile.Close()
	}

	var stderr bytes.Buffer
	// Extract flags as a list of strings
	flags := extractFlags(compilationCommand)
	for i := range filesToOptimise {
		stderr.Reset()
		inpFile, outFile := filesToOptimise[i], intermediaryFiles[i]

		// here we add -O3, despite disabling running passes in this step
		// this provided essential target library info to the IR, allowing us to optimise
		// without being at a disadvantage to a simple clang -O3 command
		command := []string{
			compilationCommand[0], inpFile, "-O3", "-emit-llvm", "-Xclang", "-disable-llvm-passes",
			"-Xclang", "-disable-llvm-optzns", "-Xclang", "-disable-O0-optnone", "-c", // https://arxiv.org/pdf/2109.08267
		}
		command = append(command, flags...)
		command = append(command, "-o", outFile)

		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			os.Remove(binName)
			return 0, 0, nil, fmt.Errorf("compilation failed: %w\nCommand: %v\nStderr: %s", err, command, stderr.String())
		}
	}

	var totalDuration int64
	for i := range intermediaryFiles {
		stderr.Reset()
		inpFile, outFile := intermediaryFiles[i], intermediaryFiles[i]

		command := []string{"opt", "-passes=" + passes, inpFile, "-o", outFile}

		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			os.Remove(binName)
			return 0, 0, nil, fmt.Errorf("optimisation failed: %w\nCommand: %v\nStderr: %s", err, command, stderr.String())
		}

		duration := cmd.ProcessState.UserTime().Microseconds()

		totalDuration += duration
	}

	newCommand := mapInputFilesToBC(compilationCommand, intermediaryFiles)

	newCommand = append(newCommand, "-O3", "-Xclang", "-disable-llvm-passes",
		"-Xclang", "-disable-llvm-optzns") // applies O3 optimisations without running passes

	if outLocal {
		localCmd := exec.Command(newCommand[0], newCommand[1:]...)
		localCmd.Stderr = &stderr
		if err := localCmd.Run(); err != nil {
			return 0, 0, nil, fmt.Errorf("final linking failed: %w\nStderr: %s", err, stderr.String())
		}
	}
	newCommand = append(newCommand, "-o", binName)

	cmd := exec.Command(newCommand[0], newCommand[1:]...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.Remove(binName)
		return 0, 0, nil, fmt.Errorf("linking failed: %w\nStderr: %s", err, stderr.String())
	}

	binFile, err = os.Open(binName)
	if err != nil {
		os.Remove(binName)
		return 0, 0, nil, fmt.Errorf("failed to open return binary: %w", err)
	}

	binSize, err := getFileSize(binFile)
	if err != nil {
		binFile.Close()
		os.Remove(binName)
		return 0, 0, nil, fmt.Errorf("failed to size binary '%s': %w", binName, err)
	}

	totalTime := time.Duration(totalDuration * int64(time.Microsecond))
	return totalTime, binSize, binFile, nil
}

// runProgramUntilStable returns the median execution time and maxrss of a file execution.
//
// Defined as a variable function for testing purposes.
var runProgramUntilStable = func(toRun *os.File, approxTime time.Duration) (float64, int64, error) {
	var times []int64
	var rsss []int64

	i := 0
	for begin := time.Now(); (i < 5 || // run at least 5 times
		!MedianConfidence(times, 0.03)) && // median is within 3%
		time.Since(begin) < 20*approxTime; { // run for at most 20*o3 time

		execution_t, rss, hash, err := runFile(toRun, approxTime*12)
		if err != nil {
			return util.MAX_FLOAT64, util.MAX_INT64, fmt.Errorf("execution failed: %v", err)
		} else if hash != trueHash {
			return util.MAX_FLOAT64, util.MAX_INT64, fmt.Errorf(
				"execution out hash: %v\ntrue out hash: %v", hash, trueHash,
			)
		}

		times = append(times, int64(execution_t))
		rsss = append(rsss, rss)
		i++
	}
	return util.Median(times), slices.Max(rsss), nil
}
