package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"
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

	// Check that paths were stored.
	db, err := m.openDB(false)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	paths := make(map[string]string)
	expected := map[string]string{"/tmp/test/path": "txt", "/tmp/typokiller": "txt"}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(locationsBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			paths[string(k)] = string(v)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if !reflect.DeepEqual(paths, expected) {
		t.Fatalf("paths = %v, want %v", paths, expected)
	}
}
