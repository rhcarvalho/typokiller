package widgets

import "github.com/nsf/termbox-go"

type Paragraph struct {
	Text    string
	handler func(Widget, termbox.Event) (Widget, error)
}

func NewParagraph(text string) Paragraph {
	return Paragraph{
		Text: text,
	}
}

func (p Paragraph) Render(x, y, w, h int) {
	i, j := x, y
	overflow := false
	for _, c := range p.Text {
		if i >= x+w {
			j++
			i = x
		}
		if j >= y+h {
			overflow = true
			break
		}
		if c == '\n' {
			j++
			i = x
			continue
		}
		termbox.SetCell(i, j, c, termbox.ColorDefault, termbox.ColorDefault)
		i++
	}
	if overflow {
		termbox.SetCell(x+w-1, y+h-1, '…', termbox.ColorWhite, termbox.ColorRed)
	}
}

func (p Paragraph) Bind(f func(w Widget, e termbox.Event) (Widget, error)) Widget {
	p.handler = f
	return p
}

func (p Paragraph) Handle(e termbox.Event) (Widget, error) {
	if p.handler != nil {
		return p.handler(p, e)
	}
	return p, nil
}
