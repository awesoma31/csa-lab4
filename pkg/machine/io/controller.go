package io

import (
	"fmt"
)

type IOController struct {
	Sched    map[int]Input // {tick:(event)}
	portsVal map[uint8]byte
	Output   []byte
}

func NewIOController(entries []TickEntry) *IOController {
	m := make(map[int]Input, len(entries))
	for _, e := range entries {
		m[e.Tick] = e.Input
	}
	return &IOController{
		Sched:    m,
		portsVal: make(map[uint8]byte),
	}
}

func (ioc *IOController) CheckTick(tick int) (irq bool, IrqNumber uint8) {
	inp, ok := ioc.Sched[tick]
	if !ok {
		return false, 0
	}
	port := uint8(inp.IntrNumber)
	ioc.portsVal[port] = toByte(inp.Value)
	delete(ioc.Sched, tick)
	return true, port
}

func (ioc *IOController) ReadPort(p uint8) byte {
	fmt.Printf("Reading %v from port %d\n", ioc.portsVal[p], p)
	return ioc.portsVal[p]
}
func (ioc *IOController) WritePort(p uint8, v byte) {
	// fmt.Printf("Writing %v to port %d | %v\n", v, p, ioc.Output)
	ioc.Output = append(ioc.Output, v)
}

func toByte(v any) byte {
	switch t := v.(type) {
	case int:
		return byte(t)
	case int64:
		return byte(t)
	case uint32:
		return byte(t)
	case string:
		if len(t) > 0 {
			return t[0]
		}
	case byte:
		return t
	}
	return 0
}
