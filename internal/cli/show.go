package cli

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/mask"
	"github.com/panic-at/envx/internal/profile"
	"github.com/panic-at/envx/internal/resolver"
)

// newShowCmd builds the "envx show" command.
func newShowCmd(opts *rootOptions) *cobra.Command {
	var reveal bool

	cmd := &cobra.Command{
		Use:   "show <profile>",
		Short: "Show the effective variables of a profile",
		Long: "show prints the effective (extends-flattened) variables of a profile.\n" +
			"Without --reveal, refs are shown as their URIs and sensitive literals\n" +
			"are masked. With --reveal, refs are resolved; sensitive values are\n" +
			"still masked (a future --unsafe flag will reveal them in full).",
		Example: "  # Refs as URIs, sensitive literals masked\n" +
			"  envx show dev\n\n" +
			"  # Resolve refs against their vaults\n" +
			"  envx show dev --reveal",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			if _, ok := cfg.Profiles[name]; !ok {
				return fmt.Errorf("profile %q does not exist", name)
			}
			eff, err := cfg.Effective(name)
			if err != nil {
				return err
			}

			keys := make([]string, 0, len(eff.Vars))
			for k := range eff.Vars {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var resolved resolver.ResolveResult
			if reveal {
				resolved = resolver.ResolveAll(cmd.Context(), resolver.DefaultRegistry(), eff)
			}

			errStyle := color.New(color.FgRed)
			maskStyle := color.New(color.FgYellow)
			if opts.noColor {
				errStyle.DisableColor()
				maskStyle.DisableColor()
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 1, ' ', 0)
			for _, k := range keys {
				display := showValue(eff.Vars[k], k, reveal, resolved, errStyle, maskStyle)
				fmt.Fprintf(tw, "%s\t= %s\n", k, display)
			}
			if err := tw.Flush(); err != nil {
				return err
			}

			parent := cfg.Profiles[name].Extends
			if parent == "" {
				parent = "-"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d variables in profile '%s' (extends: %s)\n",
				len(keys), name, parent)
			return nil
		},
	}
	cmd.Flags().BoolVar(&reveal, "reveal", false, "resolve refs and show their values")
	return cmd
}

// showValue renders a single variable for the show command. The error and
// mask styles colorize the value column; coloring only the last column keeps
// the tabwriter alignment intact.
func showValue(v profile.Var, key string, reveal bool, resolved resolver.ResolveResult,
	errStyle, maskStyle *color.Color) string {
	if !reveal {
		switch {
		case v.Type == profile.VarRef:
			return v.URI
		case v.Sensitive:
			return maskStyle.Sprint(mask.Mask(v.Value))
		default:
			return v.Value
		}
	}

	if err, failed := resolved.Errors[key]; failed {
		return errStyle.Sprintf("<error: %s>", err)
	}
	value := resolved.Values[key]
	if v.Sensitive {
		return maskStyle.Sprint(mask.Mask(value))
	}
	return value
}
