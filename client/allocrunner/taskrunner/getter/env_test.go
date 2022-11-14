package getter

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/nomad/client/interfaces"
	"github.com/hashicorp/nomad/helper/escapingfs"
	"github.com/shoenig/test/must"
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

const paramsAsJSON = `
{
  "http_read_timeout": 1000000000,
  "http_max_bytes": 2000,
  "gcs_timeout": 2000000000,
  "git_timeout": 3000000000,
  "hg_timeout": 4000000000,
  "s3_timeout": 5000000000,
  "artifact_mode": 2,
  "artifact_source": "https://example.com/file.txt",
  "artifact_destination": "local/out.txt",
  "artifact_headers": {
    "X-Nomad-Artifact": ["hi"]
  },
  "task_dir": "/path/to/task"
}`

var paramsAsStruct = &parameters{
	HTTPReadTimeout: 1 * time.Second,
	HTTPMaxBytes:    2000,
	GCSTimeout:      2 * time.Second,
	GitTimeout:      3 * time.Second,
	HgTimeout:       4 * time.Second,
	S3Timeout:       5 * time.Second,

	Mode:        getter.ClientModeFile,
	Source:      "https://example.com/file.txt",
	Destination: "local/out.txt",
	TaskDir:     "/path/to/task",
	Headers: map[string][]string{
		"X-Nomad-Artifact": {"hi"},
	},
}

func TestParameters_reader(t *testing.T) {
	p := paramsAsStruct
	reader := p.reader()
	b, err := ioutil.ReadAll(reader)
	must.NoError(t, err)
	must.EqJSON(t, paramsAsJSON, string(b))
}

func TestParameters_read(t *testing.T) {
	reader := strings.NewReader(paramsAsJSON)
	p := new(parameters)
	err := p.read(reader)
	must.NoError(t, err)
	must.Equal(t, paramsAsStruct, p)
}

func TestParameters_deadline(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		dur := paramsAsStruct.deadline()
		must.Eq(t, 1*time.Hour, dur)
	})

	t.Run("long", func(t *testing.T) {
		params := &parameters{
			HTTPReadTimeout: 1 * time.Hour,
			GCSTimeout:      2 * time.Hour,
			GitTimeout:      3 * time.Hour,
			HgTimeout:       4 * time.Hour,
			S3Timeout:       5 * time.Hour,
		}
		dur := params.deadline()
		must.Eq(t, 5*time.Hour, dur)
	})
}

func TestParameters_executes(t *testing.T) {
	cases := []struct {
		name   string
		source string
		exp    bool
	}{
		{name: "git", exp: true, source: "git::https://github.com/hashicorp/nomad"},
		{name: "hg", exp: true, source: "hg::https://example.com/nomad"},
		{name: "http", exp: false, source: "https://github.com/hashicorp/nomad"},
		{name: "s3", exp: false, source: "s3::https://s3.amazon.com/hashicorp/nomad"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &parameters{Source: tc.source}
			result := p.executes()
			must.Eq(t, tc.exp, result)
		})
	}
}

func TestParameters_client(t *testing.T) {
	ctx := context.Background()
	c := paramsAsStruct.client(ctx)
	must.NotNil(t, c)

	// security options
	must.False(t, c.Insecure)
	must.True(t, c.DisableSymlinks)
	must.Eq(t, umask, c.Umask)

	// artifact options
	must.Eq(t, "https://example.com/file.txt", c.Src)
	must.Eq(t, "local/out.txt", c.Dst)
}