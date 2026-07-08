// Command stripcomments strips comments from Go source files in place,
// preserving compiler and tool directives. Each argument is a path to a .go
// file; every named file is overwritten with its comment-stripped, gofmt'd
// form.
package main

import (
	"fmt"
	"os"

	"github.com/cristian-sima/stripcomments"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: stripcomments <file.go> [more files...]")
		os.Exit(1)
	}

	for _, path := range os.Args[1:] {
		if err := stripcomments.StripFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Println(path)
	}
}
