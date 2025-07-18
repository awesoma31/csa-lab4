package machine

import (
	"log/slog"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type microStep func(c *CPU) bool // true → macro-instruction done

const (
	opcMaxAmount   = 64 // 6 bit opcode
	modesMaxAmount = 32 // 5 bit
)

// [opcode][mode] → factory(rd,rs1,rs2)→microStep
var ucode [opcMaxAmount][modesMaxAmount]func(isa.Register, isa.Register, isa.Register) microStep

func init() {
	ucode[isa.OpNop][isa.NoOperands] = uNop
	ucode[isa.OpHalt][isa.NoOperands] = uHalt

	// data flow
	ucode[isa.OpMov][isa.MvMemReg] = uMovMemReg
	ucode[isa.OpMov][isa.MvImmReg] = uMovImmReg
	ucode[isa.OpMov][isa.MvRegReg] = uMovRegReg
	ucode[isa.OpMov][isa.MvRegMem] = uMovRegMem
	ucode[isa.OpMov][isa.MvRegIndToReg] = uMovRegIndReg
	ucode[isa.OpMov][isa.MvByteRegIndToReg] = uMovByteRegIndReg
	ucode[isa.OpMov][isa.MvRegLowToMem] = uMovRegByteToMem
	ucode[isa.OpMov][isa.MvLowRegToRegInd] = uMovByteRegToRegInd

	// JUMP BRENCH
	ucode[isa.OpCmp][isa.RegReg] = uCmpRR

	ucode[isa.OpJmp][isa.JAbsAddr] = uJump
	ucode[isa.OpJe][isa.JAbsAddr] = uJE
	ucode[isa.OpJne][isa.JAbsAddr] = uJNE
	ucode[isa.OpJg][isa.JAbsAddr] = uJG
	ucode[isa.OpJl][isa.JAbsAddr] = uJL
	ucode[isa.OpJge][isa.JAbsAddr] = uJGE
	ucode[isa.OpJle][isa.JAbsAddr] = uJLE
	ucode[isa.OpJcc][isa.JAbsAddr] = uJCC
	ucode[isa.OpJcs][isa.JAbsAddr] = uJCS

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
	ucode[isa.OpOut][isa.ByteM] = uOutCh
	ucode[isa.OpOut][isa.DigitM] = uOutD
	ucode[isa.OpOut][isa.LongM] = uOutL
	ucode[isa.OpIn][isa.ByteM] = uInCh
	ucode[isa.OpIn][isa.DigitM] = uInD

	ucode[isa.OpIRet][isa.NoOperands] = uIRet
	ucode[isa.OpIntOn][isa.NoOperands] = uIntOn
	ucode[isa.OpIntOff][isa.NoOperands] = uIntOff

}

func uIntOn(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.IsIntOn = true
		c.log.Debugf("TICK % 4d - interruptions on | %v\n", c.Tick, c.IsIntOn)
		return true
	}
}

func uIntOff(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.IsIntOn = false
		c.log.Debugf("TICK % 4d - interruptions on | %v\n", c.Tick, c.IsIntOn)
		return true
	}
}

func uIRet(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.RestoreGPRValues()
		c.inISR = false
		c.pending = false
		c.Reg.PC = c.Reg.savedPC
		c.RestoreNZVC()
		c.log.Debugf("TICK % 4d - restore register values | %v\n", c.Tick, c.ReprPC())
		c.log.Debug("------------Exiting interruption------------")
		return true
	}
}

func uInCh(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		val := c.Ioc.ReadPort(isa.PortCh)
		c.Reg.GPR[isa.RInData] = val
		c.log.Debugf("TICK % 4d - %s <- %v (%d/0x%X) | %v\n", c.Tick, isa.GetRegMnem(isa.RInData), isa.GetPortMnem(isa.PortCh), val, val, c.ReprRegVal(isa.RInData))
		return true
	}
}

func uInD(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		val := c.Ioc.ReadPort(isa.PortD)
		signedVal := int32(val)
		c.Reg.GPR[isa.RInData] = uint32(signedVal)
		c.log.Debugf("TICK % 4d - %s <- %d digit (%d/0x%X) | %v\n", c.Tick, isa.GetRegMnem(isa.RInData), isa.PortD, val, val, c.ReprRegVal(isa.RInData))
		return true
	}
}

func uOutCh(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		port := isa.PortCh
		data := c.Reg.GPR[isa.ROutData]
		c.Ioc.WritePort(port, data)
		c.log.Debugf("TICK % 4d - port %d <- %s(0x%02X) char | %v\n", c.Tick, port, isa.GetRegMnem(isa.ROutData), data, c.Ioc.Output(port))
		return true
	}
}

func uOutD(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		port := isa.PortD
		data := c.Reg.GPR[isa.ROutData]
		c.Ioc.WritePort(port, data)
		c.log.Debugf("TICK % 4d - port %d <- %s(0x%02X) digit | %v\n", c.Tick, port, isa.GetRegMnem(isa.ROutData), data, c.Ioc.Output(port))
		return true
	}
}
func uOutL(_, _, _ isa.Register) microStep {
	var hi, lo uint32
	stage := 0
	readStage := 1
	port := isa.PortL

	return func(c *CPU) bool {
		switch stage {
		case 0:
			if read32(c, &readStage, isa.ROutAddr, isa.ROutData) {
				lo = c.Reg.GPR[isa.ROutData]
				c.Ioc.WritePort(port, lo)
				c.log.Debugf("TICK % 4d - %v <- %s(0x%02X) long(lo) | %v\n", c.Tick, isa.GetPortMnem(isa.Register(port)), isa.GetRegMnem(isa.ROutData), lo, c.Ioc.Output(port))
				readStage = 1
				stage = 1
			}
		case 1:
			if read32(c, &readStage, isa.ROutAddr, isa.ROutData) {
				hi = c.Reg.GPR[isa.ROutData]
				c.Ioc.WritePort(port, hi)
				c.log.Debugf("TICK % 4d - %v <- %s(0x%02X) long(hi) | %v\n", c.Tick, isa.GetPortMnem(isa.Register(port)), isa.GetRegMnem(isa.ROutData), hi, c.Ioc.Output(port))
				return true
			}
		}
		return false
	}
}

func uCmpRR(_, rs1, rs2 isa.Register) microStep {
	return func(c *CPU) bool {
		a := c.Reg.GPR[rs1]
		b := c.Reg.GPR[rs2]
		diff := int32(a) - int32(b)

		c.N = diff < 0
		c.Z = diff == 0
		c.V = ((a^b)&(uint32(diff)^a))>>31 == 1
		c.C = a < b

		c.log.Debugf("TICK % 4d - CMP %v, %v | %v; %v %v\n", c.Tick, isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprFlags(), c.ReprRegVal(rs1), c.ReprRegVal(rs2))

		return true
	}
}

func uJCC(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if !c.C {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JCC not taken | %v; %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJCS(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.C {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JCS not taken | %v; %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}

func uJump(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.Reg.PC = c.memI[c.Reg.PC]
		c.log.Debugf("TICK % 4d - PC<-memI[0x%X]| %v\n", c.Tick, c.Reg.PC, c.ReprPC())
		return true
	}
}
func uJE(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.Z {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - no jump | %v; %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJNE(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if !c.Z {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - JNE taken; PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JNE not taken | %v; %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJG(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if !c.Z && !c.N {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - JG taken → PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JG not taken | %v %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJL(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.N {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - JL taken → PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JL not taken | %v %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJGE(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.Z || !c.N {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - JGE taken → PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JGE not taken | %v %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}
func uJLE(_, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			if c.Z || c.N {
				c.Reg.PC = c.Reg.GPR[r]
				c.log.Debugf("TICK % 4d - JLE taken → PC<-%v | %v\n", c.Tick, isa.GetRegMnem(r), c.ReprPC())
				return true
			}
			c.log.Debugf("TICK % 4d - JLE not taken | %v %v\n", c.Tick, c.ReprPC(), c.ReprFlags())
			return true
		}
		return false
	}
}

func uAndIR(rd, rs1, _ isa.Register) microStep {
	stage := 0
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[isa.RT] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(isa.RT), c.Reg.PC, c.ReprRegVal(isa.RT))
			c.Reg.PC++
			stage++
		case 1:
			c.Reg.GPR[rd] = c.Reg.GPR[rs1] & c.Reg.GPR[isa.RT]
			c.log.Debugf("TICK % 4d - %v<-%v & %X | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.Reg.GPR[isa.RT], c.ReprRegVal(rd))
			return true
		}
		return false
	}
}
func uAndRR(rd, rs1, rs2 isa.Register) microStep {
	return func(c *CPU) bool {
		c.Reg.GPR[rd] = c.Reg.GPR[rs1] & c.Reg.GPR[rs2]
		c.log.Debugf("TICK % 4d - %v<-%v & %v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
		return true
	}
}

func uAddRRR(rd, rs1, rs2 isa.Register) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpAdd)
		c.log.Debugf("TICK % 4d - %v<-%v + %v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
		return true
	}
}
func uAddRIR(rd, rs1, _ isa.Register) microStep {
	stage := 0
	r := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			MathRRR(c, rd, rs1, r, isa.OpAdd)
			return true
		}
		return false
	}
}

func uSubRRR(rd, rs1, rs2 isa.Register) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpSub)
		return true
	}
}
func uSubRIR(rd, rs1, _ isa.Register) microStep {
	stage := 0
	r := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			c.log.Debugf("TICK % 4d - %v<-%v-%v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(r), c.ReprRegVal(rd))
			MathRRR(c, rd, rs1, r, isa.OpSub)
			return true
		}
		return false
	}
}
func uMulRRR(rd, rs1, rs2 isa.Register) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpMul)
		c.log.Debugf("TICK % 4d - %v<-%v*%v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
		return true
	}
}
func uDivRRR(rd, rs1, rs2 isa.Register) microStep {
	return func(c *CPU) bool {
		MathRRR(c, rd, rs1, rs2, isa.OpDiv)
		c.log.Debugf("TICK % 4d - %v<-%v//%v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), isa.GetRegMnem(rs2), c.ReprRegVal(rd))
		return true
	}
}

func MathRRR(c *CPU, rd, rs1, rs2 isa.Register, opc uint32) {
	a := c.Reg.GPR[rs1]
	b := c.Reg.GPR[rs2]

	clrF := func() { c.N, c.Z, c.V, c.C = false, false, false, false }

	switch opc {
	case isa.OpAdd:
		res64 := uint64(a) + uint64(b)
		res := uint32(res64)

		c.Reg.GPR[rd] = res
		clrF()
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		c.C = res64 > 0xFFFF_FFFF
		c.V = ((a^b)&0x8000_0000 == 0) && ((a^res)&0x8000_0000 != 0)

	case isa.OpSub:
		res64 := uint64(a) - uint64(b)
		res := uint32(res64)

		c.Reg.GPR[rd] = res
		clrF()
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		c.C = a >= b
		c.V = ((a^b)&0x8000_0000 != 0) && ((a^res)&0x8000_0000 != 0)

	case isa.OpMul:
		prod := int64(int32(a)) * int64(int32(b))
		res := uint32(prod)

		c.Reg.GPR[rd] = res
		clrF()
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0
		c.V = prod < int64(int32(-0x8000_0000)) || prod > int64(int32(0x7FFF_FFFF))

	case isa.OpDiv:
		clrF()
		if b == 0 {
			c.V = true
			return
		}
		quot := int32(a) / int32(b)
		res := uint32(quot)

		c.Reg.GPR[rd] = res
		c.N = res&0x8000_0000 != 0
		c.Z = res == 0

	default:
		slog.Error("UNKNOWN ALU OP", "op", isa.GetOpMnemonic(opc))
	}

	aluOpLu := map[uint32]string{
		isa.OpAdd: "+",
		isa.OpSub: "-",
		isa.OpMul: "*",
		isa.OpDiv: "/",
	}

	c.log.Debugf("TICK % 4d - %v<-%v%v%v | %v %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), aluOpLu[opc], isa.GetRegMnem(rs2), c.ReprRegVal(rd), c.ReprFlags())
}

func uPopReg(rd, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.Reg.GPR[isa.SpReg]
			c.log.Debugf("TICK % 4d - %v<-%v | %v", c.Tick, isa.GetRegMnem(r), isa.GetRegMnem(isa.SpReg), c.ReprRegVal(r))
			stage++
		case 1, 2, 3, 4, 5:
			if read32(c, &stage, r, rd) {
				c.log.Debugf("TICK % 4d - SP=SP+4 | %v\n", c.Tick, c.ReprRegVal(isa.SpReg))
				c.Reg.GPR[isa.SpReg] += 4
				stage++
				return true
			}
		}
		return false
	}
}

func uPushReg(_, rs1, _ isa.Register) microStep {
	stage := 0
	readStage := 1
	rf1 := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[isa.SpReg] -= 4
			c.log.Debugf("TICK % 4d - SP=SP-4 | %v\n", c.Tick, c.ReprRegVal(isa.SpReg))
			stage++
		case 1:
			c.Reg.GPR[rf1] = c.Reg.GPR[isa.SpReg]
			c.log.Debugf("TICK % 4d - %v=%v | %v\n", c.Tick, isa.GetRegMnem(rf1), isa.GetRegMnem(isa.SpReg), c.ReprRegVal(isa.SpReg))
			stage++
		case 2:
			if write32(c, &readStage, rf1, rs1) {
				return true
			}
		}
		return false
	}
}

func uNop(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.log.Debugf("TICK % 4d - NOP, PC++ | %v\n", c.Tick, c.ReprPC())
		c.Reg.PC++
		return true
	}
}

func uHalt(_, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.log.Debugf("TICK % 4d - simultaion stopped\n", c.Tick)
		c.halted = true
		return true
	}
}

func uMovRegIndReg(rd, rs1, _ isa.Register) microStep {
	stage := 0
	r := isa.RF2
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.Reg.GPR[rs1]
			c.log.Debugf("TICK % 4d - %v<-%v | %v\n", c.Tick, isa.GetRegMnem(r), isa.GetRegMnem(rs1), c.ReprRegVal(r))
			stage++
		case 1, 2, 3, 4, 5:
			if read32(c, &stage, r, rd) {
				c.log.Debugf("TICK % 4d - %v\n", c.Tick, c.ReprRegVal(rd))
				return true
			}
		}
		return false
	}
}

func uMovByteRegIndReg(rd, rs1, _ isa.Register) microStep {
	return func(c *CPU) bool {
		addr := c.Reg.GPR[rs1]
		c.ensureDataSize(addr)
		c.Reg.GPR[rd] = uint32(c.memD[addr])
		c.log.Debugf("TICK % 4d - %s <- memD[%X] | %v\n", c.Tick, isa.GetRegMnem(rd), addr, c.ReprRegVal(rd))
		return true
	}
}
func uMovRegByteToMem(_, rs1, _ isa.Register) microStep {
	stage := 0
	r := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %s <- memI[0x%X]; PC++ | %v\n", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage++
		case 1:
			addr := c.Reg.GPR[r]
			val := byte(c.Reg.GPR[rs1] & 0xFF)
			c.ensureDataSize(addr)
			c.memD[addr] = val
			c.log.Debugf("TICK % 4d - memD[0x%X] <- %s(byte) = 0x%02X\n", c.Tick, addr, isa.GetRegMnem(rs1), val)
			return true
		}
		return false
	}
}
func uMovByteRegToRegInd(rd, rs1, _ isa.Register) microStep {
	return func(c *CPU) bool {
		val := byte(c.Reg.GPR[rs1] & 0xFF)
		addr := c.Reg.GPR[rd]
		c.ensureDataSize(addr)
		c.memD[addr] = val
		c.log.Debugf("TICK % 4d - memD[0x%X] <- %s(byte); mem[%s]<-%s(byte) = 0x%02X\n", c.Tick, addr, isa.GetRegMnem(rs1), isa.GetRegMnem(rd), isa.GetRegMnem(rs1), val)

		return true
	}
}
func uMovRegMem(_, rs1, _ isa.Register) microStep {
	stage := 0
	r := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[0x%X]; PC++ \n", c.Tick, isa.GetRegMnem(r), c.Reg.PC)
			c.Reg.PC++
			stage++
		case 1, 2, 3, 4, 5:
			if write32(c, &stage, r, rs1) {
				return true
			}
		}
		return false
	}
}
func uMovMemReg(rd, _, _ isa.Register) microStep {
	stage := 0
	r := isa.RF1
	return func(c *CPU) bool {
		switch stage {
		case 0:
			c.Reg.GPR[r] = c.memI[c.Reg.PC]
			c.log.Debugf("TICK % 4d - %v<-memI[%v], PC++ | %v", c.Tick, isa.GetRegMnem(r), c.Reg.PC, c.ReprRegVal(r))
			c.Reg.PC++
			stage = 1
		case 1, 2, 3, 4, 5:
			if read32(c, &stage, r, rd) {
				return true
			}
		}
		return false
	}
}
func uMovImmReg(rd, _, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.Reg.GPR[rd] = c.memI[c.Reg.PC]
		c.log.Debugf("TICK % 4d - %v<-#%d; PC++ | %v\n", c.Tick, isa.GetRegMnem(rd), c.memI[c.Reg.PC], c.ReprRegVal(isa.SpReg))
		c.Reg.PC++
		return true
	}
}
func uMovRegReg(rd, rs1, _ isa.Register) microStep {
	return func(c *CPU) bool {
		c.Reg.GPR[rd] = c.Reg.GPR[rs1]
		c.log.Debugf("TICK % 4d - %v<-%v | %v\n", c.Tick, isa.GetRegMnem(rd), isa.GetRegMnem(rs1), c.ReprRegVal(rd))
		return true
	}
}

// записывает 32-битное значение little-endian по адресу в regWithAddr
// возвращает true, когда все 4 байта записаны
func write32(c *CPU, stage *int, regWithAddr isa.Register, regSource isa.Register) bool {
	c.ensureDataSize(c.Reg.GPR[regWithAddr] + 3)

	val := c.Reg.GPR[regSource]
	switch *stage {
	case 1:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val)
		c.log.Debugf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n", c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
	case 2:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val >> 8)
		c.log.Debugf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n", c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
	case 3:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val >> 16)
		c.log.Debugf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n", c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
	case 4:
		c.memD[c.Reg.GPR[regWithAddr]] = byte(val >> 24)
		c.log.Debugf("TICK % 4d - memD[0x%X]<-%v | memD[0x%X]=0x%X\n", c.Tick, c.Reg.GPR[regWithAddr], isa.GetRegMnem(regSource), c.Reg.GPR[regWithAddr], c.memD[c.Reg.GPR[regWithAddr]])
		c.Reg.GPR[regWithAddr]++
		return true
	default:
		return true
	}
	*stage++
	return false
}

func read32(c *CPU, stage *int, regWithAddr isa.Register, regToStoreTo isa.Register) bool {
	c.ensureDataSize(c.Reg.GPR[regWithAddr] + 3)
	switch *stage {
	case 1:
		c.Reg.GPR[regToStoreTo] = uint32(c.memD[c.Reg.GPR[regWithAddr]])
		c.log.Debugf("TICK % 4d - %v<-memD[%X] | %v\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.Reg.GPR[regWithAddr]++
	case 2:
		c.Reg.GPR[regToStoreTo] |= uint32(c.memD[c.Reg.GPR[regWithAddr]]) << 8
		c.log.Debugf("TICK % 4d - %v<-memD[%X] | %v\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.Reg.GPR[regWithAddr]++
	case 3:
		c.Reg.GPR[regToStoreTo] |= uint32(c.memD[c.Reg.GPR[regWithAddr]]) << 16
		c.log.Debugf("TICK % 4d - %v<-memD[%X] | %v\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], c.ReprRegVal(regToStoreTo))
		c.Reg.GPR[regWithAddr]++
	case 4:
		c.Reg.GPR[regToStoreTo] |= uint32(c.memD[c.Reg.GPR[regWithAddr]]) << 24
		c.log.Debugf("TICK % 4d - %v<-memD[%X] | %v=% 4d/0x%X\n", c.Tick, isa.GetRegMnem(regToStoreTo),
			c.Reg.GPR[regWithAddr], isa.GetRegMnem(regToStoreTo), c.Reg.GPR[regToStoreTo], c.Reg.GPR[regToStoreTo])
		c.Reg.GPR[regWithAddr]++
	default:
		return true
	}
	*stage++
	return false
}
