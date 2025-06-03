package parser

import (
	"fmt"
	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

func parseStmt(p *parser) ast.Stmt {
	stmtFn, exists := stmtLu[p.currentTokenKind()]

	if exists {
		return stmtFn(p)
	}

	return parseExpressionStmt(p)
}

func parseExpressionStmt(p *parser) ast.ExpressionStmt {
	expression := parseExpr(p, defaultBp)
	p.expect(lexer.SEMI_COLON)

	return ast.ExpressionStmt{
		Expression: expression,
	}
}

func parseBlockStmt(p *parser) ast.Stmt {
	p.expect(lexer.OPEN_CURLY)
	var body []ast.Stmt

	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_CURLY {
		body = append(body, parseStmt(p))
	}

	p.expect(lexer.CLOSE_CURLY)
	return ast.BlockStmt{
		Body: body,
	}
}

func parseVarDeclStmt(p *parser) ast.Stmt {
	var explicitType ast.Type
	startToken := p.advance().Kind
	// isConstant := startToken == lexer.CONST
	symbolName := p.expectError(lexer.IDENTIFIER,
		fmt.Sprintf("Following %s expected variable name however instead recieved %s instead\n",
			lexer.TokenKindString(startToken), lexer.TokenKindString(p.currentTokenKind())))

	if p.currentTokenKind() == lexer.COLON {
		p.expect(lexer.COLON)
		explicitType = parseType(p, defaultBp)
	}

	var assignmentValue ast.Expr
	if p.currentTokenKind() != lexer.SEMI_COLON {
		p.expect(lexer.ASSIGNMENT)
		assignmentValue = parseExpr(p, assignment)
	} else if explicitType == nil {
		// panic("Missing explicit type for variable declaration.")
		p.addError("Missing explicit type for variable declaration without initial assignment.")
	}

	p.expect(lexer.SEMI_COLON)

	return ast.VarDeclarationStmt{
		// Constant:      isConstant,
		Identifier:    symbolName.Value,
		AssignedValue: assignmentValue,
		// ExplicitType:  explicitType,
	}
}

func parseFnParamsAndBody(p *parser) ([]ast.Parameter, ast.Type, []ast.Stmt) {
	functionParams := make([]ast.Parameter, 0)

	p.expect(lexer.OPEN_PAREN)
	for p.hasTokens() && p.currentTokenKind() != lexer.CLOSE_PAREN {
		paramName := p.expect(lexer.IDENTIFIER).Value
		p.expect(lexer.COLON)
		paramType := parseType(p, defaultBp)

		functionParams = append(functionParams, ast.Parameter{
			Name: paramName,
			Type: paramType,
		})

		if !p.currentToken().IsOneOfMany(lexer.CLOSE_PAREN, lexer.EOF) {
			p.expect(lexer.COMMA)
		}
	}

	p.expect(lexer.CLOSE_PAREN)
	var returnType ast.Type

	if p.currentTokenKind() == lexer.COLON {
		p.advance()
		returnType = parseType(p, defaultBp)
	}

	functionBody := ast.ExpectStmt[ast.BlockStmt](parseBlockStmt(p)).Body

	return functionParams, returnType, functionBody
}

func parseFnDeclaration(p *parser) ast.Stmt {
	p.advance()
	functionName := p.expect(lexer.IDENTIFIER).Value
	functionParams, returnType, functionBody := parseFnParamsAndBody(p)

	return ast.FunctionDeclarationStmt{
		Parameters: functionParams,
		ReturnType: returnType,
		Body:       functionBody,
		Name:       functionName,
	}
}

func parsePrintStmt(p *parser) ast.Stmt {
	p.expect(lexer.PRINT)
	p.expect(lexer.OPEN_PAREN)

	arg := parseExpr(p, defaultBp)
	// fmt.Println(arg) // This is for debugging the parser, not for code generation

	p.expect(lexer.CLOSE_PAREN)
	p.expect(lexer.SEMI_COLON)

	return ast.PrintStmt{Argument: arg}
}

// parseReadStmt is removed or commented out because read() is now an expression.
// If you *do* need `read();` as a standalone statement (which typically returns a value that is then discarded),
// you would create a ReadStmt node that contains a ReadExpr.
// func parseReadStmt(p *parser) ast.Stmt {
// 	p.expect(lexer.READ)
// 	p.expect(lexer.OPEN_PAREN)
// 	p.expect(lexer.CLOSE_PAREN)
// 	p.expect(lexer.SEMI_COLON)
// 	return ast.ReadStmt{} // Assuming you have a ReadStmt in ast/statements.go
// }

func parseReadStmt(p *parser) ast.Stmt {
	panic("impl me")
}

func parseIfStmt(p *parser) ast.Stmt {
	p.advance()
	condition := parseExpr(p, assignment)
	consequent := parseBlockStmt(p)

	var alternate ast.Stmt
	if p.currentTokenKind() == lexer.ELSE {
		p.advance()

		if p.currentTokenKind() == lexer.IF {
			alternate = parseIfStmt(p)
		} else {
			alternate = parseBlockStmt(p)
		}
	}

	return ast.IfStmt{
		Condition:  condition,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func parseReturnStmt(p *parser) ast.Stmt {
	p.expect(lexer.RETURN)

	var expr ast.Expr
	if p.currentTokenKind() != lexer.SEMI_COLON {
		// Return expression can be any valid expression up to `defalt_bp` (lowest)
		expr = parseExpr(p, defaultBp)
	}

	p.expect(lexer.SEMI_COLON)

	return ast.ReturnStmt{
		Expr: expr,
	}
}
