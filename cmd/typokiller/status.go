package main

import (
	"text/template"

	"github.com/boltdb/bolt"
)

const statusTxt = `Project: {{.ProjectName}}
Locations: (empty)
Progress: 0% (fixed 0 out of 0 typos)

Use '{{.ExecutableName}} add' to add locations to this project.
`

var statusTmpl = template.Must(template.New("status").Parse(statusTxt))

type status struct {
	ExecutableName string
	ProjectName    string
}

// Status prints information about the project to m.Stdout.
func (m *Main) Status() error {
	db, err := m.openDB(false)
	if err != nil {
		return err
	}
	defer db.Close()

	var name string
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(projectBucket))
		if b == nil {
			return ErrInvalidProject
		}
		name = string(b.Get([]byte("name")))
		return nil
	})

	return statusTmpl.Execute(m.Stdout, status{
		m.ExecutableName,
		name,
	})
}
