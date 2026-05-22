package cli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/panic-at/envx/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperBin is the path to the compiled test helper, built once by TestMain
// and used as the target command of "envx run" in these tests.
var helperBin string

// TestMain compiles testdata/helper before running the cli test suite.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "envx-cli-helper")
	if err != nil {
		fmt.Fprintln(os.Stderr, "cli test: create temp dir:", err)
		os.Exit(1)
	}
	helperBin = filepath.Join(dir, "helper")
	if runtime.GOOS == "windows" {
		helperBin += ".exe"
	}
	build := exec.Command("go", "build", "-o", helperBin, "../../testdata/helper")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "cli test: build helper: %v\n%s", err, out)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func TestRun_InjectsProfileVariables(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "set", "PORT", "8080", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "print", "PORT")
	require.NoError(t, err)
	assert.Equal(t, "PORT=8080\n", stdout)
}

func TestRun_DoesNotLeakIntoShell(t *testing.T) {
	const key = "ENVX_CLI_ISOLATION_PROBE"
	require.Empty(t, os.Getenv(key), "test precondition: probe variable must be unset")

	cfg := tempProject(t)
	mustRun(t, cfg, "set", key, "secret", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "print", key)
	require.NoError(t, err)
	assert.Equal(t, key+"=secret\n", stdout, "the child receives the variable")
	assert.Empty(t, os.Getenv(key), "the variable must not leak into the envx process")
}

func TestRun_PropagatesExitCode(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "exit", "3")
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 3, exit.Code, "the child's exit code becomes envx's exit code")
	assert.True(t, exit.Silent, "a non-zero child exit is propagated without an extra envx message")
}

func TestRun_SucceedsWithZeroExitCode(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "exit", "0")
	require.NoError(t, err, "a child exiting 0 must not surface as an error")
}

func TestRun_CommandNotFound(t *testing.T) {
	cfg := tempProject(t)
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default", "--", missing)
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 127, exit.Code, "a missing command exits 127 by shell convention")
}

func TestRun_MissingDashSeparator(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		helperBin, "print", "PORT")
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 2, exit.Code, "omitting '--' is a usage error")
}

func TestRun_NoCommandAfterDash(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default", "--")
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 2, exit.Code, "an empty command after '--' is a usage error")
}

func TestRun_MissingProfileFlag(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--", helperBin, "print", "PORT")
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 2, exit.Code, "a missing required flag is a usage error")
}

func TestRun_ProfileNotFound(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "ghost",
		"--", helperBin, "print", "PORT")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestRun_InheritsHostByDefault(t *testing.T) {
	t.Setenv("ENVX_CLI_HOST_VAR", "from-host")
	cfg := tempProject(t)

	stdout, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "print", "ENVX_CLI_HOST_VAR")
	require.NoError(t, err)
	assert.Equal(t, "ENVX_CLI_HOST_VAR=from-host\n", stdout)
}

func TestRun_NoInheritIsolatesEnvironment(t *testing.T) {
	t.Setenv("ENVX_CLI_HOST_VAR", "from-host")
	cfg := tempProject(t)

	stdout, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--no-inherit", "--", helperBin, "print", "ENVX_CLI_HOST_VAR")
	require.NoError(t, err)
	assert.Equal(t, "ENVX_CLI_HOST_VAR=\n", stdout,
		"with --no-inherit the host variable must not reach the child")
}

func TestRun_OverrideLetsProfileWin(t *testing.T) {
	t.Setenv("ENVX_CLI_SHARED", "host-value")
	cfg := tempProject(t)
	mustRun(t, cfg, "set", "ENVX_CLI_SHARED", "profile-value", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "print", "ENVX_CLI_SHARED")
	require.NoError(t, err)
	assert.Equal(t, "ENVX_CLI_SHARED=profile-value\n", stdout,
		"--override (the default) lets the profile value win")
}

func TestRun_NoOverrideLetsHostWin(t *testing.T) {
	t.Setenv("ENVX_CLI_SHARED", "host-value")
	cfg := tempProject(t)
	mustRun(t, cfg, "set", "ENVX_CLI_SHARED", "profile-value", "--profile", "default")

	stdout, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--no-override", "--", helperBin, "print", "ENVX_CLI_SHARED")
	require.NoError(t, err)
	assert.Equal(t, "ENVX_CLI_SHARED=host-value\n", stdout,
		"--no-override lets the host value win")
}

func TestRun_MutuallyExclusiveToggles(t *testing.T) {
	cfg := tempProject(t)

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--inherit", "--no-inherit", "--", helperBin, "print", "PORT")
	require.Error(t, err)

	var exit *cli.ExitError
	require.ErrorAs(t, err, &exit)
	assert.Equal(t, 2, exit.Code, "passing both --inherit and --no-inherit is a usage error")
}

func TestRun_ResolutionFailsWithoutAllowErrors(t *testing.T) {
	cfg := tempProject(t)
	// op:// is parsed but not yet resolvable, so it fails resolution.
	mustRun(t, cfg, "set", "BAD", "--ref", "op://v/i/f", "--profile", "default")
	marker := filepath.Join(t.TempDir(), "marker")

	_, _, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--", helperBin, "touch", marker)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve")
	assert.NoFileExists(t, marker, "the command must not run when resolution fails")
}

func TestRun_ResolutionFailsWithAllowErrors(t *testing.T) {
	cfg := tempProject(t)
	mustRun(t, cfg, "set", "GOOD", "ok", "--profile", "default")
	mustRun(t, cfg, "set", "BAD", "--ref", "op://v/i/f", "--profile", "default")

	stdout, stderr, err := runCmd(t, "--config", cfg, "run", "--profile", "default",
		"--allow-errors", "--", helperBin, "print", "GOOD", "BAD")
	require.NoError(t, err)
	assert.Contains(t, stdout, "GOOD=ok", "resolved variables still reach the child")
	assert.Contains(t, stdout, "BAD=\n", "the unresolved variable is omitted, so it reads empty")
	assert.Contains(t, stderr, "warning")
	assert.Contains(t, stderr, "BAD")
}
