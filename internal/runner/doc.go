// Package runner executes a command as a child process with a profile's
// resolved variables injected into its environment.
//
// The injected variables live only in the child's environment: runner never
// calls os.Setenv, so they never leak into the envx process or the user's
// shell. Run connects the child to the supplied stdio, forwards interrupt
// signals to it, enforces a grace window before escalating to SIGKILL, and
// propagates the child's exit code to the caller.
//
// BuildEnv is the pure function that merges the host environment with the
// profile's variables according to the inherit and override toggles.
package runner
