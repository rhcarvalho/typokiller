package widgets

import "github.com/nsf/termbox-go"

type Widget interface {
	Render(x, y, w, h int)
	Handle(e termbox.Event) (Widget, error)
}
