package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/rhcarvalho/typokiller/pkg/types"
)

func TestAdd(t *testing.T) {
	// Create test directory.
	d, _ := ioutil.TempDir("", "typokiller-")
	defer os.RemoveAll(d)

	m := NewTestMain()
	m.DatabasePath = filepath.Join(d, defaultDatabasePath)

	// Initialize project.
	if err := m.Init("My Project"); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	// Add paths.
	if err := m.Add("txt", "/tmp/test/path", "/tmp/typokiller"); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if err := m.Add("adoc", "/tmp/adoc"); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	// Check that paths were stored.
	db, err := m.openDB(false)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	locations := make(map[string]*types.Location)
	expected := map[string]*types.Location{
		"/tmp/test/path": {
			Path:   "/tmp/test/path",
			Format: "txt",
		},
		"/tmp/typokiller": {
			Path:   "/tmp/typokiller",
			Format: "txt",
		},
		"/tmp/adoc": {
			Path:   "/tmp/adoc",
			Format: "adoc",
		},
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(locationsBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			loc := new(types.Location)
			if err := loc.Deserialize(v); err != nil {
				return fmt.Errorf("value %q cannot be decoded: %v", v, err)
			}
			locations[string(k)] = loc
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if !reflect.DeepEqual(locations, expected) {
		t.Fatalf("locations = %v, want %v", locations, expected)
	}
}
