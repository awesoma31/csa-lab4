package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
	"github.com/sanity-io/litter"
)

func main() {
	flags := &flags{}
	flags.parseFlags()

	sourceBytes, err := os.ReadFile(flags.InPath)
	if err != nil {
		fmt.Printf("couldn't resolve %s: %v\n", flags.InPath, err)
		os.Exit(1)
	}

	program, parseErrors := parser.Parse(string(sourceBytes))
	if len(parseErrors) > 0 {
		fmt.Println("Ошибки парсера:")
		for _, err := range parseErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	printAst(program)

	cg := codegen.NewCodeGenerator()
	instructionMemory, dataMemory, debugAssembly, cgErrors := cg.Generate(program)
	if len(cgErrors) > 0 {
		for _, e := range cgErrors {
			fmt.Println("[TRANSLATE ERROR]:", e)
		}
		os.Exit(1)
	}

	printDebugAsm(debugAssembly)
	printInstrMem(instructionMemory)
	printDataMem(dataMemory)
	printSymTable(cg)

	//TODO: write bin
}

func printAst(program ast.BlockStmt) {
	fmt.Println("-------------------AST----------------------")
	litter.Dump(program)
}

func printSymTable(cg *codegen.CodeGenerator) {
	fmt.Println("-------------------SymTable--------------------------")
	fmt.Println("[var_name | addres]")
	scopeStack := cg.ScopeStack()

	for k, v := range scopeStack[0].Symbols() {
		fmt.Print(k, " | ")
		fmt.Println(v.AbsAddress)
	}
}

func printDataMem(dataMemory []byte) {
	fmt.Println("-------------------dataMemory----------------------")
	for i, val := range dataMemory {
		if i%4 == 0 {
			fmt.Println("_____")
		}
		fmt.Println(fmt.Sprintf("[0x%X|%d]:", i, i), fmt.Sprintf("0x%02X", val))
	}
}

func printInstrMem(instructionMemory []uint32) {
	fmt.Println("-------------------instructionMemory----------------------")
	for i, instr := range instructionMemory {
		// fmt.Println(instr)
		fmt.Println(
			fmt.Sprintf("[0x%04X|%04d]:", i, i),
			fmt.Sprintf("0x%08X - %d", instr, instr),
		)
	}
}

func printDebugAsm(debugAssembly []string) {
	fmt.Println("-------------------debugAssembly----------------------")
	for _, val := range debugAssembly {
		fmt.Println(val)
	}
}

type flags struct {
	InPath     string
	OutDirPath string
}

func (f *flags) parseFlags() {
	flag.StringVar(&f.InPath, "in", "", "файл исходной программы (*.lang)")
	flag.StringVar(&f.OutDirPath, "out", "out", "каталог с результатами компиляции")
	flag.Parse()

	if f.InPath == "" {
		fmt.Println("usage: translator -in=source.lang [-out dir]")
		os.Exit(1)
	}
}

// getPrettyJson форматирует JSON байты в удобочитаемую строку с отступами.
func getPrettyJson(in []byte) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, in, "", " ")
	if err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}
