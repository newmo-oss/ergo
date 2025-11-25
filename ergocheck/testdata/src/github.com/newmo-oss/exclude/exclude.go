package notarget

import (
	"errors"
)

func f() {
	_ = errors.New("error") // ok - this package is not target
}
