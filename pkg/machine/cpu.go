package machine

import (
	"fmt"
	"log/slog"

	"github.com/awesoma31/csa-lab4/pkg/machine/decoder"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type CPU struct {
	memI []uint32 // instruction memory (word-addressed)
	memD []byte   // data memory
	io   *IOBus

	// registers & latches
	reg struct {
		GPR [16]uint32
		PC  uint32
		IR  uint32
	} // + tmp for micro-code
	tmp byte

	// control
	step    microStep // current micro-routine
	inISR   bool      // already inside ISR?
	pending bool      // slot for deferred IRQ
	pendNum int
	tick    uint64
}

// vectors fixed at word-addresses 0x80, 0x84 …
var vectors = [2]uint32{0x80, 0x84}

func New(memI []uint32, memD []byte, bus *IOBus) *CPU {
	c := &CPU{memI: memI, memD: memD, io: bus}
	c.step = c.fetch() // start with fetch
	return c
}

func (c *CPU) fetch() microStep { // returns a μ-routine
	return func(c *CPU) bool {
		c.reg.IR = c.memI[c.reg.PC]
		op, mode, rd, rs1, rs2 := decoder.Dec(c.reg.IR)

		fmt.Printf("PC=%02X IR=%08X op=%s mode=%s rd=%d\n",
			c.reg.PC, c.reg.IR,
			isa.GetMnemonic(op), isa.GetAMnemonic(mode), rd)

		f := ucode[op][mode]
		if f == nil { // ─ unknown !
			fmt.Printf("[WARN  tick=%d] unknown opcode 0x%v mode 0x%v "+
				"@ PC 0x%X = 0x%08X – skipped\n",
				c.tick, isa.GetMnemonic(op), isa.GetAMnemonic(mode), c.reg.PC, c.reg.IR)

			c.reg.PC++  // just step over word
			return true // “macro-inst finished”
		}
		c.step = f(rd, rs1, rs2)
		return false
	}
}

// Run – execute at most nTicks; stop earlier on HALT
func (c *CPU) Run(nTicks uint64) {
	for c.tick = 0; c.tick < nTicks; c.tick++ {
		fmt.Printf("tick %d ", c.tick)

		//TODO: ─── devices update + IRQ sampling (always!) ─────────
		// for _, d := range c.io.Devs {
		// 	if d.step(c.tick) && !c.inISR && !c.pending {
		// 		c.pending, c.pendNum = true, d.IrqNum
		// 	}
		// }

		// ─── execute ONE micro-step ──────────────────────────
		didFinish := c.step(c)

		// ─── after macro-instruction finished … ──────────────
		if didFinish {
			if c.pending && !c.inISR {
				// enter ISR
				slog.Info("Entering Interruption")
				c.pending = false
				vec := vectors[c.pendNum]
				c.push(c.reg.PC + 1) // return address
				c.reg.PC = vec
				c.inISR = true
			}
			c.step = c.fetch()
		}
	}
}

func (c *CPU) push(v uint32) {
	// c.reg.GPR[SpReg]--
	// c.memD[c.reg.GPR[SpReg]] = v
}
func (c *CPU) pop() uint32 {
	// v := c.memD[c.reg.GPR[SpReg]]
	// c.reg.GPR[SpReg]++
	// return v
	return 0
}

// --------------------------------------------------------------------
// The ISR must execute IRET (RETurn) → here we map OpRet+NoOperands as IRET
// --------------------------------------------------------------------
func init() {
	ucode[isa.OpRet][isa.NoOperands] = func(_, _, _ int) microStep {
		return func(c *CPU) bool {
			c.reg.PC = c.pop()
			c.inISR = false
			return true
		}
	}
}

// import "fmt"
//
// type CPU struct {
// 	CU *ControlUnit
// 	DP *DataPath
// }
//
// // TODO: from cfg
// const (
// 	InstrMemSize          = 100
// 	StackStartAddr uint32 = DataMemSize
// 	DataMemSize           = 200
// )
//
// func (m *CPU) Start() {
// 	for m.CU.CurrentTick() < m.CU.TickLimit {
// 		fmt.Println("tick", m.CU.tick)
// 		m.CU.proccessNextTick()
// 	}
// 	// m.CU.Simulate()
// }
