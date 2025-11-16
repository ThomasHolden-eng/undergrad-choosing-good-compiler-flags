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

	t, s, f, err := testharness.CompileFileWithPasses(commandSplit, []int{}, 0)
	if f != nil {
		f.Close()
		defer os.Remove(f.Name())
	}

	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Time Taken: %v\n", t)
	fmt.Printf("Size in Bytes: %v\n", s)
}
