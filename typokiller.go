package typokiller

import "go/token"

// Package holds the comments of a Go package and a list of identifiers.
// The identifiers are useful to avoid false positives when spellchecking the
// comments.
type Package struct {
	Name        string `json:"PackageName"`
	Identifiers []string
	Comments    []*Comment
}

// Comment holds the text of a comment and its detailed position.
type Comment struct {
	Text           string
	Position       token.Position
	SpellingErrors []*SpellingError
	Package        *Package
}

type SpellingError struct {
	Word        string
	Offset      int
	Suggestions []string
	Action      *Action
	Comment     *Comment
}

type Action struct {
	Type        ActionType
	Replacement string
}

type ActionType int

const (
	Undefined ActionType = iota
	Ignore
	Replace
)
