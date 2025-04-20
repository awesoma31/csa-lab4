package main

import (
	"fmt"
	"strconv"
	"unicode"
)

type MathTokenType int

const (
	ILLEGAL MathTokenType = iota
	EOF
	NUMBER
	PLUS     // +
	MINUS    // -
	ASTERISK // *
	SLASH    // /
	CARET    // ^
	LPAREN   // (
	RPAREN   // )
)

type MathToken struct {
	Type  MathTokenType
	Value string
}

type MathLexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
}

func NewLexer(input string) *MathLexer {
	l := &MathLexer{input: input}
	l.readChar()
	return l
}

func (l *MathLexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readPosition])
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *MathLexer) NextToken() MathToken {
	// пропустить пробелы
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}

	var tok MathToken
	switch l.ch {
	case '+':
		tok = MathToken{Type: PLUS, Value: "+"}
	case '-':
		tok = MathToken{Type: MINUS, Value: "-"}
	case '*':
		tok = MathToken{Type: ASTERISK, Value: "*"}
	case '/':
		tok = MathToken{Type: SLASH, Value: "/"}
	case '^':
		tok = MathToken{Type: CARET, Value: "^"}
	case '(':
		tok = MathToken{Type: LPAREN, Value: "("}
	case ')':
		tok = MathToken{Type: RPAREN, Value: ")"}
	case 0:
		tok = MathToken{Type: EOF, Value: ""}
	default:
		if unicode.IsDigit(l.ch) {
			start := l.position
			for unicode.IsDigit(l.ch) || l.ch == '.' {
				l.readChar()
			}
			tok.Value = l.input[start:l.position]
			tok.Type = NUMBER
			return tok
		}
		tok = MathToken{Type: ILLEGAL, Value: string(l.ch)}
	}
	l.readChar()
	return tok
}

// --- AST узлы ---

type Expr interface {
	String(indent string) string
}

type NumberLiteral struct {
	Value float64
}

func (n *NumberLiteral) String(indent string) string {
	return fmt.Sprintf("%sNumber(%v)\n", indent, n.Value)
}

type UnaryExpr struct {
	Op   string
	Expr Expr
}

func (u *UnaryExpr) String(indent string) string {
	s := fmt.Sprintf("%sUnaryOp(%s)\n", indent, u.Op)
	s += u.Expr.String(indent + "  ")
	return s
}

type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

func (b *BinaryExpr) String(indent string) string {
	s := fmt.Sprintf("%sBinaryOp(%s)\n", indent, b.Op)
	s += b.Left.String(indent + "  ")
	s += b.Right.String(indent + "  ")
	return s
}

type MathParser struct {
	l         *MathLexer
	curToken  MathToken
	peekToken MathToken
}

func NewMathParser(input string) *MathParser {
	l := NewLexer(input)
	p := &MathParser{l: l}
	// инициализируем curToken и peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *MathParser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *MathParser) ParseMathExpr(str string) Expr {
	p.l = NewLexer(str)
	return p.ParseMath()
}

// ParseMath — точка входа разбора всего выражения
func (p *MathParser) ParseMath() Expr {
	return p.parseExpression()
}

// parseExpression обрабатывает + и -
func (p *MathParser) parseExpression() Expr {
	left := p.parseTerm()
	for p.curToken.Type == PLUS || p.curToken.Type == MINUS {
		op := p.curToken.Value
		p.nextToken()
		right := p.parseTerm()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

// parseTerm обрабатывает * и /
func (p *MathParser) parseTerm() Expr {
	left := p.parsePower()
	for p.curToken.Type == ASTERISK || p.curToken.Type == SLASH {
		op := p.curToken.Value
		p.nextToken()
		right := p.parsePower()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

// parsePower обрабатывает ^ (правый ассоциативный)
func (p *MathParser) parsePower() Expr {
	left := p.parseUnary()
	if p.curToken.Type == CARET {
		op := p.curToken.Value
		p.nextToken()
		right := p.parsePower()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

// parseUnary обрабатывает унарные + и -
func (p *MathParser) parseUnary() Expr {
	if p.curToken.Type == PLUS || p.curToken.Type == MINUS {
		op := p.curToken.Value
		p.nextToken()
		expr := p.parseUnary()
		return &UnaryExpr{Op: op, Expr: expr}
	}
	return p.parsePrimary()
}

// parsePrimary — число или скобки
func (p *MathParser) parsePrimary() Expr {
	switch p.curToken.Type {
	case NUMBER:
		val, _ := strconv.ParseFloat(p.curToken.Value, 64)
		node := &NumberLiteral{Value: val}
		p.nextToken()
		return node
	case LPAREN:
		p.nextToken()
		exp := p.parseExpression()
		if p.curToken.Type != RPAREN {
			panic("expected ')'")
		}
		p.nextToken()
		return exp
	default:
		panic(fmt.Sprintf("unexpected token %q", p.curToken.Value))
	}
}
