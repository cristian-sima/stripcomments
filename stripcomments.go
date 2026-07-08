// Package stripcomments removes comments from Go source while preserving
// compiler and tool directives (//go:embed, //go:generate, //go:build, the
// legacy // +build constraint, and similar). It parses the source with the
// standard go/parser, drops every non-directive comment group, and re-emits
// gofmt-formatted output.
package stripcomments

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
)

// directivePattern matches Go compiler/tool directive comments: //go:embed,
// //go:generate, //go:build, //go:linkname, etc., plus the legacy
// // +build constraint. These look like ordinary comments to go/parser but
// carry real semantic meaning (the go:build/legacy pair) or perform real
// work (go:embed populates the variable below it) — stripping them silently
// breaks the build or ships empty embed.FS values.
var directivePattern = regexp.MustCompile(`^//go:\S|^// \+build\b`)

// Strip removes every non-directive comment from the given Go source and
// returns gofmt-formatted output. It returns an error if the source does not
// parse.
func Strip(src []byte) ([]byte, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	file.Comments = keepDirectiveGroups(file.Comments)
	file.Doc = keepDirectiveGroup(file.Doc)

	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.Field:
			typed.Doc = keepDirectiveGroup(typed.Doc)
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.ImportSpec:
			typed.Doc = keepDirectiveGroup(typed.Doc)
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.ValueSpec:
			typed.Doc = keepDirectiveGroup(typed.Doc)
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.TypeSpec:
			typed.Doc = keepDirectiveGroup(typed.Doc)
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.GenDecl:
			typed.Doc = keepDirectiveGroup(typed.Doc)
		case *ast.FuncDecl:
			typed.Doc = keepDirectiveGroup(typed.Doc)
		}
		return true
	})

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// StripFile strips comments from the file at path and overwrites it in place,
// preserving the file's existing mode.
func StripFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	stripped, err := Strip(src)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, stripped, info.Mode())
}

func keepDirectiveGroups(groups []*ast.CommentGroup) []*ast.CommentGroup {
	var kept []*ast.CommentGroup

	for _, group := range groups {
		if filtered := keepDirectiveGroup(group); filtered != nil {
			kept = append(kept, filtered)
		}
	}

	return kept
}

func keepDirectiveGroup(group *ast.CommentGroup) *ast.CommentGroup {
	if group == nil {
		return nil
	}

	var lines []*ast.Comment
	for _, comment := range group.List {
		if directivePattern.MatchString(comment.Text) {
			lines = append(lines, comment)
		}
	}
	if len(lines) == 0 {
		return nil
	}

	return &ast.CommentGroup{List: lines}
}
