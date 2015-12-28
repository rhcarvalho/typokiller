package widgets

import "github.com/nsf/termbox-go"

type Widget interface {
	Render(x, y, w, h int)
	Bind(f func(w Widget, e termbox.Event) (Widget, error)) Widget
	Handle(e termbox.Event) (Widget, error)
}
