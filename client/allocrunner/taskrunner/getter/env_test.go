package getter

import (
	"path/filepath"

	"github.com/hashicorp/nomad/client/interfaces"
	"github.com/hashicorp/nomad/helper/escapingfs"
)

// noopReplacer is a noop version of taskenv.TaskEnv.ReplaceEnv.
type noopReplacer struct {
	taskDir string
}

func clientPath(taskDir, path string, join bool) (string, bool) {
	if !filepath.IsAbs(path) || (escapingfs.PathEscapesSandbox(taskDir, path) && join) {
		path = filepath.Join(taskDir, path)
	}
	path = filepath.Clean(path)
	if taskDir != "" && !escapingfs.PathEscapesSandbox(taskDir, path) {
		return path, false
	}
	return path, true
}

func (noopReplacer) ReplaceEnv(s string) string {
	return s
}

func (r noopReplacer) ClientPath(p string, join bool) (string, bool) {
	path, escapes := clientPath(r.taskDir, r.ReplaceEnv(p), join)
	return path, escapes
}

func noopTaskEnv(taskDir string) interfaces.EnvReplacer {
	return noopReplacer{
		taskDir: taskDir,
	}
}
