// Package lexer is an explicit deterministic finite automaton that turns
// source text into a slice of tokens. The state diagram lives in README.md;
// each sub-automaton below (comment, ident, number) is labelled with the
// states it visits.
//
// Accepting (token-emitting) states for whole tokens:
//   S_ID            -> IDENT
//   S_NUM_INT       -> NUMBER
//   S_NUM_FRAC      -> NUMBER
//   S_NUM_EXP       -> NUMBER
//   single-char final states for '(' ')' ','
//
// Whitespace and "/* ... */" block comments are consumed silently in state
// S0 and never produce a token.
package lexer

import (
	"fmt"

	"expr/internal/token"
)

// LexError is returned when the lexer rejects a character sequence.
// It satisfies the built-in `error` interface via the Error() method.
type LexError struct {
	Msg  string
	Line int
	Col  int
}

func (e *LexError) Error() string {
	return fmt.Sprintf("lexical error at %d:%d: %s", e.Line, e.Col, e.Msg)
}

// Lexer walks the source string one byte at a time. The grammar is ASCII
// only, so byte-level iteration is both correct and simpler than rune
// iteration — any non-ASCII byte becomes an "unexpected character" error.
type Lexer struct {
	src  string
	i    int // byte index of the next character to read
	line int // 1-based line of src[i]
	col  int // 1-based column of src[i]
}

// New creates a lexer positioned at the start of src.
func New(src string) *Lexer {
	return &Lexer{src: src, i: 0, line: 1, col: 1}
}

// Tokenize lexes the entire input, returning the token list (ending with
// an EOF token) or the first lexical error encountered.
func (l *Lexer) Tokenize() ([]token.Token, error) {
	var out []token.Token
	for {
		t, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		if t.Kind == token.EOF {
			return out, nil
		}
	}
}

// -----------------------------------------------------------------
// Cursor helpers
// -----------------------------------------------------------------

func (l *Lexer) hasMore() bool { return l.i < len(l.src) }

// peek returns the next byte without consuming it, or 0 at end of input.
func (l *Lexer) peek() byte {
	if l.i >= len(l.src) {
		return 0
	}
	return l.src[l.i]
}

// advance consumes the next byte and updates line/column.
func (l *Lexer) advance() byte {
	if l.i >= len(l.src) {
		return 0
	}
	c := l.src[l.i]
	l.i++
	if c == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return c
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// Identifiers in this language are only the seven function names, all
// lowercase ASCII letters. Keep the character class tight — anything else
// becomes a lexical error.
func isAlpha(c byte) bool { return c >= 'a' && c <= 'z' }

// -----------------------------------------------------------------
// FA driver — the top-level dispatch (state S0)
// -----------------------------------------------------------------

func (l *Lexer) nextToken() (token.Token, error) {
	for {
		// S0: skip whitespace.
		for l.hasMore() && isSpace(l.peek()) {
			l.advance()
		}
		startLine, startCol := l.line, l.col

		if !l.hasMore() {
			return token.Token{Kind: token.EOF, Line: startLine, Col: startCol}, nil
		}

		c := l.peek()
		switch {
		case c == '(':
			l.advance()
			return token.Token{Kind: token.LPAREN, Lexeme: "(", Line: startLine, Col: startCol}, nil
		case c == ')':
			l.advance()
			return token.Token{Kind: token.RPAREN, Lexeme: ")", Line: startLine, Col: startCol}, nil
		case c == ',':
			l.advance()
			return token.Token{Kind: token.COMMA, Lexeme: ",", Line: startLine, Col: startCol}, nil

		case c == '/':
			if err := l.consumeComment(startLine, startCol); err != nil {
				return token.Token{}, err
			}
			continue // back to S0 to look for the next real token

		case isAlpha(c):
			return l.scanIdent(startLine, startCol), nil

		case c == '+' || c == '-' || isDigit(c):
			return l.scanNumber(startLine, startCol)

		default:
			return token.Token{}, &LexError{
				Msg:  fmt.Sprintf("unexpected character '%c'", c),
				Line: startLine, Col: startCol,
			}
		}
	}
}

// -----------------------------------------------------------------
// Sub-automata
// -----------------------------------------------------------------

// consumeComment handles "/* ... */" — S0 --'/'--> S_SLASH --'*'--> S_COMMENT.
func (l *Lexer) consumeComment(startLine, startCol int) error {
	l.advance() // S0 -> S_SLASH (consumed '/')
	if l.peek() != '*' {
		return &LexError{
			Msg:  "stray '/' (did you mean '/* ... */'?)",
			Line: startLine, Col: startCol,
		}
	}
	l.advance() // S_SLASH -> S_COMMENT

	// S_COMMENT loops; '*' moves to S_COMMENT_STAR; from there '/' closes
	// the comment, '*' stays, anything else falls back to S_COMMENT.
	for {
		if !l.hasMore() {
			return &LexError{Msg: "unterminated block comment", Line: startLine, Col: startCol}
		}
		ch := l.advance()
		if ch == '*' {
			for l.peek() == '*' {
				l.advance()
			}
			if l.peek() == '/' {
				l.advance()
				return nil
			}
		}
	}
}

// scanIdent handles S0 --letter--> S_ID (loop letter) --other--> emit IDENT.
// The lexer only validates the character class; checking that the name is
// one of {add, sub, mul, div, mod, pow, tern} is the parser's job.
func (l *Lexer) scanIdent(startLine, startCol int) token.Token {
	var sb []byte
	for l.hasMore() && isAlpha(l.peek()) {
		sb = append(sb, l.advance())
	}
	return token.Token{Kind: token.IDENT, Lexeme: string(sb), Line: startLine, Col: startCol}
}

// scanNumber walks the number sub-automaton:
//   [sign]? digit+ ("." digit+)? ("e" [sign]? digit+)?
// Strict grammar: the sign and the dot each *require* a following digit.
func (l *Lexer) scanNumber(startLine, startCol int) (token.Token, error) {
	var sb []byte
	c := l.peek()

	// S_NUM_SIGN: optional leading '+'/'-' MUST be glued to a digit.
	if c == '+' || c == '-' {
		sb = append(sb, l.advance())
		if !isDigit(l.peek()) {
			return token.Token{}, &LexError{Msg: "expected digit after sign", Line: l.line, Col: l.col}
		}
	}

	// S_NUM_INT: integer digits.
	for isDigit(l.peek()) {
		sb = append(sb, l.advance())
	}

	// Optional fraction: '.' <digit-sequence>.
	if l.peek() == '.' {
		sb = append(sb, l.advance()) // S_NUM_INT -> S_NUM_DOT
		if !isDigit(l.peek()) {
			return token.Token{}, &LexError{Msg: "expected digit after '.'", Line: l.line, Col: l.col}
		}
		for isDigit(l.peek()) {
			sb = append(sb, l.advance()) // S_NUM_FRAC loop
		}
	}

	// Optional exponent: 'e' <optional-sign> <digit-sequence>. Lowercase 'e'
	// only — uppercase 'E' is rejected per the strict grammar.
	if l.peek() == 'e' {
		sb = append(sb, l.advance()) // -> S_NUM_E
		if l.peek() == '+' || l.peek() == '-' {
			sb = append(sb, l.advance()) // S_NUM_E -> S_NUM_ESIGN
		}
		if !isDigit(l.peek()) {
			return token.Token{}, &LexError{Msg: "expected digit in exponent", Line: l.line, Col: l.col}
		}
		for isDigit(l.peek()) {
			sb = append(sb, l.advance()) // S_NUM_EXP loop
		}
	}

	return token.Token{Kind: token.NUMBER, Lexeme: string(sb), Line: startLine, Col: startCol}, nil
}
