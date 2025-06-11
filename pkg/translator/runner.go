package translator

import (
	"fmt"
	"os"
	"path/filepath"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/logutil"
	"github.com/awesoma31/csa-lab4/pkg/translator/codegen"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

type Options struct {
	SrcPath string
	OutDir  string
	LogDir  string
	Debug   bool
}

func Run(opts Options) (imem []uint32, dmem []byte, err error) {
	src, err := os.ReadFile(opts.SrcPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read src: %w", err)
	}
	ast, pErr := parser.Parse(string(src))
	if len(pErr) != 0 {
		return nil, nil, fmt.Errorf("parse: %v", pErr)
	}

	cg := codegen.NewCodeGenerator()
	imem, dmem, dbgAsm, cgErr := cg.Generate(ast)
	if len(cgErr) != 0 {
		return nil, nil, fmt.Errorf("codegen: %v", cgErr)
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(opts.LogDir, 0o755); err != nil {
		return nil, nil, err
	}

	_ = bingen.SaveInstructionMemory(filepath.Join(opts.OutDir, "instr.bin"), imem)
	_ = bingen.SaveDataMemory(filepath.Join(opts.OutDir, "data.bin"), dmem)

	logutil.DumpAst(ast, filepath.Join(opts.LogDir, "ast.log"), opts.Debug)
	logutil.DumpDebugInstrLog(dbgAsm, filepath.Join(opts.LogDir, "debugIntr.log"), opts.Debug)
	logutil.DumpMemILog(imem, filepath.Join(opts.LogDir, "instr.log"), opts.Debug)
	logutil.DumpMemDLog(dmem, filepath.Join(opts.LogDir, "data.log"), opts.Debug)
	logutil.DumpSymTable(cg, filepath.Join(opts.LogDir, "symtable.log"), opts.Debug)

	fmt.Printf("binaries saved to %s\n", opts.OutDir)
	return imem, dmem, nil
}
