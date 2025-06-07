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

	//stack
	ucode[isa.OpPush][isa.SingleRegMode] = uPushReg
	ucode[isa.OpPop][isa.SingleRegMode] = uPopReg

	// MATH
	// CALL RET ...

}

func uPopReg(rd, _, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.reg.GPR[isa.RAddr] = c.reg.GPR[isa.SpReg]
			fmt.Printf("TICK %d - %v<-%v | ", c.tick, isa.GetRegMnem(isa.RAddr), isa.GetRegMnem(isa.SpReg))
			fmt.Printf("%v\n", c.ReprRegVal(isa.RAddr))
			stage++
		case 1, 2, 3, 4, 5:
			if read32LE(c, &stage, isa.RAddr, rd) {
				fmt.Printf("TICK %d - SP=SP+4 \n", c.tick)
				c.reg.GPR[isa.SpReg] += 4
				stage++
				return true
			}
		}
		return false
	}
}
func uPushReg(_, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0: // sp -4
			fmt.Printf("TICK %d - SP=SP-4 \n", c.tick)
			// fmt.Printf("rs1 %d %v\n", rs1, isa.GetRegMnem(rs1))
			c.reg.GPR[isa.SpReg] -= 4
			c.reg.GPR[isa.RAddr] = c.reg.GPR[isa.SpReg]
			c.reg.GPR[isa.RT] = c.reg.GPR[rs1]
			stage = 1
		case 1, 2, 3, 4, 5:
			if write32LE(c, &stage, isa.RAddr, rs1) {
				return true
			}
		}
		return false
	}
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
			c.reg.GPR[isa.RAddr] = c.memI[c.reg.PC]
			fmt.Printf("TICK %d - %v<-memI[%v]; PC++ | ", c.tick, isa.GetRegMnem(isa.RAddr), c.reg.PC)
			//TODO: prbly cannot pc++ on the same tick
			c.reg.PC++
			fmt.Printf("PC=%v %v=0x%X\n", c.reg.PC, isa.GetRegMnem(isa.RAddr), c.reg.GPR[isa.RAddr])
			stage = 1
		case 1, 2, 3, 4, 5: // T1–T4 – читаем 4 байта
			if read32LE(c, &stage, isa.RAddr, rd) {
				c.tick--
				return true
			}
		}
		return false
	}
}
func uMovImmReg(rd, _, _ int) microStep {
	return func(c *CPU) bool {
		c.reg.GPR[rd] = c.memI[c.reg.PC]
		fmt.Printf("TICK %d - %v<-#%d; PC++\n", c.tick, isa.GetRegMnem(rd), c.memI[c.reg.PC])
		c.reg.PC++
		return true
	}
}
func uMovRegReg(rd, rs1, _ int) microStep {
	return func(c *CPU) bool {
		c.reg.GPR[rd] = c.reg.GPR[rs1]
		c.reg.PC++
		fmt.Printf("TICK %d - %v<-%v; PC++\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rd))
		return true
	}
}

// записывает 32-битное значение little-endian по адресу в regWithAddr
// возвращает true, когда все 4 байта записаны
func write32LE(c *CPU, stage *int, regWithAddr int, regSource int) bool {
	// c.reg.GPR[regWithAddr := c.reg.GPR[regWithAddr]
	c.ensureDataSize(c.reg.GPR[regWithAddr] + 3)

	val := c.reg.GPR[regSource] // источник данных – регистр RT
	switch *stage {
	case 1:
		c.memD[c.reg.GPR[regWithAddr]] = byte(val)
		fmt.Printf("TICK %d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.tick, c.reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.reg.GPR[regWithAddr], c.memD[c.reg.GPR[regWithAddr]])
		c.reg.GPR[regWithAddr]++
	case 2:
		c.memD[c.reg.GPR[regWithAddr]] = byte(val >> 8)
		fmt.Printf("TICK %d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.tick, c.reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.reg.GPR[regWithAddr], c.memD[c.reg.GPR[regWithAddr]])
		c.reg.GPR[regWithAddr]++
	case 3:
		c.memD[c.reg.GPR[regWithAddr]] = byte(val >> 16)
		fmt.Printf("TICK %d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.tick, c.reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.reg.GPR[regWithAddr], c.memD[c.reg.GPR[regWithAddr]])
		c.reg.GPR[regWithAddr]++
	case 4:
		c.memD[c.reg.GPR[regWithAddr]] = byte(val >> 24)
		fmt.Printf("TICK %d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.tick, c.reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.reg.GPR[regWithAddr], c.memD[c.reg.GPR[regWithAddr]])
		c.reg.GPR[regWithAddr]++
		return true
	default:
		return true // все байты записаны
	}
	*stage++
	return false
}

// возвращают true, когда все 4 байта обработаны
func read32LE(c *CPU, stage *int, regWithAddr int, regToStoreTo int) bool {
	switch *stage {
	case 1:
		c.reg.GPR[regToStoreTo] = uint32(c.memD[c.reg.GPR[regWithAddr]])
		fmt.Printf("TICK %d - %v<-memD[%X] | %v\n", c.tick, isa.GetRegMnem(regToStoreTo),
			c.reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.reg.GPR[regWithAddr]++
	case 2:
		c.reg.GPR[regToStoreTo] |= uint32(c.memD[c.reg.GPR[regWithAddr]]) << 8
		fmt.Printf("TICK %d - %v<-memD[%X] | %v\n", c.tick, isa.GetRegMnem(regToStoreTo),
			c.reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.reg.GPR[regWithAddr]++
	case 3:
		c.reg.GPR[regToStoreTo] |= uint32(c.memD[c.reg.GPR[regWithAddr]]) << 16
		fmt.Printf("TICK %d - %v<-memD[%X] | %v\n", c.tick, isa.GetRegMnem(regToStoreTo),
			c.reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.reg.GPR[regWithAddr]++
	case 4:
		c.reg.GPR[regToStoreTo] |= uint32(c.memD[c.reg.GPR[regWithAddr]]) << 24
		fmt.Printf("TICK %d - %v<-memD[%X] | %v=%d/0x%X\n", c.tick, isa.GetRegMnem(regToStoreTo),
			c.reg.GPR[regWithAddr], isa.GetRegMnem(regToStoreTo), c.reg.GPR[regToStoreTo], c.reg.GPR[regToStoreTo])
		c.reg.GPR[regWithAddr]++
	default:
		return true
	}
	*stage++
	return false
}
