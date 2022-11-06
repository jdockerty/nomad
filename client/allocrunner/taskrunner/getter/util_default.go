//go:build !linux

package getter

// lockdown isolates this process to only be able to write and create files in
// the task's task directory.
//
// Only applies to Linux.
func lockdown(dir string) error {
	return nil
}
