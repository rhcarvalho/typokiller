// typokiller is your interactive tool to find and fix typos.
package main

import (
	"bufio"
	"encoding/json"
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
	arguments, _ := docopt.Parse(usage, nil, true, "typokiller 0.1", false)

	if arguments["fix"].(bool) {
		Fix()
	} else {
		Read(arguments["PATH"].([]string)...)
	}
}

// Read reads the documentation of Go packages in paths and outputs metadata to STDOUT.
func Read(paths ...string) {
	enc := json.NewEncoder(os.Stdout)
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Fatalln(err)
		}
		for _, pkg := range typokiller.ReadDir(path) {
			enc.Encode(pkg)
		}
	}
}

// Fix reads documentation metadata from STDIN and presents an interactive user
// interface to perform actions on potential misspells.
func Fix() {
	misspellings := make(chan *typokiller.Misspelling)

	// read STDIN in a new goroutine
	go func() {
		reader := bufio.NewReaderSize(os.Stdin, 64*1024*1024) // 64 MB
		var err error
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			var pkg *typokiller.Package
			if err = json.Unmarshal(line, &pkg); err != nil {
				log.Fatalf("error: %v\nline: %s\n", err, line)
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
			log.Fatalln(err)
		}
		close(misspellings)
	}()

	typokiller.Fix(misspellings)
}
