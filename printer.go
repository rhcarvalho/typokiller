package typokiller

import (
	"fmt"

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

func (tbp *TermboxPrinter) Reset() {
	tbp.x = 0
	tbp.y = 0
	tbp.ResetColors()
}

func (tbp *TermboxPrinter) Print(text string) {
	for _, c := range text {
		if c == '\n' {
			tbp.x = 0
			tbp.y++
			continue
		}
		termbox.SetCell(tbp.left+tbp.x, tbp.top+tbp.y, c, tbp.fg, tbp.bg)
		tbp.x++
		w, _ := termbox.Size()
		if tbp.x >= w-tbp.right-tbp.left-2 {
			termbox.SetCell(tbp.left+tbp.x+1, tbp.top+tbp.y, '⏎', termbox.ColorWhite, termbox.ColorRed)
			tbp.SkipLines(1)
		}
	}
}

func (tbp *TermboxPrinter) Println(text string) {
	tbp.Print(text)
	tbp.Print("\n")
}

func (tbp *TermboxPrinter) Printf(format string, a ...interface{}) {
	tbp.Print(fmt.Sprintf(format, a...))
}

func (tbp *TermboxPrinter) Bold() {
	tbp.fg |= termbox.AttrBold
}

func (tbp *TermboxPrinter) Underline() {
	tbp.fg |= termbox.AttrUnderline
}

func (tbp *TermboxPrinter) ResetColors() {
	tbp.fg = termbox.ColorDefault
	tbp.bg = termbox.ColorDefault
}

func (tbp *TermboxPrinter) SkipLines(n int) {
	tbp.y += n
	tbp.x = 0
}
