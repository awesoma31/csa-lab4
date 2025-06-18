package io

import (
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
	"reflect"
)

type Controller struct {
	Sched    map[int]Input
	portsVal map[isa.Register]uint32
	outBuf   map[uint8][]uint32
}

func (ioc *Controller) OutBufAll() map[uint8][]uint32 {
	return ioc.outBuf
}

func NewIOController(entries []TickEntry) *Controller {
	m := make(map[int]Input, len(entries))
	for _, e := range entries {
		m[e.Tick] = e.Input
	}
	return &Controller{
		Sched:    m,
		portsVal: make(map[isa.Register]uint32),
		outBuf:   make(map[uint8][]uint32),
	}
}

func (ioc *Controller) CheckTick(tick int) (bool, isa.Register) {
	inp, ok := ioc.Sched[tick]
	if !ok {
		return false, 0
	}

	port := isa.Register(inp.IrqNumber)
	ioc.portsVal[port] = toUint32(inp.Value)
	delete(ioc.Sched, tick)
	return true, port
}

func (ioc *Controller) ReadPort(p isa.Register) uint32 {
	return ioc.portsVal[p]
}

func (ioc *Controller) WritePort(p uint8, v uint32) {
	ioc.outBuf[p] = append(ioc.outBuf[p], v)
}

func (ioc *Controller) Output(port uint8) []uint32 {
	return ioc.outBuf[port]
}

func (ioc *Controller) OutputAll() []uint32 {
	var all []uint32
	for _, b := range ioc.outBuf {
		all = append(all, b...)
	}
	return all
}

func toUint32(v any) uint32 {
	switch t := v.(type) {
	case int, int8, int16, int32, int64:
		return uint32(reflect.ValueOf(t).Int())
	case uint, uint16, uint32, uint64:
		return uint32(reflect.ValueOf(t).Uint())
	case string:
		if len(t) > 0 {
			return uint32(t[0]) // 1 char
		}
	case byte:
		return uint32(t)
	}
	return 0
}
