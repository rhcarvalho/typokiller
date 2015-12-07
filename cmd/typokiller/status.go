package main

import (
	"text/template"

	"github.com/boltdb/bolt"
)

const statusTxt = `Project: {{.ProjectName}}
Locations:{{if not .Locations}} (empty){{else}}{{range .Locations}}
	{{.}}{{end}}
{{end}}
Progress: 0% (fixed 0 out of 0 typos)

{{if .Locations}}No files were read yet.
{{else}}Use '{{.ExecutableName}} add' to add locations to this project.
{{end}}`

var statusTmpl = template.Must(template.New("status").Parse(statusTxt))

type status struct {
	ExecutableName string
	ProjectName    string
	Locations      []string
}

// Status prints information about the project to m.Stdout.
func (m *Main) Status() error {
	db, err := m.openDB(false)
	if err != nil {
		return err
	}
	defer db.Close()

	var (
		name      string
		locations []string
	)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(projectBucket))
		if b == nil {
			return ErrInvalidProject
		}
		name = string(b.Get([]byte("name")))
		b = tx.Bucket([]byte(locationsBucket))
		if b == nil {
			// No locations were added.
			return nil
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			locations = append(locations, string(k))
		}
		return nil
	})

	return statusTmpl.Execute(m.Stdout, status{
		ExecutableName: m.ExecutableName,
		ProjectName:    name,
		Locations:      locations,
	})
}
