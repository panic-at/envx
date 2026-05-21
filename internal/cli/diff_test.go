package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/panic-at/envx/internal/cli"
	"github.com/panic-at/envx/internal/mask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildDiffProject creates a project with profiles "a" and "b" whose effective
// variables differ by one added, one removed and one changed key.
func buildDiffProject(t *testing.T) string {
	t.Helper()
	cfg := tempProject(t)
	mustRun(t, cfg, "profile", "add", "a")
	mustRun(t, cfg, "profile", "add", "b")
	mustRun(t, cfg, "set", "ALPHA", "1", "--profile", "a")
	mustRun(t, cfg, "set", "BETA", "2", "--profile", "a")
	mustRun(t, cfg, "set", "GAMMA", "3", "--profile", "a")
	mustRun(t, cfg, "set", "BETA", "2", "--profile", "b")
	mustRun(t, cfg, "set", "GAMMA", "changed", "--profile", "b")
	mustRun(t, cfg, "set", "DELTA", "4", "--profile", "b")
	return cfg
}

func TestDiff_BasicFixture(t *testing.T) {
	cfg := buildDiffProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "diff", "a", "b")
	require.NoError(t, err)

	want, err := os.ReadFile(filepath.Join("testdata", "expected", "diff_basic.txt"))
	require.NoError(t, err)
	assert.Equal(t, string(want), stdout)
}

func TestDiff_IdenticalProfiles(t *testing.T) {
	cfg := tempProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "diff", "default", "default")
	require.NoError(t, err)
	assert.Equal(t, "0 added, 0 removed, 0 changed\n", stdout)
}

func TestDiff_SensitiveMasked(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "profile", "add", "sec")
	mustRun(t, cfg, "set", "SECRET", "supersecret", "--profile", "sec", "--sensitive")

	stdout, _, err := runCmd(t, "--config", cfg, "diff", "default", "sec")
	require.NoError(t, err)
	assert.Contains(t, stdout, "+ SECRET = "+mask.Mask("supersecret"))
	assert.NotContains(t, stdout, "supersecret", "the raw sensitive value must not leak")
}

func TestDiff_ExtendsFlattenedBeforeDiff(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "profile", "add", "base")
	mustRun(t, cfg, "set", "BASE_VAR", "1", "--profile", "base")
	mustRun(t, cfg, "profile", "add", "child", "--extends", "base")
	mustRun(t, cfg, "set", "CHILD_VAR", "2", "--profile", "child")

	stdout, _, err := runCmd(t, "--config", cfg, "diff", "base", "child")
	require.NoError(t, err)
	// child inherits BASE_VAR through extends, so it must not appear as removed.
	assert.NotContains(t, stdout, "BASE_VAR")
	assert.Contains(t, stdout, "+ CHILD_VAR = 2")
	assert.Contains(t, stdout, "1 added, 0 removed, 0 changed")
}

func TestDiff_ProfileNotFound(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "diff", "ghost", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"ghost"`)

	_, _, err = runCmd(t, "--config", cfg, "diff", "default", "phantom")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"phantom"`)
}

func TestDiff_JSONFormat(t *testing.T) {
	cfg := buildDiffProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "diff", "a", "b", "--format", "json")
	require.NoError(t, err)

	var doc struct {
		Added   int `json:"added"`
		Removed int `json:"removed"`
		Changed int `json:"changed"`
		Changes []struct {
			Key  string `json:"key"`
			Kind string `json:"kind"`
		} `json:"changes"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &doc))
	assert.Equal(t, 1, doc.Added)
	assert.Equal(t, 1, doc.Removed)
	assert.Equal(t, 1, doc.Changed)
	require.Len(t, doc.Changes, 3)
	// Changes stay ordered by key.
	assert.Equal(t, "ALPHA", doc.Changes[0].Key)
	assert.Equal(t, "removed", doc.Changes[0].Kind)
}

func TestDiff_NoColorHasNoANSI(t *testing.T) {
	cfg := buildDiffProject(t)
	stdout, _, err := runCmd(t, "--config", cfg, "diff", "a", "b")
	require.NoError(t, err)

	ansi := regexp.MustCompile("\x1b\\[[0-9;]*m")
	assert.NotRegexp(t, ansi, stdout, "--no-color output must contain no ANSI escape codes")
}

func TestDiff_InvalidFormat(t *testing.T) {
	cfg := tempProject(t)
	_, _, err := runCmd(t, "--config", cfg, "diff", "default", "default", "--format", "xml")
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 2, exit.Code)
}
