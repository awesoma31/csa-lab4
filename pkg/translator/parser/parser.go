package parser

import (
	"fmt"
	"log"

	"github.com/awesoma31/csa-lab4/pkg/translator/ast"
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
	"github.com/sanity-io/litter"
)

type parser struct {
	tokens []lexer.Token
	pos    int

	errors []string
}

func createParser(tokens []lexer.Token) *parser {
	if len(bpLu) == 0 {
		createTokenLookups()
		createTypeTokenLookups()
	}

	p := &parser{
		tokens: tokens,
		pos:    0,
		errors: make([]string, 0), // Initialize errors slice
	}

	return p
}

func (p *parser) addError(errs ...string) {
	p.errors = append(p.errors, errs...)
}

func (p *parser) Errors() []string {
	return p.errors
}

func Parse(source string) (ast.BlockStmt, []string) {
	tokens := lexer.Tokenize(source)
	p := createParser(tokens)
	body := make([]ast.Stmt, 0)

	for p.hasTokens() {
		// body = append(body, parse_stmt(p))
		stmt := parseStmt(p)
		if stmt != nil { // Only append if a statement was successfully parsed
			body = append(body, stmt)
		} else {
			// If parse_stmt returns nil (due to error recovery), advance to avoid infinite loop
			p.advance()
		}
	}

	if len(p.errors) > 0 {
		log.Fatal(litter.Sdump(p.errors))
	}
	return ast.BlockStmt{
		Body: body,
	}, p.Errors()
}

// currentTokenKind returns the kind of the current token, or EOF if past end.
func (p *parser) currentTokenKind() lexer.TokenKind {
	if p.pos >= len(p.tokens) {
		return lexer.EOF
	}
	return p.tokens[p.pos].Kind
}

// currentToken returns the current token, or an EOF token if past end.
func (p *parser) currentToken() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF, Value: "EOF"}
	}
	return p.tokens[p.pos]
}

// advance moves the parser to the next token and returns the consumed token.
func (p *parser) advance() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Kind: lexer.EOF, Value: "EOF"}
	}
	token := p.tokens[p.pos]
	p.pos++
	return token
}

// expect checks if the current token is of the expected kind, consumes it, and returns it.
// If not, it adds an error and attempts to recover (or panics).
func (p *parser) expect(kind lexer.TokenKind) lexer.Token {
	if p.currentTokenKind() == kind {
		return p.advance()
	}
	p.addError(fmt.Sprintf("Expected token %s but got %s",
		lexer.TokenKindString(kind), lexer.TokenKindString(p.currentTokenKind())))
	// In a real parser, you might insert a dummy node or skip tokens here for recovery.
	// For now, we'll return a dummy token.
	return lexer.Token{Kind: lexer.UNKNOWN, Value: "ERROR"}
}

// expectError is a specialized expect that allows custom error messages.
func (p *parser) expectError(kind lexer.TokenKind, errMsg string) lexer.Token {
	if p.currentTokenKind() == kind {
		return p.advance()
	}
	p.addError(errMsg)
	return lexer.Token{Kind: lexer.UNKNOWN, Value: "ERROR"}
}

// hasTokens checks if there are more tokens to consume (i.e., not at EOF).
func (p *parser) hasTokens() bool {
	return p.pos < len(p.tokens) && p.currentTokenKind() != lexer.EOF
}
