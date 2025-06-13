package parser

import (
	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

type bindingPower int

const (
	defaultBp      bindingPower = iota // 0 - lowest possible precedence
	comma                              // ,
	assignment                         // =, +=, -=
	logical                            // AND, OR
	relational                         // <, <=, >, >=, ==, !=
	additive                           // +, -
	multiplicative                     // *, /, %
	unary                              // !, -, typeof (prefix)
	primary                            // Literals, Identifiers, Grouping ()
)

type stmtHandler func(p *parser) ast.Stmt
type nudHandler func(p *parser) ast.Expr

type ledHandler func(p *parser, left ast.Expr) ast.Expr

type stmtLookup map[lexer.TokenKind]stmtHandler
type nudLookup map[lexer.TokenKind]nudHandler
type ledLookup map[lexer.TokenKind]ledHandler
type bpLookup map[lexer.TokenKind]bindingPower // Stores LEFT BINDING POWER for LED tokens

var bpLu = bpLookup{}
var nudLu = nudLookup{}
var ledLu = ledLookup{}
var stmtLu = stmtLookup{}

// led registers an LED handler and its binding power.
func led(kind lexer.TokenKind, bp bindingPower, ledFn ledHandler) {
	bpLu[kind] = bp // Store the LEFT_BINDING_POWER for this operator
	ledLu[kind] = ledFn
}

// nud registers a NUD handler.
func nud(kind lexer.TokenKind, nudFn nudHandler) {
	nudLu[kind] = nudFn
}

func stmt(kind lexer.TokenKind, stmtFn stmtHandler) {
	// Statements usually have default/lowest precedence for expression parsing within them
	bpLu[kind] = defaultBp
	stmtLu[kind] = stmtFn
}

// createTokenLookups initializes the NUD, LED, and BP lookup tables.
func createTokenLookups() {
	led(lexer.ASSIGNMENT, assignment, parseAssignmentExpr)
	led(lexer.PlusEquals, assignment, parseAssignmentExpr)
	led(lexer.MinusEquals, assignment, parseAssignmentExpr)

	// Logical Operators (Left-Associative)
	led(lexer.AND, logical, parseBinaryExpr)
	led(lexer.OR, logical, parseBinaryExpr)

	// Relational Operators (Left-Associative)
	led(lexer.LESS, relational, parseBinaryExpr)
	led(lexer.LessEquals, relational, parseBinaryExpr)
	led(lexer.GREATER, relational, parseBinaryExpr)
	led(lexer.GreaterEquals, relational, parseBinaryExpr)
	led(lexer.EQUALS, relational, parseBinaryExpr)
	led(lexer.NotEquals, relational, parseBinaryExpr)

	// Additive Operators (Left-Associative)
	led(lexer.PLUS, additive, parseBinaryExpr)
	led(lexer.MINUS, additive, parseBinaryExpr) // For binary minus

	// Multiplicative Operators (Left-Associative)
	led(lexer.SLASH, multiplicative, parseBinaryExpr)
	led(lexer.STAR, multiplicative, parseBinaryExpr)
	led(lexer.PERCENT, multiplicative, parseBinaryExpr)

	// Literals & Symbols (NUDs - they start expressions)
	nud(lexer.NUMBER, parsePrimaryExpr)
	nud(lexer.STRING, parsePrimaryExpr)
	nud(lexer.IDENTIFIER, parsePrimaryExpr)
	nud(lexer.ADDSTR, parseAddStrExpr)
	nud(lexer.ADDL, parseAddLExpr)

	// Unary/Prefix Operators (NUDs)
	// The `unary` binding power is typically higher than binary operators,
	// ensuring `!x + y` parses as `(!x) + y`.
	nud(lexer.MINUS, parsePrefixExpr)
	nud(lexer.NOT, parsePrefixExpr)

	// Grouping Expression (NUD - starts a grouped expression)
	// Parentheses themselves define a grouping, their NUD handles parsing the inner expression.
	nud(lexer.OpenParen, parseGroupingExpr)

	// Built-in fnuctions
	nud(lexer.READCH, parseReadChEx)
	nud(lexer.READINT, parseReadIntEx)
	nud(lexer.LIST, parseListEx)

	// Statement handlers
	stmt(lexer.OpenCurly, parseBlockStmt)
	stmt(lexer.LET, parseVarDeclStmt)
	stmt(lexer.IF, parseIfStmt)
	stmt(lexer.WHILE, parseWhileStmt)
	stmt(lexer.PRINT, parsePrintStmt)
	stmt(lexer.INTER, parseInterStmt)
	stmt(lexer.IntOn, parseIntOnStmt)
	stmt(lexer.IntOff, parseIntOffStmt)
}
