package a

import (
	"errors"
	"fmt"

	"github.com/newmo-oss/ergo"
)

var ForPhi bool

var (
	ErrVarNew  = ergo.New("error")            // want `ergo\.New must not be used in package variable initilization, it should be replaced by ergo\.NewSentinel`
	ErrVarWrap = ergo.Wrap(ErrVarNew, "wrap") // want `ergo\.Wrap must not be used in package variable initilization, it should be replaced by errors\.Join`
)

var (
	_ = errors.New("error") // want `errors\.New must not be used in the .+ package, it should be replaced by ergo\.New`
	_ = fmt.Errorf("error") // want `fmt\.Errorf must not be used in the .+ package, it should be replaced by ergo\.Wrap`
)

var _ = func() {
	_ = errors.New("error") // want `errors\.New must not be used in the .+ package, it should be replaced by ergo\.New`
	_ = fmt.Errorf("error") // want `fmt\.Errorf must not be used in the .+ package, it should be replaced by ergo\.Wrap`
}

func forCheckDeprecatedFunc() {
	_ = errors.New("error") // want `errors\.New must not be used in the .+ package, it should be replaced by ergo\.New`
	_ = fmt.Errorf("error") // want `fmt\.Errorf must not be used in the .+ package, it should be replaced by ergo\.Wrap`
}

func forCheckFormatString() {
	err := ergo.New("error")
	_ = ergo.New("%s")                       // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%s"`
	_ = ergo.New("%d")                       // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%d"`
	_ = ergo.New("%v")                       // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%v"`
	_ = ergo.New("%T")                       // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%T"`
	_ = ergo.New("%+v")                      // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%\+v"`
	_ = ergo.New("%#v")                      // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%#v"`
	_ = ergo.New(fmt.Sprintf("%s", "error")) // ok

	_ = ergo.Wrap(err, "%s")                       // want `the message of github.com/newmo-oss/ergo.Wrap must not be format string such as "xxxx %s": "%s"`
	_ = ergo.Wrap(err, "%d")                       // want `the message of github.com/newmo-oss/ergo.Wrap must not be format string such as "xxxx %s": "%d"`
	_ = ergo.Wrap(err, "%v")                       // want `the message of github.com/newmo-oss/ergo.Wrap must not be format string such as "xxxx %s": "%v"`
	_ = ergo.Wrap(err, "%T")                       // want `the message of github.com/newmo-oss/ergo.Wrap must not be format string such as "xxxx %s": "%T"`
	_ = ergo.Wrap(err, "%+v")                      // want `the message of github.com/newmo-oss/ergo.Wrap must not be format string such as "xxxx %s": "%\+v"`
	_ = ergo.Wrap(err, "%#v")                      // want `the message of github.com/newmo-oss/ergo.Wrap must not be format string such as "xxxx %s": "%#v"`
	_ = ergo.Wrap(err, fmt.Sprintf("%s", "error")) // ok

	// typed const
	const msg1 string = "%s"
	_ = ergo.New(msg1) // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%s"`

	// expression
	const msg2 = "%" + "s"
	_ = ergo.New(msg2) // want `the message of github.com/newmo-oss/ergo.New must not be format string such as "xxxx %s": "%s"`
}

func forCheckNilErr() {
	_ = ergo.Wrap(nil, "") // want `The 1st argument of github.com/newmo-oss/ergo.Wrap must not be nil`
	code := ergo.NewCode("coe", "code")
	_ = ergo.WithCode(nil, code) // want `The 1st argument of github.com/newmo-oss/ergo.WithCode must not be nil`

	{
		var err error
		if ForPhi {
			err = ergo.New("error")
		}
		_ = ergo.Wrap(err, "")       // want `The 1st argument of github.com/newmo-oss/ergo.Wrap must not be nil`
		_ = ergo.WithCode(err, code) // want `The 1st argument of github.com/newmo-oss/ergo.WithCode must not be nil`

		if err != nil {
			// noop
		}

		if err != nil {
			_ = ergo.Wrap(err, "")       // OK
			_ = ergo.WithCode(err, code) // OK
		}

		if nil != err {
			_ = ergo.Wrap(err, "")       // OK
			_ = ergo.WithCode(err, code) // OK
		}
	}
}
