# ergo

[![Go Reference](https://pkg.go.dev/badge/github.com/newmo-oss/ergo.svg)](https://pkg.go.dev/github.com/newmo-oss/ergo)

[English](README.md)

属性、エラーコード、スタックトレースを使った構造化エラー処理ライブラリです。

## 機能

- **構造化属性**: `log/slog.Attr`でエラーにコンテキスト情報を付与します
- **エラーコード**: エラーに型付きコードを関連付けます
- **スタックトレース**: 自動でスタックトレースをキャプチャします
- **エラーラッピング**: コンテキストを保ったままエラーをラップします

## 使い方

### エラーの作成

```go
import (
    "log/slog"

    "github.com/newmo-oss/ergo"
)

// シンプルなエラー
// スタックトレースが付与されます
err := ergo.New("failed to process request")

// 属性付き
err := ergo.New("user not found", slog.String("user_id", "12345"), slog.Int("status_code", 404))
```

### エラーのラップ

```go
// コンテキストを追加してラップ
// スタックトレースがない場合は付与されます
if err := doSomething(); err != nil {
    return ergo.Wrap(err, "failed to execute operation", slog.String("operation", "process"))
}
```

### エラーコード

```go
// エラーコード定義
// NewCodeはスタックトレースからパッケージパスを自動取得するため、
// 異なるパッケージで同じキー名を使っても区別できる
var (
    ErrCodeNotFound = ergo.NewCode("NotFound", "resource not found")
    ErrCodeInvalid  = ergo.NewCode("Invalid", "invalid input")
)

func doSomething() error {
    // エラー作成
    err := ergo.New("user not found", slog.String("user_id", "12345"))

    // エラーにコード追加
    err = ergo.WithCode(err, ErrCodeNotFound)

    // エラーコード取得
    code := ergo.CodeOf(err)
    if code == ErrCodeNotFound {
        // not foundエラー処理
    }

    // エラーコードの文字列化
    // パッケージパス.キー: メッセージ の形式で出力される
    fmt.Println(code.String()) // github.com/yourorg/yourapp/service.NotFound: resource not found
}
```

### 属性の取得

```go
// 全属性を走査（親エラー含む）
for attr := range ergo.AttrsAll(err) {
    fmt.Printf("%s: %v\n", attr.Key, attr.Value)
}
```

### スタックトレース

```go
// スタックトレース取得
st := ergo.StackTraceOf(err)
if st != nil {
    fmt.Printf("Stack trace: %v\n", st)
}
```

### センチネルエラー

```go
// パッケージレベル定数には必ずNewSentinelを使います
// NewSentinelはスタックトレース、属性、エラーコードを持たない軽量なエラーを作ります
var (
    ErrInvalidInput = ergo.NewSentinel("invalid input")
    ErrTimeout      = ergo.NewSentinel("operation timeout")
)

// センチネルエラーとの比較
if errors.Is(err, ErrTimeout) {
    // タイムアウト処理
}
```

センチネルエラーについての詳細は[Dave Cheney氏の記事](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)を参照してください。

### フォーマット

```go
err := ergo.New("operation failed", slog.String("key1", "value1"), slog.Int("key2", 100))
fmt.Printf("%v\n", err)   // operation failed
fmt.Printf("%+v\n", err)  // operation failed: key1=value1,key2=100
```

## 静的解析: ergocheck

ergoの使用を統一し、ベストプラクティスをチェックする静的解析ツールです。
詳細は[ergocheck/README.ja.md](ergocheck/README.ja.md)を参照してください。

## ライセンス

MIT License - 詳細は[LICENSE](LICENSE)を参照してください。
