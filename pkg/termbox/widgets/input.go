package widgets

import "github.com/nsf/termbox-go"

// Input is a text input widget.
type Input struct {
	Label   string
	value   []rune
	pos     int
	handler func(Widget, termbox.Event) (Widget, bool)
}

func NewInput(label string, value string) Input {
	return Input{
		Label: label,
		value: []rune(value),
		pos:   len(value),
	}
}

func (in Input) Render(x, y, w, h int) {
	var i, j int
	var c rune
	for i, c = range in.Label {
		termbox.SetCell(x+i, y, c, termbox.ColorDefault, termbox.ColorDefault)
	}
	if len(in.Label) > 0 {
		i += 2
	}
	for j, c = range in.value {
		termbox.SetCell(x+i+j, y, c, termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.SetCursor(x+i+in.pos, y)
}

func (in Input) Bind(f func(w Widget, e termbox.Event) (Widget, bool)) Widget {
	in.handler = f
	return in
}

func (in Input) Handle(e termbox.Event) (Widget, bool) {
	switch e.Type {
	case termbox.EventKey:
		switch e.Key {
		case termbox.KeyArrowLeft, termbox.KeyCtrlB:
			return in.moveCursorBackward(), true
		case termbox.KeyArrowRight, termbox.KeyCtrlF:
			return in.moveCursorForward(), true
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			return in.deleteRuneBackward(), true
		case termbox.KeyDelete, termbox.KeyCtrlD:
			return in.deleteRuneForward(), true
		case termbox.KeyCtrlK:
			return in.deleteAllForward(), true
		case termbox.KeyHome, termbox.KeyCtrlA:
			return in.moveCursorStart(), true
		case termbox.KeyEnd, termbox.KeyCtrlE:
			return in.moveCursorEnd(), true
		case termbox.KeySpace:
			return in.insertRune(' '), true
		default:
			if e.Ch != 0 {
				return in.insertRune(e.Ch), true
			}
		}
	}

	if in.handler != nil {
		return in.handler(in, e)
	}
	return in, false
}

func (in Input) Value() string {
	return string(in.value)
}

func (in Input) insertRune(r rune) Widget {
	in.value = append(in.value[:in.pos], append([]rune{r}, in.value[in.pos:]...)...)
	in.pos++
	return in
}

func (in Input) deleteRuneBackward() Widget {
	if in.pos > 0 && len(in.value) > 0 {
		copy(in.value[in.pos-1:], in.value[in.pos:])
		in.value = in.value[:len(in.value)-1]
	}
	return in.moveCursorBackward()
}

func (in Input) deleteRuneForward() Widget {
	if in.pos < len(in.value) {
		copy(in.value[in.pos:], in.value[in.pos+1:])
		in.value = in.value[:len(in.value)-1]
	}
	return in
}

func (in Input) deleteAllForward() Widget {
	if in.pos < len(in.value) {
		copy(in.value[in.pos:], in.value[in.pos+1:])
		in.value = in.value[:in.pos]
	}
	return in
}

func (in Input) moveCursorBackward() Widget {
	in.pos--
	if in.pos < 0 {
		in.pos = 0
	}
	return in
}

func (in Input) moveCursorForward() Widget {
	in.pos++
	if in.pos > len(in.value) {
		in.pos = len(in.value)
	}
	return in
}

func (in Input) moveCursorStart() Widget {
	in.pos = 0
	return in
}

func (in Input) moveCursorEnd() Widget {
	in.pos = len(in.value)
	return in
}
