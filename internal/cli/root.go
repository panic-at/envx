// Package cli implements the envx command-line interface on top of Cobra.
//
// Commands never write to os.Stdout or os.Stderr directly; they use the
// writers attached to their *cobra.Command (OutOrStdout, ErrOrStderr) so that
// tests can capture output. Command failures are reported by returning an
// error from RunE — only cmd/envx/main.go decides the process exit code.
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
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", config.DefaultPath(),
		"path to the envx config file")
	cmd.PersistentFlags().BoolVar(&opts.noColor, "no-color", noColorDefault(),
		"disable colored output")

	cmd.AddCommand(
		newInitCmd(opts),
		newProfileCmd(opts),
		newSetCmd(opts),
		newShowCmd(opts),
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
