package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	bingen "github.com/awesoma31/csa-lab4/pkg/bin-gen"
	"github.com/awesoma31/csa-lab4/pkg/machine"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/machine/logger"
	"gopkg.in/yaml.v3"
)

func main() {
	configPath := flag.String("conf", "config/config.yaml", "path to cpu config yml")
	flag.Parse()
	cpu, err := loadCPUFromConfig(*configPath)
	if err != nil {
		log.Fatalf("error configuring CPU - %s", err.Error())
	}
	_ = cpu.Run()
}

func loadCPUFromConfig(cfgPath string) (*machine.CPU, error) {
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var cfg *machine.CpuConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	ioc := io.NewIOController(cfg.Schedule)

	ins, err := bingen.LoadInstructionMemory(cfg.InstrMemPath)
	if err != nil {
		return nil, err
	}
	data, err := bingen.LoadDataMemory(cfg.DataMemPath)
	if err != nil {
		return nil, err
	}

	lg := logger.New(cfg.Debug, cfg.LogFilePath)

	cfg.MemD = data
	cfg.MemI = ins
	cfg.IOC = ioc
	cfg.Logger = lg

	cpu := machine.New(cfg)
	return cpu, nil
}
