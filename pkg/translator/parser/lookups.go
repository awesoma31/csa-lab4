package parser

import (
	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

type binding_power int

// const (
//
//	defalt_bp binding_power = iota
//	lowest                  // 0
//	comma
//	assignment // 1  (=)
//	logical    // 2  (||, &&)
//	range_bp   // 3  (..)
//	relational
//	addsub // 4  (+, -)
//	additive
//	muldiv // 5  (*, /)
//	multiplicative
//	unary       // 6  (-x, !x)
//	call_member // 7  (f(), obj[x])
//	call
//	member
//	primary
//
// )
const (
	defalt_bp      binding_power = iota // 0 - lowest possible precedence
	comma                               // ,
	assignment                          // =, +=, -=
	logical                             // AND, OR
	relational                          // <, <=, >, >=, ==, !=
	additive                            // +, -
	multiplicative                      // *, /, %
	unary                               // !, -, typeof (prefix)
	call                                // function()
	member                              // . (dot), [] (computed member)
	primary                             // Literals, Identifiers, Grouping ()
)

type stmt_handler func(p *parser) ast.Stmt
type nud_handler func(p *parser) ast.Expr

// type led_handler func(p *parser, left ast.Expr, bp binding_power) ast.Expr
// LED handlers now only take the parser and the left expression.
// The binding power is determined by the operator itself and used in the parse_expr call for the right side.
type led_handler func(p *parser, left ast.Expr) ast.Expr

type stmt_lookup map[lexer.TokenKind]stmt_handler
type nud_lookup map[lexer.TokenKind]nud_handler
type led_lookup map[lexer.TokenKind]led_handler
type bp_lookup map[lexer.TokenKind]binding_power // Stores LEFT BINDING POWER for LED tokens

var bp_lu = bp_lookup{}
var nud_lu = nud_lookup{}
var led_lu = led_lookup{}
var stmt_lu = stmt_lookup{}

// led registers an LED handler and its binding power.
func led(kind lexer.TokenKind, bp binding_power, led_fn led_handler) {
	bp_lu[kind] = bp // Store the LEFT_BINDING_POWER for this operator
	led_lu[kind] = led_fn
}

// nud registers a NUD handler.
// NUDs don't have a left binding power, so we remove the bp_lu assignment here.
func nud(kind lexer.TokenKind, nud_fn nud_handler) { // Removed bp parameter
	nud_lu[kind] = nud_fn
}

func stmt(kind lexer.TokenKind, stmt_fn stmt_handler) {
	// Statements usually have default/lowest precedence for expression parsing within them
	bp_lu[kind] = defalt_bp
	stmt_lu[kind] = stmt_fn
}

// createTokenLookups initializes the NUD, LED, and BP lookup tables.
func createTokenLookups() {
	// Assignment Operators (Right-Associative: A = B = C  => A = (B = C))
	// They have the same binding power, but their parse_assignment_expr ensures right-associativity.
	led(lexer.ASSIGNMENT, assignment, parse_assignment_expr)
	led(lexer.PLUS_EQUALS, assignment, parse_assignment_expr)
	led(lexer.MINUS_EQUALS, assignment, parse_assignment_expr)

	// Logical Operators (Left-Associative)
	led(lexer.AND, logical, parse_binary_expr)
	led(lexer.OR, logical, parse_binary_expr)
	// Range operator (often low precedence, can be left or non-associative depending on language)
	led(lexer.DOT_DOT, logical, parse_range_expr) // Placing it at 'logical' level for now. Adjust if needed.

	// Relational Operators (Left-Associative)
	led(lexer.LESS, relational, parse_binary_expr)
	led(lexer.LESS_EQUALS, relational, parse_binary_expr)
	led(lexer.GREATER, relational, parse_binary_expr)
	led(lexer.GREATER_EQUALS, relational, parse_binary_expr)
	led(lexer.EQUALS, relational, parse_binary_expr)
	led(lexer.NOT_EQUALS, relational, parse_binary_expr)

	// Additive Operators (Left-Associative)
	led(lexer.PLUS, additive, parse_binary_expr)
	led(lexer.MINUS, additive, parse_binary_expr) // For binary minus

	// Multiplicative Operators (Left-Associative)
	led(lexer.SLASH, multiplicative, parse_binary_expr)
	led(lexer.ASTERISK, multiplicative, parse_binary_expr)
	led(lexer.PERCENT, multiplicative, parse_binary_expr)

	// Literals & Symbols (NUDs - they start expressions)
	nud(lexer.NUMBER, parse_primary_expr)
	nud(lexer.STRING, parse_primary_expr)
	nud(lexer.IDENTIFIER, parse_primary_expr)
	// Array literals also start expressions, like `[1, 2, 3]`
	nud(lexer.OPEN_BRACKET, parse_array_literal_expr)

	// Unary/Prefix Operators (NUDs)
	// The `unary` binding power is typically higher than binary operators,
	// ensuring `!x + y` parses as `(!x) + y`.
	nud(lexer.TYPEOF, parse_prefix_expr)
	nud(lexer.MINUS, parse_prefix_expr) // For unary minus (e.g., -5)
	nud(lexer.NOT, parse_prefix_expr)

	// Member / Computed Access / Call Operators (LEDs - they bind to a left expression)
	// Call and Member access have very high precedence.
	led(lexer.DOT, member, parse_member_expr)
	led(lexer.OPEN_BRACKET, member, parse_member_expr) // For computed properties like `obj[prop]`
	led(lexer.OPEN_PAREN, call, parse_call_expr)       // For function calls like `func()`

	// Grouping Expression (NUD - starts a grouped expression)
	// Parentheses themselves define a grouping, their NUD handles parsing the inner expression.
	nud(lexer.OPEN_PAREN, parse_grouping_expr)

	// Function Expression (NUD - `fn () {}`)
	nud(lexer.FN, parse_fn_expr)

	// New Expression (NUD - `new Class()`)
	nud(lexer.NEW, func(p *parser) ast.Expr {
		p.advance() // Consume 'new'
		// `new Class()` is typically followed by a CallExpr
		classInstantiation := parse_expr(p, call) // Parse with `call` precedence to ensure it's a call

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

	// Statement handlers
	stmt(lexer.RETURN, parse_return_stmt)
	stmt(lexer.OPEN_CURLY, parse_block_stmt)
	stmt(lexer.LET, parse_var_decl_stmt)
	stmt(lexer.CONST, parse_var_decl_stmt)
	stmt(lexer.FN, parse_fn_declaration) // Function Declaration (Statement)
	stmt(lexer.IF, parse_if_stmt)
	stmt(lexer.IMPORT, parse_import_stmt)
	stmt(lexer.FOREACH, parse_foreach_stmt)
	stmt(lexer.CLASS, parse_class_declaration_stmt)
}

func parse_return_stmt(p *parser) ast.Stmt {
	p.expect(lexer.RETURN)

	var expr ast.Expr
	if p.currentTokenKind() != lexer.SEMI_COLON {
		// Return expression can be any valid expression up to `defalt_bp` (lowest)
		expr = parse_expr(p, defalt_bp)
	}

	p.expect(lexer.SEMI_COLON)

	return ast.ReturnStmt{
		Expr: expr,
	}
}
