package isa

// ───────────────────── address-modes (5 bits) ────────────────
const (
	MvRegReg    uint32 = 0x00
	MvImmReg    uint32 = 0x01
	MvRegLowMem uint32 = 0x02
	MvRegIndReg uint32 = 0x03
	MvRegMemInd uint32 = 0x04 // mem[mem[addr]] <- reg
	RegMemFp    uint32 = 0x05
	MvMemReg    uint32 = 0x06
	MvRegMem    uint32 = 0x07
	MvImmMem    uint32 = 0x09

	IoMemReg uint32 = 0x0A
	ImmReg   uint32 = 0x0B

	StImmMode      uint32 = 0x0C
	CMPRegMode     uint32 = 0x0D
	RegReg         uint32 = 0x0E
	MvLowRegIndReg uint32 = 0x0F

	MathRRR uint32 = 0x10
	MathRMR uint32 = 0x11
	MathRIR uint32 = 0x12
	MathMMR uint32 = 0x13

	ByteM uint32 = 0x14
	WordM uint32 = 0x15

	JAbsAddr      uint32 = 0x18
	SingleRegMode uint32 = 0x1C

	InPortReg  uint32 = 0x1D
	OutRegPort uint32 = 0x1E

	NoOperands uint32 = 0x1F
)
