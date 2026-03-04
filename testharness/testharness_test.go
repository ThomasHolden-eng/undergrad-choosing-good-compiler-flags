package testharness

import (
	"6CCS3PRJ/thomasholden/util"
	"fmt"
	"os"
	"testing"
	"time"
)

func initEnv() func() {
	backupRunProgramUntilStable := runProgramUntilStable
	backupRunFile := runFile
	backupCompileFileWithPasses := compileFileWithPasses
	restore := func() {
		runProgramUntilStable = backupRunProgramUntilStable
		runFile = backupRunFile
		compileFileWithPasses = backupCompileFileWithPasses
	}

	runProgramUntilStable = func(_ *os.File, _ time.Duration) (float64, int64, error) {
		return 2.0, 1024, nil
	}
	runFile = func(_ *os.File, _ time.Duration) (time.Duration, int64, uint64, error) {
		return 2 * time.Microsecond, 1024, 1, nil
	}
	compileFileWithPasses = func(_ []string, _ string, _ int, _ bool) (time.Duration, int, *os.File, error) {
		tmpFile, err := os.CreateTemp("", "testharness_*.bin")
		if err != nil {
			return 0, 0, nil, err
		}
		return 100 * time.Millisecond, 32768, tmpFile, nil
	}

	return restore
}

func TestGetSetBaselineStats(t *testing.T) {
	tests := []struct {
		name                          string
		runFileOvveride               func()
		compileFileWithPassesOvveride func()
		errorExpected                 bool
		hashExpected                  uint64
		durationExpected              time.Duration
		memoryExpected                int64
	}{
		{
			name: "getExpectedHash error path",
			compileFileWithPassesOvveride: func() {
				compileFileWithPasses = func(_ []string, _ string, _ int, _ bool) (time.Duration, int, *os.File, error) {
					return 0, 0, nil, fmt.Errorf("error")
				}
			},
			errorExpected:    true,
			hashExpected:     0,
			durationExpected: -time.Second,
			memoryExpected:   0,
		},
		{
			name: "getExpectedExecutionStats error path",
			runFileOvveride: func() {
				runFile = func(_ *os.File, _ time.Duration) (time.Duration, int64, uint64, error) {
					return 0, 0, 0, fmt.Errorf("error")
				}
			},
			errorExpected:    true,
			hashExpected:     0,
			durationExpected: -time.Second,
			memoryExpected:   0,
		},
		{
			name:             "happy path",
			errorExpected:    false,
			hashExpected:     1,
			durationExpected: 2,
			memoryExpected:   1024,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			restore := initEnv()
			defer restore()

			if test.runFileOvveride != nil {
				test.runFileOvveride()
			}
			if test.compileFileWithPassesOvveride != nil {
				test.compileFileWithPassesOvveride()
			}
			d, m, err := GetSetBaselineStats([]string{"echo", "helloworld"}, "nil", func() {}, 1)
			if d != test.durationExpected || m != test.memoryExpected || (err != nil) != test.errorExpected {
				t.Errorf("Expected %v, %v, %v, but got %v, %v, %v",
					test.durationExpected,
					test.memoryExpected,
					test.errorExpected,
					d, m, err)
			}
			if trueHash != test.hashExpected {
				t.Errorf("Expected trueHash to be %v, got %v", test.hashExpected, trueHash)
			}
		})
	}
}

func TestCompileAndExecute(t *testing.T) {
	tests := []struct {
		name                          string
		runProgramUntilStableOvveride func()
		compileFileWithPassesOvveride func()
		errorExpected                 bool
		compilationTimeExpected       time.Duration
		executionTimeExpected         time.Duration
		binarySizeExpected            int
		memUsageExpected              int64
	}{
		{
			name: "compileFileWithPasses error path",
			compileFileWithPassesOvveride: func() {
				compileFileWithPasses = func(_ []string, _ string, _ int, _ bool) (time.Duration, int, *os.File, error) {
					return 0, 0, nil, fmt.Errorf("error")
				}
			},
			errorExpected:           true,
			compilationTimeExpected: util.MAX_DURATION,
			executionTimeExpected:   util.MAX_DURATION,
			binarySizeExpected:      0,
			memUsageExpected:        0,
		},
		{
			name: "runProgramUntilStable error path",
			runProgramUntilStableOvveride: func() {
				runProgramUntilStable = func(_ *os.File, _ time.Duration) (float64, int64, error) {
					return 0, 0, fmt.Errorf("error")
				}
			},
			errorExpected:           true,
			compilationTimeExpected: util.MAX_DURATION,
			executionTimeExpected:   util.MAX_DURATION,
			binarySizeExpected:      0,
			memUsageExpected:        0,
		},
		{
			name:                    "happy path",
			errorExpected:           false,
			compilationTimeExpected: 100 * time.Millisecond,
			executionTimeExpected:   2,
			binarySizeExpected:      32768,
			memUsageExpected:        1024,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			restore := initEnv()
			defer restore()

			if test.runProgramUntilStableOvveride != nil {
				test.runProgramUntilStableOvveride()
			}
			if test.compileFileWithPassesOvveride != nil {
				test.compileFileWithPassesOvveride()
			}

			compilationTime, executionTime, binarySize, memUsage, err := CompileAndExecute(
				"nil", 0, false, []string{"echo", "helloworld"}, 1*time.Second,
			)

			if compilationTime != test.compilationTimeExpected ||
				executionTime != test.executionTimeExpected ||
				binarySize != test.binarySizeExpected ||
				memUsage != test.memUsageExpected {
				t.Errorf("Expected %v, %v, %v, %v, but got %v, %v, %v, %v",
					test.compilationTimeExpected,
					test.executionTimeExpected,
					test.binarySizeExpected,
					test.memUsageExpected,
					compilationTime, executionTime, binarySize, memUsage)
			}
			if (err != nil) != test.errorExpected {
				t.Errorf("Expected an error, but got none")
			}
		})
	}
}
