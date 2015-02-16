package typokiller

import (
	"bytes"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
)

// AsciiDocFormat can read documentation from AsciiDoc files.
type AsciiDocFormat struct{}

// ReadDir extracts AsciiDoc-formatted documentation from files in path.
// It does not recurse into subdirectories.
func (f AsciiDocFormat) ReadDir(path string) ([]*Package, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var r []*Package
	for _, entry := range entries {
		if f.IsAsciiDocFile(entry) {
			p, err := f.ReadFile(filepath.Join(path, entry.Name()), entry)
			if err != nil {
				return r, err
			}
			r = append(r, p)
		}
	}
	return r, nil
}

// ReadFile extracts paragraphs from AsciiDoc files.
func (f AsciiDocFormat) ReadFile(path string, fi os.FileInfo) (*Package, error) {
	doc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p := &Package{Name: fi.Name()}
	offset := 0
	for _, paragraph := range bytes.SplitAfter(doc, []byte("\n\n")) {
		p.Documentation = append(p.Documentation, &Text{
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

// IsAsciiDocFile returns true for AsciiDoc files, false otherwise.
func (f AsciiDocFormat) IsAsciiDocFile(fi os.FileInfo) bool {
	return fi.Mode().IsRegular() && filepath.Ext(fi.Name()) == ".adoc"
}
