package isa

// ──────────────────────── registers (4 bits) ─────────────────
type Register int

const (
	RA       Register = iota //0
	RM1                      //1
	RM2                      //2
	RAddr                    //3
	RD                       //4
	ROutAddr                 //5
	ROutData                 //6
	R6                       //7
	RInData                  //8
	RC                       //9
	SpReg                    //10
	RT                       //11
	RT2                      //12
	ZERO                     //13
	R7                       //14
	R8                       //15

	// not addressable
	RF1
	RF2
)

const (
	PortD  Register = 0
	PortCh          = 1
	PortL           = 2
)
