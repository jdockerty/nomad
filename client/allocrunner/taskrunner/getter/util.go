package getter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/nomad/client/interfaces"
	"github.com/hashicorp/nomad/nomad/structs"
)

const (
	// githubPrefixSSH is the prefix for downloading via git using ssh from GitHub.
	githubPrefixSSH = "git@github.com:"
)

func getURL(taskEnv interfaces.EnvReplacer, artifact *structs.TaskArtifact) (string, *Error) {
	source := taskEnv.ReplaceEnv(artifact.GetterSource)

	// fixup GitHub SSH URL such as git@github.com:hashicorp/nomad.git
	gitSSH := false
	if strings.HasPrefix(source, githubPrefixSSH) {
		gitSSH = true
		source = source[len(githubPrefixSSH):]
	}

	u, err := url.Parse(source)
	if err != nil {
		return "", &Error{
			URL:         artifact.GetterSource,
			Err:         fmt.Errorf("failed to parse source URL %q: %v", artifact.GetterSource, err),
			Recoverable: false,
		}
	}

	// build the URL by substituting as necessary
	q := u.Query()
	for k, v := range artifact.GetterOptions {
		q.Set(k, taskEnv.ReplaceEnv(v))
	}
	u.RawQuery = q.Encode()

	// add the prefix back if necessary
	sourceURL := u.String()
	if gitSSH {
		sourceURL = fmt.Sprintf("%s%s", githubPrefixSSH, sourceURL)
	}

	return sourceURL, nil
}

func getDestination(env interfaces.EnvReplacer, artifact *structs.TaskArtifact) (string, *Error) {
	destination, escapes := env.ClientPath(artifact.RelativeDest, true)
	if escapes {
		return "", &Error{
			URL:         artifact.GetterSource,
			Err:         fmt.Errorf("artifact destination path escapes alloc directory"),
			Recoverable: false,
		}
	}
	return destination, nil
}

func getMode(artifact *structs.TaskArtifact) getter.ClientMode {
	switch artifact.GetterMode {
	case structs.GetterModeFile:
		return getter.ClientModeFile
	case structs.GetterModeDir:
		return getter.ClientModeDir
	default:
		return getter.ClientModeAny
	}
}

func getHeaders(env interfaces.EnvReplacer, artifact *structs.TaskArtifact) map[string][]string {
	m := artifact.GetterHeaders
	if len(m) == 0 {
		return nil
	}
	headers := make(http.Header, len(m))
	for k, v := range m {
		headers.Set(k, env.ReplaceEnv(v))
	}
	return headers
}

func getTaskDir(env interfaces.EnvReplacer) string {
	p, _ := env.ClientPath("stub", false)
	return filepath.Dir(p)
}

func minimalVars() []string {
	getOr := func(key, value string) string {
		if v := os.Getenv(key); v != "" {
			value = v
		}
		return fmt.Sprintf("%s=%s", key, value)
	}
	return []string{
		getOr("PATH", "/usr/local/bin:/usr/bin:/bin"),
	}
}

func runCmd(env *parameters) *Error {
	bin, err := exec.LookPath("nomad")
	if err != nil {
		return &Error{
			URL:         env.Source,
			Err:         fmt.Errorf("nomad executable not found: %v", err),
			Recoverable: false,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	uid, gid := credentials()

	cmd := exec.CommandContext(ctx, bin, ProcessName)
	cmd.Env = minimalVars()
	cmd.Stdin = env.reader()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
	output, err := cmd.CombinedOutput()
	fmt.Println("output", string(output))
	if err != nil {
		return &Error{
			URL:         env.Source,
			Err:         fmt.Errorf("getter subprocess failed: %v", err),
			Recoverable: true,
		}
	}
	return nil
}
