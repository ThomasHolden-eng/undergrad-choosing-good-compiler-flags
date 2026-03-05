package testharness

import (
	"os"
	"testing"
	"time"

	"github.com/google/shlex"
)

// compilePassFunc is defined to prevent repeat definitions for overrides.
type compilePassFunc func([]string, string, int, bool) (time.Duration, int, *os.File, error)

func TestGetFileSize(t *testing.T) {
	zeroFile, err := os.CreateTemp("", "testharness_*.bin")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "testharness_*.bin")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	bytesToWrite := make([]byte, 1024) // 1KB of 0 data
	tmpFile.Write(bytesToWrite)

	deletedFile, _ := os.CreateTemp("", "testharness_*.bin")
	deletedFile.Close()
	os.Remove(deletedFile.Name())

	tests := []struct {
		name          string
		file          *os.File
		sizeExpected  int
		errorExpected bool
	}{
		{
			name:          "valid file of 0 size",
			file:          zeroFile,
			sizeExpected:  0,
			errorExpected: false,
		},
		{
			name:          "valid file of non-zero size",
			file:          tmpFile,
			sizeExpected:  1024,
			errorExpected: false,
		},
		{
			name:          "error when file does not exist",
			file:          deletedFile,
			errorExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size, err := getFileSize(test.file)
			if err == nil && test.errorExpected {
				t.Errorf("expected to fail getting file size")
			} else if err != nil && !test.errorExpected {
				t.Errorf("failed to get file size unexpectedly: %v", err)
			} else if err == nil && size != test.sizeExpected {
				t.Errorf("got file size of %d, but expected %d", size, test.sizeExpected)
			}
		})
	}
}

func TestExtractInputFiles(t *testing.T) {
	validCmd, err := shlex.Split(
		"clang -I./example/PolyBenchC-4.2.1/utilities -I./example/PolyBenchC-4.2.1/datamining/" +
			"correlation ./example/PolyBenchC-4.2.1/utilities/polybench.c ./example/PolyBenchC" +
			"-4.2.1/datamining/correlation/correlation.c",
	)
	if err != nil {
		t.Fatalf("failed to split command: %v", err)
	}

	tests := []struct {
		name    string
		command []string
		files   []string
	}{
		{
			name:    "no files in command",
			command: []string{"echo", "helloworld"},
			files:   []string{},
		},
		{
			name:    "empty command",
			command: []string{},
			files:   []string{},
		},
		{
			name:    "only files in command",
			command: []string{"helloworld.c", "hello_world.cpp", "helloworld.cc"},
			files:   []string{"helloworld.c", "hello_world.cpp", "helloworld.cc"},
		},
		{
			name:    "valid example command",
			command: validCmd,
			files: []string{
				"./example/PolyBenchC-4.2.1/utilities/polybench.c",
				"./example/PolyBenchC-4.2.1/datamining/correlation/correlation.c",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files := extractInputFiles(test.command)
			if len(files) != len(test.files) {
				t.Errorf("got %d files, but expected %d", len(files), len(test.files))
			}
			for i, file := range files {
				if file != test.files[i] {
					t.Errorf("got file %s, but expected %s", file, test.files[i])
				}
			}
		})
	}
}

func TestExtractFlags(t *testing.T) {
	validCmd, err := shlex.Split(
		"clang -I./example/PolyBenchC-4.2.1/utilities -I./example/PolyBenchC-4.2.1/datamining/" +
			"correlation ./example/PolyBenchC-4.2.1/utilities/polybench.c ./example/PolyBenchC" +
			"-4.2.1/datamining/correlation/correlation.c",
	)
	if err != nil {
		t.Fatalf("failed to split command: %v", err)
	}

	tests := []struct {
		name    string
		command []string
		flags   []string
	}{
		{
			name:    "no files in command",
			command: []string{"run", "echo", "helloworld"},
			flags:   []string{"echo", "helloworld"},
		},
		{
			name:    "empty command",
			command: []string{},
			flags:   []string{},
		},
		{
			name:    "only files in command",
			command: []string{"run", "helloworld.c", "hello_world.cpp", "helloworld.cc"},
			flags:   []string{},
		},
		{
			name:    "valid example command",
			command: validCmd,
			flags: []string{
				"-I./example/PolyBenchC-4.2.1/utilities",
				"-I./example/PolyBenchC-4.2.1/datamining/correlation",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags := extractFlags(test.command)
			if len(flags) != len(test.flags) {
				t.Fatalf("got %d flag(s), but expected %d", len(flags), len(test.flags))
			}
			for i, flag := range flags {
				if flag != test.flags[i] {
					t.Errorf("got file %s, but expected %s", flag, test.flags[i])
				}
			}
		})
	}
}

func TestMapInputFilesToBC(t *testing.T) {
	validCmd, err := shlex.Split(
		"clang -I./example/PolyBenchC-4.2.1/utilities -I./example/PolyBenchC-4.2.1/datamining/" +
			"correlation ./example/PolyBenchC-4.2.1/utilities/polybench.c ./example/PolyBenchC" +
			"-4.2.1/datamining/correlation/correlation.c",
	)
	if err != nil {
		t.Fatalf("failed to split command: %v", err)
	}

	tests := []struct {
		name    string
		command []string
		files   []string
		outCmd  []string
	}{
		{
			name:    "no files in command",
			command: []string{"echo", "helloworld"},
			outCmd:  []string{"echo", "helloworld"},
		},
		{
			name:    "empty command",
			command: []string{},
			files:   []string{},
			outCmd:  []string{},
		},
		{
			name:    "only files in command",
			command: []string{"helloworld.c", "hello_world.cpp", "helloworld.cc"},
			files:   []string{"helloworld1.bc", "hello_world2.bc", "helloworld3.bc"},
			outCmd:  []string{"helloworld1.bc", "hello_world2.bc", "helloworld3.bc"},
		},
		{
			name:    "valid example command",
			command: validCmd,
			files: []string{
				"./example/PolyBenchC-4.2.1/utilities/polybench.bc",
				"./example/PolyBenchC-4.2.1/datamining/correlation/correlation.bc",
			},
			outCmd: []string{"clang", "-I./example/PolyBenchC-4.2.1/utilities", "-I./example/PolyBenchC-4.2.1/datamining/" +
				"correlation", "./example/PolyBenchC-4.2.1/utilities/polybench.bc", "./example/PolyBenchC" +
				"-4.2.1/datamining/correlation/correlation.bc"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outCmd := mapInputFilesToBC(test.command, test.files)
			if len(outCmd) != len(test.outCmd) {
				t.Fatalf("got command of length %d, but expected %d", len(outCmd), len(test.outCmd))
			}
			for i, param := range outCmd {
				if param != test.outCmd[i] {
					t.Errorf("got command %v, but expected %v", outCmd, test.outCmd)
				}
			}
		})
	}
}

func TestRunFile(t *testing.T) {
	deletedFile, err := os.CreateTemp("", "testharness_*.bin")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	os.Remove(deletedFile.Name())

	validFile, err := createShellScriptWithOutput("success")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(validFile.Name())

	tests := []struct {
		name          string
		file          *os.File
		errorExpected bool
	}{
		{
			name:          "invalid file",
			file:          deletedFile,
			errorExpected: true,
		},
		{
			name:          "valid file",
			file:          validFile,
			errorExpected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			duration, memUsage, _, err := runFile(test.file, 2*time.Second)

			if test.errorExpected && err == nil {
				t.Errorf("expected an error, but got nil")
			} else if !test.errorExpected && err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			} else if duration < 0 {
				t.Errorf("expected duration to be non-negative, got %v", duration)
			} else if memUsage < 0 {
				t.Errorf("expected memory usage to be non-negative, got %d", memUsage)
			}
		})
	}
}

func TestCalculateHash(t *testing.T) {
	stream1 := []byte("hello world")
	stream2 := []byte("")
	stream3 := []byte("helloworld")

	tests := []struct {
		name   string
		first  []byte
		second []byte
		match  bool
	}{
		{
			name:   "same stream",
			first:  stream1,
			second: stream1,
			match:  true,
		},
		{
			name:   "two empty streams",
			first:  stream2,
			second: stream2,
			match:  true,
		},
		{
			name:   "subtly different streams",
			first:  stream1,
			second: stream3,
			match:  false,
		},
		{
			name:   "completely different streams",
			first:  stream1,
			second: stream2,
			match:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hash1 := calculateHash(test.first)
			hash2 := calculateHash(test.second)
			if test.match != (hash1 == hash2) {
				t.Errorf("expected hashes to match, but got %v and %v", hash1, hash2)
			}
		})
	}
}

func TestSetTrueHash(t *testing.T) {
	before := trueHash
	setTrueHash(before + 1)
	if trueHash != before+1 {
		t.Errorf("expected trueHash to be %d, but got %d", before+1, trueHash)
	}
}

func TestGetExpectedHash(t *testing.T) {
	tests := []struct {
		name          string
		first         compilePassFunc
		second        compilePassFunc
		match         bool
		errorExpected bool
	}{
		{
			name:          "same output",
			first:         buildCompileFileWithPassesOut("hello world", nil, false),
			second:        buildCompileFileWithPassesOut("hello world", nil, false),
			match:         true,
			errorExpected: false,
		},
		{
			name:          "two empty outputs",
			first:         buildCompileFileWithPassesOut("", nil, false),
			second:        buildCompileFileWithPassesOut("", nil, false),
			match:         true,
			errorExpected: false,
		},
		{
			name:          "compilation error path",
			first:         buildCompileFileWithPassesOut("hello world", os.ErrClosed, false),
			second:        buildCompileFileWithPassesOut("helloworld", os.ErrClosed, false),
			match:         true,
			errorExpected: true,
		},
		{
			name:          "run error path",
			first:         buildCompileFileWithPassesOut("hello world", nil, true),
			second:        buildCompileFileWithPassesOut("hello world", nil, true),
			match:         true,
			errorExpected: true,
		},
	}

	backupCompileFileWithPasses := compileFileWithPasses
	defer func() { compileFileWithPasses = backupCompileFileWithPasses }()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compileFileWithPasses = test.first
			hash1, mem1, err1 := getExpectedHash([]string{}, "")

			compileFileWithPasses = test.second
			hash2, mem2, err2 := getExpectedHash([]string{}, "")

			if test.match != (hash1 == hash2) {
				t.Errorf("expected hashes to match, but got %v and %v", hash1, hash2)
			} else if err1 != nil && err2 != nil && mem1 > 0 && mem2 > 0 {
				t.Errorf("expected memory usage to be above 0, got %v and %v", mem1, mem2)
			} else if test.errorExpected && (err1 == nil || err2 == nil) {
				t.Errorf("expected an error, but got nil: %v and %v", err1, err2)
			} else if !test.errorExpected && (err1 != nil || err2 != nil) {
				t.Fatalf("did not expect an error, but got: %v and %v", err1, err2)
			}
		})
	}
}

func TestGetExpectedExecutionStats(t *testing.T) {
	tests := []struct {
		name          string
		override      compilePassFunc
		hash          uint64
		errorExpected bool
	}{
		{
			name:          "happy path",
			override:      buildCompileFileWithPassesOut("hello world", nil, false),
			hash:          calculateHash([]byte("hello world\n")),
			errorExpected: false,
		},
		{
			name:          "compilation error path",
			override:      buildCompileFileWithPassesOut("hello world", os.ErrClosed, false),
			errorExpected: true,
		},
		{
			name:          "run error path",
			override:      buildCompileFileWithPassesOut("hello world", nil, true),
			errorExpected: true,
		},
	}

	backupCompileFileWithPasses := compileFileWithPasses
	defer func() { compileFileWithPasses = backupCompileFileWithPasses }()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compileFileWithPasses = test.override
			setTrueHash(test.hash)
			time, err := getExpectedExecutionStats([]string{}, "")

			if test.errorExpected != (err != nil) {
				t.Errorf("expected error mismatch, got %v, expected? %v", err, test.errorExpected)
			} else if err == nil && time < 0 {
				t.Errorf("expected time > 0, got %v", time)
			}
		})
	}
}

func TestCompileFileWithPasses(t *testing.T) {
	validFile, err := os.CreateTemp("", "testharness_*.c")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(validFile.Name())
	if _, err := validFile.WriteString("int main() { return 0; }\n"); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	validFile.Sync()
	validFile.Close()

	tests := []struct {
		name               string
		compilationCommand []string
		passes             string
		errorExpected      bool
	}{
		{
			name:               "empty command",
			compilationCommand: []string{},
			passes:             "",
			errorExpected:      true,
		},
		{
			name:               "invalid command",
			compilationCommand: []string{"invalidcmd", validFile.Name()},
			passes:             "",
			errorExpected:      true,
		},
		{
			name:               "valid compilation",
			compilationCommand: []string{"clang", validFile.Name()},
			passes:             "default<O1>",
			errorExpected:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			duration, size, file, err := compileFileWithPasses(test.compilationCommand, test.passes, 0, false)

			if file != nil {
				defer os.Remove(file.Name())
				defer file.Close()
			}

			if !test.errorExpected && err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			}

			if err == nil {
				if test.errorExpected {
					t.Errorf("expected an error, but got nil")
				}
				if file == nil {
					t.Errorf("expected a valid file reference, got nil")
				}
				if size <= 0 {
					t.Errorf("expected binary size > 0, got %d", size)
				}
				if duration < 0 {
					t.Errorf("expected duration >= 0, got %v", duration)
				}
			}
		})
	}
}

func TestRunProgramUntilStable(t *testing.T) {
	deletedFile, err := createShellScriptWithOutput("failure")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	os.Remove(deletedFile.Name())

	validFile, err := createShellScriptWithOutput("success")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(validFile.Name())

	invalidFile, err := createShellScriptWithOutput("failure")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(validFile.Name())

	tests := []struct {
		name          string
		file          *os.File
		hash          uint64
		errorExpected bool
	}{
		{
			name:          "happy path",
			file:          validFile,
			hash:          calculateHash([]byte("success\n")),
			errorExpected: false,
		},
		{
			name:          "run error path",
			file:          deletedFile,
			errorExpected: true,
		},
		{
			name:          "invalid hash",
			file:          invalidFile,
			errorExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setTrueHash(test.hash)
			execTime, maxRSS, err := runProgramUntilStable(test.file, time.Hour)

			if test.errorExpected != (err != nil) {
				t.Errorf("expected error mismatch, got %v, expected? %v", err, test.errorExpected)
			} else if err == nil && execTime < 0 {
				t.Errorf("expected time > 0, got %v", execTime)
			} else if err == nil && maxRSS < 0 {
				t.Errorf("expected maxRSS > 0, got %v", maxRSS)
			}
		})
	}
}

// createShellScriptWithOutput creates a shell script with the given output.
//
// Returns the file and an error.
func createShellScriptWithOutput(out string) (*os.File, error) {
	validFile, err := os.CreateTemp("", "testharness_valid_*.sh")
	if err != nil {
		return nil, err
	}
	validFile.Chmod(0755)

	_, err = validFile.WriteString("#!/bin/sh\necho '" + out + "'\n")
	if err != nil {
		return nil, err
	}
	validFile.Sync()
	validFile.Close()

	return validFile, nil
}

// buildCompileFileWithPassesOut builds and returns a compileFileWithPasses function override.
func buildCompileFileWithPassesOut(out string, err error, delete bool) compilePassFunc {
	return func(_ []string, _ string, _ int, _ bool) (time.Duration, int, *os.File, error) {
		file, _ := createShellScriptWithOutput(out)
		if delete {
			os.Remove(file.Name())
		}
		return 0, 0, file, err
	}
}
