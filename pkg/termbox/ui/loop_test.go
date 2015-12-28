package ui

import (
	"testing"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/rhcarvalho/typokiller/pkg/termbox/widgets"
)

type testWidget struct {
	// calls []string
}

func (tw *testWidget) Render(x, y, w, h int) {
	// tw.calls = append(tw.calls,
	// 	fmt.Sprintf("Render(%v, %v, %v, %v)", x, y, w, h))
}

func (tw *testWidget) Handle(e termbox.Event) (widgets.Widget, error) {
	// tw.calls = append(tw.calls,
	// 	fmt.Sprintf("Handle(%#v)", e))
	return tw, nil
}

// testBox implements termboxer.
type testBox struct{}

func (testBox) Init() error { return nil }
func (testBox) Close()      {}

func (testBox) SetCursor(x, y int) {}
func (testBox) HideCursor()        {}

func (testBox) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {}
func (testBox) Size() (width int, height int)                       { return 80, 24 }
func (testBox) Clear(fg, bg termbox.Attribute) error                { return nil }
func (testBox) Flush() error                                        { return nil }

func (testBox) PollEvent() termbox.Event { return termbox.Event{Type: termbox.EventNone} }
func (testBox) Interrupt()               {}

func (testBox) SetInputMode(mode termbox.InputMode) termbox.InputMode {
	return mode
}
func (testBox) SetOutputMode(mode termbox.OutputMode) termbox.OutputMode {
	return mode
}

func TestStopImmediately(t *testing.T) {
	loop := loop{}
	done := make(chan bool)
	go func() {
		loop.Stop()
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Millisecond):
		t.Errorf("test took too long")
	}
}

func TestOnlyOneLoop(t *testing.T) {
	widget := &testWidget{}
	loop := loop{
		termbox: testBox{},
		view:    widget,
	}

	errs := make(chan error, 2)
	go func() { errs <- loop.Start() }()
	go func() { errs <- loop.Start() }()

	done := make(chan bool)
	go func() {
		if err := <-errs; err != ErrAlreadyRunning {
			t.Errorf("loop.Start() = %v; want %v", err, ErrAlreadyRunning)
		}
		loop.Stop()
		<-errs
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Millisecond):
		t.Errorf("test took too long")
	}
}
