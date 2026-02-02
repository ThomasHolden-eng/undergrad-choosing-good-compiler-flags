package testharness

import (
	"6CCS3PRJ/thomasholden/util"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// CompileFileWithPasses takes a file reference, pass list & unique id
//
// Returns compilation time, binary size, file reference & error
//
// Caller must resolve returned executable reference (close and delete)
func CompileFileWithPasses(compilationCommand []string, passes string, id int, outLocal bool) (time.Duration, int, *os.File, error) {
	if len(compilationCommand) == 0 {
		return 0, 0, nil, fmt.Errorf("empty command passed")
	}

	binFile, err := os.CreateTemp("", fmt.Sprintf("bin-%d-*", id))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	binName := binFile.Name()
	binFile.Close()

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
		"-Xclang", "-disable-llvm-optzns")

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

// Takes a file reference
// Returns execution time, memory usage & error
func RunFile(file *os.File, maxDuration time.Duration) (time.Duration, int64, error) {
	var stderr bytes.Buffer

	ctx := context.Background()
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	file.Chmod(0755)
	cmd := exec.CommandContext(ctx, file.Name())
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("execution failed: %w\nStderr: %s", err, stderr.String())
	}

	duration := cmd.ProcessState.UserTime().Microseconds()

	// Will fail within windows
	memUsage := func() int64 {
		if usage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
			return usage.Maxrss
		}
		return 0
	}()

	totalTime := time.Duration(duration * int64(time.Microsecond))
	return totalTime, memUsage, nil
}

// GetSequenceExecutionTime returns only the execution time of a pass sequence.
func GetSequenceExecutionTime(compilationCommand []string, passes string, outLocal bool) (time.Duration, error) {
	_, _, f, err := CompileFileWithPasses(compilationCommand, passes, 0, outLocal)
	defer os.Remove(f.Name())
	defer f.Close() // defer is LIFO

	if err != nil {
		return time.Second, fmt.Errorf("compilation failure: %w", err)
	}

	timeList := make([]time.Duration, 3)
	for i := range 3 {
		execTime, _, err := RunFile(f, time.Hour*24)
		if err != nil {
			return time.Second, fmt.Errorf("execution failure: %w", err)
		}
		timeList[i] = execTime
	}

	return time.Duration(util.Median(timeList)), nil
}
