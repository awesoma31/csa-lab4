package codegen

// =============================================================================
// ISA Definition
// =============================================================================

// Format: [Opcode (8 bits)][Addressing Mode (4 bits)][(3 bits)][RegS1 (3 bits)][RegS2 (3 bits)][Unused (11 bits)]
// Note: Actual bit packing depends on the mode. Not all fields are always used.
func encodeInstructionWord(opcode, mode uint32, dest, s1, s2 int) uint32 {
	word := opcode << 24 // Opcode in bits 31-24
	word |= mode << 20   // Addressing mode in bits 23-20

	if opcode == OP_MOV && mode == AM_SPOFFS_REG {
		// Special case for MOV with stack pointer offset mode
		// Uses bits 19-0 for offset (signed immediate)
		if dest != -1 {
			word |= uint32(dest&0x7) << 17 // RegD in bits 19-17 (3 bits)
		} else {
			panic("dest = -1 in SPOFFS")
		}
		if s1 != -1 {
			// Encode offset in remaining bits (16-0)
			// Assuming regS1 contains the offset value in this mode
			offset := uint32(s1) & 0x1FFFFF // 21-bit offset (signed)
			word |= offset
		} else {
			panic("s1 = -1 in SPOFFS")
		}
		return word
	}

	// Standard encoding for other instructions/modes
	if dest != -1 {
		word |= uint32(dest) << 17 // RegD in bits 19-17
	}
	if s1 != -1 {
		word |= uint32(s1) << 14 // RegS1 in bits 16-14
	}
	if s2 != -1 {
		word |= uint32(s2) << 11 // RegS2 in bits 13-11
	}

	return word
}

const (
	RA = iota
	RM1
	RM2
	R3
	R4
	RPRINT
	RREAD
	R7

	// SP_REG Special purpose registers (conceptual, might not be directly addressable by instructions)
	SP_REG // Stack Pointer
	FP_REG // Frame Pointer
	// Program Counter is usually implicit
)

// Opcode constants (1 byte: 0x00 - 0xFF)
const (
	OP_NOP  uint32 = 0x00
	OP_MOV  uint32 = 0x01
	OP_ADD  uint32 = 0x02
	OP_SUB  uint32 = 0x03
	OP_MUL  uint32 = 0x04
	OP_DIV  uint32 = 0x05
	OP_NEG  uint32 = 0x06
	OP_NOT  uint32 = 0x07
	OP_HALT uint32 = 0x08

	OP_PUSH uint32 = 0x10
	OP_POP  uint32 = 0x11

	OP_JMP  uint32 = 0x20
	OP_CALL uint32 = 0x21
	OP_RET  uint32 = 0x22

	OP_IN  uint32 = 0x30
	OP_OUT uint32 = 0x31

	OP_JE  uint32 = 0x40
	OP_JNE uint32 = 0x41
	OP_JG  uint32 = 0x42
	OP_JL  uint32 = 0x43
	OP_JGE uint32 = 0x44
	OP_JLE uint32 = 0x45
	OP_JA  uint32 = 0x46
	OP_JB  uint32 = 0x47
	OP_JAE uint32 = 0x48
	OP_JBE uint32 = 0x49
	OP_CMP uint32 = 0x50
)

var opcodeMnemonics = map[uint32]string{}
var amMnemonics = map[uint32]string{}
var registerMnemonics = map[int]string{}

func init() {
	// OPCODE MNEMONIC
	opcodeMnemonics[OP_HALT] = "HALT"
	opcodeMnemonics[OP_MOV] = "MOV"
	opcodeMnemonics[OP_ADD] = "ADD"
	opcodeMnemonics[OP_SUB] = "SUB"
	opcodeMnemonics[OP_MUL] = "MUL"
	opcodeMnemonics[OP_DIV] = "DIV"
	opcodeMnemonics[OP_NEG] = "NEG"
	opcodeMnemonics[OP_NOT] = "NOT"
	opcodeMnemonics[OP_NOP] = "NO_OP"

	opcodeMnemonics[OP_PUSH] = "PUSH"
	opcodeMnemonics[OP_POP] = "POP"

	opcodeMnemonics[OP_JMP] = "JMP"
	opcodeMnemonics[OP_CALL] = "CALL"
	opcodeMnemonics[OP_RET] = "RET"

	opcodeMnemonics[OP_IN] = "IN"
	opcodeMnemonics[OP_OUT] = "OUT"

	opcodeMnemonics[OP_JE] = "JE"
	opcodeMnemonics[OP_JNE] = "JNE"
	opcodeMnemonics[OP_JG] = "JG"
	opcodeMnemonics[OP_JL] = "JL"
	opcodeMnemonics[OP_JGE] = "JGE"
	opcodeMnemonics[OP_JLE] = "JLE"
	opcodeMnemonics[OP_JA] = "JA"
	opcodeMnemonics[OP_JB] = "JB"
	opcodeMnemonics[OP_JAE] = "JAE"
	opcodeMnemonics[OP_JBE] = "JBE"
	opcodeMnemonics[OP_CMP] = "CMP"

	// ADDRESS MODE MNEMONIC
	amMnemonics[AM_REG_REG] = "REG_REG"
	amMnemonics[AM_IMM_REG] = "IMM_REG"
	amMnemonics[AM_IMM_MEM] = "IMM_MEM"
	amMnemonics[AM_MEM_ABS_REG] = "MEM_ABS_REG"
	amMnemonics[AM_REG_MEM_ABS] = "REG_MEM_ABS"
	//TODO:
	// amMnemonics[AM_MEM_FP_REG] = "MEM_FP_REG"
	amMnemonics[AM_REG_MEM_FP] = "REG_MEM_FP"
	amMnemonics[AM_SINGLE_REG] = "SINGLE_REG"
	amMnemonics[AM_IMM_PORT_REG] = "IMM_PORT_REG"
	amMnemonics[AM_REG_PORT_IMM] = "REG_PORT_IMM"
	amMnemonics[AM_ABS_ADDR] = "ABS_ADDR"
	amMnemonics[AM_NO_OPERANDS] = "NO_OPERANDS"
	amMnemonics[AM_MEM_REG] = "MEM_REG"
	amMnemonics[AM_SPOFFS_REG] = "SPOFFS_REG"
	amMnemonics[AM_MEM_MEM] = "MEM_MEM"
	amMnemonics[AM_REG_MEM] = "REG_MEM"
	amMnemonics[MATH_R_R_R] = "MATH_R_R_R"
	amMnemonics[MATH_R_M_R] = "MATH_R_M_R"
	amMnemonics[MATH_M_M_R] = "MATH_M_M_R"
	amMnemonics[MATH_R_I_R] = "MATH_R_I_R"
	amMnemonics[AM_JE] = "JE_AM"
	amMnemonics[AM_JNE] = "JNE_AM"
	amMnemonics[AM_JG] = "JG_AM"
	amMnemonics[AM_JL] = "JL_AM"
	amMnemonics[AM_JGE] = "JGE_AM"
	amMnemonics[AM_JLE] = "JLE_AM"
	amMnemonics[AM_JMP_ABS] = "JMP_ABS"
	amMnemonics[AM_JMP_REG] = "JMP_REG"
	amMnemonics[AM_JMP_MEM] = "JMP_MEM"
	amMnemonics[AM_CMP] = "CMP_AM"
	amMnemonics[AM_CALL_ABS] = "CALL_ABS"
	amMnemonics[AM_CALL_REG] = "CALL_REG"
	amMnemonics[AM_CALL_MEM] = "CALL_MEM"
	amMnemonics[AM_RET_IMM] = "RET_IMM"

	// REGISTER MNEMONIC
	registerMnemonics[RA] = "RA"
	registerMnemonics[RM1] = "RM1"
	registerMnemonics[RM2] = "RM2"
	registerMnemonics[R3] = "R3"
	registerMnemonics[R4] = "R4"
	registerMnemonics[RPRINT] = "RPRINT"
	registerMnemonics[RREAD] = "RREAD"
	registerMnemonics[R7] = "R7"
	registerMnemonics[SP_REG] = "SP"
	registerMnemonics[FP_REG] = "FP"
}

func GetMnemonic(opcode uint32) string {
	if mnemonic, ok := opcodeMnemonics[opcode]; ok {
		return mnemonic
	}
	return "UNKNOWN"
}

func GetAMnemonic(mode uint32) string {
	if mnemonic, ok := amMnemonics[mode]; ok {
		return mnemonic
	}
	return "UNKNOWN_AM"
}

func GetRegisterMnemonic(reg int) string {
	if reg == -1 {
		return ""
	}
	if mnemonic, ok := registerMnemonics[reg]; ok {
		return mnemonic
	}
	return "UNKNOWN_REG"
}

// Addressing Mode / Operand Type constants (4 bits: 0x0 - 0xF)
const (
	AM_REG_REG     uint32 = 0x0 // Register to Register (e.g., ADD R0, R1)
	AM_IMM_REG     uint32 = 0x1 // Immediate to Register (e.g., MOV R0, #123)
	AM_MEM_ABS_REG uint32 = 0x2 // Absolute Memory to Register (e.g., MOV R0, [0x1000])
	AM_REG_MEM_ABS uint32 = 0x3 // Register to Absolute Memory (e.g., MOV [0x1000], R0)
	// AM_MEM_FP_REG   uint32 = 0x4 // FP+Offset Memory to Register (e.g., MOV R0, [FP+8])
	AM_REG_MEM_FP   uint32 = 0x5 // Register to FP+Offset Memory (e.g., MOV [FP+8], R0)
	AM_SINGLE_REG   uint32 = 0x6 // Single Register operand (e.g., PUSH R0, NEG R0)
	AM_IMM_PORT_REG uint32 = 0x7 // Immediate (port number) to Register (e.g., IN R0, #PORT_ID)
	AM_REG_PORT_IMM uint32 = 0x8 // Register to Immediate (port number) (e.g., OUT #PORT_ID, R0)
	AM_ABS_ADDR     uint32 = 0x9 // Absolute Address (for JMP, CALL)
	AM_NO_OPERANDS  uint32 = 0xA // No operands (e.g., HALT, RET)

	AM_MEM_REG    uint32 = 0xB // Memory (absolute address) to Register (MOV rd, [addr])
	AM_SPOFFS_REG uint32 = 0xC // Stack Pointer Offset to Register (MOV rd, [(sp)+offs])
	AM_MEM_MEM    uint32 = 0xD // Memory to Memory (MOV [dest_addr], [source_addr])
	AM_REG_MEM    uint32 = 0xE // Register to Memory (MOV [addr], rs)

	MATH_R_R_R uint32 = 0x1 // Math op: Reg, Reg, Reg (ADD rd, rs1, rs2) - This is a placeholder, needs careful bit packing for 3 regs
	MATH_R_M_R uint32 = 0x2 // Math op: Reg, Mem, Reg (ADD rd, rs1, [addr])
	MATH_M_M_R uint32 = 0x3 // Math op: Mem, Mem, Reg (ADD rd, [addr1], [addr2])
	MATH_R_I_R uint32 = 0x4

	//FIXME: размерность
	AM_JE  uint32 = 0x12 // Jump if Equal
	AM_JNE uint32 = 0x13 // Jump if Not Equal
	AM_JG  uint32 = 0x14 // Jump if Greater (signed)
	AM_JL  uint32 = 0x15 // Jump if Less (signed)
	AM_JGE uint32 = 0x16 // Jump if Greater or Equal (signed)
	AM_JLE uint32 = 0x17 // Jump if Less or Equal (signed)
	// AM_JA  uint32 = 0x18 // Jump if Above (unsigned)
	// AM_JB  uint32 = 0x19 // Jump if Below (unsigned)
	// AM_JAE uint32 = 0x1A // Jump if Above or Equal (unsigned)
	// AM_JBE uint32 = 0x1B // Jump if Below or Equal (unsigned)

	AM_JMP_ABS uint32 = 0x1C // Jump Absolute (JMP addr)
	AM_JMP_REG uint32 = 0x1D // Jump Register (JMP reg)
	AM_JMP_MEM uint32 = 0x1E // Jump Memory (JMP [addr])

	AM_CMP uint32 = 0x1F // Compare (CMP rs1, rs2)

	AM_CALL_ABS uint32 = 0x20 // Call Absolute (CALL addr)
	AM_CALL_REG uint32 = 0x21 // Call Register (CALL reg)
	AM_CALL_MEM uint32 = 0x22 // Call Memory (CALL [addr])

	AM_RET_IMM uint32 = 0x23 // Return with Immediate offset (RET imm)
	AM_IMM_MEM uint32 = 0x24
)
