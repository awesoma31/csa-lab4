package decoder

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

type Decoded struct {
	Opcode uint32 // 6 бит [31:26]
	Mode   uint32 // 5 бит [25:21]

	Rd  int // 4 бит [20:17]
	Rs1 int // 4 бит [16:13]
	Rs2 int // 4 бит [12:9]

	Low9 uint16 // младшие 9 бит [8:0] – сдвиги/непосредственные данные
}

// DecodeInstructionWord извлекает все части согласно
// формату из шапки файла isa.go.
func DecodeInstructionWord(w uint32) Decoded {
	const (
		opcodeMask = 0xFC000000 // 111111 00..0
		modeMask   = 0x03E00000 // 000000 11111 ..0
		rdMask     = 0x001E0000 //              1111 ..0
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

		Low9: uint16(w & low9Mask),
	}
}

// String даёт удобное представление (опционально).
func (d Decoded) String() string {
	return fmt.Sprintf(
		"%s  %s   rd:%s  rs1:%s  rs2:%s}",
		isa.GetMnemonic(d.Opcode),
		isa.GetAMnemonic(d.Mode),
		isa.GetRegisterMnemonic(d.Rd),
		isa.GetRegisterMnemonic(d.Rs1),
		isa.GetRegisterMnemonic(d.Rs2),
	)
}
