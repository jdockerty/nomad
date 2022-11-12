package getter

import (
	"fmt"
	"os"
	"os/user"

	"github.com/hashicorp/nomad/helper/subproc"
)

const (
	ProcessName = "getter"
)

func init() {
	subproc.Do(ProcessName, func() int {

		// get client and artifact configuration from standard IO
		env := new(environment)
		if err := env.read(os.Stdin); err != nil {
			fail("failed to read configuration: %v", err)
			return subproc.ExitFailure
		}

		// force quit after maximum timeout exceeded
		subproc.Expire(env.timeout())

		u, err := user.Current()
		if err != nil {
			fail("failed to lookup user: %v", err)
		}
		fmt.Println("u is", u.Username, "uid", u.Uid, "gid", u.Gid, "taskdir", env.TaskDir)

		fmt.Println("env is", os.Environ())

		fi, err := os.Stat(env.TaskDir)
		if err != nil {
			fail("failed to stat task dir: %v", err)
		}

		fmt.Println("isDir", fi.IsDir())

		// sandbox the filesystem for this process
		if err := lockdown(env.TaskDir); err != nil {
			fail("failed to sandbox getter process: %v", err)
			return subproc.ExitFailure
		}

		fmt.Printf("config %#v\n", env)

		c := env.client()

		if err := c.Get(); err != nil {
			fail("failed to download artifact: %v", err)
			return subproc.ExitFailure
		}

		fmt.Printf("success!")

		return subproc.ExitSuccess
	})
}

func fail(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}
