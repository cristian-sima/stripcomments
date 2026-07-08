// Package stripcomments removes comments from Go source while preserving the
// ones that carry meaning. Compiler and tool directives (//go:embed,
// //go:generate, //go:build, the legacy // +build constraint, and similar)
// are always kept. By default the doc-comment block above every exported
// identifier and the package doc are kept too; the DropExported and
// DropPackage options remove those as well. Everything else — body comments,
// trailing inline comments, floating comments, and unexported docs — is
// stripped. Output is re-emitted gofmt-formatted.
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

// StripComments removes every non-directive comment from the given Go source
// and returns gofmt-formatted output. It returns an error if the source does
// not parse.
func StripComments(src []byte, opts ...Option) ([]byte, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	file.Comments = keepDirectiveGroups(file.Comments)
	if !cfg.preservePackage {
		file.Doc = keepDirectiveGroup(file.Doc)
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.Field:
			if !(cfg.preserveExported && fieldExported(typed)) {
				typed.Doc = keepDirectiveGroup(typed.Doc)
			}
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.ImportSpec:
			typed.Doc = keepDirectiveGroup(typed.Doc)
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.ValueSpec:
			if !(cfg.preserveExported && anyExported(typed.Names)) {
				typed.Doc = keepDirectiveGroup(typed.Doc)
			}
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.TypeSpec:
			if !(cfg.preserveExported && typed.Name.IsExported()) {
				typed.Doc = keepDirectiveGroup(typed.Doc)
			}
			typed.Comment = keepDirectiveGroup(typed.Comment)
		case *ast.GenDecl:
			if !(cfg.preserveExported && genDeclHasExported(typed)) {
				typed.Doc = keepDirectiveGroup(typed.Doc)
			}
		case *ast.FuncDecl:
			if !(cfg.preserveExported && typed.Name.IsExported()) {
				typed.Doc = keepDirectiveGroup(typed.Doc)
			}
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
// preserving the file's existing mode. It applies the same options as
// StripComments.
func StripFile(path string, opts ...Option) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	stripped, err := StripComments(src, opts...)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, stripped, info.Mode())
}

func anyExported(names []*ast.Ident) bool {
	for _, name := range names {
		if name.IsExported() {
			return true
		}
	}
	return false
}

func fieldExported(field *ast.Field) bool {
	if len(field.Names) == 0 {
		if ident := embeddedIdent(field.Type); ident != nil {
			return ident.IsExported()
		}
		return false
	}
	return anyExported(field.Names)
}

func embeddedIdent(expr ast.Expr) *ast.Ident {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed
	case *ast.StarExpr:
		return embeddedIdent(typed.X)
	case *ast.SelectorExpr:
		return typed.Sel
	}
	return nil
}

func genDeclHasExported(decl *ast.GenDecl) bool {
	for _, spec := range decl.Specs {
		switch typed := spec.(type) {
		case *ast.ValueSpec:
			if anyExported(typed.Names) {
				return true
			}
		case *ast.TypeSpec:
			if typed.Name.IsExported() {
				return true
			}
		}
	}
	return false
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
