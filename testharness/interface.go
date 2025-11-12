package testharness

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Takes a file reference, pass list & unique id
// Returns compilation time, binary size, file reference & error
func CompileFileWithPasses(file *os.File, passes []int, id int) (time.Duration, int, *os.File, error) {
	inputFileName := file.Name()

	passStrings := make([]string, len(passes))
	for i, pass := range passes {
		passStrings[i] = IntToFlag(pass)
	}

	optFile, err := os.CreateTemp("", fmt.Sprintf("opt-%d-*.ll", id))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	optName := optFile.Name()
	optFile.Close()

	objFile, err := os.CreateTemp("", fmt.Sprintf("obj-%d-*.o", id))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to create second temporary file: %w", err)
	}
	objName := objFile.Name()
	objFile.Close()

	binFile, err := os.CreateTemp("", fmt.Sprintf("bin-%d-*", id))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to create third temporary file: %w", err)
	}
	binName := binFile.Name()
	binFile.Close()

	defer os.Remove(optName)
	defer os.Remove(objName)

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	optArgs := []string{"\\time", "-p", "opt"}
	optArgs = append(optArgs, "-passes="+strings.Join(passStrings, ","))
	optArgs = append(optArgs, "-o", optName, inputFileName)
	cmd := exec.Command(optArgs[0], optArgs[1:]...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return 0, 0, nil, fmt.Errorf("opt failed: %w\nStderr: %s", err, stderr.String())
	}

	optDuration, err := extractUserTime(stdout.String())
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to extract time from: %s", stdout.String())
	}

	stderr.Reset()
	stdout.Reset()

	llcArgs := []string{"\\time", "-p", "llc", "-filetype=obj", "-o", objName, optName}
	cmdLlc := exec.Command(llcArgs[0], llcArgs[1:]...)
	cmdLlc.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmdLlc.Run(); err != nil {
		return 0, 0, nil, fmt.Errorf("llc failed: %w\nStderr: %s", err, stderr.String())
	}

	llcDuration, err := extractUserTime(stdout.String())
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to extract time from: %s", stdout.String())
	}

	stderr.Reset()
	stdout.Reset()

	clangArgs := []string{"\\time", "-p", "clang", "-o", binName, objName}
	cmdLink := exec.Command(clangArgs[0], clangArgs[1:]...)
	cmdLink.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmdLink.Run(); err != nil {
		return 0, 0, nil, fmt.Errorf("link (clang) failed: %w\nStderr: %s", err, stderr.String())
	}

	clangDuration, err := extractUserTime(stdout.String())
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to extract time from: %s", stdout.String())
	}

	finalFile, err := os.Open(binName)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to open final binary '%s': %w", binName, err)
	}

	binSize, err := getFileSize(finalFile)
	if err != nil {
		finalFile.Close()
		return 0, 0, nil, fmt.Errorf("failed to size binary '%s': %w", binName, err)
	}

	totalUserTime := time.Duration(optDuration + llcDuration + clangDuration)
	return totalUserTime, binSize, finalFile, nil
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
