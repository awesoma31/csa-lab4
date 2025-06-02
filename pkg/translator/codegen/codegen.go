package codegen

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
)

// =============================================================================
// CodeGenerator Structure and Core Methods
// =============================================================================

const WordSizeBytes = 4

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

type Scope struct {
	symbols map[string]SymbolEntry
}

func (sc *Scope) Symbols() map[string]SymbolEntry {
	return sc.symbols
}

// CodeGenerator  Main code generator structure
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

func (cg *CodeGenerator) ScopeStack() []Scope {
	return cg.scopeStack
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
	// fmt.Printf("added var-[%s] to scope %d with value=%d\n", entry.Name, len(cg.scopeStack), entry.NumberValue)
	cg.currentScope().symbols[entry.Name] = entry
}

// emitInstruction adds an instruction to the instruction memory.
func (cg *CodeGenerator) emitInstruction(opcode, mode uint32, dest, s1, s2 int) {
	instructionWord := encodeInstructionWord(opcode, mode, dest, s1, s2)
	cg.instructionMemory = append(cg.instructionMemory, instructionWord)
	cg.debugAssembly = append(
		cg.debugAssembly,
		fmt.Sprintf("[0x%04X] - %08X - (Opc: %02s, Mode: %s, D:%s, S1:%s, S2:%s)",
			cg.nextInstructionAddr,
			instructionWord,
			GetMnemonic(opcode),
			GetAMnemonic(mode),
			GetRegisterMnemonic(dest),
			GetRegisterMnemonic(s1),
			GetRegisterMnemonic(s2),
		),
	)
	cg.nextInstructionAddr++
}

func (cg *CodeGenerator) emitMov(mode uint32, dest, s1, s2 int) {
	// var instructionWord uint32
	switch mode {
	case AM_REG_REG: // reg to reg
		cg.emitInstruction(OP_MOV, AM_REG_REG, dest, s1, s2)
	case AM_IMM_REG: // imm to reg; s1=imm
		cg.emitInstruction(OP_MOV, AM_IMM_REG, dest, -1, -1)
		cg.emitImmediate(uint32(s1))
	case AM_MEM_REG: // mem to reg; s1=addr
		cg.emitInstruction(OP_MOV, mode, dest, -1, -1)
		cg.emitImmediate(uint32(s1))
	case AM_SPOFFS_REG: // sp+offs to reg
		cg.emitInstruction(OP_MOV, mode, dest, s1, -1)
	case AM_MEM_MEM: // mem to mem
		cg.emitInstruction(OP_MOV, mode, -1, -1, -1)
		cg.emitImmediate(uint32(s1))
		cg.emitImmediate(uint32(s2))
	case AM_REG_MEM: // reg to mem; dest=addr, s1=reg
		cg.emitInstruction(OP_MOV, mode, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	}
	// cg.nextInstructionAddr++
}

// emitImmediate adds an immediate value as an operand to the instruction memory.
func (cg *CodeGenerator) emitImmediate(value uint32) {
	cg.instructionMemory = append(cg.instructionMemory, value)
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("[0x%04X] - %08X - (Imm)", cg.nextInstructionAddr, value))
	cg.nextInstructionAddr++
}

// ReserveWord reserves space for a word in instruction memory and returns its address.
func (cg *CodeGenerator) ReserveWord() uint32 {
	addr := cg.nextInstructionAddr
	cg.emitInstruction(OP_NOP, AM_NO_OPERANDS, -1, -1, -1) // Emit a NOP
	//   TODO: should it nextInstructionAddr++?
	return addr
}

// PatchWord updates a previously emitted word at a given address in instruction memory.
func (cg *CodeGenerator) PatchWord(address, value uint32) {
	if address >= uint32(len(cg.instructionMemory)) {
		cg.addError(fmt.Sprintf("Attempted to patch address %d out of bounds (instruction memory size %d).", address, len(cg.instructionMemory)))
		return
	}
	// NOTE: displaying in the wrong address bcs  extra lines of debug
	cg.instructionMemory[address] = value
	originalLine := cg.debugAssembly[address]
	// cg.debugAssembly[address] = fmt.Sprintf("%s ->[%08X] PATCHED to: 0x%08X", originalLine, address, value)
	fmt.Printf("%s ->[%08X] PATCHED to: 0x%08X", originalLine, address, value)
}

func (cg *CodeGenerator) PatchDebugAssemblyByAddress(targetAddress uint32, newContent string) {
	addressHex := fmt.Sprintf("0x%08X", targetAddress)
	re, err := regexp.Compile("^" + regexp.QuoteMeta(addressHex) + ":")
	if err != nil {
		cg.addError(fmt.Sprintf("Failed to compile regex for address %s: %v", addressHex, err))
		return
	}

	foundAndPatched := false
	for i, line := range cg.debugAssembly {
		if re.MatchString(line) {
			cg.debugAssembly[i] = fmt.Sprintf("%s -> PATCHED: %s", line, newContent)
			foundAndPatched = true
		}
	}

	if !foundAndPatched {
		cg.addError(fmt.Sprintf("No debug assembly line found for address %s to patch.", addressHex))
	}
}

// addString adds a string literal to data memory.
// It stores: [len(4 bytes)][string_bytes...][padding].
// The entire block must start at a word-aligned byte-address.
// TODO: return start of str(len).
// FIXME: check str len in datamem
func (cg *CodeGenerator) addString(s string) uint32 {
	//Align current data address to the next word boundary (if not already aligned)
	currentByteAddr := cg.nextDataAddr
	alignmentPadding := (WordSizeBytes - int(currentByteAddr%WordSizeBytes)) % WordSizeBytes
	for range alignmentPadding {
		cg.dataMemory = append(cg.dataMemory, 0)
		cg.nextDataAddr++
	}

	// This is the word-aligned byte-address where the string's length word will start
	stringBlockStartAddr := cg.nextDataAddr

	strBytes := []byte(strings.Trim(s, `"`))

	// Store length as an uint32 (4 bytes)
	length := byte(len(strBytes))
	buf := make([]byte, WordSizeBytes)
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
	remainingBytes := int(cg.nextDataAddr % WordSizeBytes)
	if remainingBytes != 0 {
		padding := WordSizeBytes - remainingBytes
		for range padding {
			cg.dataMemory = append(cg.dataMemory, 0) // Add padding bytes
			cg.nextDataAddr++
		}
	}

	return stringBlockStartAddr
}

// addNumberData adds an uint32 number to data memory.
// This function ensures the number is stored at a word-aligned byte-address.
// Returns the byte-address where the number is stored.
func (cg *CodeGenerator) addNumberData(val int32) uint32 {
	// Ensure current data address is aligned to the next word boundary
	allignDataMem(cg)

	dataAddr := cg.nextDataAddr // This is now a word-aligned byte-address
	buf := make([]byte, WordSizeBytes)
	binary.LittleEndian.PutUint32(buf, uint32(val)) // Assuming LittleEndian

	cg.dataMemory = append(cg.dataMemory, buf...)
	cg.nextDataAddr += WordSizeBytes // Increment by the size of the number in bytes

	return dataAddr
}

func allignDataMem(cg *CodeGenerator) {
	currentByteAddr := cg.nextDataAddr
	alignmentPadding := (WordSizeBytes - int(currentByteAddr%WordSizeBytes)) % WordSizeBytes
	for range alignmentPadding {
		cg.dataMemory = append(cg.dataMemory, 0) // Add padding bytes
		cg.nextDataAddr++
	}
}
