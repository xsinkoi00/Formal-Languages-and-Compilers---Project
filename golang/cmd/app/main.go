// Expression Transformation Project — Go implementation.
// Reads a prefix function-composition expression from stdin and writes the
// equivalent infix mathematical expression (with the minimum number of
// parentheses) to stdout.
//
// Pipeline: stdin -> Lexer -> Parser -> AST -> Printer -> stdout
//
// Exit codes:
//
//	0  success
//	2  lexical error
//	3  syntax error
//	4  I/O error
//
// Debug flags:
//
//	--tokens  dump the token stream and exit
//	--ast     dump the parsed AST and exit
package main

import (
	"fmt"
	"io"
	"os"

	"expr/internal/ast"
	"expr/internal/lexer"
	"expr/internal/parser"
	"expr/internal/printer"
)

func main() {
	srcBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "I/O error:", err)
		os.Exit(4)
	}
	src := string(srcBytes)
	args := os.Args[1:]

	tokens, err := lexer.New(src).Tokenize()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if hasFlag(args, "--tokens") {
		for _, t := range tokens {
			fmt.Println(t)
		}
		return
	}

	tree, err := parser.New(tokens).Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	if hasFlag(args, "--ast") {
		fmt.Println(ast.Dump(tree))
		return
	}

	fmt.Println(printer.Pretty(tree))
}

// hasFlag reports whether the argument list contains the given flag.
// A tiny helper keeps us free of the `flag` package until we actually
// need named options.
func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}
