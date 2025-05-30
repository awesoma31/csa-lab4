package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
)

// =============================================================================
// CodeGenerator Structure and Core Methods
// =============================================================================

type SymbolEntry struct {
	Name       string
	Type       ast.Type // Changed to ast.Type interface
	MemoryArea string   // "data", "stack", "code" (for functions)
	Address    uint32   // Absolute address in data memory or code memory (word-address)
	Offset     int      // Offset from FP for stack variables (in bytes, negative)
	Size       int      // Size in bytes (4 for int, length + 4 for string)
	IsGlobal   bool     // Global or local variable
	// For functions:
	NumParams     int
	LocalVarCount int // Total size of local variables in bytes
}

// CodeGenerator: Main code generator structure
type CodeGenerator struct {
	// Output segments
	instructionMemory []uint32 // Machine words for instruction memory
	dataMemory        []byte   // Machine words for data memory
	// For debug output
	debugAssembly []string // Assembly mnemonics with addresses

	// State of the code generator
	scopeStack          []map[string]SymbolEntry // Stack of symbol tables for scopes
	nextInstructionAddr uint32                   // Next free address in instruction memory (word-addresses)
	nextDataAddr        uint32                   // Next free address in data memory (word-addresses)
	errors              []string
}

// NewCodeGenerator creates a new instance of CodeGenerator.
func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		instructionMemory:   make([]uint32, 0),
		dataMemory:          make([]byte, 0),
		debugAssembly:       make([]string, 0),
		scopeStack:          make([]map[string]SymbolEntry, 0),
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

// GetMachineCode returns the generated machine code.
func (cg *CodeGenerator) GetMachineCode() []uint32 {
	return cg.instructionMemory
}

// GetDataMemory returns the generated data memory.
func (cg *CodeGenerator) GetDataMemory() []byte {
	return cg.dataMemory
}

// GetDebugAssembly returns the human-readable assembly for debugging.
func (cg *CodeGenerator) GetDebugAssembly() []string {
	return cg.debugAssembly
}

// GetErrors returns any errors encountered during code generation.
func (cg *CodeGenerator) GetErrors() []string {
	return cg.errors
}

func (cg *CodeGenerator) addError(msg string) {
	cg.errors = append(cg.errors, msg)
}

// pushScope adds a new scope to the scope stack.
func (cg *CodeGenerator) pushScope() {
	cg.scopeStack = append(cg.scopeStack, make(map[string]SymbolEntry))
}

// popScope removes the current scope from the scope stack.
// It also generates code to deallocate local variables from the stack.
func (cg *CodeGenerator) popScope() {
	// TODO:
	// if len(cg.scopeStack) <= 1 {
	// 	cg.addError("Attempted to pop global scope.")
	// 	return
	// }

	// Calculate total size of locals in this scope
	currentScope := cg.scopeStack[len(cg.scopeStack)-1]
	// localBytesDeallocated := 0 // We'll manage currentStackOffset instead
	for _, entry := range currentScope {
		if entry.MemoryArea == "stack" {
			// localBytesDeallocated += entry.Size // Sum up sizes
		}
	}
	// cg.emitInstruction(encodeInstructionWord(OP_ADD, AM_IMM_REG, int(SP_REG), -1, -1), fmt.Sprintf("ADD SP, #%d (pop scope)\n", localBytesDeallocated))\n\t// cg.emitImmediate(uint32(localBytesDeallocated))\n

	cg.scopeStack = cg.scopeStack[:len(cg.scopeStack)-1]
	// When popping a scope, restore the stack offset to what it was before this scope
	// This requires tracking the offset *per scope*. For simplicity
}

// lookupSymbol searches for a symbol in the current scope stack (from innermost to outermost).
func (cg *CodeGenerator) lookupSymbol(name string) (SymbolEntry, bool) {
	for i := len(cg.scopeStack) - 1; i >= 0; i-- {
		if entry, ok := cg.scopeStack[i][name]; ok {
			return entry, true
		}
	}
	return SymbolEntry{}, false
}

// addSymbol adds a symbol to the current (innermost) scope.
func (cg *CodeGenerator) addSymbol(name string, entry SymbolEntry) {
	currentScope := cg.scopeStack[len(cg.scopeStack)-1]
	if _, exists := currentScope[name]; exists {
		cg.addError(fmt.Sprintf("Symbol '%s' already declared in this scope.", name))
		return
	}
	currentScope[name] = entry
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
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("[%08X] - %08X - (Imm)", cg.nextInstructionAddr, value))
	cg.nextInstructionAddr++
}

// ReserveWord reserves space for a word in instruction memory and returns its address.
func (cg *CodeGenerator) ReserveWord() uint32 {
	addr := cg.nextInstructionAddr
	cg.emitInstruction(0, 0, -1, -1, -1) // Emit a NOP (0x00) as a placeholder
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

// addString adds a string literal to data memory (Pascal string format: length + characters).
func (cg *CodeGenerator) addString(s string) uint32 {
	stringAddr := cg.nextDataAddr
	strBytes := []byte(s)

	length := byte(len(strBytes))

	cg.dataMemory = append(cg.dataMemory, length)
	cg.nextDataAddr++

	for _, char := range strBytes {
		cg.dataMemory = append(cg.dataMemory, byte(char))
		cg.nextDataAddr++
	}
	return stringAddr
}
