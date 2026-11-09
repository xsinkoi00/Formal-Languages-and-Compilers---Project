// Package greeter builds greeting strings. It lives under internal/ so that
// only this module can import it — matching the risk_audit layout where
// business logic sits in internal/<domain>.
package greeter

import "fmt"

// Greet returns a friendly greeting for the given name.
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
