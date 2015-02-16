// typokiller is your interactive tool to find and fix typos.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"

	docopt "github.com/docopt/docopt-go"
	"github.com/rhcarvalho/typokiller"
)

func main() {
	usage := `Usage:
  typokiller read [options] PATH ...
  typokiller fix

Interactive tool to find and fix typos in codebases.

Options:
  -h --help     Show this usage help
  --format=EXT  Document format [default: go]
  --version     Show version

Commands:
  read       For each PATH, read the documentation of Go packages and outputs metadata to STDOUT
  fix        Reads spelling error information from STDIN and allows for interative patching

Available formats:
  go         Go source code
  adoc       AsciiDoc documents
`
	arguments, _ := docopt.Parse(usage, nil, true, "typokiller 0.2", false)

	var err error
	if arguments["fix"].(bool) {
		err = Fix()
	} else {
		format := arguments["--format"].(string)
		err = Read(format, arguments["PATH"].([]string)...)
	}
	if err != nil {
		if pe, ok := err.(*os.PathError); ok && pe.Err == syscall.EPIPE {
			// ignore broken pipe
		} else {
			log.Fatalln("error:", err)
		}
	}
}

// Read reads the documentation in paths and outputs metadata to STDOUT.
func Read(format string, paths ...string) error {
	enc := json.NewEncoder(os.Stdout)
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		var readDirer typokiller.ReadDirer
		switch format {
		case "adoc":
			readDirer = typokiller.AsciiDocFormat{}
		default:
			readDirer = typokiller.GoFormat{}
		}
		pkgs, err := readDirer.ReadDir(path)
		if err != nil {
			return err
		}
		for _, pkg := range pkgs {
			err = enc.Encode(pkg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Fix reads documentation metadata from STDIN and presents an interactive user
// interface to perform actions on potential misspells.
func Fix() error {
	misspellings := make(chan *typokiller.Misspelling)
	errs := make(chan error)

	// read STDIN in a new goroutine
	go func() {
		defer close(misspellings)
		defer close(errs)

		reader := bufio.NewReaderSize(os.Stdin, 64*1024*1024) // 64 MB
		var err error

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			var pkg *typokiller.Package
			if err = json.Unmarshal(line, &pkg); err != nil {
				errs <- fmt.Errorf("parsing '%s': %v", line, err)
				continue
			}

			for _, text := range pkg.Documentation {
				text.Package = pkg
				for _, misspelling := range text.Misspellings {
					misspelling.Text = text
					misspellings <- misspelling
				}
			}
		}

		if err != nil && err != io.EOF {
			errs <- err
			return
		}
	}()

	return typokiller.Fix(misspellings, errs)
}
