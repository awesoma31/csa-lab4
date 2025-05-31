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
		cg.generateExpr(s.Expression, R0)
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
	// case ast.WhileStmt: // Added WhileStmt handling
	// 	cg.generateWhileStmt(s)
	// Add other statement types as needed (Loop, etc.)
	default:
		cg.addError(fmt.Sprintf("Unsupported statement type: %T", s))
	}
}

// generateExpr generates code for a given expression, leaving its result in specified register.
func (cg *CodeGenerator) generateExpr(expr ast.Expr, rd int) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		cg.emitInstruction(OP_MOV, AM_IMM_REG, rd, int(e.Value), -1)

	case ast.BinaryExpr:
		// cg.generateExpr(e.Left)
		// // 2. Сохранить результат левого операнда на стек (или в другой регистр, если доступен).
		// // Это необходимо, потому что правый операнд тоже будет использовать R0.
		// cg.emitInstruction(OP_PUSH, AM_REG_STACK_OFF, R0, -1, -1, 0) // PUSH R0
		//
		// // 3. Сгенерировать код для правого операнда (результат в R0).
		// cg.generateExpr(e.Right)
		// // 4. Загрузить результат левого операнда со стека в другой регистр (например, R1).
		// cg.emitInstruction(OP_POP, AM_REG_STACK_OFF, R1, -1, -1, 0) // POP R1
		//
		// // 5. Выполнить операцию между R1 (левый операнд) и R0 (правый операнд), результат в R0.
		// // R0 = R1 op R0
		// switch e.Operator.Kind {
		// case lexer.PLUS:
		// 	cg.emitInstruction(OP_ADD, AM_REG_REG, R0, R1, R0) // ADD R0, R1, R0 (R0 = R1 + R0)
		// case lexer.MINUS:
		// 	cg.emitInstruction(OP_SUB, AM_REG_REG, R0, R1, R0) // SUB R0, R1, R0 (R0 = R1 - R0)
		// case lexer.ASTERISK:
		// 	cg.emitInstruction(OP_MUL, AM_REG_REG, R0, R1, R0) // MUL R0, R1, R0 (R0 = R1 * R0)
		// case lexer.SLASH:
		// 	cg.emitInstruction(OP_DIV, AM_REG_REG, R0, R1, R0) // DIV R0, R1, R0 (R0 = R1 / R0)
		// // ... другие операторы: EQUALS, NOT_EQUALS, LESS, GREATER, AND, OR, и т.д.
		// default:
		// 	cg.addError(fmt.Sprintf("Unsupported binary operator: %s", e.Operator.Value))
		// }

	case ast.SymbolExpr: // Для чтения значения переменной (например, если 'a' используется в выражении 'b = a + 1')
		// symbol, found := cg.lookupSymbol(e.Value)
		// if !found {
		// 	cg.addError(fmt.Sprintf("Undeclared variable: %s", e.Value))
		// 	// Загрузить 0 в R0 для восстановления
		// 	cg.emitInstruction(OP_MOV, AM_IMM_REG, R0, -1, -1, 0)
		// 	return
		// }
		// if symbol.IsGlobal {
		// 	cg.emitInstruction(OP_LDR, AM_MEM_ABS_REG, R0, -1, -1, symbol.Address) // LDR R0, [global_addr]
		// } else {
		// 	cg.emitInstruction(OP_LDR, AM_MEM_FP_OFF_REG, R0, -1, -1, uint32(symbol.Offset)) // LDR R0, [FP + offset]
		// }

	// ... другие типы выражений, например StringExpr, CallExpr, PrefixExpr и т.д.
	// (StringExpr уже была рассмотрена в предыдущем ответе)
	case ast.ArrayLiteral:
		// TODO: реализовать генерацию кода для литералов массивов.
		// Это может включать выделение памяти в dataMemory и заполнение ее элементами.
		// Результат в R0 будет адрес первого элемента или метаданные массива.
		cg.addError("ArrayLiteral code generation not implemented.")
	case ast.NewExpr:
		cg.addError("NewExpr code generation not implemented.")
	case ast.FunctionExpr:
		cg.addError("FunctionExpr code generation not implemented.")
	case ast.ComputedExpr:
		cg.addError("ComputedExpr code generation not implemented.")
	case ast.RangeExpr:
		cg.addError("RangeExpr code generation not implemented.")

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

	// cg.popScope() // Pop global scope
}

// generateVarDeclStmt generates code for a variable declaration.
func (cg *CodeGenerator) generateVarDeclStmt(s ast.VarDeclarationStmt) {
	// Check if the variable is already declared in the current scope
	if _, found := cg.currentScope().symbols[s.Identifier]; found {
		cg.addError(fmt.Sprintf("Variable '%s' already declared in this scope.", s.Identifier))
		return
	}

	symbolEntry := SymbolEntry{
		Name: s.Identifier,
	}

	// Determine the type and size based on the assigned value (simple inference)
	if s.AssignedValue != nil {
		switch asVal := s.AssignedValue.(type) {
		case ast.NumberExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WORD_SIZE_BYTES
			//TODO: big numbers go to 0
			symbolEntry.NumberValue = int32(asVal.Value)

			if len(cg.scopeStack) == 1 { // if global
				symbolEntry.MemoryArea = "data"
				symbolEntry.AbsAddress = cg.addNumberData(uint32(asVal.Value))
			} else { // It's a local variable (on the stack)
				//TODO:
				symbolEntry.MemoryArea = "stack"
				// Ensure the offset is aligned for stack variables
				alignmentPadding := (WORD_SIZE_BYTES - (cg.currentFrameOffset % WORD_SIZE_BYTES)) % WORD_SIZE_BYTES
				cg.currentFrameOffset += alignmentPadding // Add padding to the offset

				symbolEntry.FPOffset = cg.currentFrameOffset     // Assign current aligned offset
				cg.currentFrameOffset += symbolEntry.SizeInBytes // Increment offset for the next local variable
			}

			cg.addSymbolToScope(symbolEntry)
			return

		case ast.StringExpr:
			// in word store ptr, in word stored len and in bytes stored chars
			symbolEntry.Type = ast.IntType // Default type, consider type inference later (could be string, bool etc.)
			symbolEntry.SizeInBytes = WORD_SIZE_BYTES
			symbolEntry.NumberValue = int32(cg.nextDataAddr) + int32(symbolEntry.SizeInBytes)

			//store as pointer
			if len(cg.scopeStack) == 1 { // Global
				symbolEntry.MemoryArea = "data"
				symbolEntry.AbsAddress = cg.addNumberData(uint32(symbolEntry.NumberValue))
			} else { // Local (on stack)
				symbolEntry.MemoryArea = "stack"
				alignmentPadding := (WORD_SIZE_BYTES - (cg.currentFrameOffset % WORD_SIZE_BYTES)) % WORD_SIZE_BYTES
				cg.currentFrameOffset += alignmentPadding
				symbolEntry.FPOffset = cg.currentFrameOffset
				cg.currentFrameOffset += symbolEntry.SizeInBytes
			}
			cg.addSymbolToScope(symbolEntry)

			// Generate code to assign the initial value. This happens after symbol is added.
			assignExpr := ast.AssignmentExpr{
				Assigne:       ast.SymbolExpr{Value: s.Identifier},
				AssignedValue: s.AssignedValue,
			}
			cg.generateAssignExpr(assignExpr)

		case ast.AssignmentExpr:
			//TODO: Handle cases like `let x = y = 5;` or `let x = (y = 5);` properly if needed.
			// Currently, this should be handled by the default case below.
			cg.addError("Nested assignment expressions in declaration are not directly supported yet via specific case.")
			return
		default: // This will now handle StringExpr and other expressions
			cg.addError("unimpl default gen var decl ")
			// symbolEntry.Type = ast.IntType // Default type, consider type inference later (could be string, bool etc.)
			// symbolEntry.SizeInBytes = WORD_SIZE_BYTES
			//
			// if len(cg.scopeStack) == 1 { // Global
			// 	symbolEntry.MemoryArea = "data"
			// 	allignDataMem(cg)
			// 	symbolEntry.AbsAddress = cg.nextDataAddr // Assign before reserving
			// 	for range symbolEntry.SizeInBytes {
			// 		cg.dataMemory = append(cg.dataMemory, 0) // Initialize with zeros (reserving space for the pointer/value)
			// 		cg.nextDataAddr++
			// 	}
			// 	allignDataMem(cg) // Align after reserving
			// } else { // Local (on stack)
			// 	symbolEntry.MemoryArea = "stack"
			// 	alignmentPadding := (WORD_SIZE_BYTES - (cg.currentFrameOffset % WORD_SIZE_BYTES)) % WORD_SIZE_BYTES
			// 	cg.currentFrameOffset += alignmentPadding
			// 	symbolEntry.FPOffset = cg.currentFrameOffset
			// 	cg.currentFrameOffset += symbolEntry.SizeInBytes
			// }
			// cg.addSymbolToScope(symbolEntry)
			//
			// // Generate code to assign the initial value. This happens after symbol is added.
			// assignExpr := ast.AssignmentExpr{
			// 	Assigne:       ast.SymbolExpr{Value: s.Identifier},
			// 	AssignedValue: s.AssignedValue,
			// }
			// cg.generateAssignExpr(assignExpr)
		}
	} else {
		cg.addError(fmt.Sprintf("All variables should be initialized: %s", s.Identifier))
		return
	}

}

// generateAssignExpr generates code for assignment expressions.
func (cg *CodeGenerator) generateAssignExpr(e ast.AssignmentExpr) {
	// Evaluate the right-hand side expression, result is left in R0
	cg.generateExpr(e.AssignedValue)

	switch target := e.Assigne.(type) {
	case ast.SymbolExpr: // Assignment to a simple variable
		symbol, found := cg.lookupSymbol(target.Value) // lookupSymbol searches through the entire scope stack
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in assignment.", target.Value))
			return
		}

		// Store the value from R0 into the variable's memory location based on MemoryArea
		if symbol.MemoryArea == "data" { // Global variable
			// MOV [absolute_byte_address], R0
			cg.emitInstruction(OP_MOV, AM_REG_MEM_ABS, -1, R0, -1)
			cg.emitImmediate(symbol.AbsAddress) // Absolute byte-address operand (must be word-aligned)
		} else if symbol.MemoryArea == "stack" { // Local variable (on the stack)
			// MOV [FP+offset_in_bytes], R0
			cg.emitInstruction(OP_MOV, AM_REG_MEM_FP, -1, R0, -1)
			cg.emitImmediate(uint32(symbol.FPOffset)) // Byte-offset from FP as operand (must be word-aligned)
		} else {
			cg.addError(fmt.Sprintf("Unknown memory area for symbol '%s': %s", symbol.Name, symbol.MemoryArea))
			return
		}
	default:
		cg.addError(fmt.Sprintf("Unsupported assignment target type: %T", target))
	}
}

// VisitExpressionStmt handles statements that are just expressions (e.g., function calls).
func (cg *CodeGenerator) VisitExpressionStmt(s ast.ExpressionStmt) {
	cg.generateExpr(s.Expression)
}

// generateBlockStmt handles code blocks (scopes).
func (cg *CodeGenerator) generateBlockStmt(s ast.BlockStmt) {
	cg.pushScope() // Create a new scope for the block
	// TODO: Handle local variable allocation/deallocation on stack here for block-scoped variables
	for _, stmt := range s.Body {
		cg.generateStmt(stmt)
	}
	cg.popScope() // Exit the scope
}

// generateIfStmt handles if-else statements.
func (cg *CodeGenerator) generateIfStmt(s ast.IfStmt) {
	// Evaluate condition (result in R0, 0 for false, non-zero for true)
	cg.generateExpr(s.Condition)

	// JUMP IF FALSE: JMP R0, else_label (if R0 is 0, jump to else_label)
	// We assume OP_JMP with AM_REG_REG implies a conditional jump if RegD (R0) is 0.
	cg.emitInstruction(OP_JMP, AM_REG_REG, R0, -1, -1) // JMP if R0 is 0 (false)
	jumpToFalseTargetPlaceholder := cg.ReserveWord()   // Reserve space for jump target word-address

	// Visit 'then' block
	cg.generateStmt(s.Consequent)

	if s.Alternate != nil {
		// After 'then' block, jump over 'else' block
		cg.emitInstruction(OP_JMP, AM_ABS_ADDR, -1, -1, -1)
		jumpToEndIfTargetPlaceholder := cg.ReserveWord() // Reserve space for jump target word-address after else

		// Patch the jump target for the 'if' condition (if R0 was false, jump here)
		cg.PatchWord(jumpToFalseTargetPlaceholder, cg.nextInstructionAddr)

		// Visit 'else' block
		cg.generateStmt(s.Alternate)

		// Patch the jump target for the jump after 'then' block
		cg.PatchWord(jumpToEndIfTargetPlaceholder, cg.nextInstructionAddr)
	} else {
		// Patch the jump target for the 'if' condition (if R0 was false, jump here)
		cg.PatchWord(jumpToFalseTargetPlaceholder, cg.nextInstructionAddr)
	}
}

// generateNumberExpr generates code for number literals.
func (cg *CodeGenerator) generateNumberExpr(e ast.NumberExpr) {
	// Move the immediate value (number literal) into R0 (accumulator)
	cg.emitInstruction(OP_MOV, AM_IMM_REG, R0, -1, -1)
	cg.emitImmediate(uint32(e.Value)) // Emit the numeric literal directly as an immediate operand
}

// generateStringExpr generates code for string literals.
func (cg *CodeGenerator) generateStringExpr(e ast.StringExpr) {
	// Store string in data memory and load its byte-address (pointer) into R0.
	stringAddr := cg.addString(e.Value) // addString handles writing string to dataMemory (bytes) and alignment

	cg.emitInstruction(OP_MOV, AM_IMM_REG, R0, -1, -1) // MOV R0, #string_byte_address
	cg.emitImmediate(stringAddr)                       // Emit the byte-address as an immediate word
}

// generateIdentifierExpr generates code for identifier (variable) access.
func (cg *CodeGenerator) generateIdentifierExpr(e ast.SymbolExpr) {
	symbol, found := cg.lookupSymbol(e.Value)
	if !found {
		cg.addError(fmt.Sprintf("Undeclared variable '%s' used in expression.", e.Value))
		return
	}

	// Load the value from memory into R0 based on MemoryArea and its byte-address/offset
	if symbol.MemoryArea == "data" { // Global variable
		// MOV R0, [absolute_byte_address] - CPU will fetch 4 bytes (one word)
		cg.emitInstruction(OP_MOV, AM_MEM_ABS_REG, R0, -1, -1)
		cg.emitImmediate(symbol.AbsAddress) // Absolute byte-address operand (must be word-aligned)
	} else if symbol.MemoryArea == "stack" { // Local variable (on the stack)
		// MOV R0, [FP+offset_in_bytes] - CPU will fetch 4 bytes (one word)
		cg.emitInstruction(OP_MOV, AM_MEM_FP_REG, R0, -1, -1)
		cg.emitImmediate(uint32(symbol.FPOffset)) // Byte-offset from FP as operand (must be word-aligned)
	} else {
		cg.addError(fmt.Sprintf("Unknown memory area for symbol '%s': %s", symbol.Name, symbol.MemoryArea))
		return
	}
}

// generateBinaryExpr generates code for binary operations.
func (cg *CodeGenerator) generateBinaryExpr(e ast.BinaryExpr) {
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
func (cg *CodeGenerator) generateUnaryExpr(e ast.PrefixExpr) {
	cg.generateExpr(e.Right) // Evaluate operand, result in R0

	switch e.Operator.Kind { // Access Kind from lexer.Token
	case lexer.MINUS: // Unary negation (minus sign)
		cg.emitInstruction(OP_NEG, AM_SINGLE_REG, R0, -1, -1)
	case lexer.NOT: // Logical NOT
		cg.emitInstruction(OP_NOT, AM_SINGLE_REG, R0, -1, -1)
	default:
		cg.addError(fmt.Sprintf("Unsupported unary operator: %s", e.Operator.Value)) // Use Operator.Value
	}
}

// generateCallExpr generates code for function calls.
func (cg *CodeGenerator) generateCallExpr(e ast.CallExpr) {
	// Example: Handling built-in 'print' and 'input' functions.
	// For actual function calls, you'd push arguments, then CALL instruction,
	// and handle return values.

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
		cg.generateExpr(e.Arguments[0])                         // Evaluate argument, result in R0 (should contain value or string address)
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
