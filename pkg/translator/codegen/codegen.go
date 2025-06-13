package codegen

import (
	"encoding/binary"
	"fmt"
	"regexp"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
)

const (
	DataMemSize                   = 200
	StackStartAddr         uint32 = DataMemSize
	WordSizeBytes                 = 4
	intVectorTableBaseAddr uint32 = 0
	maxInterrupts          uint32 = 2
)

type SymbolEntry struct {
	Name        string
	Type        ast.Type // currently doesnt mean much but should be set
	MemoryArea  string
	AbsAddress  uint32
	SizeInBytes int
	NumberValue int32 // str addr if points to str
	LongValue   int64
	StringValue string
	IsStr       bool
	IsLong      bool
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
	heapPtrAddr         uint32
	errors              []string
}

func (cg *CodeGenerator) NextInstructionAddres() uint32 {
	return cg.nextInstructionAddr
}

func NewCodeGenerator() *CodeGenerator {
	zeros := make([]uint32, maxInterrupts)
	memI := make([]uint32, 0)
	memI = append(memI, zeros...) //reserve space for 2 addreses of interruptions
	cg := &CodeGenerator{
		instructionMemory:   memI,
		dataMemory:          make([]byte, 0),
		debugAssembly:       make([]string, 0),
		scopeStack:          make([]Scope, 0),
		nextInstructionAddr: intVectorTableBaseAddr + maxInterrupts,
		nextDataAddr:        0,
		errors:              make([]string, 0),
	}
	cg.heapPtrAddr = cg.addNumberData(0)
	return cg
}

func (cg *CodeGenerator) ScopeStack() []Scope {
	return cg.scopeStack
}

func (cg *CodeGenerator) Generate(program ast.BlockStmt) ([]uint32, []byte, []string, []string) {
	cg.VisitProgram(&program)

	heapStart := cg.nextDataAddr
	binary.LittleEndian.PutUint32(
		cg.dataMemory[cg.heapPtrAddr:cg.heapPtrAddr+4],
		uint32(heapStart),
	)
	return cg.instructionMemory, cg.dataMemory, cg.debugAssembly, cg.errors
}

func (cg *CodeGenerator) currentScope() *Scope {
	if len(cg.scopeStack) == 0 {
		panic("scope stack is 0")
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
	cg.currentScope().symbols[entry.Name] = entry
	// fmt.Println("added to scope", entry.Name, cg.currentScope().symbols[entry.Name])
}

// emitInstruction encodes an instruction to the instruction memory and stores dubeg info in debugAssembly.
func (cg *CodeGenerator) emitInstruction(opcode, mode uint32, dest, s1, s2 int) {
	instructionWord := isa.EncodeInstructionWord(opcode, mode, dest, s1, s2)
	cg.instructionMemory = append(cg.instructionMemory, instructionWord)
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

func (cg *CodeGenerator) emitMov(mode uint32, dest, s1, s2 int) {
	// var instructionWord uint32
	switch mode {
	case isa.MvRegReg: // reg to reg
		cg.emitInstruction(isa.OpMov, isa.MvRegReg, dest, s1, s2)
	case isa.MvImmReg: // imm to reg; s1=imm dest = rd
		cg.emitInstruction(isa.OpMov, isa.MvImmReg, dest, -1, -1)
		cg.emitImmediate(uint32(s1))
	case isa.MvMemReg: // mem to reg; s1=addr
		cg.emitInstruction(isa.OpMov, mode, dest, -1, -1)
		cg.emitImmediate(uint32(s1))
	case isa.MvRegMemInd:
		cg.emitInstruction(isa.OpMov, mode, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	case isa.MvLowRegToRegInd: // mem[rd]<-rs1(low)
		cg.emitInstruction(isa.OpMov, isa.MvLowRegToRegInd, dest, s1, -1)
	case isa.MvRegMem: // reg to mem; dest=addr, s1=reg
		cg.emitInstruction(isa.OpMov, mode, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	case isa.MvRegIndToReg:
		cg.emitInstruction(isa.OpMov, isa.MvRegIndToReg, dest, s1, -1)
	case isa.MvRegToRegInd: // mem[dest]<-rs1
		cg.emitInstruction(isa.OpMov, isa.MvRegToRegInd, dest, s1, -1)
	case isa.MvByteRegIndToReg:
		cg.emitInstruction(isa.OpMov, isa.MvByteRegIndToReg, dest, s1, -1)
	case isa.MvRegLowToMem:
		cg.emitInstruction(isa.OpMov, isa.MvRegLowToMem, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	default:
		panic("unknown mov mode")
	}
	// cg.nextInstructionAddr++
}

// emitImmediate adds an immediate value as an operand to the instruction memory.
func (cg *CodeGenerator) emitImmediate(value uint32) {
	cg.instructionMemory = append(cg.instructionMemory, value)
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("[0x%04X] - %08X - Imm", cg.nextInstructionAddr, value))
	cg.nextInstructionAddr++
}

// ReserveWord reserves space for a word in instruction memory and returns its address.
func (cg *CodeGenerator) ReserveWord() uint32 {
	addr := cg.nextInstructionAddr
	cg.emitInstruction(isa.OpNop, isa.NoOperands, -1, -1, -1) // Emit a NOP
	return addr
}

// PatchWord updates a previously emitted word at a given address in instruction memory.
func (cg *CodeGenerator) PatchWord(address, value uint32) {
	if address >= uint32(len(cg.instructionMemory)) {
		cg.addError(fmt.Sprintf("Attempted to patch address %d out of bounds (instruction memory size %d).", address, len(cg.instructionMemory)))
		return
	}
	cg.instructionMemory[address] = value
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
// It stores: [len(1 bytes)][string_bytes...][padding].
// The entire starts at a word-aligned byte-address.
func (cg *CodeGenerator) addString(s string) uint32 {
	var strStartAddr uint32 // points to 1st byte of str(len)
	allignDataMem(cg)

	strStartAddr = cg.nextDataAddr
	if len(s) > 255 {
		cg.addError("String too long for Pascal-style string (max 255 characters)")
		return 0
	}

	cg.dataMemory = append(cg.dataMemory, byte(len(s)))
	cg.nextDataAddr++
	for i := range len(s) {
		cg.dataMemory = append(cg.dataMemory, s[i])
		cg.nextDataAddr++
	}

	allignDataMem(cg)

	return strStartAddr
}

func (cg *CodeGenerator) addLongData(v int64) uint32 {
	lo := int32(v & 0xFFFF_FFFF)         // младшее 32 бита
	hi := int32((v >> 32) & 0xFFFF_FFFF) // старшие 32 бита
	addr := cg.addNumberData(lo)         // [addr+0 … addr+3] ← lo
	cg.addNumberData(hi)                 // [addr+4 … addr+7] ← hi
	return addr                          // адрес всей 8-байтовой константы
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
