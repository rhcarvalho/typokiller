package main

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/rhcarvalho/typokiller/pkg/read"
)

// Read reads the documentation in paths and outputs metadata to STDOUT.
func (m *Main) Read(format string, paths ...string) error {
	if err := m.Add(format, paths...); err != nil {
		return err
	}

	db, err := m.openDB(false)
	if err != nil {
		return err
	}
	defer db.Close()

	enc := json.NewEncoder(m.Stdout)

	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(locationsBucket))
		if b == nil {
			return errors.New("no locations were added")
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			path, err := filepath.Abs(string(k))
			if err != nil {
				return err
			}
			var dirReader read.DirReader
			switch format {
			case "adoc":
				dirReader = read.AsciiDocFormat{}
			case "go":
				dirReader = read.GoFormat{}
			default:
				dirReader = read.PlainTextFormat{}
			}
			pkgs, err := dirReader.ReadDir(path)
			if err != nil {
				return err
			}
			for _, pkg := range pkgs {
				err = enc.Encode(pkg)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
