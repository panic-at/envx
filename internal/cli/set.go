package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/profile"
	"github.com/panic-at/envx/internal/resolver"
)

// envKeyPattern is the CLI-level rule for variable names: the conventional
// uppercase POSIX environment variable form.
const envKeyPattern = `^[A-Z_][A-Z0-9_]*$`

var envKeyRe = regexp.MustCompile(envKeyPattern)

// newSetCmd builds the "envx set" command.
func newSetCmd(opts *rootOptions) *cobra.Command {
	var (
		profileName string
		ref         string
		sensitive   bool
	)

	cmd := &cobra.Command{
		Use:   "set <KEY> [value]",
		Short: "Set a variable in a profile",
		Long: "set defines a variable as either a literal value (positional) or a\n" +
			"resolver reference (--ref). An existing variable is overwritten.",
		Example: "  # A literal value\n" +
			"  envx set PORT 8080 --profile dev\n\n" +
			"  # A secret literal, masked in show/diff output\n" +
			"  envx set API_KEY s3cret --profile dev --sensitive\n\n" +
			"  # A reference resolved at run time from a vault\n" +
			"  envx set DB_PASSWORD --ref op://vault/db/password --profile dev",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if !envKeyRe.MatchString(key) {
				return fmt.Errorf("invalid variable name %q (must match %s)", key, envKeyPattern)
			}

			hasValue := len(args) == 2
			switch {
			case hasValue && ref != "":
				return errors.New("cannot set both a literal value and --ref")
			case !hasValue && ref == "":
				return errors.New("provide a literal value or --ref <uri>")
			}

			v := profile.Var{Sensitive: sensitive}
			if ref != "" {
				scheme, err := uriScheme(ref)
				if err != nil {
					return err
				}
				if _, ok := resolver.DefaultRegistry().Get(scheme); !ok {
					return fmt.Errorf("unknown resolver scheme %q in --ref %q", scheme, ref)
				}
				v.Type = profile.VarRef
				v.URI = ref
			} else {
				v.Type = profile.VarLiteral
				v.Value = args[1]
			}

			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			p, ok := cfg.Profiles[profileName]
			if !ok {
				return fmt.Errorf("profile %q does not exist", profileName)
			}
			if p.Vars == nil {
				p.Vars = map[string]profile.Var{}
			}
			p.Vars[key] = v
			cfg.Profiles[profileName] = p
			if err := cfg.Save(opts.configPath); err != nil {
				return err
			}

			if ref != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Set %s (ref) in profile '%s'\n", key, profileName)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Set %s in profile '%s'\n", key, profileName)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "target profile (required)")
	cmd.Flags().StringVar(&ref, "ref", "", "resolver URI instead of a literal value")
	cmd.Flags().BoolVar(&sensitive, "sensitive", false, "mark the value as sensitive")
	_ = cmd.MarkFlagRequired("profile")
	return cmd
}

// uriScheme extracts the scheme component (the text before "://") of a
// resolver reference URI.
func uriScheme(uri string) (string, error) {
	i := strings.Index(uri, "://")
	if i <= 0 {
		return "", fmt.Errorf("invalid --ref URI %q: missing scheme", uri)
	}
	return uri[:i], nil
}
