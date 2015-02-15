package typokiller

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	termbox "github.com/nsf/termbox-go"
)

// Fix turns the terminal into an interactive UI for fixing typos.
func Fix(misspellings <-chan *Misspelling, errs <-chan error) error {
	ui := NewFixUI()

	// read misspellings channel in a goroutine
	go func() {
		for misspelling := range misspellings {
			ui.Misspellings = append(ui.Misspellings, misspelling)
			ui.Draw()
		}
		ui.DoneLoadingInput = true
		ui.Draw()
	}()

	// block this goroutine in the UI mainloop
	return ui.Mainloop(errs)
}

// FixUI has the state necessary in the UI.
type FixUI struct {
	Misspellings     []*Misspelling
	Index            int
	Printer          *TermboxPrinter
	DoneLoadingInput bool
}

// NewFixUI creates a new FixUI.
func NewFixUI() *FixUI {
	ui := &FixUI{
		Printer: NewTermboxPrinter(5, 3, 5, 3),
	}
	return ui
}

// Mainloop draws the current state in the terminal and waits for user input.
func (ui *FixUI) Mainloop(errs <-chan error) error {
	// initialize termbox
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()
	termbox.HideCursor()
	termbox.SetOutputMode(termbox.Output256)
	ui.Draw()

	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	// loop until there's an upstream error or user request to quit
	for {
		select {
		case err, ok := <-errs:
			if ok {
				return err
			}
		case ev := <-events:
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					return nil
				case termbox.KeyArrowUp, termbox.KeyArrowRight:
					ui.Next()
				case termbox.KeyArrowDown, termbox.KeyArrowLeft:
					ui.Previous()
				default:
					switch ev.Ch {
					case 'i':
						ui.Ignore()
					case 'I':
						ui.IgnoreAll()
					case 'r':
						ui.Replace()
					case 'R':
						ui.ReplaceAll()
					case 'e':
						ui.Edit()
					case 'E':
						ui.EditAll()
					case 'n':
						ui.NextUndefined()
					case 'a':
						ui.Apply()
					case 'q':
						return nil
					}
				}
			case termbox.EventError:
				return ev.Err
			}
			ui.Draw()
		}
	}
}

// Write implements the io.Writer interface.
func (ui *FixUI) Write(p []byte) (int, error) {
	return ui.Printer.Write(p)
}

// Next advances to the next misspell.
func (ui *FixUI) Next() {
	ui.Index++
	if !(ui.Index < len(ui.Misspellings)) {
		ui.Index = 0
	}
}

// NextUndefined advances to the next misspell that has an Undefined action.
func (ui *FixUI) NextUndefined() {
	start := ui.Index
	for {
		ui.Index++
		if !(ui.Index < len(ui.Misspellings)) {
			ui.Index = 0
		}
		m := ui.Misspellings[ui.Index]
		if m.Action.Type == Undefined {
			break
		}
		if ui.Index == start {
			ui.Printer.fg = termbox.ColorGreen
			fmt.Fprintln(ui, "all done")
			break
		}
	}
}

// Previous goes back to the previous misspell.
func (ui *FixUI) Previous() {
	ui.Index--
	if ui.Index < 0 {
		ui.Index = len(ui.Misspellings) - 1
	}
}

// Ignore ignores the current misspell.
func (ui *FixUI) Ignore() {
	m := ui.Misspellings[ui.Index]
	m.Action = Action{Type: Ignore}
	ui.NextUndefined()
}

// Replace replaces the current misspell with a suggestion.
func (ui *FixUI) Replace() {
	m := ui.Misspellings[ui.Index]
	m.Action = Action{
		Type:        Replace,
		Replacement: m.Suggestions[ui.ReadIntegerInRange(1, len(m.Suggestions))-1]}
	ui.NextUndefined()
}

// Edit replaces the current misspell with custom text.
func (ui *FixUI) Edit() {
	m := ui.Misspellings[ui.Index]
	m.Action = Action{
		Type:        Replace,
		Replacement: ui.ReadString()}
	ui.NextUndefined()
}

// IgnoreAll ignores all misspells with Undefined action that matches the
// current word.
func (ui *FixUI) IgnoreAll() {
	m := ui.Misspellings[ui.Index]
	word := m.Word
	for i := ui.Index; i < len(ui.Misspellings); i++ {
		m = ui.Misspellings[i]
		if m.Word == word && m.Action.Type == Undefined {
			m.Action = Action{Type: Ignore}
		}
	}
	ui.NextUndefined()
}

// ReplaceAll replaces all occurrences of the current word with a suggestion.
func (ui *FixUI) ReplaceAll() {
	m := ui.Misspellings[ui.Index]
	word := m.Word
	replacement := m.Suggestions[ui.ReadIntegerInRange(1, len(m.Suggestions))-1]
	for i := ui.Index; i < len(ui.Misspellings); i++ {
		m = ui.Misspellings[i]
		if m.Word == word && m.Action.Type == Undefined {
			m.Action = Action{Type: Replace, Replacement: replacement}
		}
	}
	ui.NextUndefined()
}

// EditAll replaces all occurrences of the current word with custom text.
func (ui *FixUI) EditAll() {
	m := ui.Misspellings[ui.Index]
	word := m.Word
	replacement := ui.ReadString()
	for i := ui.Index; i < len(ui.Misspellings); i++ {
		m = ui.Misspellings[i]
		if m.Word == word && m.Action.Type == Undefined {
			m.Action = Action{Type: Replace, Replacement: replacement}
		}
	}
	ui.NextUndefined()
}

// Apply applies marked changes to disk.
func (ui *FixUI) Apply() {
	defer termbox.PollEvent() // stay visible until user presses a key
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	ui.DrawBorders()
	ui.Printer.Reset()
	ui.Printer.fg = termbox.ColorGreen
	fmt.Fprintln(ui, "applying changes")
	termbox.Flush()

	status := make(chan string)
	go Apply(ui.Misspellings, status)

	ui.Printer.fg = termbox.ColorYellow
	for s := range status {
		fmt.Fprint(ui, s)
		termbox.Flush()
	}
	fmt.Fprint(ui, "\ndone")
	ui.Printer.ResetColors()
	termbox.Flush()
}

// ReadIntegerInRange interactively reads an integer within the range [a, b].
func (ui *FixUI) ReadIntegerInRange(a, b int) int {
start:
	ui.Printer.fg |= termbox.AttrBold
	fmt.Fprintf(ui, "\nenter number in range [%d, %d]: ", a, b)
	ui.Printer.ResetColors()
	termbox.Flush()
	ui.Printer.fg = termbox.ColorMagenta
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
				fmt.Fprint(ui, string(ev.Ch))
				termbox.Flush()
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
	ui.Printer.ResetColors()
	i, err := strconv.Atoi(string(v))
	if err != nil || !(a <= i && i <= b) {
		fmt.Fprintln(ui, " → invalid number, try again")
		termbox.Flush()
		v = nil
		goto start
	}
	return i
}

// ReadString interactively reads an arbitrary string.
func (ui *FixUI) ReadString() string {
	ui.Printer.fg |= termbox.AttrBold
	fmt.Fprint(ui, "\nreplace with: ")
	ui.Printer.ResetColors()
	termbox.Flush()
	ui.Printer.fg = termbox.ColorMagenta
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
					ui.Printer.x--
					fmt.Fprint(ui, " ")
					ui.Printer.x--
					termbox.Flush()
				}
			default:
				if unicode.IsPrint(ev.Ch) || ev.Key == termbox.KeySpace {
					v = append(v, ev.Ch)
					fmt.Fprint(ui, string(ev.Ch))
					termbox.Flush()
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
	ui.Printer.ResetColors()
	return string(v)
}

// Draw draws the current state of the UI.
func (ui *FixUI) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	ui.DrawBorders()

	tp := ui.Printer
	tp.Reset()

	if len(ui.Misspellings) == 0 {
		if ui.DoneLoadingInput {
			fmt.Fprint(ui, "No spelling errors!")
		} else {
			fmt.Fprint(ui, "Loading data...")
		}
		return
	}

	fmt.Fprint(ui, "Spelling error ")
	tp.Bold()
	fmt.Fprintf(ui, "%d", ui.Index+1)
	tp.ResetColors()
	fmt.Fprint(ui, " of ")
	tp.Bold()
	fmt.Fprintf(ui, "%d", len(ui.Misspellings))
	if !ui.DoneLoadingInput {
		fmt.Fprint(ui, "+")
	}
	tp.ResetColors()

	m := ui.Misspellings[ui.Index]
	text := m.Text

	tp.SkipLines(2)
	tp.Underline()
	fmt.Fprintf(ui, "%s:%d:%d\n", text.Position.Filename, text.Position.Line, text.Position.Column)

	tp.SkipLines(1)
	tmp := tp.y
	tp.fg = 0xf0
	fmt.Fprintln(ui, text.Content)
	tmp, tp.y = tp.y, tmp

	tp.x += m.Offset - strings.LastIndex(text.Content[:m.Offset], "\n") - 1
	tp.y += strings.Count(text.Content[:m.Offset], "\n")
	tp.fg = termbox.ColorRed | termbox.AttrBold
	fmt.Fprint(ui, m.Word)
	tp.ResetColors()
	tp.y = tmp

	tp.SkipLines(1)
	fmt.Fprint(ui, "Suggestions: ")
	for i, suggestion := range m.Suggestions {
		if i > 0 {
			fmt.Fprint(ui, ", ")
		}
		fmt.Fprintf(ui, "[%d] %s", i+1, suggestion)
	}
	tp.SkipLines(2)

	fmt.Fprint(ui, "Actions: ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "r")
	tp.ResetColors()
	fmt.Fprint(ui, "eplace, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "R")
	tp.ResetColors()
	fmt.Fprint(ui, "eplace all, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "i")
	tp.ResetColors()
	fmt.Fprint(ui, "gnore, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "I")
	tp.ResetColors()
	fmt.Fprint(ui, "gnore all, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "e")
	tp.ResetColors()
	fmt.Fprint(ui, "dit, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "E")
	tp.ResetColors()
	fmt.Fprint(ui, "dit all, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "n")
	tp.ResetColors()
	fmt.Fprint(ui, "ext undefined, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "a")
	tp.ResetColors()
	fmt.Fprint(ui, "pply, ")
	tp.fg |= termbox.AttrUnderline
	fmt.Fprint(ui, "q")
	tp.ResetColors()
	fmt.Fprintln(ui, "uit")

	if m.Action.Type != Undefined {
		tp.SkipLines(1)
		tp.fg = termbox.ColorBlue
		switch m.Action.Type {
		case Ignore:
			fmt.Fprintln(ui, "ignored")
		case Replace:
			fmt.Fprintf(ui, "replace with '%s'\n", m.Action.Replacement)
		}
		tp.ResetColors()
	}
}

// DrawBorders draws a rectangular border around the screen.
func (ui *FixUI) DrawBorders() {
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
