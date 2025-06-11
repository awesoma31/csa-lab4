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
	"github.com/sanity-io/litter"
	"gopkg.in/yaml.v2"
)

type cpuConfig struct {
	InstrMemPath     string         `yaml:"instruction_bin"`
	DataMemPath      string         `yaml:"data_bin"`
	TickLimit        int            `yaml:"tick_limit"`
	Schedule         []io.TickEntry `yaml:"schedule"`
	MaxInterruptions int            `yaml:"max_interruptions"`
}

func RunGolden(t *testing.T, dir string, tickLimit int) {
	t.Helper()
	src := filepath.Join(dir, "src.lang")
	out := filepath.Join(dir, "logs") // сохраняем внутрь каталога теста

	_, _, err := translator.Run(translator.Options{
		SrcPath: src, OutDir: out,
		Debug:     false, // не шумим в stdout
		DumpFiles: true,
	})
	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	// root, err := getProjectRoot()
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }
	// dir := filepath.Join(root, goldenDir)
	//
	// sourceLangPath := filepath.Join(dir, "src.lang")
	// memIPath := filepath.Join(dir, "instr.bin")
	// memDPath := filepath.Join(dir, "data.bin")
	cfgPath := filepath.Join(dir, "config.yaml")
	// // fmt.Println(os.L)
	// // println(sourceLangPath, memIPath, memDPath, cfgPath)
	//
	// translator.Translate(sourceLangPath, memIPath, memDPath)
	//
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	var cfg cpuConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err.Error())
	}

	ioc := io.NewIOController(cfg.Schedule)
	litter.Dump(cfg.Schedule)

	ins, err := bingen.LoadInstructionMemory(cfg.InstrMemPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	data, err := bingen.LoadDataMemory(cfg.DataMemPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	cpu := machine.New(ins, data, ioc, cfg.MaxInterruptions, cfg.TickLimit)
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
