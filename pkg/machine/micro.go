package machine

import (
	"fmt"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type microStep func(c *CPU) bool // true → macro-instruction done

const (
	opcodesMaxAmount = 64 // 6 bit opcode
	modesMaxAmount   = 32 // 5 bit
)

// [opcode][mode] → factory(rd,rs1,rs2)→microStep
var ucode [opcodesMaxAmount][modesMaxAmount]func(int, int, int) microStep

func init() {
	// ─── NOP (single tick) ───────────────────────────────
	ucode[isa.OpNop][isa.NoOperands] = uNop
	ucode[isa.OpHalt][isa.NoOperands] = uHalt

	// ─── MOV rd, [memAbs] (2 words, 3 ticks) ─────────────
	ucode[isa.OpMov][isa.MvMemReg] = uMovMemReg
	ucode[isa.OpMov][isa.MvImmReg] = uMovImmReg
	ucode[isa.OpMov][isa.MvRegReg] = uMovRegReg

}

func uNop(_, _, _ int) microStep {
	return func(c *CPU) bool {
		fmt.Printf("TICK %d - NOP, PC++\n", c.tick)
		c.reg.PC++
		return true
	}
}

func uHalt(_ int, _ int, _ int) microStep {
	return func(c *CPU) bool {
		fmt.Println("simultaion stopped")
		os.Exit(0)
		return true
	}
}

//	func uMovMemReg(rd, _, _ int) microStep {
//		stage := 0
//		var b [4]byte
//		return func(c *CPU) bool {
//			switch stage {
//			case 0: // T0 – fetch 2-nd word (адрес)
//				c.reg.PC++
//				c.reg.GPR[isa.RAddr] = c.memI[c.reg.PC]
//				stage++
//			case 1: // T1 – read byte 0
//				b[0] = c.memD[c.reg.GPR[isa.RAddr]]
//				stage++
//			case 2: // T2 – read byte 1
//				b[1] = c.memD[c.reg.GPR[isa.RAddr]+1]
//				stage++
//			case 3: // T3 – read byte 2
//				b[2] = c.memD[c.reg.GPR[isa.RAddr]+2]
//				stage++
//			case 4: // T4 – read byte 3
//				b[3] = c.memD[c.reg.GPR[isa.RAddr]+3]
//				stage++
//			case 5: // T5 – assemble and write back
//				val := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
//				c.reg.GPR[rd] = val
//				c.reg.PC++
//				return true
//			}
//			return false
//		}
//	}
func uMovMemReg(rd, _, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0: // fetch addr
			c.reg.PC++
			c.reg.GPR[isa.RAddr] = c.memI[c.reg.PC]
			stage++
		default:
			var val uint32
			if c.read32(c.reg.GPR[isa.RAddr], &stage, &val) {
				c.reg.GPR[rd] = val
				c.reg.PC++
				return true
			}
		}
		return false
	}
}
func uMovImmReg(rd, _, _ int) microStep {
	return func(c *CPU) bool {
		c.reg.PC++
		fmt.Printf("TICK %d - reg%v<-#%d\n", c.tick, isa.GetRegisterMnemonic(rd), c.memI[c.reg.PC])
		c.reg.GPR[rd] = c.memI[c.reg.PC]
		c.reg.PC++
		return true
	}
}

func uMovRegReg(rd, rs1, _ int) microStep {
	return func(c *CPU) bool {
		c.reg.GPR[rd] = c.reg.GPR[rs1]
		c.reg.PC++
		return true
	}
}

// read32 consumes 4 u-tacts and возвращает true, когда слово готово
func (c *CPU) read32(addr uint32, stage *int, out *uint32) bool {
	switch *stage {
	case 0:
		c.tmp = c.memD[addr]
		*stage++
	case 1:
		c.tmp |= c.memD[addr+1] << 8
		*stage++
	case 2:
		c.tmp |= c.memD[addr+2] << 16
		*stage++
	case 3:
		c.tmp |= c.memD[addr+3] << 24
		*stage++
	default:
		*out = uint32(c.tmp)
		return true
	}
	return false
}
