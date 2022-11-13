package getter

import (
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
			subproc.Log("failed to read configuration: %v", err)
			return subproc.ExitFailure
		}

		// create context with the overall timeout
		ctx, cancel := subproc.Context(env.deadline())
		defer cancel()

		// force quit after maximum timeout exceeded
		subproc.SetExpiration(ctx)

		dir := env.TaskDir
		executes := env.executes()

		// sandbox the filesystem for this process
		if err := lockdown(dir, executes); err != nil {
			subproc.Log("failed to sandbox getter process: %v", err)
			return subproc.ExitFailure
		}

		// create the go-getter client
		// options were already transformed into url query parameters
		// headers were already replaced and are usable now
		c := env.client(ctx)

		// run the go-getter client
		if err := c.Get(); err != nil {
			subproc.Log("failed to download artifact: %v", err)
			return subproc.ExitFailure
		}

		subproc.Log("artifact download was a success")
		return subproc.ExitSuccess
	})
}
