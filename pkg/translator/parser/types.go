package parser

import (
	"fmt"
	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

type typeNudHandler func(p *parser) ast.Type
type typeLedHandler func(p *parser, left ast.Type) ast.Type

type typeNudLookup map[lexer.TokenKind]typeNudHandler
type typeLedLookup map[lexer.TokenKind]typeLedHandler
type typeBpLookup map[lexer.TokenKind]bindingPower

var typeBpLu = typeBpLookup{}
var typeNudLu = typeNudLookup{}
var typeLedLu = typeLedLookup{}

// func typeLed(kind lexer.TokenKind, bp bindingPower, ledFn typeLedHandler) {
// 	typeBpLu[kind] = bp
// 	typeLedLu[kind] = ledFn
// }

func typeNud(kind lexer.TokenKind, nudFn typeNudHandler) {
	typeNudLu[kind] = nudFn
}

func createTypeTokenLookups() {

	typeNud(lexer.IDENTIFIER, func(p *parser) ast.Type {
		return ast.SymbolType{
			Value: p.advance().Value,
		}
	})

	typeNud(lexer.OpenBracket, func(p *parser) ast.Type {
		p.advance()
		p.expect(lexer.CloseBracket)
		insideType := parseType(p, defaultBp)

		return ast.ListType{
			Underlying: insideType,
		}
	})
}

// parseType is the Pratt parser for type expressions.
// TODO: delete
func parseType(p *parser, rbp bindingPower) ast.Type {
	tokenKind := p.currentTokenKind()
	nudFn, exists := typeNudLu[tokenKind]

	if !exists {
		p.addError(fmt.Sprintf("type: NUD Handler expected for token %s\n", lexer.TokenKindString(tokenKind)))
		panic(p.errors[len(p.errors)-1])
	}

	left := nudFn(p)

	for typeBpLu[p.currentTokenKind()] > rbp {
		tokenKind = p.currentTokenKind()
		ledFn, exists := typeLedLu[tokenKind]

		if !exists {
			p.addError(fmt.Sprintf("type: LED Handler expected for token %s\n", lexer.TokenKindString(tokenKind)))
			panic(p.errors[len(p.errors)-1])
		}

		left = ledFn(p, left) // Removed bp argument
	}

	return left
}
