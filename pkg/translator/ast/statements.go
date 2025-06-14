package ast

type BlockStmt struct {
	Body []Stmt
}

func (b BlockStmt) stmt() {}

// VarDeclarationStmt Var declare
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

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (n WhileStmt) stmt() {}

type InterruptionStmt struct {
	IrqNumber int
	Body      Stmt
}

type IntOnStmt struct {
}

func (n IntOnStmt) stmt() {}

type IntOffStmt struct {
}

func (n IntOffStmt) stmt() {}

func (n InterruptionStmt) stmt() {}

type PrintStmt struct {
	Argument Expr
}

func (n PrintStmt) stmt() {}

type ReadStmt struct {
	Argument Expr
}

func (n ReadStmt) stmt() {}

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
