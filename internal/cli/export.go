package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/exporter"
	"github.com/panic-at/envx/internal/profile"
	"github.com/panic-at/envx/internal/resolver"
)

// newExportCmd builds the "envx export" command.
func newExportCmd(opts *rootOptions) *cobra.Command {
	var (
		profileName string
		format      string
		output      string
		allowErrors bool
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a resolved profile in a consumable format",
		Long: "export resolves every variable of a profile and serializes the\n" +
			"result as a dotenv file, JSON object or POSIX shell script.\n" +
			"If any variable fails to resolve, export aborts unless --allow-errors\n" +
			"is given, in which case the failing variables are omitted with a\n" +
			"warning on stderr.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// --profile is required; a missing value is a usage error (exit 2).
			if profileName == "" {
				return &ExitError{Code: 2, Err: errors.New("required flag --profile not set")}
			}
			exp, err := exporter.Get(exporter.Format(format))
			if err != nil {
				return &ExitError{Code: 2, Err: err}
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
					return fmt.Errorf("failed to resolve %d variable(s): %s (use --allow-errors to export anyway)",
						len(failed), strings.Join(failed, ", "))
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: omitting %d unresolved variable(s): %s\n",
					len(failed), strings.Join(failed, ", "))
			}

			if output == "" {
				return exp.Export(cmd.OutOrStdout(), result.Values)
			}

			var buf bytes.Buffer
			if err := exp.Export(&buf, result.Values); err != nil {
				return err
			}
			perm := os.FileMode(0o644)
			if hasSensitive(eff, result.Values) {
				perm = 0o600
			}
			if err := os.WriteFile(output, buf.Bytes(), perm); err != nil {
				return fmt.Errorf("write export file %s: %w", output, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Exported profile '%s' (%s) to %s\n", profileName, format, output)
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile to export (required)")
	cmd.Flags().StringVar(&format, "format", string(exporter.FormatDotenv),
		"output format: "+strings.Join(formatNames(), ", "))
	cmd.Flags().StringVar(&output, "output", "", "write to a file instead of stdout")
	cmd.Flags().BoolVar(&allowErrors, "allow-errors", false,
		"export even if some variables fail to resolve")
	return cmd
}

// sortedErrorKeys returns the keys of errs in ascending lexical order.
func sortedErrorKeys(errs map[string]error) []string {
	keys := make([]string, 0, len(errs))
	for k := range errs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// hasSensitive reports whether any exported variable is marked sensitive in
// the profile, which decides the permissions of an --output file.
func hasSensitive(p profile.Profile, values map[string]string) bool {
	for k := range values {
		if p.Vars[k].Sensitive {
			return true
		}
	}
	return false
}

// formatNames returns the supported export formats as strings, for help text.
func formatNames() []string {
	all := exporter.All()
	names := make([]string, len(all))
	for i, f := range all {
		names[i] = string(f)
	}
	return names
}
