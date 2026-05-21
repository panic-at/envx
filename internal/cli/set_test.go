package cli_test

import (
	"testing"

	"github.com/panic-at/envx/internal/config"
	"github.com/panic-at/envx/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_Literal(t *testing.T) {
	cfg := tempProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "set", "FOO", "bar", "--profile", "default")
	require.NoError(t, err)
	assert.Equal(t, "Set FOO in profile 'default'\n", stdout)

	loaded, err := config.Load(cfg)
	require.NoError(t, err)
	v := loaded.Profiles["default"].Vars["FOO"]
	assert.Equal(t, profile.VarLiteral, v.Type)
	assert.Equal(t, "bar", v.Value)
}

func TestSet_Ref(t *testing.T) {
	cfg := tempProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "set", "API_KEY", "--ref", "env://HOME", "--profile", "default")
	require.NoError(t, err)
	assert.Equal(t, "Set API_KEY (ref) in profile 'default'\n", stdout)

	loaded, err := config.Load(cfg)
	require.NoError(t, err)
	v := loaded.Profiles["default"].Vars["API_KEY"]
	assert.Equal(t, profile.VarRef, v.Type)
	assert.Equal(t, "env://HOME", v.URI)
}

func TestSet_Sensitive(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "PWD_TOKEN", "s3cret", "--profile", "default", "--sensitive")
	require.NoError(t, err)

	loaded, err := config.Load(cfg)
	require.NoError(t, err)
	assert.True(t, loaded.Profiles["default"].Vars["PWD_TOKEN"].Sensitive)
}

func TestSet_RefUnknownScheme(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "X", "--ref", "weird://a/b", "--profile", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown resolver scheme")
}

func TestSet_RefMissingScheme(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "X", "--ref", "no-scheme-here", "--profile", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing scheme")
}

func TestSet_InvalidKey(t *testing.T) {
	cfg := tempProject(t)
	for _, key := range []string{"lower", "1LEAD", "has-dash", "WITH SPACE"} {
		_, _, err := runCmd(t, "--config", cfg, "set", key, "v", "--profile", "default")
		require.Errorf(t, err, "key %q should be rejected", key)
		assert.Contains(t, err.Error(), "invalid variable name")
	}
}

func TestSet_ProfileNotFound(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "FOO", "bar", "--profile", "ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestSet_ValueAndRefTogether(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "FOO", "bar", "--ref", "env://X", "--profile", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "both")
}

func TestSet_NeitherValueNorRef(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "FOO", "--profile", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "literal value or --ref")
}

func TestSet_RequiredProfileFlagMissing(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "set", "FOO", "bar")
	require.Error(t, err)
}

func TestSet_Overwrite(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "set", "FOO", "one", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "set", "FOO", "two", "--profile", "default")
	require.NoError(t, err)
	assert.Equal(t, "Set FOO in profile 'default'\n", stdout)

	loaded, err := config.Load(cfg)
	require.NoError(t, err)
	assert.Equal(t, "two", loaded.Profiles["default"].Vars["FOO"].Value)
}
