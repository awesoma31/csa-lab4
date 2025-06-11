package main

import (
	"flag"
	"fmt"
	"github.com/awesoma31/csa-lab4/pkg/logutil"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

var instrAmount int = 0

func main() {
	flags := &flags{}
	flags.parseFlags()

	sourceBytes, err := os.ReadFile(flags.InPath)
	if err != nil {
		fmt.Printf("couldn't resolve %s: %v\n", flags.InPath, err)
		os.Exit(1)
	}

	ast, parseErrors := parser.Parse(string(sourceBytes))
	if len(parseErrors) > 0 {
		fmt.Println("Ошибки парсера:")
		for _, err := range parseErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	cg := codegen.NewCodeGenerator()
	memI, memD, debugAssembly, cgErrors := cg.Generate(ast)
	if len(cgErrors) > 0 {
		for _, e := range cgErrors {
			fmt.Println("[TRANSLATE ERROR]:", e)
		}
		os.Exit(1)
	}

	instrAmount = int(cg.NextInstructionAddres())

	if flags.Debug {
		logutil.PrintAst(ast)
		logutil.PrintDebugAsm(debugAssembly)
		logutil.PrintInstrMem(memI)
		logutil.PrintDataMem(memD)
		logutil.PrintSymTable(cg)
	}

	logDirPath := "logs"

	if err := os.MkdirAll(logDirPath, 0o755); err != nil {
		log.Fatalf("Error creating log directory %s: %v", logDirPath, err)
	}
	logutil.WriteAstToFile(ast, filepath.Join(logDirPath, "ast.log"))
	logutil.WriteDataMemLogToFile(memD, "logs/memD.log")
	logutil.WriteDebugInstrLogToFile(debugAssembly, filepath.Join(logDirPath, "instr.log"))
	logutil.WriteInstrMemLogToFile(memI, instrAmount, filepath.Join(logDirPath, "memI.log"))
	logutil.WriteSymTableLogToFile(cg, filepath.Join(logDirPath, "symtable.log"))

	memIPath := "bin/instr.bin"
	err = bingen.SaveInstructionMemory(memIPath, memI)
	if err != nil {
		slog.Error(fmt.Sprintf("error writeing instr mem bin - %s", err.Error()))
		log.Fatal("fatal")
	}
	slog.Info(fmt.Sprintf("instructionMemory saved to %s", memIPath))

	memDPath := "bin/data.bin"
	err = bingen.SaveDataMemory(memDPath, memD)
	if err != nil {
		slog.Error(fmt.Sprintf("error writeing data mem bin - %s", err.Error()))
	}
	slog.Info(fmt.Sprintf("data memory saved to %s", memDPath))
}

func Translate(inPath string, memIPath, memDPath string) {
	sourceBytes, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Printf("couldn't resolve %s: %v\n", inPath, err)
		os.Exit(1)
	}

	ast, parseErrors := parser.Parse(string(sourceBytes))
	if len(parseErrors) > 0 {
		fmt.Println("Parse errors:")
		for _, err := range parseErrors {
			fmt.Println("-", err)
		}
		os.Exit(1)
	}

	// printAst(program)

	cg := codegen.NewCodeGenerator()
	memI, memD, _, cgErrors := cg.Generate(ast)
	if len(cgErrors) > 0 {
		for _, e := range cgErrors {
			fmt.Println("[TRANSLATE ERROR]:", e)
		}
		os.Exit(1)
	}

	// instrAmount = int(cg.NextInstructionAddres())

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
	}
	// slog.Info(fmt.Sprintf("data memory saved to %s", dataMemPath))

}

type flags struct {
	InPath     string
	OutDirPath string
	Debug      bool
}

func (f *flags) parseFlags() {
	flag.StringVar(&f.InPath, "in", "", "source file path")
	flag.StringVar(&f.OutDirPath, "out", "out", "directory to save bin files ")
	flag.BoolVar(&f.Debug, "debug", false, "print dumps to stdout")
	flag.Parse()

	if f.InPath == "" {
		fmt.Println("usage: translator -in=source-path [-out dir] [-debug]")
		os.Exit(1)
	}
}
