package parser

import (
	"fmt"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
	"github.com/sanity-io/litter"
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
	startToken := p.advance().Kind
	symbolName := p.expectError(lexer.IDENTIFIER,
		fmt.Sprintf("Following %s expected variable name however instead recieved %s instead\n",
			lexer.TokenKindString(startToken), lexer.TokenKindString(p.currentTokenKind())))

	var assignmentValue ast.Expr
	if p.currentTokenKind() != lexer.SemiColon {
		p.expect(lexer.ASSIGNMENT)
		assignmentValue = parseExpr(p, assignment)
	}

	p.expect(lexer.SemiColon)

	return ast.VarDeclarationStmt{
		Identifier:    symbolName.Value,
		AssignedValue: assignmentValue,
	}
}
func parseLongDeclStmt(p *parser) ast.Stmt {
	startToken := p.advance().Kind
	symbolName := p.expectError(lexer.IDENTIFIER,
		fmt.Sprintf("Following %s expected variable name however instead recieved %s instead\n",
			lexer.TokenKindString(startToken), lexer.TokenKindString(p.currentTokenKind())))

	p.expect(lexer.ASSIGNMENT)
	assignmentValue := parseExpr(p, assignment)
	// var assignmentValue ast.Expr
	// if p.currentTokenKind() != lexer.SemiColon {
	// 	p.expect(lexer.ASSIGNMENT)
	// 	assignmentValue = parseExpr(p, assignment)
	// }

	p.expect(lexer.SemiColon)

	return ast.VarDeclarationStmt{
		Identifier:    symbolName.Value,
		AssignedValue: assignmentValue,
	}
}

func parsePrintStmt(p *parser) ast.Stmt {
	p.expect(lexer.PRINT)
	p.expect(lexer.OpenParen)

	arg := parseExpr(p, defaultBp)

	p.expect(lexer.CloseParen)
	p.expect(lexer.SemiColon)

	return ast.PrintStmt{Argument: arg}
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
	p.expect(lexer.WHILE)
	cond := parseExpr(p, assignment)
	body := parseBlockStmt(p)
	return ast.WhileStmt{Condition: cond, Body: body}
}

func parseInterStmt(p *parser) ast.Stmt {
	p.expect(lexer.INTER)
	n := parseExpr(p, assignment)
	var irqN int
	switch a := n.(type) {
	case ast.NumberExpr:
		irqN = int(a.Value)
	default:
		p.addError(fmt.Sprint("interruption number must be a number, got: ", litter.Sdump(n)))
	}

	b := parseBlockStmt(p)
	return ast.InterruptionStmt{IrqNumber: irqN, Body: b}
}
func parseIntOnStmt(p *parser) ast.Stmt {
	p.expect(lexer.IntOn)
	p.expect(lexer.SemiColon)
	return ast.IntOnStmt{}
}

func parseIntOffStmt(p *parser) ast.Stmt {
	p.expect(lexer.IntOff)
	p.expect(lexer.SemiColon)
	return ast.IntOffStmt{}
}
