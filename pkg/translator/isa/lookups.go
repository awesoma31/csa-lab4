package isa

// ────────────────── pretty-printing tables ───────────────────
var (
	opcodeMnemonics   = map[uint32]string{}
	amMnemonics       = map[uint32]string{}
	registerMnemonics = map[int]string{}
	portMnemonics     = map[int]string{}
)

func init() {
	// opcodes → strings
	for k, v := range map[uint32]string{
		OpNop:  "NOP",
		OpMov:  "MOV",
		OpPush: "PUSH",
		OpPop:  "POP",
		// OpNeg: "NEG",
		// OpNot: "NOT",
		OpHalt: "HALT",

		OpAdd: "ADD",
		OpSub: "SUB",
		OpMul: "MUL",
		OpDiv: "DIV",

		OpCmp: "CMP",
		OpAnd: "AND",

		OpIn: "IN", OpOut: "OUT",

		OpIRet:   "IRet",
		OpIntOn:  "IntOn",
		OpIntOff: "IntOff",

		OpJmp: "JMP",
		// OpCall: "CALL",
		// OpRet: "RET",
		OpJe: "JE", OpJne: "JNE", OpJg: "JG", OpJl: "JL",
		OpJge: "JGE", OpJle: "JLE", OpJa: "JA", OpJb: "JB",
		OpJae: "JAE", OpJbe: "JBE",
	} {
		opcodeMnemonics[k] = v
	}

	// address-modes → strings (only the ones that appear in code-gen)
	for k, v := range map[uint32]string{
		MvRegReg:         "MvRegReg",
		MvImmReg:         "MvImmReg",
		MvRegIndReg:      "MvRegIndReg",
		MvRegMemInd:      "MvRegMemInd",
		MvLowRegToRegInd: "MvLowRegToRegInd",
		MvMemReg:         "MvMemReg",
		MvRegMem:         "MvRegMem",
		MvRegLowToMem:    "MvRegLowMem",
		MvImmMem:         "MvImmMem",
		ImmReg:           "ImmReg",
		RegReg:           "RegReg",
		MvByteRegIndReg:  "MvLowRegIndReg",
		ByteM:            "Byte",
		DigitM:           "Digit",

		MathRRR: "MathRRR",
		MathRMR: "MathRMR",
		MathRIR: "MathRIR",

		JAbsAddr:      "JAbsAddr",
		SingleRegMode: "SingleReg",

		NoOperands: "NoOperands",
	} {
		amMnemonics[k] = v
	}

	// registers → strings (0-15)
	for k, v := range map[int]string{
		RA:       "RA",
		RM1:      "RM1",
		RM2:      "RM2",
		RAddr:    "RAddr",
		RD:       "RD",
		RC:       "RC",
		ZERO:     "zero",
		ROutData: "ROutData",
		ROutAddr: "ROutAddr",
		RInData:  "RInData",
		R6:       "R6",
		SpReg:    "SP",
		RT2:      "RT2",
		RT:       "RT",
	} {
		registerMnemonics[k] = v
	}

	for k, v := range map[int]string{
		PortD:  "port Digit",
		PortCh: "port Char",
	} {
		portMnemonics[k] = v
	}
}

func GetOpMnemonic(op uint32) string {
	return opcodeMnemonics[op]
}
func GetAMnemonic(m uint32) string {
	return amMnemonics[m]
}
func GetPortMnem(r int) string {
	return portMnemonics[r]
}
func GetRegMnem(r int) string {
	if r < 0 {
		return ""
	}
	if s, ok := registerMnemonics[r]; ok {
		return s
	}
	return "R?"
}
