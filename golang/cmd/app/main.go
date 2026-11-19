// Expression Transformation Project — Go implementation.
// Reads a prefix function-composition expression from stdin and
// writes the equivalent infix mathematical expression to stdout.
//
// Pipeline: stdin -> Lexer -> Parser -> AST -> Printer -> stdout
// Sta
//
// Exit codes:
//
//	0  success
//	2  lexical error
//	3  syntax error  (coming in Phase 2)
//	4  I/O error
//
// Debug flags:
//
//	--tokens  dump the token stream and exit
package main

import (
	"fmt"
	"io"
	"os"

	"expr/internal/lexer"
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
	_ = tokens
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
