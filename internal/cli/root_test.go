package cli_test

import (
	"os"
	"testing"

	"github.com/panic-at/envx/internal/cli"
	"github.com/panic-at/envx/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoot_NoArgsShowsHelp(t *testing.T) {
	stdout, _, err := runCmd(t)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available Commands")
	assert.Contains(t, stdout, "init")
	assert.Contains(t, stdout, "show")
}

func TestRoot_VersionFlag(t *testing.T) {
	stdout, _, err := runCmd(t, "--version")
	require.NoError(t, err)
	assert.Contains(t, stdout, "envx")
	assert.Contains(t, stdout, version.Version)
}

func TestRoot_UnknownCommand(t *testing.T) {
	_, _, err := runCmd(t, "definitely-not-a-command")
	require.Error(t, err)
}

func TestRoot_NoColorDefaultHonorsEnv(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		cmd := cli.NewRootCmd()
		assert.Equal(t, "true", cmd.PersistentFlags().Lookup("no-color").DefValue)
	})
	t.Run("unset", func(t *testing.T) {
		// t.Setenv registers restoration of the original value; removing it
		// afterwards exercises the unset branch of noColorDefault.
		t.Setenv("NO_COLOR", "placeholder")
		require.NoError(t, os.Unsetenv("NO_COLOR"))
		cmd := cli.NewRootCmd()
		assert.Equal(t, "false", cmd.PersistentFlags().Lookup("no-color").DefValue)
	})
}
