package codegen

// =============================================================================
// ISA Definition
// =============================================================================

// Helper function to combine opcode and addressing mode into the first instruction word
// Format: [Opcode (8 bits)][Addressing Mode (4 bits)][RegD (3 bits)][RegS1 (3 bits)][RegS2 (3 bits)][Unused (11 bits)]
// Note: Actual bit packing depends on the mode. Not all fields are always used.
func encodeInstructionWord(opcode, mode uint32, regD, regS1, regS2 int) uint32 {
	word := opcode << 24 // Opcode in bits 31-24
	word |= mode << 20   // Addressing mode in bits 23-20
	if regD != -1 {
		word |= uint32(regD) << 17 // RegD in bits 19-17
	}
	if regS1 != -1 {
		word |= uint32(regS1) << 14 // RegS1 in bits 16-14
	}
	if regS2 != -1 {
		word |= uint32(regS2) << 11 // RegS2 in bits 13-11
	}
	return word
}

const (
	R0 = 0 // General purpose register, often used as accumulator
	R1 = 1 // General purpose register
	R2 = 2
	R3 = 3
	R4 = 4
	R5 = 5
	R6 = 6
	R7 = 7

	// Special purpose registers (conceptual, might not be directly addressable by instructions)
	SP_REG = 8 // Stack Pointer
	FP_REG = 9 // Frame Pointer
	PC_REG     // Program Counter is usually implicit
)

// Opcode constants (1 byte: 0x00 - 0xFF)
const (
	OP_HALT uint32 = 0x00
	OP_MOV  uint32 = 0x01
	OP_ADD  uint32 = 0x02
	OP_SUB  uint32 = 0x03
	OP_MUL  uint32 = 0x04
	OP_DIV  uint32 = 0x05
	OP_NEG  uint32 = 0x06
	OP_NOT  uint32 = 0x07

	OP_PUSH uint32 = 0x10
	OP_POP  uint32 = 0x11

	OP_JMP  uint32 = 0x20
	OP_CALL uint32 = 0x21
	OP_RET  uint32 = 0x22

	OP_IN  uint32 = 0x30
	OP_OUT uint32 = 0x31

	//TODO: For vector operations
	// OP_VADD           uint32 = 0x40 // Vector Add
	// ...
)

var opcodeMnemonics = map[uint32]string{}
var amMnemonics = map[uint32]string{}
var registerMnemonics = map[int]string{}

// init function is called automatically when the package is initialized.
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

	opcodeMnemonics[OP_PUSH] = "PUSH"
	opcodeMnemonics[OP_POP] = "POP"

	opcodeMnemonics[OP_JMP] = "JMP"
	opcodeMnemonics[OP_CALL] = "CALL"
	opcodeMnemonics[OP_RET] = "RET"

	opcodeMnemonics[OP_IN] = "IN"
	opcodeMnemonics[OP_OUT] = "OUT"

	// ADDRESS MODE MNEMONIC
	amMnemonics[AM_REG_REG] = "REG_REG"
	amMnemonics[AM_IMM_REG] = "IMM_REG"
	amMnemonics[AM_MEM_ABS_REG] = "MEM_ABS_REG"
	amMnemonics[AM_REG_MEM_ABS] = "REG_MEM_ABS"
	amMnemonics[AM_MEM_FP_REG] = "MEM_FP_REG"
	amMnemonics[AM_REG_MEM_FP] = "REG_MEM_FP"
	amMnemonics[AM_SINGLE_REG] = "SINGLE_REG"
	amMnemonics[AM_IMM_PORT_REG] = "IMM_PORT_REG"
	amMnemonics[AM_REG_PORT_IMM] = "REG_PORT_IMM"
	amMnemonics[AM_ABS_ADDR] = "ABS_ADDR"
	amMnemonics[AM_NO_OPERANDS] = "NO_OPERANDS"

	// REGISTER MNEMONIC
	registerMnemonics[R0] = "R0"
	registerMnemonics[R1] = "R1"
	registerMnemonics[R2] = "R2"
	registerMnemonics[R3] = "R3"
	registerMnemonics[R4] = "R4"
	registerMnemonics[R5] = "R5"
	registerMnemonics[R6] = "R6"
	registerMnemonics[R7] = "R7"
	registerMnemonics[SP_REG] = "SP_REG"
	registerMnemonics[FP_REG] = "FP_REG"
}

// GetMnemonic returns the string mnemonic for a given opcode.
// If the opcode is not found, it returns "UNKNOWN".
func GetMnemonic(opcode uint32) string {
	if mnemonic, ok := opcodeMnemonics[opcode]; ok {
		return mnemonic
	}
	return "UNKNOWN"
}

// GetAMnemonic returns the string mnemonic for a given addressing mode.
// If the mode is not found, it returns "UNKNOWN_AM".
func GetAMnemonic(mode uint32) string {
	if mnemonic, ok := amMnemonics[mode]; ok {
		return mnemonic
	}
	return "UNKNOWN_AM"
}

func GetRegisterMnemonic(reg int) string {
	if reg == -1 {
		return "" // Represents an unused register field
	}
	if mnemonic, ok := registerMnemonics[reg]; ok {
		return mnemonic
	}
	return "UNKNOWN_REG"
}

// Addressing Mode / Operand Type constants (4 bits: 0x0 - 0xF)
// Эти биты будут идти сразу за опкодом в первом слове инструкции.
const (
	AM_REG_REG      uint32 = 0x0 // Register to Register (e.g., ADD R0, R1)
	AM_IMM_REG      uint32 = 0x1 // Immediate to Register (e.g., MOV R0, #123)
	AM_MEM_ABS_REG  uint32 = 0x2 // Absolute Memory to Register (e.g., MOV R0, [0x1000])
	AM_REG_MEM_ABS  uint32 = 0x3 // Register to Absolute Memory (e.g., MOV [0x1000], R0)
	AM_MEM_FP_REG   uint32 = 0x4 // FP+Offset Memory to Register (e.g., MOV R0, [FP+8])
	AM_REG_MEM_FP   uint32 = 0x5 // Register to FP+Offset Memory (e.g., MOV [FP+8], R0)
	AM_SINGLE_REG   uint32 = 0x6 // Single Register operand (e.g., PUSH R0, NEG R0)
	AM_IMM_PORT_REG uint32 = 0x7 // Immediate (port number) to Register (e.g., IN R0, #PORT_ID)
	AM_REG_PORT_IMM uint32 = 0x8 // Register to Immediate (port number) (e.g., OUT #PORT_ID, R0)
	AM_ABS_ADDR     uint32 = 0x9 // Absolute Address (for JMP, CALL)
	AM_NO_OPERANDS  uint32 = 0xA // No operands (e.g., HALT, RET)

	// Добавьте больше, если нужны другие режимы адресации (например, индексная, косвенная через регистр)
)
