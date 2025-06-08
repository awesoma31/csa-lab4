package io

import "fmt"

type IOController struct {
	Sched    map[int]Input    // расписание IRQ
	portsVal map[uint8]byte   // “регистры”-входы (последний принятый байт)
	outBuf   map[uint8][]byte // ***отдельный буфер на каждый порт***
}

func (ioc *IOController) OutBufAll() map[uint8][]byte {
	return ioc.outBuf
}

func NewIOController(entries []TickEntry) *IOController {
	m := make(map[int]Input, len(entries))
	for _, e := range entries {
		m[e.Tick] = e.Input
	}
	return &IOController{
		Sched:    m,
		portsVal: make(map[uint8]byte),
		outBuf:   make(map[uint8][]byte),
	}
}

/* ——— вход (IRQ) остаётся как был ——— */

func (ioc *IOController) CheckTick(tick int) (bool, uint8) {
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
	return ioc.portsVal[p]
}

func (ioc *IOController) WritePort(p uint8, v byte) {
	ioc.outBuf[p] = append(ioc.outBuf[p], v)
}

func (ioc *IOController) Output(port uint8) []byte {
	return ioc.outBuf[port]
}

func (ioc *IOController) OutputAll() []byte {
	var all []byte
	for _, b := range ioc.outBuf {
		all = append(all, b...)
	}
	return all
}

func toByte(v any) byte {
	switch t := v.(type) {
	case int, int8, int16, int32, int64:
		return byte(fmt.Sprint(t)[0])
	case uint, uint16, uint32, uint64:
		return byte(fmt.Sprint(t)[0])
	case string:
		if len(t) > 0 {
			return t[0]
		}
	case byte:
		return t
	}
	return 0
}
