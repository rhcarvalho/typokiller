package typokiller

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
)

// ReadDir extracts documentation metadata from Go files in path.
// This includes documentation comments and known identifiers.
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
			begin := fset.Position(c.Pos())
			end := fset.Position(c.End())
			b, err := ioutil.ReadFile(begin.Filename)
			if err != nil {
				panic(err)
			}
			text := string(b[begin.Offset:end.Offset])
			p.Comments = append(p.Comments, &Comment{Text: text, Position: begin})
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
