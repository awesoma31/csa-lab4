package io

import (
	"reflect"
)

type Controller struct {
	Sched    map[int]Input
	portsVal map[uint8]byte
	outBuf   map[uint8][]byte
}

func (ioc *Controller) OutBufAll() map[uint8][]byte {
	return ioc.outBuf
}

func NewIOController(entries []TickEntry) *Controller {
	m := make(map[int]Input, len(entries))
	for _, e := range entries {
		m[e.Tick] = e.Input
	}
	return &Controller{
		Sched:    m,
		portsVal: make(map[uint8]byte),
		outBuf:   make(map[uint8][]byte),
	}
}

func (ioc *Controller) CheckTick(tick int) (bool, uint8) {
	inp, ok := ioc.Sched[tick]
	if !ok {
		return false, 0
	}

	port := byte(inp.IrqNumber)
	ioc.portsVal[port] = toByte(inp.Value)
	delete(ioc.Sched, tick)
	return true, port
}

func (ioc *Controller) ReadPort(p uint8) byte {
	//TODO: отрицательные числа
	return ioc.portsVal[p]
}

func (ioc *Controller) WritePort(p uint8, v byte) {
	ioc.outBuf[p] = append(ioc.outBuf[p], v)
}

func (ioc *Controller) Output(port uint8) []byte {
	return ioc.outBuf[port]
}

func (ioc *Controller) OutputAll() []byte {
	var all []byte
	for _, b := range ioc.outBuf {
		all = append(all, b...)
	}
	return all
}

func toByte(v any) byte {
	switch t := v.(type) {
	case int, int8, int16, int32, int64:
		return byte(reflect.ValueOf(t).Int() & 0xFF)
	case uint, uint16, uint32, uint64:
		return byte(reflect.ValueOf(t).Uint() & 0xFF)
	case string:
		if len(t) > 0 {
			return t[0]
		}
	case byte:
		return t
	}
	return 0
}
