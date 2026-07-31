package stripcomments

type Option func(*config)

type config struct {
	preserveExported   bool
	preserveUnexported bool
	preservePackage    bool
}

func defaultConfig() config {
	return config{
		preserveExported:   true,
		preserveUnexported: true,
		preservePackage:    true,
	}
}

func keepsDoc(cfg config, isExported bool) bool {
	if isExported {
		return cfg.preserveExported
	}
	return cfg.preserveUnexported
}

func DropExported() Option {
	return func(c *config) { c.preserveExported = false }
}

func DropUnexported() Option {
	return func(c *config) { c.preserveUnexported = false }
}

func DropPackage() Option {
	return func(c *config) { c.preservePackage = false }
}
