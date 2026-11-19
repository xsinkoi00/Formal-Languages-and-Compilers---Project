// Package ast defines the abstract-syntax tree nodes shared by the parser
// and the printer. Two concrete kinds exist: Num (a numeric literal leaf)
// and Op (an operator node with 2 or 3 children). A tiny helper, Dump,
// prints the tree in a readable form for the --ast debug flag.
package ast

import (
	"fmt"
	"strings"
)

// Node is the common interface every AST node implements. The unexported
// `node()` method is a Go idiom for a "sealed" interface: only types in
// this package can satisfy it, so a type switch elsewhere knows the two
// cases (*Num, *Op) are exhaustive.
type Node interface {
	node()
}

// Num is a numeric-literal leaf. The original source lexeme is kept
// verbatim so the printer can reproduce it ("5e2" stays "5e2").
type Num struct {
	Text string
}

func (*Num) node() {}

// Op is an operator node.
//
//   - Symbol is the infix spelling the printer will emit
//     (one of "+", "-", "*", "/", "%", "^", "?:").
//   - Prec is 1 (lowest, "?:") .. 4 (highest, "^") and drives the
//     parenthesisation rule in the printer.
//   - RightAssoc marks right-associative operators ("^" and "?:").
//   - Children holds 2 items for binary operators, 3 for the ternary.
type Op struct {
	Symbol     string
	Prec       int
	RightAssoc bool
	Children   []Node
}

func (*Op) node() {}

// Dump formats the tree as an indented, human-readable block — used by the
// --ast debug flag and by anyone eyeballing a parse.
func Dump(n Node) string {
	var sb strings.Builder
	dump(&sb, n, 0)
	return sb.String()
}

func dump(sb *strings.Builder, n Node, indent int) {
	pad := strings.Repeat("  ", indent)
	switch v := n.(type) {
	case *Num:
		fmt.Fprintf(sb, "%sNum('%s')", pad, v.Text)
	case *Op:
		assoc := "L"
		if v.RightAssoc {
			assoc = "R"
		}
		fmt.Fprintf(sb, "%sOp('%s', prec=%d, %s)", pad, v.Symbol, v.Prec, assoc)
		for _, c := range v.Children {
			sb.WriteByte('\n')
			dump(sb, c, indent+1)
		}
	}
}
