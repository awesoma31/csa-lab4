// package codegen
//
// // =============================================================================
// // ISA Definition
// // =============================================================================
//
// // Format: [Opcode (8 bits)][Addressing Mode (4 bits)][(3 bits)][RegS1 (3 bits)][RegS2 (3 bits)][Unused (11 bits)]
// // Note: Actual bit packing depends on the mode. Not all fields are always used.
// func encodeInstructionWord(opcode, mode uint32, dest, s1, s2 int) uint32 {
// 	word := opcode << 24 // Opcode in bits 31-24
// 	word |= mode << 20   // Addressing mode in bits 23-20
//
// 	if opcode == OP_MOV && mode == SPOFFS_REG {
// 		// Special case for MOV with stack pointer offset mode
// 		// Uses bits 19-0 for offset (signed immediate)
// 		if dest != -1 {
// 			word |= uint32(dest&0x7) << 17 // RegD in bits 19-17 (3 bits)
// 		} else {
// 			panic("dest = -1 in SPOFFS")
// 		}
// 		if s1 != -1 {
// 			// Encode offset in remaining bits (16-0)
// 			// Assuming regS1 contains the offset value in this mode
// 			offset := uint32(s1) & 0x1FFFFF // 21-bit offset (signed)
// 			word |= offset
// 		} else {
// 			panic("s1 = -1 in SPOFFS")
// 		}
// 		return word
// 	}
//
// 	// Standard encoding for other instructions/modes
// 	if dest != -1 {
// 		word |= uint32(dest) << 17 // RegD in bits 19-17
// 	}
// 	if s1 != -1 {
// 		word |= uint32(s1) << 14 // RegS1 in bits 16-14
// 	}
// 	if s2 != -1 {
// 		word |= uint32(s2) << 11 // RegS2 in bits 13-11
// 	}
//
// 	return word
// }
//
// const (
// 	RA = iota
// 	RM1
// 	RM2
// 	R3
// 	R4
// 	RPRINT
// 	RREAD
// 	R7
//
// 	// SP_REG Special purpose registers (conceptual, might not be directly addressable by instructions)
// 	SP_REG // Stack Pointer
// 	FP_REG // Frame Pointer
// 	// Program Counter is usually implicit
// )
//
// // Opcode constants (1 byte: 0x00 - 0xFF)
// const (
// 	OP_NOP  uint32 = 0x00
// 	OP_MOV  uint32 = 0x01
// 	OP_ADD  uint32 = 0x02
// 	OP_SUB  uint32 = 0x03
// 	OP_MUL  uint32 = 0x04
// 	OP_DIV  uint32 = 0x05
// 	OP_NEG  uint32 = 0x06
// 	OP_NOT  uint32 = 0x07
// 	OP_HALT uint32 = 0x08
//
// 	OP_PUSH uint32 = 0x10
// 	OP_POP  uint32 = 0x11
//
// 	OP_CALL uint32 = 0x21
// 	OP_RET  uint32 = 0x22
//
// 	OP_IN  uint32 = 0x30
// 	OP_OUT uint32 = 0x31
//
// 	OP_JMP uint32 = 0x20
// 	OP_JE  uint32 = 0x40
// 	OP_JNE uint32 = 0x41
// 	OP_JG  uint32 = 0x42
// 	OP_JL  uint32 = 0x43
// 	OP_JGE uint32 = 0x44
// 	OP_JLE uint32 = 0x45
// 	OP_JA  uint32 = 0x46
// 	OP_JB  uint32 = 0x47
// 	OP_JAE uint32 = 0x48
// 	OP_JBE uint32 = 0x49
//
// 	OP_CMP uint32 = 0x50
// )
//
// var opcodeMnemonics = map[uint32]string{}
// var amMnemonics = map[uint32]string{}
// var registerMnemonics = map[int]string{}
//
// func init() {
// 	// OPCODE MNEMONIC
// 	opcodeMnemonics[OP_HALT] = "HALT"
// 	opcodeMnemonics[OP_MOV] = "MOV"
// 	opcodeMnemonics[OP_ADD] = "ADD"
// 	opcodeMnemonics[OP_SUB] = "SUB"
// 	opcodeMnemonics[OP_MUL] = "MUL"
// 	opcodeMnemonics[OP_DIV] = "DIV"
// 	opcodeMnemonics[OP_NEG] = "NEG"
// 	opcodeMnemonics[OP_NOT] = "NOT"
// 	opcodeMnemonics[OP_NOP] = "NO_OP"
//
// 	opcodeMnemonics[OP_PUSH] = "PUSH"
// 	opcodeMnemonics[OP_POP] = "POP"
//
// 	opcodeMnemonics[OP_JMP] = "JMP"
// 	opcodeMnemonics[OP_CALL] = "CALL"
// 	opcodeMnemonics[OP_RET] = "RET"
//
// 	opcodeMnemonics[OP_IN] = "IN"
// 	opcodeMnemonics[OP_OUT] = "OUT"
//
// 	opcodeMnemonics[OP_JE] = "JE"
// 	opcodeMnemonics[OP_JNE] = "JNE"
// 	opcodeMnemonics[OP_JG] = "JG"
// 	opcodeMnemonics[OP_JL] = "JL"
// 	opcodeMnemonics[OP_JGE] = "JGE"
// 	opcodeMnemonics[OP_JLE] = "JLE"
// 	opcodeMnemonics[OP_JA] = "JA"
// 	opcodeMnemonics[OP_JB] = "JB"
// 	opcodeMnemonics[OP_JAE] = "JAE"
// 	opcodeMnemonics[OP_JBE] = "JBE"
// 	opcodeMnemonics[OP_CMP] = "CMP"
//
// 	// ADDRESS MODE MNEMONIC
// 	amMnemonics[REG_REG] = "REG_REG"
// 	amMnemonics[IMM_REG] = "IMM_REG"
// 	amMnemonics[IMM_MEM] = "IMM_MEM"
// 	amMnemonics[MEM_ABS_REG] = "MEM_ABS_REG"
// 	amMnemonics[REG_MEM_ABS] = "REG_MEM_ABS"
// 	//TODO:
// 	// amMnemonics[MEM_FP_REG] = "MEM_FP_REG"
// 	amMnemonics[REG_MEM_FP] = "REG_MEM_FP"
// 	amMnemonics[SINGLE_REG] = "SINGLE_REG"
// 	amMnemonics[IMM_PORT_REG] = "IMM_PORT_REG"
// 	amMnemonics[REG_PORT_IMM] = "REG_PORT_IMM"
// 	amMnemonics[ABS_ADDR] = "ABS_ADDR"
// 	amMnemonics[NO_OPERANDS] = "NO_OPERANDS"
// 	amMnemonics[MEM_REG] = "MEM_REG"
// 	amMnemonics[SPOFFS_REG] = "SPOFFS_REG"
// 	amMnemonics[MEM_MEM] = "MEM_MEM"
// 	amMnemonics[REG_MEM] = "REG_MEM"
// 	amMnemonics[MATH_R_R_R] = "MATH_R_R_R"
// 	amMnemonics[MATH_R_M_R] = "MATH_R_M_R"
// 	amMnemonics[MATH_M_M_R] = "MATH_M_M_R"
// 	amMnemonics[MATH_R_I_R] = "MATH_R_I_R"
// 	// amMnemonics[JE] = "JE_AM"
// 	// amMnemonics[JNE] = "JNE_AM"
// 	// amMnemonics[JG] = "JG_AM"
// 	// amMnemonics[JL] = "JL_AM"
// 	// amMnemonics[JGE] = "JGE_AM"
// 	// amMnemonics[JLE] = "JLE_AM"
// 	// amMnemonics[JMP_ABS] = "JMP_ABS"
// 	// amMnemonics[JMP_REG] = "JMP_REG"
// 	// amMnemonics[JMP_MEM] = "JMP_MEM"
// 	amMnemonics[CMP] = "CMP_AM"
// 	amMnemonics[CALL_ABS] = "CALL_ABS"
// 	// amMnemonics[CALL_REG] = "CALL_REG"
// 	// amMnemonics[CALL_MEM] = "CALL_MEM"
// 	// amMnemonics[RET_IMM] = "RET_IMM"
//
// 	// REGISTER MNEMONIC
// 	registerMnemonics[RA] = "RA"
// 	registerMnemonics[RM1] = "RM1"
// 	registerMnemonics[RM2] = "RM2"
// 	registerMnemonics[R3] = "R3"
// 	registerMnemonics[R4] = "R4"
// 	registerMnemonics[RPRINT] = "RPRINT"
// 	registerMnemonics[RREAD] = "RREAD"
// 	registerMnemonics[R7] = "R7"
// 	registerMnemonics[SP_REG] = "SP"
// 	registerMnemonics[FP_REG] = "FP"
// }
//
// func GetMnemonic(opcode uint32) string {
// 	if mnemonic, ok := opcodeMnemonics[opcode]; ok {
// 		return mnemonic
// 	}
// 	return "UNKNOWN"
// }
//
// func GetAMnemonic(mode uint32) string {
// 	if mnemonic, ok := amMnemonics[mode]; ok {
// 		return mnemonic
// 	}
// 	return "UNKNOWN_AM"
// }
//
// func GetRegisterMnemonic(reg int) string {
// 	if reg == -1 {
// 		return ""
// 	}
// 	if mnemonic, ok := registerMnemonics[reg]; ok {
// 		return mnemonic
// 	}
// 	return "UNKNOWN_REG"
// }
//
// // Addressing Mode / Operand Type constants (4 bits: 0x0 - 0xF)
// const (
// 	REG_REG     uint32 = 0x0 // Register to Register (e.g., ADD R0, R1)
// 	IMM_REG     uint32 = 0x1 // Immediate to Register (e.g., MOV R0, #123)
// 	MEM_ABS_REG uint32 = 0x2 // Absolute Memory to Register (e.g., MOV R0, [0x1000])
// 	REG_MEM_ABS uint32 = 0x3 // Register to Absolute Memory (e.g., MOV [0x1000], R0)
// 	// MEM_FP_REG   uint32 = 0x4 // FP+Offset Memory to Register (e.g., MOV R0, [FP+8])
// 	REG_MEM_FP   uint32 = 0x5 // Register to FP+Offset Memory (e.g., MOV [FP+8], R0)
// 	SINGLE_REG   uint32 = 0x6 // Single Register operand (e.g., PUSH R0, NEG R0)
// 	IMM_PORT_REG uint32 = 0x7 // Immediate (port number) to Register (e.g., IN R0, #PORT_ID)
// 	REG_PORT_IMM uint32 = 0x8 // Register to Immediate (port number) (e.g., OUT #PORT_ID, R0)
// 	ABS_ADDR     uint32 = 0x9 // Absolute Address (for JMP, CALL)
// 	NO_OPERANDS  uint32 = 0xA // No operands (e.g., HALT, RET)
//
// 	MEM_REG    uint32 = 0xB // Memory (absolute address) to Register (MOV rd, [addr])
// 	SPOFFS_REG uint32 = 0xC // Stack Pointer Offset to Register (MOV rd, [(sp)+offs])
// 	MEM_MEM    uint32 = 0xD // Memory to Memory (MOV [dest_addr], [source_addr])
// 	REG_MEM    uint32 = 0xE // Register to Memory (MOV [addr], rs)
//
// 	MATH_R_R_R uint32 = 0x1 // Math op: Reg, Reg, Reg (ADD rd, rs1, rs2) - This is a placeholder, needs careful bit packing for 3 regs
// 	MATH_R_M_R uint32 = 0x2 // Math op: Reg, Mem, Reg (ADD rd, rs1, [addr])
// 	MATH_M_M_R uint32 = 0x3 // Math op: Mem, Mem, Reg (ADD rd, [addr1], [addr2])
// 	MATH_R_I_R uint32 = 0x4
//
// 	// JE  uint32 = 0x12 // Jump if Equal
// 	// JNE uint32 = 0x13 // Jump if Not Equal
// 	// JG  uint32 = 0x14 // Jump if Greater (signed)
// 	// JL  uint32 = 0x15 // Jump if Less (signed)
// 	// JGE uint32 = 0x16 // Jump if Greater or Equal (signed)
// 	// JLE uint32 = 0x17 // Jump if Less or Equal (signed)
// 	// JA  uint32 = 0x18 // Jump if Above (unsigned)
// 	// JB  uint32 = 0x19 // Jump if Below (unsigned)
// 	// JAE uint32 = 0x1A // Jump if Above or Equal (unsigned)
// 	// JBE uint32 = 0x1B // Jump if Below or Equal (unsigned)
//
// 	// JMP_ABS uint32 = 0x1C // Jump Absolute (JMP addr)
// 	// JMP_REG uint32 = 0x1D // Jump Register (JMP reg)
// 	// JMP_MEM uint32 = 0x1E // Jump Memory (JMP [addr])
//
// 	CMP uint32 = 0x1F // Compare (CMP rs1, rs2)
//
// 	CALL_ABS uint32 = 0x20 // Call Absolute (CALL addr)
// 	// CALL_REG uint32 = 0x21 // Call Register (CALL reg)
// 	// CALL_MEM uint32 = 0x22 // Call Memory (CALL [addr])
//
// 	// RET_IMM uint32 = 0x23 // Return with Immediate offset (RET imm)
// 	IMM_MEM uint32 = 0x24
// )

package codegen

// ─────────────────────────────────────────────────────────────
// Layout (big-endian view)
//
//  31          26 25       21 20    17 16    13 12     9   8-0
//  ┌────────────┬──────────┬────────┬────────┬────────┬──────┐
//  │  OPCODE 6b │ MODE 5b  │  RD 4b │ RS1 4b │ RS2 4b │ ...  │
//  └────────────┴──────────┴────────┴────────┴────────┴──────┘
//
//  • subsequent 32-bit words are used for immediates / addresses
//    depending on MODE.
// ─────────────────────────────────────────────────────────────

// ──────────────────────── registers (4 bits) ─────────────────
const (
	RA = iota
	RM1
	RM2
	R3
	R4
	R_OUT_ADDR
	R_OUT_DATA
	R_IN_ADDR
	R_IN_DATA
	R_COUNTER
	SP_REG
	FP_REG
	R5
	R6
)

// ──────────────────────── op-codes (6 bits) ──────────────────
const (
	// 0x00-0x0F: data move & misc
	OP_NOP  uint32 = 0x00
	OP_MOV  uint32 = 0x01
	OP_PUSH uint32 = 0x02
	OP_POP  uint32 = 0x03
	OP_NEG  uint32 = 0x04
	OP_NOT  uint32 = 0x05
	OP_HALT uint32 = 0x06

	// 0x10-0x1F: arithmetic / logical
	OP_ADD uint32 = 0x10
	OP_SUB uint32 = 0x11
	OP_MUL uint32 = 0x12
	OP_DIV uint32 = 0x13
	OP_CMP uint32 = 0x14

	// 0x18-0x1F: I/O
	OP_IN  uint32 = 0x18
	OP_OUT uint32 = 0x19

	// 0x20-0x2F: control-flow (unconditional & calls)
	OP_JMP  uint32 = 0x20
	OP_CALL uint32 = 0x21
	OP_RET  uint32 = 0x22

	// 0x30-0x3F: conditional jumps (depend on flags)
	OP_JE  uint32 = 0x30
	OP_JNE uint32 = 0x31
	OP_JG  uint32 = 0x32
	OP_JL  uint32 = 0x33
	OP_JGE uint32 = 0x34
	OP_JLE uint32 = 0x35
	OP_JA  uint32 = 0x36
	OP_JB  uint32 = 0x37
	OP_JAE uint32 = 0x38
	OP_JBE uint32 = 0x39
)

// ───────────────────── address-modes (5 bits) ────────────────
const (
	// plain register/register/immediate moves
	MV_REG_REG     uint32 = 0x00
	MV_IMM_REG     uint32 = 0x01
	MV_MEM_ABS_REG uint32 = 0x02
	MV_REG_MEM_ABS uint32 = 0x03
	SPOFFS_REG     uint32 = 0x04 // rd ← [SP+offs]   /  [SP+offs] ← rs
	REG_MEM_FP     uint32 = 0x05 // frame-pointer relative
	MV_MEM_REG     uint32 = 0x06 // rd ← [addr]
	MV_REG_MEM     uint32 = 0x07 // [addr] ← rs
	MV_MEM_MEM     uint32 = 0x08 // [dst] ← [src]
	MV_IMM_MEM     uint32 = 0x09 // [addr] ← imm

	IO_MEM_REG uint32 = 0x0A
	IO_IMM_REG uint32 = 0x0B

	// arithmetic “ternary” forms
	MATH_R_R_R uint32 = 0x10 // rd ← rs1 op rs2
	MATH_R_M_R uint32 = 0x11 // rd ← rs1 op [addr]
	MATH_R_I_R uint32 = 0x12 // rd ← rs1 op imm
	MATH_M_M_R uint32 = 0x13 // rd ← [a1] op [a2]

	// jumps / calls get their own modes if you wish
	J_ABS_ADDR   uint32 = 0x18 // the next word contains absolute address
	SINGLE_REG   uint32 = 0x1C // PUSH reg, POP reg, NEG reg, NOT reg
	IN_PORT_REG  uint32 = 0x1D // IN reg, #port
	OUT_REG_PORT uint32 = 0x1E // OUT #port, reg

	NO_OPERANDS uint32 = 0x1F // HALT, NOP, RET
)

// ─────────────────── encode helper (common path) ─────────────
func encodeInstructionWord(opcode, mode uint32, rd, rs1, rs2 int) uint32 {
	word := opcode<<26 | mode<<21

	if rd >= 0 {
		word |= uint32(rd&0xF) << 17
	}
	if rs1 >= 0 {
		word |= uint32(rs1&0xF) << 13
	}
	if rs2 >= 0 {
		word |= uint32(rs2&0xF) << 9
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
		OP_NOP: "NOP", OP_MOV: "MOV", OP_PUSH: "PUSH", OP_POP: "POP",
		OP_NEG: "NEG", OP_NOT: "NOT", OP_HALT: "HALT",

		OP_ADD: "ADD", OP_SUB: "SUB", OP_MUL: "MUL", OP_DIV: "DIV", OP_CMP: "CMP",

		OP_IN: "IN", OP_OUT: "OUT",

		OP_JMP: "JMP", OP_CALL: "CALL", OP_RET: "RET",
		OP_JE: "JE", OP_JNE: "JNE", OP_JG: "JG", OP_JL: "JL",
		OP_JGE: "JGE", OP_JLE: "JLE", OP_JA: "JA", OP_JB: "JB",
		OP_JAE: "JAE", OP_JBE: "JBE",
	} {
		opcodeMnemonics[k] = v
	}

	// address-modes → strings (only the ones that appear in code-gen)
	for k, v := range map[uint32]string{
		MV_REG_REG: "MV_REG_REG", MV_IMM_REG: "MV_IMM_REG",
		MV_MEM_ABS_REG: "MV_MEM_ABS_REG", MV_REG_MEM_ABS: "MV_REG_MEM_ABS",
		SPOFFS_REG: "SPOFFS_REG", REG_MEM_FP: "REG_MEM_FP",
		MV_MEM_REG: "MV_MEM_REG", MV_REG_MEM: "MV_REG_MEM", MV_MEM_MEM: "MV_MEM_MEM",
		MV_IMM_MEM: "MV_IMM_MEM",

		MATH_R_R_R: "MATH_R_R_R", MATH_R_M_R: "MATH_R_M_R",
		MATH_R_I_R: "MATH_R_I_R", MATH_M_M_R: "MATH_M_M_R",

		J_ABS_ADDR: "J_ABS_ADDR", SINGLE_REG: "SINGLE_REG",
		IN_PORT_REG: "IN_PORT_REG", OUT_REG_PORT: "OUT_REG_PORT",
		NO_OPERANDS: "NO_OPERANDS",
	} {
		amMnemonics[k] = v
	}

	// registers → strings (0-15)
	for k, v := range map[int]string{
		RA: "RA", RM1: "RM1", RM2: "RM2", R3: "R3", R4: "R4",
		R_OUT_ADDR: "R_OUT_ADDR", R_IN_ADDR: "R_IN_ADDR",
		SP_REG: "SP", FP_REG: "FP",
	} {
		registerMnemonics[k] = v
	}
}

// helper getters (unchanged interface)
func GetMnemonic(op uint32) string { return opcodeMnemonics[op] }
func GetAMnemonic(m uint32) string { return amMnemonics[m] }
func GetRegisterMnemonic(r int) string {
	if r < 0 {
		return ""
	}
	if s, ok := registerMnemonics[r]; ok {
		return s
	}
	return "R?"
}
