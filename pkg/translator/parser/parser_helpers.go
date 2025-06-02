package parser

import (
	"github.com/awesoma31/csa-lab4/pkg/translator/lexer"
)

func (p *parser) nextToken() lexer.Token {
	return p.tokens[p.pos+1]
}

func (p *parser) previousToken() lexer.Token {
	return p.tokens[p.pos-1]
}
