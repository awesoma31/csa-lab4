package isa

import "maps"

// ────────────────── pretty-printing tables ───────────────────
var (
	opcodeMnemonics   = map[uint32]string{}
	amMnemonics       = map[uint32]string{}
	registerMnemonics = map[Register]string{}
	portMnemonics     = map[Register]string{}
)

func init() {
	maps.Copy(opcodeMnemonics, map[uint32]string{
		OpNop:  "NOP",
		OpMov:  "MOV",
		OpPush: "PUSH",
		OpPop:  "POP",

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

		OpJe: "JE", OpJne: "JNE", OpJg: "JG", OpJl: "JL",
		OpJge: "JGE", OpJle: "JLE", OpJa: "JA", OpJb: "JB",
		OpJae: "JAE", OpJbe: "JBE", OpJcc: "JCC", OpJcs: "JCS",
	})

	maps.Copy(amMnemonics, map[uint32]string{
		MvRegReg:          "MvRegReg",
		MvImmReg:          "MvImmReg",
		MvRegIndToReg:     "MvRegIndToReg",
		MvRegMemInd:       "MvRegToMemInd",
		MvLowRegToRegInd:  "MvLowRegToRegInd",
		MvMemReg:          "MvMemReg",
		MvRegMem:          "MvRegMem",
		MvRegLowToMem:     "MvRegLowMem",
		MvImmMem:          "MvImmMem",
		MvRegToRegInd:     "MvRegToRegInd",
		ImmReg:            "ImmReg",
		RegReg:            "RegReg",
		MvByteRegIndToReg: "MvLowRegIndToReg",
		ByteM:             "Byte",
		DigitM:            "Digit",
		LongM:             "Long",

		MathRRR: "MathRRR",
		MathRMR: "MathRMR",
		MathRIR: "MathRIR",

		JAbsAddr:      "JAbsAddr",
		SingleRegMode: "SingleReg",

		NoOperands: "NoOperands",
	})

	maps.Copy(registerMnemonics, map[Register]string{
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
		R7:       "R7",
		R8:       "R8",
		SpReg:    "SP",
		RT2:      "RT2",
		RT:       "RT",
		RF1:      "RF1",
		RF2:      "RF2",
	})

	maps.Copy(portMnemonics, map[Register]string{
		PortD:  "port Digit",
		PortCh: "port Char",
		PortL:  "port Long",
	})
}

func GetOpMnemonic(op uint32) string {
	return opcodeMnemonics[op]
}
func GetAMnemonic(m uint32) string {
	return amMnemonics[m]
}
func GetPortMnem(r Register) string {
	return portMnemonics[r]
}
func GetRegMnem(r Register) string {
	if r < 0 {
		return ""
	}
	if s, ok := registerMnemonics[r]; ok {
		return s
	}
	return "R?"
}
