package subproc

import (
	"context"
	"fmt"
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

// Log the given message to standard error.
func Log(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Context creates a context setup with the given timeout.
func Context(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, timeout)
}

// SetExpiration is used to ensure the process terminates, once ctx
// is complete. A short grace period is added to allow any cleanup
// to happen first.
func SetExpiration(ctx context.Context) {
	const graceful = 5 * time.Second
	go func() {
		<-ctx.Done()
		time.Sleep(graceful)
		os.Exit(ExitTimeout)
	}()
}
