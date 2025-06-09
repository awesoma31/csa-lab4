package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
	"github.com/sanity-io/litter"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

// generateStmt generates code for a given statement.
func (cg *CodeGenerator) generateStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case ast.VarDeclarationStmt:
		cg.generateVarDeclStmt(s)
	case ast.ExpressionStmt:
		cg.genEx(s.Expression, isa.RA)
	case ast.BlockStmt:
		cg.generateBlockStmt(s)
	case ast.IfStmt:
		cg.generateIfStmt(s)
	case ast.PrintStmt:
		cg.genPrintStmt(s)
	case ast.WhileStmt:
		cg.generateWhileStmt(s)
	case ast.InterruptionStmt:
		cg.generateInterStmt(s)
	default:
		cg.addError(fmt.Sprintf("Unsupported statement type: %T", s))
	}
}

func (cg *CodeGenerator) generateInterStmt(s ast.InterruptionStmt) {
	irqN := s.IrqNumber
	if irqN > int(maxInterrupts) {
		cg.addError(fmt.Sprintf("Invalid interruption number, must be between 0 and %d", maxInterrupts))
		return
	}
	cg.debugAssembly = append(cg.debugAssembly, fmt.Sprintf("INTERRUPTION %d STMT", irqN))

	cg.instructionMemory[irqN] = cg.nextInstructionAddr

	switch t := s.Body.(type) {
	case ast.BlockStmt:
		cg.generateBlockStmt(t)
	default:
		cg.addError(fmt.Sprint("interruption body must be block stmt, got: ", litter.Sdump(t)))
	}

	cg.emitInstruction(isa.OpIRet, isa.NoOperands, irqN, -1, -1)
}

func (cg *CodeGenerator) generateWhileStmt(s ast.WhileStmt) {
	conditionAddr := cg.nextInstructionAddr
	cg.debugAssembly = append(cg.debugAssembly, "WHILE STATEMENT CONDITION:")

	var operator lexer.Token
	switch cond := s.Condition.(type) {
	case ast.BinaryExpr:
		operator = cond.Operator
	default:
		// TODO: Добавить обработку для других типов условий (например, bool переменных)
		cg.addError(fmt.Sprintf("unsupported condition type in while loop: %T", cond))
		return
	}

	cg.genEx(s.Condition, -1)

	var jmpToEndOpc uint32
	switch operator.Kind {
	case lexer.EQUALS:
		jmpToEndOpc = isa.OpJne // Jump if Not Equal
	case lexer.NotEquals:
		jmpToEndOpc = isa.OpJe // Jump if Equal
	case lexer.GREATER:
		jmpToEndOpc = isa.OpJle // Jump if Less or Equal
	case lexer.LESS:
		jmpToEndOpc = isa.OpJge // Jump if Greater or Equal
	case lexer.GREATER_EQUALS:
		jmpToEndOpc = isa.OpJl // Jump if Less
	case lexer.LessEquals:
		jmpToEndOpc = isa.OpJg // Jump if Greater
	default:
		cg.addError(fmt.Sprintf("unsupported operator in while condition: %s", operator.Value))
		jmpToEndOpc = isa.OpJmp
	}

	cg.emitInstruction(jmpToEndOpc, isa.JAbsAddr, -1, -1, -1)
	addrToPatchEnd := cg.ReserveWord()

	cg.debugAssembly = append(cg.debugAssembly, "WHILE STMT BODY:")
	cg.generateStmt(s.Body)

	cg.emitInstruction(isa.OpJmp, isa.JAbsAddr, -1, -1, -1)
	cg.emitImmediate(conditionAddr) // Адрес для возврата к проверке условия

	afterLoopAddr := cg.nextInstructionAddr
	cg.PatchWord(addrToPatchEnd, afterLoopAddr)
	cg.debugAssembly = append(cg.debugAssembly, " # END OF WHILE STMT")
}

func (cg *CodeGenerator) genPrintStmt(s ast.PrintStmt) {
	cg.debugAssembly = append(cg.debugAssembly, "PRINT STMT")
	switch arg := s.Argument.(type) {
	case ast.StringExpr:
		cg.genEx(s.Argument, isa.ROutAddr)
		strLen := len(arg.Value)
		if strLen > 255 {
			cg.addError(fmt.Sprintf("String length cannot be more than 1 byte (255): %d - %s", strLen, arg.Value))
			return
		}
		cg.emitMov(isa.MvImmReg, isa.RC, strLen, -1)

		cmpAddr := cg.nextInstructionAddr
		cg.emitInstruction(isa.OpCmp, isa.RegReg, -1, isa.RC, isa.ZERO)
		cg.emitInstruction(isa.OpJe, isa.JAbsAddr, -1, -1, -1)
		jToEndAddr := cg.ReserveWord()
		cg.emitMov(isa.MvLowRegIndReg, isa.ROutData, isa.ROutAddr, -1)
		cg.emitInstruction(isa.OpOut, isa.ByteM, isa.PortCh, -1, -1)
		cg.emitInstruction(isa.OpSub, isa.MathRIR, isa.RC, isa.RC, -1)
		cg.emitImmediate(1)
		cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.ROutAddr, isa.ROutAddr, -1)
		cg.emitImmediate(1)
		cg.emitInstruction(isa.OpJmp, isa.JAbsAddr, -1, -1, -1)
		cg.emitImmediate(cmpAddr)
		afterEndAddr := cg.nextInstructionAddr
		cg.PatchWord(jToEndAddr, afterEndAddr)

	case ast.SymbolExpr:
		symbol, found := cg.lookupSymbol(arg.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable in print expr: %s", arg.Value))
			return
		}

		if !symbol.IsStr {
			cg.genEx(arg, isa.ROutData)
			cg.emitInstruction(isa.OpOut, isa.WordM, isa.PortD, -1, -1)
		} else {
			cg.genEx(s.Argument, isa.ROutAddr)
			cg.emitMov(isa.MvRegIndReg, isa.RC, isa.ROutAddr, -1) // mov rc <- mem[routaddr]
			cg.emitInstruction(isa.OpAnd, isa.ImmReg, isa.RC, isa.RC, -1)
			cg.emitImmediate(0xFF)
			// add 1 to routaddr bcs generateExpr will store ptr to str len initially
			cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.ROutAddr, isa.ROutAddr, -1)
			cg.emitImmediate(1)
			// routaddr = addr of str 1 char
			// rcounter addr len

			_, found := cg.lookupSymbol(arg.Value)
			if !found {
				cg.addError(fmt.Sprintf("Undeclared variable in print expr: %s", arg.Value))
			}

			cmpAddr := cg.nextInstructionAddr
			cg.emitInstruction(isa.OpCmp, isa.RegReg, -1, isa.RC, isa.ZERO)
			cg.emitInstruction(isa.OpJe, isa.JAbsAddr, -1, -1, -1)
			jToEndAddr := cg.ReserveWord()
			cg.emitMov(isa.MvLowRegIndReg, isa.ROutData, isa.ROutAddr, -1)

			cg.emitInstruction(isa.OpOut, isa.ByteM, isa.PortCh, -1, -1)
			cg.emitInstruction(isa.OpSub, isa.MathRIR, isa.RC, isa.RC, -1)
			cg.emitImmediate(1)
			cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.ROutAddr, isa.ROutAddr, -1)
			cg.emitImmediate(1)
			cg.emitInstruction(isa.OpJmp, isa.JAbsAddr, -1, -1, -1)
			cg.emitImmediate(cmpAddr)

			afterEndAddr := cg.nextInstructionAddr
			cg.PatchWord(jToEndAddr, afterEndAddr)
		}

	case ast.NumberExpr:
		cg.genEx(arg, isa.ROutData)
		cg.emitInstruction(isa.OpOutB, isa.NoOperands, isa.PortD, -1, -1)

	default:
		cg.addError(fmt.Sprintf("Unsupported argument type for print: %T", arg))
	}
}

// genEx generates code for a given expression, leaving its result in specified register.
func (cg *CodeGenerator) genEx(expr ast.Expr, rd int) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		cg.emitMov(isa.MvImmReg, rd, int(e.Value), -1)

	case ast.BinaryExpr:
		cg.genEx(e.Left, isa.RM1)
		cg.emitPushReg(isa.RM1)

		cg.genEx(e.Right, isa.RM2)

		cg.emitPopToReg(isa.RM1)

		var opcode uint32
		switch e.Operator.Kind {
		case lexer.PLUS:
			opcode = isa.OpAdd
		case lexer.MINUS:
			opcode = isa.OpSub
		case lexer.STAR:
			opcode = isa.OpMul
		case lexer.SLASH:
			opcode = isa.OpDiv
		case lexer.EQUALS, lexer.NotEquals, lexer.GREATER, lexer.GREATER_EQUALS, lexer.LESS, lexer.LessEquals:
			opcode = isa.OpCmp
			rd = -1
		default:
			cg.addError(fmt.Sprintf("Unsupported binary operator: %s", e.Operator.Value))
			return
		}
		if opcode != isa.OpCmp {
			cg.emitInstruction(opcode, isa.MathRRR, rd, isa.RM1, isa.RM2) // rd = RM1 op RM2
		} else {
			cg.emitInstruction(opcode, isa.RegReg, rd, isa.RM1, isa.RM2) // rd = RM1 op RM2

		}

	case ast.SymbolExpr: // Для чтения значения переменной
		symbol, found := cg.lookupSymbol(e.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable in assign expr: %s", e.Value))
			if rd != -1 {
				cg.emitInstruction(isa.OpMov, isa.MvImmReg, rd, -1, -1)
				cg.emitImmediate(0) // Загрузить 0 в rd для восстановления
			}
			return
		}
		cg.emitInstruction(isa.OpMov, isa.MvMemReg, rd, -1, -1)
		cg.emitImmediate(symbol.AbsAddress)
	case ast.FunctionExpr:
		cg.addError("FunctionExpr code generation not implemented.")
	case ast.StringExpr:
		cg.genStringEx(e, rd)
	case ast.PrefixExpr:
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
			cg.genEx(newExpr, rd)
		default:
			panic("unimpl prefix functionality, only unary with numbers work for now")
		}
	case ast.AssignmentExpr:
		cg.genAssignEx(e, rd)
	case ast.ReadChExpr:
		cg.genReadChEx(rd)
	case ast.ReadIntExpr:
		cg.genReadIntEx(rd)
	default:
		cg.addError(fmt.Sprintf("Unsupported expression type: %T", e))
	}
}

func (cg *CodeGenerator) emitPushReg(reg int) {
	cg.emitInstruction(isa.OpPush, isa.SingleRegMode, -1, reg, -1)
}

func (cg *CodeGenerator) emitPopToReg(reg int) {
	cg.emitInstruction(isa.OpPop, isa.SingleRegMode, reg, -1, -1)
}

// if rd is RInData then it will not move read value to destination register
// instead, just keep it in RInData
func (cg *CodeGenerator) genReadChEx(rd int) {
	cg.debugAssembly = append(cg.debugAssembly, "READ_CHAR EXPR")
	cg.emitInstruction(isa.OpIn, isa.ByteM, isa.PortCh, -1, -1)
	if rd != isa.RInData {
		cg.emitMov(isa.MvRegReg, rd, isa.RInData, -1)
	}
}

// if rd is RInData then it will not move read value to destination register
// instead, just keep it in RInData
func (cg *CodeGenerator) genReadIntEx(rd int) {
	cg.debugAssembly = append(cg.debugAssembly, "READ DIGIT EXPR")
	cg.emitInstruction(isa.OpIn, isa.WordM, isa.PortD, -1, -1)
	if rd != isa.RInData {
		cg.emitMov(isa.MvRegReg, rd, isa.RInData, -1)
	}
}

// VisitProgram generates code for the entire program.
func (cg *CodeGenerator) VisitProgram(p *ast.BlockStmt) {
	// Initialize global scope
	halted := false
	cg.pushScope()

	for _, stmt := range p.Body {
		switch stmt.(type) {
		case ast.InterruptionStmt:
			if !halted {
				cg.emitInstruction(isa.OpHalt, isa.NoOperands, -1, -1, -1)
				halted = true
			}
			cg.generateStmt(stmt)
		default:
			cg.generateStmt(stmt)
		}
	}

	if !halted {
		cg.emitInstruction(isa.OpHalt, isa.NoOperands, -1, -1, -1)
	}
}

func (cg *CodeGenerator) generateVarDeclStmt(s ast.VarDeclarationStmt) {
	if _, found := cg.currentScope().symbols[s.Identifier]; found {
		cg.addError(fmt.Sprintf("Variable '%s' already declared in this scope.", s.Identifier))
		return
	}

	symbolEntry := SymbolEntry{
		Name: s.Identifier,
	}

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
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = int32(strAddr)
			symbolEntry.AbsAddress = ptrAddr
			symbolEntry.IsStr = true
			symbolEntry.MemoryArea = "data"
			cg.addSymbolToScope(symbolEntry)
		case ast.ReadChExpr:
			println("readchar")
			strAddr := cg.addString("")
			ptrAddr := cg.addNumberData(int32(strAddr))

			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = int32(strAddr)
			symbolEntry.AbsAddress = ptrAddr
			symbolEntry.IsStr = true
			symbolEntry.MemoryArea = "data"
			println("before add to scope")
			cg.addSymbolToScope(symbolEntry)
			cg.dataMemory[strAddr] = 1 // read 1 char -> len = 1

			cg.genReadChEx(isa.RInData)
			cg.emitInstruction(isa.OpMov, isa.MvRegLowMem, -1, isa.RInData, -1)
			cg.emitImmediate(strAddr + 1) // store after len byte
			return

		case ast.ReadIntExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.IsStr = false
			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addNumberData(0) // reserve space for int
			cg.addSymbolToScope(symbolEntry)

			cg.genReadIntEx(isa.RInData)

			cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, isa.RInData, -1)
			cg.emitImmediate(symbolEntry.AbsAddress)
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
			cg.genAssignEx(assign, isa.RA)
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
			cg.genAssignEx(assign, isa.RA)
			return

			// cg.generateExpr(assignedVal, RA)
		// case *ast.ReadChExpr:
		// 	//TODO: straight to mem
		// 	cg.genReadChEx(isa.RT)
		default:
			cg.addError(fmt.Sprintf("unknown case of generating var declaration - %T", assignedVal))
		}
	} else {
		cg.addError(fmt.Sprintf("All variables should be initialized: %s - is undefined", s.Identifier))
		return
	}

}

// genAssignEx generates code for assignment expressions.
func (cg *CodeGenerator) genAssignEx(e ast.AssignmentExpr, rd int) {
	// Evaluate the right-hand side expression, result is left in rd
	cg.genEx(e.AssignedValue, rd)
	switch target := e.Assigne.(type) {
	case ast.SymbolExpr:
		symbol, found := cg.lookupSymbol(target.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in assignment.", target.Value))
			return
		}

		cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, rd, -1)
		cg.emitImmediate(symbol.AbsAddress)
	default:
		cg.addError(fmt.Sprintf("Unsupported assignment target type: %T", target))
	}
}

// generateBlockStmt handles code blocks
func (cg *CodeGenerator) generateBlockStmt(s ast.BlockStmt) {
	for _, stmt := range s.Body {
		cg.generateStmt(stmt)
	}
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

	cg.genEx(s.Condition, -1) // cmp generates inside, flags must be set

	// determines type of jump based on operator in reverse order (> -> jump less_equals to else block)
	var jmpToAltOpc uint32
	switch operator.Kind {
	case lexer.EQUALS:
		jmpToAltOpc = isa.OpJne
	case lexer.NotEquals:
		jmpToAltOpc = isa.OpJe
	case lexer.GREATER:
		jmpToAltOpc = isa.OpJle
	case lexer.LESS:
		jmpToAltOpc = isa.OpJge
	case lexer.GREATER_EQUALS:
		jmpToAltOpc = isa.OpJl
	case lexer.LessEquals:
		jmpToAltOpc = isa.OpJg
	default:
		jmpToAltOpc = isa.OpJmp
	}
	cg.emitInstruction(jmpToAltOpc, isa.JAbsAddr, -1, -1, -1)
	addrToPatchElse := cg.nextInstructionAddr
	cg.emitImmediate(0)

	cg.debugAssembly = append(cg.debugAssembly, "IF STMT CONSEQUENCE:")
	cg.generateStmt(s.Consequent)

	var addrOfAddrToJumpAfterElse uint32 = 4294967295
	if s.Alternate != nil {
		cg.emitInstruction(isa.OpJmp, isa.JAbsAddr, -1, -1, -1)
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

// genStringEx generates code for string literals.
func (cg *CodeGenerator) genStringEx(e ast.StringExpr, rd int) {
	stringAddr := cg.addString(e.Value)

	cg.emitInstruction(isa.OpMov, isa.MvImmReg, rd, -1, -1)
	cg.emitImmediate(stringAddr + 1)
}

// TODO: check

// generateUnaryExpr generates code for unary operations.
func (cg *CodeGenerator) generateUnaryExpr(e ast.PrefixExpr, rd int) {
	cg.genEx(e.Right, isa.RA) // Evaluate operand, result in R0

	switch e.Operator.Kind { // Access Kind from lexer.Token
	case lexer.MINUS: // Unary negation (minus sign)
		cg.emitInstruction(isa.OpNeg, isa.SingleRegMode, rd, -1, -1)
	case lexer.NOT: // Logical NOT
		cg.emitInstruction(isa.OpNot, isa.SingleRegMode, rd, -1, -1)
	default:
		cg.addError(fmt.Sprintf("Unsupported unary operator: %s", e.Operator.Value)) // Use Operator.Value
	}
}
