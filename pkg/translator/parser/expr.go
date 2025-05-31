package parser

import (
	"fmt"
	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
	"strconv"
)

// bp is the right binding power limit.
func parse_expr(p *parser, rbp binding_power) ast.Expr {
	tokenKind := p.currentTokenKind()
	nud_fn, exists := nud_lu[tokenKind]

	if !exists {
		p.addError(fmt.Sprintf("NUD Handler expected for token %s\n", lexer.TokenKindString(tokenKind)))
		return nil
		// panic(p.errors[len(p.errors)-1])
	}

	// 1. Call the Null Denotation (NUD) to parse the left-hand side of the expression.
	// This handles literals, identifiers, prefix operators, and grouped expressions.
	left := nud_fn(p)

	// 2. Loop to parse Left Denotation (LED) operators.
	// This loop continues as long as the current token's LEFT_BINDING_POWER
	// is greater than the RIGHT_BINDING_POWER_LIMIT passed to this parse_expr call.
	// This is the core of precedence enforcement.
	for bp_lu[p.currentTokenKind()] > rbp {
		tokenKind = p.currentTokenKind()
		led_fn, exists := led_lu[tokenKind]

		if !exists {
			p.addError(fmt.Sprintf("LED Handler expected for token %s\n", lexer.TokenKindString(tokenKind)))
			panic(p.errors[len(p.errors)-1])
		}

		// Call the Left Denotation (LED), passing the already parsed 'left' expression.
		// The LED function is responsible for consuming the current operator token
		// and then parsing its right-hand side operand(s).
		left = led_fn(p, left) // Removed 'bp' as an argument to led_fn; it's handled internally by specific LED handlers.
	}

	return left
}

// parse_prefix_expr handles unary operators (e.g., -x, !y, typeof z)
func parse_prefix_expr(p *parser) ast.Expr {
	operatorToken := p.advance()
	// Parse the right-hand side operand with the 'unary' precedence.
	// This ensures that `!a + b` parses as `(!a) + b`, not `!(a + b)`.
	expr := parse_expr(p, unary)

	return ast.PrefixExpr{
		Operator: operatorToken,
		Right:    expr,
	}
}

// parse_assignment_expr handles assignment operators (e.g., x = y, x += y)
func parse_assignment_expr(p *parser, left ast.Expr) ast.Expr { // Removed bp, it's determined by the operator's own LBP
	operatorToken := p.advance() // Consume the assignment token (=, +=, etc.)

	// Assignment is right-associative (e.g., a = b = c parses as a = (b = c)).
	// To achieve this, the right-hand side expression is parsed with a binding power
	// *one less* than the assignment operator's own binding power.
	// This ensures that the next assignment operator on the right will bind to its
	// right operand before the current assignment operator finalizes.
	rhs := parse_expr(p, bp_lu[operatorToken.Kind]-1)

	return ast.AssignmentExpr{
		Assigne:       left,
		AssignedValue: rhs,
	}
}

// parse_range_expr handles range operators (e.g., 1..10)
func parse_range_expr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	operatorToken := p.advance() // Consume the '..' token

	// Range operator typically has very low precedence, but its right-hand side
	// should consume expressions according to the range operator's own precedence.
	// This makes expressions like `1 .. 5 + 2` parse as `1 .. (5 + 2)`.
	// For most left-associative operators, this means using its own binding power.
	right := parse_expr(p, bp_lu[operatorToken.Kind])

	return ast.RangeExpr{
		Lower: left,
		Upper: right,
	}
}

// parse_binary_expr handles standard binary operators (+, -, *, /, ==, <, etc.)
func parse_binary_expr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	operatorToken := p.advance() // Consume the operator token (+, *, ==, etc.)

	// For left-associative binary operators (like +, -, *, /, ==, <, etc.),
	// the right-hand side expression must be parsed with a right binding power
	// equal to the current operator's LEFT_BINDING_POWER.
	// This ensures that operators of equal or lower precedence on the right will *not*
	// be consumed by this `parse_expr` call, allowing the current operator to bind.
	// For example, in `2 + 3 * 4`:
	// 1. `parse_expr` parses `2`.
	// 2. Encounters `+`. Its LBP is `additive`. `additive > current_rbp`.
	// 3. Calls `parse_binary_expr` for `+`.
	// 4. `parse_binary_expr` consumes `+`.
	// 5. Calls `parse_expr` for the right side (`3 * 4`) with `additive` as `rbp`.
	//    a. `parse_expr` parses `3`.
	//    b. Encounters `*`. Its LBP is `multiplicative`. `multiplicative > additive`.
	//    c. Calls `parse_binary_expr` for `*`.
	//    d. `parse_binary_expr` consumes `*`.
	//    e. Calls `parse_expr` for `4` with `multiplicative` as `rbp`.
	//    f. `parse_expr` parses `4`.
	//    g. No more tokens with LBP > `multiplicative`. Returns `4`.
	//    h. `parse_binary_expr` for `*` gets `4`, constructs `3 * 4`.
	//    i. Returns `3 * 4`.
	// 6. `parse_binary_expr` for `+` gets `(3 * 4)`, constructs `2 + (3 * 4)`.
	// This ensures correct precedence.
	right := parse_expr(p, bp_lu[operatorToken.Kind])

	return ast.BinaryExpr{
		Left:     left,
		Operator: operatorToken,
		Right:    right,
	}
}

func parse_primary_expr(p *parser) ast.Expr {
	switch p.currentTokenKind() {
	case lexer.NUMBER:
		number, err := strconv.ParseUint(p.advance().Value, 10, 32)
		if err != nil {
			p.addError(fmt.Sprintf("Failed to parse number: %v", err))
			// Handle error gracefully or return a dummy node
			return ast.NumberExpr{Value: 0} // Or panic based on your error strategy
		}
		return ast.NumberExpr{
			Value: uint32(number),
		}
	case lexer.STRING:
		return ast.StringExpr{
			Value: p.advance().Value,
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

func parse_member_expr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	dotOrBracketToken := p.advance() // Consume DOT or OPEN_BRACKET

	isComputed := dotOrBracketToken.Kind == lexer.OPEN_BRACKET

	if isComputed {
		// For computed members (e.g., obj[prop]), the property itself is an expression.
		// It should be parsed with the lowest possible precedence (0 or defalt_bp)
		// to consume the entire expression inside the brackets.
		rhs := parse_expr(p, defalt_bp)
		p.expect(lexer.CLOSE_BRACKET)
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

func parse_array_literal_expr(p *parser) ast.Expr {
	p.expect(lexer.OPEN_BRACKET)
	arrayContents := make([]ast.Expr, 0)

	// Array elements are typically parsed with a precedence higher than comma,
	// but allowing for full expressions.
	// `logical` (or a similar low precedence) is often a good choice here.
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_BRACKET {
		arrayContents = append(arrayContents, parse_expr(p, assignment)) // Use assignment for array elements, common practice

		if p.currentTokenKind() == lexer.COMMA {
			p.advance() // Consume the comma
		} else if p.currentTokenKind() != lexer.CLOSE_BRACKET {
			// If not a comma and not closing bracket, something is wrong
			p.addError(fmt.Sprintf("Expected comma or closing bracket in array literal, but got %s\n", lexer.TokenKindString(p.currentTokenKind())))
			break // Try to recover by breaking the loop
		}
	}

	p.expect(lexer.CLOSE_BRACKET)

	return ast.ArrayLiteral{
		Contents: arrayContents,
	}
}

func parse_grouping_expr(p *parser) ast.Expr {
	p.expect(lexer.OPEN_PAREN)
	// Parse the expression inside the parentheses with the lowest possible precedence (0 or defalt_bp)
	// so it consumes the entire inner expression.
	expr := parse_expr(p, defalt_bp)
	p.expect(lexer.CLOSE_PAREN)
	return expr
}

func parse_call_expr(p *parser, left ast.Expr) ast.Expr { // Removed bp
	p.advance() // Consume the OPEN_PAREN token
	arguments := make([]ast.Expr, 0)

	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		// Arguments are typically parsed with the lowest possible precedence (e.g., assignment or logical)
		// to allow full expressions within the arguments.
		arguments = append(arguments, parse_expr(p, assignment))

		if p.currentTokenKind() == lexer.COMMA {
			p.advance() // Consume the comma
		} else if p.currentTokenKind() != lexer.CLOSE_PAREN {
			p.addError(fmt.Sprintf("Expected comma or closing parenthesis in call arguments, but got %s\n", lexer.TokenKindString(p.currentTokenKind())))
			break // Try to recover
		}
	}

	p.expect(lexer.CLOSE_PAREN)
	return ast.CallExpr{
		Method:    left,
		Arguments: arguments,
	}
}

func parse_fn_expr(p *parser) ast.Expr {
	p.expect(lexer.FN)
	functionParams, returnType, functionBody := parse_fn_params_and_body(p)

	return ast.FunctionExpr{
		Parameters: functionParams,
		ReturnType: returnType,
		Body:       functionBody,
	}
}
