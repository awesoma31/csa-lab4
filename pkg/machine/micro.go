package machine

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type microStep func(c *CPU) bool // true → macro-instruction done

// [opcode][mode] → factory(rd,rs1,rs2)→microStep
var ucode [64][32]func(int, int, int) microStep

func init() {
	// ─── NOP (single tick) ───────────────────────────────
	ucode[isa.OpNop][isa.NoOperands] = func(_, _, _ int) microStep {
		return func(c *CPU) bool {
			c.reg.PC++
			return true
		}
	}

	// HALT
	ucode[isa.OpHalt][isa.NoOperands] = uHalt
	// func(i1, i2, i3 int) microStep {
	// 	fmt.Printf("Opc: %s\n", isa.GetMnemonic(isa.OpHalt))
	// 	return func(c *CPU) bool {
	// 		slog.Info("HALT, stopping simulation")
	// 		os.Exit(0)
	// 		return true
	// 	}
	// }

	// ─── MOV rd, [memAbs] (2 words, 3 ticks) ─────────────
	// ucode[isa.OpMov][isa.MvMemReg] = func(rd, _, _ int) microStep {
	// 	stage := 0
	// 	var addr uint32
	// 	return func(c *CPU) bool {
	// 		switch stage {
	// 		case 0: // T0 – fetch 2-nd word
	// 			c.reg.PC++
	// 			addr = c.memI[c.reg.PC]
	// 			stage = 1
	// 		case 1: // T1 – memory read
	// 			c.tmp = c.memD[addr]
	// 			stage = 2
	// 		case 2: // T2 – write-back, PC advance
	// 			//TODO: read 4 bytes
	// 			c.reg.GPR[rd] = uint32(c.tmp)
	// 			c.reg.PC += 2
	// 			return true
	// 		}
	// 		return false
	// 	}
	// }
	// ucode[isa.OpMov][isa.MvImmReg] = uMovImmReg
	// ucode[isa.OpMov][isa.MvRegReg] = uMovRegReg

	// ─── IN rd, #port (single-word, single-tick) ─────────
	// ucode[isa.OpIn][InPortReg] = func(rd, _, _ int) microStep {
	// 	return func(c *CPU) bool {
	// 		port := int((c.reg.IR >> 9) & 0xF)
	// 		val := c.io.Devs[port].Load()
	// 		c.reg.GPR[rd] = val
	// 		c.reg.PC++
	// 		return true
	// 	}
	// }

	// OUT #port, rs (single tick)
	// ucode[isa.OpOut][OutRegPort] = func(_, rs, _ int) microStep {
	// 	return func(c *CPU) bool {
	// 		port := int((c.reg.IR >> 9) & 0xF)
	// 		c.io.Devs[port].Store(byte(c.reg.GPR[rs]))
	// 		c.reg.PC++
	// 		return true
	// 	}
	// }
}

func uHalt(_ int, _ int, _ int) microStep {
	return func(c *CPU) bool {
		slog.Info("HALT, stopping simulation")
		os.Exit(0)
		return true
	}
}
func uMovImmReg(rd int, _ int, _ int) microStep {
	// fmt.Println("MV IMM REG DECODED")
	stage := 0
	var addr uint32
	return func(c *CPU) bool {
		switch stage {
		case 0: // T0 – fetch 2-nd word
			fmt.Println("MV stage 0")
			c.reg.PC++
			addr = c.memI[c.reg.PC]
			stage = 1
		case 1: // T1 – memory read
			fmt.Println("MV stage 1")
			c.tmp = c.memD[addr]
			stage = 2
		case 2: // T2 – write-back, PC advance
			fmt.Println("MV stage 2")
			//TODO: read 4 bytes
			c.reg.GPR[rd] = uint32(c.tmp)
			c.reg.PC += 2
			return true
		}
		return false
	}
}

func uMovRegReg(rd, rs1, _ int) microStep {
	return func(c *CPU) bool {
		fmt.Println("MV REG REG")
		//c.reg[rd] = c.reg[rs1]
		//c.reg.PC++
		return true
	}
}
