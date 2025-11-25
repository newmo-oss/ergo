package ergo

// for test
func New(string, ...any) error {
	return nil
}

// for test
func Wrap(error, string, ...any) error {
	return nil
}

type Code struct{}

func NewCode(key, message string) Code {
	return Code{}
}

func WithCode(err error, code Code) error {
	return nil
}
