// Command stripcomments strips comments from Go source files in place,
// preserving compiler and tool directives. By default it also keeps the doc
// comments above exported identifiers and the package doc; pass -drop-exported
// or -drop-package to remove those too. Each non-flag argument is a path to a
// .go file, overwritten with its comment-stripped, gofmt'd form.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cristian-sima/stripcomments"
)

func main() {
	dropExported := flag.Bool("drop-exported", false, "also strip doc comments above exported identifiers")
	dropPackage := flag.Bool("drop-package", false, "also strip the package doc comment")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: stripcomments [-drop-exported] [-drop-package] <file.go> [more files...]")
		flag.PrintDefaults()
	}
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var opts []stripcomments.Option
	if *dropExported {
		opts = append(opts, stripcomments.DropExported())
	}
	if *dropPackage {
		opts = append(opts, stripcomments.DropPackage())
	}

	for _, path := range files {
		if err := stripcomments.StripFile(path, opts...); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Println(path)
	}
}
