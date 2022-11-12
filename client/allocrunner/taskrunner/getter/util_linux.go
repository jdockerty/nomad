//go:build linux

package getter

import (
	"fmt"

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
// dir - the task directory
//
// Only applies to Linux, when active.
func lockdown(dir string) error {
	if !landlock.Available() {
		return nil
	}
	paths := []*landlock.Path{
		landlock.DNS(),
		landlock.Certs(),
		landlock.Shared(),
		landlock.Dir("/bin", "rx"),
		landlock.Dir("/usr/bin", "rx"),
		landlock.Dir("/usr/local/bin", "rx"),
		landlock.Dir(dir, "rwc"),
	}
	locker := landlock.New(paths...)
	fmt.Println("LOCKER", locker)
	return locker.Lock(landlock.Enforce)
}
