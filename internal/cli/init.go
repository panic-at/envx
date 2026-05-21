package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/profile"
)

// defaultProfileName is the name of the empty profile created by envx init.
const defaultProfileName = "default"

// newInitCmd builds the "envx init" command.
func newInitCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new envx project in the current directory",
		Long: "init creates the .envx directory and a starter config.yaml " +
			"containing a single empty profile named 'default'.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := opts.configPath

			switch _, err := os.Stat(path); {
			case err == nil:
				// TODO: support a --force flag to overwrite an existing config.
				return fmt.Errorf("config already exists at %s", path)
			case !errors.Is(err, fs.ErrNotExist):
				return fmt.Errorf("check config path %s: %w", path, err)
			}

			cfg := config.New()
			cfg.Profiles[defaultProfileName] = profile.Profile{}
			if err := cfg.Save(path); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized envx config at %s\n", path)
			return nil
		},
	}
}
