package ui

import (
	"testing"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/rhcarvalho/typokiller/pkg/termbox/widgets"
)

func TestStopImmediately(t *testing.T) {
	loop := Loop{}
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
	loop := NewLoop(testWidget{})

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

type testWidget struct{}

func (tw testWidget) Render(x, y, w, h int) {}
func (tw testWidget) Bind(f func(w widgets.Widget, e termbox.Event) (widgets.Widget, bool)) widgets.Widget {
	return tw
}
func (tw testWidget) Handle(e termbox.Event) (widgets.Widget, bool) { return tw, false }
