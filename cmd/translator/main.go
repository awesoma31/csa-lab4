package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

func main() {

	var codeFile string
	if len(os.Args) > 1 {
		codeFile = os.Args[1]
	}
	sourceBytes, err := os.ReadFile(codeFile)
	if err != nil {
		sourceBytes, _ = os.ReadFile("examples/00.lang")
	}
	source := string(sourceBytes)
	astTree := parser.Parse(source)

	astJson, _ := json.Marshal(astTree)
	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, []byte(astJson), "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	a, _ := getPrettyJson(astJson)
	// fmt.Println(string(astJson))
	fmt.Println(a)
	// fmt.Println(string(prettyJson.String()))
}

func getPrettyJson(in []byte) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(in), "", " ")
	if err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}
