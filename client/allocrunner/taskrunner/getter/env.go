package getter

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/nomad/helper"
)

// parameters is encoded by the Nomad client and decoded by the getter sub-process
// so it can know what to do. We use standard IO instead of parameters variables
// because the job-submitter has control over the parameters and that is scary,
// see https://www.opencve.io/cve/CVE-2022-41716.
type parameters struct {
	// Config
	HTTPReadTimeout time.Duration `json:"http_read_timeout"`
	HTTPMaxBytes    int64         `json:"http_max_bytes"`
	GCSTimeout      time.Duration `json:"gcs_timeout"`
	GitTimeout      time.Duration `json:"git_timeout"`
	HgTimeout       time.Duration `json:"hg_timeout"`
	S3Timeout       time.Duration `json:"s3_timeout"`

	// Artifact
	Mode        getter.ClientMode   `json:"artifact_mode"`
	Source      string              `json:"artifact_source"`
	Destination string              `json:"artifact_destination"`
	Headers     map[string][]string `json:"artifact_headers"`

	// Task Environment
	TaskDir string `json:"task_dir"`
}

func (e *parameters) reader() io.Reader {
	b, err := json.Marshal(e)
	if err != nil {
		b = nil
	}
	return strings.NewReader(string(b))
}

func (e *parameters) read(r io.Reader) error {
	return json.NewDecoder(r).Decode(e)
}

func (e *parameters) timeout() time.Duration {
	max := time.Duration(0)
	max = helper.Max(max, e.HTTPReadTimeout)
	max = helper.Max(max, e.GCSTimeout)
	max = helper.Max(max, e.GitTimeout)
	max = helper.Max(max, e.HgTimeout)
	max = helper.Max(max, e.S3Timeout)
	max += 1 * time.Second
	return max
}

// executes returns true if go-getter will be used in a mode that
// requires the use of exec
func (e *parameters) executes() bool {
	if strings.HasPrefix(e.Source, "git::") {
		return true
	}
	if strings.HasPrefix(e.Source, "hg::") {
		return true
	}
	return false
}

const (
	// blocks from downloading executables (?)
	umask = 060000000
)

func (e *parameters) client() *getter.Client {
	httpGetter := &getter.HttpGetter{
		Netrc:  true,
		Client: cleanhttp.DefaultClient(),
		Header: e.Headers,

		// Do not support the custom X-Terraform-Get header and
		// associated logic.
		XTerraformGetDisabled: true,

		// Disable HEAD requests as they can produce corrupt files when
		// retrying a download of a resource that has changed.
		// hashicorp/go-getter#219
		DoNotCheckHeadFirst: true,

		// Read timeout for HTTP operations. Must be long enough to
		// accommodate large/slow downloads.
		ReadTimeout: e.HTTPReadTimeout,

		// Maximum download size. Must be large enough to accommodate
		// large downloads.
		MaxBytes: e.HTTPMaxBytes,
	}
	return &getter.Client{
		Ctx:             nil, // maybe use this?
		Src:             e.Source,
		Dst:             e.Destination,
		Mode:            e.Mode,
		Umask:           umask,
		Insecure:        false,
		DisableSymlinks: false,
		Getters: map[string]getter.Getter{
			"git": &getter.GitGetter{
				Timeout: e.GitTimeout,
			},
			"hg": &getter.HgGetter{
				Timeout: e.HgTimeout,
			},
			"gcs": &getter.GCSGetter{
				Timeout: e.GCSTimeout,
			},
			"s3": &getter.S3Getter{
				Timeout: e.S3Timeout,
			},
			"http":  httpGetter,
			"https": httpGetter,
		},
	}
}
