package box

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

// testBox implements termboxer.
type testBox struct {
	w, h     int
	backBuf  []termbox.Cell
	frontBuf []termbox.Cell
	// log stores all method calls.
	log []string
}

func (tb *testBox) Init() error {
	tb.log = append(tb.log, "Init()")
	tb.backBuf = make([]termbox.Cell, tb.w*tb.h)
	tb.frontBuf = make([]termbox.Cell, tb.w*tb.h)
	return nil
}
func (tb *testBox) Close() {
	tb.log = append(tb.log, "Close()")
}

func (tb *testBox) SetCursor(x, y int) {
	tb.log = append(tb.log, fmt.Sprintf("SetCursor(%v, %v)", x, y))
}
func (tb *testBox) HideCursor() {
	tb.log = append(tb.log, "HideCursor()")
}

func (tb *testBox) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	tb.log = append(tb.log, fmt.Sprintf("SetCell(%v, %v, %v, %v, %v)", x, y, ch, fg, bg))
	w, _ := tb.Size()
	tb.backBuf[x+y*w] = termbox.Cell{Ch: ch, Fg: fg, Bg: bg}
}
func (tb *testBox) CellBuffer() []termbox.Cell {
	tb.log = append(tb.log, "CellBuffer()")
	return tb.backBuf
}
func (tb *testBox) Size() (width int, height int) {
	tb.log = append(tb.log, "Size()")
	return tb.w, tb.h
}
func (tb *testBox) Clear(fg, bg termbox.Attribute) error {
	tb.log = append(tb.log, fmt.Sprintf("Clear(%v, %v)", fg, bg))
	for _, c := range tb.backBuf {
		c.Fg = fg
		c.Bg = bg
	}
	return nil
}
func (tb *testBox) Flush() error {
	tb.log = append(tb.log, "Flush()")
	copy(tb.frontBuf, tb.backBuf)
	return nil
}

func (tb *testBox) PollEvent() termbox.Event {
	tb.log = append(tb.log, "PollEvent()")
	return termbox.Event{Type: termbox.EventNone}
}
func (tb *testBox) Interrupt() {
	tb.log = append(tb.log, "Interrupt()")
}

func (tb *testBox) SetInputMode(mode termbox.InputMode) termbox.InputMode {
	tb.log = append(tb.log, fmt.Sprintf("SetInputMode(%v)", mode))
	return mode
}
func (tb *testBox) SetOutputMode(mode termbox.OutputMode) termbox.OutputMode {
	tb.log = append(tb.log, fmt.Sprintf("SetOutputMode(%v)", mode))
	return mode
}
