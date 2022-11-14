package getter

import (
	"errors"
	"testing"

	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/shoenig/test/must"
)

func TestUtil_getURL(t *testing.T) {
	cases := []struct {
		name     string
		artifact *structs.TaskArtifact
		expURL   string
		expErr   *Error
	}{{
		name:     "basic http",
		artifact: &structs.TaskArtifact{GetterSource: "example.com"},
		expURL:   "example.com",
		expErr:   nil,
	}, {
		name:     "bad url",
		artifact: &structs.TaskArtifact{GetterSource: "::example.com"},
		expURL:   "",
		expErr: &Error{
			URL:         "::example.com",
			Err:         errors.New(`failed to parse source URL "::example.com": parse "::example.com": missing protocol scheme`),
			Recoverable: false,
		},
	}}

	env := &noopReplacer{taskDir: "/path/to/task"}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getURL(env, tc.artifact)
			must.Equal(t, tc.expErr, err)
			must.Eq(t, tc.expURL, result)
		})
	}
}
