package ast

type BlockStmt struct {
	Body []Stmt
}

func (b BlockStmt) stmt() {}

// Var declare
type VarDeclarationStmt struct {
	Identifier    string
	AssignedValue Expr
	// ExplicitType  Type // Используем новый интерфейс Type
}

func (n VarDeclarationStmt) stmt() {}

type ExpressionStmt struct {
	Expression Expr
}

func (n ExpressionStmt) stmt() {}

type Parameter struct {
	Name string
	Type Type
}

// Function
type FunctionDeclarationStmt struct {
	Name       string
	Parameters []Parameter
	Body       []Stmt
	ReturnType Type
}

func (n FunctionDeclarationStmt) stmt() {}

type ReturnStmt struct {
	Expr Expr
}

func (n ReturnStmt) stmt() {}

type IfStmt struct {
	Condition  Expr
	Consequent Stmt
	Alternate  Stmt
}

func (n IfStmt) stmt() {}

type ImportStmt struct {
	Name string
	From string
}

func (n ImportStmt) stmt() {}

type ForeachStmt struct {
	Value    string
	Index    bool
	Iterable Expr
	Body     []Stmt
}

func (n ForeachStmt) stmt() {}

type ClassDeclarationStmt struct {
	Name string
	Body []Stmt
}

func (n ClassDeclarationStmt) stmt() {}
