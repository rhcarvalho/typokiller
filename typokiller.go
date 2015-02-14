package typokiller

import "go/token"

// Package holds the documentation of a Go package and a list of identifiers.
// The identifiers are useful to avoid false positives when spellchecking the
// documentation.
type Package struct {
	Name          string `json:"PackageName"`
	Identifiers   []string
	Documentation []*Text
}

// Text holds some documentation text.
type Text struct {
	Content      string
	Position     token.Position
	Misspellings []*Misspelling
	Package      *Package `json:"-"`
}

// Misspelling holds information about a potential misspell.
type Misspelling struct {
	Word        string
	Offset      int
	Suggestions []string
	Action      Action
	Text        *Text `json:"-"`
}

// Action represents the user action towards a misspell.
type Action struct {
	Type        ActionType
	Replacement string
}

// ActionType is one of Undefined, Ignore or Replace.
type ActionType int

const (
	Undefined ActionType = iota
	Ignore
	Replace
)
