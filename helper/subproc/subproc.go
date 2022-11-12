package subproc

import (
	"os"
	"time"
)

const (
	// ExitSuccess indicates the subprocess completed successfully.
	ExitSuccess = iota

	// ExitFailure indicates the subprocess terminated unsuccessfully.
	ExitFailure

	// ExitTimeout indicates the subprocess timed out before completion.
	ExitTimeout
)

// MainFunc is the function that runs for this sub-process.
//
// The return value is a process exit code.
type MainFunc func() int

// Do f if nomad was launched as, "nomad [name]". The sub-process will exit
// without running any other part of Nomad.
func Do(name string, f MainFunc) {
	if len(os.Args) > 1 && os.Args[1] == name {
		os.Exit(f())
	}
}

// Expiration waits for timeout to elapse, then causes the process to exit with
// the ExitTimeout exit code.
func Expiration(timeout time.Duration) {
	go func() {
		time.Sleep(timeout)
		os.Exit(ExitTimeout)
	}()
}
