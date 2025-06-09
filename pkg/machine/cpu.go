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

	Ioc *io.Controller

	Reg struct {
		GPR      [14]uint32
		PC, IR   uint32
		savedGPR [14]uint32
		savedPC  uint32
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
	maxInt    int
}

func New(memI []uint32, memD []byte, ioc *io.Controller, mi int, tickLimit int) *CPU {
	c := &CPU{
		memI:      slices.Clone(memI),
		memD:      slices.Clone(memD),
		Ioc:       ioc,
		TickLimit: tickLimit,
		halted:    false,
		maxInt:    mi,
	}

	if StackStart < uint32(len(c.memD)) {
		StackStart = uint32(len(c.memD) + StackSize)
	}
	c.Reg.GPR[isa.SpReg] = StackStart
	c.Reg.PC = uint32(mi) // skip vector table addreses go to prog

	c.step = c.fetch()
	return c
}

// Run – execute at most nTicks; stop earlier on HALT
func (c *CPU) Run() {
	for c.Tick = 0; c.Tick < c.TickLimit; c.Tick++ {
		if c.halted {
			break
		}

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
			if b >= 32 && b <= 126 { // Диапазон печатаемых ASCII-символов
				strBuilder.WriteByte(b)
			} else {
				strBuilder.WriteString("*") // Непечатаемые -> '*'
			}
		}
		strOutput := strBuilder.String()

		var hexVals []string
		for _, b := range buf {
			hexVals = append(hexVals, fmt.Sprintf("%X", b))
		}
		hexOutput := strings.Join(hexVals, ", ")

		var byteVals []string
		for _, b := range buf {
			byteVals = append(byteVals, fmt.Sprintf("%d", b))
		}
		byteOutput := strings.Join(byteVals, ", ")

		fmt.Printf("port % 2d| %s\n", port, strOutput)
		fmt.Printf("       |%s\n", hexOutput)
		fmt.Printf("       |%s\n", byteOutput)
	}
}

func (c *CPU) fetch() microStep {
	return func(c *CPU) bool {
		c.Reg.IR = c.memI[c.Reg.PC]
		c.Reg.PC++
		op, mode, rd, rs1, rs2 := decoder.Dec(c.Reg.IR)

		f := ucode[op][mode]
		fmt.Printf("TICK % 4d @ 0x%08X -  %v %v; PC++ | %v\n", c.Tick, c.Reg.IR, isa.GetOpMnemonic(op), isa.GetAMnemonic(mode), c.ReprPC())
		if f == nil {
			slog.Error("unknown instruction", "PC", c.Reg.PC-1, "IR", c.Reg.IR)
			log.Fatal()
			return false
		}
		c.step = f(rd, rs1, rs2)
		return false
	}
}

func (c *CPU) raiseIRQ(vec uint8) {
	if c.inISR || c.pending {
		fmt.Printf("interruption ignored, either in one or one is already pending, %v\n", vec)
		return
	}
	c.pending, c.pendNum = true, int(vec)
}
func (c *CPU) enterISR() {
	fmt.Printf("------------Entering Interruption %d------------\n", c.pendNum)
	c.Reg.savedPC = c.Reg.PC
	//TODO: save flags
	c.Reg.PC = c.memI[c.pendNum] // (пока) вектор = номер
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

func (c *CPU) ReprPC() string {
	return fmt.Sprintf("PC=% 4d/0x%X", c.Reg.PC, c.Reg.PC)
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

func (c *CPU) ReprRegVal(r int) any {
	return fmt.Sprintf("%v=% 4d/0x%X", isa.GetRegMnem(r), c.Reg.GPR[r], c.Reg.GPR[r])
}

func (c *CPU) ensureDataSize(last uint32) {
	if last < uint32(len(c.memD)) {
		return
	}
	need := last - uint32(len(c.memD)) + 1
	c.memD = append(c.memD, make([]byte, need)...)
}
