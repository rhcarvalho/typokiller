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
  typokiller scan PATH ...
  typokiller fix

Find comments in Go source files and interactively fix typos.

Options:
  -h --help  Show this usage help
  --version  Show version

Commands:
  scan       Outputs comments for the packages found
  fix        Reads spelling error information from STDIN and allows for interative patching`
	arguments, _ := docopt.Parse(usage, nil, true, "typokiller 0.1", false)

	// fix typos mode
	if arguments["fix"].(bool) {
		reader := bufio.NewReaderSize(os.Stdin, 64*1024*1024) // 64 MB
		var spellingErrors []*typokiller.SpellingError
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

			for _, c := range pkg.Comments {
				c.Package = pkg
				for _, s := range c.SpellingErrors {
					s.Comment = c
					spellingErrors = append(spellingErrors, s)
				}
			}
		}
		if err != nil && err != io.EOF {
			log.Fatalln(err)
		}
		typokiller.IFix(spellingErrors)
	}

	// scan comments mode
	enc := json.NewEncoder(os.Stdout)
	for _, path := range arguments["PATH"].([]string) {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Fatalln(err)
		}
		for _, pkg := range typokiller.ReadDir(path) {
			enc.Encode(pkg)
		}
	}
}
