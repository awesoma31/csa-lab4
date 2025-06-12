package testingutil

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/machine/logger"
	"github.com/awesoma31/csa-lab4/pkg/translator"
	"gopkg.in/yaml.v2"
)

var update = flag.Bool("u", false, "rewrite *.golden files")

func RunGolden(t *testing.T, dir string, tickLimit int) {
	t.Helper()

	src := filepath.Join(dir, "src.lang")
	cfgPath := filepath.Join(dir, "config.yaml")
	outDir := filepath.Join(dir)
	logDir := filepath.Join(dir, "logs")
	goldenOutputPath := filepath.Join(dir, "output.golden")

	_, _, err := translator.Run(translator.Options{
		SrcPath: src,
		OutDir:  outDir,
		LogDir:  logDir,
		Debug:   false,
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
	gotOutput := cpu.Run()

	if *update {
		t.Log("updating golden file...")
		err := os.WriteFile(goldenOutputPath, []byte(gotOutput), 0644)
		if err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		return
	}

	wantOutputBytes, err := os.ReadFile(goldenOutputPath)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %v", goldenOutputPath, err)
	}
	wantOutput := string(wantOutputBytes)

	gotOutput = strings.TrimSuffix(gotOutput, "\n")
	wantOutput = strings.TrimSuffix(wantOutput, "\n")

	if gotOutput != wantOutput {
		diff := cmpLines(gotOutput, wantOutput)
		t.Errorf("output mismatch (-got +want):\n%s", diff)
	}
}

func cmpLines(got, want string) string {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")

	var diff strings.Builder
	maxLen := max(len(wantLines), len(gotLines))

	for i := range maxLen {
		currentGotLine := ""
		if i < len(gotLines) {
			currentGotLine = gotLines[i]
		}
		currentWantLine := ""
		if i < len(wantLines) {
			currentWantLine = wantLines[i]
		}

		if currentGotLine != currentWantLine {
			if i < len(gotLines) {
				fmt.Fprintf(&diff, "-%s\n", currentGotLine)
			}
			if i < len(wantLines) {
				fmt.Fprintf(&diff, "+%s\n", currentWantLine)
			}
		}
	}
	return diff.String()
}
