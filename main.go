package main

import (
	"6CCS3PRJ/thomasholden/testharness"
	"bufio"
	"fmt"
	"os"

	"github.com/google/shlex"
)

func main() {
	fmt.Print("Enter command: ")
	reader := bufio.NewReader(os.Stdin)
	commandText, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("text input error: %v\n", err)
		return
	}

	commandSplit, err := shlex.Split(commandText)
	if err != nil {
		fmt.Printf("text lex error: %v\n", err)
		return
	}

	compileT, size, file, err := testharness.CompileFileWithPasses(commandSplit, []int{}, 0)
	if err != nil {
		fmt.Printf("compilation error: %v\n", err)
	}
	if file != nil {
		fmt.Println("Compiled!")
		fmt.Printf("Time Taken: %v\n", compileT)
		fmt.Printf("Size in Bytes: %v\n", size)

		defer os.Remove(file.Name())

		executionT, memoryUsage, err := testharness.RunFile(file)
		file.Close()
		if err != nil {
			fmt.Printf("execution error: %v\n", err)
		} else {
			fmt.Println("Executed!")
			fmt.Printf("Time Taken: %v\n", executionT)
			fmt.Printf("Size in Bytes: %v\n", memoryUsage)
		}
	}
}
