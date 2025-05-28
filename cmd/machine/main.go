package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	if len(args) != 3 {
		panic("Wrong arguments: machine.py <code_file> <input_file>")
	}

	args = os.Args[1:]
	code_file := args[0]
	input_file := args[1]

	simulation()

	fmt.Print(code_file, input_file)
}

func simulation() {
	controlUnit := NewControlUnit()
	dataPath := NewDataPath()

	fmt.Print(controlUnit, dataPath)
}
