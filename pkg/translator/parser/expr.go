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

func parseReadExpr(p *parser) ast.Expr {
	p.expect(lexer.READ)
	p.expect(lexer.OpenParen)
	p.expect(lexer.CloseParen)
	return ast.ReadExpr{} // Возвращаем новый AST-узел ReadExpr
}

// parseRangeExpr handles range operators (e.g., 1..10)
// TODO: delete
func parseRangeExpr(p *parser, left ast.Expr) ast.Expr {
	operatorToken := p.advance()
	right := parseExpr(p, bpLu[operatorToken.Kind])

	return ast.RangeExpr{
		Lower: left,
		Upper: right,
	}
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
		return ast.SymbolExpr{
			Value: p.advance().Value,
		}
	default:
		p.addError(fmt.Sprintf("Cannot create primary_expr from %s\n", lexer.TokenKindString(p.currentTokenKind())))
		// Important: In a real parser, you need robust error recovery here,
		// otherwise, this panic will stop everything.
		panic(p.errors[len(p.errors)-1])
	}
}

// TODO: delete
func parseMemberExpr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	dotOrBracketToken := p.advance() // Consume DOT or OPEN_BRACKET

	isComputed := dotOrBracketToken.Kind == lexer.OpenBracket

	if isComputed {
		// For computed members (e.g., obj[prop]), the property itself is an expression.
		// It should be parsed with the lowest possible precedence (0 or default_bp)
		// to consume the entire expression inside the brackets.
		rhs := parseExpr(p, defaultBp)
		p.expect(lexer.CloseBracket)
		return ast.ComputedExpr{
			Member:   left,
			Property: rhs,
		}
	}

	// For direct members (e.g., obj.property)
	return ast.MemberExpr{
		Member:   left,
		Property: p.expect(lexer.IDENTIFIER).Value,
	}
}

// TODO: delete
func parseArrayLiteralExpr(p *parser) ast.Expr {
	p.expect(lexer.OpenBracket)
	arrayContents := make([]ast.Expr, 0)

	// Array elements are typically parsed with a precedence higher than comma,
	// but allowing for full expressions.
	// `logical` (or a similar low precedence) is often a good choice here.
	for p.hasTokens() && p.currentTokenKind() != lexer.CloseBracket {
		arrayContents = append(arrayContents, parseExpr(p, assignment)) // Use assignment for array elements, common practice

		if p.currentTokenKind() == lexer.COMMA {
			p.advance() // Consume the comma
		} else if p.currentTokenKind() != lexer.CloseBracket {
			// If not a comma and not closing bracket, something is wrong
			p.addError(fmt.Sprintf("Expected comma or closing bracket in array literal, but got %s\n", lexer.TokenKindString(p.currentTokenKind())))
			break // Try to recover by breaking the loop
		}
	}

	p.expect(lexer.CloseBracket)

	return ast.ArrayLiteral{
		Contents: arrayContents,
	}
}

func parseGroupingExpr(p *parser) ast.Expr {
	p.expect(lexer.OpenParen)
	// Parse the expression inside the parentheses with the lowest possible precedence (0 or default_bp)
	// so it consumes the entire inner expression.
	expr := parseExpr(p, defaultBp)
	p.expect(lexer.CloseParen)
	return expr
}

func parseCallExpr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	p.advance() // Consume the OPEN_PAREN token
	arguments := make([]ast.Expr, 0)

	for p.hasTokens() && p.currentTokenKind() != lexer.CloseParen {
		// Arguments are typically parsed with the lowest possible precedence (e.g., assignment or logical)
		// to allow full expressions within the arguments.
		arguments = append(arguments, parseExpr(p, assignment))

		if p.currentTokenKind() == lexer.COMMA {
			p.advance() // Consume the comma
		} else if p.currentTokenKind() != lexer.CloseParen {
			p.addError(fmt.Sprintf("Expected comma or closing parenthesis in call arguments, but got %s\n", lexer.TokenKindString(p.currentTokenKind())))
			break // Try to recover
		}
	}

	p.expect(lexer.CloseParen)
	return ast.CallExpr{
		Method:    left,
		Arguments: arguments,
	}
}

func parseFnExpr(p *parser) ast.Expr {
	p.expect(lexer.FN)
	functionParams, returnType, functionBody := parseFnParamsAndBody(p)

	return ast.FunctionExpr{
		Parameters: functionParams,
		ReturnType: returnType,
		Body:       functionBody,
	}
}
