package getter

import (
	"errors"
	"testing"

	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/shoenig/test/must"
)

func TestError_Error(t *testing.T) {
	cases := []struct {
		name string
		err  *Error
		exp  string
	}{
		{"object nil", nil, "<nil>"},
		{"error nil", new(Error), "<nil>"},
		{"has error", &Error{Err: errors.New("oops")}, "oops"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := Error{Err: tc.err}
			result := e.Error()
			must.Eq(t, tc.exp, result)
		})
	}
}

func TestError_IsRecoverable(t *testing.T) {
	var _ structs.Recoverable = (*Error)(nil)
	must.True(t, (&Error{Recoverable: true}).IsRecoverable())
	must.False(t, (&Error{Recoverable: false}).IsRecoverable())
}
