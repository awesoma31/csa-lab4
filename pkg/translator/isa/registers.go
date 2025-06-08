package isa

// ──────────────────────── registers (4 bits) ─────────────────
const (
	RA = iota // 0
	RM1
	RM2
	RAddr
	R4
	ROutAddr
	ROutData
	RInAddr
	RInData
	RC
	SpReg
	FpReg
	RT
	ZERO        // 13
	ONE         // 14
	placeholder // 15
)

const (
	PORT1 = R4
	PORT2 = FpReg
)
