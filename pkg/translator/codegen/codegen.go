package codegen

import (
	"encoding/binary"
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
)

// Constants for memory management and interrupt handling.
const (
	WordSizeBytes          = 4
	intVectorTableBaseAddr = 0
	maxInterrupts          = 2
	maxStringLength        = 255 // Max characters for Pascal-style string
)

// --- Symbol Management ---

// SymbolEntry represents an entry in the symbol table.
type SymbolEntry struct {
	Name        string
	Type        ast.Type // Type of the symbol (e.g., int, bool, string)
	MemoryArea  string
	AbsAddress  uint32 // Absolute memory address
	SizeInBytes int    // Size of the data in bytes
	NumberValue int32  // Numeric value for integer constants or string address
	LongValue   int64  // For 64-bit integers
	StringValue string // Raw string value
	IsStr       bool
	IsRead      bool // Indicates if the symbol has been read/used
	IsLong      bool // Indicates if the symbol is a 64-bit integer
}

// Scope manages a collection of symbols within a particular scope.
type Scope struct {
	symbols map[string]SymbolEntry
}

// Symbols returns the map of symbols in the current scope.
func (sc *Scope) Symbols() map[string]SymbolEntry {
	return sc.symbols
}

// --- Code Generator Core ---

// CodeGenerator handles the translation of AST into machine code and data.
type CodeGenerator struct {
	instructionMemory []uint32 // Machine words for instruction memory
	dataMemory        []byte   // Machine bytes for data memory
	debugAssembly     []string // Assembly mnemonics with addresses for debugging

	scopeStack          []Scope  // Stack of scopes for symbol resolution
	nextInstructionAddr uint32   // Next free address in instruction memory (word-addresses)
	nextDataAddr        uint32   // Next free address in data memory (byte-addresses)
	heapPtrAddr         uint32   // Address of the heap pointer in data memory
	errors              []string // List of errors encountered during code generation
}

// NewCodeGenerator creates and initializes a new CodeGenerator.
func NewCodeGenerator() *CodeGenerator {
	// Reserve space for interrupt vector table
	instructionMemory := make([]uint32, maxInterrupts)

	cg := &CodeGenerator{
		instructionMemory:   instructionMemory,
		dataMemory:          make([]byte, 0),
		debugAssembly:       make([]string, 0),
		scopeStack:          make([]Scope, 0),
		nextInstructionAddr: intVectorTableBaseAddr + maxInterrupts, // Start after vector table
		nextDataAddr:        0,
		errors:              make([]string, 0),
	}
	// Initialize heap pointer in data memory
	cg.heapPtrAddr = cg.addNumberData(0) // Allocate space for initial heap pointer (0)
	return cg
}

// NextInstructionAddress returns the next available instruction memory address.
func (cg *CodeGenerator) NextInstructionAddress() uint32 {
	return cg.nextInstructionAddr
}

// ScopeStack returns the current scope stack.
func (cg *CodeGenerator) ScopeStack() []Scope {
	return cg.scopeStack
}

// Generate starts the code generation process from the AST.
// It returns the instruction memory, data memory, debug assembly, and any errors.
func (cg *CodeGenerator) Generate(program ast.BlockStmt) ([]uint32, []byte, []string, []string) {
	cg.VisitProgram(&program) // Assuming VisitProgram is the entry point for AST traversal

	// After code generation, update the heap pointer in data memory
	heapStart := cg.nextDataAddr
	binary.LittleEndian.PutUint32(
		cg.dataMemory[cg.heapPtrAddr:cg.heapPtrAddr+WordSizeBytes],
		heapStart,
	)
	return cg.instructionMemory, cg.dataMemory, cg.debugAssembly, cg.errors
}

// currentScope returns a pointer to the topmost scope on the stack.
// Panics if the scope stack is empty.
func (cg *CodeGenerator) currentScope() *Scope {
	if len(cg.scopeStack) == 0 {
		panic("scope stack is empty: cannot access current scope")
	}
	return &cg.scopeStack[len(cg.scopeStack)-1]
}

// lookupSymbol searches for a symbol, starting from the current scope and moving up.
// Returns the symbol entry and true if found, otherwise an empty entry and false.
func (cg *CodeGenerator) lookupSymbol(name string) (SymbolEntry, bool) {
	for i := len(cg.scopeStack) - 1; i >= 0; i-- {
		scope := cg.scopeStack[i]
		if entry, ok := scope.symbols[name]; ok {
			return entry, true
		}
	}
	return SymbolEntry{}, false
}

// addSymbolToScope adds a symbol entry to the current scope.
func (cg *CodeGenerator) addSymbolToScope(entry SymbolEntry) {
	cg.currentScope().symbols[entry.Name] = entry
}

// --- Instruction Emission ---

// emitInstruction encodes and appends an instruction word to instruction memory.
// It also records debug information for assembly.
func (cg *CodeGenerator) emitInstruction(opcode, mode uint32, dest, s1, s2 isa.Register) {
	instructionWord := isa.EncodeInstructionWord(opcode, mode, dest, s1, s2)
	cg.instructionMemory = append(cg.instructionMemory, instructionWord)

	// Determine mnemonic for destination operand (register or port)
	var rdMnem string
	switch opcode {
	case isa.OpIn, isa.OpOut:
		rdMnem = isa.GetPortMnem(dest)
	default:
		rdMnem = isa.GetRegMnem(dest)
	}

	cg.debugAssembly = append(
		cg.debugAssembly,
		fmt.Sprintf("[0x%04X] - %08X - Opc: %02s, Mode: %s, D:%s, S1:%s, S2:%s",
			cg.nextInstructionAddr,
			instructionWord,
			isa.GetOpMnemonic(opcode),
			isa.GetAMnemonic(mode),
			rdMnem,
			isa.GetRegMnem(s1),
			isa.GetRegMnem(s2),
		),
	)
	cg.nextInstructionAddr++
}

// emitMov emits a MOV instruction based on the specified mode.
func (cg *CodeGenerator) emitMov(mode uint32, dest, s1, s2 isa.Register) {
	switch mode {
	case isa.MvRegReg: // reg to reg: MOV D, S1
		cg.emitInstruction(isa.OpMov, isa.MvRegReg, dest, s1, s2)
	case isa.MvImmReg: // imm to reg: MOV D, #Imm (S1 is immediate value)
		cg.emitInstruction(isa.OpMov, isa.MvImmReg, dest, -1, -1)
		cg.emitImmediate(uint32(s1))
	case isa.MvMemReg: // mem to reg: MOV D, [S1] (S1 is memory address)
		cg.emitInstruction(isa.OpMov, mode, dest, -1, -1)
		cg.emitImmediate(uint32(s1))
	case isa.MvRegMemInd: // reg to mem indexed: MOV [D], S1 (dest is address, S1 is register)
		cg.emitInstruction(isa.OpMov, mode, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	case isa.MvLowRegToRegInd: // mem[D] <- S1 (low part of register)
		cg.emitInstruction(isa.OpMov, isa.MvLowRegToRegInd, dest, s1, -1)
	case isa.MvRegMem: // reg to mem: MOV [D], S1 (dest is address, S1 is register)
		cg.emitInstruction(isa.OpMov, mode, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	case isa.MvRegIndToReg: // reg indexed to reg: MOV D, [S1]
		cg.emitInstruction(isa.OpMov, isa.MvRegIndToReg, dest, s1, -1)
	case isa.MvRegToRegInd: // reg to reg indexed: MOV [D], S1 (dest is register, S1 is register)
		cg.emitInstruction(isa.OpMov, isa.MvRegToRegInd, dest, s1, -1)
	case isa.MvByteRegIndToReg: // byte from reg indexed to reg: MOV.B D, [S1]
		cg.emitInstruction(isa.OpMov, isa.MvByteRegIndToReg, dest, s1, -1)
	case isa.MvRegLowToMem: // low part of reg to mem: MOV.L [D], S1
		cg.emitInstruction(isa.OpMov, isa.MvRegLowToMem, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	default:
		cg.addError(fmt.Sprintf("unknown MOV mode encountered: %d", mode))
	}
}

// emitImmediate adds an immediate value as an operand to the instruction memory.
func (cg *CodeGenerator) emitImmediate(value uint32) {
	cg.instructionMemory = append(cg.instructionMemory, value)
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("[0x%04X] - %08X - Imm", cg.nextInstructionAddr, value))
	cg.nextInstructionAddr++
}

// ReserveWord reserves space for a word in instruction memory and returns its address.
// It emits a NOP instruction as a placeholder.
func (cg *CodeGenerator) ReserveWord() uint32 {
	addr := cg.nextInstructionAddr
	cg.emitInstruction(isa.OpNop, isa.NoOperands, -1, -1, -1) // Emit a NOP placeholder
	return addr
}

// PatchWord updates a previously emitted word at a given address in instruction memory.
func (cg *CodeGenerator) PatchWord(address, value uint32) {
	if address >= uint32(len(cg.instructionMemory)) {
		cg.addError(fmt.Sprintf("attempted to patch address 0x%X out of bounds (instruction memory size 0x%X).", address, len(cg.instructionMemory)))
		return
	}
	cg.instructionMemory[address] = value
	// Update debug assembly for the patched instruction, if desired.
	// For simplicity, this refactor doesn't re-generate the debug string,
	// but in a real scenario, you might want to update it to reflect the patched value.
}

// --- Data Memory Management ---

// addString adds a Pascal-style string literal to data memory.
// Format: [len(1 byte)][string_bytes...][padding to word-align].
// Returns the byte address where the string (including length byte) starts.
func (cg *CodeGenerator) addString(s string) uint32 {
	cg.alignDataMemory() // Ensure word alignment before adding string

	strStartAddr := cg.nextDataAddr
	if len(s) > maxStringLength {
		cg.addError(fmt.Sprintf("string too long (%d characters) for Pascal-style string (max %d characters)", len(s), maxStringLength))
		return 0 // Return an invalid address
	}

	cg.dataMemory = append(cg.dataMemory, byte(len(s)))
	cg.nextDataAddr++

	for _, charByte := range []byte(s) {
		cg.dataMemory = append(cg.dataMemory, charByte)
		cg.nextDataAddr++
	}

	cg.alignDataMemory() // Ensure word alignment after adding string

	return strStartAddr
}

// addLongData adds a 64-bit integer to data memory.
// It stores two 32-bit words (low part then high part).
// Returns the address of the low part.
func (cg *CodeGenerator) addLongData(v int64) uint32 {
	lo := int32(v & 0xFFFF_FFFF)         // Lower 32 bits
	hi := int32((v >> 32) & 0xFFFF_FFFF) // Upper 32 bits

	addr := cg.addNumberData(lo) // Store low part
	cg.addNumberData(hi)         // Store high part immediately after
	return addr
}

// addNumberData adds a 32-bit integer to data memory.
// This function ensures the number is stored at a word-aligned byte-address.
// Returns the address where the number is stored.
func (cg *CodeGenerator) addNumberData(val int32) uint32 {
	cg.alignDataMemory() // Ensure word alignment

	dataAddr := cg.nextDataAddr
	buf := make([]byte, WordSizeBytes)
	binary.LittleEndian.PutUint32(buf, uint32(val)) // Convert int32 to uint32 for binary encoding

	cg.dataMemory = append(cg.dataMemory, buf...)
	cg.nextDataAddr += WordSizeBytes

	return dataAddr
}

// alignDataMemory adds zero-padding to data memory until the next address is word-aligned.
func (cg *CodeGenerator) alignDataMemory() {
	currentByteAddr := cg.nextDataAddr
	alignmentPadding := (WordSizeBytes - int(currentByteAddr%WordSizeBytes)) % WordSizeBytes
	for i := 0; i < alignmentPadding; i++ {
		cg.dataMemory = append(cg.dataMemory, 0)
		cg.nextDataAddr++
	}
}

// --- Error Handling ---

// addError appends an error message to the internal error list.
func (cg *CodeGenerator) addError(msg string) {
	cg.errors = append(cg.errors, msg)
}

// --- Scope Stack Management ---

// pushScope adds a new empty scope to the scope stack.
func (cg *CodeGenerator) pushScope() {
	cg.scopeStack = append(cg.scopeStack, Scope{symbols: make(map[string]SymbolEntry)})
}

// popScope removes the top scope from the scope stack.
// Panics if the scope stack is empty.
// func (cg *CodeGenerator) popScope() {
// 	if len(cg.scopeStack) == 0 {
// 		panic("scope stack is empty: cannot pop scope")
// 	}
// 	cg.scopeStack = cg.scopeStack[:len(cg.scopeStack)-1]
// }

// --- Getters for Generated Output ---

// GetMachineCode returns the generated instruction memory.
func (cg *CodeGenerator) GetMachineCode() []uint32 {
	return cg.instructionMemory
}

// GetDataMemory returns the generated data memory.
func (cg *CodeGenerator) GetDataMemory() []byte {
	return cg.dataMemory
}

// GetDebugAssembly returns the generated debug assembly mnemonics.
func (cg *CodeGenerator) GetDebugAssembly() []string {
	return cg.debugAssembly
}

// GetErrors returns any errors encountered during code generation.
func (cg *CodeGenerator) GetErrors() []string {
	return cg.errors
}
