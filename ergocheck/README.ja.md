# ergocheck

[English](README.md)

`ergo`パッケージの使用を統一し、ベストプラクティスをチェックする静的解析ツールです。

## 検出項目

### 1. `errors.New`と`fmt.Errorf`の使用

`errors.New`や`fmt.Errorf`の使用を検出し、`ergo.New`や`ergo.Wrap`への置き換えを推奨します。

```go
// NG
err := errors.New("user not found")
err := fmt.Errorf("failed: %w", err)

// OK
err := ergo.New("user not found")
err := ergo.Wrap(err, "failed")
```

### 2. エラーメッセージ内のフォーマット文字列

`ergo.New`や`ergo.Wrap`のメッセージ引数にフォーマット文字列（`%s`, `%d`, `%v`など）が含まれていないかチェックします。
動的な値は`slog.Attr`で渡すべきです。

```go
// NG
err := ergo.New("user %s not found")
err := ergo.Wrap(err, "failed with code %d")

// OK
err := ergo.New("user not found", slog.String("user_id", userID))
err := ergo.Wrap(err, "failed", slog.Int("code", code))
```

### 3. `nil`エラーの検出

`ergo.Wrap`や`ergo.WithCode`に`nil`エラーが渡されていないかチェックします。

```go
// NG
err := ergo.Wrap(nil, "failed")
err := ergo.WithCode(nil, code)

// OK
if err != nil {
    err = ergo.Wrap(err, "failed")
}
```

### 4. パッケージ変数初期化での`ergo.New`使用

パッケージレベルの変数初期化で`ergo.New`が使われている場合、`ergo.NewSentinel`への置き換えを推奨します。

```go
// NG
var ErrNotFound = ergo.New("not found")

// OK
var ErrNotFound = ergo.NewSentinel("not found")
```

## インストール

```bash
go install github.com/newmo-oss/ergo/ergocheck/cmd/ergocheck@latest
```

## 使い方

### 基本的な使い方

```bash
# 特定のパッケージで実行
go vet -vettool=$(which ergocheck) -ergocheck.packages='github.com/yourorg/yourapp/...' ./...
```

### パッケージの除外

```bash
# vendorなど特定のパッケージを除外
go vet -vettool=$(which ergocheck) \
    -ergocheck.packages='github.com/yourorg/yourapp/...' \
    -ergocheck.excludes='github.com/yourorg/yourapp/vendor/...' \
    ./...
```

### フラグ

- `-ergocheck.packages`: チェック対象のパッケージを正規表現で指定
- `-ergocheck.excludes`: 除外するパッケージを正規表現で指定

## ライセンス

MIT License - 詳細は[LICENSE](../LICENSE)を参照してください。
