package main

import (
	"6CCS3PRJ/thomasholden/geneticalgorithm"

	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Print("Enter command: ")
	reader := bufio.NewReader(os.Stdin)
	// e.g. clang++ -I./example/string_concat/include ./example/string_concat/src/TestStringCons.cpp
	commandText, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("text input error: %v\n", err)
		return
	}

	gaPopulation, err := geneticalgorithm.CreatePopulation(30, 200, commandText)
	if err != nil {
		fmt.Printf("ga error: %v\n", err)
		return
	}
	gaPopulation.RunNGenerations(2)

	seq, err := gaPopulation.Best()
	if err != nil {
		fmt.Printf("ga error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", seq)
}
