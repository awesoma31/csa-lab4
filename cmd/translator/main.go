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
	fmt.Println(line)
	l := "1+2*3^4-(5)"
	// parser := NewMathParser(line)
	// ast := parser.ParseMath()
	fmt.Println(l)
	parser := NewMathParser(l)
	ast := parser.ParseMathExpr(l)
	fmt.Println("\nAST:")
	fmt.Print(ast.String(""))
}
