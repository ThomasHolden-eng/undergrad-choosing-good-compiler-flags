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
	gaPopulation.RunNGenerations(1)

	seq, val, err := gaPopulation.Best()
	if err != nil {
		fmt.Printf("ga error: %v\n", err)
		return
	}
	fmt.Printf("%s\n%v\n", seq, val)
}
