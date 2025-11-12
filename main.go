package main

import (
	"6CCS3PRJ/thomasholden/testharness"
	"fmt"
	"os"
)

func main() {
	helloWorldFile, err := os.Open("helloworld.ll")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	t, s, f, err := testharness.CompileFileWithPasses(helloWorldFile, []int{}, 0)
	f.Close()
	defer os.Remove(f.Name())

	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Time Taken: %v\n", t)
	fmt.Printf("Size in Bytes: %v\n", s)
}
