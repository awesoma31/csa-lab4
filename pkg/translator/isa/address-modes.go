package isa

// ───────────────────── address-modes (5 bits) ────────────────
const (
	// plain register/register/immediate moves
	MvRegReg uint32 = 0x00
	MvImmReg uint32 = 0x01
	// MvMemAbsReg uint32 = 0x02
	MvRegIndReg   uint32 = 0x03
	MvSpOffsToReg uint32 = 0x04 // rd ← [SP+offs]   /  [SP+offs] ← rs
	RegMemFp      uint32 = 0x05 // frame-pointer relative
	MvMemReg      uint32 = 0x06 // rd ← [addr]
	MvRegMem      uint32 = 0x07 // [addr] ← rs
	//TODO: delete
	MvMemMem uint32 = 0x08 // [dst] ← [src]
	MvImmMem uint32 = 0x09 // [addr] ← imm

	IoMemReg uint32 = 0x0A
	ImmReg   uint32 = 0x0B

	StImmMode  uint32 = 0x0C
	CMPRegMode uint32 = 0x0D
	RegReg     uint32 = 0x0E

	// arithmetic “ternary” forms
	MathRRR uint32 = 0x10 // rd ← rs1 op rs2
	MathRMR uint32 = 0x11 // rd ← rs1 op [addr]
	MathRIR uint32 = 0x12 // rd ← rs1 op imm
	MathMMR uint32 = 0x13 // rd ← [a1] op [a2]

	// jumps / calls get their own modes if you wish
	JAbsAddr      uint32 = 0x18 // the next word contains absolute address
	SingleRegMode uint32 = 0x1C // PUSH reg, POP reg, NEG reg, NOT reg

	InPortReg  uint32 = 0x1D // IN reg, #port
	OutRegPort uint32 = 0x1E // OUT #port, reg

	NoOperands uint32 = 0x1F // HALT, NOP, RET
)
