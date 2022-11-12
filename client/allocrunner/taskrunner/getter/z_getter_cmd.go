package getter

import (
	"fmt"
	"os"

	"github.com/hashicorp/nomad/helper/subproc"
)

const (
	ProcessName = "getter"
)

func init() {
	subproc.Do(ProcessName, func() int {

		// get client and artifact configuration from standard IO
		env := new(parameters)
		if err := env.read(os.Stdin); err != nil {
			fail("failed to read configuration: %v", err)
			return subproc.ExitFailure
		}

		// force quit after maximum timeout exceeded
		subproc.Expiration(env.timeout())

		// sandbox the filesystem for this process
		if err := lockdown(env.TaskDir); err != nil {
			fail("failed to sandbox getter process: %v", err)
			return subproc.ExitFailure
		}

		// create the go-getter client
		// options were already transformed into url query parameters
		// headers were already replaced and are usable now
		c := env.client()

		// run the go-getter client
		if err := c.Get(); err != nil {
			fail("failed to download artifact: %v", err)
			return subproc.ExitFailure
		}

		log("artifact download was a success")
		return subproc.ExitSuccess
	})
}

func fail(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func log(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}
