package typokiller

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	termbox "github.com/nsf/termbox-go"
)

func Fix(misspellings <-chan *Misspelling) {
	ui := NewFixUI()

	// read misspellings channel in a goroutine
	go func() {
		for misspelling := range misspellings {
			ui.Misspellings = append(ui.Misspellings, misspelling)
			ui.Draw()
		}
	}()

	// block this goroutine in the UI mainloop
	ui.Mainloop()
}

type FixUI struct {
	Misspellings []*Misspelling
	Index        int
	Printer      *TermboxPrinter
}

func NewFixUI() *FixUI {
	ui := &FixUI{
		Printer: NewTermboxPrinter(5, 5, 5, 5),
	}
	return ui
}

func (t *FixUI) Mainloop() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.HideCursor()
	termbox.SetOutputMode(termbox.Output256)
	t.Draw()

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyArrowUp, termbox.KeyArrowRight:
				t.Next()
			case termbox.KeyArrowDown, termbox.KeyArrowLeft:
				t.Previous()
			default:
				switch ev.Ch {
				case 'i':
					t.Ignore()
				case 'I':
					t.IgnoreAll()
				case 'r':
					t.Replace()
				case 'R':
					t.ReplaceAll()
				case 'e':
					t.Edit()
				case 'E':
					t.EditAll()
				case 'n':
					t.NextUndefined()
				case 'a':
					t.Apply()
				case 'q':
					break mainloop
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		t.Draw()
	}
}

// Write implements the io.Writer interface.
func (t *FixUI) Write(p []byte) (int, error) {
	return t.Printer.Write(p)
}

func (t *FixUI) Next() {
	t.Index++
	if !(t.Index < len(t.Misspellings)) {
		t.Index = 0
	}
}

func (t *FixUI) NextUndefined() {
	start := t.Index
	for {
		t.Index++
		if !(t.Index < len(t.Misspellings)) {
			t.Index = 0
		}
		m := t.Misspellings[t.Index]
		if m.Action.Type == Undefined {
			break
		}
		if t.Index == start {
			t.Printer.fg = termbox.ColorGreen
			fmt.Fprintln(t, "all done")
			break
		}
	}
}

func (t *FixUI) Previous() {
	t.Index--
	if t.Index < 0 {
		t.Index = len(t.Misspellings) - 1
	}
}

func (t *FixUI) Ignore() {
	m := t.Misspellings[t.Index]
	m.Action = Action{Type: Ignore}
	t.NextUndefined()
}

func (t *FixUI) Replace() {
	m := t.Misspellings[t.Index]
	m.Action = Action{
		Type:        Replace,
		Replacement: m.Suggestions[t.ReadIntegerInRange(1, len(m.Suggestions))-1]}
	t.NextUndefined()
}

func (t *FixUI) Edit() {
	m := t.Misspellings[t.Index]
	m.Action = Action{
		Type:        Replace,
		Replacement: t.ReadString()}
	t.NextUndefined()
}

func (t *FixUI) IgnoreAll() {
	m := t.Misspellings[t.Index]
	word := m.Word
	for i := t.Index; i < len(t.Misspellings); i++ {
		m = t.Misspellings[i]
		if m.Word == word && m.Action.Type == Undefined {
			m.Action = Action{Type: Ignore}
		}
	}
	t.NextUndefined()
}

func (t *FixUI) ReplaceAll() {
	m := t.Misspellings[t.Index]
	word := m.Word
	replacement := m.Suggestions[t.ReadIntegerInRange(1, len(m.Suggestions))-1]
	for i := t.Index; i < len(t.Misspellings); i++ {
		m = t.Misspellings[i]
		if m.Word == word && m.Action.Type == Undefined {
			m.Action = Action{Type: Replace, Replacement: replacement}
		}
	}
	t.NextUndefined()
}

func (t *FixUI) EditAll() {
	m := t.Misspellings[t.Index]
	word := m.Word
	replacement := t.ReadString()
	for i := t.Index; i < len(t.Misspellings); i++ {
		m = t.Misspellings[i]
		if m.Word == word && m.Action.Type == Undefined {
			m.Action = Action{Type: Replace, Replacement: replacement}
		}
	}
	t.NextUndefined()
}

func (t *FixUI) Apply() {
	defer termbox.PollEvent() // stay visible until user presses a key
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	t.DrawBorders()
	t.Printer.Reset()
	t.Printer.fg = termbox.ColorGreen
	fmt.Fprint(t, "applying changes")
	termbox.Flush()

	status := make(chan string)
	go Apply(t.Misspellings, status)

	t.Printer.fg = termbox.ColorYellow
	for s := range status {
		fmt.Fprint(t, s)
		termbox.Flush()
	}
	fmt.Fprint(t, "done")
	t.Printer.ResetColors()
	termbox.Flush()
}

func (t *FixUI) ReadIntegerInRange(a, b int) int {
start:
	t.Printer.fg |= termbox.AttrBold
	fmt.Fprintf(t, "\nenter number in range [%d, %d]: ", a, b)
	t.Printer.ResetColors()
	termbox.Flush()
	t.Printer.fg = termbox.ColorMagenta
	var v []rune
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEnter {
				break mainloop
			}
			if ev.Ch >= '0' && ev.Ch <= '9' {
				v = append(v, ev.Ch)
				fmt.Fprint(t, string(ev.Ch))
				termbox.Flush()
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
	t.Printer.ResetColors()
	i, err := strconv.Atoi(string(v))
	if err != nil || !(a <= i && i <= b) {
		fmt.Fprintln(t, " → invalid number, try again")
		termbox.Flush()
		v = nil
		goto start
	}
	return i
}

func (t *FixUI) ReadString() string {
	t.Printer.fg |= termbox.AttrBold
	fmt.Fprint(t, "\nreplace with: ")
	t.Printer.ResetColors()
	termbox.Flush()
	t.Printer.fg = termbox.ColorMagenta
	var v []rune
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEnter:
				break mainloop
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if len(v) > 0 {
					v = v[:len(v)-1]
					t.Printer.x--
					fmt.Fprint(t, " ")
					t.Printer.x--
					termbox.Flush()
				}
			default:
				if unicode.IsPrint(ev.Ch) || ev.Key == termbox.KeySpace {
					v = append(v, ev.Ch)
					fmt.Fprint(t, string(ev.Ch))
					termbox.Flush()
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
	t.Printer.ResetColors()
	return string(v)
}

func (t *FixUI) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	t.DrawBorders()

	tp := t.Printer
	tp.Reset()

	if len(t.Misspellings) == 0 {
		fmt.Fprint(t, "No spelling errors!")
		return
	}

	fmt.Fprint(t, "Spelling error ")
	tp.Bold()
	fmt.Fprintf(t, "%d", t.Index+1)
	tp.ResetColors()
	fmt.Fprint(t, " of ")
	tp.Bold()
	fmt.Fprintf(t, "%d", len(t.Misspellings))
	tp.ResetColors()

	m := t.Misspellings[t.Index]
	text := m.Text

	tp.SkipLines(2)
	tp.Underline()
	fmt.Fprintf(t, "%s:%d:%d\n", text.Position.Filename, text.Position.Line, text.Position.Column)

	tp.SkipLines(1)
	tmp := tp.y
	tp.fg = 0xf0
	fmt.Fprintln(t, text.Content)
	tmp, tp.y = tp.y, tmp

	tp.x += m.Offset - strings.LastIndex(text.Content[:m.Offset], "\n") - 1
	tp.y += strings.Count(text.Content[:m.Offset], "\n")
	tp.fg = termbox.ColorRed | termbox.AttrBold
	fmt.Fprint(t, m.Word)
	tp.ResetColors()
	tp.y = tmp

	tp.SkipLines(1)
	fmt.Fprint(t, "Suggestions: ")
	for i, suggestion := range m.Suggestions {
		if i > 0 {
			fmt.Fprint(t, ", ")
		}
		fmt.Fprintf(t, "[%d] %s", i+1, suggestion)
	}
	tp.SkipLines(2)

	fmt.Fprint(t, "Actions: ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "r")
	tp.ResetColors()
	fmt.Fprint(t, "eplace, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "R")
	tp.ResetColors()
	fmt.Fprint(t, "eplace all, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "i")
	tp.ResetColors()
	fmt.Fprint(t, "gnore, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "I")
	tp.ResetColors()
	fmt.Fprint(t, "gnore all, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "e")
	tp.ResetColors()
	fmt.Fprint(t, "dit, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "E")
	tp.ResetColors()
	fmt.Fprint(t, "dit all, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "n")
	tp.ResetColors()
	fmt.Fprint(t, "ext undefined, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "a")
	tp.ResetColors()
	fmt.Fprint(t, "pply, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(t, "q")
	tp.ResetColors()
	fmt.Fprintln(t, "uit")

	if m.Action.Type != Undefined {
		tp.SkipLines(1)
		tp.fg = termbox.ColorBlue
		switch m.Action.Type {
		case Ignore:
			fmt.Fprintln(t, "ignored")
		case Replace:
			fmt.Fprintf(t, "replace with '%s'\n", m.Action.Replacement)
		}
		tp.ResetColors()
	}
}

// DrawBorders draws a rectangular border around the screen.
func (t *FixUI) DrawBorders() {
	x, y := 1, 1
	w, h := termbox.Size()
	c := ' '
	fg := termbox.ColorDefault
	bg := termbox.Attribute(0xf7)
	// draw top and bottom borders
	for i := x; i < w-x; i++ {
		termbox.SetCell(i, y, c, fg, bg)
		termbox.SetCell(i, h-y, c, fg, bg)
	}
	// draw left and right borders
	for j := y; j < h-y; j++ {
		termbox.SetCell(x, j, c, fg, bg)
		termbox.SetCell(x+1, j, c, fg, bg)
		termbox.SetCell(w-x-2, j, c, fg, bg)
		termbox.SetCell(w-x-1, j, c, fg, bg)
	}
}
