package stripcomments

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStrip_InlineTrailingComment(t *testing.T) {
	Convey("Given source with an inline trailing comment", t, func() {
		src := `package p

func F() {
	x := 1 // comment
	_ = x
}
`
		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			Convey("Then the comment is stripped", func() {
				So(string(out), ShouldNotContainSubstring, "comment")
			})
			Convey("Then the code line survives", func() {
				So(string(out), ShouldContainSubstring, "x := 1")
			})
		})
	})
}

func TestStrip_StandaloneAndBlockComments(t *testing.T) {
	Convey("Given source with a doc comment and a block comment", t, func() {
		src := `package p

// doc comment
func f() {
	/* block comment */
	x := 1
	_ = x
}
`
		Convey("When stripped with DropUnexported", func() {
			out, err := StripComments([]byte(src), DropUnexported())
			So(err, ShouldBeNil)

			Convey("Then all comments are stripped", func() {
				So(string(out), ShouldNotContainSubstring, "comment")
			})
		})
	})
}

func TestStrip_StringLiteralsUntouched(t *testing.T) {
	Convey("Given source with comment-like content inside string literals", t, func() {
		src := "package p\n\n" +
			`func F() {` + "\n" +
			"\turl := \"http://example.com\" // trailing comment\n" +
			"\tblock := \"/* not a comment */\"\n" +
			"\traw := `has // and /* inside */ too`\n" +
			"\t_ = url\n\t_ = block\n\t_ = raw\n}\n"

		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the string literal with // survives", func() {
				So(got, ShouldContainSubstring, `"http://example.com"`)
			})
			Convey("Then the string literal with /* */ survives", func() {
				So(got, ShouldContainSubstring, `"/* not a comment */"`)
			})
			Convey("Then the raw string literal survives untouched", func() {
				So(got, ShouldContainSubstring, "`has // and /* inside */ too`")
			})
			Convey("Then the real trailing comment is stripped", func() {
				So(got, ShouldNotContainSubstring, "trailing comment")
			})
		})
	})
}

func TestStrip_PreservesBuildConstraints(t *testing.T) {
	Convey("Given source with build constraints and a regular comment", t, func() {
		src := `//go:build windows
// +build windows

package p

func F() {
	x := 1 // comment
	_ = x
}
`
		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the //go:build constraint survives", func() {
				So(got, ShouldContainSubstring, "//go:build windows")
			})
			Convey("Then the legacy +build constraint survives", func() {
				So(got, ShouldContainSubstring, "// +build windows")
			})
			Convey("Then the regular comment is still stripped", func() {
				So(got, ShouldNotContainSubstring, "// comment")
			})
		})
	})
}

func TestStrip_PreservesGoEmbedDirective(t *testing.T) {
	Convey("Given source with a go:embed directive and a regular comment", t, func() {
		src := `package p

import "embed"

//go:embed data/*.json
var schemaFiles embed.FS

func F() {
	x := 1 // comment
	_ = x
}
`
		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the //go:embed directive survives", func() {
				So(got, ShouldContainSubstring, "//go:embed data/*.json")
			})
			Convey("Then //go:embed appears exactly once", func() {
				So(strings.Count(got, "//go:embed"), ShouldEqual, 1)
			})
			Convey("Then the regular comment is still stripped", func() {
				So(got, ShouldNotContainSubstring, "// comment")
			})
		})
	})
}

func TestStrip_PreservesDirectiveAdjacentToDocComment(t *testing.T) {
	Convey("Given source with a doc comment directly above a go:embed directive", t, func() {
		src := `package p

import "embed"

// schemaFiles holds every schema on disk.
//go:embed data/*.json
var schemaFiles embed.FS
`
		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the //go:embed directive survives", func() {
				So(got, ShouldContainSubstring, "//go:embed data/*.json")
			})
			Convey("Then the adjacent doc comment is still stripped", func() {
				So(got, ShouldNotContainSubstring, "holds every schema")
			})
		})
	})
}

func TestStrip_PreservesDirectiveAfterDocParagraphAndBlankCommentLine(t *testing.T) {
	Convey("Given source with a doc paragraph, blank comment line, and two go:embed directives", t, func() {
		src := `package p

// tessPatterns is bundled into the binary so no setup is required on
// the host.
//
//go:embed tessdata/patterns.txt
var tessPatterns []byte

//go:embed tessdata/words.txt
var tessWords []byte
`
		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			got := string(out)

			Convey("Then the first //go:embed directive survives", func() {
				So(got, ShouldContainSubstring, "//go:embed tessdata/patterns.txt")
			})
			Convey("Then the second //go:embed directive survives", func() {
				So(got, ShouldContainSubstring, "//go:embed tessdata/words.txt")
			})
			Convey("Then the doc paragraph is still stripped", func() {
				So(got, ShouldNotContainSubstring, "bundled into the binary")
			})
			Convey("Then the stripped output is gofmt-formatted", func() {
				formatted, formatErr := format.Source(out)
				So(formatErr, ShouldBeNil)
				So(string(out), ShouldEqual, string(formatted))
			})
		})
	})
}

func TestStrip_OutputIsGofmtFormatted(t *testing.T) {
	Convey("Given source with a struct field trailing comment", t, func() {
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
		Convey("When stripped", func() {
			out, err := StripComments([]byte(src))
			So(err, ShouldBeNil)

			Convey("Then the stripped output is gofmt-formatted", func() {
				formatted, formatErr := format.Source(out)
				So(formatErr, ShouldBeNil)
				So(string(out), ShouldEqual, string(formatted))
			})
		})
	})
}

func TestStripFile_OverwritesInPlace(t *testing.T) {
	Convey("Given a file on disk with an inline comment", t, func() {
		dir := t.TempDir()
		path := filepath.Join(dir, "sample.go")
		src := "package p\n\nfunc F() {\n\tx := 1 // comment\n\t_ = x\n}\n"
		writeErr := os.WriteFile(path, []byte(src), 0o644)
		So(writeErr, ShouldBeNil)

		Convey("When StripFile processes it", func() {
			err := StripFile(path)
			So(err, ShouldBeNil)

			got, readErr := os.ReadFile(path)
			So(readErr, ShouldBeNil)

			Convey("Then the comment is stripped on disk", func() {
				So(string(got), ShouldNotContainSubstring, "comment")
			})
			Convey("Then the code survives on disk", func() {
				So(string(got), ShouldContainSubstring, "x := 1")
			})
		})
	})
}
