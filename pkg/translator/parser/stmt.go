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
	p.expect(lexer.SemiColon)

	return ast.ExpressionStmt{
		Expression: expression,
	}
}

func parseBlockStmt(p *parser) ast.Stmt {
	p.expect(lexer.OpenCurly)
	var body []ast.Stmt

	for p.hasTokens() && p.currentTokenKind() != lexer.CloseCurly {
		body = append(body, parseStmt(p))
	}

	p.expect(lexer.CloseCurly)
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
	if p.currentTokenKind() != lexer.SemiColon {
		p.expect(lexer.ASSIGNMENT)
		assignmentValue = parseExpr(p, assignment)
	} else if explicitType == nil {
		// panic("Missing explicit type for variable declaration.")
		p.addError("Missing explicit type for variable declaration without initial assignment.")
	}

	p.expect(lexer.SemiColon)

	return ast.VarDeclarationStmt{
		// Constant:      isConstant,
		Identifier:    symbolName.Value,
		AssignedValue: assignmentValue,
		// ExplicitType:  explicitType,
	}
}

func parseFnParamsAndBody(p *parser) ([]ast.Parameter, ast.Type, []ast.Stmt) {
	functionParams := make([]ast.Parameter, 0)

	p.expect(lexer.OpenParen)
	for p.hasTokens() && p.currentTokenKind() != lexer.CloseParen {
		paramName := p.expect(lexer.IDENTIFIER).Value
		p.expect(lexer.COLON)
		paramType := parseType(p, defaultBp)

		functionParams = append(functionParams, ast.Parameter{
			Name: paramName,
			Type: paramType,
		})

		if !p.currentToken().IsOneOfMany(lexer.CloseParen, lexer.EOF) {
			p.expect(lexer.COMMA)
		}
	}

	p.expect(lexer.CloseParen)
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
	p.expect(lexer.OpenParen)

	arg := parseExpr(p, defaultBp)
	// fmt.Println(arg) // This is for debugging the parser, not for code generation

	p.expect(lexer.CloseParen)
	p.expect(lexer.SemiColon)

	return ast.PrintStmt{Argument: arg}
}

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

func parseWhileStmt(p *parser) ast.Stmt {
	p.advance()
	cond := parseExpr(p, assignment)
	body := parseBlockStmt(p)
	return ast.WhileStmt{Condition: cond, Body: body}
}

func parseReturnStmt(p *parser) ast.Stmt {
	p.expect(lexer.RETURN)

	var expr ast.Expr
	if p.currentTokenKind() != lexer.SemiColon {
		// Return expression can be any valid expression up to `defalt_bp` (lowest)
		expr = parseExpr(p, defaultBp)
	}

	p.expect(lexer.SemiColon)

	return ast.ReturnStmt{
		Expr: expr,
	}
}
