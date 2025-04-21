package main

import (
	"os"

	"github.com/awesoma31/csa-lab4/cmd/translator/lexer"
)

func main() {
	bytes, err := os.ReadFile("examples/00.lang")
	if err != nil {
		panic(err)
	}
	text := string(bytes)

	t := lexer.Tokenize(text)

	for _, tok := range t {
		tok.Debug()
	}
}
