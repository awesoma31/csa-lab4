package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Print("Enter expression: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Fprintln(os.Stderr, "No input")
		os.Exit(1)
	}
	line := scanner.Text()
	parser := NewMathParser(line)
	ast := parser.ParseMath()
	fmt.Println("\nAST:")
	fmt.Print(ast.String(""))
}
