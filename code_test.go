package ergo_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/newmo-oss/ergo"
)

func TestWithCode(t *testing.T) {
	t.Parallel()

	var (
		parentAttrs = attrs(t, "key1", 100)
		parent      = ergo.New("error", parentAttrs...)
	)

	cases := map[string]struct {
		parent error
		code   ergo.Code // codes are declared in ergo_stacktrace_test.go

		wantErrString string
		wantNil       bool
	}{
		"parent nil":     {nil, codeA, "", true},
		"parent not nil": {parent, codeA, "github.com/newmo-oss/ergo_test.A: code A message: error", false},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ergo.WithCode(tt.parent, tt.code)

			switch {
			case tt.wantNil && err != nil:
				t.Fatal("expect WithCode returned nil, but got", err)
			case !tt.wantNil && err == nil:
				t.Fatal("WithCode return nil")
			case err == nil:
				return
			}

			// error string
			if got := err.Error(); got != tt.wantErrString {
				t.Errorf("err.Error does not match: (got, want) = (%q, %q)", got, tt.wantErrString)
			}

			// code
			if got := ergo.CodeOf(err); got != tt.code {
				t.Errorf("CodeOf does not match: (got, want) = (%q, %q)", got, tt.code)
			}

			// unwrap
			if got := errors.Unwrap(err); got != tt.parent {
				t.Error("Unwrap must be return parent")
			}

			{ // AttrsAll
				got := slices.Collect(ergo.AttrsAll(err))

				if diff := cmp.Diff(got, parentAttrs); diff != "" {
					t.Error("AttrsAll does not match:", diff)
				}

				for range ergo.AttrsAll(err) {
					break
				}
			}
		})
	}
}
