// Command helper is a tiny, cross-platform test fixture for the envx runner.
//
// It is never installed or shipped: the runner and CLI test suites compile it
// on demand (see the TestMain functions) and exec it as the target command of
// "envx run". Its subcommands let tests observe the injected environment and
// exercise exit codes, cancellation and signal handling.
//
// Usage:
//
//	helper print KEY [KEY...]   print "KEY=value" for each environment variable
//	helper exit N               exit immediately with status N
//	helper sleep N              sleep N seconds, then exit 0
//	helper sigwait              await SIGINT/SIGTERM, report it, exit 0
//	helper sigignore            install a no-op signal handler, then block
//	helper touch PATH           create a marker file at PATH, then exit 0
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run dispatches a subcommand and returns the process exit code. Keeping the
// logic out of main makes every branch reachable without spawning a process.
func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "helper: missing subcommand")
		return 2
	}
	switch args[0] {
	case "print":
		for _, name := range args[1:] {
			fmt.Printf("%s=%s\n", name, os.Getenv(name))
		}
		return 0
	case "exit":
		return cmdExit(args[1:])
	case "sleep":
		return cmdSleep(args[1:])
	case "sigwait":
		return cmdSigwait()
	case "sigignore":
		return cmdSigignore()
	case "touch":
		return cmdTouch(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "helper: unknown subcommand %q\n", args[0])
		return 2
	}
}

// cmdExit exits with the status code given as its single argument.
func cmdExit(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "helper exit: want exactly one argument")
		return 2
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper exit: %v\n", err)
		return 2
	}
	return n
}

// cmdSleep sleeps for the given number of seconds; tests cancel it early to
// exercise context cancellation and signal forwarding.
func cmdSleep(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "helper sleep: want exactly one argument")
		return 2
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper sleep: %v\n", err)
		return 2
	}
	time.Sleep(time.Duration(n) * time.Second)
	return 0
}

// cmdSigwait installs a SIGINT/SIGTERM handler, signals readiness on stdout and
// blocks until a signal arrives, which it reports before exiting cleanly.
func cmdSigwait() int {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	fmt.Println("ready")
	sig := <-ch
	fmt.Printf("got signal: %s\n", sig)
	return 0
}

// cmdSigignore installs a handler that swallows SIGINT/SIGTERM and then blocks
// forever, forcing the runner to fall back to SIGKILL after its grace window.
func cmdSigignore() int {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	fmt.Println("ready")
	for {
		<-ch
	}
}

// cmdTouch creates a marker file, letting tests prove whether the command ran.
func cmdTouch(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "helper touch: want exactly one argument")
		return 2
	}
	if err := os.WriteFile(args[0], []byte("marker"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "helper touch: %v\n", err)
		return 1
	}
	return 0
}
