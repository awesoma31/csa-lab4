package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"gopkg.in/yaml.v3"
)

type config struct {
	InstrMemPath string `yaml:"instruction_bin"`
	DataMemPath  string `yaml:"data_bin"`
	TickLimit    int    `yaml:"tick_limit"`
}

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to cpu config yml")
	machine, err := loadCPUFromConfig(*configPath)
	if err != nil {
		slog.Error(fmt.Sprintf("error configuring CPU - %s", err.Error()))
		log.Fatal()

	}
	slog.Info("CPU ocnfigured succesfully, starting simulation")
	machine.Run(1000)

	// machine.Start()
}

func loadCPUFromConfig(path string) (*machine.CPU, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	ins, err := bingen.LoadInstructionMemory(cfg.InstrMemPath)
	if err != nil {
		return nil, err
	}
	data, err := bingen.LoadDataMemory(cfg.DataMemPath)
	if err != nil {
		return nil, err
	}

	//TODO: bus
	cpu := machine.New(ins, data, nil)
	return cpu, nil
}
