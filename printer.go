package typokiller

import (
	"unicode/utf8"

	termbox "github.com/nsf/termbox-go"
)

// TermboxPrinter is an abstraction on top of termbox to facilitate outputting
// text in a text-based terminal.
type TermboxPrinter struct {
	x, y        int               // current cursor position (column, line)
	left, right int               // left and right margins
	top, bottom int               // top and bottom margins
	fg, bg      termbox.Attribute // foreground and background colors
}

// NewTermboxPrinter creates a new TermboxPrinter.
func NewTermboxPrinter(left, top, right, bottom int) *TermboxPrinter {
	return &TermboxPrinter{left: left, top: top, right: right, bottom: bottom}
}

// Reset resets the printer to its initial state.
func (tp *TermboxPrinter) Reset() {
	tp.x = 0
	tp.y = 0
	tp.ResetColors()
}

// ResetColors resets the printer colors to their default values.
func (tp *TermboxPrinter) ResetColors() {
	tp.fg = termbox.ColorDefault
	tp.bg = termbox.ColorDefault
}

// Write implements the io.Writer interface.
func (tp *TermboxPrinter) Write(p []byte) (n int, err error) {
	// TODO reshape bytes to fit screen width before proceeding
	for len(p) > 0 {
		r, size := utf8.DecodeRune(p)
		tp.WriteRune(r)
		p = p[size:]
		n += size
	}
	return
}

// WriteRune prints a single rune in the current printer position and advance
// one character.
func (tp *TermboxPrinter) WriteRune(r rune) (n int, err error) {
	n = utf8.RuneLen(r)
	if r == '\n' {
		tp.NewLine()
		return
	}
	w, _ := termbox.Size()
	maxX := w - tp.right - tp.left - 1
	if tp.x >= maxX {
		// TODO new line rune should be introduced by reshape method,
		// only when needed.
		termbox.SetCell(tp.left+maxX, tp.top+tp.y, '⏎', termbox.ColorWhite, termbox.ColorRed)
		tp.NewLine()
	}
	termbox.SetCell(tp.left+tp.x, tp.top+tp.y, r, tp.fg, tp.bg)
	tp.x++
	return
}

// NewLine advances the printer to the beginning of the next line.
func (tp *TermboxPrinter) NewLine() {
	tp.SkipLines(1)
}

// SkipLines is equivalent to calling NewLine n times.
func (tp *TermboxPrinter) SkipLines(n int) {
	tp.x = 0
	tp.y += n
}

// Bold makes the printer print bold characters.
func (tp *TermboxPrinter) Bold() {
	tp.fg |= termbox.AttrBold
}

// Underline makes the printer print underlined characters.
func (tp *TermboxPrinter) Underline() {
	tp.fg |= termbox.AttrUnderline
}
