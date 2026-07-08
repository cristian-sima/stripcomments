package stripcomments

// Option configures which comments StripComments drops. By default
// StripComments preserves the doc-comment block above every exported
// identifier and the package doc; use the Drop options to remove them too.
type Option func(*config)

type config struct {
	preserveExported bool
	preservePackage  bool
}

func defaultConfig() config {
	return config{preserveExported: true, preservePackage: true}
}

// DropExported also strips the doc-comment block above exported identifiers
// (func, type, const, var, and exported struct fields / interface methods),
// which are preserved by default.
func DropExported() Option {
	return func(c *config) { c.preserveExported = false }
}

// DropPackage also strips the package doc block (the "// Package ..." comment
// above the package clause), which is preserved by default.
func DropPackage() Option {
	return func(c *config) { c.preservePackage = false }
}
