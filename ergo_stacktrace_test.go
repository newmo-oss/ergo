package ergo_test

import (
	"log/slog"

	"github.com/newmo-oss/ergo"
)

func newErrorForTest(msg string, attrs ...slog.Attr) error {
	return do(func() error { // lib/go/ergo/ergo_stacktrace_test.go:10
		return ergo.New(msg, attrs...) // lib/go/ergo/ergo_stacktrace_test.go:11
	})
}

func wrapErrorForTest(parent error, msg string, attrs ...slog.Attr) error {
	return do(func() error { // lib/go/ergo/ergo_stacktrace_test.go:16
		return ergo.Wrap(parent, msg, attrs...) // lib/go/ergo/ergo_stacktrace_test.go:17
	})
}

func do(f func() error) error {
	return f() // lib/go/ergo/ergo_stacktrace_test.go:28
}

var codeA = ergo.NewCode("A", "code A message") // lib/go/ergo/ergo_stacktrace_test.go:31
