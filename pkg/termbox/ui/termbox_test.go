package ui

import "github.com/nsf/termbox-go"

func init() {
	tb = testTermbox{}
}

// testTermbox implements termboxer.
type testTermbox struct{}

func (testTermbox) Init() error { return nil }
func (testTermbox) Close()      {}

func (testTermbox) SetCursor(x, y int) {}
func (testTermbox) HideCursor()        {}

func (testTermbox) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {}
func (testTermbox) Size() (width int, height int)                       { return 80, 24 }
func (testTermbox) Clear(fg, bg termbox.Attribute) error                { return nil }
func (testTermbox) Flush() error                                        { return nil }

func (testTermbox) PollEvent() termbox.Event { return termbox.Event{Type: termbox.EventNone} }
func (testTermbox) Interrupt()               {}

func (testTermbox) SetInputMode(mode termbox.InputMode) termbox.InputMode {
	return mode
}
func (testTermbox) SetOutputMode(mode termbox.OutputMode) termbox.OutputMode {
	return mode
}
