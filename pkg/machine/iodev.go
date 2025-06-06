package machine

import "sync/atomic"

type Device struct {
	ID        int
	IrqNum    int
	sched     []tickVal // sorted by Tick
	cursor    int
	inValue   atomic.Uint32
	OutBuffer []byte
}

type tickVal struct {
	Tick uint64
	Val  byte
}

func (d *Device) step(tick uint64) (raisedIRQ bool) {
	if d.cursor < len(d.sched) && tick == d.sched[d.cursor].Tick {
		d.inValue.Store(uint32(d.sched[d.cursor].Val))
		d.cursor++
		raisedIRQ = true
	}
	return
}
func (d *Device) Load() uint32 { return d.inValue.Load() }
func (d *Device) Store(b byte) { d.OutBuffer = append(d.OutBuffer, b) }

type IOBus struct{ Devs [2]*Device }

func NewIOBus(port0, port1 []tickVal) *IOBus {
	return &IOBus{
		Devs: [2]*Device{
			{ID: 0, IrqNum: 0, sched: port0},
			{ID: 1, IrqNum: 1, sched: port1},
		},
	}
}
