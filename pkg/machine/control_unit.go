package machine

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/awesoma31/csa-lab4/pkg/machine/decoder"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

const (
	IFetch int = 0
)

type ControlUnit struct {
	DP *DataPath

	IR uint32
	PC int
	SP uint32

	tick      int
	TickLimit int

	n, z, v, c          bool
	interuptionsEnabled bool

	step int
}

func (cu *ControlUnit) Simulate() {
	for cu.CurrentTick() < cu.TickLimit {
		cu.proccessNextTick()
	}
}

var decoded decoder.Decoded

func (cu *ControlUnit) proccessNextTick() {
	if cu.PC >= len(cu.DP.InstrMem) {
		log.Fatal("PC exceeds instruction memory size")
	}
	word := cu.DP.InstrMem[cu.PC]

	switch cu.step {
	case IFetch:
		decoded = decoder.DecodeInstructionWord(word)
		cu.IR = decoded.Opcode //latch IR

		slog.Info(decoded.String())
		cu.PC++
		cu.step++
	case 1:
		switch cu.IR {
		case isa.OpHalt:
			slog.Info("halt, stopping...")
			os.Exit(0)
		case isa.OpNop:
			cu.step = IFetch
			cu.PC++
			cu.NextTick()
		case isa.OpMov:
			slog.Info("MOV ")
			switch decoded.Mode {
			case isa.MvMemReg:
				//prbly should take 2 ticks
				slog.Info("MvMemReg")
				word = cu.DP.InstrMem[cu.PC]
				slog.Info(fmt.Sprintf("load from addr %X", word))
				cu.PC++

				// cu.DP.setRegister(decoded.Rd, cu.DP.DataMem[word])
			case isa.MvImmReg:
				slog.Info("Imm Reg")
				imm := cu.DP.InstrMem[cu.PC]
				slog.Info(fmt.Sprintf("imm val %v", imm))
				cu.PC++
				cu.DP.setRegister(decoded.Rd, imm)
				cu.step = IFetch

			default:
				cu.step = IFetch
				cu.PC++
			}
		default:
			cu.step = IFetch
			cu.PC++
			cu.NextTick()
		}
	default:
		cu.step = IFetch
	}

	cu.NextTick()
}

func (cu *ControlUnit) NextTick() {
	cu.tick++
}

func (cu *ControlUnit) CurrentTick() int {
	return cu.tick
}
