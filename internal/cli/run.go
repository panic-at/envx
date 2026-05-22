package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/resolver"
	"github.com/panic-at/envx/internal/runner"
)

// newRunCmd builds the "envx run" command.
func newRunCmd(opts *rootOptions) *cobra.Command {
	var (
		profileName string
		inherit     bool
		noInherit   bool
		override    bool
		noOverride  bool
		allowErrors bool
	)

	cmd := &cobra.Command{
		Use:   "run --profile <name> -- <command> [args...]",
		Short: "Run a command with a profile's variables injected",
		Long: "run resolves every variable of a profile and executes a command as a\n" +
			"child process with those variables in its environment. The variables\n" +
			"live only in the child: they never touch the envx process or your\n" +
			"shell.\n\n" +
			"Everything after the '--' separator is passed to the command verbatim.\n" +
			"The child inherits this terminal's stdin, stdout and stderr; SIGINT and\n" +
			"SIGTERM are forwarded to it, and its exit code becomes envx's exit code.",
		Example: "  # Run a server with the dev profile injected\n" +
			"  envx run --profile dev -- ./server --addr :8080\n\n" +
			"  # Inspect the injected environment\n" +
			"  envx run --profile dev -- printenv\n\n" +
			"  # Run with only the profile's variables, nothing inherited\n" +
			"  envx run --profile ci --no-inherit -- ./test.sh",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --profile is required; a missing value is a usage error.
			if profileName == "" {
				return &ExitError{Code: 2, Err: errors.New("required flag --profile not set")}
			}

			command, err := commandAfterDash(cmd, args)
			if err != nil {
				return err
			}

			inheritFinal, err := resolveToggle(cmd, "inherit", "no-inherit", true)
			if err != nil {
				return err
			}
			overrideFinal, err := resolveToggle(cmd, "override", "no-override", true)
			if err != nil {
				return err
			}

			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[profileName]; !ok {
				return fmt.Errorf("profile %q does not exist", profileName)
			}
			eff, err := cfg.Effective(profileName)
			if err != nil {
				return err
			}

			result := resolver.ResolveAll(cmd.Context(), resolver.DefaultRegistry(), eff)
			if len(result.Errors) > 0 {
				failed := sortedErrorKeys(result.Errors)
				if !allowErrors {
					return fmt.Errorf("failed to resolve %d variable(s): %s (use --allow-errors to run anyway)",
						len(failed), strings.Join(failed, ", "))
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: omitting %d unresolved variable(s): %s\n",
					len(failed), strings.Join(failed, ", "))
			}

			code, runErr := runner.Run(cmd.Context(), runner.Options{
				Command:  command,
				Vars:     result.Values,
				Inherit:  inheritFinal,
				Override: overrideFinal,
				Stdin:    cmd.InOrStdin(),
				Stdout:   cmd.OutOrStdout(),
				Stderr:   cmd.ErrOrStderr(),
			})
			if runErr != nil {
				// The command never started: report why, with the
				// shell-convention exit code from the runner.
				return &ExitError{Code: code, Err: runErr}
			}
			if code != 0 {
				// The command ran and chose a non-zero status. Propagate it
				// silently: the child has already produced its own output.
				return &ExitError{Code: code, Silent: true,
					Err: fmt.Errorf("command exited with status %d", code)}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to run with (required)")
	cmd.Flags().BoolVar(&inherit, "inherit", true, "include the host environment in the child")
	cmd.Flags().BoolVar(&noInherit, "no-inherit", false, "run with only the profile's variables")
	cmd.Flags().BoolVar(&override, "override", true, "profile variables win conflicts with the host")
	cmd.Flags().BoolVar(&noOverride, "no-override", false, "host variables win conflicts with the profile")
	cmd.Flags().BoolVar(&allowErrors, "allow-errors", false, "run even if some variables fail to resolve")
	return cmd
}

// commandAfterDash extracts the command to execute from the positional
// arguments. envx run requires an explicit "--" separator, with the command
// after it and nothing before it. A violation is a usage error (exit 2).
func commandAfterDash(cmd *cobra.Command, args []string) ([]string, error) {
	dash := cmd.ArgsLenAtDash()
	if dash == -1 {
		return nil, &ExitError{Code: 2,
			Err: errors.New("missing '--' separator before the command (usage: envx run --profile <name> -- <command>)")}
	}
	if dash > 0 {
		return nil, &ExitError{Code: 2,
			Err: fmt.Errorf("unexpected arguments before '--': %s", strings.Join(args[:dash], " "))}
	}
	command := args[dash:]
	if len(command) == 0 {
		return nil, &ExitError{Code: 2, Err: errors.New("no command given after '--'")}
	}
	return command, nil
}

// resolveToggle reads a pair of opposed boolean flags — a positive flag and its
// "no-" counterpart — and returns the effective value. Setting both is a usage
// error. When neither is set, def is returned.
func resolveToggle(cmd *cobra.Command, pos, neg string, def bool) (bool, error) {
	posSet := cmd.Flags().Changed(pos)
	negSet := cmd.Flags().Changed(neg)
	if posSet && negSet {
		return false, &ExitError{Code: 2,
			Err: fmt.Errorf("--%s and --%s are mutually exclusive", pos, neg)}
	}
	switch {
	case posSet:
		v, _ := cmd.Flags().GetBool(pos)
		return v, nil
	case negSet:
		v, _ := cmd.Flags().GetBool(neg)
		return !v, nil
	default:
		return def, nil
	}
}
