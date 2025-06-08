package machine

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"slices"

	"github.com/awesoma31/csa-lab4/pkg/machine/decoder"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

const (
	StackSize = 0x100
)

var (
	StackStart uint32 = 0x100
)

type CPU struct {
	memI []uint32
	memD []byte

	Ioc *io.IOController

	Reg struct {
		GPR      [14]uint32
		PC, IR   uint32
		savedGPR [14]uint32
	}
	N, Z, V, C bool
	IF         bool

	step      microStep // current micro-routine
	inISR     bool
	pending   bool // slot for deferred IRQ
	pendNum   int
	Tick      int
	TickLimit int
	halted    bool
}

func New(memI []uint32, memD []byte, ioc *io.IOController, tickLimit int) *CPU {
	c := &CPU{
		memI:      slices.Clone(memI),
		memD:      slices.Clone(memD),
		Ioc:       ioc,
		TickLimit: tickLimit,
		halted:    false,
	}

	if StackStart < uint32(len(c.memD)) {
		StackStart = uint32(len(c.memD) + StackSize)
	}
	c.Reg.GPR[isa.SpReg] = StackStart

	c.step = c.fetch()
	return c
}

// Run – execute at most nTicks; stop earlier on HALT
func (c *CPU) Run() {
	for c.Tick = 0; c.Tick < c.TickLimit; c.Tick++ {
		if c.halted {
			break
		}
		//TODO: ─── devices update + IRQ sampling (always!) ─────────

		if gotIrq, irqNumber := c.Ioc.CheckTick(c.Tick); gotIrq {
			c.raiseIRQ(irqNumber)
		}

		finished := c.step(c)

		if finished {
			if c.pending && !c.inISR {
				c.enterISR()
			}

			c.step = c.fetch()
		}
	}
	c.PrintAllPortOutputs()
}

func (c *CPU) PrintAllPortOutputs() {
	fmt.Println("───── Port Outputs ─────")
	for port, buf := range c.Ioc.OutBufAll() {
		if len(buf) == 0 {
			continue
		}

		var strBuilder strings.Builder
		for _, b := range buf {
			if b >= 32 && b <= 126 {
				strBuilder.WriteByte(b)
			} else {
				strBuilder.WriteString("*") // непечатаемые → точка
			}
		}
		strOutput := strBuilder.String()

		// числовое представление
		var hexVals []string
		for _, b := range buf {
			hexVals = append(hexVals, fmt.Sprintf("%2X", b))
		}
		hexOutput := strings.Join(hexVals, " ")

		var byteVals []string
		for _, b := range buf {
			byteVals = append(byteVals, fmt.Sprintf("%3d", b))
		}
		byteOutput := strings.Join(byteVals, " ")

		// финальный вывод
		fmt.Printf("port %d: %s\n", port, strOutput)
		fmt.Printf("        %s\n", hexOutput)
		fmt.Printf("        %s\n", byteOutput)
	}
}

func (c *CPU) fetch() microStep {
	return func(c *CPU) bool {

		// if !c.inISR && c.pending
		//go check if interruption request at this tick or irq

		c.Reg.IR = c.memI[c.Reg.PC]
		c.Reg.PC++
		op, mode, rd, rs1, rs2 := decoder.Dec(c.Reg.IR)

		f := ucode[op][mode]
		fmt.Printf("TICK %d @ 0x%08X -  %v %v; PC++ | %v\n", c.Tick, c.Reg.IR, isa.GetOpMnemonic(op), isa.GetAMnemonic(mode), c.ReprPC())
		if f == nil {
			slog.Error("unknown instruction", "PC", c.Reg.PC-1, "IR", c.Reg.IR)
			log.Fatal()
			return false
		}
		c.step = f(rd, rs1, rs2)
		return false
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
	// return 0
}

func (c *CPU) raiseIRQ(vec uint8) {
	if c.inISR || c.pending {
		fmt.Printf("interruption ignored, either in one or one is already pending, %v\n", vec)
		return
	}
	c.pending, c.pendNum = true, int(vec)
}
func (c *CPU) enterISR() {
	fmt.Printf("Entering Interruption, nmb=%d", c.pendNum)
	c.push(c.Reg.PC)             // адрес возврата
	c.Reg.PC = uint32(c.pendNum) // (пока) вектор = номер
	c.SaveGPRValues()
	c.inISR = true
	c.pending = false
}

func (c *CPU) SaveGPRValues() {
	for i := range len(c.Reg.GPR) {
		c.Reg.savedGPR[i] = c.Reg.GPR[i]
	}
}
func (c *CPU) RestoreGPRValues() {
	for i := range len(c.Reg.savedGPR) {
		c.Reg.GPR[i] = c.Reg.savedGPR[i]
	}
}

// --------------------------------------------------------------------
// The ISR must execute IRET (RETurn) → here we map OpRet+NoOperands as IRET
// --------------------------------------------------------------------
func init() {
	ucode[isa.OpRet][isa.NoOperands] = func(_, _, _ int) microStep {
		return func(c *CPU) bool {
			c.Reg.PC = c.pop()
			c.inISR = false
			return true
		}
	}
}

func (c *CPU) Repr() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "tick=%-6d PC=%02X IR=%08X\n",
		c.Tick, c.Reg.PC, c.Reg.IR)

	c.ReprFlags()

	for i, v := range c.Reg.GPR {
		_, _ = fmt.Fprintf(&b, " R%-2d:%08X", i, v)
		if (i+1)%4 == 0 {
			b.WriteByte('\n')
		}
	}
	if len(c.Reg.GPR)%4 != 0 {
		b.WriteByte('\n')
	}
	return b.String()
}

func (c *CPU) ReprPC() string {
	return fmt.Sprintf("PC=%d/0x%X", c.Reg.PC, c.Reg.PC)
}

func (c *CPU) ReprFlags() string {
	boolToIntStr := func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	}

	return fmt.Sprintf("N=%s,Z=%s,V=%s,C=%s",
		boolToIntStr(c.N),
		boolToIntStr(c.Z),
		boolToIntStr(c.V),
		boolToIntStr(c.C),
	)
}

func (c *CPU) DumpState(stage string) {
	fmt.Printf("TICK %d\n", c.Tick)
	fmt.Println("────────────────────────────────────────────")
	fmt.Printf("[INSTR] PC=0x%02X  IR=0x%08X  %s\n",
		c.Reg.PC, c.Reg.IR, Disasm(c.Reg.IR))
	fmt.Printf("[μSTEP] %s\n", stage)
	fmt.Println("[REGS ]")
	fmt.Printf(" PC = 0x%02X  SP = 0x%02X  IR = 0x%08X\n", c.Reg.PC, c.Reg.GPR[isa.SpReg], c.Reg.IR)
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

func (c *CPU) ReprRegVal(r int) any {
	return fmt.Sprintf("%v=%d/0x%X", isa.GetRegMnem(r), c.Reg.GPR[r], c.Reg.GPR[r])
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
