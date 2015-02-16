package typokiller

// ReadDirer is implemented by any value that has a ReadDir method, which
// defines how to read documentation in a given format.
type ReadDirer interface {
	ReadDir(string) ([]*Package, error)
}
