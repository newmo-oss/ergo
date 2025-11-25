package ergo

import (
	"errors"
	"fmt"

	"github.com/newmo-oss/go-caller"
)

// Code is an error code that can be associated with an error using [WithCode].
type Code struct {
	pkgpath string
	key     string
	message string
}

// NewCode creates new error code with the key and the message.
func NewCode(key, message string) Code {
	code := Code{
		key:     key,
		message: message,
	}
	st := caller.New(1)
	if len(st) > 0 {
		code.pkgpath = st[0].PkgPath()
	}
	return code
}

// IsZero returns whether code is zero value or not.
func (code Code) IsZero() bool {
	return code == Code{}
}

// String implements [fmt.Stringer].
func (code Code) String() string {
	key := code.key
	if code.pkgpath != "" {
		key = code.pkgpath + "." + key
	}
	return fmt.Sprintf("%s: %s", key, code.message)
}

// PkgPath returns the import path of the package the code belongs to
// The import path which is from the stack trace of calling [NewCode].
func (code Code) PkgPath() string {
	return code.pkgpath
}

// Key returns the key of error code.
func (code Code) Key() string {
	return code.key
}

// Message returns the message of error code.
func (code Code) Message() string {
	return code.message
}

type codedError struct {
	parent error
	code   Code
}

func (err *codedError) Error() string {
	return fmt.Sprintf("%s: %s", err.code, err.parent.Error())
}

func (err *codedError) Unwrap() error {
	return err.parent
}

// WithCode creates a new error which associated with the given code and the parent error.
// The code can be obtained via [CodeOf].
func WithCode(err error, code Code) error {
	if err == nil {
		return nil
	}

	return &codedError{
		parent: err,
		code:   code,
	}
}

// CodeOf returns the associated code with the error.
func CodeOf(err error) Code {
	var codedError *codedError
	if errors.As(err, &codedError) {
		return codedError.code
	}
	return Code{}
}
