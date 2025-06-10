package machine

import (
	"fmt"
	"log/slog"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type microStep func(c *CPU) bool // true → macro-instruction done

const (
	opcMaxAmount   = 64 // 6 bit opcode
	modesMaxAmount = 32 // 5 bit
)

// [opcode][mode] → factory(rd,rs1,rs2)→microStep
var ucode [opcMaxAmount][modesMaxAmount]func(int, int, int) microStep

func init() {
	ucode[isa.OpNop][isa.NoOperands] = uNop
	ucode[isa.OpHalt][isa.NoOperands] = uHalt

	// data flow
	ucode[isa.OpMov][isa.MvMemReg] = uMovMemReg
	ucode[isa.OpMov][isa.MvImmReg] = uMovImmReg
	ucode[isa.OpMov][isa.MvRegReg] = uMovRegReg
	ucode[isa.OpMov][isa.MvImmMem] = uMovImmMem
	ucode[isa.OpMov][isa.MvRegMem] = uMovRegMem
	ucode[isa.OpMov][isa.MvRegIndReg] = uMovRegIndReg
	ucode[isa.OpMov][isa.MvLowRegIndReg] = uMovLowRegIndReg
	ucode[isa.OpMov][isa.MvRegLowMem] = uMovRegLowMem

	// JUMP BRENCH
	ucode[isa.OpCmp][isa.RegReg] = uCmpRR

	ucode[isa.OpJmp][isa.JAbsAddr] = uJump
	ucode[isa.OpJe][isa.JAbsAddr] = uJE
	ucode[isa.OpJne][isa.JAbsAddr] = uJNE
	ucode[isa.OpJg][isa.JAbsAddr] = uJG
	ucode[isa.OpJl][isa.JAbsAddr] = uJL
	ucode[isa.OpJge][isa.JAbsAddr] = uJGE
	ucode[isa.OpJle][isa.JAbsAddr] = uJLE

	//stack
	ucode[isa.OpPush][isa.SingleRegMode] = uPushReg
	ucode[isa.OpPop][isa.SingleRegMode] = uPopReg

	// MATH
	ucode[isa.OpAdd][isa.MathRRR] = uAddRRR
	ucode[isa.OpAdd][isa.MathRIR] = uAddRIR
	ucode[isa.OpSub][isa.MathRRR] = uSubRRR
	ucode[isa.OpSub][isa.MathRIR] = uSubRIR
	ucode[isa.OpMul][isa.MathRRR] = uMulRRR
	ucode[isa.OpDiv][isa.MathRRR] = uDivRRR

	// LOGICAL
	ucode[isa.OpAnd][isa.ImmReg] = uAndIR
	ucode[isa.OpAnd][isa.RegReg] = uAndRR

	// IO
	ucode[isa.OpOut][isa.ByteM] = uOutB
	ucode[isa.OpOut][isa.WordM] = uOutW
	ucode[isa.OpIn][isa.ByteM] = uInB
	ucode[isa.OpIn][isa.WordM] = uInW

	ucode[isa.OpIRet][isa.NoOperands] = uIRet

}

func uIRet(_, _, _ int) microStep {
	return func(c *CPU) bool {
		c.RestoreGPRValues()
		c.inISR = false
		c.pending = false
		c.Reg.PC = c.Reg.savedPC
		c.RestoreNZVC()
		fmt.Printf("TICK % 4d - restore register values | %v\n", c.Tick, c.ReprPC())
		fmt.Println("------------Exiting interruption------------")
		return true
	}
}

func uInB(_, _, _ int) microStep {
	return func(c *CPU) bool {
		val := c.Ioc.ReadPort(isa.PortCh)
		c.Reg.GPR[isa.RInData] = uint32(val)
		fmt.Printf("TICK % 4d - %s <- %v (%d/0x%X) | %v\n",
			c.Tick, isa.GetRegMnem(isa.RInData), isa.GetPortMnem(isa.PortCh), val, val, c.ReprRegVal(isa.RInData))
		return true
	}
}

// TODO: check, now reads 1 byte from digit port
func uInW(rd, _, _ int) microStep {
	return func(c *CPU) bool {
		val := c.Ioc.ReadPort(isa.PortCh)
		c.Reg.GPR[isa.RInData] = uint32(val)
		fmt.Printf("TICK % 4d - %s <- %d word (%d/0x%X) | %v\n",
			c.Tick, isa.GetRegMnem(isa.RInData), isa.PortD, val, val, c.ReprRegVal(isa.RInData))
		return true
	}
}

func uOutB(rd, _, _ int) microStep {
	return func(c *CPU) bool {
		//TODO: if print then always 1
		port := uint8(1)
		data := byte(c.Reg.GPR[isa.ROutData])
		c.Ioc.WritePort(port, data)
		fmt.Printf("TICK % 4d - port %d <- %s (0x%02X) byte | %v\n",
			c.Tick, port, isa.GetRegMnem(isa.ROutData), data, c.Ioc.Output(port))
		return true
	}
}

func uOutW(rd, _, _ int) microStep {
	return func(c *CPU) bool {
		port := uint8(0)
		data := c.Reg.GPR[isa.ROutData]

		numStr := fmt.Sprintf("%d", int32(data))

		for _, char := range numStr {
			c.Ioc.WritePort(port, byte(char))
		}

		//TODO: check strange output in %v ioc output
		fmt.Printf("TICK % 4d - port %d <- %s  | %v\n",
			c.Tick, port, isa.GetRegMnem(isa.ROutData), c.Ioc.Output(port))
		return true
	}
}

func uCmpRR(_, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		a := uint32(c.Reg.GPR[rs1])
		b := uint32(c.Reg.GPR[rs2])
		diff := int32(a) - int32(b)

		//TODO: check
		c.N = diff < 0                          // negative
		c.Z = diff == 0                         // zero
		c.V = ((a^b)&(uint32(diff)^a))>>31 == 1 // overflow
		c.C = a < b                             // borrow/carry

		fmt.Printf("TICK % 4d - CMP %v, %v | %v; %v %v\n", c.Tick, isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprFlags(), c.ReprRegVal(rs1), c.ReprRegVal(rs2))

		return true
	}
}

func uJump(_, _, _ int) microStep {
	//TODO: should be 1 tick?
	stage := 0
	r := isa.RT
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]| %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			// c.Reg.PC++
			stage++
		case 1:
			c.Reg.PC = c.Reg.GPR[r]
			fmt.Printf("TICK % 4d - PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
			return true
		}
		return false
	}
}
func uJE(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.Z {
				c.Reg.PC = c.Reg.GPR[r]
				fmt.Printf("TICK % 4d - PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			fmt.Printf("TICK % 4d - no jump | %v; %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJNE(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if !c.Z {
				c.Reg.PC = c.Reg.GPR[r]
				fmt.Printf("TICK % 4d - JNE taken; PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			fmt.Printf("TICK % 4d - JNE not taken | %v; %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJG(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if !c.Z && !c.N {
				c.Reg.PC = c.Reg.GPR[r]
				fmt.Printf("TICK % 4d - JG taken → PC<-%v | %v\n",
					c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			fmt.Printf("TICK % 4d - JG not taken | %v %v\n",
				c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJL(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.N {
				c.Reg.PC = c.Reg.GPR[r]
				fmt.Printf("TICK % 4d - JL taken → PC<-%v | %v\n",
					c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			fmt.Printf("TICK % 4d - JL not taken | %v %v\n",
				c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJGE(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.Z || !c.N {
				c.Reg.PC = c.Reg.GPR[r]
				fmt.Printf("TICK % 4d - JGE taken → PC<-%v | %v\n",
					c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			fmt.Printf("TICK % 4d - JGE not taken | %v %v\n",
				c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJLE(_, _, _ int) microStep {
	stage := 0
	r := isa.RAddr
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.Z || c.N {
				c.Reg.PC = c.Reg.GPR[r]
				fmt.Printf("TICK % 4d - JLE taken → PC<-%v | %v\n",
					c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			fmt.Printf("TICK % 4d - JLE not taken | %v %v\n",
				c.Tick, c.ReprPC(), c.ReprFlags())
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
			c.Reg.GPR[isa.RT] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(isa.RT), c.Reg.PC, c.ReprRegVal(isa.RT))
			c.Reg.PC++
			stage++
		case 1:
			c.Reg.GPR[rd] = c.Reg.GPR[rs1] & c.Reg.GPR[isa.RT]
			fmt.Printf("TICK % 4d - %v<-%v & %X | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.Reg.GPR[isa.RT], c.ReprRegVal(rd))
			return true
		}
		return false
	}
}
func uAndRR(rd, rs1, rs2 int) microStep {
	return func(c *CPU) bool {
		c.Reg.GPR[rd] = c.Reg.GPR[rs1] & c.Reg.GPR[rs2]
		fmt.Printf("TICK % 4d - %v<-%v & %v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
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
			c.Reg.GPR[isa.RT] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(isa.RT), c.Reg.PC, c.ReprRegVal(isa.RT))
			c.Reg.PC++
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
func uSubRIR(rd, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[isa.RT] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(isa.RT), c.Reg.PC, c.ReprRegVal(isa.RT))
			c.Reg.PC++
			stage++
		case 1:
			MathRRR(c, rd, rs1, isa.RT, isa.OpSub)
			return true
		}
		return false
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

func MathRRR(c *CPU, rd, rs1, rs2 int, opc uint32) {
	//TODO: check flag setting
	a := uint32(c.Reg.GPR[rs1])
	b := uint32(c.Reg.GPR[rs2])

	clrF := func() { c.N, c.Z, c.V, c.C = false, false, false, false }

	switch opc {
	case isa.OpAdd:
		res64 := uint64(a) + uint64(b)
		res := uint32(res64)

		c.Reg.GPR[rd] = res
		clrF()
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		c.C = res64 > 0xFFFF_FFFF                                    // carry-out (unsigned)
		c.V = ((a^b)&0x8000_0000 == 0) && ((a^res)&0x8000_0000 != 0) // signed overflow

	case isa.OpSub:
		res64 := uint64(a) - uint64(b)
		res := uint32(res64)

		c.Reg.GPR[rd] = res
		clrF()
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		c.C = a >= b // 1 = no borrow
		c.V = ((a^b)&0x8000_0000 != 0) && ((a^res)&0x8000_0000 != 0)

	case isa.OpMul:
		prod := int64(int32(a)) * int64(int32(b))
		res := uint32(prod)

		c.Reg.GPR[rd] = res
		clrF()
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		c.V = prod < int64(int32(-0x8000_0000)) || prod > int64(int32(0x7FFF_FFFF))
		// C не определяем – остаётся 0

	case isa.OpDiv:
		clrF()
		if b == 0 {
			c.V = true // деление на ноль → overflow
			return
		}
		quot := int32(a) / int32(b)
		res := uint32(quot)

		c.Reg.GPR[rd] = res
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		// V, C уже обнулены

	default:
		slog.Error("UNKNOWN ALU OP", "op", isa.GetOpMnemonic(opc))
	}

	aluOpLu := map[uint32]string{
		isa.OpAdd: "+",
		isa.OpSub: "-",
		isa.OpMul: "*",
		isa.OpDiv: "/",
	}

	fmt.Printf("TICK % 4d - %v<-(%v%v%v) | %v %v\n",
		c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), aluOpLu[opc], isa.GetRegMnem(rs2), c.ReprRegVal(rd), c.ReprFlags())
}

func uPopReg(rd, _, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[isa.RAddr] = c.Reg.GPR[isa.SpReg]
			fmt.Printf("TICK % 4d - %v<-%v | ", c.Tick, isa.GetRegMnem(isa.RAddr), isa.GetRegMnem(isa.SpReg))
			fmt.Printf("%v\n", c.ReprRegVal(isa.RAddr))
			stage++
		case 1, 2, 3, 4, 5:
			if read32LE(c, &stage, isa.RAddr, rd) {
				fmt.Printf("TICK % 4d - SP=SP+4 \n", c.Tick)
				c.Reg.GPR[isa.SpReg] += 4
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
			fmt.Printf("TICK % 4d - SP=SP-4 \n", c.Tick)
			// fmt.Printf("rs1 % 4d %v\n", rs1, isa.GetRegMnem(rs1))
			c.Reg.GPR[isa.SpReg] -= 4
			c.Reg.GPR[isa.RAddr] = c.Reg.GPR[isa.SpReg]
			c.Reg.GPR[isa.RT] = c.Reg.GPR[rs1]
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
		fmt.Printf("TICK % 4d - NOP, PC++\n", c.Tick)
		c.Reg.PC++
		return true
	}
}

func uHalt(_, _, _ int) microStep {
	return func(c *CPU) bool {
		fmt.Printf("TICK % 4d - simultaion stopped\n", c.Tick)
		c.halted = true
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
			c.Reg.GPR[isa.RAddr] = c.Reg.GPR[rs1]
			fmt.Printf("TICK % 4d - %v<-*%v | %v\n",
				c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.ReprRegVal(rd))
			stage++
		case 1, 2, 3, 4, 5: // читаем 4 байта LE -> rd
			if read32LE(c, &stage, isa.RAddr, rd) {
				fmt.Printf("TICK % 4d - %v\n", c.Tick, c.ReprRegVal(rd))
				return true // микро-рутина завершена
			}
		}
		return false
	}
}
func uMovLowRegIndReg(rd, rs1, _ int) microStep {
	return func(c *CPU) bool {
		addr := c.Reg.GPR[rs1]
		c.ensureDataSize(addr)
		c.Reg.GPR[rd] = uint32(c.memD[addr])
		fmt.Printf("TICK % 4d - %s <- memD[%X] | %v\n",
			c.Tick, isa.GetRegMnem(rd), addr, c.ReprRegVal(rd))
		return true
	}
}
func uMovRegLowMem(_, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[isa.RAddr] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %s <- memI[0x%X]; PC++ | %v\n",
				c.Tick, isa.GetRegMnem(isa.RAddr), c.Reg.PC, c.ReprRegVal(isa.RAddr))
			c.Reg.PC++
			stage++
		case 1:
			addr := c.Reg.GPR[isa.RAddr]
			val := byte(c.Reg.GPR[rs1] & 0xFF)
			c.ensureDataSize(addr)
			c.memD[addr] = val
			fmt.Printf("TICK % 4d - memD[0x%X] <- %s[0:8] = 0x%02X\n",
				c.Tick, addr, isa.GetRegMnem(rs1), val)
			return true
		}
		return false
	}
}
func uMovRegMem(_, rs1, _ int) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0: // sp -4
			c.Reg.GPR[isa.RAddr] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[0x%X]; PC++ \n", c.Tick, isa.GetRegMnem(isa.RAddr), c.Reg.PC)
			c.Reg.PC++
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
			c.Reg.GPR[isa.RAddr] = c.memI[c.Reg.PC]
			fmt.Printf("TICK % 4d - %v<-memI[%v]; PC++ | ", c.Tick, isa.GetRegMnem(isa.RAddr), c.Reg.PC)
			//TODO: prbly cannot pc++ on the same tick
			c.Reg.PC++
			fmt.Printf("%v %v=0x%X\n", c.ReprPC(), isa.GetRegMnem(isa.RAddr), c.Reg.GPR[isa.RAddr])
			stage = 1
		case 1, 2, 3, 4, 5: // T1–T4 – читаем 4 байта
			if read32LE(c, &stage, isa.RAddr, rd) {
				//TODO: check of need tick--
				c.Tick--
				return true
			}
		}
		return false
	}
}
func uMovImmReg(rd, _, _ int) microStep {
	return func(c *CPU) bool {
		c.Reg.GPR[rd] = c.memI[c.Reg.PC]
		fmt.Printf("TICK % 4d - %v<-#% 4d; PC++ | %v\n", c.Tick, isa.GetRegMnem(rd), c.memI[c.Reg.PC], c.ReprRegVal(isa.SpReg))
		c.Reg.PC++
		return true
	}
}
func uMovRegReg(rd, rs1, _ int) microStep {
	return func(c *CPU) bool {
		c.Reg.GPR[rd] = c.Reg.GPR[rs1]
		fmt.Printf("TICK % 4d - %v<-%v; PC++ | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.ReprRegVal(rd))
		return true
	}
}

// записывает 32-битное значение little-endian по адресу в regWithAddr
// возвращает true, когда все 4 байта записаны
func write32LE(c *CPU, stage *int, regWithAddr int, regSource int) bool {
	// c.reg.GPR[regWithAddr := c.reg.GPR[regWithAddr]
	c.ensureDataSize(c.Reg.GPR[regWithAddr] + 3)

	val := c.Reg.GPR[regSource] // источник данных – регистр RT
	switch *stage {
	case 1:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val)
		fmt.Printf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
	case 2:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val >> 8)
		fmt.Printf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
	case 3:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val >> 16)
		fmt.Printf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
	case 4:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val >> 24)
		fmt.Printf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n",
			c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
		return true
	default:
		return true // все байты записаны
	}
	*stage++
	return false
}

// возвращают true, когда все 4 байта обработаны
func read32LE(c *CPU, stage *int, regWithAddr int, regToStoreTo int) bool {
	c.ensureDataSize(c.Reg.GPR[regWithAddr] + 3)
	switch *stage {
	case 1:
		c.Reg.GPR[regToStoreTo] = uint32(c.memD[c.Reg.GPR[regWithAddr]])
		fmt.Printf("TICK % 4d - %v<-memD[%X] | %v\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.Reg.GPR[regWithAddr]++
	case 2:
		c.Reg.GPR[regToStoreTo] |= uint32(c.memD[c.Reg.GPR[regWithAddr]]) << 8
		fmt.Printf("TICK % 4d - %v<-memD[%X] | %v\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.Reg.GPR[regWithAddr]++
	case 3:
		c.Reg.GPR[regToStoreTo] |= uint32(c.memD[c.Reg.GPR[regWithAddr]]) << 16
		fmt.Printf("TICK % 4d - %v<-memD[%X] | %v\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.Reg.GPR[regWithAddr]++
	case 4:
		c.Reg.GPR[regToStoreTo] |= uint32(c.memD[c.Reg.GPR[regWithAddr]]) << 24
		fmt.Printf("TICK % 4d - %v<-memD[%X] | %v=% 4d/0x%X\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], isa.GetRegMnem(regToStoreTo), c.Reg.GPR[regToStoreTo], c.Reg.GPR[regToStoreTo])
		c.Reg.GPR[regWithAddr]++
	default:
		return true
	}
	*stage++
	return false
}
