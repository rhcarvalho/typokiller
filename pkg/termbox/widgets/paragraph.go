package widgets

import "github.com/nsf/termbox-go"

type Paragraph struct {
	Text    string
	handler func(Widget, termbox.Event) (Widget, bool)
}

func NewParagraph(text string) Paragraph {
	return Paragraph{
		Text: text,
	}
}

func (p Paragraph) Render(x, y, w, h int) {
	if i, j, overflow := p.render(x, y, w, h); overflow {
		termbox.SetCell(i, j, '…', termbox.ColorWhite, termbox.ColorRed)
	}
}

func (p Paragraph) render(x, y, w, h int) (i, j int, overflow bool) {
	i, j = x, y
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
	if i >= x+w {
		j++
		i = x
	}
	if overflow {
		i, j = x+w-1, y+h-1
	}
	return
}

func (p Paragraph) Bind(f func(w Widget, e termbox.Event) (Widget, bool)) Widget {
	p.handler = f
	return p
}

func (p Paragraph) Handle(e termbox.Event) (Widget, bool) {
	if p.handler != nil {
		return p.handler(p, e)
	}
	return p, false
}
