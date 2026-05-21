// Command envx is the entrypoint for the envx CLI.
//
// It is intentionally minimal: all behaviour lives in internal packages so it
// can be unit tested. main only builds the command tree, executes it and
// decides the process exit code — no other file in the project calls os.Exit.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/panic-at/envx/internal/cli"
)

func main() {
	err := cli.NewRootCmd().Execute()
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "envx:", err)

	var exit *cli.ExitError
	if errors.As(err, &exit) {
		os.Exit(exit.Code)
	}
	os.Exit(1)
}
