package cli_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/panic-at/envx/internal/cli"
)

// newTestCmd returns a fresh root command with stdout and stderr redirected to
// in-memory buffers for inspection.
func newTestCmd(t *testing.T) (cmd *cobra.Command, stdout, stderr *bytes.Buffer) {
	t.Helper()
	cmd = cli.NewRootCmd()
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd, stdout, stderr
}

// runCmd builds a fresh root command and executes it with args. Colored output
// is disabled so that captured output is deterministic. It returns the
// captured stdout, stderr and the execution error.
func runCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd, out, errBuf := newTestCmd(t)
	cmd.SetArgs(append([]string{"--no-color"}, args...))
	err = cmd.Execute()
	return out.String(), errBuf.String(), err
}

// tempProject creates an initialized envx project under a temporary directory
// and returns the path to its config file. Cleanup is automatic via t.TempDir.
func tempProject(t *testing.T) string {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), ".envx", "config.yaml")
	if _, _, err := runCmd(t, "--config", cfgPath, "init"); err != nil {
		t.Fatalf("tempProject: init failed: %v", err)
	}
	return cfgPath
}
