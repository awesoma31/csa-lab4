package main

type ControlUnit interface {
	ProcessNextTick()
}

type controlUnit struct {
	// TODO: program = slice of commands
	Program         string
	Program_counter int
	DataPath        *dataPath

	//"Текущее модельное время процессора (в тактах). Инициализируется нулём."
	tick int
}

func (c *controlUnit) ProcessNextTick() {
	panic("unimplemented")
}

func NewControlUnit() ControlUnit {
	return &controlUnit{}
}
