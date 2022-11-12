package getter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/helper/users"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/shoenig/test/must"
)

func setupDir(t *testing.T) (string, string) {
	uid, gid := users.NobodyIDs()

	allocDir := t.TempDir()
	taskDir := filepath.Join(allocDir, "local")
	topDir := filepath.Dir(allocDir)

	must.NoError(t, os.Chown(topDir, int(uid), int(gid)))
	must.NoError(t, os.Chmod(topDir, 0o755))

	must.NoError(t, os.Chown(allocDir, int(uid), int(gid)))
	must.NoError(t, os.Chmod(allocDir, 0o755))

	must.NoError(t, os.Mkdir(taskDir, 0o755))
	must.NoError(t, os.Chown(taskDir, int(uid), int(gid)))
	must.NoError(t, os.Chmod(taskDir, 0o755))
	return allocDir, taskDir
}

func TestSandbox_Get(t *testing.T) {
	sbox := New(&config.ArtifactConfig{
		HTTPReadTimeout: 0,
		HTTPMaxBytes:    0,
		GCSTimeout:      0,
		GitTimeout:      0,
		HgTimeout:       0,
		S3Timeout:       0,
	})

	_, taskDir := setupDir(t)
	env := &noopReplacer{taskDir: taskDir}
	fmt.Println("taskDir", taskDir)

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
