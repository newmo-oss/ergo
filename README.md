# ergo

[![Go Reference](https://pkg.go.dev/badge/github.com/newmo-oss/ergo.svg)](https://pkg.go.dev/github.com/newmo-oss/ergo)

[日本語](README.ja.md)

A structured error handling library for Go with attributes, error codes, and stack traces.

## Features

- **Structured Attributes**: Attach contextual information to errors using `log/slog.Attr`
- **Error Codes**: Associate typed codes with errors
- **Stack Traces**: Automatically capture stack traces
- **Error Wrapping**: Wrap errors while preserving context

## Installation

```bash
go get github.com/newmo-oss/ergo
```

## Usage

### Creating Errors

```go
import (
    "log/slog"

    "github.com/newmo-oss/ergo"
)

// Simple error
// Stack trace is automatically attached
err := ergo.New("failed to process request")

// With attributes
err := ergo.New("user not found", slog.String("user_id", "12345"), slog.Int("status_code", 404))
```

### Wrapping Errors

```go
// Wrap with additional context
// Stack trace is attached if not present
if err := doSomething(); err != nil {
    return ergo.Wrap(err, "failed to execute operation", slog.String("operation", "process"))
}
```

### Error Codes

```go
// Define error codes
// NewCode automatically captures the package path from the stack trace,
// so codes with the same key name from different packages can be distinguished
var (
    ErrCodeNotFound = ergo.NewCode("NotFound", "resource not found")
    ErrCodeInvalid  = ergo.NewCode("Invalid", "invalid input")
)

func doSomething() error {
    // Create error
    err := ergo.New("user not found", slog.String("user_id", "12345"))

    // Add code to error
    err = ergo.WithCode(err, ErrCodeNotFound)

    // Retrieve error code
    code := ergo.CodeOf(err)
    if code == ErrCodeNotFound {
        // handle not found error
    }

    // String representation of error code
    // Formatted as: package-path.Key: Message
    fmt.Println(code.String()) // github.com/yourorg/yourapp/service.NotFound: resource not found
}
```

### Retrieving Attributes

```go
// Iterate over all attributes (including parent errors)
for attr := range ergo.AttrsAll(err) {
    fmt.Printf("%s: %v\n", attr.Key, attr.Value)
}
```

### Stack Traces

```go
// Get stack trace
st := ergo.StackTraceOf(err)
if st != nil {
    fmt.Printf("Stack trace: %v\n", st)
}
```

### Sentinel Errors

```go
// Always use NewSentinel for package-level constants
// NewSentinel creates lightweight errors without stack traces, attributes, or error codes
var (
    ErrInvalidInput = ergo.NewSentinel("invalid input")
    ErrTimeout      = ergo.NewSentinel("operation timeout")
)

// Compare with sentinel errors
if errors.Is(err, ErrTimeout) {
    // handle timeout
}
```

For more details on sentinel errors, see [Dave Cheney's article](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully).

### Formatting

```go
err := ergo.New("operation failed", slog.String("key1", "value1"), slog.Int("key2", 100))
fmt.Printf("%v\n", err)   // operation failed
fmt.Printf("%+v\n", err)  // operation failed: key1=value1,key2=100
```

## Static Analysis: ergocheck

A static analyzer that enforces consistent usage of `ergo` and checks for best practices.
See [ergocheck/README.md](ergocheck/README.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Issues and Pull Requests are welcome.
