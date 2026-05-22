package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/version"
)

// rootOptions holds the values of the persistent (global) flags, shared by
// reference with every subcommand.
type rootOptions struct {
	configPath string
	noColor    bool
}

// ExitError wraps an error with the process exit code it should produce.
// cmd/envx/main.go inspects returned errors with errors.As to choose the exit
// code; any error that is not an *ExitError maps to exit code 1.
//
// Silent suppresses the "envx: ..." message main.go normally prints. It is set
// when a child command run by "envx run" exits non-zero: the child has already
// reported its own failure, so envx only needs to mirror the exit code.
type ExitError struct {
	Code   int
	Err    error
	Silent bool
}

// Error returns the wrapped error's message.
func (e *ExitError) Error() string { return e.Err.Error() }

// Unwrap returns the wrapped error so errors.Is and errors.As see through it.
func (e *ExitError) Unwrap() error { return e.Err }

// NewRootCmd builds the root cobra command with every subcommand wired in.
//
// Output goes to the command's configured writers; callers and tests may
// redirect them with SetOut and SetErr. The returned command is ready to
// Execute.
func NewRootCmd() *cobra.Command {
	opts := &rootOptions{}

	cmd := &cobra.Command{
		Use:   "envx",
		Short: "envx manages layered environment-variable profiles",
		Long: "envx manages environment variables as named, inheritable profiles.\n" +
			"Values may be stored inline or referenced from external secret stores.",
		Example: "  envx init\n" +
			"  envx profile add dev\n" +
			"  envx set DATABASE_URL postgres://localhost/dev --profile dev\n" +
			"  envx run --profile dev -- ./server",
		Version:       version.Info(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	// Flag-parsing failures (unknown flag, missing required flag, ...) are
	// usage errors and exit with code 2.
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &ExitError{Code: 2, Err: err}
	})

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", config.DefaultPath(),
		"path to the envx config file")
	cmd.PersistentFlags().BoolVar(&opts.noColor, "no-color", noColorDefault(),
		"disable colored output")

	cmd.AddCommand(
		newInitCmd(opts),
		newProfileCmd(opts),
		newSetCmd(opts),
		newShowCmd(opts),
		newDiffCmd(opts),
		newExportCmd(opts),
		newRunCmd(opts),
	)
	return cmd
}

// noColorDefault reports whether colored output should be disabled by default,
// honouring the NO_COLOR convention: any presence of the variable disables
// color regardless of its value.
func noColorDefault() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}
