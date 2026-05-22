package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperBin is the path to the compiled test helper, built once by TestMain.
var helperBin string

// TestMain compiles testdata/helper into a temporary binary that the
// integration tests exec as the target command of Run.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "envx-runner-helper")
	if err != nil {
		fmt.Fprintln(os.Stderr, "runner test: create temp dir:", err)
		os.Exit(1)
	}
	helperBin = filepath.Join(dir, "helper")
	if runtime.GOOS == "windows" {
		helperBin += ".exe"
	}
	build := exec.Command("go", "build", "-o", helperBin, "../../testdata/helper")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "runner test: build helper: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// syncBuffer is a bytes.Buffer guarded by a mutex so a test can read a child's
// output while Run is still writing to it from another goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// runResult bundles the two return values of Run for delivery over a channel.
type runResult struct {
	code int
	err  error
}

func TestRun_InjectsProfileVariables(t *testing.T) {
	var out bytes.Buffer
	code, err := Run(context.Background(), Options{
		Command: []string{helperBin, "print", "FOO"},
		Vars:    map[string]string{"FOO": "bar"},
		HostEnv: []string{},
		Inherit: true,
		Stdout:  &out,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "FOO=bar\n", out.String())
}

func TestRun_DoesNotLeakIntoParentEnvironment(t *testing.T) {
	const key = "ENVX_RUNNER_ISOLATION_PROBE"
	require.Empty(t, os.Getenv(key), "test precondition: probe variable must be unset")

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{
		Command: []string{helperBin, "print", key},
		Vars:    map[string]string{key: "secret"},
		HostEnv: []string{},
		Inherit: true,
		Stdout:  &out,
	})
	require.NoError(t, err)

	assert.Equal(t, key+"=secret\n", out.String(), "the child receives the variable")
	assert.Empty(t, os.Getenv(key), "the variable must not leak into the envx process")
}

func TestRun_InheritsHostEnvironment(t *testing.T) {
	t.Setenv("ENVX_RUNNER_HOST_VAR", "from-host")

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{
		Command: []string{helperBin, "print", "ENVX_RUNNER_HOST_VAR"},
		Inherit: true,
		Stdout:  &out,
	})
	require.NoError(t, err)
	assert.Equal(t, "ENVX_RUNNER_HOST_VAR=from-host\n", out.String())
}

func TestRun_NoInheritHidesHostEnvironment(t *testing.T) {
	t.Setenv("ENVX_RUNNER_HOST_VAR", "from-host")

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{
		Command: []string{helperBin, "print", "ENVX_RUNNER_HOST_VAR"},
		Inherit: false,
		Stdout:  &out,
	})
	require.NoError(t, err)
	assert.Equal(t, "ENVX_RUNNER_HOST_VAR=\n", out.String(),
		"with --no-inherit the host variable must not reach the child")
}

func TestRun_OverrideLetsProfileWin(t *testing.T) {
	t.Setenv("ENVX_RUNNER_SHARED", "host-value")

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{
		Command:  []string{helperBin, "print", "ENVX_RUNNER_SHARED"},
		Vars:     map[string]string{"ENVX_RUNNER_SHARED": "profile-value"},
		Inherit:  true,
		Override: true,
		Stdout:   &out,
	})
	require.NoError(t, err)
	assert.Equal(t, "ENVX_RUNNER_SHARED=profile-value\n", out.String())
}

func TestRun_NoOverrideLetsHostWin(t *testing.T) {
	t.Setenv("ENVX_RUNNER_SHARED", "host-value")

	var out bytes.Buffer
	_, err := Run(context.Background(), Options{
		Command:  []string{helperBin, "print", "ENVX_RUNNER_SHARED"},
		Vars:     map[string]string{"ENVX_RUNNER_SHARED": "profile-value"},
		Inherit:  true,
		Override: false,
		Stdout:   &out,
	})
	require.NoError(t, err)
	assert.Equal(t, "ENVX_RUNNER_SHARED=host-value\n", out.String())
}

func TestRun_PropagatesExitCode(t *testing.T) {
	code, err := Run(context.Background(), Options{
		Command: []string{helperBin, "exit", "3"},
		HostEnv: []string{},
	})
	require.NoError(t, err, "a child that ran is not a Run error, whatever its code")
	assert.Equal(t, 3, code)
}

func TestRun_NoCommand(t *testing.T) {
	code, err := Run(context.Background(), Options{})
	require.Error(t, err)
	assert.Equal(t, 2, code)
}

func TestRun_CommandNotFound(t *testing.T) {
	code, err := Run(context.Background(), Options{
		Command: []string{"envx-no-such-command-zzz"},
		HostEnv: []string{},
	})
	require.Error(t, err)
	assert.Equal(t, 127, code, "a missing command exits 127 by shell convention")
	assert.Contains(t, err.Error(), "command not found")
}

func TestRun_ContextCancellationKillsChild(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	code, err := Run(ctx, Options{
		Command: []string{helperBin, "sleep", "10"},
		HostEnv: []string{},
	})
	elapsed := time.Since(start)

	require.NoError(t, err, "a child that started returns no Run error even when cancelled")
	assert.NotEqual(t, 0, code, "a killed child does not exit cleanly")
	assert.Less(t, elapsed, 5*time.Second,
		"the child must be killed when ctx expires, not left to sleep 10s")
}

func TestRun_ForwardsSignalToChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX signal delivery between processes is not available on Windows")
	}
	out := &syncBuffer{}
	resCh := make(chan runResult, 1)
	go func() {
		code, err := Run(context.Background(), Options{
			Command: []string{helperBin, "sigwait"},
			HostEnv: []string{},
			Stdout:  out,
		})
		resCh <- runResult{code, err}
	}()

	// The helper prints "ready" only after installing its own handler; by
	// then Run has already registered its signal.Notify (it does so before
	// starting the child), so the signal cannot be missed.
	require.Eventually(t, func() bool {
		return strings.Contains(out.String(), "ready")
	}, 5*time.Second, 10*time.Millisecond, "helper never became ready")

	signalSelf(t)

	select {
	case res := <-resCh:
		require.NoError(t, res.err)
		assert.Equal(t, 0, res.code, "the helper exits 0 after handling the forwarded signal")
		assert.Contains(t, out.String(), "got signal: interrupt",
			"the child must receive the signal envx forwarded")
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after the signal was forwarded")
	}
}

func TestRun_KillsChildAfterGraceWindow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX signal delivery between processes is not available on Windows")
	}
	out := &syncBuffer{}
	resCh := make(chan runResult, 1)
	start := time.Now()
	go func() {
		code, err := Run(context.Background(), Options{
			Command: []string{helperBin, "sigignore"},
			HostEnv: []string{},
			Stdout:  out,
			Grace:   300 * time.Millisecond,
		})
		resCh <- runResult{code, err}
	}()

	require.Eventually(t, func() bool {
		return strings.Contains(out.String(), "ready")
	}, 5*time.Second, 10*time.Millisecond, "helper never became ready")

	signalSelf(t)

	select {
	case res := <-resCh:
		require.NoError(t, res.err)
		assert.Equal(t, 137, res.code,
			"a child that ignores the signal is SIGKILLed, reporting 128+9")
		assert.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond,
			"the child must be given its full grace window before the kill")
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after the grace window expired")
	}
}

// signalSelf sends SIGINT to the test process. Run's signal handler, active
// for the duration of the call, intercepts it and forwards it to the child, so
// the test process itself is never terminated.
func signalSelf(t *testing.T) {
	t.Helper()
	self, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, self.Signal(os.Interrupt))
}
