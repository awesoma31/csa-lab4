package io

type TickEntry struct {
	Tick  int   `yaml:"tick"`
	Input Input `yaml:"input"`
}

type Input struct {
	IrqNumber int `yaml:"interrupt"`
	Value     any `yaml:"value"`
}
