package codegen

import (
	"encoding/binary"
	"fmt"
	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
	"regexp"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
)

const WordSizeBytes = 4

type SymbolEntry struct {
	Name        string
	Type        ast.Type
	MemoryArea  string // "data", "stack", "code" (for functions)
	AbsAddress  uint32 // addr of it in dataMemory
	FPOffset    int
	SizeInBytes int
	NumberValue int32
	StringValue string
	IsStr       bool

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

func (cg *CodeGenerator) NextInstructionAddres() uint32 {
	return cg.nextInstructionAddr
}

const (
	InstrMemSize          = 100
	StackStartAddr uint32 = DataMemSize
	DataMemSize           = 200
)

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
	//TODO: jmp addr mode
	instructionWord := isa.EncodeInstructionWord(opcode, mode, dest, s1, s2)
	cg.instructionMemory = append(cg.instructionMemory, instructionWord)
	cg.debugAssembly = append(
		cg.debugAssembly,
		fmt.Sprintf("[0x%04X] - %08X - Opc: %02s, Mode: %s, D:%s, S1:%s, S2:%s",
			cg.nextInstructionAddr,
			instructionWord,
			isa.GetOpMnemonic(opcode),
			isa.GetAMnemonic(mode),
			isa.GetRegMnem(dest),
			isa.GetRegMnem(s1),
			isa.GetRegMnem(s2),
		),
	)
	if opcode == isa.OpPush {
		fmt.Printf("push word 0x%08X\n", instructionWord)
	}
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
	case isa.MvSpOffsToReg: // sp+offs to reg
		cg.emitInstruction(isa.OpMov, mode, dest, s1, -1)
	case isa.MvMemMem: // mem to mem
		cg.emitInstruction(isa.OpMov, mode, -1, -1, -1)
		cg.emitImmediate(uint32(s1))
		cg.emitImmediate(uint32(s2))
	case isa.MvRegMem: // reg to mem; dest=addr, s1=reg
		cg.emitInstruction(isa.OpMov, mode, -1, s1, -1)
		cg.emitImmediate(uint32(dest))
	case isa.MvRegIndReg:
		cg.emitInstruction(isa.OpMov, isa.MvRegIndReg, dest, s1, -1)
	case isa.MvLowRegIndReg:
		cg.emitInstruction(isa.OpMov, isa.MvLowRegIndReg, dest, s1, -1)
	default:
		panic("unknown mov mode")
	}
	// cg.nextInstructionAddr++
}

func fits17(v uint32) bool { return v>>17 == 0 }

// port ∈ [0..3] → 2 бита. 19 младших бит — “что угодно” (имм-данные для OUT IMM).
const IoNoImmVal = 0

func encodeIOWord(opcode, mode uint32, port uint8, imm int32) uint32 {
	if port > 3 {
		panic("port number must be 0-3")
	}
	word := uint32(opcode)<<26 | mode<<21 | uint32(port)<<19
	if mode == isa.ImmReg {
		word |= uint32(imm) & 0x7FFFF // 19 бит
	}
	return word
}

func (cg *CodeGenerator) emitOut(mode uint32, port uint8, value int32) {
	switch mode {
	case isa.IoMemReg:
		cg.emitOutMemReg(port)
	case isa.ImmReg:
		cg.emitOutImm(port, value)
	}

}

// ───────────────────────── OUT ──────────────────────────
func (cg *CodeGenerator) emitOutMemReg(port uint8) {
	word := encodeIOWord(isa.OpOut, isa.IoMemReg, port, IoNoImmVal)
	cg.instructionMemory = append(cg.instructionMemory, word)
	cg.debugAssembly = append(cg.debugAssembly,
		fmt.Sprintf("[0x%04X] - %08X - Opc: OUT mem[R_OUT_ADDR]->(R_OUT_DATA)->port[%d]", cg.nextInstructionAddr, word, port))
	cg.nextInstructionAddr++
}

func (cg *CodeGenerator) emitOutImm(port uint8, value int32) {
	word := encodeIOWord(isa.OpOut, isa.ImmReg, port, value)
	cg.instructionMemory = append(cg.instructionMemory, word)
	cg.debugAssembly = append(cg.debugAssembly,
		fmt.Sprintf("[0x%04X] OUT #%d → port[%d]  0x%08X", cg.nextDataAddr, value, port, word))
	cg.nextInstructionAddr++
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
