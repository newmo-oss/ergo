package ergo

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"slices"
	"strings"

	"github.com/newmo-oss/go-caller"
)

type defaultError struct {
	parent     error
	msg        string
	attrs      []slog.Attr
	stacktrace caller.StackTrace
}

func (err *defaultError) Error() string {
	if err.msg == "" {
		return err.parent.Error()
	}

	var suffix string
	if err.parent != nil {
		suffix = ": " + err.parent.Error()
	}

	return err.msg + suffix
}

func (err *defaultError) Unwrap() error {
	return err.parent
}

func (err *defaultError) Format(s fmt.State, verb rune) {
	if verb == 'v' && s.Flag('+') {
		if len(err.attrs) == 0 {
			fmt.Fprint(s, err.msg)
			if err.parent != nil {
				if err.msg != "" {
					fmt.Fprint(s, ": ")
				}
				errf, ok := err.parent.(interface{ Format(fmt.State, rune) })
				if ok {
					errf.Format(s, verb)
				} else {
					fmt.Fprint(s, err.parent.Error())
				}
			}
			return
		}

		values := make([]string, len(err.attrs))
		for i, attr := range err.attrs {
			values[i] = attr.String()
		}

		if err.msg != "" {
			fmt.Fprintf(s, "%s: %s", err.msg, strings.Join(values, ","))
		} else {
			fmt.Fprint(s, strings.Join(values, ","))
		}
		if err.parent != nil {
			fmt.Fprint(s, ": ")
			errf, ok := err.parent.(interface{ Format(fmt.State, rune) })
			if ok {
				errf.Format(s, verb)
			} else {
				fmt.Fprint(s, err.parent.Error())
			}
		}
		return
	}
	fmt.Fprint(s, err.Error())
}

func newDefaultError(msg string, attrs ...slog.Attr) error {
	return &defaultError{
		msg: msg,
		// Expanding the slice to variadic argument is just a shallow copy of the original slice.
		// Because it is not goroutine safe, the attrs must be cloned.
		// see: https://go.dev/play/p/Q5a9oG_Uv2i
		attrs:      slices.Clone(attrs),
		stacktrace: caller.New(2),
	}
}

// New creates a new error with given attributes.
// The attributes can be retrieved using [AttrsAll].
// The error has a stacktrace of the callers.
// The stacktrace can be obtained via [StackTraceOf].
// If you would like to create an error into a package variable, you must use [NewSentinel].
func New(msg string, attrs ...slog.Attr) error {
	return newDefaultError(msg, attrs...)
}

// NewSentinel creates a new [sentinel error].
// The sentinel error does not have a stacktrace, any attributes and an error code.
//
// [sentinel error]: https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully#sentinel%20errors
func NewSentinel(msg string) error {
	return errors.New(msg)
}

// Wrap creates a new wrapped error of the parent error with given attributes.
// The attributes can be retrieved using [AttrsAll].
// The error has a stacktrace of the callers.
// The stacktrace can be obtained via [StackTraceOf].
// Calling errors.Unwrap with the wrapped error returns the parent error.
func Wrap(parent error, msg string, attrs ...slog.Attr) error {
	err := &defaultError{
		parent: parent,
		msg:    msg,
		attrs:  slices.Clone(attrs),
	}

	if st := StackTraceOf(parent); st == nil {
		err.stacktrace = caller.New(1)
	}

	return err
}

// AttrsAll returns an iterator that iterates over the attributes
// of the given error and its parent ones.
// If the parent of the error has the same attribute key as the children,
// the children are iterated over the parent.
func AttrsAll(err error) iter.Seq[slog.Attr] {
	return attrsAll(err, make(map[string]struct{}))
}

func attrsAll(err error, done map[string]struct{}) iter.Seq[slog.Attr] {
	return func(yield func(slog.Attr) bool) {
		var defaultError *defaultError
		if !errors.As(err, &defaultError) {
			return
		}

		for _, attr := range defaultError.attrs {
			if _, ok := done[attr.Key]; ok {
				continue
			}

			done[attr.Key] = struct{}{}
			if !yield(attr) {
				return
			}
		}

		for attr := range attrsAll(defaultError.parent, done) {
			if !yield(attr) {
				return
			}
		}
	}
}

// StackTraceOf returns stacktrace of the given error.
// If err does not have stacktrace, StackTraceOf returns nil.
func StackTraceOf(err error) caller.StackTrace {
	var defaultError *defaultError
	if !errors.As(err, &defaultError) {
		return nil
	}

	if st := defaultError.stacktrace; st != nil {
		return st
	}

	return StackTraceOf(defaultError.parent)
}
