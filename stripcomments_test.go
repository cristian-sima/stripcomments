package stripcomments

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStrip_InlineTrailingComment(t *testing.T) {
	src := `package p

func F() {
	x := 1 // comment
	_ = x
}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	if strings.Contains(string(out), "comment") {
		t.Errorf("expected comment to be stripped, got:\n%s", out)
	}
	if !strings.Contains(string(out), "x := 1") {
		t.Errorf("expected code line to survive, got:\n%s", out)
	}
}

func TestStrip_StandaloneAndBlockComments(t *testing.T) {
	src := `package p

// doc comment
func f() {
	/* block comment */
	x := 1
	_ = x
}
`
	out, err := StripComments([]byte(src), DropUnexported())
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	if strings.Contains(string(out), "comment") {
		t.Errorf("expected all comments to be stripped, got:\n%s", out)
	}
}

func TestStrip_StringLiteralsUntouched(t *testing.T) {
	src := "package p\n\n" +
		`func F() {` + "\n" +
		"\turl := \"http://example.com\" // trailing comment\n" +
		"\tblock := \"/* not a comment */\"\n" +
		"\traw := `has // and /* inside */ too`\n" +
		"\t_ = url\n\t_ = block\n\t_ = raw\n}\n"

	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, `"http://example.com"`) {
		t.Errorf("expected string literal with // to survive, got:\n%s", got)
	}
	if !strings.Contains(got, `"/* not a comment */"`) {
		t.Errorf("expected string literal with /* */ to survive, got:\n%s", got)
	}
	if !strings.Contains(got, "`has // and /* inside */ too`") {
		t.Errorf("expected raw string literal to survive untouched, got:\n%s", got)
	}
	if strings.Contains(got, "trailing comment") {
		t.Errorf("expected real trailing comment to be stripped, got:\n%s", got)
	}
}

func TestStrip_PreservesBuildConstraints(t *testing.T) {
	src := `//go:build windows
// +build windows

package p

func F() {
	x := 1 // comment
	_ = x
}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "//go:build windows") {
		t.Errorf("expected //go:build constraint to survive, got:\n%s", got)
	}
	if !strings.Contains(got, "// +build windows") {
		t.Errorf("expected legacy +build constraint to survive, got:\n%s", got)
	}
	if strings.Contains(got, "// comment") {
		t.Errorf("expected regular comment to still be stripped, got:\n%s", got)
	}
}

func TestStrip_PreservesGoEmbedDirective(t *testing.T) {
	src := `package p

import "embed"

//go:embed data/*.json
var schemaFiles embed.FS

func F() {
	x := 1 // comment
	_ = x
}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "//go:embed data/*.json") {
		t.Errorf("expected //go:embed directive to survive, got:\n%s", got)
	}
	if strings.Count(got, "//go:embed") != 1 {
		t.Errorf("expected //go:embed to appear exactly once, got:\n%s", got)
	}
	if strings.Contains(got, "// comment") {
		t.Errorf("expected regular comment to still be stripped, got:\n%s", got)
	}
}

func TestStrip_PreservesDirectiveAdjacentToDocComment(t *testing.T) {
	src := `package p

import "embed"

// schemaFiles holds every schema on disk.
//go:embed data/*.json
var schemaFiles embed.FS
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "//go:embed data/*.json") {
		t.Errorf("expected //go:embed directive to survive, got:\n%s", got)
	}
	if strings.Contains(got, "holds every schema") {
		t.Errorf("expected the adjacent doc comment to still be stripped, got:\n%s", got)
	}
}

func TestStrip_PreservesDirectiveAfterDocParagraphAndBlankCommentLine(t *testing.T) {
	src := `package p

// tessPatterns is bundled into the binary so no setup is required on
// the host.
//
//go:embed tessdata/patterns.txt
var tessPatterns []byte

//go:embed tessdata/words.txt
var tessWords []byte
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "//go:embed tessdata/patterns.txt") {
		t.Errorf("expected first //go:embed directive to survive, got:\n%s", got)
	}
	if !strings.Contains(got, "//go:embed tessdata/words.txt") {
		t.Errorf("expected second //go:embed directive to survive, got:\n%s", got)
	}
	if strings.Contains(got, "bundled into the binary") {
		t.Errorf("expected the doc paragraph to still be stripped, got:\n%s", got)
	}

	formatted, err := format.Source(out)
	if err != nil {
		t.Fatalf("format.Source on stripped output failed: %v", err)
	}
	if string(out) != string(formatted) {
		t.Errorf("stripped output is not gofmt-formatted.\ngot:\n%s\nwant:\n%s", out, formatted)
	}
}

func TestStrip_OutputIsGofmtFormatted(t *testing.T) {
	src := `package p

import "fmt"

type T struct {
	Name string // comment
	Age  int
}

func F(t T) {
	if t.Age > 0 {
		fmt.Println(t.Name)
	}
}
`
	out, err := StripComments([]byte(src))
	if err != nil {
		t.Fatalf("StripComments returned error: %v", err)
	}

	formatted, err := format.Source(out)
	if err != nil {
		t.Fatalf("format.Source on stripped output failed: %v", err)
	}

	if string(out) != string(formatted) {
		t.Errorf("stripped output is not gofmt-formatted.\ngot:\n%s\nwant:\n%s", out, formatted)
	}
}

func TestStripFile_OverwritesInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	src := "package p\n\nfunc F() {\n\tx := 1 // comment\n\t_ = x\n}\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile setup failed: %v", err)
	}

	if err := StripFile(path); err != nil {
		t.Fatalf("StripFile returned error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after processing failed: %v", err)
	}
	if strings.Contains(string(got), "comment") {
		t.Errorf("expected file on disk to have comment stripped, got:\n%s", got)
	}
	if !strings.Contains(string(got), "x := 1") {
		t.Errorf("expected code to survive on disk, got:\n%s", got)
	}
}
