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
	for i, c := range p.Text {
		termbox.SetCell(x+i, y, c, termbox.ColorDefault, termbox.ColorDefault)
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
