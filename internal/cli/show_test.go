package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupShowProject builds a project with one profile holding a plain literal,
// a sensitive literal and a ref, for the show tests to render.
func setupShowProject(t *testing.T) string {
	t.Helper()
	cfg := tempProject(t)
	mustRun(t, cfg, "profile", "add", "dev")
	mustRun(t, cfg, "set", "FOO", "bar", "--profile", "dev")
	mustRun(t, cfg, "set", "SECRET", "hunter2", "--profile", "dev", "--sensitive")
	mustRun(t, cfg, "set", "REF", "--ref", "op://v/i/f", "--profile", "dev")
	return cfg
}

func TestShow_WithoutReveal(t *testing.T) {
	cfg := setupShowProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "show", "dev")
	require.NoError(t, err)

	want := "FOO    = bar\n" +
		"REF    = op://v/i/f\n" +
		"SECRET = h***\n" +
		"3 variables in profile 'dev' (extends: -)\n"
	assert.Equal(t, want, stdout)
}

func TestShow_WithReveal(t *testing.T) {
	cfg := setupShowProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "show", "dev", "--reveal")
	require.NoError(t, err)

	assert.Contains(t, stdout, "FOO    = bar")
	// A 1Password ref is parsed but not yet resolvable.
	assert.Contains(t, stdout, "REF    = <error:")
	assert.Contains(t, stdout, "not implemented")
	// Sensitive values stay masked even with --reveal.
	assert.Contains(t, stdout, "SECRET = h***")
}

func TestShow_RevealResolvesEnvRef(t *testing.T) {
	cfg := tempProject(t)
	t.Setenv("ENVX_SHOW_TEST_VAR", "resolved-value")
	mustRun(t, cfg, "set", "V", "--ref", "env://ENVX_SHOW_TEST_VAR", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "show", "default", "--reveal")
	require.NoError(t, err)
	assert.Contains(t, stdout, "V = resolved-value")
}

func TestShow_RevealResolutionErrorInline(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "set", "MISSING", "--ref", "env://ENVX_DEFINITELY_UNSET_VAR_XZ", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "show", "default", "--reveal")
	require.NoError(t, err, "a resolution error must not fail the command")
	assert.Contains(t, stdout, "MISSING = <error:")
}

func TestShow_ProfileNotFound(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "show", "ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestShow_ExtendsFooterAndInheritedVars(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "profile", "add", "dev")
	mustRun(t, cfg, "set", "BASE", "base-value", "--profile", "dev")
	mustRun(t, cfg, "profile", "add", "prod", "--extends", "dev")
	mustRun(t, cfg, "set", "OWN", "own-value", "--profile", "prod")

	stdout, _, err := runCmd(t, "--config", cfg, "show", "prod")
	require.NoError(t, err)

	want := "BASE = base-value\n" +
		"OWN  = own-value\n" +
		"2 variables in profile 'prod' (extends: dev)\n"
	assert.Equal(t, want, stdout)
}

func TestShow_EmptyProfile(t *testing.T) {
	cfg := tempProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "show", "default")
	require.NoError(t, err)
	assert.Equal(t, "0 variables in profile 'default' (extends: -)\n", stdout)
}
