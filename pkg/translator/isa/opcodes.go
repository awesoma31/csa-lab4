package isa

// ──────────────────────── op-codes (6 bits) ──────────────────
const (
	// 0x00-0x0F: data move & misc
	OpNop  uint32 = 0x00
	OpMov  uint32 = 0x01
	OpPush uint32 = 0x02
	OpPop  uint32 = 0x03
	OpNeg  uint32 = 0x04
	OpNot  uint32 = 0x05
	OpHalt uint32 = 0x06

	// 0x10-0x1F: arithmetic / logical
	OpAdd uint32 = 0x10
	OpSub uint32 = 0x11
	OpMul uint32 = 0x12
	OpDiv uint32 = 0x13
	OpCmp uint32 = 0x14

	// 0x18-0x1F: I/O
	OpIn   uint32 = 0x18
	OpOut  uint32 = 0x19
	OpOutD uint32 = 0x1A

	// 0x20-0x2F: control-flow (unconditional & calls)
	OpJmp  uint32 = 0x20
	OpCall uint32 = 0x21
	OpRet  uint32 = 0x22
	OpAnd  uint32 = 0x23
	OpIRet uint32 = 0x24

	// 0x30-0x3F: conditional jumps (depend on flags)
	OpJe  uint32 = 0x30
	OpJne uint32 = 0x31
	OpJg  uint32 = 0x32
	OpJl  uint32 = 0x33
	OpJge uint32 = 0x34
	OpJle uint32 = 0x35
	OpJa  uint32 = 0x36
	OpJb  uint32 = 0x37
	OpJae uint32 = 0x38
	OpJbe uint32 = 0x39
)
