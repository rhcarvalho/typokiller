package box

import "github.com/nsf/termbox-go"

// termboxer is an interface covering the parts of the termbox API used in this
// package.
type termboxer interface {
	// Initialization.
	Init() error
	Close()

	// Cursor.
	SetCursor(x, y int)
	HideCursor()

	// Back buffer.
	SetCell(x, y int, ch rune, fg, bg termbox.Attribute)
	CellBuffer() []termbox.Cell
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

// tb implements termboxer.
type tb struct{}

func (tb) Init() error { return termbox.Init() }
func (tb) Close()      { termbox.Close() }

func (tb) SetCursor(x, y int) { termbox.SetCursor(x, y) }
func (tb) HideCursor()        { termbox.HideCursor() }

func (tb) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) { termbox.SetCell(x, y, ch, fg, bg) }
func (tb) CellBuffer() []termbox.Cell                          { return termbox.CellBuffer() }
func (tb) Size() (width int, height int)                       { return termbox.Size() }
func (tb) Clear(fg, bg termbox.Attribute) error                { return termbox.Clear(fg, bg) }
func (tb) Flush() error                                        { return termbox.Flush() }

func (tb) PollEvent() termbox.Event { return termbox.PollEvent() }
func (tb) Interrupt()               { termbox.Interrupt() }

func (tb) SetInputMode(mode termbox.InputMode) termbox.InputMode {
	return termbox.SetInputMode(mode)
}
func (tb) SetOutputMode(mode termbox.OutputMode) termbox.OutputMode {
	return termbox.SetOutputMode(mode)
}
