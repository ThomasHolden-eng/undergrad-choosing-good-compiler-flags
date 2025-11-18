package testharness

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CompileFileWithPasses takes a file reference, pass list & unique id
//
// Returns compilation time, binary size, file reference & error
//
// Caller must resolve returned executable reference (close and delete)
func CompileFileWithPasses(compilationCommand []string, passes []int, id int) (time.Duration, int, *os.File, error) {
	if len(compilationCommand) == 0 {
		return 0, 0, nil, fmt.Errorf("empty command passed")
	}

	passStrings := make([]string, len(passes))
	for i, pass := range passes {
		passStrings[i] = IntToFlag(pass)
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

	for i := range filesToOptimise {
		stderr.Reset()
		inpFile, outFile := filesToOptimise[i], intermediaryFiles[i]

		command := []string{}
		command = append(command, compilationCommand[0], inpFile)
		command = append(command, []string{"-emit-llvm", "-Xclang", "-disable-llvm-passes", "-Xclang", "-disable-llvm-optzns"}...)
		command = append(command, "-o", outFile)

		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			os.Remove(binName)
			return 0, 0, nil, fmt.Errorf("compilation failed: %w\nStderr: %s", err, stderr.String())
		}
	}

	combinedPasses := strings.Join(passStrings, ",")
	var totalDuration float32
	for i := range intermediaryFiles {
		stderr.Reset()
		inpFile, outFile := intermediaryFiles[i], intermediaryFiles[i]

		command := []string{"/usr/bin/time", "-p",
			"opt", "-passes=" + combinedPasses, inpFile, "-o", outFile}

		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			os.Remove(binName)
			return 0, 0, nil, fmt.Errorf("compilation failed: %w\nStderr: %s", err, stderr.String())
		}

		duration, err := extractUserTime(stderr.String())
		if err != nil {
			os.Remove(binName)
			return 0, 0, nil, fmt.Errorf("failed to extract time from: %s", stderr.String())
		}
		totalDuration += duration
	}

	mapInputFilesToBC(compilationCommand, intermediaryFiles)

	compilationCommand = append(compilationCommand, "-o", binName)

	cmd := exec.Command(compilationCommand[0], compilationCommand[1:]...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.Remove(binName)
		return 0, 0, nil, fmt.Errorf("compilation failed: %w\nStderr: %s", err, stderr.String())
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

	totalTime := time.Duration(totalDuration * float32(time.Second))
	return totalTime, binSize, binFile, nil
}

// Takes a file reference and unique id
// Returns execution time, memory usage & error
func RunFile(file *os.File, id int) (time.Duration, int, error) {
	return 0, 0, nil
}

func FlagToInt(s string) int {
	return 0
}

func IntToFlag(n int) string {
	return ""
}
