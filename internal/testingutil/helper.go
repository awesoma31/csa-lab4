package testingutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
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
	cfg.IOC = ioc
	cfg.MemD = data
	cfg.MemI = ins

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

func getProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	currentDir := filepath.Dir(filename)

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("go.mod not found in any parent directory of %s", filepath.Dir(filename))
		}
		currentDir = parentDir
	}
}
