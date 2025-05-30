package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

// =============================================================================
// AST Traversal and Code Generation Methods
// =============================================================================

// generateStmt generates code for a given statement.
func (cg *CodeGenerator) generateStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case ast.VarDeclarationStmt:
		cg.generateVarDeclStmt(s)
	case ast.ExpressionStmt:
		cg.generateExpr(s.Expression)
		// For an expression statement, the result in R0 is usually discarded.
		// If it's the result of an assignment, the assignment handled saving.
	case ast.BlockStmt:
		cg.generateBlockStmt(s)
	case ast.FunctionDeclarationStmt:
		cg.generateFunctionDeclarationStmt(s)
	case ast.ReturnStmt:
		cg.generateReturnStmt(s)
	case ast.IfStmt:
		cg.generateIfStmt(s)
	// Add other statement types as needed (Loop, etc.)
	default:
		cg.addError(fmt.Sprintf("Unsupported statement type: %T", s))
	}
}

// generateExpr generates code for a given expression, leaving its result in R0.
func (cg *CodeGenerator) generateExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		cg.generateNumberExpr(e)
	case ast.StringExpr:
		cg.generateStringExpr(e)
	case ast.SymbolExpr:
		cg.generateSymbolExpr(e)
	case ast.BinaryExpr:
		cg.generateBinaryExpr(e)
	case ast.AssignmentExpr:
		cg.generateAssignmentExpr(e)
	case ast.PrefixExpr:
		cg.generatePrefixExpr(e)
	case ast.CallExpr:
		cg.generateCallExpr(e)
	// Add other expression types (ArrayLiteral, NewExpr, etc.)
	default:
		cg.addError(fmt.Sprintf("Unsupported expression type: %T", e))
	}
}

// --- Specific Statement Generators ---

func (cg *CodeGenerator) generateVarDeclStmt(stmt ast.VarDeclarationStmt) {
	// Determine if it's a local or global variable
	isGlobal := len(cg.scopeStack) == 1 // If only global scope is on stack

	var entry SymbolEntry
	var varSize int = 4 // Default to 4 bytes for integer

	// Infer type from assigned value if no explicit type (or use explicit type)
	// For simplicity, assuming integer by default, or string for string literals.
	if stmt.AssignedValue != nil {
		if _, ok := stmt.AssignedValue.(ast.StringExpr); ok {
			entry.Type = "string"
			varSize = 4 // Variable stores a pointer/address to string in data memory (4 bytes)
		} else {
			entry.Type = "integer"
		}
	} else {
		// If no assigned value, assume default type (e.g., integer) or require explicit type
		entry.Type = "integer"
	}

	if isGlobal {
		// Global variables are already pre-scanned and added to symbol table
		// Just need to handle initial assignment.
		sym, ok := cg.findSymbol(stmt.Identifier)
		if !ok {
			cg.addError(fmt.Sprintf("Global variable %s not found in pre-scan.", stmt.Identifier))
			return
		}
		entry = sym // Use the pre-allocated entry

		if stmt.AssignedValue != nil {
			cg.generateExpr(stmt.AssignedValue) // Result in R0 (either direct value or address of string literal)
			cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_MEM_ABS, int(R0), -1, -1), fmt.Sprintf("MOV [0x%X], R0 ; Init global %s", entry.Address, entry.Name))
			cg.emitImmediate(entry.Address)
		}
	} else {
		// Local variable: allocate space on stack and add to current scope's symbol table
		cg.currentStackOffset -= varSize // Stack grows downwards (e.g., -4, -8, etc. from FP)
		entry = SymbolEntry{
			Name:       stmt.Identifier,
			Type:       entry.Type, // Use inferred type
			MemoryArea: "stack",
			Offset:     cg.currentStackOffset,
			Size:       varSize,
			IsGlobal:   false,
		}
		cg.scopeStack[len(cg.scopeStack)-1][stmt.Identifier] = entry // Add to current (top) scope

		// Allocate space on stack for the variable by adjusting SP
		// This is usually done once per function for all locals.
		// For now, we do it per var. A more advanced approach would calculate total size.
		cg.emitInstruction(encodeInstructionWord(OP_SUB, AM_IMM_REG, int(SP_REG), -1, -1), fmt.Sprintf("SUB SP, #%d ; Allocate local var %s", varSize, entry.Name))
		cg.emitImmediate(uint32(varSize))

		if stmt.AssignedValue != nil {
			cg.generateExpr(stmt.AssignedValue) // Result in R0
			// Move value from R0 to stack location (FP + offset)
			cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_MEM_FP, int(R0), -1, -1), fmt.Sprintf("MOV [FP+%d], R0 ; Init local %s", entry.Offset, entry.Name))
			cg.emitImmediate(uint32(entry.Offset)) // Offset itself is the immediate operand
		}
	}
}

func (cg *CodeGenerator) generateBlockStmt(stmt ast.BlockStmt) {
	cg.pushScope()
	cg.emitDebug("; --- Entering Block Scope ---")
	// Store current stack offset to restore it later for this scope
	oldStackOffset := cg.currentStackOffset

	for _, s := range stmt.Body {
		cg.generateStmt(s)
	}

	cg.emitDebug("; --- Exiting Block Scope ---")
	// Deallocate space for local variables declared *within this block*
	// by moving SP back to where it was before this block.
	// The difference in stack offsets gives the size of locals in this block.
	bytesToDeallocate := oldStackOffset - cg.currentStackOffset
	if bytesToDeallocate > 0 {
		cg.emitInstruction(encodeInstructionWord(OP_ADD, AM_IMM_REG, int(SP_REG), -1, -1), fmt.Sprintf("ADD SP, #%d ; Deallocate block locals", bytesToDeallocate))
		cg.emitImmediate(uint32(bytesToDeallocate))
	}
	cg.currentStackOffset = oldStackOffset // Restore stack offset
	cg.popScope()
}

// --- Specific Expression Generators ---

func (cg *CodeGenerator) generateNumberExpr(expr ast.NumberExpr) {
	// For simplicity, treating float64 values as 32-bit integers if within range,
	// or potentially storing them in data segment and loading address.
	// For now, assume it fits directly as an immediate.
	val := uint32(expr.Value) // Cast to uint32
	cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_IMM_REG, int(R0), -1, -1), fmt.Sprintf("MOV R0, #%g", expr.Value))
	cg.emitImmediate(val)
}

func (cg *CodeGenerator) generateStringExpr(expr ast.StringExpr) {
	// Pascal string: [length (1 word)][char1 (1 byte)][char2 (1 byte)]...
	// Allocate space in data memory for the string literal
	strLen := len(expr.Value)
	// Calculate words needed: 1 for length, then ceil(strLen / 4) for chars (assuming 4 bytes per word)
	// If 1 char = 1 byte, 4 chars per word

	//TODO:
	// numCharWords := (strLen + 3) / 4 // Ceiling division for bytes
	// totalWords := 1 + numCharWords   // 1 for length + words for chars

	// Store current data address, this will be the address of the string literal
	stringAddr := cg.nextDataAddr
	cg.emitDebug(fmt.Sprintf("DATA %04X: PSTR_LITERAL \"%s\"", stringAddr, expr.Value))

	// Emit length word
	cg.emitData(uint32(strLen), fmt.Sprintf("; PSTR length for \"%s\"", expr.Value))

	// Emit characters word by word (4 chars per word)
	for i := 0; i < strLen; i += 4 {
		word := uint32(0)
		for j := 0; j < 4; j++ {
			if i+j < strLen {
				word |= uint32(expr.Value[i+j]) << (uint(j) * 8) // Little-endian char packing
			}
		}
		cg.emitData(word, fmt.Sprintf("; Chars \"%s\"", expr.Value[i:min(i+4, strLen)]))
	}

	// Result of StringExpr is the address of the string literal in data memory
	cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_IMM_REG, int(R0), -1, -1), fmt.Sprintf("MOV R0, #0x%X ; Address of string \"%s\"", stringAddr, expr.Value))
	cg.emitImmediate(stringAddr)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (cg *CodeGenerator) generateSymbolExpr(expr ast.SymbolExpr) {
	sym, ok := cg.findSymbol(expr.Value)
	if !ok {
		cg.addError(fmt.Sprintf("Undefined symbol: %s", expr.Value))
		return
	}

	switch sym.MemoryArea {
	case "data": // Global variable
		cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_MEM_ABS_REG, int(R0), -1, -1), fmt.Sprintf("MOV R0, [0x%X] ; Load global var %s", sym.Address, sym.Name))
		cg.emitImmediate(sym.Address)
	case "stack": // Local variable
		cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_MEM_FP_REG, int(R0), -1, -1), fmt.Sprintf("MOV R0, [FP%+d] ; Load local var %s", sym.Offset, sym.Name))
		cg.emitImmediate(uint32(sym.Offset)) // Offset is the immediate operand for FP-relative
	default:
		cg.addError(fmt.Sprintf("Unsupported memory area for symbol %s: %s", sym.Name, sym.MemoryArea))
	}
}

func (cg *CodeGenerator) generateBinaryExpr(expr ast.BinaryExpr) {
	// Generate code for Left operand (result in R0)
	cg.generateExpr(expr.Left)
	cg.emitInstruction(encodeInstructionWord(OP_PUSH, AM_SINGLE_REG, int(R0), -1, -1), "PUSH R0") // Save Left result

	// Generate code for Right operand (result in R0)
	cg.generateExpr(expr.Right)

	cg.emitInstruction(encodeInstructionWord(OP_POP, AM_SINGLE_REG, int(R1), -1, -1), "POP R1") // Restore Left result into R1

	// Perform the operation: R0 = R1 OP R0 (Left OP Right)
	var mnemonic string
	var opcode uint32
	switch expr.Operator.Kind {
	case lexer.PLUS:
		opcode = OP_ADD
		mnemonic = "ADD R0, R1"
	case lexer.DASH:
		opcode = OP_SUB
		mnemonic = "SUB R1, R0 ; R1 (left) - R0 (right)"                                                                      // Subtraction order matters
		cg.emitInstruction(encodeInstructionWord(opcode, AM_REG_REG, int(R1), int(R0), -1), mnemonic)                         // R1 = R1 - R0
		cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_REG, int(R1), int(R0), -1), "MOV R0, R1 ; Move result to R0") // Move R1 to R0
		return                                                                                                                // Handled specially
	case lexer.STAR:
		opcode = OP_MUL
		mnemonic = "MUL R0, R1"
	case lexer.SLASH:
		opcode = OP_DIV
		mnemonic = "DIV R1, R0 ; R1 (left) / R0 (right)" // Division order matters
		// For integer division, some architectures might use specific registers (e.g., EDX:EAX for x86)
		// For simplicity, assume R0 = R1 / R0
		cg.emitInstruction(encodeInstructionWord(opcode, AM_REG_REG, int(R1), int(R0), -1), mnemonic)                         // R1 = R1 / R0
		cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_REG, int(R1), int(R0), -1), "MOV R0, R1 ; Move result to R0") // Move R1 to R0
		return                                                                                                                // Handled specially
	default:
		cg.addError(fmt.Sprintf("Unsupported binary operator: %s", lexer.TokenKindString(expr.Operator.Kind)))
		return
	}
	cg.emitInstruction(encodeInstructionWord(opcode, AM_REG_REG, int(R0), int(R1), -1), mnemonic) // R0 = R0 OP R1 (Right OP Left)
}

func (cg *CodeGenerator) generateAssignmentExpr(expr ast.AssignmentExpr) {
	// Evaluate the right-hand side first (result in R0)
	cg.generateExpr(expr.AssignedValue)

	// The assignee must be a SymbolExpr (variable) or a MemberExpr (array/object member)
	switch assignee := expr.Assigne.(type) {
	case ast.SymbolExpr:
		sym, ok := cg.findSymbol(assignee.Value)
		if !ok {
			cg.addError(fmt.Sprintf("Undefined symbol for assignment: %s", assignee.Value))
			return
		}
		switch sym.MemoryArea {
		case "data": // Global variable
			cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_MEM_ABS, int(R0), -1, -1), fmt.Sprintf("MOV [0x%X], R0 ; Assign to global %s", sym.Address, sym.Name))
			cg.emitImmediate(sym.Address)
		case "stack": // Local variable
			cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_MEM_FP, int(R0), -1, -1), fmt.Sprintf("MOV [FP%+d], R0 ; Assign to local %s", sym.Offset, sym.Name))
			cg.emitImmediate(uint32(sym.Offset))
		default:
			cg.addError(fmt.Sprintf("Unsupported memory area for assignment to symbol %s: %s", sym.Name, sym.MemoryArea))
		}
	// case ast.MemberExpr: // For arrays/objects
	// 	cg.generateMemberAssignment(assignee) // This would be more complex
	default:
		cg.addError(fmt.Sprintf("Unsupported assignee type for assignment: %T", assignee))
	}
}

func (cg *CodeGenerator) generatePrefixExpr(expr ast.PrefixExpr) {
	cg.generateExpr(expr.Right) // Result in R0
	var opcode uint32
	var mnemonic string
	switch expr.Operator.Kind {
	case lexer.DASH: // Unary minus
		opcode = OP_NEG
		mnemonic = "NEG R0"
	case lexer.NOT: // Logical NOT
		opcode = OP_NOT
		mnemonic = "NOT R0" // Assumes R0 contains 0 or 1 for boolean
	default:
		cg.addError(fmt.Sprintf("Unsupported prefix operator: %s", lexer.TokenKindString(expr.Operator.Kind)))
		return
	}
	cg.emitInstruction(encodeInstructionWord(opcode, AM_SINGLE_REG, int(R0), -1, -1), mnemonic)
}

func (cg *CodeGenerator) generateFunctionDeclarationStmt(stmt ast.FunctionDeclarationStmt) {
	// Store current instruction address as the function's entry point
	funcAddr := cg.nextInstructionAddr
	cg.emitDebug(fmt.Sprintf("%s_FUNC_START:", stmt.Name))

	// Add function to global symbol table (functions are global)
	cg.scopeStack[0][stmt.Name] = SymbolEntry{
		Name:       stmt.Name,
		Type:       "function",
		MemoryArea: "code",
		Address:    funcAddr,
		NumParams:  len(stmt.Parameters),
	}

	cg.pushScope() // New scope for function parameters and local variables

	// Prologue: Save old FP, set new FP, allocate space for locals
	// Parameters are usually accessed via positive offsets from FP (FP+4, FP+8...)
	// Local variables via negative offsets from FP (FP-4, FP-8...)
	// cg.emitInstruction(encodeInstructionWord(OP_PUSH, AM_SINGLE_REG, int(FP_REG), -1, -1), "PUSH FP")
	// cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_REG, int(SP_REG), int(FP_REG), -1), "MOV FP, SP")

	// Handle parameters: Add them to the current scope.
	// Parameters are pushed onto the stack *before* the call.
	// They are at positive offsets from FP.
	// Example: FP+4 for first param, FP+8 for second (if 4-byte words/params)
	paramOffset := 4 // First parameter is at FP + 4 (after pushed FP)
	for _, param := range stmt.Parameters {
		cg.scopeStack[len(cg.scopeStack)-1][param.Name] = SymbolEntry{
			Name:       param.Name,
			Type:       "integer", // Or infer from AST type
			MemoryArea: "stack",
			Offset:     paramOffset,
			Size:       4, // Assuming 4 bytes per parameter
			IsGlobal:   false,
		}
		paramOffset += 4
	}

	// Calculate and allocate space for local variables (this needs a two-pass approach or dynamic stack adjustment)
	// For now, let's simplify: `VarDeclarationStmt` in `generateVarDeclStmt` will adjust SP
	// This means `SUB SP, #size` will happen every time a local is declared.
	// A more efficient way: pre-calculate total local size for function body and subtract once.
	cg.currentStackOffset = 0 // Reset stack offset for locals within this function frame

	// Generate code for function body
	for _, s := range stmt.Body {
		cg.generateStmt(s)
	}

	// Epilogue (ensure return is handled, otherwise implicit return/fallthrough)
	// If the last statement isn't `ReturnStmt`, ensure a default return.
	// A common pattern is to just put the epilogue after body generation,
	// and `ReturnStmt` jumps to it or duplicates it.
	cg.emitDebug(fmt.Sprintf("%s_FUNC_END:", stmt.Name))
	// cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_REG, int(FP_REG), int(SP_REG), -1), "MOV SP, FP") // Restore SP
	// cg.emitInstruction(encodeInstructionWord(OP_POP, AM_SINGLE_REG, int(FP_REG), -1, -1), "POP FP")           // Restore old FP
	cg.emitInstruction(encodeInstructionWord(OP_RET, AM_NO_OPERANDS, -1, -1, -1), "RET")

	cg.popScope() // Pop function scope
}

func (cg *CodeGenerator) generateReturnStmt(stmt ast.ReturnStmt) {
	if stmt.Expr != nil {
		cg.generateExpr(stmt.Expr) // Evaluate return expression, result in R0
		// R0 usually holds the return value.
	}
	// Jump to function epilogue or directly perform epilogue actions
	// For simplicity, we just generate RET. Assumes epilogue is right after body.
	// cg.emitInstruction(encodeInstructionWord(OP_MOV, AM_REG_REG, int(FP_REG), int(SP_REG), -1), "MOV SP, FP ; Restore SP for return") // Restore SP
	// cg.emitInstruction(encodeInstructionWord(OP_POP, AM_SINGLE_REG, int(FP_REG), -1, -1), "POP FP ; Restore FP for return")           // Restore old FP
	cg.emitInstruction(encodeInstructionWord(OP_RET, AM_NO_OPERANDS, -1, -1, -1), "RET")
}

func (cg *CodeGenerator) generateCallExpr(expr ast.CallExpr) {
	// 1. Resolve the function being called (e.g., a SymbolExpr for function name)
	funcSym, ok := expr.Method.(ast.SymbolExpr)
	if !ok {
		cg.addError("Call expression method must be a symbol (function name).")
		return
	}
	funcEntry, ok := cg.findSymbol(funcSym.Value)
	if !ok || funcEntry.Type != "function" {
		cg.addError(fmt.Sprintf("Function %s not found or not a function.", funcSym.Value))
		return
	}

	// 2. Push arguments onto the stack in reverse order (for C-like calling convention)
	// This means the first argument is at the lowest address (FP+4) after call
	for i := len(expr.Arguments) - 1; i >= 0; i-- {
		arg := expr.Arguments[i]
		cg.generateExpr(arg) // Result in R0
		cg.emitInstruction(encodeInstructionWord(OP_PUSH, AM_SINGLE_REG, int(R0), -1, -1), fmt.Sprintf("PUSH R0 ; Push arg %d for %s", i, funcSym.Value))
	}

	// 3. Call the function
	cg.emitInstruction(encodeInstructionWord(OP_CALL, AM_ABS_ADDR, -1, -1, -1), fmt.Sprintf("CALL 0x%X ; Call function %s", funcEntry.Address, funcSym.Value))
	cg.emitImmediate(funcEntry.Address)

	// 4. Clean up arguments from stack
	// Add SP, #num_bytes_of_args
	argBytes := uint32(len(expr.Arguments) * 4) // Assuming 4 bytes per argument
	if argBytes > 0 {
		cg.emitInstruction(encodeInstructionWord(OP_ADD, AM_IMM_REG, int(SP_REG), -1, -1), fmt.Sprintf("ADD SP, #%d ; Pop args from stack", argBytes))
		cg.emitImmediate(argBytes)
	}
	// The return value (if any) is expected to be in R0
}

func (cg *CodeGenerator) generateIfStmt(stmt ast.IfStmt) {
	// Evaluate condition (result in R0, typically 0 for false, non-zero for true)
	cg.generateExpr(stmt.Condition)

	// We need labels for jumps
	//TODO:
	// elseLabel := cg.nextInstructionAddr + 2 // After JMP_IF_ZERO and its immediate
	endIfLabel := uint32(0) // Will be filled later

	// Jump to else branch if condition is zero (false)
	// (Or a CMP/JE instruction if your ISA supports conditional jumps based on flags)
	// For simplicity, let's use a "jump if zero" instruction.
	// Assuming R0 holds the boolean result (0 or 1)
	cg.emitInstruction(encodeInstructionWord(OP_JMP, AM_ABS_ADDR, int(R0), -1, -1), "JMP_IF_ZERO R0, #ELSE_LABEL") // This is a conceptual instruction.
	// You need specific conditional jump instructions for your ISA (e.g., JZ, JNE, etc.)
	// For this example, let's assume a conditional jump that uses R0:
	// JZ_REG_ABS_ADDR (Opcode, RegToCheck, AddrMode, Address)
	elseJumpTargetPlaceholder := cg.nextInstructionAddr + 1                                                    // Address of the immediate for the jump
	cg.emitInstruction(encodeInstructionWord(OP_JMP, AM_ABS_ADDR, -1, -1, -1), "JMP_IF_ZERO R0, #ELSE_TARGET") // Opcode for conditional jump to an absolute address
	cg.emitImmediate(0xDEADBEEF)                                                                               // Placeholder for else target address

	cg.emitDebug("; --- If Consequent Block ---")
	// Generate code for consequent block
	cg.generateStmt(stmt.Consequent)

	if stmt.Alternate != nil {
		// If there's an else or else if, jump to end of if after consequent
		//TODO:
		// endIfJumpTargetPlaceholder := cg.nextInstructionAddr + 1 // Address of the immediate for the jump
		cg.emitInstruction(encodeInstructionWord(OP_JMP, AM_ABS_ADDR, -1, -1, -1), "JMP #END_IF_LABEL")
		cg.emitImmediate(0xDEADBEEF)        // Placeholder for end-if target address
		endIfLabel = cg.nextInstructionAddr // Current instruction address is the end of IF

		// Fill in the else jump target
		cg.instructionMemory[elseJumpTargetPlaceholder] = endIfLabel // Update placeholder with actual else target
		cg.debugAssembly[elseJumpTargetPlaceholder] = fmt.Sprintf("CODE %04X: %08X (Immediate -> 0x%X) ; ELSE_TARGET", elseJumpTargetPlaceholder, endIfLabel, endIfLabel)

		cg.emitDebug("; --- If Alternate Block ---")
		// Generate code for alternate (else/else-if)
		cg.generateStmt(stmt.Alternate)

		// After alternate, the endIfLabel is the current address
	} else {
		// No alternate, the end of the if statement is right after the consequent
		endIfLabel = cg.nextInstructionAddr // Current instruction address is the end of IF
		// Fill in the else jump target
		cg.instructionMemory[elseJumpTargetPlaceholder] = endIfLabel // Update placeholder with actual else target
		cg.debugAssembly[elseJumpTargetPlaceholder] = fmt.Sprintf("CODE %04X: %08X (Immediate -> 0x%X) ; END_OF_IF_NO_ELSE", elseJumpTargetPlaceholder, endIfLabel, endIfLabel)
	}

	if stmt.Alternate != nil {
		// Fill in the end-if jump target (if there was an else branch)
		// endIfJumpTargetPlaceholder is valid only if there was an explicit JMP #END_IF_LABEL
		//TODO:
		// if cg.instructionMemory[endIfJumpTargetPlaceholder] == 0xDEADBEEF { // Check if it's still the placeholder
		// 	cg.instructionMemory[endIfJumpTargetPlaceholder] = cg.nextInstructionAddr
		// 	cg.debugAssembly[endIfJumpTargetPlaceholder] = fmt.Sprintf("CODE %04X: %08X (Immediate -> 0x%X) ; END_IF_LABEL", endIfJumpTargetPlaceholder, cg.nextInstructionAddr, cg.nextInstructionAddr)
		// }
	}
}

// =============================================================================
// Helper Functions for Generating Machine Code (low-level encoding)
// These functions will be called by your generateExpr/generateStmt methods.
// =============================================================================

// You need to design the exact bit layout for each instruction based on your ISA.
// The following are simplified examples based on the `encodeInstructionWord` helper.

// HALT instruction (no operands)
func encodeHALT() uint32 {
	return encodeInstructionWord(OP_HALT, AM_NO_OPERANDS, -1, -1, -1)
}

// MOV operations
func encodeMOV_IMM_REG(regD int) uint32 { // Immediate value is next word
	return encodeInstructionWord(OP_MOV, AM_IMM_REG, regD, -1, -1)
}
func encodeMOV_REG_REG(regS, regD int) uint32 {
	return encodeInstructionWord(OP_MOV, AM_REG_REG, regD, regS, -1)
}
func encodeMOV_MEM_ABS_REG(regD int) uint32 { // Absolute address is next word
	return encodeInstructionWord(OP_MOV, AM_MEM_ABS_REG, regD, -1, -1)
}
func encodeMOV_REG_MEM_ABS(regS int) uint32 { // Absolute address is next word
	return encodeInstructionWord(OP_MOV, AM_REG_MEM_ABS, -1, regS, -1)
}

// func encodeMOV_MEM_FP_REG(regD int) uint32 { // FP offset is next word
// 	return encodeInstructionWord(OP_MOV, AM_MEM_FP_REG, regD, -1, -1)
// }
// func encodeMOV_REG_MEM_FP(regS int) uint32 { // FP offset is next word
// 	return encodeInstructionWord(OP_MOV, AM_REG_MEM_FP, -1, regS, -1)
// }

// Arithmetic operations (REG_REG mode)
func encodeADD_REG_REG(regD, regS int) uint32 {
	return encodeInstructionWord(OP_ADD, AM_REG_REG, regD, regS, -1)
}
func encodeSUB_REG_REG(regD, regS int) uint32 {
	return encodeInstructionWord(OP_SUB, AM_REG_REG, regD, regS, -1)
}
func encodeMUL_REG_REG(regD, regS int) uint32 {
	return encodeInstructionWord(OP_MUL, AM_REG_REG, regD, regS, -1)
}
func encodeDIV_REG_REG(regD, regS int) uint32 {
	return encodeInstructionWord(OP_DIV, AM_REG_REG, regD, regS, -1)
}

// Unary operations
func encodeNEG_REG(regD int) uint32 {
	return encodeInstructionWord(OP_NEG, AM_SINGLE_REG, regD, -1, -1)
}
func encodeNOT_REG(regD int) uint32 {
	return encodeInstructionWord(OP_NOT, AM_SINGLE_REG, regD, -1, -1)
}

// Stack operations
func encodePUSH_REG(regS int) uint32 {
	return encodeInstructionWord(OP_PUSH, AM_SINGLE_REG, regS, -1, -1)
}
func encodePOP_REG(regD int) uint32 {
	return encodeInstructionWord(OP_POP, AM_SINGLE_REG, regD, -1, -1)
}

// Control Flow
func encodeJMP_ABS() uint32 { // Absolute address is next word
	return encodeInstructionWord(OP_JMP, AM_ABS_ADDR, -1, -1, -1)
}
func encodeCALL_ABS() uint32 { // Absolute address is next word
	return encodeInstructionWord(OP_CALL, AM_ABS_ADDR, -1, -1, -1)
}
func encodeRET() uint32 {
	return encodeInstructionWord(OP_RET, AM_NO_OPERANDS, -1, -1, -1)
}

// Input/Output (Port-mapped)
func encodeIN_PORT_REG(regD int) uint32 { // Port ID is next word
	return encodeInstructionWord(OP_IN, AM_IMM_PORT_REG, regD, -1, -1)
}
func encodeOUT_REG_PORT(regS int) uint32 { // Port ID is next word
	return encodeInstructionWord(OP_OUT, AM_REG_PORT_IMM, regS, -1, -1)
}
