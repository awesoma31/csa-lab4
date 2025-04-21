package main

import (
	"github.com/awesoma31/csa-lab4/cmd/translator/lexer"
)

func main() {
	t := lexer.Tokenize("1 + 2 -  -3")

	for _, tok := range t {
		tok.Debug()
	}
}
