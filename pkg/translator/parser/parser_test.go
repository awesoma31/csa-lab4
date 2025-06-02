package parser_test

import (
	"testing"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
)

// ─────────────────────────────────────────────────────────────────────────────
//
//	src:
//	    let b = 256;
//	    let c = 605;
//	    let d = 2000;
//	    let a = (2 + b) * c - d;
//
// ─────────────────────────────────────────────────────────────────────────────
//
//	Ожидаемая форма AST для выражения переменной a:
//
//	      (-)
//	    /      \
//	  (*)       d
//	 /   \
//
// ( + )   c
// /   \
// 2     b
// ─────────────────────────────────────────────────────────────────────────────
func TestComplexMathExpressionTree(t *testing.T) {
	src := `
		let b = 256;
		let c = 605;
		let d = 2000;
		let a = (2 + b) * c - d;
	`

	prog, errs := parser.Parse(src)
	if len(errs) != 0 {
		t.Fatalf("parser returned errors: %v", errs)
	}

	if len(prog.Body) != 4 {
		t.Fatalf("want 4 top-level statements, got %d", len(prog.Body))
	}

	last, ok := prog.Body[3].(ast.VarDeclarationStmt)
	if !ok || last.Identifier != "a" {
		t.Fatalf("4-й stmt должен быть VarDeclaration a, got %#v", prog.Body[3])
	}

	// ─── 3. корень выражения – вычитание (… - d) ────────────────────────────
	sub, ok := last.AssignedValue.(ast.BinaryExpr)
	if !ok || sub.Operator.Kind != lexer.MINUS {
		t.Fatalf("a := must be MINUS BinaryExpr, got %#v", last.AssignedValue)
	}

	// ─── 4. RHS вычитания – идентификатор d ─────────────────────────────────
	if rhs, ok := sub.Right.(ast.SymbolExpr); !ok || rhs.Value != "d" {
		t.Fatalf("right operand should be symbol d, got %#v", sub.Right)
	}

	// ─── 5. LHS вычитания – умножение ((2+b)*c) ─────────────────────────────
	mul, ok := sub.Left.(ast.BinaryExpr)
	if !ok || mul.Operator.Kind != lexer.STAR {
		t.Fatalf("left operand should be STAR BinaryExpr, got %#v", sub.Left)
	}

	// ─── 6. RHS умножения – идентификатор c ─────────────────────────────────
	if rhs, ok := mul.Right.(ast.SymbolExpr); !ok || rhs.Value != "c" {
		t.Fatalf("right operand of * should be symbol c, got %#v", mul.Right)
	}

	// ─── 7. LHS умножения – сложение (2+b) ──────────────────────────────────
	add, ok := mul.Left.(ast.BinaryExpr)
	if !ok || add.Operator.Kind != lexer.PLUS {
		t.Fatalf("left operand of * should be PLUS BinaryExpr, got %#v", mul.Left)
	}

	// 7.1 левый операнд сложения – число 2
	if lNum, ok := add.Left.(ast.NumberExpr); !ok || lNum.Value != 2 {
		t.Fatalf("left operand of + should be number 2, got %#v", add.Left)
	}

	// 7.2 правый операнд сложения – символ b
	if rSym, ok := add.Right.(ast.SymbolExpr); !ok || rSym.Value != "b" {
		t.Fatalf("right operand of + should be symbol b, got %#v", add.Right)
	}
}
