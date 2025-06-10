package ast

import (
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

// --------------------
// Literal Expressions
// --------------------

type NumberExpr struct {
	Value int32
}

func (n NumberExpr) expr() {}

type StringExpr struct {
	Value string
}

func (n StringExpr) expr() {}

type SymbolExpr struct {
	Value string
}

func (n SymbolExpr) expr() {}

type ReadChExpr struct{}

func (n ReadChExpr) expr() {}

type ReadIntExpr struct{}

func (n ReadIntExpr) expr() {}

type ListEx struct {
	Size int
}

func (n ListEx) expr() {}

type ArrayIndexEx struct {
	Target Expr
	Index  Expr
}

func (n ArrayIndexEx) expr() {}

// TODO: should be only standalone stmt so remove
type PrintExpr struct {
	Argument Expr
}

func (n PrintExpr) expr() {}

// --------------------
// Complex Expressions
// --------------------

type BinaryExpr struct {
	Left     Expr
	Operator lexer.Token
	Right    Expr
}

func (n BinaryExpr) expr() {}

type AssignmentExpr struct {
	Assigne       Expr
	AssignedValue Expr
}

func (n AssignmentExpr) expr() {}

type PrefixExpr struct {
	Operator lexer.Token
	Right    Expr
}

func (n PrefixExpr) expr() {}

type MemberExpr struct {
	Member   Expr
	Property string
}

func (n MemberExpr) expr() {}

type CallExpr struct {
	Method    Expr
	Arguments []Expr
}

func (n CallExpr) expr() {}

type ComputedExpr struct {
	Member   Expr
	Property Expr
}

func (n ComputedExpr) expr() {}

type RangeExpr struct {
	Lower Expr
	Upper Expr
}

func (n RangeExpr) expr() {}

type FunctionExpr struct {
	Parameters []Parameter
	Body       []Stmt
	ReturnType Type
}

func (n FunctionExpr) expr() {}

type ArrayLiteral struct {
	Contents []Expr
}

func (n ArrayLiteral) expr() {}

type NewExpr struct {
	Instantiation CallExpr
}

func (n NewExpr) expr() {}
