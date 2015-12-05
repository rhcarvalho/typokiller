package read

import "github.com/rhcarvalho/typokiller/pkg/types"

// DirReader is implemented by any value that has a ReadDir method, which
// defines how to read documentation in a given format.
type DirReader interface {
	ReadDir(string) ([]*types.Package, error)
}
