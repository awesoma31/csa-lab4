package parser_test

import (
	"testing"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
	"github.com/awesoma31/csa-lab4/pkg/translator/parser"
	"github.com/google/go-cmp/cmp"
)

const src = `intOff;
			let a = (5 + 3) * 2 - 10 / 5 + 4;
			print(a);
`

func TestAstBuild(t *testing.T) {
	got, pErr := parser.Parse(src)
	if len(pErr) != 0 {
		t.Fatal("parse errors")
	}

	want := ast.BlockStmt{
		Body: []ast.Stmt{
			ast.IntOffStmt{},
			ast.VarDeclarationStmt{
				Identifier:    "a",
				AssignedValue: buildExpectedExpr(), // (5+3)*2 - 10/5 + 4
			},
			ast.PrintStmt{Argument: ast.SymbolExpr{Value: "a"}}, // print(a);
		},
	}

	var opts []cmp.Option

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("AST mismatch (-want +got):\n%s", diff)
	}
}

func buildExpectedExpr() ast.Expr {
	// 5 + 3
	add53 := ast.BinaryExpr{
		Left:     ast.NumberExpr{Value: 5},
		Operator: lexer.Token{Kind: lexer.PLUS, Value: "+"},
		Right:    ast.NumberExpr{Value: 3},
	}
	// (5+3) * 2
	mul := ast.BinaryExpr{
		Left:     add53,
		Operator: lexer.Token{Kind: lexer.STAR, Value: "*"},
		Right:    ast.NumberExpr{Value: 2},
	}
	// 10 / 5
	div := ast.BinaryExpr{
		Left:     ast.NumberExpr{Value: 10},
		Operator: lexer.Token{Kind: lexer.SLASH, Value: "/"},
		Right:    ast.NumberExpr{Value: 5},
	}
	// (mul) - (div)
	sub := ast.BinaryExpr{
		Left:     mul,
		Operator: lexer.Token{Kind: lexer.MINUS, Value: "-"},
		Right:    div,
	}
	// (sub) + 4
	return ast.BinaryExpr{
		Left:     sub,
		Operator: lexer.Token{Kind: lexer.PLUS, Value: "+"},
		Right:    ast.NumberExpr{Value: 4},
	}
}

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
		t.Fatalf("4-й stmt must VarDeclaration, got %#v", prog.Body[3])
	}

	sub, ok := last.AssignedValue.(ast.BinaryExpr)
	if !ok || sub.Operator.Kind != lexer.MINUS {
		t.Fatalf("a := must be MINUS BinaryExpr, got %#v", last.AssignedValue)
	}

	if rhs, ok := sub.Right.(ast.SymbolExpr); !ok || rhs.Value != "d" {
		t.Fatalf("right operand should be symbol d, got %#v", sub.Right)
	}

	mul, ok := sub.Left.(ast.BinaryExpr)
	if !ok || mul.Operator.Kind != lexer.STAR {
		t.Fatalf("left operand should be STAR BinaryExpr, got %#v", sub.Left)
	}

	if rhs, ok := mul.Right.(ast.SymbolExpr); !ok || rhs.Value != "c" {
		t.Fatalf("right operand of * should be symbol c, got %#v", mul.Right)
	}

	add, ok := mul.Left.(ast.BinaryExpr)
	if !ok || add.Operator.Kind != lexer.PLUS {
		t.Fatalf("left operand of * should be PLUS BinaryExpr, got %#v", mul.Left)
	}

	if lNum, ok := add.Left.(ast.NumberExpr); !ok || lNum.Value != 2 {
		t.Fatalf("left operand of + should be number 2, got %#v", add.Left)
	}

	if rSym, ok := add.Right.(ast.SymbolExpr); !ok || rSym.Value != "b" {
		t.Fatalf("right operand of + should be symbol b, got %#v", add.Right)
	}
}
