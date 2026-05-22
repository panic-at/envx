//go:build windows

package runner

import "os"

// forwardSignal relays sig to the child process.
//
// Windows has no POSIX signal delivery: a process cannot be sent SIGINT or
// SIGTERM by another process the way it can on Unix. The only action the Go
// runtime can guarantee here is an unconditional termination, so any
// termination signal is mapped to a kill. Graceful Ctrl+C handling for child
// processes on Windows would require console control events and is out of
// scope for the MVP.
func forwardSignal(proc *os.Process, _ os.Signal) error {
	return proc.Kill()
}

// signalExitCode always reports ok false on Windows: process exit there does
// not carry a Unix signal number, so the caller falls back to
// ProcessState.ExitCode.
func signalExitCode(_ *os.ProcessState) (int, bool) {
	return 0, false
}
