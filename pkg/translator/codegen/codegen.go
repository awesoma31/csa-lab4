package codegen

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
)

// =============================================================================
// CodeGenerator Structure and Core Methods
// =============================================================================

const WORD_SIZE_BYTES = 4

type SymbolEntry struct {
	Name        string
	Type        ast.Type
	MemoryArea  string // "data", "stack", "code" (for functions)
	AbsAddress  uint32 // Absolute address in data memory or code memory (word-address)
	FPOffset    int    // Offset from FP for stack variables (in bytes, negative)
	SizeInBytes int
	NumberValue int32
	StringValue string

	NumParams     int
	LocalVarCount int // Total size of local variables in bytes

}

// CodeGenerator: Main code generator structure
type CodeGenerator struct {
	// Output segments
	instructionMemory []uint32 // Machine words for instruction memory
	dataMemory        []byte   // Machine words for data memory
	debugAssembly     []string // Assembly mnemonics with addresses

	// State of the code generator
	scopeStack          []Scope
	nextInstructionAddr uint32 // Next free address in instruction memory (word-addresses)
	nextDataAddr        uint32 // Next free address in data memory (byte-addresses)
	errors              []string
	currentFrameOffset  int
}

type Scope struct {
	symbols map[string]SymbolEntry
}

func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		instructionMemory:   make([]uint32, 0),
		dataMemory:          make([]byte, 0),
		debugAssembly:       make([]string, 0),
		scopeStack:          make([]Scope, 0),
		nextInstructionAddr: 0,
		nextDataAddr:        0,
		errors:              make([]string, 0),
	}
	return cg
}

func (cg *CodeGenerator) Generate(program ast.BlockStmt) ([]uint32, []byte, []string, []string) {
	cg.VisitProgram(&program)

	return cg.instructionMemory, cg.dataMemory, cg.debugAssembly, cg.errors
}

func (cg *CodeGenerator) GetMachineCode() []uint32 {
	return cg.instructionMemory
}
func (cg *CodeGenerator) GetDataMemory() []byte {
	return cg.dataMemory
}
func (cg *CodeGenerator) GetDebugAssembly() []string {
	return cg.debugAssembly
}
func (cg *CodeGenerator) GetErrors() []string {
	return cg.errors
}
func (cg *CodeGenerator) addError(msg string) {
	cg.errors = append(cg.errors, msg)
}

func (cg *CodeGenerator) pushScope() {
	cg.scopeStack = append(cg.scopeStack, Scope{symbols: make(map[string]SymbolEntry)})
}

func (cg *CodeGenerator) popScope() {
	if len(cg.scopeStack) > 1 {
		// TODO: Если это выход из функции, здесь ли нужно добавить код для деаллокации стека
		// и восстановления SP/FP.
		cg.scopeStack = cg.scopeStack[:len(cg.scopeStack)-1]
	} else {
		cg.addError("Attempted to pop global scope.")
	}
}

func (cg *CodeGenerator) currentScope() *Scope {
	if len(cg.scopeStack) == 0 {
		return nil //TODO: Или паника, или ошибка
	}
	return &cg.scopeStack[len(cg.scopeStack)-1]
}

func (cg *CodeGenerator) lookupSymbol(name string) (SymbolEntry, bool) {
	for i := len(cg.scopeStack) - 1; i >= 0; i-- {
		scope := cg.scopeStack[i]
		if entry, ok := scope.symbols[name]; ok {
			return entry, true
		}
	}
	return SymbolEntry{}, false
}

func (cg *CodeGenerator) addSymbolToScope(entry SymbolEntry) {
	fmt.Printf("added var-[%s] to scope %d with value=%d\n", entry.Name, len(cg.scopeStack), entry.NumberValue)
	cg.currentScope().symbols[entry.Name] = entry
}

// emitInstruction adds an instruction to the instruction memory.
func (cg *CodeGenerator) emitInstruction(opcode, mode uint32, regD, regS1, regS2 int) {
	instructionWord := encodeInstructionWord(opcode, mode, regD, regS1, regS2)
	cg.instructionMemory = append(cg.instructionMemory, instructionWord)
	cg.debugAssembly = append(
		cg.debugAssembly,
		fmt.Sprintf("[0x%08X] - %08X - (Opc: %02s, Mode: %s, D:%s, S1:%s, S2:%s)",
			cg.nextInstructionAddr,
			instructionWord,
			GetMnemonic(opcode),
			GetAMnemonic(mode),
			GetRegisterMnemonic(regD),
			GetRegisterMnemonic(regS1),
			GetRegisterMnemonic(regS2),
		),
	)
	cg.nextInstructionAddr++
}

// emitImmediate adds an immediate value as an operand to the instruction memory.
func (cg *CodeGenerator) emitImmediate(value uint32) {
	cg.instructionMemory = append(cg.instructionMemory, value)
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("[0x%08X] - %08X - (Imm)", cg.nextInstructionAddr, value))
	cg.nextInstructionAddr++
}

// ReserveWord reserves space for a word in instruction memory and returns its address.
func (cg *CodeGenerator) ReserveWord() uint32 {
	addr := cg.nextInstructionAddr
	cg.emitInstruction(OP_NOP, AM_NO_OPERANDS, -1, -1, -1) // Emit a NOP
	return addr
}

// PatchWord updates a previously emitted word at a given address in instruction memory.
func (cg *CodeGenerator) PatchWord(address, value uint32) {
	if address >= uint32(len(cg.instructionMemory)) {
		cg.addError(fmt.Sprintf("Attempted to patch address %d out of bounds (instruction memory size %d).", address, len(cg.instructionMemory)))
		return
	}
	cg.instructionMemory[address] = value
	// Update debug assembly line for patching
	originalLine := cg.debugAssembly[address]
	cg.debugAssembly[address] = fmt.Sprintf("%s -> PATCHED to %08X", originalLine, value)
}

// addString adds a string literal to data memory.
// It stores: [length_uint32 (4 bytes)][string_bytes...][padding_bytes_to_align_next_word].
// The entire block must start at a word-aligned byte-address.
// TODO: return start of str(len).
func (cg *CodeGenerator) addString(s string) uint32 {
	//Align current data address to the next word boundary (if not already aligned)
	currentByteAddr := cg.nextDataAddr
	alignmentPadding := (WORD_SIZE_BYTES - int(currentByteAddr%WORD_SIZE_BYTES)) % WORD_SIZE_BYTES
	for range alignmentPadding {
		cg.dataMemory = append(cg.dataMemory, 0)
		cg.nextDataAddr++
	}

	// This is the word-aligned byte-address where the string's length word will start
	stringBlockStartAddr := cg.nextDataAddr

	strBytes := []byte(strings.Trim(s, `"`))

	// Store length as a uint32 (4 bytes)
	length := byte(len(strBytes))
	buf := make([]byte, WORD_SIZE_BYTES)
	// binary.LittleEndian.PutUint32(buf, length) // Using LittleEndian for byte order
	buf = append(buf, length)

	cg.dataMemory = append(cg.dataMemory, buf...)
	cg.nextDataAddr += 1

	// The address of the actual string characters starts now (this is what the variable will point to)
	// charStartAddr := cg.nextDataAddr // This might be needed if you want the pointer to point AFTER the length

	// Store string characters (byte by byte)
	cg.dataMemory = append(cg.dataMemory, strBytes...)
	cg.nextDataAddr += uint32(len(strBytes))

	// Add padding after the string characters to align the *next* data item
	// This ensures subsequent data variables also start on a word boundary.
	remainingBytes := int(cg.nextDataAddr % WORD_SIZE_BYTES)
	if remainingBytes != 0 {
		padding := WORD_SIZE_BYTES - remainingBytes
		for range padding {
			cg.dataMemory = append(cg.dataMemory, 0) // Add padding bytes
			cg.nextDataAddr++
		}
	}

	return stringBlockStartAddr
}

// addNumberData adds a uint32 number to data memory.
// This function ensures the number is stored at a word-aligned byte-address.
// Returns the byte-address where the number is stored.
func (cg *CodeGenerator) addNumberData(val uint32) uint32 {
	// Ensure current data address is aligned to the next word boundary
	allignDataMem(cg)

	dataAddr := cg.nextDataAddr // This is now a word-aligned byte-address
	buf := make([]byte, WORD_SIZE_BYTES)
	binary.LittleEndian.PutUint32(buf, val) // Assuming LittleEndian

	cg.dataMemory = append(cg.dataMemory, buf...)
	cg.nextDataAddr += WORD_SIZE_BYTES // Increment by the size of the number in bytes

	return dataAddr
}

func allignDataMem(cg *CodeGenerator) {
	currentByteAddr := cg.nextDataAddr
	alignmentPadding := (WORD_SIZE_BYTES - int(currentByteAddr%WORD_SIZE_BYTES)) % WORD_SIZE_BYTES
	for range alignmentPadding {
		cg.dataMemory = append(cg.dataMemory, 0) // Add padding bytes
		cg.nextDataAddr++
	}
}
