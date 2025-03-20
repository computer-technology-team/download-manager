package types

type ErrorMsg struct {
	Err     error
	ErrInfo *string
}

func (e ErrorMsg) Error() string {
	return e.Err.Error()
}
