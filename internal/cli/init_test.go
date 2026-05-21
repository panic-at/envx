package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/panic-at/envx/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesConfig(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), ".envx", "config.yaml")

	stdout, _, err := runCmd(t, "--config", cfgPath, "init")
	require.NoError(t, err)
	assert.Equal(t, "Initialized envx config at "+cfgPath+"\n", stdout)
	assert.FileExists(t, cfgPath)
}

func TestInit_GeneratedYAML(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), ".envx", "config.yaml")
	_, _, err := runCmd(t, "--config", cfgPath, "init")
	require.NoError(t, err)

	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version: 1")
	assert.Contains(t, string(data), "default:")

	// The generated file must be loadable and valid per the config schema.
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, config.CurrentVersion, cfg.Version)
	assert.Contains(t, cfg.Profiles, "default")
}

func TestInit_FailsIfExists(t *testing.T) {
	cfgPath := tempProject(t)

	_, _, err := runCmd(t, "--config", cfgPath, "init")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
