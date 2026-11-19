// Package parser is a recursive-descent parser over the prefix grammar:
//
//	expr ::= NUMBER
//	       | IDENT '(' expr (',' expr)+ ')'
//
// The input language is pure prefix (functions and parentheses only), so
// operator precedence never enters the parsing picture — a one-function
// recursive descent is enough. All precedence/associativity work happens
// later, in the printer package.
package parser

import (
	"fmt"

	"expr/internal/ast"
	"expr/internal/token"
)

// ParseError is returned on a syntactically invalid program. It carries
// the source position of the offending token so the driver can report it.
type ParseError struct {
	Msg  string
	Line int
	Col  int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("syntax error at %d:%d: %s", e.Line, e.Col, e.Msg)
}

// fnInfo is the metadata the parser attaches to each known function name
// when it builds an Op node. Stored once in the `functions` table below.
type fnInfo struct {
	symbol     string
	prec       int
	rightAssoc bool
	arity      int
}

// functions maps every function name the language recognises to the
// operator metadata the parser bakes into the AST.
var functions = map[string]fnInfo{
	"tern": {"?:", 1, true, 3},
	"add":  {"+", 2, false, 2},
	"sub":  {"-", 2, false, 2},
	"mul":  {"*", 3, false, 2},
	"div":  {"/", 3, false, 2},
	"mod":  {"%", 3, false, 2},
	"pow":  {"^", 4, true, 2},
}

// Parser walks a pre-lexed token slice.
type Parser struct {
	toks []token.Token
	i    int
}

// New creates a parser positioned at the first token.
func New(tokens []token.Token) *Parser {
	return &Parser{toks: tokens, i: 0}
}

// Parse consumes the whole token stream and returns the AST. If anything
// but EOF is left over after the top-level expression, that's a syntax
// error.
func (p *Parser) Parse() (ast.Node, error) {
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	trailing := p.peek()
	if trailing.Kind != token.EOF {
		return nil, &ParseError{
			Msg:  fmt.Sprintf("unexpected trailing %s '%s'", trailing.Kind, trailing.Lexeme),
			Line: trailing.Line, Col: trailing.Col,
		}
	}
	return node, nil
}

// -----------------------------------------------------------------
// Cursor helpers
// -----------------------------------------------------------------

func (p *Parser) peek() token.Token { return p.toks[p.i] }

func (p *Parser) advance() token.Token {
	t := p.toks[p.i]
	if t.Kind != token.EOF {
		p.i++
	}
	return t
}

func (p *Parser) expect(kind token.Kind, what string) (token.Token, error) {
	t := p.peek()
	if t.Kind != kind {
		return token.Token{}, &ParseError{
			Msg:  fmt.Sprintf("expected %s, got %s '%s'", what, t.Kind, t.Lexeme),
			Line: t.Line, Col: t.Col,
		}
	}
	return p.advance(), nil
}

// -----------------------------------------------------------------
// Grammar rules
// -----------------------------------------------------------------

func (p *Parser) parseExpr() (ast.Node, error) {
	t := p.peek()

	switch t.Kind {
	case token.NUMBER:
		p.advance()
		return &ast.Num{Text: t.Lexeme}, nil
	case token.IDENT:
		return p.parseCall()
	default:
		return nil, &ParseError{
			Msg:  fmt.Sprintf("expected number or function name, got %s '%s'", t.Kind, t.Lexeme),
			Line: t.Line, Col: t.Col,
		}
	}
}

func (p *Parser) parseCall() (ast.Node, error) {
	name := p.advance() // IDENT

	info, ok := functions[name.Lexeme]
	if !ok {
		return nil, &ParseError{
			Msg:  fmt.Sprintf("unknown function '%s'", name.Lexeme),
			Line: name.Line, Col: name.Col,
		}
	}

	if _, err := p.expect(token.LPAREN, "'(' after function name"); err != nil {
		return nil, err
	}

	first, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	args := []ast.Node{first}

	for p.peek().Kind == token.COMMA {
		p.advance()
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	if _, err := p.expect(token.RPAREN, "')' to close function call"); err != nil {
		return nil, err
	}

	if len(args) != info.arity {
		return nil, &ParseError{
			Msg: fmt.Sprintf("'%s' expects %d argument(s), got %d",
				name.Lexeme, info.arity, len(args)),
			Line: name.Line, Col: name.Col,
		}
	}

	return &ast.Op{
		Symbol:     info.symbol,
		Prec:       info.prec,
		RightAssoc: info.rightAssoc,
		Children:   args,
	}, nil
}
