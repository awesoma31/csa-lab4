package isa

// ──────────────────────── registers (4 bits) ─────────────────
const (
	Ra          = iota // 0
	RM1                // 1
	RM2                //2
	RAddr              //3
	R4                 //4
	ROutAddr           //5
	ROutData           //6
	R6                 //7
	RInData            //8
	RC                 //9
	SpReg              //10
	RT                 //11
	RT2                //12
	ZERO               // 13
	ONE                // 14
	placeholder        // 15
)

const (
	PortCh = 1
	PortD  = 0
)
