package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestMain represents a test wrapper for Main that records output.
type TestMain struct {
	*Main
	Stdin  bytes.Buffer
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// NewMain returns a new instance of Main.
func NewTestMain() *TestMain {
	m := &TestMain{Main: NewMain()}
	m.Main.ExecutableName = "typokiller"
	m.Main.Stdin = &m.Stdin
	m.Main.Stdout = &m.Stdout
	m.Main.Stderr = &m.Stderr
	return m
}

func TestStatus(t *testing.T) {
	// Create test directory.
	d, _ := ioutil.TempDir("", "typokiller-")
	defer os.RemoveAll(d)

	m := NewTestMain()
	m.DatabasePath = filepath.Join(d, defaultDatabasePath)

	// Status before Init is an error.
	if err := m.Status(); err != ErrNoProject {
		t.Fatalf("unexpected error: %v\n", err)
	}

	// Initialize project.
	if err := m.Init("My Project"); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	// Expected status of empty project.
	var b bytes.Buffer
	if err := statusTmpl.Execute(&b, status{
		ExecutableName: m.ExecutableName,
		ProjectName:    "My Project",
	}); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	// Check status of empty project.
	if err := m.Status(); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if got, want := m.Stdout.String(), b.String(); got != want {
		t.Fatalf("'typokiller status' returned:\n%s\nwant:\n%s\n", got, want)
	}
}
