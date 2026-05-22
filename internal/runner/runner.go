// Package runner executes a command as a child process with a profile's
// resolved variables injected into its environment.
//
// The injected variables live only in the child's environment: runner never
// calls os.Setenv, so they never leak into the envx process or the user's
// shell. The runner connects the child to the supplied stdio, forwards
// interrupt signals to it, enforces a grace window before escalating to
// SIGKILL, and propagates the child's exit code to the caller.
package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// defaultGrace is how long Run waits after forwarding a termination signal to
// the child before escalating to an unconditional kill.
const defaultGrace = 5 * time.Second

// Options configures a single Run invocation.
type Options struct {
	// Command is the executable followed by its arguments. Command[0] is
	// looked up on PATH. It must not be empty.
	Command []string
	// Vars holds the profile's resolved variables to inject, keyed by name.
	Vars map[string]string
	// HostEnv is the host environment in os.Environ form. A nil value means
	// "use os.Environ()"; pass a non-nil (possibly empty) slice to control
	// inheritance deterministically, e.g. in tests.
	HostEnv []string
	// Inherit includes the host environment in the child's environment.
	Inherit bool
	// Override lets profile variables win key collisions with the host
	// environment; it has no effect when Inherit is false.
	Override bool
	// Stdin, Stdout and Stderr are connected directly to the child. Passing
	// the real *os.File values yields true TTY inheritance; tests pass
	// buffers instead. A nil stream behaves as exec.Cmd documents.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// Grace overrides the wait between forwarding a termination signal and
	// killing the child. A non-positive value selects defaultGrace.
	Grace time.Duration
}

// Run executes opts.Command as a child process and returns its exit code.
//
// The returned error is non-nil only when the command could not be started;
// in that case the exit code follows shell convention (127 not found, 126 not
// executable, 1 otherwise). Once the child has started, Run always returns a
// nil error and the child's exit code — including when the child is terminated
// by a signal, in which case the code is 128+signum.
//
// Run forwards SIGINT and SIGTERM received by the envx process to the child
// rather than dying immediately, giving it a grace window to exit before
// sending an unconditional kill. Cancelling ctx kills the child.
func Run(ctx context.Context, opts Options) (int, error) {
	if len(opts.Command) == 0 {
		return 2, errors.New("runner: no command given")
	}

	host := opts.HostEnv
	if host == nil {
		host = os.Environ()
	}

	grace := opts.Grace
	if grace <= 0 {
		grace = defaultGrace
	}

	cmd := exec.CommandContext(ctx, opts.Command[0], opts.Command[1:]...)
	cmd.Env = BuildEnv(host, opts.Vars, opts.Inherit, opts.Override)
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	// Register the signal handler before Start so that a signal arriving the
	// instant the child appears is never missed.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	if err := cmd.Start(); err != nil {
		return startExitCode(err), startError(err, opts.Command[0])
	}

	// done closes when the child has been waited on; the signal goroutine
	// uses it to stop without leaking.
	done := make(chan struct{})
	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		forwardSignals(cmd.Process, sigCh, grace, done)
	}()

	waitErr := cmd.Wait()
	close(done)
	<-loopDone

	return waitExitCode(waitErr)
}

// forwardSignals relays termination signals from sigCh to the child process.
// After forwarding the first signal it starts the grace timer; if the child
// has not been reaped (done not closed) when the timer fires, it kills the
// child outright. It returns once done is closed.
func forwardSignals(proc *os.Process, sigCh <-chan os.Signal, grace time.Duration, done <-chan struct{}) {
	select {
	case <-done:
		return
	case sig := <-sigCh:
		_ = forwardSignal(proc, sig)
		select {
		case <-done:
		case <-time.After(grace):
			_ = proc.Kill()
		}
	}
}

// startNotFound reports whether err from cmd.Start means the command could not
// be located — either a PATH lookup miss or a path to a missing file.
func startNotFound(err error) bool {
	return errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist)
}

// startError turns a cmd.Start failure into a caller-facing error, recognising
// the common "not found" and "not executable" cases.
func startError(err error, name string) error {
	switch {
	case startNotFound(err):
		return fmt.Errorf("command not found: %s", name)
	case errors.Is(err, fs.ErrPermission):
		return fmt.Errorf("permission denied: %s", name)
	default:
		return fmt.Errorf("start command %s: %w", name, err)
	}
}

// startExitCode maps a cmd.Start failure to a shell-convention exit code: 127
// for a missing command, 126 for one that exists but is not executable, and 1
// for anything else.
func startExitCode(err error) int {
	switch {
	case startNotFound(err):
		return 127
	case errors.Is(err, fs.ErrPermission):
		return 126
	default:
		return 1
	}
}

// waitExitCode extracts the child's exit code from the error returned by
// cmd.Wait. A child terminated by a signal yields 128+signum. The returned
// error is non-nil only for a Wait failure that is not an ordinary exit.
func waitExitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if code, ok := signalExitCode(exitErr.ProcessState); ok {
			return code, nil
		}
		return exitErr.ProcessState.ExitCode(), nil
	}
	return 1, fmt.Errorf("wait for command: %w", err)
}
