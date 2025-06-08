package machine

import (
	"fmt"
	"log/slog"
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
	ucode[isa.OpMov][isa.MvRegIndReg] = uMovRegIndReg

	// JUMP BRENCH
	ucode[isa.OpCmp][isa.RegReg] = uCmpRR

	ucode[isa.OpJe][isa.JAbsAddr] = uJumpEquals

	//stack
	ucode[isa.OpPush][isa.SingleRegMode] = uPushReg
	ucode[isa.OpPop][isa.SingleRegMode] = uPopReg

	// MATH
	ucode[isa.OpAdd][isa.MathRRR] = uAddRRR
	ucode[isa.OpAdd][isa.MathRIR] = uAddRIR
	ucode[isa.OpSub][isa.MathRRR] = uSubRRR
	ucode[isa.OpMul][isa.MathRRR] = uMulRRR
	ucode[isa.OpDiv][isa.MathRRR] = uDivRRR

	// LOGICAL
	ucode[isa.OpAnd][isa.ImmReg] = uAndIR
	ucode[isa.OpAnd][isa.RegReg] = uAndRR

	// CALL RET ...

	ucode[isa.OpOut][isa.SingleRegMode] = uDivRRR

}

func uCmpRR(_, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		a := uint32(c.reg.GPR[rs1])
		b := uint32(c.reg.GPR[rs2])
		diff := int32(a) - int32(b)

		//TODO: check
		c.N = diff < 0                          // negative
		c.Z = diff == 0                         // zero
		c.V = ((a^b)&(uint32(diff)^a))>>31 == 1 // overflow
		c.C = a < b                             // borrow/carry

		fmt.Printf("TICK %d - CMP %v, %v | %v\n", c.tick, isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprFlags())

		return true
	}
}

func uJumpEquals(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.reg.GPR[r] = c.memI[c.reg.PC]
			fmt.Printf("TICK %d - %v<-memI[0x%X]; PC++ | %v\n", c.tick, isa.GetRegMnem(r), c.reg.PC, c.ReprRegVal(r))
			c.reg.PC++
			stage++
		case 1:
			if c.Z {
				c.reg.PC = c.reg.GPR[r]
				fmt.Printf("TICK %d - PC<-%v | PC=%v\n", c.tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.reg.PC++
			fmt.Printf("TICK %d - %v, no jump; PC++ | %v\n", c.tick, c.ReprFlags(), c.ReprPC())
			return true
		}
		return false
	}
}

func uAndIR(rd, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.reg.GPR[isa.RT] = c.memI[c.reg.PC]
			fmt.Printf("TICK %d - %v<-memI[0x%X]; PC++ | %v\n", c.tick, isa.GetRegMnem(isa.RT), c.reg.PC, c.ReprRegVal(isa.RT))
			c.reg.PC++
			stage++
		case 1:
			c.reg.GPR[rd] = c.reg.GPR[rs1] & c.reg.GPR[isa.RT]
			fmt.Printf("TICK %d - %v<-%v & %X | %v\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.reg.GPR[isa.RT], c.ReprRegVal(rd))
			return true
		}
		return false
	}
}
func uAndRR(rd, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		c.reg.GPR[rd] = c.reg.GPR[rs1] & c.reg.GPR[rs2]
		fmt.Printf("TICK %d - %v<-%v & %v | %v\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
		return true
	}
}

func uAddRRR(rd, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpAdd)
		return true
	}
}
func uAddRIR(rd, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.reg.GPR[isa.RT] = c.memI[c.reg.PC]
			fmt.Printf("TICK %d - %v<-memI[0x%X]; PC++ | %v\n", c.tick, isa.GetRegMnem(isa.RT), c.reg.PC, c.ReprRegVal(isa.RT))
			c.reg.PC++
			stage++
		case 1:
			MathRRR(c, rd, rs1, isa.RT, isa.OpAdd)
			return true
		}
		return false
	}
}

func uSubRRR(rd, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpSub)
		return true
	}
}
func uMulRRR(rd, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpMul)
		return true
	}
}
func uDivRRR(rd, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpDiv)
		return true
	}
}

func MathRRR(c *CPU, rd int, rs1 int, rs2 int, opc uint32) {
	//TODO: flags
	switch opc {
	case isa.OpAdd:
		c.reg.GPR[rd] = uint32(int32(c.reg.GPR[rs1]) + int32(c.reg.GPR[rs2]))
		fmt.Printf("TICK %d - %v<-%v+%v | %v\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
	case isa.OpSub:
		c.reg.GPR[rd] = uint32(int32(c.reg.GPR[rs1]) + int32(c.reg.GPR[rs2]))
		fmt.Printf("TICK %d - %v<-%v+%v | %v\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
	case isa.OpMul:
		//TODO: check or dont do in 1 tact
		c.reg.GPR[rd] = uint32(int32(c.reg.GPR[rs1]) * int32(c.reg.GPR[rs2]))
		fmt.Printf("TICK %d - %v<-%v+%v | %v\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
	case isa.OpDiv:
		//TODO: check or dont do in 1 tact
		c.reg.GPR[rd] = uint32(int32(c.reg.GPR[rs1]) / int32(c.reg.GPR[rs2]))
		fmt.Printf("TICK %d - %v<-%v+%v | %v\n", c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
	default:
		slog.Error(fmt.Sprintf("UNKNOWN ALU OP - %v", isa.GetOpMnemonic(opc)))

	}
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
func uMovRegIndReg(rd, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.reg.GPR[isa.RAddr] = c.reg.GPR[rs1]
			fmt.Printf("TICK %d - %v<-*%v | addr=%v\n",
				c.tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.ReprRegVal(isa.RAddr))
			stage = 1
		case 1, 2, 3, 4, 5: // читаем 4 байта LE -> rd
			if read32LE(c, &stage, isa.RAddr, rd) {
				fmt.Printf("TICK %d - %v\n", c.tick, c.ReprRegVal(rd))
				return true // микро-рутина завершена
			}
		}
		return false
	}
}
func uMovRegMem(_, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0: // sp -4
			c.reg.GPR[isa.RAddr] = c.memI[c.reg.PC]
			fmt.Printf("TICK %d - %v<-memI[0x%X]; PC++ \n", c.tick, isa.GetRegMnem(isa.RAddr), c.reg.PC)
			c.reg.PC++
			stage++
		case 1, 2, 3, 4, 5:
			if write32LE(c, &stage, isa.RAddr, rs1) {
				return true
			}
		}
		return false
	}
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
			fmt.Printf("%v %v=0x%X\n", c.ReprPC(), isa.GetRegMnem(isa.RAddr), c.reg.GPR[isa.RAddr])
			stage = 1
		case 1, 2, 3, 4, 5: // T1–T4 – читаем 4 байта
			if read32LE(c, &stage, isa.RAddr, rd) {
				//TODO: check of need tick--
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
		fmt.Printf("TICK %d - %v<-#%d; PC++ | %v\n", c.tick, isa.GetRegMnem(rd), c.memI[c.reg.PC], c.ReprRegVal(isa.SpReg))
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
