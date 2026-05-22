package cli

import (
	"fmt"
	"regexp"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/profile"
)

// profileNamePattern is the CLI-level rule for new profile names: a leading
// letter followed by letters, digits, hyphens or underscores. It is stricter
// than the config schema, which the saved file is also validated against.
const profileNamePattern = `^[a-zA-Z][a-zA-Z0-9_-]*$`

var profileNameRe = regexp.MustCompile(profileNamePattern)

// newProfileCmd builds the "envx profile" command group.
func newProfileCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles",
		Long: "profile groups the subcommands that create and list profiles.\n" +
			"Run it with a subcommand: 'profile add' or 'profile list'.",
		Example: "  envx profile add staging\n" +
			"  envx profile add prod --extends staging\n" +
			"  envx profile list",
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newProfileAddCmd(opts), newProfileListCmd(opts))
	return cmd
}

// newProfileAddCmd builds the "envx profile add" command.
func newProfileAddCmd(opts *rootOptions) *cobra.Command {
	var extends string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new, empty profile",
		Long: "add creates a new, empty profile. With --extends the profile\n" +
			"inherits every variable of an existing parent profile.",
		Example: "  # A standalone profile\n" +
			"  envx profile add dev\n\n" +
			"  # A profile that inherits from another\n" +
			"  envx profile add prod --extends dev",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if !profileNameRe.MatchString(name) {
				return fmt.Errorf("invalid profile name %q (must match %s)", name, profileNamePattern)
			}

			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			if _, exists := cfg.Profiles[name]; exists {
				return fmt.Errorf("profile %q already exists", name)
			}
			if extends != "" {
				if _, ok := cfg.Profiles[extends]; !ok {
					return fmt.Errorf("parent profile %q does not exist", extends)
				}
			}

			cfg.Profiles[name] = profile.Profile{Extends: extends}
			if err := cfg.Save(opts.configPath); err != nil {
				return err
			}

			if extends != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Added profile '%s' extending '%s'\n", name, extends)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Added profile '%s'\n", name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&extends, "extends", "", "parent profile to inherit from")
	return cmd
}

// newProfileListCmd builds the "envx profile list" command.
func newProfileListCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		Long: "list prints every profile with its variable count and the parent\n" +
			"profile it extends, if any.",
		Example: "  envx profile list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}

			names := make([]string, 0, len(cfg.Profiles))
			for name := range cfg.Profiles {
				names = append(names, name)
			}
			sort.Strings(names)

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "  NAME\tVARS\tEXTENDS")
			for _, name := range names {
				p := cfg.Profiles[name]
				extends := p.Extends
				if extends == "" {
					extends = "-"
				}
				fmt.Fprintf(tw, "  %s\t%d\t%s\n", name, len(p.Vars), extends)
			}
			return tw.Flush()
		},
	}
}
