package main

import (
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const (
	defaultDatabasePath = ".typokiller/project"
	defaultFileMode     = 0664
	defaultDirMode      = 0755
)

const (
	projectBucket = "Project"
)

// Init initializes a typokiller project. It creates a new database in the
// defaultDatabasePath and sets the project name.
// It is valid to reinitialize a project, effectively renaming it.
func (m *Main) Init(name string) error {
	err := os.MkdirAll(".typokiller", defaultDirMode)
	if err != nil {
		return err
	}

	db, err := openDB(true)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(projectBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put([]byte("name"), []byte(name))
		if err != nil {
			return fmt.Errorf("put \"name\": %s", err)
		}
		return nil
	})
}

// openDB opens a Bolt database with default options. ErrNoProject is returned
// if create is false and the defaultDatabasePath does not exist.
func openDB(create bool) (*bolt.DB, error) {
	if !create {
		_, err := os.Stat(defaultDatabasePath)
		if os.IsNotExist(err) {
			return nil, ErrNoProject
		}
	}
	opts := &bolt.Options{
		Timeout: 1 * time.Second,
	}
	return bolt.Open(defaultDatabasePath, defaultFileMode, opts)
}
