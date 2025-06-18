package isa

// ─────────────────────────────────────────────────────────────
// Layout (big-endian view)
//
//	31          26 25       21 20    17 16    13 12     9   8-0
//	┌────────────┬──────────┬────────┬────────┬────────┬──────┐
//	│  OPCODE 6b │ MODE 5b  │  RD 4b │ RS1 4b │ RS2 4b │ ...  │
//	└────────────┴──────────┴────────┴────────┴────────┴──────┘
//
//	• subsequent 32-bit words are used for immediates / addresses
//	  depending on MODE.
//
// ─────────────────────────────────────────────────────────────
const OpcodeOffset = 26
const ModeOffset = 21
const DestRegOffset = 17
const Rs1Offset = 13
const Rs2Offset = 9

func EncodeInstructionWord(opcode, mode uint32, rd, rs1, rs2 Register) uint32 {
	word := opcode<<OpcodeOffset | mode<<ModeOffset

	if rd >= 0 {
		word |= uint32(rd&0xF) << DestRegOffset
	}
	if rs1 >= 0 {
		word |= uint32(rs1&0xF) << Rs1Offset
	}
	if rs2 >= 0 {
		word |= uint32(rs2&0xF) << Rs2Offset
	}

	return word
}
