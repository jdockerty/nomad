//go:build linux

package getter

import (
	"runtime"

	"github.com/shoenig/go-landlock"
)

// lockdown isolates this process to only be able to write and create files in
// the task's task directory.
//
// Only applies to Linux.
func lockdown(dir string) error {
	switch runtime.GOOS {
	case "linux":
		enforce := landlock.IfElse(
			landlock.Available(),
			landlock.Enforce,
			landlock.Ignore,
		)
		return landlock.New(
			landlock.DNS(),
			landlock.Certs(),
			landlock.Dir(dir, "rwc"),
		).Lock(enforce)
	default:
		return nil
	}
}
