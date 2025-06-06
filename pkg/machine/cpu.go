package machine

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/machine/decoder"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type CPU struct {
	memI []uint32
	memD []byte
	io   *IOBus

	reg struct {
		GPR        [16]uint32
		PC, IR, SP uint32
	}
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

		f := ucode[op][mode]
		fmt.Printf("TICK %d - 0x%08X -  %v, mode %v \n", c.tick, c.reg.IR, isa.GetOpMnemonic(op), isa.GetAMnemonic(mode))
		// if f == nil { // ─ unknown !
		// 	fmt.Printf("[WARN  tick=%d] unknown opcode 0x%v mode 0x%v "+
		// 		"@ PC 0x%X = 0x%08X – skipped\n",
		// 		c.tick, isa.GetMnemonic(op), isa.GetAMnemonic(mode), c.reg.PC, c.reg.IR)
		//
		// 	c.reg.PC++  // just step over word
		// 	return true // “macro-inst finished”
		// }
		if f == nil {
			slog.Warn("unknown instruction", "pc", c.reg.PC, "ir", c.reg.IR)
			c.reg.PC++   // пропускаем слово
			return false // ← fetch в следующем такте
		}
		c.step = f(rd, rs1, rs2)
		return false
	}
}

// Run – execute at most nTicks; stop earlier on HALT
func (c *CPU) Run(nTicks uint64) {
	for c.tick = 0; c.tick < nTicks; c.tick++ {
		// fmt.Printf("tick %d - ", c.tick)
		// fmt.Print(c.Repr())
		// c.DumpState("before microstep")

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

//	func (c *CPU) Repr() string {
//		s := fmt.Sprintf("tick=%d PC=0x%02X IR=0x%08X\n", c.tick, c.reg.PC, c.reg.IR)
//		for i := range len(c.reg.GPR) {
//			s += fmt.Sprintf(" R%-2d: 0x%08X\n", i, c.reg.GPR[i])
//		}
//		return s
//	}
func (c *CPU) Repr() string {
	var b strings.Builder
	fmt.Fprintf(&b, "tick=%-6d PC=%02X IR=%08X\n",
		c.tick, c.reg.PC, c.reg.IR)

	// флаги, если будут
	// fmt.Fprintf(&b, "NZVC=%d%d%d%d  inISR=%v  pending=%v\n",
	//    c.reg.Z, c.reg.N, c.reg.V, c.reg.C, c.inISR, c.pending)

	for i, v := range c.reg.GPR {
		fmt.Fprintf(&b, " R%-2d:%08X", i, v)
		if (i+1)%4 == 0 {
			b.WriteByte('\n')
		}
	}
	// if len(c.reg.GPR)%4 != 0 {
	// 	b.WriteByte('\n')
	// }
	return b.String()
}

func (c *CPU) DumpState(stage string) {
	fmt.Printf("TICK %d\n", c.tick)
	fmt.Println("────────────────────────────────────────────")
	fmt.Printf("[INSTR] PC=0x%02X  IR=0x%08X  %s\n",
		c.reg.PC, c.reg.IR, Disasm(c.reg.IR))
	fmt.Printf("[μSTEP] %s\n", stage)
	fmt.Println("[REGS ]")
	fmt.Printf(" PC = 0x%02X  SP = 0x%02X  IR = 0x%08X\n", c.reg.PC, c.reg.SP, c.reg.IR)
	// for i := 0; i < 16; i += 4 {
	// 	for j := 0; j < 4; j++ {
	// 		fmt.Printf(" R%-2d= 0x%08X  ", i+j, c.reg.GPR[i+j])
	// 	}
	// 	fmt.Println()
	// }
	// fmt.Println("[IO    ]")
	// for i, dev := range c.io.Devs {
	// 	fmt.Printf(" PORT %d IN = %q  OUT = %q\n", i, dev.InBuf, dev.OutBuf)
	// }
	fmt.Printf(" INTERRUPT PENDING: %v  IN_ISR: %v\n", c.pending, c.inISR)
	fmt.Println("────────────────────────────────────────────")
}

func Disasm(ir uint32) string {
	op, mode, rd, rs1, _ := decoder.Dec(ir)
	switch {
	case op == isa.OpMov && mode == isa.MvImmReg:
		return fmt.Sprintf("MOV r%d<-#0x%X", rd, ir&0xFFFF)
	case op == isa.OpMov && mode == isa.MvRegReg:
		return fmt.Sprintf("MOV r%d<-r%d", rd, rs1)
	case op == isa.OpMov && mode == isa.MvMemReg:
		return fmt.Sprintf("MOV r%d<-[addr]", rd)
	// ...
	default:
		return fmt.Sprintf("??? op=%d mode=%d", op, mode)
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
