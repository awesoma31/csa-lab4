package logutil

import (
	"fmt"
	"log"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/sanity-io/litter"
)

func PrintAst(program ast.BlockStmt) {
	fmt.Println("-------------------AST----------------------")
	litter.Dump(program)
}

func PrintSymTable(cg *codegen.CodeGenerator) {
	fmt.Println("-------------------SymTable--------------------------")
	fmt.Println("[var_name | addres]")
	scopeStack := cg.ScopeStack()

	for k, v := range scopeStack[0].Symbols() {
		fmt.Print(k, " | ")
		fmt.Printf(" %X\n", v.AbsAddress)
	}
}

func PrintDataMem(dataMemory []byte) {
	fmt.Println("-------------------dataMemory----------------------")
	for i, val := range dataMemory {
		if i%4 == 0 {
			fmt.Println("_____")
		}
		fmt.Println(fmt.Sprintf("[0x%X|%d]:", i, i), fmt.Sprintf("0x%02X", val))
	}
}

func PrintInstrMem(instructionMemory []uint32) {

	fmt.Println("-------------------instructionMemory----------------------")
	for i, instr := range instructionMemory {
		if i <= len(instructionMemory) {
			fmt.Println(
				fmt.Sprintf("[0x%04X|%04d]:", i, i),
				fmt.Sprintf("0x%08X - %d", instr, instr),
			)
		}
	}
}

func PrintDebugAsm(debugAssembly []string) {
	fmt.Println("-------------------debugAssembly----------------------")
	for _, val := range debugAssembly {
		fmt.Println(val)
	}
}

func DumpAst(program ast.BlockStmt, filePath string, debug bool) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file %s: %v", filePath, err)
		return
	}
	defer closeFile(file)
	astString := litter.Sdump(program)
	_, err = fmt.Fprint(file, astString)
	if err != nil {
		log.Printf("Error writing AST to file %s: %v", filePath, err)
	}

	if debug {
		PrintAst(program)
	}
}

func DumpSymTable(cg *codegen.CodeGenerator, filePath string, debug bool) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file %s: %v", filePath, err)
		return
	}
	defer closeFile(file)

	_, _ = fmt.Fprintf(file, "[var_name | addres]\n")
	scopeStack := cg.ScopeStack()

	for k, v := range scopeStack[0].Symbols() {
		_, _ = fmt.Fprintf(file, "%s |  %X\n", k, v.AbsAddress)
	}

	if debug {
		PrintSymTable(cg)
	}
}

// DumpMemDLog writes the data memory to a specified file.
func DumpMemDLog(dataMemory []byte, filePath string, debug bool) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file %s: %v", filePath, err)
		return
	}
	defer closeFile(file)
	for i, val := range dataMemory {
		if i%4 == 0 {
			_, _ = fmt.Fprintf(file, "_____\n")
		}
		_, _ = fmt.Fprintf(file, "[0x%X|%d]: 0x%02X\n", i, i, val)
	}
	if debug {
		PrintDataMem(dataMemory)
	}
}

// DumpMemILog writes the instruction memory to a specified file.
func DumpMemILog(instructionMemory []uint32, filePath string, debug bool) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file %s: %v", filePath, err)
		return
	}
	defer closeFile(file)

	for i, instr := range instructionMemory {
		if i <= len(instructionMemory) {
			_, _ = fmt.Fprintf(file, "[0x%04X|%04d]: 0x%08X - %d\n", i, i, instr, instr)
		}
	}
	if debug {
		PrintInstrMem(instructionMemory)
	}
}

// DumpDebugInstrLog writes the debug assembly to a specified file.
func DumpDebugInstrLog(debugAssembly []string, filePath string, debug bool) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file %s: %v", filePath, err)
		return
	}
	defer closeFile(file)
	for _, val := range debugAssembly {
		_, _ = fmt.Fprintf(file, "%s\n", val)
	}
	if debug {
		PrintDebugAsm(debugAssembly)
	}
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Printf("Error closing file %s: %v", f.Name(), err)
	}
}
