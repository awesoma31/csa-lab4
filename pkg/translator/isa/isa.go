package isa

// ─────────────────────────────────────────────────────────────
// Layout (big-endian view)
//
//	31          26 25       21 20    17 16    13 12     9   8-0
//	┌────────────┬──────────┬────────┬────────┬────────┬──────┐
//	│  OPCODE 6b │ MODE 5b  │  RD 4b │ RS1 4b │ RS2 4b │ ...  │
//	└────────────┴──────────┴────────┴────────┴────────┴──────┘
//
//	• subsequent 32-bit words are used for immediates / addresses
//	  depending on MODE.
//
// ─────────────────────────────────────────────────────────────
const OpcodeOffset = 26
const ModeOffset = 21
const DestRegOffset = 17
const Rs1Offset = 13
const Rs2Offset = 9

func EncodeInstructionWord(opcode, mode uint32, rd, rs1, rs2 int) uint32 {
	word := opcode<<OpcodeOffset | mode<<ModeOffset

	if rd >= 0 {
		word |= uint32(rd&0xF) << DestRegOffset
	}
	if rs1 >= 0 {
		word |= uint32(rs1&0xF) << Rs1Offset
	}
	if rs2 >= 0 {
		word |= uint32(rs2&0xF) << Rs2Offset
	}

	// lower 9 bits are left zero – filled later if the
	// particular address-mode packs an immediate offset here.
	return word
}

// ────────────────── pretty-printing tables ───────────────────
var (
	opcodeMnemonics   = map[uint32]string{}
	amMnemonics       = map[uint32]string{}
	registerMnemonics = map[int]string{}
)

func init() {
	// opcodes → strings
	for k, v := range map[uint32]string{
		OpNop: "NOP", OpMov: "MOV", OpPush: "PUSH", OpPop: "POP",
		OpNeg: "NEG", OpNot: "NOT", OpHalt: "HALT",

		OpAdd: "ADD", OpSub: "SUB", OpMul: "MUL", OpDiv: "DIV", OpCmp: "CMP",

		OpIn: "IN", OpOut: "OUT",

		OpJmp: "JMP", OpCall: "CALL", OpRet: "RET",
		OpJe: "JE", OpJne: "JNE", OpJg: "JG", OpJl: "JL",
		OpJge: "JGE", OpJle: "JLE", OpJa: "JA", OpJb: "JB",
		OpJae: "JAE", OpJbe: "JBE",
	} {
		opcodeMnemonics[k] = v
	}

	// address-modes → strings (only the ones that appear in code-gen)
	for k, v := range map[uint32]string{
		MvRegReg: "MV_REG_REG", MvImmReg: "MV_IMM_REG",
		// MvMemAbsReg: "MV_MEM_ABS_REG",
		MvRegMemAbs:   "MV_REG_MEM_ABS",
		MvSpOffsToReg: "SPOFFS_REG", RegMemFp: "REG_MEM_FP",
		MvMemReg: "MV_MEM_REG", MvRegMem: "MV_REG_MEM", MvMemMem: "MV_MEM_MEM",
		MvImmMem: "MV_IMM_MEM",

		MathRRR: "MATH_R_R_R", MathRMR: "MATH_R_M_R",
		MathRIR: "MATH_R_I_R", MathMMR: "MATH_M_M_R",

		JAbsAddr: "J_ABS_ADDR", SingleRegMode: "SINGLE_REG",
		InPortReg: "IN_PORT_REG", OutRegPort: "OUT_REG_PORT",
		NoOperands: "NO_OPERANDS",
	} {
		amMnemonics[k] = v
	}

	// registers → strings (0-15)
	for k, v := range map[int]string{
		RA: "RA", RM1: "RM1", RM2: "RM2", RAddr: "R3", R4: "R4",
		ROutAddr: "R_OUT_ADDR", RInAddr: "R_IN_ADDR",
		SpReg: "SP", FpReg: "FP",
	} {
		registerMnemonics[k] = v
	}
}

func GetOpMnemonic(op uint32) string { return opcodeMnemonics[op] }
func GetAMnemonic(m uint32) string   { return amMnemonics[m] }
func GetRegisterMnemonic(r int) string {
	if r < 0 {
		return ""
	}
	if s, ok := registerMnemonics[r]; ok {
		return s
	}
	return "R?"
}
