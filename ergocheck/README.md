# ergocheck

[日本語](README.ja.md)

A static analyzer that enforces consistent usage of the `ergo` package and checks for best practices.

## Checks

### 1. Usage of `errors.New` and `fmt.Errorf`

Detects usage of `errors.New` and `fmt.Errorf` and recommends replacing them with `ergo.New` and `ergo.Wrap`.

```go
// NG
err := errors.New("user not found")
err := fmt.Errorf("failed: %w", err)

// OK
err := ergo.New("user not found")
err := ergo.Wrap(err, "failed")
```

### 2. Format Strings in Error Messages

Checks if error message arguments in `ergo.New` or `ergo.Wrap` contain format strings (`%s`, `%d`, `%v`, etc.).
Dynamic values should be passed as `slog.Attr`.

```go
// NG
err := ergo.New("user %s not found")
err := ergo.Wrap(err, "failed with code %d")

// OK
err := ergo.New("user not found", slog.String("user_id", userID))
err := ergo.Wrap(err, "failed", slog.Int("code", code))
```

### 3. Nil Error Detection

Checks if `nil` errors are passed to `ergo.Wrap` or `ergo.WithCode`.

```go
// NG
err := ergo.Wrap(nil, "failed")
err := ergo.WithCode(nil, code)

// OK
if err != nil {
    err = ergo.Wrap(err, "failed")
}
```

### 4. Using `ergo.New` in Package Variable Initialization

When `ergo.New` is used in package-level variable initialization, recommends replacing it with `ergo.NewSentinel`.

```go
// NG
var ErrNotFound = ergo.New("not found")

// OK
var ErrNotFound = ergo.NewSentinel("not found")
```

## Installation

```bash
go install github.com/newmo-oss/ergo/ergocheck/cmd/ergocheck@latest
```

## Usage

### Basic Usage

```bash
# Run on specific packages
go vet -vettool=$(which ergocheck) -ergocheck.packages='github.com/yourorg/yourapp/...' ./...
```

### Excluding Packages

```bash
# Exclude specific packages like vendor
go vet -vettool=$(which ergocheck) \
    -ergocheck.packages='github.com/yourorg/yourapp/...' \
    -ergocheck.excludes='github.com/yourorg/yourapp/vendor/...' \
    ./...
```

### Flags

- `-ergocheck.packages`: Specify target packages as a regular expression
- `-ergocheck.excludes`: Specify packages to exclude as a regular expression

## License

MIT License - see [LICENSE](../LICENSE) for details.
