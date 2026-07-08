# gostripcomments

Strip comments from Go source while keeping the ones that matter.

`gostripcomments` parses Go with the standard `go/parser`, drops every comment,
and re-emits gofmt-formatted output. Compiler and tool **directives** are
preserved, because removing them silently breaks builds or ships empty values:

- `//go:build` / `// +build` constraints
- `//go:embed` (populates the variable below it)
- `//go:generate`, `//go:linkname`, and any other `//go:` directive

String and raw-string literals that happen to contain `//` or `/* */` are left
untouched — only real comments are removed.

## Install

```sh
go install github.com/cristian-sima/gostripcomments/cmd/gostripcomments@latest
```

## CLI

Overwrites each named file in place:

```sh
gostripcomments path/to/file.go another.go
```

## Library

```go
import "github.com/cristian-sima/gostripcomments"

// Strip returns gofmt-formatted source with comments removed.
out, err := stripcomments.Strip(src)

// StripFile overwrites a file in place, preserving its mode.
err := stripcomments.StripFile("path/to/file.go")
```

## Strip a whole tree

```sh
find . -name '*.go' -not -path './vendor/*' -print0 \
  | xargs -0 gostripcomments
```

## License

MIT — see [LICENSE](LICENSE).
