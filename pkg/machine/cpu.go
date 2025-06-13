package machine

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/machine/decoder"
	"github.com/awesoma31/csa-lab4/pkg/machine/io"
	"github.com/awesoma31/csa-lab4/pkg/machine/logger"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

const (
	StackSize = 0x100
)

var (
	StackStart uint32 = 0x100
)

type CpuConfig struct {
	InstrMemPath     string         `yaml:"instruction_bin"`
	DataMemPath      string         `yaml:"data_bin"`
	TickLimit        int            `yaml:"tick_limit"`
	Schedule         []io.TickEntry `yaml:"schedule"`
	MaxInterruptions int            `yaml:"max_interruptions"`
	Debug            bool           `yaml:"debug"`
	LogFilePath      string         `yaml:"log_file"`

	IOC  *io.Controller
	MemI []uint32
	MemD []byte

	Logger *logger.Logger
}

type CPU struct {
	memI []uint32
	memD []byte

	Ioc *io.Controller

	Reg struct {
		GPR      [16]uint32
		PC, IR   uint32
		savedGPR [16]uint32
		savedPC  uint32
	}
	N, Z, V, C     bool
	savedNZVCFlags uint8
	IsIntOn        bool

	step      microStep // current micro-routine
	inISR     bool
	pending   bool
	pendNum   int
	Tick      int
	TickLimit int
	halted    bool
	maxInt    int

	log *logger.Logger
}

func New(cfg *CpuConfig) *CPU {
	c := &CPU{
		memI:      cfg.MemI,
		memD:      cfg.MemD,
		Ioc:       cfg.IOC,
		TickLimit: cfg.TickLimit,
		halted:    false,
		maxInt:    cfg.MaxInterruptions,
		IsIntOn:   true,
		log:       cfg.Logger,
	}

	//TODO: stack = last sata addr + stack size
	if StackStart < uint32(len(c.memD)) {
		StackStart = uint32(len(c.memD) + StackSize)
	}
	c.Reg.GPR[isa.SpReg] = StackStart
	c.Reg.PC = uint32(c.maxInt)

	c.step = c.fetch()
	return c
}

func (c *CPU) Run() string {
	for c.Tick = 0; c.Tick < c.TickLimit; c.Tick++ {
		if c.halted {
			break
		}

		if gotIrq, irqNumber := c.Ioc.CheckTick(c.Tick); gotIrq {
			c.raiseIRQ(irqNumber)
		}

		finished := c.step(c)

		if finished {
			if c.pending && !c.inISR && c.IsIntOn {
				c.enterISR()
			}

			c.step = c.fetch()
		}
	}

	// c.PrintAllPortOutputs()
	fmt.Println("ticks", c.Tick)
	fmt.Println(c.GetFormattedPortOutputs())
	return c.GetFormattedPortOutputs()
}

func (c *CPU) GetFormattedPortOutputs() string {
	var sb strings.Builder

	for port, buf := range c.Ioc.OutBufAll() {
		if len(buf) == 0 {
			continue
		}

		switch port {
		case isa.PortCh:
			sb.WriteString(isa.GetPortMnem(isa.PortCh))
			sb.WriteString("| ")
			for _, w := range buf {
				ch := byte(w)
				if ch >= 32 && ch <= 126 {
					sb.WriteByte(ch)
				} else {
					sb.WriteByte('*')
				}
			}
			sb.WriteByte('\n')

		case isa.PortD:
			sb.WriteString(isa.GetPortMnem(isa.PortD))
			sb.WriteString("| ")
			for i, w := range buf {
				if i > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString(strconv.FormatInt(int64(int32(w)), 10))
			}
			sb.WriteByte('\n')

		case isa.PortL:
			sb.WriteString(isa.GetPortMnem(isa.PortL))
			sb.WriteString("| ")
			if len(buf)%2 != 0 {
				sb.WriteString("(warn: odd words) ")
			}
			for i := 0; i+1 < len(buf); i += 2 {
				lo := uint64(buf[i])
				hi := uint64(buf[i+1])
				val := int64(hi<<32 | lo) // sign-extend
				if i > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString(strconv.FormatInt(val, 10))
			}
			sb.WriteByte('\n')

		default:
			sb.WriteString(fmt.Sprintf("Port %d| ", port))
			for _, w := range buf {
				sb.WriteString(fmt.Sprintf("0x%X ", w))
			}
			sb.WriteByte('\n')
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (c *CPU) fetch() microStep {
	return func(c *CPU) bool {
		c.Reg.IR = c.memI[c.Reg.PC]
		c.Reg.PC++
		op, mode, rd, rs1, rs2 := decoder.Dec(c.Reg.IR)

		f := ucode[op][mode]
		c.log.Debug(
			fmt.Sprintf("TICK % 4d @ 0x%08X -  %v %v; PC++ | %v\n", c.Tick, c.Reg.IR, isa.GetOpMnemonic(op), isa.GetAMnemonic(mode), c.ReprPC()),
		)
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
		c.log.Debugf("interruption ignored, either in one or one is already pending, %v\n", vec)
		return
	}
	c.pending, c.pendNum = true, int(vec)
}
func (c *CPU) enterISR() {
	// c.log.Debug(c.DumpState())
	c.log.Debugf("------------Entering Interruption %d, value=%v/0x%X------------\n", c.pendNum, c.Ioc.ReadPort(byte(c.pendNum)), c.Ioc.ReadPort(byte(c.pendNum)))
	c.Reg.savedPC = c.Reg.PC
	c.SaveNZVC()
	c.Reg.PC = c.memI[c.pendNum]
	c.SaveGPRValues()
	c.inISR = true
	c.pending = false
	// c.log.Debug(c.DumpState())

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

func (c *CPU) SaveNZVC() {
	c.savedNZVCFlags = 0
	if c.N {
		c.savedNZVCFlags |= (1 << 0) // Set bit 0 for N
	}
	if c.Z {
		c.savedNZVCFlags |= (1 << 1) // Set bit 1 for Z
	}
	if c.V {
		c.savedNZVCFlags |= (1 << 2) // Set bit 2 for V
	}
	if c.C {
		c.savedNZVCFlags |= (1 << 3) // Set bit 3 for C
	}
}

func (c *CPU) RestoreNZVC() {
	c.N = (c.savedNZVCFlags&(1<<0) != 0)
	c.Z = (c.savedNZVCFlags&(1<<1) != 0)
	c.V = (c.savedNZVCFlags&(1<<2) != 0)
	c.C = (c.savedNZVCFlags&(1<<3) != 0)
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

func (c *CPU) ReprRegVal(r int) any {
	return fmt.Sprintf("%v=%d/0x%X", isa.GetRegMnem(r), c.Reg.GPR[r], c.Reg.GPR[r])
}

func (c *CPU) ensureDataSize(last uint32) {
	if last < uint32(len(c.memD)) {
		return
	}
	need := last - uint32(len(c.memD)) + 1
	c.memD = append(c.memD, make([]byte, need)...)
}

func (c *CPU) DumpState() string {
	var sb strings.Builder
	sb.WriteString("---------- CPU State Dump ----------\n")
	sb.WriteString(fmt.Sprintf("Tick: %d/%d | ", c.Tick, c.TickLimit))
	sb.WriteString(fmt.Sprintf("PC: %d(0x%X) | ", c.Reg.PC, c.Reg.PC))
	sb.WriteString(fmt.Sprintf("IR: %d(0x%X) | ", c.Reg.IR, c.Reg.IR))
	sb.WriteString(fmt.Sprintf("Flags: %s\n", c.ReprFlags()))
	sb.WriteString(fmt.Sprintf("Interrupts On:%t | In ISR:%t | Pending:%t(num: %d)\n", c.IsIntOn, c.inISR, c.pending, c.pendNum))

	for i := range len(c.Reg.GPR) {
		sb.WriteString(fmt.Sprintf("%s: %d (0x%X)\n", isa.GetRegMnem(i), c.Reg.GPR[i], c.Reg.GPR[i]))
	}

	// sb.WriteString("\n--- Stack Pointer (SP) and Stack Content (top 16 bytes) ---\n")
	// spVal := c.Reg.GPR[isa.SpReg] // Assuming SpReg is the index for the stack pointer register
	// sb.WriteString(fmt.Sprintf("SP (R%d): %d (0x%X)\n", isa.SpReg, spVal, spVal))
	//
	// stackDumpStart := max(int(spVal)-8, 0)
	// stackDumpEnd := min(int(spVal)+8, len(c.memD))
	//
	// if stackDumpEnd > stackDumpStart {
	// 	sb.WriteString("Stack (data memory around SP):\n")
	// 	for i := stackDumpStart; i < stackDumpEnd; i++ {
	// 		if i%8 == 0 {
	// 			sb.WriteString(fmt.Sprintf("0x%08X: ", i))
	// 		}
	// 		sb.WriteString(fmt.Sprintf("%02X ", c.memD[i]))
	// 		if (i+1)%8 == 0 || i == stackDumpEnd-1 {
	// 			sb.WriteString("\n")
	// 		}
	// 	}
	// } else {
	// 	sb.WriteString("Stack content not accessible or empty.\n")
	// }

	sb.WriteString("--- Saved Registers (for ISR) ---\n")
	for i := range len(c.Reg.savedGPR) {
		sb.WriteString(fmt.Sprintf("%s: %d (0x%X)\n", isa.GetRegMnem(i), c.Reg.GPR[i], c.Reg.GPR[i]))
	}
	sb.WriteString(fmt.Sprintf("Saved PC: %d (0x%X)\n", c.Reg.savedPC, c.Reg.savedPC))
	sb.WriteString(fmt.Sprintf("Saved Flags: %08b\n", c.savedNZVCFlags))

	sb.WriteString("------------------------------------\n")
	return sb.String()
}
