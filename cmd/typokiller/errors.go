package main

import "errors"

var (
	// ErrNoProject is returned when typokiller is used before a project has been
	// initialized.
	ErrNoProject = errors.New("not a typokiller project")

	// ErrInvalidProject is returned when typokiller's project data is invalid.
	ErrInvalidProject = errors.New("the current project is invalid")
)
