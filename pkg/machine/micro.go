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
	ucode[isa.OpNop][isa.NoOperands] = uNop
	ucode[isa.OpHalt][isa.NoOperands] = uHalt

	// ─── MOV rd, [memAbs] (2 words, 3 ticks) ─────────────
	ucode[isa.OpMov][isa.MvMemReg] = uMovMemReg
	ucode[isa.OpMov][isa.MvImmReg] = uMovImmReg
	ucode[isa.OpMov][isa.MvRegReg] = uMovRegReg
	ucode[isa.OpMov][isa.MvImmMem] = uMovImmMem
	ucode[isa.OpMov][isa.MvRegMem] = uMovRegMem
	ucode[isa.OpMov][isa.MvMemMem] = uMovMemMem

	// JUMP BRENCH

	// MATH
	// CALL RET ...

}

func uNop(_, _, _ int) microStep {
	return func(c *CPU) bool {
		fmt.Printf("TICK %d - NOP, PC++\n", c.tick)
		c.reg.PC++
		return true
	}
}

func uHalt(_, _, _ int) microStep {
	return func(c *CPU) bool {
		fmt.Println("simultaion stopped")
		os.Exit(0)
		return true
	}
}

func uMovMemMem(rd, _, _ int) microStep {
	panic("unimplemeted")
}
func uMovImmMem(rd, _, _ int) microStep {
	panic("unimplemeted")
}
func uMovRegMem(rd, _, _ int) microStep {
	panic("unimplemeted")
}
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
