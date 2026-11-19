// Package token defines the lexical tokens emitted by the lexer and the
// identifiers used to classify them. Keeping this in its own package means
// the parser and lexer can depend on it without depending on each other.
package token

import "fmt"

// Kind is the category of a token. It is a small enum implemented the
// idiomatic Go way: a named integer type with `iota`-generated constants.
type Kind int

const (
	IDENT Kind = iota
	NUMBER
	LPAREN
	RPAREN
	COMMA
	EOF
)

// String returns the human-readable name of the kind. Implementing the
// fmt.Stringer interface lets the Printf family render kinds nicely.
func (k Kind) String() string {
	switch k {
	case IDENT:
		return "IDENT"
	case NUMBER:
		return "NUMBER"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case COMMA:
		return "COMMA"
	case EOF:
		return "EOF"
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// Token is a single lexical token with its source position.
//
// The lexeme is kept exactly as it appeared in the source so that numeric
// literals round-trip untouched (e.g. "5e2" stays "5e2", not "500.0").
type Token struct {
	Kind   Kind
	Lexeme string
	Line   int // 1-based
	Col    int // 1-based
}

// String formats a token like IDENT('add')@1:4 — matches the Python/Java
// debug dumps so the --tokens output is identical across implementations.
func (t Token) String() string {
	return fmt.Sprintf("%s('%s')@%d:%d", t.Kind, t.Lexeme, t.Line, t.Col)
}
