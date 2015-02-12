package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"

	docopt "github.com/docopt/docopt-go"
)

// Package holds the comments of a Go package and a list of identifiers.
// The identifiers are useful to avoid false positives when spellchecking the
// comments.
type Package struct {
	Name        string `json:"PackageName"`
	Identifiers []string
	Comments    []*Comment
}

// Comment holds the text of a comment and its detailed position.
type Comment struct {
	Text     string
	Position token.Position
}

// ReadDir extracts comments of Go files.
func ReadDir(path string) []*Package {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatalln(err)
	}
	var r []*Package
	for name, pkg := range pkgs {
		r = append(r, ReadPackage(name, pkg, fset))
	}
	return r
}

// ReadPackage extracts comments of a Go package.
func ReadPackage(name string, pkg *ast.Package, fset *token.FileSet) *Package {
	p := &Package{Name: name}
	for _, f := range pkg.Files {
		// Collect comments
		for _, c := range f.Comments {
			p.Comments = append(p.Comments, &Comment{c.Text(), fset.Position(c.Pos())})
		}

		// Collect identifiers
		ast.Inspect(pkg, func(n ast.Node) bool {
			if ident, isIdent := n.(*ast.Ident); isIdent {
				p.Identifiers = append(p.Identifiers, ident.String())
			}
			return true
		})
	}
	return p
}

func main() {
	usage := `Usage: typokiller [PATH ...]

Find comments in Go source files.

Options:
  -h --help  Show this usage help
  --version  Show version`
	arguments, _ := docopt.Parse(usage, nil, true, "typokiller 0.1", false)

	w := json.NewEncoder(os.Stdout)
	for _, path := range arguments["PATH"].([]string) {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Fatalln(err)
		}
		for _, pkg := range ReadDir(path) {
			w.Encode(pkg)
		}
	}
}
