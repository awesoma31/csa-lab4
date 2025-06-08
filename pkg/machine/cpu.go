package machine

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/machine/decoder"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
	"slices"
)

var (
	StackStart uint32 = 0x0300
)

type CPU struct {
	memI []uint32
	memD []byte
	io   *IOBus

	reg struct {
		GPR    [16]uint32
		PC, IR uint32
	}
	tmp byte

	// control
	step    microStep // current micro-routine
	inISR   bool      // already inside ISR?
	pending bool      // slot for deferred IRQ
	pendNum int
	tick    uint64
}

func (c *CPU) ReprRegVal(r int) any {
	return fmt.Sprintf("%v=%d/0x%X", isa.GetRegMnem(r), c.reg.GPR[r], c.reg.GPR[r])
}

// vectors fixed at word-addresses 0x80, 0x84 …
var vectors = [2]uint32{0x80, 0x84}

//	func New(memI []uint32, memD []byte, bus *IOBus) *CPU {
//		//TODO: mem capacity
//		mI := make([]uint32, 0, 1000)
//		mD := make([]byte, 0, 1000)
//		mI = append(mI, memI...)
//		mD = append(mD, memD...)
//		c := &CPU{memI: mI, memD: mD, io: bus}
//		c.step = c.fetch() // start with fetch
//		return c
//	}
func New(memI []uint32, memD []byte, bus *IOBus) *CPU {
	c := &CPU{
		memI: slices.Clone(memI),
		memD: slices.Clone(memD),
		io:   bus,
	}

	// Стек растёт вниз, поэтому SP ставим сразу «ниже» данных.
	if StackStart < uint32(len(c.memD)) {
		StackStart = uint32(len(c.memD) + 0x100)
	}
	c.reg.GPR[isa.SpReg] = StackStart

	c.step = c.fetch()
	return c
}

func (c *CPU) fetch() microStep { // returns a μ-routine
	return func(c *CPU) bool {
		c.reg.IR = c.memI[c.reg.PC]
		c.reg.PC++
		op, mode, rd, rs1, rs2 := decoder.Dec(c.reg.IR)

		f := ucode[op][mode]
		// if op == isa.OpPush {
		// 	fmt.Printf("decoded push, reg n %d=%v\n", rs1, isa.GetRegMnem(rs1))
		// }
		fmt.Printf("TICK %d - 0x%08X -  %v %v; PC++ | %v\n", c.tick, c.reg.IR, isa.GetOpMnemonic(op), isa.GetAMnemonic(mode), c.ReprPC())
		if f == nil {
			slog.Warn("unknown instruction", "PC", c.reg.PC-1, "IR", c.reg.IR)
			return false // ← fetch в следующем такте
		}
		c.step = f(rd, rs1, rs2)
		return false
	}
}

// Run – execute at most nTicks; stop earlier on HALT
func (c *CPU) Run(nTicks uint64) {
	for c.tick = 0; c.tick < nTicks; c.tick++ {
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
	panic("unimpl")
	// c.reg.GPR[SpReg]--
	// c.memD[c.reg.GPR[SpReg]] = v
}
func (c *CPU) pop() uint32 {
	panic("unimpl")
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

func (c *CPU) ReprPC() string {
	return fmt.Sprintf("PC=%X", c.reg.PC)
}

func (c *CPU) DumpState(stage string) {
	fmt.Printf("TICK %d\n", c.tick)
	fmt.Println("────────────────────────────────────────────")
	fmt.Printf("[INSTR] PC=0x%02X  IR=0x%08X  %s\n",
		c.reg.PC, c.reg.IR, Disasm(c.reg.IR))
	fmt.Printf("[μSTEP] %s\n", stage)
	fmt.Println("[REGS ]")
	fmt.Printf(" PC = 0x%02X  SP = 0x%02X  IR = 0x%08X\n", c.reg.PC, c.reg.GPR[isa.SpReg], c.reg.IR)
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

func (c *CPU) ensureDataSize(last uint32) {
	if last < uint32(len(c.memD)) {
		return
	}
	need := last - uint32(len(c.memD)) + 1
	c.memD = append(c.memD, make([]byte, need)...)
}
