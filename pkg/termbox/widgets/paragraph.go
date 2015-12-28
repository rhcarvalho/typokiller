package widgets

import "github.com/nsf/termbox-go"

type Paragraph struct {
	Text string
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

func (p Paragraph) Handle(e termbox.Event) (Widget, error) {
	return p, nil
}
