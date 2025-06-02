package ast

import "github.com/awesoma31/csa-lab4/pkg/translator/helpers"

type Stmt interface {
	stmt()
}

type Expr interface {
	expr()
}

type Type interface {
	_type()
	String() string
}

func ExpectExpr[T Expr](expr Expr) T {
	return helpers.ExpectType[T](expr)
}

func ExpectStmt[T Stmt](expr Stmt) T {
	return helpers.ExpectType[T](expr)
}
