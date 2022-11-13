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
func Context(timeout time.Duration) context.Context {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, timeout)
	return ctx
}

// SetExpiration waits for ctx to expire, then force terminates the
// process with ExitTimeout shortly after. If ctx has no deadline
// no expiration is set.
func SetExpiration(ctx context.Context) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return
	}
	go func() {
		duration := deadline.Sub(time.Now())
		duration += 3 * time.Second // graceful shutdown
		time.Sleep(duration)
		os.Exit(ExitTimeout)
	}()
}
