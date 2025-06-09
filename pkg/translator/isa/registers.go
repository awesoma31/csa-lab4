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
	R6
	RInData
	RC
	SpReg
	R5
	RT
	ZERO        // 13
	ONE         // 14
	placeholder // 15
)

const (
	PortCh = 1
	PortD  = 0
)
