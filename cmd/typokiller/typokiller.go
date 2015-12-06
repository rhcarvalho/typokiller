// typokiller is your interactive tool to find and fix typos.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/rhcarvalho/typokiller/pkg/fix"
	"github.com/rhcarvalho/typokiller/pkg/read"
	"github.com/rhcarvalho/typokiller/pkg/types"
)

// Main represents the main program execution.
type Main struct {
	ExecutableName string
	DatabasePath   string
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
}

// NewMain returns a new instance of Main connect to the standard input/output.
func NewMain() *Main {
	return &Main{
		ExecutableName: os.Args[0],
		DatabasePath:   defaultDatabasePath,
		Stdin:          os.Stdin,
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
	}
}

func main() {
	usage := `Exterminate typos. Now.

typokiller is a tool to find and fix typos in text files, source code, and documentation.

Usage:
  typokiller init [NAME]
  typokiller status
  typokiller read [options] PATH ...
  typokiller fix

Options:
  -h --help     Show this usage help
  --format=EXT  Document format [default: go]
  --version     Show version

Commands:
  init       Initializes a new typo-hunting project
  status     Reports project status
  read       For each PATH, read the documentation of Go packages and outputs metadata to STDOUT
  fix        Reads spelling error information from STDIN and allows for interative patching

Available formats:
  go         Go source code
  adoc       AsciiDoc documents
`
	args, _ := docopt.Parse(usage, nil, true, "typokiller 0.3", false)

	main := NewMain()
	var err error

	switch {
	case args["init"]:
		name, _ := args["NAME"].(string)
		err = main.Init(name)
	case args["status"]:
		err = main.Status()
	case args["read"]:
		format := args["--format"].(string)
		err = main.Read(format, args["PATH"].([]string)...)
	case args["fix"]:
		err = main.Fix()
	default:
		fmt.Fprintln(main.Stdout, usage)
		os.Exit(1)
	}

	if err != nil {
		if pe, ok := err.(*os.PathError); ok && pe.Err == syscall.EPIPE {
			// ignore broken pipe
		} else {
			fmt.Fprintf(main.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}
}

// Read reads the documentation in paths and outputs metadata to STDOUT.
func (m *Main) Read(format string, paths ...string) error {
	enc := json.NewEncoder(os.Stdout)
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		var dirReader read.DirReader
		switch format {
		case "adoc":
			dirReader = read.AsciiDocFormat{}
		default:
			dirReader = read.GoFormat{}
		}
		pkgs, err := dirReader.ReadDir(path)
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
func (m *Main) Fix() error {
	misspellings := make(chan *types.Misspelling)
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

			var pkg *types.Package
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

	return fix.Fix(misspellings, errs)
}
