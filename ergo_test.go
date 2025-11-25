package ergo_test

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/newmo-oss/ergo"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		msg    string
		attrs  []slog.Attr
		format string

		wantString string
	}{
		"message %v":    {"error message", nil, "%v", "error message"},
		"message %+v":   {"error message", nil, "%+v", "error message"},
		"one attr %v":   {"error message", attrs(t, "key1", 100), "%v", "error message"},
		"one attr %+v":  {"error message", attrs(t, "key1", 100), "%+v", "error message: key1=100"},
		"two attrs %v":  {"error message", attrs(t, "key1", 100, "key2", "value2"), "%v", "error message"},
		"two attrs %+v": {"error message", attrs(t, "key1", 100, "key2", "value2"), "%+v", "error message: key1=100,key2=value2"},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ergo.New(tt.msg, tt.attrs...)

			// error string
			if got := fmt.Sprintf(tt.format, err); got != tt.wantString {
				t.Errorf("fmt.Sprintf(%q, err) does not match: (got, want) = (%q, %q)", tt.format, got, tt.wantString)
			}

			// unwrap
			if got := errors.Unwrap(err); got != nil {
				t.Errorf("Unwrap must be nil but got %v", got)
			}

			{ // AttrsAll
				got := slices.Collect(ergo.AttrsAll(err))
				if diff := cmp.Diff(got, tt.attrs); diff != "" {
					t.Error("AttrsAll does not match:", diff)
				}

				// test whether checking return value of yield
				for range ergo.AttrsAll(err) {
					break
				}
			}
		})
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	var (
		parentAttrs = attrs(t, "key1", 100)
		parent      = ergo.New("error", parentAttrs...)
	)

	cases := map[string]struct {
		parent error
		msg    string
		attrs  []slog.Attr
		format string

		wantString string
	}{
		"parent nil %v":  {nil, "wrap", nil, "%v", "wrap"},
		"parent nil %+v": {nil, "wrap", nil, "%+v", "wrap"},
		"parent %v":      {parent, "wrap", nil, "%v", "wrap: error"},
		"parent %+v":     {parent, "wrap", nil, "%+v", "wrap: error: key1=100"},
		"one attr %v":    {parent, "wrap", attrs(t, "key2", "value2"), "%v", "wrap: error"},
		"one attr %+v":   {parent, "wrap", attrs(t, "key2", "value2"), "%+v", "wrap: key2=value2: error: key1=100"},
		"two attr %v":    {parent, "wrap", attrs(t, "key2", "value2", "key3", true), "%v", "wrap: error"},
		"two attr %+v":   {parent, "wrap", attrs(t, "key2", "value2", "key3", true), "%+v", "wrap: key2=value2,key3=true: error: key1=100"},
		"same attr %v":   {parent, "wrap", attrs(t, "key2", "value2", "key1", true), "%v", "wrap: error"},
		"same attr %+v":  {parent, "wrap", attrs(t, "key2", "value2", "key1", true), "%+v", "wrap: key2=value2,key1=true: error: key1=100"},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ergo.Wrap(tt.parent, tt.msg, tt.attrs...)

			// error string
			if got := fmt.Sprintf(tt.format, err); got != tt.wantString {
				t.Errorf("fmt.Sprintf(%q, err) does not match: (got, want) = (%q, %q)", tt.format, got, tt.wantString)
			}

			// unwrap
			if got := errors.Unwrap(err); got != tt.parent {
				t.Error("Unwrap must be return parent")
			}

			allAttrs := tt.attrs
			if tt.parent != nil {
				for _, attr := range parentAttrs {
					if !slices.ContainsFunc(allAttrs, func(a slog.Attr) bool {
						return a.Key == attr.Key
					}) {
						allAttrs = append(allAttrs, attr)
					}
				}
			}

			{ // AttrsAll
				got := slices.Collect(ergo.AttrsAll(err))

				if diff := cmp.Diff(got, allAttrs); diff != "" {
					t.Error("AttrsAll does not match:", diff)
				}

				// test whether checking return value of yield
				for range ergo.AttrsAll(err) {
					break
				}
			}
		})
	}
}

func TestStackTraceOf(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err     error
		want    string
		wantNil bool
	}{
		"New":           {newErrorForTest("error"), "[ergo_stacktrace_test.go:11 ergo_stacktrace_test.go:22 ergo_stacktrace_test.go:10]", false},
		"Wrap(nil)":     {wrapErrorForTest(nil, "wrap"), "[ergo_stacktrace_test.go:17 ergo_stacktrace_test.go:22 ergo_stacktrace_test.go:16]", false},
		"Wrap(New)":     {wrapErrorForTest(newErrorForTest("error"), "wrap"), "[ergo_stacktrace_test.go:11 ergo_stacktrace_test.go:22 ergo_stacktrace_test.go:10]", false},
		"nil":           {nil, "", true},
		"no stacktrace": {errors.New("error"), "", true},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			st := ergo.StackTraceOf(tt.err)
			switch {
			case tt.wantNil && st != nil:
				t.Fatal("expect StackTraceOf returned nil, but got", st)
			case !tt.wantNil && st == nil:
				t.Fatal("StackTraceOf return nil")
			case st == nil:
				return
			}

			if got := fmt.Sprintf("%v", st[:3]); got != tt.want {
				t.Errorf("StackTrace does not match: (got, want) = (%q, %q)", got, tt.want)
			}
		})
	}
}

func TestNil(t *testing.T) {
	t.Parallel()

	t.Run("AttrsAll", func(t *testing.T) {
		t.Parallel()
		// test whether checking return value of yield
		for range ergo.AttrsAll(nil) {
			break
		}
	})

	t.Run("CodeOf", func(t *testing.T) {
		t.Parallel()
		if got := ergo.CodeOf(nil); !got.IsZero() {
			t.Error(`ergo.CodeOf(nil) must return ""`)
		}
	})
}

func attrs(t *testing.T, args ...any) []slog.Attr {
	t.Helper()
	if len(args)%2 != 0 {
		t.Fatal("invalid arguments")
	}

	var attrs []slog.Attr
	for kv := range slices.Chunk(args, 2) {
		key, ok := kv[0].(string)
		if !ok {
			t.Fatal("invalid key", kv[0])
		}
		attrs = append(attrs, slog.Any(key, kv[1]))
	}

	return attrs
}
