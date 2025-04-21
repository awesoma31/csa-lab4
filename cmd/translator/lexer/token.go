package lexer

import "slices"

import "fmt"

type Token struct {
	Kind  TokenKind
	Value string
}

type TokenKind int

const (
	EOF TokenKind = iota
	TRUE
	FALSE
	NUMBER
	STRING
	IDENTIFIER

	// Grouping & Braces
	OPEN_BRACKET
	CLOSE_BRACKET
	OPEN_CURLY
	CLOSE_CURLY
	OPEN_PAREN
	CLOSE_PAREN

	// Equivilance
	ASSIGNMENT
	EQUALS
	NOT_EQUALS
	NOT

	// Conditional
	LESS
	LESS_EQUALS
	GREATER
	GREATER_EQUALS

	// Logical
	OR
	AND

	// Symbols
	DOT
	DOT_DOT
	SEMI_COLON
	COLON
	QUESTION
	COMMA

	// Shorthand
	//PLUS_PLUS
	//MINUS_MINUS
	//PLUS_EQUALS
	//MINUS_EQUALS
	// NULLISH_ASSIGNMENT // ??=

	//Maths
	PLUS  // +
	DASH  // -
	SLASH // /
	STAR  // *
	// PERCENT // %

	// Reserved Keywords
	VAR
	//CONST
	//CLASS
	//NEW
	//IMPORT
	//FROM
	FN
	IF
	ELSE
	//FOREACH
	WHILE
	FOR
	//EXPORT
	//TYPEOF
	//IN

	// Misc
	NUM_TOKENS
)

var reserved_lu map[string]TokenKind = map[string]TokenKind{
	"true":  TRUE,
	"false": FALSE,
	"var":   VAR,
	"fn":    FN,
	"if":    IF,
	"else":  ELSE,
	"while": WHILE,
	"for":   FOR,
}

func (tk Token) IsOneOfMany(expectedTokens ...TokenKind) bool {
	return slices.Contains(expectedTokens, tk.Kind)
}

func (token Token) Debug() {
	if token.Kind == IDENTIFIER || token.Kind == NUMBER || token.Kind == STRING {
		fmt.Printf("%s(%s)\n", TokenKindString(token.Kind), token.Value)
	} else {
		fmt.Printf("%s()\n", TokenKindString(token.Kind))
	}
}

func TokenKindString(kind TokenKind) string {
	switch kind {
	case EOF:
		return "eof"
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
	case OPEN_BRACKET:
		return "open_bracket"
	case CLOSE_BRACKET:
		return "close_bracket"
	case OPEN_CURLY:
		return "open_curly"
	case CLOSE_CURLY:
		return "close_curly"
	case OPEN_PAREN:
		return "open_paren"
	case CLOSE_PAREN:
		return "close_paren"
	case ASSIGNMENT:
		return "assignment"
	case EQUALS:
		return "equals"
	case NOT_EQUALS:
		return "not_equals"
	case NOT:
		return "not"
	case LESS:
		return "less"
	case LESS_EQUALS:
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
	case DOT_DOT:
		return "dot_dot"
	case SEMI_COLON:
		return "semi_colon"
	case COLON:
		return "colon"
	case QUESTION:
		return "question"
	case COMMA:
		return "comma"
	case PLUS:
		return "plus"
	case DASH:
		return "dash"
	case SLASH:
		return "slash"
	case STAR:
		return "star"
	case VAR:
		return "VAR"
	case FN:
		return "fn"
	case IF:
		return "if"
	case ELSE:
		return "else"
	case FOR:
		return "for"
	case WHILE:
		return "while"
	default:
		return fmt.Sprintf("unknown(%d)", kind)
	}
}

func newUniqueToken(kind TokenKind, value string) Token {
	return Token{
		kind, value,
	}
}
