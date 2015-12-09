package main

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/rhcarvalho/typokiller/pkg/read"
	"github.com/rhcarvalho/typokiller/pkg/types"
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

	jsonEnc := json.NewEncoder(m.Stdout)

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(locationsBucket))
		if b == nil {
			return errors.New("no locations were added")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			loc := new(types.Location)
			if err := loc.Deserialize(v); err != nil {
				return err
			}

			if loc.IsRead {
				if err := encodePackages(jsonEnc, loc.Packages); err != nil {
					return err
				}
				continue
			}

			// TODO do not convert path to absolute.
			path, err := filepath.Abs(loc.Path)
			if err != nil {
				return err
			}

			var dirReader read.DirReader
			switch loc.Format {
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

			loc.IsRead = true
			loc.Packages = pkgs

			v, err := loc.Serialize()
			if err != nil {
				return err
			}
			if err := b.Put(k, v); err != nil {
				return err
			}

			if err := encodePackages(jsonEnc, pkgs); err != nil {
				return err
			}
		}
		return nil
	})
}

func encodePackages(enc *json.Encoder, pkgs []*types.Package) error {
	for _, pkg := range pkgs {
		if err := enc.Encode(pkg); err != nil {
			return err
		}
	}
	return nil
}
