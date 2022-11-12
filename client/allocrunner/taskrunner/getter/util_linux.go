//go:build linux

package getter

import (
	"os/exec"

	"github.com/hashicorp/nomad/helper/users"
	"github.com/shoenig/go-landlock"
)

// credentials returns the UID and GID of the user the child process will run as.
// On Linux this will be the nobody user.
func credentials() (uint32, uint32) {
	uid, gid := users.NobodyIDs()
	return uid, gid
}

// findGetterBins returns absolute paths to the executable binaries go-getter
// may use under the hood. These paths will need to be allowed to execute under
// landlock.
func findGetterBins() []*landlock.Path {
	var paths []*landlock.Path
	lookup := func(name string) {
		if path, err := exec.LookPath(name); err == nil {
			paths = append(paths, landlock.File(path, "rx"))
		}
	}
	lookup("getent")
	lookup("git")
	lookup("hg")
	return paths
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
		landlock.Dir(dir, "rwc"),
	}
	paths = append(paths, findGetterBins()...)
	return landlock.New(paths...).Lock(landlock.Enforce)
}
