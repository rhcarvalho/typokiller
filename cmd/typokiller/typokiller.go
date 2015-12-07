// typokiller is your interactive tool to find and fix typos.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/rhcarvalho/typokiller/pkg/fix"
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
  typokiller add [--format=txt|go|adoc] PATH ...
  typokiller read [--format=txt|go|adoc] [PATH ...]
  typokiller fix

Options:
  -h --help     Show this usage help
  --version     Show version

Commands:
  init       Initializes a new typo-hunting project
  status     Reports project status
  add        Add files or directories to the current project
  read       For each PATH, read the documentation of Go packages and outputs metadata to STDOUT
  fix        Reads spelling error information from STDIN and allows for interative patching

Available formats:
  txt        Plain text (default)
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
	case args["add"]:
		format, _ := args["--format"].(string)
		paths := args["PATH"].([]string)
		err = main.Add(format, paths...)
	case args["read"]:
		format, _ := args["--format"].(string)
		paths := args["PATH"].([]string)
		err = main.Read(format, paths...)
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
