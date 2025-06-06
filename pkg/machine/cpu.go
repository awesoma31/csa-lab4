package machine

import "fmt"

type Machine struct {
	CU *ControlUnit
	DP *DataPath
}

// TODO: from cfg
const (
	InstrMemSize          = 100
	StackStartAddr uint32 = DataMemSize
	DataMemSize           = 200
)

func (m *Machine) Start() {
	for m.CU.CurrentTick() < m.CU.TickLimit {
		fmt.Println("tick", m.CU.tick)
		m.CU.proccessNextTick()
	}
	// m.CU.Simulate()
}
