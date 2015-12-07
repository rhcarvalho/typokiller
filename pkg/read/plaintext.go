package read

import (
	"bytes"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rhcarvalho/typokiller/pkg/types"
)

// PlainTextFormat reads plain text files.
type PlainTextFormat struct{}

// ReadDir reads text from files in path. It does not recurse into
// subdirectories.
func (f PlainTextFormat) ReadDir(path string) ([]*types.Package, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var r []*types.Package
	for _, entry := range entries {
		p, err := f.ReadFile(filepath.Join(path, entry.Name()), entry)
		if err != nil {
			return r, err
		}
		r = append(r, p)
	}
	return r, nil
}

// ReadFile extracts paragraphs text files.
func (f PlainTextFormat) ReadFile(path string, fi os.FileInfo) (*types.Package, error) {
	doc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p := &types.Package{Name: fi.Name()}
	offset := 0
	for _, paragraph := range bytes.SplitAfter(doc, []byte("\n\n")) {
		p.Documentation = append(p.Documentation, &types.Text{
			Content: string(paragraph),
			Position: token.Position{
				Filename: path,
				Offset:   offset,
				Line:     bytes.Count(doc[:offset], []byte("\n")) + 1,
				Column:   1,
			},
		})
		offset += len(paragraph)
	}
	return p, nil
}
