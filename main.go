package main

import (
	"6CCS3PRJ/thomasholden/geneticalgorithm"
	"flag"
	"time"

	"bufio"
	"fmt"
	"os"
)

// command line args:
// -p population size
// -l max sequence length
// -t maximum run time e.g. 1s 1m 1h
// -c command
func main() {
	popSize := flag.Int("p", 30, "population size")
	maxLength := flag.Int("l", 200, "max sequence length")
	maxTime := flag.Duration("t", 2*time.Hour, "maximum run time (e.g., 10s, 1m)")
	command := flag.String("c", "", "command")

	flag.Parse()

	commandText := *command
	if commandText == "" {
		fmt.Print("Enter command: ")
		reader := bufio.NewReader(os.Stdin)
		// e.g. clang++ -I./example/string_concat/include ./example/string_concat/src/TestStringCons.cpp
		newCommand, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("text input error: %v\n", err)
			return
		}
		commandText = newCommand
	}

	gaPopulation, err := geneticalgorithm.CreatePopulation(*popSize, *maxLength, commandText)
	if err != nil {
		fmt.Printf("ga error: %v\n", err)
		return
	}
	generations := gaPopulation.RunGenerationsForDuration(*maxTime)

	seq, err := gaPopulation.Best()
	if err != nil {
		fmt.Printf("ga error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", seq)
	fmt.Printf("Converged in %d generations.\n", generations)
}
