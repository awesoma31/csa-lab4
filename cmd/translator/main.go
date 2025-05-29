package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

func main() {
	//user, err := user.Current()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Hello %s! This is the Monkey programming language!\n",
	//	user.Username)
	//fmt.Printf("Feel free to type in commands\n")
	//repl.Start(os.Stdin, os.Stdout)

	var codeFile string
	if len(os.Args) > 1 {
		codeFile = os.Args[1]
	}
	sourceBytes, err := os.ReadFile(codeFile)
	if err != nil {
		sourceBytes, _ = os.ReadFile("examples/00.lang")
	}
	input := string(sourceBytes)

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("parser errors:")
		for _, e := range p.Errors() {
			fmt.Println("\t", e)
		}
		return
	}

	fmt.Println(getJsonAst(program))
}

func getJsonAst(program *ast.Program) string {
	astJson, _ := json.Marshal(program)
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(astJson), "", " ")
	if err != nil {
		panic(err)
	}
	a, _ := getPrettyJson(astJson)
	return a
}

func getPrettyJson(in []byte) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(in), "", " ")
	if err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}
