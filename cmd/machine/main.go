package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/sanity-io/litter"
	"gopkg.in/yaml.v3"
)

type cpuConfig struct {
	InstrMemPath     string         `yaml:"instruction_bin"`
	DataMemPath      string         `yaml:"data_bin"`
	TickLimit        int            `yaml:"tick_limit"`
	Schedule         []io.TickEntry `yaml:"schedule"`
	MaxInterruptions int            `yaml:"max_interruptions"`
}

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to cpu config yml")
	cpu, err := loadCPUFromConfig(*configPath)
	if err != nil {
		slog.Error(fmt.Sprintf("error configuring CPU - %s", err.Error()))
		log.Fatal()
	}
	slog.Info("CPU ocnfigured succesfully, starting simulation")
	fmt.Println("--------------------------------------------------------------")
	cpu.Run()
}

func loadCPUFromConfig(path string) (*machine.CPU, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg cpuConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	ioc := io.NewIOController(cfg.Schedule)
	litter.Dump(cfg.Schedule)

	ins, err := bingen.LoadInstructionMemory(cfg.InstrMemPath)
	if err != nil {
		return nil, err
	}
	data, err := bingen.LoadDataMemory(cfg.DataMemPath)
	if err != nil {
		return nil, err
	}

	cpu := machine.New(ins, data, ioc, cfg.MaxInterruptions, cfg.TickLimit)
	return cpu, nil
}
