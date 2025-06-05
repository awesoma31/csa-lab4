package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

// generateStmt generates code for a given statement.
func (cg *CodeGenerator) generateStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case ast.VarDeclarationStmt:
		cg.generateVarDeclStmt(s)
	case ast.ExpressionStmt:
		cg.generateExpr(s.Expression, RA)
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
	case ast.PrintStmt: // Обработка оператора Print
		cg.generatePrintStmt(s)
	// case ast.WhileStmt: // Added WhileStmt handling
	// 	cg.generateWhileStmt(s)
	// Add other statement types as needed (Loop, etc.)
	default:
		cg.addError(fmt.Sprintf("Unsupported statement type: %T", s))
	}
}

// generatePrintStmt handles PrintStmt (if `print` can be a statement like `print("hello world");`).
func (cg *CodeGenerator) generatePrintStmt(s ast.PrintStmt) {
	cg.generateExpr(s.Argument, RPRINT) // str addr is in RPRINT [len]byte
	switch arg := s.Argument.(type) {
	case ast.StringExpr:
		strLen := len(arg.Value)
		cg.emitInstruction(OP_ADD, MATH_R_I_R, RPRINT, RPRINT, -1) // RPRINT is ptr to the 1 char
		cg.emitImmediate(1)
		for range strLen {
			cg.emitInstruction(OP_OUT, AM_SINGLE_REG, RPRINT, -1, -1)  // char mem[RPRINT] to out
			cg.emitInstruction(OP_ADD, MATH_R_I_R, RPRINT, RPRINT, -1) // RPRINT++
			cg.emitImmediate(1)
		}
	case ast.SymbolExpr:
		panic("print var not implemented")
		// val, found := cg.currentScope().Symbols()[arg.Value]
		// if !found {
		// 	cg.addError(fmt.Sprintf("could not find variable %s in scope", arg.Value))
		// }
		//TODO:

	}

}

// generateExpr generates code for a given expression, leaving its result in specified register.
func (cg *CodeGenerator) generateExpr(expr ast.Expr, rd int) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		cg.emitMov(AM_IMM_REG, rd, int(e.Value), -1)

	case ast.BinaryExpr:
		//TODO: check push
		cg.generateExpr(e.Left, RM1)
		cg.emitInstruction(OP_PUSH, AM_SINGLE_REG, RM1, -1, -1)

		cg.generateExpr(e.Right, RM2)

		cg.emitInstruction(OP_POP, AM_SINGLE_REG, RM1, -1, -1)

		var opcode uint32
		switch e.Operator.Kind {
		case lexer.PLUS:
			opcode = OP_ADD
		case lexer.MINUS:
			opcode = OP_SUB
		case lexer.STAR:
			opcode = OP_MUL
		case lexer.SLASH:
			opcode = OP_DIV
		case lexer.EQUALS, lexer.NOT_EQUALS, lexer.GREATER, lexer.GREATER_EQUALS, lexer.LESS, lexer.LESS_EQUALS:
			opcode = OP_CMP
			rd = -1
		default:
			cg.addError(fmt.Sprintf("Unsupported binary operator: %s", e.Operator.Value))
			return
		}
		// TODO: check addres mode
		cg.emitInstruction(opcode, AM_REG_REG, rd, RM1, RM2) // R0 = R1 op R0

	case ast.SymbolExpr: // Для чтения значения переменной
		symbol, found := cg.lookupSymbol(e.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable in assign expr: %s", e.Value))
			if rd != -1 {
				cg.emitInstruction(OP_MOV, AM_IMM_REG, rd, -1, -1)
				cg.emitImmediate(0) // Загрузить 0 в rd для восстановления
			}
			return
		}
		if symbol.MemoryArea == "data" {
			cg.emitInstruction(OP_MOV, AM_MEM_REG, rd, -1, -1) // MOV rd, [symbol_addr] // rd<-var
			cg.emitImmediate(symbol.AbsAddress)
		} else if symbol.MemoryArea == "stack" {
			//TODO:
			// cg.emitInstruction(OP_MOV, AM_MEM_FP_REG, rd, -1, -1) // LDR rd, [FP + offset]
			// cg.emitImmediate(uint32(symbol.FPOffset))
		} else {
			cg.addError(fmt.Sprintf("Unknown memory area for symbol '%s': %s", symbol.Name, symbol.MemoryArea))
		}
	case ast.FunctionExpr:
		cg.addError("FunctionExpr code generation not implemented.")
	case ast.StringExpr:
		// panic("impl me StringExpr")
		cg.generateStringExpr(e, rd) // Обновление: передаем rd
	case ast.PrefixExpr: // Для унарных операторов
		newExpr := ast.NumberExpr{}
		switch a := e.Right.(type) {
		case ast.NumberExpr:
			switch e.Operator.Kind {
			case lexer.MINUS:
				newExpr.Value = -a.Value
			case lexer.PLUS:
				newExpr.Value = a.Value
			default:
				cg.addError(fmt.Sprintf("unknown unary prefix %v", e.Operator.Kind))
				return
			}
			cg.generateExpr(newExpr, rd)

		default:
			panic("unimpl prefix functionality, only unary with numbers work for now")
		}
		// cg.generateUnaryExpr(e, rd) // Обновление: передаем rd
	case ast.CallExpr: // Для вызовов функций
		cg.generateCallExpr(e) // Обновление: передаем rd
	case ast.AssignmentExpr: // Для выражений присваивания
		cg.generateAssignExpr(e, rd) // Обновление: передаем rd
	case ast.PrintExpr: // Обработка выражения Print
		cg.generatePrintExpr(e) // Вызываем тот же генератор, что и для PrintStmt
	case ast.ReadExpr: // <-- НОВОЕ: Обработка выражения ReadExpr
		//TODO: var then should store ptr to string
		cg.generateReadExpr()
	default:
		cg.addError(fmt.Sprintf("Unsupported expression type: %T", e))
	}
}

func (cg *CodeGenerator) generateReadExpr() {
	panic("impl me assign o read()")
	cg.emitInstruction(OP_IN, AM_SINGLE_REG, RREAD, -1, -1)
}

func (cg *CodeGenerator) generatePrintExpr(e ast.PrintExpr) {
	panic("unimplemented")
}

// VisitProgram generates code for the entire program.
func (cg *CodeGenerator) VisitProgram(p *ast.BlockStmt) {
	// Initialize global scope
	cg.pushScope()

	for _, stmt := range p.Body {
		cg.generateStmt(stmt)
	}

	cg.emitInstruction(OP_HALT, 0, -1, -1, -1)

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
		switch assignedVal := s.AssignedValue.(type) {
		case ast.NumberExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			//TODO: big numbers go to 0
			symbolEntry.NumberValue = assignedVal.Value

			if len(cg.scopeStack) == 1 { // if global
				symbolEntry.MemoryArea = "data"
				symbolEntry.AbsAddress = cg.addNumberData(assignedVal.Value)
			} else { // It's a local variable (on the stack)
				//TODO:
				symbolEntry.MemoryArea = "stack"
				// Ensure the offset is aligned for stack variables
				alignmentPadding := (WordSizeBytes - (cg.currentFrameOffset % WordSizeBytes)) % WordSizeBytes
				cg.currentFrameOffset += alignmentPadding // Add padding to the offset

				symbolEntry.FPOffset = cg.currentFrameOffset     // Assign current aligned offset
				cg.currentFrameOffset += symbolEntry.SizeInBytes // Increment offset for the next local variable
			}

			cg.addSymbolToScope(symbolEntry)
			return

		case ast.StringExpr:
			strAddr := cg.addString(assignedVal.Value)
			ptrAddr := cg.addNumberData(int32(strAddr))
			// in word store ptr, in word stored len and in bytes stored chars
			symbolEntry.Type = ast.IntType // Default type, consider type inference later (could be string, bool etc.)
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = int32(strAddr)
			symbolEntry.AbsAddress = ptrAddr

			//store as pointer
			if len(cg.scopeStack) == 1 {
				symbolEntry.MemoryArea = "data"
				// symbolEntry.AbsAddress = cg.addNumberData(symbolEntry.NumberValue)
			} else {
				//TODO:
			}
			cg.addSymbolToScope(symbolEntry)
		case ast.ReadExpr:
			strAddr := cg.addString("") // reserve space for reading 1 char
			ptrAddr := cg.addNumberData(int32(strAddr))

			symbolEntry.Type = ast.IntType // PTR to str
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = int32(strAddr)
			//TODO: check when local, stack
			symbolEntry.AbsAddress = ptrAddr

			if len(cg.scopeStack) == 1 { // if global
				symbolEntry.MemoryArea = "data"
			} else { // It's a local variable (on the stack)
				//TODO:
			}

			cg.addSymbolToScope(symbolEntry)
			cg.generateReadExpr()

			cg.emitMov(AM_IMM_REG, RA, 1, -1)            // imm to reg; s1=imm dest = rd
			cg.emitMov(AM_REG_MEM, int(strAddr), RA, -1) // reg to mem; dest=addr, s1=reg
			cg.emitMov(AM_REG_MEM, int(strAddr+1), RREAD, -1)
			return

		case ast.BinaryExpr:
			//TODO: string check for operations

			// cg.generateExpr(s.AssignedValue, R0)
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes

			if len(cg.scopeStack) == 1 { // global
				symbolEntry.MemoryArea = "data"
				symbolEntry.AbsAddress = cg.addNumberData(0) // placeholder = 0
			} else { // local
				// TODO:stack
				/* аналогичный код для stack-переменных */
			}

			cg.addSymbolToScope(symbolEntry)

			assign := ast.AssignmentExpr{
				Assigne:       ast.SymbolExpr{Value: s.Identifier},
				AssignedValue: s.AssignedValue,
			}
			cg.generateAssignExpr(assign, RA)
			return
		case ast.PrefixExpr:
			//TODO: for now movs expr to placeholder
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes

			if len(cg.scopeStack) == 1 { // global
				symbolEntry.MemoryArea = "data"
				symbolEntry.AbsAddress = cg.addNumberData(0) // placeholder = 0
			} else { // local
				// TODO:stack
				/* аналогичный код для stack-переменных */
			}

			cg.addSymbolToScope(symbolEntry)

			assign := ast.AssignmentExpr{
				Assigne:       ast.SymbolExpr{Value: s.Identifier},
				AssignedValue: s.AssignedValue,
			}
			cg.generateAssignExpr(assign, RA)
			return

			// cg.generateExpr(assignedVal, RA)
		default:
			cg.addError(fmt.Sprintf("unknown case of generating var declaration - %T", assignedVal))
		}
	} else {
		cg.addError(fmt.Sprintf("All variables should be initialized: %s - is undefined", s.Identifier))
		return
	}

}

// generateAssignExpr generates code for assignment expressions.
func (cg *CodeGenerator) generateAssignExpr(e ast.AssignmentExpr, rd int) {
	// Evaluate the right-hand side expression, result is left in rd
	cg.generateExpr(e.AssignedValue, rd)

	switch target := e.Assigne.(type) {
	case ast.SymbolExpr: // Assignment to a simple variable
		symbol, found := cg.lookupSymbol(target.Value) // lookupSymbol searches through the entire scope stack
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in assignment.", target.Value))
			return
		}

		// Store the value from rd into the variable's memory location based on MemoryArea
		if symbol.MemoryArea == "data" { // Global variable
			// MOV [absolute_byte_address], rd
			cg.emitInstruction(OP_MOV, AM_REG_MEM_ABS, -1, rd, -1)
			cg.emitImmediate(symbol.AbsAddress) // Absolute byte-address operand (must be word-aligned)
		} else if symbol.MemoryArea == "stack" { // Local variable (on the stack)
			// MOV [FP+offset_in_bytes], rd
			cg.emitInstruction(OP_MOV, AM_REG_MEM_FP, -1, rd, -1)
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
func (cg *CodeGenerator) VisitExpressionStmt(s ast.ExpressionStmt, rd int) {
	cg.generateExpr(s.Expression, rd)
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

func (cg *CodeGenerator) generateIfStmt(s ast.IfStmt) {
	cg.debugAssembly = append(cg.debugAssembly, "IF STATEMENT CONDITION:")
	var operator lexer.Token
	switch a := s.Condition.(type) {
	case ast.BinaryExpr:
		operator = a.Operator
	case ast.SymbolExpr:
		//TODO: check for 1 or 0
		cg.addError(fmt.Sprintf("checking for variable bool is not implemented - %v", a.Value))
	case ast.NumberExpr:
		cg.addError(fmt.Sprintf("just number in if condition - %v", a.Value))
	}

	cg.generateExpr(s.Condition, -1) // cmp generates inside, flags must be set

	// determines type of jump based on operator in reverse order (> -> jump less_equals to else block)
	var jmpToAltOpc uint32
	switch operator.Kind {
	case lexer.EQUALS:
		jmpToAltOpc = OP_JNE
	case lexer.NOT_EQUALS:
		jmpToAltOpc = OP_JE
	case lexer.GREATER:
		jmpToAltOpc = OP_JLE
	case lexer.LESS:
		jmpToAltOpc = OP_JGE
	case lexer.GREATER_EQUALS:
		jmpToAltOpc = OP_JL
	case lexer.LESS_EQUALS:
		jmpToAltOpc = OP_JG
	default:
		jmpToAltOpc = OP_JMP
	}
	cg.emitInstruction(jmpToAltOpc, AM_JMP_ABS, -1, -1, -1)
	addrToPatchElse := cg.nextInstructionAddr
	cg.emitImmediate(0)

	cg.debugAssembly = append(cg.debugAssembly, "IF STMT CONSEQUENSE:")
	cg.generateStmt(s.Consequent)

	var addrOfAddrToJumpAfterElse uint32 = 4294967295
	if s.Alternate != nil {
		cg.emitInstruction(OP_JMP, AM_JMP_ABS, -1, -1, -1)
		addrOfAddrToJumpAfterElse = cg.nextInstructionAddr
		cg.emitImmediate(0)
	}
	elseBlockAddr := cg.nextInstructionAddr

	if s.Alternate != nil {
		cg.debugAssembly = append(cg.debugAssembly, "IF STMT ALTERNATE:")
		cg.generateStmt(s.Alternate)
	}

	endOfElseAddr := cg.nextInstructionAddr

	cg.PatchWord(addrToPatchElse, elseBlockAddr)
	if s.Alternate != nil {
		if addrOfAddrToJumpAfterElse == 4294967295 {
			panic("addr defined not gud")
		}
		cg.PatchWord(addrOfAddrToJumpAfterElse, endOfElseAddr)
	}

}

// generateStringExpr generates code for string literals.
func (cg *CodeGenerator) generateStringExpr(e ast.StringExpr, rd int) {
	// Store string in data memory and load its byte-address (pointer) into R0.
	stringAddr := cg.addString(e.Value) // addString handles writing string to dataMemory (bytes) and alignment

	cg.emitInstruction(OP_MOV, AM_IMM_REG, rd, -1, -1) // MOV R0, #string_byte_address
	cg.emitImmediate(stringAddr)                       // Emit the byte-address as an immediate word
}

// TODO: check
// generateUnaryExpr generates code for unary operations.
func (cg *CodeGenerator) generateUnaryExpr(e ast.PrefixExpr, rd int) {
	cg.generateExpr(e.Right, RA) // Evaluate operand, result in R0

	switch e.Operator.Kind { // Access Kind from lexer.Token
	case lexer.MINUS: // Unary negation (minus sign)
		cg.emitInstruction(OP_NEG, AM_SINGLE_REG, rd, -1, -1)
	case lexer.NOT: // Logical NOT
		cg.emitInstruction(OP_NOT, AM_SINGLE_REG, rd, -1, -1)
	default:
		cg.addError(fmt.Sprintf("Unsupported unary operator: %s", e.Operator.Value)) // Use Operator.Value
	}
}

// TODO: check
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
		cg.generateExpr(e.Arguments[0], RA)                     // Evaluate argument, result in R0 (should contain value or string address)
		cg.emitInstruction(OP_OUT, AM_REG_PORT_IMM, -1, RA, -1) // OUT #0, R0 (assuming port 0 for console output)
		cg.emitImmediate(0)                                     // Port address for output
	case "input":
		if len(e.Arguments) != 0 {
			cg.addError("Input function expects no arguments.")
			return
		}
		cg.emitInstruction(OP_IN, AM_IMM_PORT_REG, RA, -1, -1) // IN R0, #0 (assuming port 0 for console input)
		cg.emitImmediate(0)                                    // Port address for input
	default:
		// For user-defined functions, lookup the function symbol and CALL its address.
		cg.addError(fmt.Sprintf("Unsupported or undeclared function call: %s", calleeName))
	}
}

// TODO: check
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

// TODO: check
// generateReturnStmt handles return statements. (Stub)
func (cg *CodeGenerator) generateReturnStmt(s ast.ReturnStmt) {
	// TODO: Implement return statement.
	// This would involve:
	// 1. Evaluating the return expression (if any) and placing result in a designated return register (e.g., R0).
	// 2. Deallocating local stack frame.
	// 3. Emitting a RET instruction.
	if s.Expr != nil {
		cg.generateExpr(s.Expr, RA) // Evaluate return value into R0
	}
	cg.emitInstruction(OP_RET, AM_NO_OPERANDS, -1, -1, -1)
	cg.addError("Return statement code generation not fully implemented (stack cleanup).")
}
