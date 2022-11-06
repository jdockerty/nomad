package getter

// Error is a RecoverableError used to include the URL along with the underlying
// fetching error.
type Error struct {
	URL         string
	Err         error
	Recoverable bool
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) IsRecoverable() bool {
	return e.Recoverable
}
