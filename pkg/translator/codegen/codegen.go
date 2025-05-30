package codegen

import (
	"fmt"
	"github.com/awesoma31/csa-lab4/pkg/translator/ast" // Assume ast package is in this path
)

// SymbolEntry represents an entry in the symbol table.
type SymbolEntry struct {
	Name     string
	Address  uint32 // Memory address or stack offset
	IsConst  bool
	Type     ast.Type
	IsGlobal bool // Indicates if the symbol is global
}

// CodeGenerator is responsible for translating AST into machine code.
type CodeGenerator struct {
	instructionMemory   []uint32                 // Generated machine code instructions
	dataMemory          []uint32                 // Allocated data memory
	debugAssembly       []string                 // Human-readable assembly for debugging
	scopeStack          []map[string]SymbolEntry // Stack of symbol tables for scopes (local and nested)
	nextInstructionAddr uint32
	nextDataAddr        uint32
	errors              []string
}

// NewCodeGenerator creates a new instance of CodeGenerator.
func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		instructionMemory:   make([]uint32, 0),
		dataMemory:          make([]uint32, 0),
		debugAssembly:       make([]string, 0),
		scopeStack:          make([]map[string]SymbolEntry, 0),
		nextInstructionAddr: 0,
		nextDataAddr:        0,
		errors:              make([]string, 0),
	}
	cg.pushScope() // Push the initial global scope
	return cg
}

// GetMachineCode returns the generated machine code.
func (cg *CodeGenerator) GetMachineCode() []uint32 {
	return cg.instructionMemory
}

// GetDataMemory returns the generated data memory.
func (cg *CodeGenerator) GetDataMemory() []uint32 {
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

// pushScope creates a new scope.
func (cg *CodeGenerator) pushScope() {
	cg.scopeStack = append(cg.scopeStack, make(map[string]SymbolEntry))
}

// popScope removes the current scope.
func (cg *CodeGenerator) popScope() {
	if len(cg.scopeStack) <= 1 { // Ensure we don't pop the global scope
		cg.addError("Attempted to pop global scope (or no scope exists).")
		return
	}
	cg.scopeStack = cg.scopeStack[:len(cg.scopeStack)-1]
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
	// For simplicity, we add to the current scope.
	// If you want to differentiate global/local at declaration, you'd check len(cg.scopeStack)
	// and add to scopeStack[0] for global, or scopeStack[len-1] for local.
	// Given your current AST, VarDeclarationStmt is visited globally first.
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
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("%08X - %08X - (Opcode: %02X, Mode: %X, D:%d, S1:%d, S2:%d)", cg.nextInstructionAddr, instructionWord, opcode, mode, regD, regS1, regS2))
	cg.nextInstructionAddr++
}

// emitOperand adds an operand (e.g., immediate value, address) to the instruction memory.
func (cg *CodeGenerator) emitOperand(operand uint32) {
	cg.instructionMemory = append(cg.instructionMemory, operand)
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("%08X - %08X - (Operand)", cg.nextInstructionAddr, operand))
	cg.nextInstructionAddr++
}

// ReserveWord reserves space for a word and returns its address.
func (cg *CodeGenerator) ReserveWord() uint32 {
	addr := cg.nextInstructionAddr
	cg.emitInstruction(0, 0, -1, -1, -1) // Emit a NOP or 0-filled word as a placeholder
	return addr
}

// PatchWord updates a previously emitted word at a given address.
func (cg *CodeGenerator) PatchWord(address, value uint32) {
	if address >= uint32(len(cg.instructionMemory)) {
		cg.addError(fmt.Sprintf("Attempted to patch address %d out of bounds.", address))
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

	// Pascal string format: length (1 word) + characters (N words)
	// For simplicity, assume 1 char per word for now. Adjust as needed.
	length := uint32(len(strBytes)) // Number of characters

	// Emit length as the first word
	cg.dataMemory = append(cg.dataMemory, length)
	cg.nextDataAddr++

	// Emit characters, one per word (adjust if you pack multiple chars into a word)
	for _, char := range strBytes {
		cg.dataMemory = append(cg.dataMemory, uint32(char))
		cg.nextDataAddr++
	}
	return stringAddr
}
