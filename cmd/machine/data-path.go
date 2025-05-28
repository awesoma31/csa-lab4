package main

type DataPath interface {
	Zero() int
}

type dataPath struct {
}

func (dp *dataPath) Zero() int {
	// TODO:
	return 0
}

func NewDataPath() DataPath {
	return &dataPath{}
}
