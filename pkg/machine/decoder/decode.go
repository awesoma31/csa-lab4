package decoder

import (
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type Decoded struct {
	Opcode uint32 // 6 бит [31:26]
	Mode   uint32 // 5 бит [25:21]

	Rd  int // 4 бит [20:17]
	Rs1 int // 4 бит [16:13]
	Rs2 int // 4 бит [12:9]

}

func DecodeInstructionWord(w uint32) Decoded {
	const (
		opcodeMask = 0xFC000000
		modeMask   = 0x03E00000
		rdMask     = 0x001E0000
		rs1Mask    = 0x0001E000
		rs2Mask    = 0x00001E00
		low9Mask   = 0x000001FF
	)

	return Decoded{
		Opcode: (w & opcodeMask) >> isa.OpcodeOffset,
		Mode:   (w & modeMask) >> 21,

		Rd:  int((w & rdMask) >> 17),
		Rs1: int((w & rs1Mask) >> 13),
		Rs2: int((w & rs2Mask) >> 9),
	}
}

func Dec(w uint32) (op, mode uint32, rd, rs1, rs2 int) {
	d := DecodeInstructionWord(w)

	op = d.Opcode
	mode = d.Mode
	rd = d.Rd
	rs1 = d.Rs1
	rs2 = d.Rs2
	return
}
