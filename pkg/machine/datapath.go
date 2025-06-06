package machine

type DataPath struct {
	InstrMem []uint32
	DataMem  []byte
	Regs     *Registers

	RA uint32
}

type Registers struct {
	RA, RM1, RM2 uint32
}

func (d *DataPath) setRegister(destReg int, data uint32) {
	d.RA = data
}
