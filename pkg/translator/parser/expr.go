package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

// bp is the right binding power limit.
func parseExpr(p *parser, rbp bindingPower) ast.Expr {
	tokenKind := p.currentTokenKind()
	nudFn, exists := nudLu[tokenKind]

	if !exists {
		p.addError(fmt.Sprintf("NUD Handler expected for token %s\n", lexer.TokenKindString(tokenKind)))
		return nil
		// panic(p.errors[len(p.errors)-1])
	}

	// 1. Call the Null Denotation (NUD) to parse the left-hand side of the expression.
	// This handles literals, identifiers, prefix operators, and grouped expressions.
	left := nudFn(p)

	// 2. Loop to parse Left Denotation (LED) operators.
	// This loop continues as long as the current token's LEFT_BINDING_POWER
	// is greater than the RIGHT_BINDING_POWER_LIMIT passed to this parse_expr call.
	// This is the core of precedence enforcement.
	for bpLu[p.currentTokenKind()] > rbp {
		tokenKind = p.currentTokenKind()
		ledFn, exists := ledLu[tokenKind]

		if !exists {
			p.addError(fmt.Sprintf("LED Handler expected for token %s\n", lexer.TokenKindString(tokenKind)))
			panic(p.errors[len(p.errors)-1])
		}

		// Call the Left Denotation (LED), passing the already parsed 'left' expression.
		// The LED function is responsible for consuming the current operator token
		// and then parsing its right-hand side operand(s).
		left = ledFn(p, left) // Removed 'bp' as an argument to led_fn; it's handled internally by specific LED handlers.
	}

	return left
}

// parsePrefixExpr handles unary operators (e.g., -x, !y, typeof z)
func parsePrefixExpr(p *parser) ast.Expr {
	operatorToken := p.advance()
	// Parse the right-hand side operand with the 'unary' precedence.
	// This ensures that `!a + b` parses as `(!a) + b`, not `!(a + b)`.
	expr := parseExpr(p, unary)

	return ast.PrefixExpr{
		Operator: operatorToken,
		Right:    expr,
	}
}

// parseAssignmentExpr handles assignment operators (e.g., x = y, x += y)
func parseAssignmentExpr(p *parser, left ast.Expr) ast.Expr { // Removed bp, it's determined by the operator's own LBP
	operatorToken := p.advance() // Consume the assignment token (=, +=, etc.)

	// Assignment is right-associative (e.g., a = b = c parses as a = (b = c)).
	// To achieve this, the right-hand side expression is parsed with a binding power
	// *one less* than the assignment operator's own binding power.
	// This ensures that the next assignment operator on the right will bind to its
	// right operand before the current assignment operator finalizes.
	rhs := parseExpr(p, bpLu[operatorToken.Kind]-1)

	return ast.AssignmentExpr{
		Assigne:       left,
		AssignedValue: rhs,
	}
}

func parsePrintExpr(p *parser) ast.Expr {
	p.expect(lexer.PRINT)
	p.expect(lexer.OpenParen)
	arg := parseExpr(p, defaultBp)
	p.expect(lexer.CloseParen)
	return ast.PrintExpr{Argument: arg}
}

func parseReadChEx(p *parser) ast.Expr {
	p.expect(lexer.READCH)
	p.expect(lexer.OpenParen)
	p.expect(lexer.CloseParen)
	return ast.ReadChExpr{}
}
func parseReadIntEx(p *parser) ast.Expr {
	p.expect(lexer.READINT)
	p.expect(lexer.OpenParen)
	p.expect(lexer.CloseParen)
	return ast.ReadIntExpr{}
}

func parseListEx(p *parser) ast.Expr {
	p.expect(lexer.LIST)
	p.expect(lexer.OpenParen)
	arg := parseExpr(p, primary)
	var n int
	switch a := arg.(type) {
	case ast.NumberExpr:
		n = int(a.Value)
	default:
		p.addError(fmt.Sprintf("list argument should be a number, got %v", a))
		return nil
	}
	p.expect(lexer.CloseParen)
	return ast.ListEx{Size: n}
}

// parseBinaryExpr handles standard binary operators (+, -, *, /, ==, <, etc.)
func parseBinaryExpr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	operatorToken := p.advance() // Consume the operator token (+, *, ==, etc.)

	right := parseExpr(p, bpLu[operatorToken.Kind])

	return ast.BinaryExpr{
		Left:     left,
		Operator: operatorToken,
		Right:    right,
	}
}

func parsePrimaryExpr(p *parser) ast.Expr {
	switch p.currentTokenKind() {
	case lexer.NUMBER:
		number, err := strconv.ParseUint(p.advance().Value, 10, 32)
		if err != nil {
			p.addError(fmt.Sprintf("Failed to parse number: %v", err))
			return ast.NumberExpr{Value: 0}
		}
		return ast.NumberExpr{
			Value: int32(number),
		}
	case lexer.STRING:
		s := strings.Trim(p.advance().Value, `"`)
		return ast.StringExpr{
			Value: s,
		}
	case lexer.IDENTIFIER:
		identTok := p.advance()
		sym := ast.SymbolExpr{Value: identTok.Value}

		/* ---------- arr[i] ---------- */
		if p.currentTokenKind() == lexer.OpenBracket {
			p.advance()                    // '['
			idx := parseExpr(p, defaultBp) // expr
			p.expect(lexer.CloseBracket)
			return ast.ArrayIndexEx{
				Target: sym,
				Index:  idx,
			}
		}

		// /* ---------- ident(...)  ---------- */
		// if p.currentTokenKind() == lexer.OpenParen {
		// 	p.advance() // '('
		// 	var args []ast.Expr
		// 	if p.currentTokenKind() != lexer.OpenParen {
		// 		args = append(args, p.parseExpr())
		// 		for p.currentTokenKind() == lexer.COMMA {
		// 			p.advance()
		// 			args = append(args, p.parseExpr())
		// 		}
		// 	}
		// 	p.expect(lexer.CloseParen)
		// 	return ast.FunctionExpr{
		// 		Name: identTok.Value,
		// 		Args: args,
		// 	}
		// }

		return sym
	default:
		p.addError(fmt.Sprintf("Cannot create primary_expr from %s\n", lexer.TokenKindString(p.currentTokenKind())))
		return ast.StringExpr{}
	}
}

// // TODO: delete
// func parseArrayLiteralExpr(p *parser) ast.Expr {
// 	p.expect(lexer.OpenBracket)
// 	arrayContents := make([]ast.Expr, 0)
//
// 	// Array elements are typically parsed with a precedence higher than comma,
// 	// but allowing for full expressions.
// 	// `logical` (or a similar low precedence) is often a good choice here.
// 	for p.hasTokens() && p.currentTokenKind() != lexer.CloseBracket {
// 		arrayContents = append(arrayContents, parseExpr(p, assignment)) // Use assignment for array elements, common practice
//
// 		if p.currentTokenKind() == lexer.COMMA {
// 			p.advance() // Consume the comma
// 		} else if p.currentTokenKind() != lexer.CloseBracket {
// 			// If not a comma and not closing bracket, something is wrong
// 			p.addError(fmt.Sprintf("Expected comma or closing bracket in array literal, but got %s\n", lexer.TokenKindString(p.currentTokenKind())))
// 			break // Try to recover by breaking the loop
// 		}
// 	}
//
// 	p.expect(lexer.CloseBracket)
//
// 	return ast.ArrayLiteral{
// 		Contents: arrayContents,
// 	}
// }

func parseGroupingExpr(p *parser) ast.Expr {
	p.expect(lexer.OpenParen)
	expr := parseExpr(p, defaultBp)
	p.expect(lexer.CloseParen)
	return expr
}

// func parseCallExpr(p *parser, left ast.Expr) ast.Expr { // Removed bp
// 	p.advance() // Consume the OPEN_PAREN token
// 	arguments := make([]ast.Expr, 0)
//
// 	for p.hasTokens() && p.currentTokenKind() != lexer.CloseParen {
// 		// Arguments are typically parsed with the lowest possible precedence (e.g., assignment or logical)
// 		// to allow full expressions within the arguments.
// 		arguments = append(arguments, parseExpr(p, assignment))
//
// 		if p.currentTokenKind() == lexer.COMMA {
// 			p.advance() // Consume the comma
// 		} else if p.currentTokenKind() != lexer.CloseParen {
// 			p.addError(fmt.Sprintf("Expected comma or closing parenthesis in call arguments, but got %s\n", lexer.TokenKindString(p.currentTokenKind())))
// 			break // Try to recover
// 		}
// 	}
//
// 	p.expect(lexer.CloseParen)
// 	return ast.CallExpr{
// 		Method:    left,
// 		Arguments: arguments,
// 	}
// }
