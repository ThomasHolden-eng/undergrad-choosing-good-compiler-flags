package main

import (
	"6CCS3PRJ/thomasholden/geneticalgorithm"

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

	pool := geneticalgorithm.NewWorkerPool(commandSplit, 1, 1)
	results := pool.StartPool()

	fmt.Printf("%v\n", results)
}
