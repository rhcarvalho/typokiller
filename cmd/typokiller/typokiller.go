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

	docopt "github.com/docopt/docopt-go"
	"github.com/rhcarvalho/typokiller"
)

func main() {
	usage := `Usage:
  typokiller read PATH ...
  typokiller fix

Interactive tool to find and fix typos in codebases.

Options:
  -h --help  Show this usage help
  --version  Show version

Commands:
  read       For each PATH, read the documentation of Go packages and outputs metadata to STDOUT
  fix        Reads spelling error information from STDIN and allows for interative patching`
	arguments, _ := docopt.Parse(usage, nil, true, "typokiller 0.2", false)

	var err error
	if arguments["fix"].(bool) {
		err = Fix()
	} else {
		err = Read(arguments["PATH"].([]string)...)
	}
	if err != nil {
		log.Fatalln("error:", err)
	}
}

// Read reads the documentation of Go packages in paths and outputs metadata to STDOUT.
func Read(paths ...string) error {
	enc := json.NewEncoder(os.Stdout)
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		pkgs, err := typokiller.ReadDir(path)
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
