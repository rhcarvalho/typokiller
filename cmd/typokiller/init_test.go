package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	// Create test directory.
	d, _ := ioutil.TempDir("", "typokiller-")
	defer os.RemoveAll(d)

	m := NewTestMain()
	m.DatabasePath = filepath.Join(d, defaultDatabasePath)

	// Initialize project.
	if err := m.Init("My Project"); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if stdout := m.Stdout.String(); stdout != "" {
		t.Fatalf("'typokiller init' should not write to stdout, got:\n%s\n", stdout)
	}
	if stderr := m.Stderr.String(); stderr != "" {
		t.Fatalf("'typokiller init' should not write to stderr, got:\n%s\n", stderr)
	}

	// Reinitialize project.
	if err := m.Init("My Project 2"); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if stdout := m.Stdout.String(); stdout != "" {
		t.Fatalf("'typokiller init' should not write to stdout, got:\n%s\n", stdout)
	}
	if stderr := m.Stderr.String(); stderr != "" {
		t.Fatalf("'typokiller init' should not write to stderr, got:\n%s\n", stderr)
	}
}
