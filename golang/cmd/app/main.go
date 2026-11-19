// Expression Transformation Project — Go implementation.
// Reads a prefix function-composition expression from stdin and (once
// wired up) writes the equivalent infix mathematical expression to stdout.

package main

import (
	"io"
	"os"
)

func main() {
	if _, err := io.ReadAll(os.Stdin); err != nil {
		os.Exit(4)
	}
}
