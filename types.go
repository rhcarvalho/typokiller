package typokiller

// Package holds the documentation of a Go package and a list of identifiers.
// The identifiers are useful to avoid false positives when spellchecking the
// documentation.
// type Package struct {
// 	Name          string `json:"PackageName"`
// 	Identifiers   []string
// 	Documentation []*Text
// }
//
// // Text holds some documentation text.
// type Text struct {
// 	Content      string
// 	Position     token.Position
// 	Misspellings []*Misspelling
// 	Package      *Package `json:"-"`
// }
//
// // Misspelling holds information about a potential misspell.
// type Misspelling struct {
// 	Word        string
// 	Offset      int
// 	Suggestions []string
// 	Action      Action
// 	Text        *Text `json:"-"`
// }
//
// // Action represents the user action towards a misspell.
// type Action struct {
// 	Type        ActionType
// 	Replacement string
// }
//
// // ActionType is one of Undefined, Ignore or Replace.
// type ActionType int
//
// const (
// 	Undefined ActionType = iota
// 	Ignore
// 	Replace
// )

// -----------------------

// Project is a spellchecking project in typokiller.
type Project struct {
	Files    map[string]*File `json:"files"`
	Metadata *Metadata        `json:"metadata"`
}

// NewProject creates an empty Project with all fields initialized.
func NewProject() *Project {
	return &Project{
		Files:    make(map[string]*File),
		Metadata: &Metadata{},
	}
}

// FIXME
// **** Missing place to store text that will be passed to the spellchecker ****

// A File holds a list of potential Misspellings in the file at Path, and a list
// of Replacements that should take place to fix the misspellings.
type File struct {
	Path         string         `json:"-"`
	Metadata     *Metadata      `json:"metadata"`
	Fragments    []*Fragment    `json:"fragments"`
	Misspellings []*Misspelling `json:"misspellings,omitempty"`
	Replacements []*Replacement `json:"replacements,omitempty"`
}

type Fragment struct {
	Text   string `json:"text"`
	File   *File  `json:"-"`
	Offset int    `json:"offset"` // file offset
}

// A Misspelling represents a potentially misspelled Term at a given Offset of
// File. Processed is true after this term has been ignored or replaced, perhaps
// with one of the Suggestions.
type Misspelling struct {
	Term        string   `json:"term"`
	File        *File    `json:"-"`
	Offset      int      `json:"offset"` // file offset
	Suggestions []string `json:"suggestions"`
	Processed   bool     `json:"processed"`
}

// A Replacement defines an action to replace a substring of File, from From
// to To, with the string With.
type Replacement struct {
	File *File  `json:"-"`
	From int    `json:"from"`
	To   int    `json:"to"`
	With string `json:"with"`
}

// Metadata includes project or file-wide metadata.
type Metadata struct {
	// Ignored is a list of terms to be ignored when spellchecking.
	Ignored []string `json:"ignored"`
}
