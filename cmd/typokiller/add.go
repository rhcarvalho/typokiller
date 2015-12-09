package main

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/rhcarvalho/typokiller/pkg/types"
)

// Add adds the files and directories in the list of paths to the current
// project.
func (m *Main) Add(format string, paths ...string) error {
	db, err := m.openDB(false)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(locationsBucket))
		if err != nil {
			return fmt.Errorf("add location: %s", err)
		}
		for _, path := range paths {
			loc := types.Location{
				Path:   path,
				Format: format,
			}
			v, err := loc.Serialize()
			if err != nil {
				return err
			}
			if err := b.Put([]byte(path), v); err != nil {
				return err
			}
		}
		return nil
	})
}
