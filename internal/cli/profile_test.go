package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileAdd_Success(t *testing.T) {
	cfg := tempProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "profile", "add", "dev")
	require.NoError(t, err)
	assert.Equal(t, "Added profile 'dev'\n", stdout)
}

func TestProfileAdd_Duplicate(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "profile", "add", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestProfileAdd_InvalidName(t *testing.T) {
	cfg := tempProject(t)
	for _, name := range []string{"1bad", "has space", "bad!", "_lead"} {
		_, _, err := runCmd(t, "--config", cfg, "profile", "add", name)
		require.Errorf(t, err, "name %q should be rejected", name)
		assert.Contains(t, err.Error(), "invalid profile name")
	}
}

func TestProfileAdd_ExtendsValidParent(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "profile", "add", "dev")
	require.NoError(t, err)

	stdout, _, err := runCmd(t, "--config", cfg, "profile", "add", "prod", "--extends", "dev")
	require.NoError(t, err)
	assert.Equal(t, "Added profile 'prod' extending 'dev'\n", stdout)
}

func TestProfileAdd_ExtendsMissingParent(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "profile", "add", "prod", "--extends", "ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestProfileList_OnlyDefault(t *testing.T) {
	cfg := tempProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "profile", "list")
	require.NoError(t, err)

	want := "  NAME     VARS  EXTENDS\n" +
		"  default  0     -\n"
	assert.Equal(t, want, stdout)
}

func TestProfileList_MultipleProfiles(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "profile", "add", "dev")
	mustRun(t, cfg, "profile", "add", "prod", "--extends", "dev")
	mustRun(t, cfg, "set", "A", "1", "--profile", "dev")
	mustRun(t, cfg, "set", "B", "2", "--profile", "dev")

	stdout, _, err := runCmd(t, "--config", cfg, "profile", "list")
	require.NoError(t, err)

	want := "  NAME     VARS  EXTENDS\n" +
		"  default  0     -\n" +
		"  dev      2     -\n" +
		"  prod     0     dev\n"
	assert.Equal(t, want, stdout)
}

// mustRun runs a command against cfg and fails the test on error. It keeps
// multi-step test setup readable.
func mustRun(t *testing.T, cfg string, args ...string) {
	t.Helper()
	full := append([]string{"--config", cfg}, args...)
	if _, _, err := runCmd(t, full...); err != nil {
		t.Fatalf("setup command %v failed: %v", args, err)
	}
}
