package main

import (
	"bytes"
	"encoding/binary"
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

	//cg := NewCG()
	//var dataMem []uint32 // собираем параллельно
	//dataMem := make([]Word, 4096) // или сколько нужно
	//cg := &CodeGen{
	//	Sym:  NewSymTab(),
	//	Data: &dataMem,
	//}
	//
	//cg.Sym.nextData = 0x1000
	//
	//for _, stmt := range astTree.Body {
	//	fmt.Println()
	//	switch n := stmt.(type) {
	//
	//	case *ast.VarDeclarationStmt:
	//		cg.genVarDecl(n)
	//	default:
	//		panic("only var decls in demo")
	//	}
	//}
	//
	//saveWords("program.bin", cg.Instr)
	//saveWords("data.bin", dataMem)
	//err = os.WriteFile("program.txt", []byte(strings.Join(cg.Report, "\n")), 0644)
	//if err != nil {
	//	return
	//}
	// litter.Dump(ast)
}

func getPrettyJson(in []byte) (string, error) {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, []byte(in), "", " ")
	if err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}

func saveWords(name string, words []uint32) {
	f, _ := os.Create(name)
	defer f.Close()
	for _, w := range words {
		_ = binary.Write(f, binary.LittleEndian, w)
	}
}
