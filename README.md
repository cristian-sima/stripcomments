<p align="center">
  <img src="assets/logo.svg" width="140" height="140" alt="stripcomments logo">
</p>

<h1 align="center">stripcomments</h1>

<p align="center">
  <b>Tired of dropping the AI comments by hand?</b><br>
  Say goodbye to useless comments — keep the code, sweep the noise.
</p>

---

Every AI-generated file arrives buried in `// this function does X` narration.
`stripcomments` parses Go with the standard `go/parser`, drops every comment,
and re-emits gofmt-formatted output — in one command. Compiler and tool
**directives** are preserved, because removing them silently breaks builds or
ships empty values:

- `//go:build` / `// +build` constraints
- `//go:embed` (populates the variable below it)
- `//go:generate`, `//go:linkname`, and any other `//go:` directive

String and raw-string literals that happen to contain `//` or `/* */` are left
untouched — only real comments are removed.

## How it works

```mermaid
flowchart LR
    A["source.go<br/>+ AI comments"] --> B["go/parser<br/>ParseComments"]
    B --> C{comment is a<br/>//go: directive?}
    C -->|yes| K["keep"]
    C -->|no| D["drop"]
    K --> E["go/format<br/>gofmt output"]
    D --> E
    E --> F["clean source.go<br/>directives intact"]

    classDef keep fill:#a7f3d0,stroke:#059669,color:#064e3b;
    classDef drop fill:#fecaca,stroke:#dc2626,color:#7f1d1d;
    class K keep;
    class D drop;
```

**Before → after:**

```go
// UserID uniquely identifies a user.   ← swept
// It is generated at signup.           ← swept
//go:generate stringer -type=Role       ← kept
func NewUser() {}                        // returns nothing  ← swept
```

```go
//go:generate stringer -type=Role
func NewUser() {}
```

## Install

```sh
go install github.com/cristian-sima/stripcomments/cmd/stripcomments@latest
```

## CLI

Overwrites each named file in place:

```sh
stripcomments path/to/file.go another.go
```

## Library

```go
import "github.com/cristian-sima/stripcomments"

// StripComments returns gofmt-formatted source with comments removed.
out, err := stripcomments.StripComments(src)

// StripFile overwrites a file in place, preserving its mode.
err := stripcomments.StripFile("path/to/file.go")
```

## Strip a whole tree

```sh
find . -name '*.go' -not -path './vendor/*' -print0 \
  | xargs -0 stripcomments
```

## License

MIT — see [LICENSE](LICENSE).
