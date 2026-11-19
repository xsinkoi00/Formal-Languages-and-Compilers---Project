// Package printer renders an AST back to an infix mathematical expression
// using the minimum number of parentheses.
//
// For every operator child of an operator parent:
//
//	let pp = parent.Prec, pc = child.Prec
//	wrap the child in '(' ... ')' iff
//	    pc <  pp,  OR
//	    pc == pp AND the child sits on the "wrong" side for the
//	                 parent's associativity:
//	                   - parent is right-assoc -> the LEFT child needs parens
//	                   - parent is left-assoc  -> the RIGHT child needs parens
//
// A Num child never needs parentheses. For the ternary (symbol "?:") the
// middle operand sits between '?' and ':' and is syntactically unambiguous
// regardless of its own precedence, so it is never wrapped.
//
// Spacing — chosen to reproduce all four assignment examples byte-for-byte:
//
//	'+' '-' '*' '/' '%'   -> surrounded by single spaces
//	'^' and '?:'          -> no spaces
package printer

import (
	"expr/internal/ast"
)

// spaced lists the operator symbols that get surrounded by single spaces
// in the output. A map-of-empty-structs is Go's idiomatic "set".
var spaced = map[string]struct{}{
	"+": {}, "-": {}, "*": {}, "/": {}, "%": {},
}

// side records which position a child occupies relative to its parent —
// used only by renderChild to apply the associativity rule.
type side int

const (
	leftSide side = iota
	rightSide
)

// Pretty returns the minimal-parentheses infix form of n.
func Pretty(n ast.Node) string {
	switch v := n.(type) {
	case *ast.Num:
		return v.Text
	case *ast.Op:
		if v.Symbol == "?:" {
			cond := renderChild(v, v.Children[0], leftSide)
			then := Pretty(v.Children[1]) // middle is always safe
			els := renderChild(v, v.Children[2], rightSide)
			return cond + "?" + then + ":" + els
		}
		// Binary operator.
		left := renderChild(v, v.Children[0], leftSide)
		right := renderChild(v, v.Children[1], rightSide)
		sep := v.Symbol
		if _, s := spaced[v.Symbol]; s {
			sep = " " + v.Symbol + " "
		}
		return left + sep + right
	}
	return "" // unreachable: Node is a sealed interface with exactly two cases.
}

// renderChild formats one child and wraps it in parens iff the
// precedence/associativity rule demands it.
func renderChild(parent *ast.Op, child ast.Node, s side) string {
	rendered := Pretty(child)

	childOp, ok := child.(*ast.Op)
	if !ok {
		return rendered // Num never needs parentheses.
	}

	needsParens := false
	switch {
	case childOp.Prec < parent.Prec:
		needsParens = true
	case childOp.Prec == parent.Prec:
		// Same precedence: parens only on the side that fights the
		// parent's associativity.
		if parent.RightAssoc && s == leftSide {
			needsParens = true
		} else if !parent.RightAssoc && s == rightSide {
			needsParens = true
		}
	}
	if needsParens {
		return "(" + rendered + ")"
	}
	return rendered
}
