package ui

import (
	"errors"
	"sync"

	"github.com/nsf/termbox-go"
	"github.com/rhcarvalho/typokiller/pkg/termbox/widgets"
)

// Errors returned by Loop.
var (
	ErrExit           = errors.New("exit")
	ErrNoView         = errors.New("no view to be rendered")
	ErrAlreadyRunning = errors.New("a loop is already running")
)

// Loop represents a UI loop that renders the current state on the screen and
// handle events.
type Loop struct {
	termbox termboxer
	view    widgets.Widget
	handler func(Loop, termbox.Event) bool
}

// NewLoop returns a new UI loop. Call its Start() method to start it.
func NewLoop(view widgets.Widget) Loop {
	return Loop{
		termbox: tb{},
		view:    view,
	}
}

var (
	// semaphore ensures there's only one loop running at a time.
	semaphore = make(chan struct{}, 1)
	// stop signals a running loop to stop.
	stop = make(chan struct{})
)

// Start starts the UI loop. There can only be one loop running at a time.
func (l Loop) Start() error {
	select {
	case semaphore <- struct{}{}:
		// Semaphore acquired, no other loop is running.
		defer func() {
			<-semaphore
		}()
	default:
		// Semaphore is busy, return error.
		return ErrAlreadyRunning
	}

	l.setup()
	defer l.teardown()

	var wg sync.WaitGroup
	defer wg.Wait()

	events := make(chan termbox.Event)
	// Continuously poll events from termbox
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			select {
			case <-stop:
				close(events)
				return
			case events <- l.termbox.PollEvent():
			}
		}
	}()

	for {
		err := l.render()
		if err != nil {
			return err
		}
		e, ok := <-events
		if !ok {
			return nil
		}
		if l, err = l.handle(e); err != nil {
			if err == ErrExit {
				return nil
			}
			return err
		}
	}
}

// Stop stops the loop. It returns immediately if no loop is running or blocks
// until a running loop handles the stop signal.
func (l Loop) Stop() {
	select {
	case semaphore <- struct{}{}:
		// Semaphore acquired, no loop is running.
		defer func() {
			<-semaphore
		}()
	default:
		// Semaphore is busy, interrupt pending termbox.PollEvent and
		// send stop signal.
		l.termbox.Interrupt()
		stop <- struct{}{}
	}

}

func (l Loop) Bind(f func(Loop, termbox.Event) bool) Loop {
	l.handler = f
	return l
}

func (l Loop) setup() {
	// Initialize termbox.
	if err := l.termbox.Init(); err != nil {
		panic(err)
	}
	l.termbox.HideCursor()
	l.termbox.SetOutputMode(termbox.Output256)
}

func (l Loop) render() error {
	if l.view == nil {
		return ErrNoView
	}
	l.termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer l.termbox.Flush()
	w, h := l.termbox.Size()
	l.view.Render(0, 0, w, h)
	return nil
}

func (l Loop) teardown() {
	l.termbox.Close()
}

func (l Loop) handle(e termbox.Event) (Loop, error) {
	if l.handler != nil {
		if !l.handler(l, e) {
			return l, nil
		}
	}
	if l.view == nil {
		return l, ErrNoView
	}
	var err error
	l.view, err = l.view.Handle(e)
	return l, err
}
