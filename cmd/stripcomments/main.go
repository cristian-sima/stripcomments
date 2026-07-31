package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cristian-sima/stripcomments"
)

func main() {
	dropExported := flag.Bool("drop-exported", false, "also strip doc comments above exported identifiers")
	dropUnexported := flag.Bool("drop-unexported", false, "also strip doc comments above unexported identifiers")
	dropPackage := flag.Bool("drop-package", false, "also strip the package doc comment")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: stripcomments [-drop-exported] [-drop-unexported] [-drop-package] <file.go> [more files...]")
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
	if *dropUnexported {
		opts = append(opts, stripcomments.DropUnexported())
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
