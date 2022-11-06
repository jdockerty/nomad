package getter

import (
	"fmt"
	"testing"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/nomad/structs"
)

func TestSandbox_Get(t *testing.T) {
	sbox := New(&config.ArtifactConfig{
		HTTPReadTimeout: 0,
		HTTPMaxBytes:    0,
		GCSTimeout:      0,
		GitTimeout:      0,
		HgTimeout:       0,
		S3Timeout:       0,
	})

	taskDir := t.TempDir()
	env := &noopReplacer{taskDir: taskDir}

	artifact := &structs.TaskArtifact{
		GetterSource:  "https://github.com/shoenig/test.git",
		GetterOptions: nil,
		GetterHeaders: nil,
		GetterMode:    "auto",
		RelativeDest:  "local/test",
	}

	err := sbox.Get(env, artifact)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
	} else {
		fmt.Println("no error")
	}
}
