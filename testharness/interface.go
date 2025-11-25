package testharness

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// CompileFileWithPasses takes a file reference, pass list & unique id
//
// Returns compilation time, binary size, file reference & error
//
// Caller must resolve returned executable reference (close and delete)
func CompileFileWithPasses(compilationCommand []string, passes []int, id int, intPassF func(int) string, outLocal bool) (time.Duration, int, *os.File, error) {
	if len(compilationCommand) == 0 {
		return 0, 0, nil, fmt.Errorf("empty command passed")
	}

	passStrings := make([]string, len(passes))
	for i, pass := range passes {
		passStrings[i] = intPassF(pass)
	}

	binFile, err := os.CreateTemp("", fmt.Sprintf("bin-%d-*", id))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	binName := binFile.Name()
	binFile.Close()

	var stderr bytes.Buffer

	filesToOptimise := extractInputFiles(compilationCommand)

	var intermediaryFiles []string
	for range filesToOptimise {
		bcFile, err := os.CreateTemp("", fmt.Sprintf("opt-%d-*.bc", id))
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to create temporary file: %w", err)
		}
		intermediaryFiles = append(intermediaryFiles, bcFile.Name())
		bcFile.Close()
	}
	defer cleanFiles(intermediaryFiles)

	flags := extractFlags(compilationCommand)
	for i := range filesToOptimise {
		stderr.Reset()
		inpFile, outFile := filesToOptimise[i], intermediaryFiles[i]

		command := []string{
			compilationCommand[0], inpFile, "-emit-llvm", "-Xclang", "-disable-llvm-passes",
			"-Xclang", "-disable-llvm-optzns", "-Xclang", "-disable-O0-optnone", "-c",
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

	combinedPasses := strings.Join(passStrings, ",")
	var totalDuration int64
	for i := range intermediaryFiles {
		stderr.Reset()
		inpFile, outFile := intermediaryFiles[i], intermediaryFiles[i]

		// Add O3 here to run backend optimisations, otherwise we optimise only IR
		command := []string{"opt", "-passes=" + combinedPasses, inpFile, "-o", outFile}

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

	newCommand = append(newCommand, "-O3")
	if !outLocal {
		newCommand = append(newCommand, "-o", binName)
	}

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
func RunFile(file *os.File) (time.Duration, int64, error) {
	var stderr bytes.Buffer

	file.Chmod(0755)
	cmd := exec.Command(file.Name())
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
