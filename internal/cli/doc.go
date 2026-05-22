// Package cli implements the envx command-line interface on top of Cobra.
//
// NewRootCmd builds the command tree; each subcommand lives in its own file.
// Commands never write to os.Stdout or os.Stderr directly — they use the
// writers attached to their *cobra.Command (OutOrStdout, ErrOrStderr) so that
// tests can capture output.
//
// Command failures are reported by returning an error from RunE; only
// cmd/envx/main.go decides the process exit code. An *ExitError carries an
// explicit exit code, letting commands distinguish usage errors, resolution
// failures and the propagated exit status of a child process.
package cli
