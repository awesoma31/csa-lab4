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
	call                               // function()
	member                             // . (dot), [] (computed member)
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
// NUDs don't have a left binding power, so we remove the bpLu assignment here.
func nud(kind lexer.TokenKind, nudFn nudHandler) { // Removed bp parameter
	nudLu[kind] = nudFn
}

func stmt(kind lexer.TokenKind, stmtFn stmtHandler) {
	// Statements usually have default/lowest precedence for expression parsing within them
	bpLu[kind] = defaultBp
	stmtLu[kind] = stmtFn
}

// createTokenLookups initializes the NUD, LED, and BP lookup tables.
func createTokenLookups() {
	// Assignment Operators (Right-Associative: A = B = C  => A = (B = C))
	// They have the same binding power, but their parse_assignment_expr ensures right-associativity.
	led(lexer.ASSIGNMENT, assignment, parseAssignmentExpr)
	led(lexer.PLUS_EQUALS, assignment, parseAssignmentExpr)
	led(lexer.MINUS_EQUALS, assignment, parseAssignmentExpr)

	// Logical Operators (Left-Associative)
	led(lexer.AND, logical, parseBinaryExpr)
	led(lexer.OR, logical, parseBinaryExpr)
	// Range operator (often low precedence, can be left or non-associative depending on language)
	led(lexer.DOT_DOT, logical, parseRangeExpr) // Placing it at 'logical' level for now. Adjust if needed.

	// Relational Operators (Left-Associative)
	led(lexer.LESS, relational, parseBinaryExpr)
	led(lexer.LESS_EQUALS, relational, parseBinaryExpr)
	led(lexer.GREATER, relational, parseBinaryExpr)
	led(lexer.GREATER_EQUALS, relational, parseBinaryExpr)
	led(lexer.EQUALS, relational, parseBinaryExpr)
	led(lexer.NOT_EQUALS, relational, parseBinaryExpr)

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
	// Array literals also start expressions, like `[1, 2, 3]`
	nud(lexer.OPEN_BRACKET, parseArrayLiteralExpr)

	// Unary/Prefix Operators (NUDs)
	// The `unary` binding power is typically higher than binary operators,
	// ensuring `!x + y` parses as `(!x) + y`.
	nud(lexer.TYPEOF, parsePrefixExpr)
	nud(lexer.MINUS, parsePrefixExpr) // For unary minus (e.g., -5)
	nud(lexer.NOT, parsePrefixExpr)

	// Member / Computed Access / Call Operators (LEDs - they bind to a left expression)
	// Call and Member access have very high precedence.
	led(lexer.DOT, member, parseMemberExpr)
	led(lexer.OPEN_BRACKET, member, parseMemberExpr) // For computed properties like `obj[prop]`
	led(lexer.OPEN_PAREN, call, parseCallExpr)       // For function calls like `func()`

	// Grouping Expression (NUD - starts a grouped expression)
	// Parentheses themselves define a grouping, their NUD handles parsing the inner expression.
	nud(lexer.OPEN_PAREN, parseGroupingExpr)

	// Function Expression (NUD - `fn () {}`)
	nud(lexer.FN, parseFnExpr)

	// New Expression (NUD - `new Class()`)
	nud(lexer.NEW, func(p *parser) ast.Expr {
		p.advance() // Consume 'new'
		// `new Class()` is typically followed by a CallExpr
		classInstantiation := parseExpr(p, call) // Parse with `call` precedence to ensure it's a call

		// Type assertion for robustness
		callExpr, ok := classInstantiation.(ast.CallExpr)
		if !ok {
			p.addError("Expected a call expression after 'new' keyword")
			// Return a dummy node or panic based on error strategy
			return ast.NewExpr{Instantiation: ast.CallExpr{}}
		}

		return ast.NewExpr{
			Instantiation: callExpr,
		}
	})

	// Built-in fnuctions
	nud(lexer.PRINT, parsePrintExpr)
	nud(lexer.READ, parseReadExpr)
	// stmt(lexer.PRINT, parsePrintStmt)
	// stmt(lexer.READ, parseReadStmt)

	// Statement handlers
	stmt(lexer.RETURN, parseReturnStmt)
	stmt(lexer.OPEN_CURLY, parseBlockStmt)
	stmt(lexer.LET, parseVarDeclStmt)
	// stmt(lexer.CONST, parseVarDeclStmt)
	stmt(lexer.FN, parseFnDeclaration) // Function Declaration (Statement)
	stmt(lexer.IF, parseIfStmt)
	// stmt(lexer.IMPORT, parseImportStmt)
	// stmt(lexer.FOREACH, parseForeachStmt)
	// stmt(lexer.CLASS, parseClassDeclarationStmt)
	stmt(lexer.PRINT, parsePrintStmt) // This would be for `print("hello");` as a statement
	// stmt(lexer.READ, parseReadStmt) // if read is standalone
}
