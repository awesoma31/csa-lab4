package testingutil

import (
	"os"
	"path/filepath"
	"testing"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/machine/logger"
	"github.com/awesoma31/csa-lab4/pkg/translator"
	"gopkg.in/yaml.v2"
)

func RunGolden(t *testing.T, dir string, tickLimit int) {
	t.Helper()

	src := filepath.Join(dir, "src.lang")
	cfgPath := filepath.Join(dir, "config.yaml")
	outDir := filepath.Join(dir, "logs")

	_, _, err := translator.Run(translator.Options{
		SrcPath:   src,
		OutDir:    outDir,
		Debug:     false,
		DumpFiles: true,
	})
	if err != nil {
		cwd, _ := os.Getwd()
		t.Fatalf("translate: %v, \n CWD- %s", err, cwd)
	}
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("cpu config, %s", err.Error())
	}
	var cfg *machine.CpuConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err.Error())
	}

	ioc := io.NewIOController(cfg.Schedule)

	ins, err := bingen.LoadInstructionMemory(cfg.InstrMemPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	data, err := bingen.LoadDataMemory(cfg.DataMemPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	lg := logger.New(cfg.Debug, cfg.LogFilePath)

	cfg.IOC = ioc
	cfg.MemD = data
	cfg.MemI = ins
	cfg.Logger = lg

	cpu := machine.New(cfg)
	cpu.Run()

	// // ── сравниваем вывод ──────────────────────────────────────
	// gotOut := ioc.OutputAll()
	// wantOut, _ := os.ReadFile(filepath.Join(goldenDir, "out.txt"))
	// if diff := cmpBytes(gotOut, wantOut); diff != "" {
	// 	t.Fatalf("stdout mismatch (-got +want):\n%s", diff)
	// }
	//
	// // ── трасса ────────────────────────────────────────────────
	// gotTrace := cpu.Trace() // верни slice []string из CPU
	// wantTrace, _ := os.ReadFile(filepath.Join(goldenDir, "trace.log"))
	// if needUpdate() {
	// 	panic("unimpl update golden")
	// 	// os.WriteFile(...)
	// }
	// if diff := cmpLines(gotTrace, wantTrace); diff != "" {
	// 	t.Fatal(diff)
	// }
}
