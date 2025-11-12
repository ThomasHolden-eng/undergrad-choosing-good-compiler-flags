package main

import (
	"6CCS3PRJ/thomasholden/testharness"
	"fmt"
	"os"
)

func main() {
	helloWorldFile, err := os.Open("helloworld.ll")
	if err != nil {
		return
	}
	fmt.Println(testharness.CompileFileWithPasses(helloWorldFile, []int{}, 0))
}
