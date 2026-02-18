package testharness

import (
	"6CCS3PRJ/thomasholden/util"
	"fmt"
	"os"
	"time"
)

// GetSetBaselineStats returns the execution time and memory usage for the baseline while setting the
// ground truth hash value.
func GetSetBaselineStats(compilationCommand []string, passes string, noiseFunc func(), workers int) (time.Duration, int64, error) {
	trueHash, mem, err := getExpectedHash(compilationCommand, passes)
	if err != nil {
		return -time.Second, 0, fmt.Errorf("failed basic execution, is the command formatted correctly?:\n%v", err.Error())
	}
	setTrueHash(trueHash)
	fmt.Print("Successfully compiled -O3 ")

	for range workers - 1 {
		go noiseFunc()
	}
	o3Time, err := getExpectedExecutionStats(compilationCommand, passes)
	if err != nil {
		return -time.Second, 0, fmt.Errorf("failed basic execution, is the command formatted correctly?:\n%v", err.Error())
	}
	fmt.Println(": Successfully executed -O3, waiting on noise.")

	return o3Time, mem, nil
}

// CompileAndExecute manages interfacing with the test harness to compile and get a stable
// time for execution of a pass sequence.
//
// Returns the stats of the sequence.
func CompileAndExecute(passes string, diff int, outLocal bool, SourceCmd []string, approxTime time.Duration) (
	time.Duration, time.Duration, int, int64, error) {
	compilation_t, binary_size, newFile, err := compileFileWithPasses(
		SourceCmd,
		passes,
		diff, outLocal,
	)
	if err != nil {
		return util.MAX_DURATION, util.MAX_DURATION, 0, 0, err
	}
	defer os.Remove(newFile.Name())

	execution_t, maxrss, err := runProgramUntilStable(newFile, approxTime)
	if err != nil {
		return util.MAX_DURATION, util.MAX_DURATION, 0, 0, err
	}
	newFile.Close()

	return compilation_t, time.Duration(execution_t), binary_size, maxrss, nil
}
