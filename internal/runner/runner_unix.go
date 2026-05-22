//go:build !windows

package runner

import (
	"os"
	"syscall"
)

// forwardSignal relays sig to the child process. On Unix every supported
// signal can be delivered directly.
func forwardSignal(proc *os.Process, sig os.Signal) error {
	return proc.Signal(sig)
}

// signalExitCode reports the 128+signum exit code of a process terminated by a
// signal, following shell convention. It returns ok false for a process that
// exited normally, leaving the caller to use ProcessState.ExitCode.
func signalExitCode(state *os.ProcessState) (int, bool) {
	ws, ok := state.Sys().(syscall.WaitStatus)
	if !ok || !ws.Signaled() {
		return 0, false
	}
	return 128 + int(ws.Signal()), true
}
