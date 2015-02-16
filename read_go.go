package typokiller

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

// GoFormat can read documentation from Go source code.
type GoFormat struct{}

// ReadDir extracts documentation metadata from Go files in path.
// This includes documentation comments and known identifiers.
func (f GoFormat) ReadDir(path string) ([]*Package, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var r []*Package
	for _, pkg := range pkgs {
		p, err := f.ReadPackage(pkg, fset)
		if err != nil {
			return r, err
		}
		r = append(r, p)
	}
	return r, nil
}

// ReadPackage extracts comments of a Go package.
func (GoFormat) ReadPackage(pkg *ast.Package, fset *token.FileSet) (*Package, error) {
	p := &Package{Name: pkg.Name}
	for _, f := range pkg.Files {
		// Collect comments
		for _, c := range f.Comments {
			begin := fset.Position(c.Pos())
			end := fset.Position(c.End())
			b, err := ioutil.ReadFile(begin.Filename)
			if err != nil {
				return nil, err
			}
			content := string(b[begin.Offset:end.Offset])
			p.Documentation = append(p.Documentation, &Text{Content: content, Position: begin})
		}

		// Collect identifiers
		ast.Inspect(pkg, func(n ast.Node) bool {
			if ident, isIdent := n.(*ast.Ident); isIdent {
				p.Identifiers = append(p.Identifiers, ident.String())
			}
			return true
		})
	}
	return p, nil
}
