package translator

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

func Translate(inPath string, memIPath, memDPath string) {
	sourceBytes, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Printf("couldn't resolve %s: %v\n", inPath, err)
		os.Exit(1)
	}

	progAst, parseErrors := parser.Parse(string(sourceBytes))
	if len(parseErrors) > 0 {
		fmt.Println("Parse errors:")
		for _, err := range parseErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	// printAst(program)

	cg := codegen.NewCodeGenerator()
	memI, memD, _, cgErrors := cg.Generate(progAst)
	if len(cgErrors) > 0 {
		for _, e := range cgErrors {
			fmt.Println("[TRANSLATE ERROR]:", e)
		}
		os.Exit(1)
	}

	// printDebugAsm(debugAssembly)
	// printInstrMem(memI)
	// printDataMem(memD)
	// printSymTable(cg)

	instrMemPath := memIPath
	err = bingen.SaveInstructionMemory(instrMemPath, memI)
	if err != nil {
		slog.Error(fmt.Sprintf("error writeing instr mem bin - %s", err.Error()))
		log.Fatal("fatal")
	}
	// slog.Info(fmt.Sprintf("instructionMemory saved to %s", instrMemPath))

	dataMemPath := memDPath
	err = bingen.SaveDataMemory(dataMemPath, memD)
	if err != nil {
		slog.Error(fmt.Sprintf("error writeing data mem bin - %s", err.Error()))
		log.Fatal()
	}
	// slog.Info(fmt.Sprintf("data memory saved to %s", dataMemPath))

}
