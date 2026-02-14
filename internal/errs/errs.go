package errs

import "errors"

const (
	CodeTestFailed = 1
	CodeUsage      = 2
)

type codedError struct {
	code int
	msg  string
	err  error
}

func (e *codedError) Error() string {
	if e.err == nil {
		return e.msg
	}
	if e.msg == "" {
		return e.err.Error()
	}
	return e.msg + ": " + e.err.Error()
}

func (e *codedError) Unwrap() error {
	return e.err
}

func (e *codedError) ExitCode() int {
	return e.code
}

func New(code int, msg string, err error) error {
	return &codedError{code: code, msg: msg, err: err}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var withCode interface{ ExitCode() int }
	if errors.As(err, &withCode) {
		return withCode.ExitCode()
	}
	return CodeTestFailed
}
