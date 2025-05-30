package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer" // Lexer might not be directly used here if only AST is traversed
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
	// case ast.WhileStmt: // Добавил обработку WhileStmt
	// 	cg.generateWhileStmt(s)
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
	case ast.SymbolExpr: // Renamed from IdentifierExpr
		cg.generateIdentifierExpr(e)
	case ast.BinaryExpr:
		cg.generateBinaryExpr(e)
	case ast.PrefixExpr: // Renamed from UnaryExpr
		cg.generateUnaryExpr(e)
	case ast.CallExpr:
		cg.generateCallExpr(e)
	case ast.AssignmentExpr: // Добавил обработку AssignmentExpr
		cg.generateAssignExpr(e)
	default:
		cg.addError(fmt.Sprintf("Unsupported expression type: %T", e))
	}
}

// VisitProgram generates code for the entire program.
func (cg *CodeGenerator) VisitProgram(p *ast.BlockStmt) {
	// Initialize global scope
	cg.pushScope()

	for _, stmt := range p.Body {
		cg.generateStmt(stmt)
	}

	cg.emitInstruction(OP_HALT, AM_NO_OPERANDS, -1, -1, -1)

	cg.popScope() // Pop global scope
}

func (cg *CodeGenerator) generateVarDeclStmt(s ast.VarDeclarationStmt) {
	if _, found := cg.lookupSymbol(s.Identifier); found {
		cg.addError(fmt.Sprintf("Variable '%s' already declared.", s.Identifier))
		return
	}

	//TODO: complete math ion declaration
	sizeInWords := 1

	if s.AssignedValue != nil {
		// Basic type inference (e.g., if it's a NumberExpr, assume int)
		switch s.AssignedValue.(type) {
		case ast.NumberExpr:
			s.ExplicitType = ast.IntType
		case ast.StringExpr:
			s.ExplicitType = ast.StringType
		default:
			s.ExplicitType = ast.IntType // Default to int if no clear type or assignment
		}
	} else {
		cg.addError(fmt.Sprintf("All variables should be initialised: %s", s.Identifier))
	}

	symbolEntry := SymbolEntry{
		Name:       s.Identifier,
		Address:    cg.nextDataAddr,
		Type:       s.ExplicitType,
		MemoryArea: "data", // Global variables are in data
		Size:       sizeInWords,
	}
	cg.addSymbol(s.Identifier, symbolEntry)
	cg.nextDataAddr += uint32(sizeInWords)

	// Ensure dataMemory has enough capacity or append zeros
	for range sizeInWords {
		cg.dataMemory = append(cg.dataMemory, 0)
	}

	// Create a dummy AssignmentExpr to reuse existing logic
	assignExpr := ast.AssignmentExpr{
		Assigne:       ast.SymbolExpr{Value: s.Identifier},
		AssignedValue: s.AssignedValue,
	}
	cg.generateAssignExpr(assignExpr)
}

// generateAssignExpr generates code for assignment expressions.
func (cg *CodeGenerator) generateAssignExpr(e ast.AssignmentExpr) {
	// Evaluate the right-hand side expression, result in R0
	//TODO: check
	cg.generateExpr(e.AssignedValue)

	switch target := e.Assigne.(type) {
	case ast.SymbolExpr: // Assignment to a simple variable
		symbol, found := cg.lookupSymbol(target.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in assignment.", target.Value))
			return
		}

		// Store the value from R0 into the variable's memory location.
		//TODO: not check for global but check for scope
		if symbol.IsGlobal {
			cg.emitInstruction(OP_MOV, AM_REG_MEM_ABS, -1, R0, -1) // MOV [target_addr], R0
			cg.emitImmediate(symbol.Address)                       // Absolute address operand
		} else {
			// Local variable, store to FP+offset
			// offset := symbol.Address - cg.scopeStack[0][symbol.Name].Address // This logic needs to be revisited for true FP-relative offsets
			cg.emitInstruction(OP_MOV, AM_REG_MEM_FP, -1, R0, -1) // MOV [FP+offset], R0
			cg.emitImmediate(uint32(symbol.Offset))               // Using symbol.Offset directly
		}
	default:
		cg.addError(fmt.Sprintf("Unsupported assignment target type: %T", target))
	}
}

// VisitExpressionStmt handles statements that are just expressions (e.g., function calls).
func (cg *CodeGenerator) VisitExpressionStmt(s ast.ExpressionStmt) { // Changed to value receiver
	cg.generateExpr(s.Expression)
}

// generateBlockStmt handles code blocks (scopes).
func (cg *CodeGenerator) generateBlockStmt(s ast.BlockStmt) { // Changed to value receiver
	cg.pushScope() // Create a new scope for the block
	// TODO: Handle local variable allocation/deallocation on stack here
	for _, stmt := range s.Body {
		cg.generateStmt(stmt)
	}
	cg.popScope() // Exit the scope
}

// generateIfStmt handles if-else statements.
func (cg *CodeGenerator) generateIfStmt(s ast.IfStmt) { // Changed to value receiver
	// Evaluate condition (result in R0, 0 for false, non-zero for true)
	cg.generateExpr(s.Condition)

	// JUMP IF FALSE: JMPF R0, else_label (if R0 is 0, jump to else_label)
	// We'll use a JMP instruction followed by a conditional check.
	// Current instruction set might need a dedicated JMPF or use compare/jump.
	// For simplicity, assuming JMP opcode can handle conditional branching if R0 is 0.
	// This will jump if R0 == 0.
	cg.emitInstruction(OP_JMP, AM_REG_REG, R0, -1, -1) // JMP if R0 is 0 (false)
	jumpToFalseTargetPlaceholder := cg.ReserveWord()   // Reserve space for jump target address

	// Visit 'then' block
	cg.generateStmt(s.Consequent)

	if s.Alternate != nil {
		// After 'then' block, jump over 'else' block
		cg.emitInstruction(OP_JMP, AM_ABS_ADDR, -1, -1, -1)
		jumpToEndIfTargetPlaceholder := cg.ReserveWord() // Reserve space for jump target address after else

		// Patch the jump target for the 'if' condition (if false, jump here)
		cg.PatchWord(jumpToFalseTargetPlaceholder, cg.nextInstructionAddr)

		// Visit 'else' block
		cg.generateStmt(s.Alternate)

		// Patch the jump target for the jump after 'then' block
		cg.PatchWord(jumpToEndIfTargetPlaceholder, cg.nextInstructionAddr)
	} else {
		// Patch the jump target for the 'if' condition (if false, jump here)
		cg.PatchWord(jumpToFalseTargetPlaceholder, cg.nextInstructionAddr)
	}
}

// generateWhileStmt handles while loops.
// func (cg *CodeGenerator) generateWhileStmt(s ast.WhileStmt) { // Changed to value receiver
//
//		loopStartAddr := cg.nextInstructionAddr // Mark the start of the loop
//
//		cg.generateExpr(s.Condition) // Evaluate condition (result in R0)
//
//		// JMPF R0, loop_end_label (jump if R0 is 0)
//		cg.emitInstruction(OP_JMP, AM_REG_REG, R0, -1, -1) // JMP if R0 is 0 (false)
//		jumpToEndLoopTargetPlaceholder := cg.ReserveWord() // Reserve space for jump target address
//
//		cg.generateBlockStmt(ast.BlockStmt{Body: s.Body}) // Visit loop body, treating it as a block
//
//		// Jump back to the start of the loop
//		cg.emitInstruction(OP_JMP, AM_ABS_ADDR, -1, -1, -1)
//		cg.emitImmediate(loopStartAddr)
//
//		// Patch the jump target for the 'while' condition (if false, jump here)
//		cg.PatchWord(jumpToEndLoopTargetPlaceholder, cg.nextInstructionAddr)
//	}
//
// generateNumberExpr generates code for number literals.
func (cg *CodeGenerator) generateNumberExpr(e ast.NumberExpr) { // Changed to value receiver
	// Move the immediate value into R0 (accumulator)
	cg.emitInstruction(OP_MOV, AM_IMM_REG, R0, -1, -1)
	cg.emitImmediate(uint32(e.Value))
}

// generateStringExpr generates code for string literals.
func (cg *CodeGenerator) generateStringExpr(e ast.StringExpr) { // Changed to value receiver
	// Store string in data memory and load its address into R0.
	stringAddr := cg.addString(e.Value) // This function handles writing string to dataMemory

	cg.emitInstruction(OP_MOV, AM_IMM_REG, R0, -1, -1) // MOV R0, #string_address
	cg.emitImmediate(stringAddr)
}

// generateIdentifierExpr generates code for identifier (variable) access.
func (cg *CodeGenerator) generateIdentifierExpr(e ast.SymbolExpr) { // Changed to value receiver
	symbol, found := cg.lookupSymbol(e.Value) // e.Name -> e.Value
	if !found {
		cg.addError(fmt.Sprintf("Undeclared variable '%s' used in expression.", e.Value))
		return
	}

	if symbol.IsGlobal {
		cg.emitInstruction(OP_MOV, AM_MEM_ABS_REG, R0, -1, -1) // MOV R0, [target_addr]
		cg.emitImmediate(symbol.Address)                       // Absolute address operand
	} else {
		// Local variable, load from FP+offset
		// offset := symbol.Address - cg.scopeStack[0][symbol.Name].Address // This logic needs to be revisited for true FP-relative offsets
		cg.emitInstruction(OP_MOV, AM_MEM_FP_REG, R0, -1, -1) // MOV R0, [FP+offset]
		cg.emitImmediate(uint32(symbol.Offset))               // Using symbol.Offset directly
	}
}

// generateBinaryExpr generates code for binary operations.
func (cg *CodeGenerator) generateBinaryExpr(e ast.BinaryExpr) { // Changed to value receiver
	// Evaluate left-hand side, result in R0
	cg.generateExpr(e.Left)
	cg.emitInstruction(OP_PUSH, AM_SINGLE_REG, R0, -1, -1) // Push R0 to stack

	// Evaluate right-hand side, result in R0
	cg.generateExpr(e.Right)
	cg.emitInstruction(OP_MOV, AM_REG_REG, R1, R0, -1) // Move R0 to R1 (operand 2)

	cg.emitInstruction(OP_POP, AM_SINGLE_REG, R0, -1, -1) // Pop stack to R0 (operand 1)

	var opcode uint32
	switch e.Operator.Kind { // Access Kind from lexer.Token
	case lexer.PLUS:
		opcode = OP_ADD
	case lexer.MINUS:
		opcode = OP_SUB
	case lexer.ASTERISK: // For multiplication
		opcode = OP_MUL
	case lexer.SLASH: // For division
		opcode = OP_DIV
	default:
		cg.addError(fmt.Sprintf("Unsupported binary operator: %s", e.Operator.Value)) // Use Operator.Value for string
		return
	}

	// Perform operation: R0 = R0 op R1
	cg.emitInstruction(opcode, AM_REG_REG, R0, R0, R1)
}

// generateUnaryExpr generates code for unary operations.
func (cg *CodeGenerator) generateUnaryExpr(e ast.PrefixExpr) { // Changed to value receiver (PrefixExpr)
	cg.generateExpr(e.Right) // Evaluate operand, result in R0

	switch e.Operator.Kind { // Access Kind from lexer.Token
	case lexer.MINUS: // Unary negation
		cg.emitInstruction(OP_NEG, AM_SINGLE_REG, R0, -1, -1)
	// case lexer.BANG: // Logical NOT
	// 	cg.emitInstruction(OP_NOT, AM_SINGLE_REG, R0, -1, -1)
	default:
		cg.addError(fmt.Sprintf("Unsupported unary operator: %s", e.Operator.Value)) // Use Operator.Value
	}
}

// generateCallExpr generates code for function calls.
func (cg *CodeGenerator) generateCallExpr(e ast.CallExpr) { // Changed to value receiver
	// Example: Handling built-in 'print' and 'input' functions.
	// For actual function calls, you'd push arguments, then CALL instruction,
	// and handle return values.

	// Assuming e.Method is an IdentifierExpr for simplicity
	calleeName := ""
	if symbolExpr, ok := e.Method.(ast.SymbolExpr); ok {
		calleeName = symbolExpr.Value
	} else {
		cg.addError(fmt.Sprintf("Unsupported callee type in function call: %T", e.Method))
		return
	}

	switch calleeName {
	case "print":
		if len(e.Arguments) != 1 {
			cg.addError("Print function expects exactly one argument.")
			return
		}
		cg.generateExpr(e.Arguments[0])                         // Evaluate argument, result in R0
		cg.emitInstruction(OP_OUT, AM_REG_PORT_IMM, -1, R0, -1) // OUT #0, R0 (assuming port 0 for console output)
		cg.emitImmediate(0)                                     // Port address for output
	case "input":
		if len(e.Arguments) != 0 {
			cg.addError("Input function expects no arguments.")
			return
		}
		cg.emitInstruction(OP_IN, AM_IMM_PORT_REG, R0, -1, -1) // IN R0, #0 (assuming port 0 for console input)
		cg.emitImmediate(0)                                    // Port address for input
	default:
		// For user-defined functions, lookup the function symbol and CALL its address.
		cg.addError(fmt.Sprintf("Unsupported or undeclared function call: %s", calleeName))
	}
}

// generateFunctionDeclarationStmt handles function declarations. (Stub)
func (cg *CodeGenerator) generateFunctionDeclarationStmt(s ast.FunctionDeclarationStmt) {
	// TODO: Implement function code generation.
	// This would involve:
	// 1. Storing the current instruction address as the function's entry point.
	// 2. Adding the function to the symbol table.
	// 3. Pushing a new scope for function parameters and local variables.
	// 4. Generating code for the function body.
	// 5. Popping the function's scope.
	// 6. Emitting a RET instruction.
	cg.addError(fmt.Sprintf("Function declaration '%s' code generation not yet implemented.", s.Name))
}

// generateReturnStmt handles return statements. (Stub)
func (cg *CodeGenerator) generateReturnStmt(s ast.ReturnStmt) {
	// TODO: Implement return statement.
	// This would involve:
	// 1. Evaluating the return expression (if any) and placing result in a designated return register (e.g., R0).
	// 2. Deallocating local stack frame.
	// 3. Emitting a RET instruction.
	if s.Expr != nil {
		cg.generateExpr(s.Expr) // Evaluate return value into R0
	}
	cg.emitInstruction(OP_RET, AM_NO_OPERANDS, -1, -1, -1)
	cg.addError("Return statement code generation not fully implemented (stack cleanup).")
}
