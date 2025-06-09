package lexer

import "slices"

import "fmt"

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
	// CONST TODO: delete
	CONST
	CLASS
	NEW
	IMPORT
	FROM
	FN
	IF
	ELSE
	FOREACH
	WHILE
	FOR
	EXPORT
	TYPEOF
	IN

	INTER

	PRINT
	READ
	QOUTE

	RETURN

	UNKNOWN
)

// RESERVED WORDS FOUND HERE
var reservedLu = map[string]TokenKind{
	"true":    TRUE,
	"false":   FALSE,
	"null":    NULL,
	"let":     LET,
	"const":   CONST,
	"class":   CLASS,
	"new":     NEW,
	"import":  IMPORT,
	"from":    FROM,
	"fn":      FN,
	"if":      IF,
	"else":    ELSE,
	"foreach": FOREACH,
	"while":   WHILE,
	"for":     FOR,
	"export":  EXPORT,
	"typeof":  TYPEOF,
	"in":      IN,
	"print":   PRINT,
	"read":    READ,
	"inter":   INTER,
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
	// case QUESTION:
	// 	return "question"
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
	// case CONST:
	// 	return "const"
	// case CLASS:
	// 	return "class"
	// case NEW:
	// 	return "new"
	// case IMPORT:
	// 	return "import"
	// case FROM:
	// 	return "from"
	case FN:
		return "fn"
	case IF:
		return "if"
	case ELSE:
		return "else"
	// case FOREACH:
	// 	return "foreach"
	case FOR:
		return "for"
	case WHILE:
		return "while"
	// case EXPORT:
	// 	return "export"
	// case IN:
	// 	return "in"
	case PRINT:
		return "print"
	case READ:
		return "read"
	case QOUTE:
		return `"`
	default:
		return fmt.Sprintf("unknown(%d)", kind)
	}
}

func newUniqueToken(kind TokenKind, value string) Token {
	return Token{
		kind, value,
	}
}
