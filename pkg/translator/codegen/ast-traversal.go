package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/isa"
	"github.com/sanity-io/litter"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

func (cg *CodeGenerator) VisitProgram(p *ast.BlockStmt) {
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

// generateBlockStmt handles code blocks
func (cg *CodeGenerator) generateBlockStmt(s ast.BlockStmt) {
	for _, stmt := range s.Body {
		cg.generateStmt(stmt)
	}
}

func (cg *CodeGenerator) generateStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case ast.VarDeclarationStmt:
		cg.genVarDeclStmt(s)
	case ast.ExpressionStmt:
		cg.genEx(s.Expression, isa.RA)
	case ast.BlockStmt:
		cg.generateBlockStmt(s)
	case ast.IfStmt:
		cg.genIfStmt(s)
	case ast.PrintStmt:
		cg.genPrintStmt(s)
	case ast.WhileStmt:
		cg.generateWhileStmt(s)
	case ast.InterruptionStmt:
		cg.generateInterStmt(s)
	case ast.IntOnStmt:
		cg.genIntOnStmt()
	case ast.IntOffStmt:
		cg.genIntOffStmt()
	default:
		cg.addError(fmt.Sprintf("Unsupported statement type: %T", s))
	}
}

func (cg *CodeGenerator) genIntOffStmt() {
	cg.emitInstruction(isa.OpIntOff, isa.NoOperands, -1, -1, -1)
}

func (cg *CodeGenerator) genIntOnStmt() {
	cg.emitInstruction(isa.OpIntOn, isa.NoOperands, -1, -1, -1)
}

func (cg *CodeGenerator) generateInterStmt(s ast.InterruptionStmt) {
	irqN := s.IrqNumber
	if irqN > maxInterrupts {
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

	cg.emitInstruction(isa.OpIRet, isa.NoOperands, isa.Register(irqN), -1, -1)
}

func (cg *CodeGenerator) generateWhileStmt(s ast.WhileStmt) {
	conditionAddr := cg.nextInstructionAddr
	cg.debugAssembly = append(cg.debugAssembly, "WHILE STATEMENT CONDITION:")

	var operator lexer.Token
	switch cond := s.Condition.(type) {
	case ast.BinaryExpr:
		operator = cond.Operator
	default:
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
	case lexer.GreaterEquals:
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
	cg.emitImmediate(conditionAddr)

	afterLoopAddr := cg.nextInstructionAddr
	cg.PatchWord(addrToPatchEnd, afterLoopAddr)
	cg.debugAssembly = append(cg.debugAssembly, " # END OF WHILE STMT")
}

func (cg *CodeGenerator) genPrintStmt(s ast.PrintStmt) {
	cg.debugAssembly = append(cg.debugAssembly, "PRINT STMT")
	switch arg := s.Argument.(type) {
	case ast.StringExpr:
		cg.genStringExPl1(arg, isa.ROutAddr)
		strLen := len(arg.Value)
		if strLen > 255 {
			cg.addError(fmt.Sprintf("String length cannot be more than 1 byte (255): %d - %s", strLen, arg.Value))
			return
		}
		cg.emitMov(isa.MvImmReg, isa.RC, isa.Register(strLen), -1)

		cmpAddr := cg.nextInstructionAddr
		cg.emitInstruction(isa.OpCmp, isa.RegReg, -1, isa.RC, isa.ZERO)
		cg.emitInstruction(isa.OpJe, isa.JAbsAddr, -1, -1, -1)
		jToEndAddr := cg.ReserveWord()
		cg.emitMov(isa.MvByteRegIndToReg, isa.ROutData, isa.ROutAddr, -1)
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
			cg.addError(fmt.Sprintf("Undeclared variable in print expr: %s; %v", arg.Value, litter.Sdump(cg.scopeStack)))
			return
		}

		if !symbol.IsStr && !symbol.IsLong {
			cg.genEx(arg, isa.ROutData)
			cg.emitInstruction(isa.OpOut, isa.DigitM, isa.PortD, -1, -1)
			return
		} else if symbol.IsLong {

			cg.emitMov(isa.MvImmReg, isa.ROutAddr, isa.Register(symbol.AbsAddress), -1)
			cg.emitInstruction(isa.OpOut, isa.LongM, isa.PortL, -1, -1)
			return

		} else {
			cg.genEx(s.Argument, isa.ROutAddr)
			cg.emitMov(isa.MvRegIndToReg, isa.RC, isa.ROutAddr, -1) // mov rc <- mem[routaddr]
			cg.emitInstruction(isa.OpAnd, isa.ImmReg, isa.RC, isa.RC, -1)
			cg.emitImmediate(0xFF)
			// add 1 to routaddr bcs generateExpr will store ptr to str len initially
			cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.ROutAddr, isa.ROutAddr, -1)
			cg.emitImmediate(1)
			// routaddr = addr of str 1 char
			// rcounter addr len

			cmpAddr := cg.nextInstructionAddr
			cg.emitInstruction(isa.OpCmp, isa.RegReg, -1, isa.RC, isa.ZERO)
			cg.emitInstruction(isa.OpJe, isa.JAbsAddr, -1, -1, -1)
			jToEndAddr := cg.ReserveWord()
			cg.emitMov(isa.MvByteRegIndToReg, isa.ROutData, isa.ROutAddr, -1)

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
		cg.emitInstruction(isa.OpOut, isa.NoOperands, isa.PortD, -1, -1)

	default:
		cg.addError(fmt.Sprintf("Unsupported argument type for print: %T", arg))
	}
}

// genEx generates code for a given expression, leaving its result in specified register.
func (cg *CodeGenerator) genEx(expr ast.Expr, rd isa.Register) {
	switch e := expr.(type) {
	case ast.NumberExpr:
		cg.emitMov(isa.MvImmReg, rd, isa.Register(e.Value), -1)
	case ast.LongNumberExpr:
		cg.addError("generating long expr is not supported")

	case ast.ArrayIndexEx:
		cg.genArrayAddress(e, isa.RAddr)
		cg.emitMov(isa.MvByteRegIndToReg, rd, isa.RAddr, -1)

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
		case lexer.EQUALS, lexer.NotEquals, lexer.GREATER, lexer.GreaterEquals, lexer.LESS, lexer.LessEquals:
			opcode = isa.OpCmp
			rd = -1
		default:
			cg.addError(fmt.Sprintf("Unsupported binary operator: %s", e.Operator.Value))
			return
		}
		if opcode != isa.OpCmp {
			cg.emitInstruction(opcode, isa.MathRRR, rd, isa.RM1, isa.RM2)
		} else {
			cg.emitInstruction(opcode, isa.RegReg, rd, isa.RM1, isa.RM2)

		}

	case ast.SymbolExpr:
		symbol, found := cg.lookupSymbol(e.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable in assign expr: %s", e.Value))
			if rd != -1 {
				cg.emitInstruction(isa.OpMov, isa.MvImmReg, rd, -1, -1)
				cg.emitImmediate(0)
			}
			return
		}
		cg.emitInstruction(isa.OpMov, isa.MvMemReg, rd, -1, -1)
		cg.emitImmediate(symbol.AbsAddress)
	case ast.StringExpr:
		cg.genStringExPl1(e, rd)
	case ast.CallExpr:
		switch e.Name {
		case lexer.TokenKindString(lexer.ADDSTR):
			cg.genAddStrc(e, rd)
		default:
			cg.addError(fmt.Sprintf("unknown func name %s", e.Name))
		}
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

func (cg *CodeGenerator) FindSymbol(arg ast.SymbolExpr) *SymbolEntry {
	if s1, found := cg.currentScope().symbols[arg.Value]; found {
		return &s1
	}
	panic("undeclared variable")
}
func (cg *CodeGenerator) FindSymbolFromEx(arg ast.Expr) *SymbolEntry {
	switch e := arg.(type) {
	case ast.SymbolExpr:
		if s1, found := cg.currentScope().symbols[e.Value]; found {
			return &s1
		}
		panic("undeclared variable")
	default:
		panic("unknown")
	}
}

func (cg *CodeGenerator) emitPushReg(reg isa.Register) {
	cg.emitInstruction(isa.OpPush, isa.SingleRegMode, -1, reg, -1)
}

func (cg *CodeGenerator) emitPopToReg(reg isa.Register) {
	cg.emitInstruction(isa.OpPop, isa.SingleRegMode, reg, -1, -1)
}

// if rd is -1 then it will not move read value to destination register, instead just keep it in RInData
func (cg *CodeGenerator) genReadChEx(rd isa.Register) {
	cg.debugAssembly = append(cg.debugAssembly, "READ_CHAR EXPR")
	cg.emitInstruction(isa.OpIn, isa.ByteM, isa.PortCh, -1, -1)
	if rd != -1 {
		cg.emitMov(isa.MvRegReg, rd, isa.RInData, -1)
	}
}

// if rd is RInData then it will not move read value to destination register
// instead, just keep it in RInData
func (cg *CodeGenerator) genReadIntEx(rd isa.Register) {
	cg.debugAssembly = append(cg.debugAssembly, "READ DIGIT EXPR")
	cg.emitInstruction(isa.OpIn, isa.DigitM, isa.PortD, -1, -1)
	if rd != isa.RInData {
		cg.emitMov(isa.MvRegReg, rd, isa.RInData, -1)
	}
}

func (cg *CodeGenerator) genVarDeclStmt(s ast.VarDeclarationStmt) {
	if _, found := cg.currentScope().symbols[s.Identifier]; found {
		cg.addError(fmt.Sprintf("Variable '%s' already declared ", s.Identifier))
		return
	}

	symbolEntry := SymbolEntry{
		Name: s.Identifier,
	}

	if s.AssignedValue != nil {
		switch assignedVal := s.AssignedValue.(type) {
		case ast.LongNumberExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes * 2
			symbolEntry.NumberValue = int32(assignedVal.Value)
			symbolEntry.LongValue = assignedVal.Value
			symbolEntry.IsLong = true

			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addLongData(symbolEntry.LongValue)
			cg.addSymbolToScope(symbolEntry)
		case ast.NumberExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = assignedVal.Value

			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addNumberData(assignedVal.Value)
			cg.addSymbolToScope(symbolEntry)

		case ast.StringExpr:
			strAddr := cg.addString(assignedVal.Value)
			ptrAddr := cg.addNumberData(int32(strAddr))
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = int32(strAddr)
			symbolEntry.AbsAddress = ptrAddr
			symbolEntry.IsStr = true
			symbolEntry.MemoryArea = "data"
			cg.addSymbolToScope(symbolEntry)
		case ast.ReadChExpr:
			strAddr := cg.addString("")
			ptrAddr := cg.addNumberData(int32(strAddr))

			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.NumberValue = int32(strAddr)
			symbolEntry.AbsAddress = ptrAddr
			symbolEntry.IsStr = true
			symbolEntry.IsRead = true
			symbolEntry.MemoryArea = "data"
			cg.addSymbolToScope(symbolEntry)
			cg.dataMemory[strAddr] = 1 // read 1 char -> len = 1

			cg.genReadChEx(-1)
			cg.emitInstruction(isa.OpMov, isa.MvRegLowToMem, -1, isa.RInData, -1)
			cg.emitImmediate(strAddr + 1) // store after len byte
			return

		case ast.ReadIntExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.IsStr = false
			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addNumberData(0)
			cg.addSymbolToScope(symbolEntry)

			cg.genReadIntEx(isa.RInData)

			cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, isa.RInData, -1)
			cg.emitImmediate(symbolEntry.AbsAddress)
			return

		case ast.BinaryExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes

			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addNumberData(0)

			cg.addSymbolToScope(symbolEntry)

			assign := ast.AssignmentExpr{
				Assigne:       ast.SymbolExpr{Value: s.Identifier},
				AssignedValue: s.AssignedValue,
			}
			cg.genAssignEx(assign, isa.RA)
			return
		case ast.PrefixExpr:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes

			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addNumberData(0)
			cg.addSymbolToScope(symbolEntry)

			assign := ast.AssignmentExpr{
				Assigne:       ast.SymbolExpr{Value: s.Identifier},
				AssignedValue: s.AssignedValue,
			}
			cg.genAssignEx(assign, isa.RA)
			return
		case ast.ListEx:
			listPtr := cg.nextDataAddr
			cg.dataMemory = append(cg.dataMemory, make([]byte, assignedVal.Size)...)
			cg.nextDataAddr += uint32(assignedVal.Size)
			// allignDataMem(cg)

			ptrAddr := cg.addNumberData(int32(listPtr))

			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.AbsAddress = ptrAddr
			symbolEntry.MemoryArea = "data"
			symbolEntry.IsStr = false
			cg.addSymbolToScope(symbolEntry)
		case ast.ArrayIndexEx:
			symbolEntry.Type = ast.IntType
			symbolEntry.SizeInBytes = WordSizeBytes
			symbolEntry.MemoryArea = "data"
			symbolEntry.AbsAddress = cg.addNumberData(0)
			cg.addSymbolToScope(symbolEntry)

			cg.genEx(assignedVal, isa.RT2)

			cg.emitInstruction(isa.OpMov, isa.MvRegLowToMem, -1, isa.RT2, -1)
			cg.emitImmediate(symbolEntry.AbsAddress)
		case ast.CallExpr:
			switch assignedVal.Name {
			case lexer.TokenKindString(lexer.ADDSTR):
				if len(assignedVal.Args) != 2 {
					cg.addError(fmt.Sprintf("addStr( , ) must have 2 arguments, got %d", len(assignedVal.Args)))
					return
				}
				ptrAddr := cg.addNumberData(0)
				sym := SymbolEntry{
					Name:        s.Identifier,
					Type:        ast.IntType,
					SizeInBytes: WordSizeBytes,
					AbsAddress:  ptrAddr,
					IsStr:       true,
					MemoryArea:  "data",
				}
				cg.addSymbolToScope(sym)

				cg.genAddStrc(assignedVal, isa.RA)

				cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, isa.RA, -1)
				cg.emitImmediate(ptrAddr)
			case lexer.TokenKindString(lexer.ADDL):
				if len(assignedVal.Args) != 2 {
					cg.addError(fmt.Sprintf("%s( , ) must have 2 arguments, got %d", lexer.TokenKindString(lexer.ADDL), len(assignedVal.Args)))
					return
				}
				if _, ok := assignedVal.Args[0].(ast.SymbolExpr); !ok {
					panic("argument must be variable")
				}
				if _, ok := assignedVal.Args[1].(ast.SymbolExpr); !ok {
					panic("")
				}
				arg1e := assignedVal.Args[0].(ast.SymbolExpr)
				arg2e := assignedVal.Args[1].(ast.SymbolExpr)
				s1, found1 := cg.lookupSymbol(arg1e.Value)
				s2, found2 := cg.lookupSymbol(arg2e.Value)
				if !found1 || !found2 {
					cg.addError(fmt.Sprintf("Undeclared variable in print expr: %s or %s; %v", arg1e.Value, arg2e.Value, litter.Sdump(cg.scopeStack)))
					return
				}
				ptrAddr := cg.addLongData(s1.LongValue + s2.LongValue)
				sym := SymbolEntry{
					Name:        s.Identifier,
					Type:        ast.IntType,
					SizeInBytes: WordSizeBytes * 2,
					AbsAddress:  ptrAddr,
					IsLong:      true,
					MemoryArea:  "data",
				}
				cg.addSymbolToScope(sym)

			default:
				cg.addError(fmt.Sprintf("unknown function name %v", assignedVal.Name))
			}

		default:
			cg.addError(fmt.Sprintf("unknown case of generating var declaration - %T", assignedVal))
		}
	} else {
		cg.addError(fmt.Sprintf("All variables should be initialized: %s - is undefined", s.Identifier))
		return
	}

}

func (cg *CodeGenerator) genAddStrc(call ast.CallExpr, rd isa.Register) {
	var s1Len byte
	var s2Len byte
	var s1str string
	var s2str string
	if len(call.Args) != 2 {
		cg.addError("addStr() needs 2 arguments")
		return
	}
	switch a1 := call.Args[0].(type) {
	case ast.SymbolExpr:
		s1, found := cg.lookupSymbol(a1.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in addStr func.", a1.Value))
			return
		}
		s1Addr := s1.NumberValue

		s1Len = cg.dataMemory[s1Addr]
		s1bytes := cg.dataMemory[s1Addr+1 : s1Addr+1+int32(s1Len)]
		s1str = string(s1bytes)
	default:
		cg.addError("Using not symbol expr in addStr is not supported")
		return
	}

	switch a2 := call.Args[1].(type) {
	case ast.SymbolExpr:
		s2, found := cg.lookupSymbol(a2.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in addStr func.", a2.Value))
			return
		}
		s2Addr := s2.NumberValue
		s2Len = cg.dataMemory[s2Addr]
		s2bytes := cg.dataMemory[s2Addr+1 : s2Addr+1+int32(s2Len)]
		s2str = string(s2bytes)
	default:
		cg.addError("Using not symbol expr in addStr is not supported")
		return
	}

	newStr := s1str + s2str
	cg.genStringEx(ast.StringExpr{Value: newStr}, rd)
}

func (cg *CodeGenerator) genAssignEx(e ast.AssignmentExpr, rd isa.Register) {
	// Evaluate the right-hand side expression, result is left in rd
	switch r := e.AssignedValue.(type) {
	case ast.ReadChExpr:
		trg := cg.FindSymbolFromEx(e.Assigne)
		trg.IsRead = true
		trg.IsStr = true
		trg.IsLong = false
		cg.emitInstruction(isa.OpIn, isa.ByteM, isa.PortCh, -1, -1)

		cg.emitInstruction(isa.OpMov, isa.MvRegLowToMem, -1, isa.RInData, -1)
		cg.emitImmediate(uint32(trg.NumberValue))
		return

	case ast.CallExpr:
		switch r.Name {
		case lexer.TokenKindString(lexer.ADDL):
			targetS := cg.FindSymbolFromEx(e.Assigne)
			cg.genAddLongAssign(r.Args, e.Assigne.(ast.SymbolExpr))

			cg.emitInstruction(isa.OpMov, isa.MvRegIndToReg, isa.RA, rd, -1)
			cg.emitMov(isa.MvRegMem, isa.Register(targetS.AbsAddress), isa.RA, -1)

			cg.emitInstruction(isa.OpAdd, isa.MathRIR, rd, rd, -1)
			cg.emitImmediate(4)
			cg.emitInstruction(isa.OpMov, isa.MvRegIndToReg, isa.RA, rd, -1)
			cg.emitMov(isa.MvRegMem, isa.Register(targetS.AbsAddress+4), isa.RA, -1)
			return
		default:
			cg.addError(fmt.Sprintf("unknown func name %s", r.Name))
		}

	default:
		cg.genEx(e.AssignedValue, rd)
	}

	switch target := e.Assigne.(type) {
	case ast.SymbolExpr:
		symbol, found := cg.lookupSymbol(target.Value)
		if !found {
			cg.addError(fmt.Sprintf("Undeclared variable '%s' used in assignment.", target.Value))
			return
		}

		cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, rd, -1)
		cg.emitImmediate(symbol.AbsAddress)
	case ast.ArrayIndexEx:
		cg.genAssignArray(e, target)

	default:
		cg.addError(fmt.Sprintf("Unsupported assignment target type: %T", target))
	}
}

func (cg *CodeGenerator) genAssignArray(e ast.AssignmentExpr, target ast.ArrayIndexEx) {
	regWithAddr := isa.RAddr
	regWithVal := isa.RA
	cg.genEx(e.AssignedValue, regWithVal)
	cg.genArrayAddress(target, regWithAddr)
	cg.emitInstruction(isa.OpMov, isa.MvLowRegToRegInd, regWithAddr, regWithVal, -1)
}

// calculates addr of array element and stores it in rd
func (cg *CodeGenerator) genArrayAddress(ix ast.ArrayIndexEx, rd isa.Register) {
	cg.genEx(ix.Target, isa.RM1)
	cg.genEx(ix.Index, isa.RM2)

	cg.emitInstruction(isa.OpAdd, isa.MathRRR, rd, isa.RM1, isa.RM2)
}

func (cg *CodeGenerator) genIfStmt(s ast.IfStmt) {
	cg.debugAssembly = append(cg.debugAssembly, "IF STATEMENT CONDITION:")
	var operator lexer.Token
	switch a := s.Condition.(type) {
	case ast.BinaryExpr:
		operator = a.Operator
	case ast.SymbolExpr:
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
	case lexer.GreaterEquals:
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

// genStringExPl1 generates code to move string literal's ptr to rd. rd <- strAddr + 1
func (cg *CodeGenerator) genStringExPl1(e ast.StringExpr, rd isa.Register) {
	stringAddr := cg.addString(e.Value)

	cg.emitInstruction(isa.OpMov, isa.MvImmReg, rd, -1, -1)
	cg.emitImmediate(stringAddr + 1)
}

// genStringEx generates code to move string len ptr to rd. rd <- strAddr
func (cg *CodeGenerator) genStringEx(e ast.StringExpr, rd isa.Register) {
	stringAddr := cg.addString(e.Value)

	cg.emitInstruction(isa.OpMov, isa.MvImmReg, rd, -1, -1)
	cg.emitImmediate(stringAddr)
}

func (cg *CodeGenerator) genAddLongAssign(args []ast.Expr, target ast.SymbolExpr) {
	addrA := cg.FindSymbol(args[0].(ast.SymbolExpr)).AbsAddress
	addrB := cg.FindSymbol(args[1].(ast.SymbolExpr)).AbsAddress
	addrT := cg.FindSymbol(target).AbsAddress

	cg.emitMov(isa.MvImmReg, isa.RM1, isa.Register(addrA), -1)
	cg.emitMov(isa.MvImmReg, isa.RM2, isa.Register(addrB), -1)

	cg.emitMov(isa.MvRegIndToReg, isa.RT, isa.RM1, -1)
	cg.emitMov(isa.MvRegIndToReg, isa.R6, isa.RM2, -1)
	cg.emitInstruction(isa.OpAdd, isa.MathRRR, isa.RA, isa.RT, isa.R6)

	cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.RM1, isa.RM1, -1)
	cg.emitImmediate(4)
	cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.RM2, isa.RM2, -1)
	cg.emitImmediate(4)

	cg.emitMov(isa.MvRegIndToReg, isa.RT, isa.RM1, -1)
	cg.emitMov(isa.MvRegIndToReg, isa.R6, isa.RM2, -1)
	cg.emitInstruction(isa.OpAdd, isa.MathRRR, isa.RT2, isa.RT, isa.R6)

	cg.emitInstruction(isa.OpJcc, isa.JAbsAddr, -1, -1, -1)
	skip := cg.ReserveWord()
	cg.emitInstruction(isa.OpAdd, isa.MathRIR, isa.RT2, isa.RT2, -1)
	cg.emitImmediate(1)
	cg.PatchWord(skip, cg.nextInstructionAddr)

	cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, isa.RA, -1)
	cg.emitImmediate(addrT)
	cg.emitInstruction(isa.OpMov, isa.MvRegMem, -1, isa.RT2, -1)
	cg.emitImmediate(addrT + 4)
}
