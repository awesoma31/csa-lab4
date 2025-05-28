package codegen

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/cmd/translator/ast"
)

type CodeGen struct {
	Sym    *SymTab
	Instr  []Word
	Report []string
	regCnt int // простой рег-аллокатор: R0..R7
	Data   *[]Word
}

func NewCG() *CodeGen { return &CodeGen{Sym: NewSymTab()} }

func (cg *CodeGen) fresh() int {
	r := cg.regCnt
	cg.regCnt = (cg.regCnt + 1) % 8
	return r
}

func (cg *CodeGen) emit(w Word, mnemonic string) {
	addr := len(cg.Instr)
	cg.Instr = append(cg.Instr, w)
	cg.Report = append(cg.Report,
		fmt.Sprintf("%04X - %08X - %s", addr, w, mnemonic))
}

// ---- обработка узлов AST ----
func (cg *CodeGen) genNumber(n *ast.NumberExpr) int {
	r := cg.fresh()
	w0 := Pack(OPC_ADDN, 1, Word(r), 0, 0) // ADDN dst, imm
	desc := ImmDescriptor(uint32(n.Value))
	cg.emit(w0, fmt.Sprintf("load #%d -> R%d", int(n.Value), r))
	cg.emit(desc, "imm")
	return r
}

func (cg *CodeGen) genVarDecl(v *ast.VarDeclarationStmt) {
	// allocate in data memory
	if _, ok := cg.Sym.vars[v.Identifier]; ok {
		panic("duplicate")
	}
	// init value
	var init uint32
	if num, ok := v.AssignedValue.(*ast.NumberExpr); ok {
		init = uint32(num.Value)
	}
	addr := cg.Sym.Alloc(v.Identifier, init, cg.Data)
	// если инициализация не константная, нужно сгенерировать код Store
	if _, ok := v.AssignedValue.(*ast.NumberExpr); !ok {
		r := cg.genExpr(v.AssignedValue)
		w0 := Pack(OPC_ST, 1, Word(r), 0, 0)
		cg.emit(w0, fmt.Sprintf("ST R%d -> [%04X]", r, addr))
		cg.emit(MemDescriptor(addr), "mem")
	}
}

func (cg *CodeGen) genExpr(e ast.Expr) int {
	switch n := e.(type) {
	case *ast.NumberExpr:
		return cg.genNumber(n)
	case *ast.SymbolExpr:
		addr := cg.Sym.AddrOf(n.Value)
		r := cg.fresh()
		w0 := Pack(OPC_LD, 1, Word(r), 0, 0)
		cg.emit(w0, fmt.Sprintf("LD [%04X] -> R%d", addr, r))
		cg.emit(MemDescriptor(addr), "mem")
		return r
	case *ast.BinaryExpr:
		left := cg.genExpr(n.Left)
		right := cg.genExpr(n.Right)
		dst := cg.fresh()
		var opc Word
		var mnem string
		switch n.Operator.Value {
		case "+":
			opc = OPC_ADDN
			mnem = "ADD"
		case "-":
			opc = OPC_SUB
			mnem = "SUB"
		case "*":
			opc = OPC_MULN
			mnem = "MUL"
		default:
			panic("op")
		}
		w0 := Pack(opc, 2, Word(dst), 0, 0)
		cg.emit(w0, fmt.Sprintf("%s R%d,R%d -> R%d", mnem, left, right, dst))
		cg.emit(RegDescriptor(Word(left)), "src1")
		cg.emit(RegDescriptor(Word(right)), "src2")
		return dst
	default:
		panic("expr")
	}
}
