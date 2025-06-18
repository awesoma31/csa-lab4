package isa

const (
	OpNop  uint32 = 0x00
	OpMov  uint32 = 0x01
	OpPush uint32 = 0x02
	OpPop  uint32 = 0x03
	OpHalt uint32 = 0x06

	OpAdd uint32 = 0x10
	OpSub uint32 = 0x11
	OpMul uint32 = 0x12
	OpDiv uint32 = 0x13
	OpCmp uint32 = 0x14

	OpIn uint32 = 0x18
	// 0x19
	OpOut uint32 = 0x1A
	// 0x1B
	OpIntOn  uint32 = 0x1C
	OpIntOff uint32 = 0x1D

	OpJmp  uint32 = 0x20
	OpAnd  uint32 = 0x23
	OpIRet uint32 = 0x24

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
	OpJcc uint32 = 0x3A
	OpJcs uint32 = 0x3B
)
