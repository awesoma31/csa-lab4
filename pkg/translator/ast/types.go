package ast

// TypeKind represents the kind of type (e.g., int, string, bool).
type TypeKind int

const (
	TypeUnknown  TypeKind = iota // 0
	TypeInt                      // 1
	TypeString                   // 2
	TypeBool                     // 3
	TypeFunction                 // 4
	TypeList                     // 5
	TypeSymbol                   // 6 // Used for identifiers in type position
	// Add more as needed
)

// String returns the string representation of the TypeKind.
func (tk TypeKind) String() string {
	switch tk {
	case TypeInt:
		return "int"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeFunction:
		return "function"
	case TypeList:
		return "list"
	case TypeSymbol:
		return "symbol"
	default:
		return "unknown"
	}
}

// SymbolType represents a named type (e.g., "int", "string", "MyClass").
type SymbolType struct {
	Value string   // The name of the type (e.g., "int", "string")
	Kind  TypeKind // The kind of this symbol type (e.g., TypeInt, TypeString)
}

func (t SymbolType) _type() {}

// String implements the Type interface for SymbolType.
func (t SymbolType) String() string {
	// If Kind is more specific, use it. Otherwise, use the Value.
	if t.Kind != TypeUnknown && t.Kind != TypeSymbol {
		return t.Kind.String()
	}
	return t.Value
}

// ListType represents a list/array type (e.g., "[]int", "[]string").
type ListType struct {
	Underlying Type // The type of elements in the list (e.g., IntType for "[]int")
}

func (t ListType) _type() {}

// String implements the Type interface for ListType.
func (t ListType) String() string {
	if t.Underlying == nil {
		return "[]unknown"
	}
	return "[]" + t.Underlying.String()
}

// Predefined types for convenience
var (
	IntType    = SymbolType{Value: "int", Kind: TypeInt}
	StringType = SymbolType{Value: "string", Kind: TypeString}
	BoolType   = SymbolType{Value: "bool", Kind: TypeBool}
	// Add more predefined types as needed
)
