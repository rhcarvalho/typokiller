package ui

import "github.com/nsf/termbox-go"

var tb termboxer = defaultTermbox{}

// termboxer is an interface covering the parts of the termbox API used in
// typokiller.
type termboxer interface {
	// Initialization.
	Init() error
	Close()

	// Cursor.
	SetCursor(x, y int)
	HideCursor()

	// Back buffer.
	SetCell(x, y int, ch rune, fg, bg termbox.Attribute)
	// CellBuffer() []termbox.Cell
	Size() (width int, height int)
	Clear(fg, bg termbox.Attribute) error
	Flush() error
	// Sync() error

	// Events.
	PollEvent() termbox.Event
	Interrupt()
	// ParseEvent(data []byte) termbox.Event
	// PollRawEvent(data []byte) termbox.Event

	// IO mode.
	SetInputMode(mode termbox.InputMode) termbox.InputMode
	SetOutputMode(mode termbox.OutputMode) termbox.OutputMode
}

// defaultTermbox implements termboxer.
type defaultTermbox struct{}

func (defaultTermbox) Init() error { return termbox.Init() }
func (defaultTermbox) Close()      { termbox.Close() }

func (defaultTermbox) SetCursor(x, y int) { termbox.SetCursor(x, y) }
func (defaultTermbox) HideCursor()        { termbox.HideCursor() }

func (defaultTermbox) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x, y, ch, fg, bg)
}
func (defaultTermbox) Size() (width int, height int)        { return termbox.Size() }
func (defaultTermbox) Clear(fg, bg termbox.Attribute) error { return termbox.Clear(fg, bg) }
func (defaultTermbox) Flush() error                         { return termbox.Flush() }

func (defaultTermbox) PollEvent() termbox.Event { return termbox.PollEvent() }
func (defaultTermbox) Interrupt()               { termbox.Interrupt() }

func (defaultTermbox) SetInputMode(mode termbox.InputMode) termbox.InputMode {
	return termbox.SetInputMode(mode)
}
func (defaultTermbox) SetOutputMode(mode termbox.OutputMode) termbox.OutputMode {
	return termbox.SetOutputMode(mode)
}
