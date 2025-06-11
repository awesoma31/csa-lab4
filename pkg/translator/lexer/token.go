package lexer

import (
	"fmt"
	"slices"
)

type TokenKind int

const (
	EOF TokenKind = iota
	NULL
	TRUE
	FALSE
	NUMBER
	STRING
	IDENTIFIER

	// Grouping & Braces
	OpenBracket
	CloseBracket
	OpenCurly
	CloseCurly
	OpenParen
	CloseParen

	// Equivalence
	ASSIGNMENT
	EQUALS
	NotEquals
	NOT

	// Conditional

	LESS
	LessEquals
	GREATER
	GREATER_EQUALS

	// Logical
	OR
	AND

	// Symbols
	DOT
	DotDot
	SemiColon
	COLON
	QUESTION
	COMMA

	// Shorthand
	PlusPlus
	MinusMinus
	PlusEquals
	MinusEquals
	NullishAssignment // ??=

	//Maths
	PLUS
	MINUS // -
	SLASH
	STAR
	PERCENT

	// Reserved Keywords
	LET
	IF
	ELSE
	WHILE
	IN

	INTER
	IntOn
	IntOff

	PRINT
	READCH
	READINT
	ADDSTR

	LIST

	UNKNOWN
)

// RESERVED WORDS FOUND HERE
var reservedLu = map[string]TokenKind{
	"true":    TRUE,
	"false":   FALSE,
	"let":     LET,
	"if":      IF,
	"else":    ELSE,
	"while":   WHILE,
	"print":   PRINT,
	"read":    READCH,
	"readInt": READINT,
	"inter":   INTER,
	"intOn":   IntOn,
	"intOff":  IntOff,
	"list":    LIST,
	"addStr":  ADDSTR,
}

type Token struct {
	Kind  TokenKind
	Value string
}

func (tk Token) IsOneOfMany(expectedTokens ...TokenKind) bool {
	return slices.Contains(expectedTokens, tk.Kind)
}

func (tk Token) Debug() {
	if tk.Kind == IDENTIFIER || tk.Kind == NUMBER || tk.Kind == STRING {
		fmt.Printf("%s(%s)\n", TokenKindString(tk.Kind), tk.Value)
	} else {
		fmt.Printf("%s()\n", TokenKindString(tk.Kind))
	}
}

func TokenKindString(kind TokenKind) string {
	switch kind {
	case EOF:
		return "eof"
	case NULL:
		return "null"
	case NUMBER:
		return "number"
	case STRING:
		return "string"
	case TRUE:
		return "true"
	case FALSE:
		return "false"
	case IDENTIFIER:
		return "identifier"
	case OpenBracket:
		return "open_bracket"
	case CloseBracket:
		return "close_bracket"
	case OpenCurly:
		return "open_curly"
	case CloseCurly:
		return "close_curly"
	case OpenParen:
		return "open_paren"
	case CloseParen:
		return "close_paren"
	case ASSIGNMENT:
		return "assignment"
	case EQUALS:
		return "equals"
	case NotEquals:
		return "not_equals"
	case NOT:
		return "not"
	case LESS:
		return "less"
	case LessEquals:
		return "less_equals"
	case GREATER:
		return "greater"
	case GREATER_EQUALS:
		return "greater_equals"
	case OR:
		return "or"
	case AND:
		return "and"
	case DOT:
		return "dot"
	case DotDot:
		return "dot_dot"
	case SemiColon:
		return "semi_colon"
	case COLON:
		return "colon"
	case COMMA:
		return "comma"
	case PlusPlus:
		return "plus_plus"
	case MinusMinus:
		return "minus_minus"
	case PlusEquals:
		return "plus_equals"
	case MinusEquals:
		return "minus_equals"
	case NullishAssignment:
		return "nullish_assignment"
	case PLUS:
		return "plus"
	case MINUS:
		return "minus"
	case SLASH:
		return "slash"
	case STAR:
		return "star"
	case PERCENT:
		return "percent"
	case LET:
		return "let"
	case IF:
		return "if"
	case ELSE:
		return "else"
	case WHILE:
		return "while"
	case PRINT:
		return "print"
	case READCH:
		return "read"
	case ADDSTR:
		return "addStr"
	default:
		return fmt.Sprintf("unknown(%d)", kind)
	}
}

func newUniqueToken(kind TokenKind, value string) Token {
	return Token{
		kind, value,
	}
}
