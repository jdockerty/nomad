//go:build linux

package getter

import (
	"github.com/hashicorp/nomad/helper/users"
	"github.com/shoenig/go-landlock"
)

// credentials returns the UID and GID of the user the child process will run as.
// On Linux this will be the nobody user.
func credentials() (uint32, uint32) {
	uid, gid := users.NobodyIDs()
	return uid, gid
}

// lockdown isolates this process to only be able to write and create files in
// the task's task directory.
//
// Only applies to Linux.
func lockdown(dir string) error {
	mode := landlock.IfElse(
		landlock.Available(),
		landlock.Enforce,
		landlock.Ignore,
	)
	return landlock.New(
		landlock.DNS(),
		landlock.Certs(),
		landlock.Dir(dir, "rwc"),
		landlock.File("/bin/getent", "rx"),
	).Lock(mode)
}
