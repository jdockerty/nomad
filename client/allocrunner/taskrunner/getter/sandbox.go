package getter

import (
	"fmt"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/client/interfaces"
	"github.com/hashicorp/nomad/nomad/structs"
)

// Sandbox is used for launching "getter" sub-process helpers for downloading
// artifacts. A Nomad client creates one of these and the task runner will call
// Get per artifact. Think "one site per browser tab" security model.
type Sandbox interface {
	Get(interfaces.EnvReplacer, *structs.TaskArtifact) error
}

// New creates a Sandbox with the given ArtifactConfig.
func New(ac *config.ArtifactConfig) Sandbox {
	return &sandbox{ac: ac}
}

type sandbox struct {
	ac *config.ArtifactConfig
}

func (s *sandbox) Get(env interfaces.EnvReplacer, artifact *structs.TaskArtifact) error {
	source, err := getURL(env, artifact)
	if err != nil {
		return err
	}

	destination, err := getDestination(env, artifact)
	if err != nil {
		return err
	}

	mode := getMode(artifact)
	headers := getHeaders(env, artifact)
	dir := getTaskDir(env)

	fmt.Println("dir:", dir)

	environ := &environment{
		HTTPReadTimeout: s.ac.HTTPReadTimeout,
		HTTPMaxBytes:    s.ac.HTTPMaxBytes,
		GCSTimeout:      s.ac.GCSTimeout,
		GitTimeout:      s.ac.GitTimeout,
		HgTimeout:       s.ac.HgTimeout,
		S3Timeout:       s.ac.S3Timeout,
		Mode:            mode,
		Source:          source,
		Destination:     destination,
		Headers:         headers,
		TaskDir:         dir,
	}

	err = runCmd(environ)
	if err != nil {
		return err
	}

	return nil // yay!
}
