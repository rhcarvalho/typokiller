package typokiller

import (
	"container/heap"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode"

	termbox "github.com/nsf/termbox-go"
)

type TermboxUI struct {
	Misspellings []*Misspelling
	Index        int
	Printer      *TermboxPrinter
}

func NewTermboxUI(misspellings []*Misspelling) *TermboxUI {
	ui := &TermboxUI{
		Misspellings: misspellings,
		Printer:      NewTermboxPrinter(5, 5, 5, 5),
	}
	return ui
}

// Write implements the io.Writer interface.
func (t *TermboxUI) Write(p []byte) (int, error) {
	return t.Printer.Write(p)
}

func (t *TermboxUI) Next() {
	t.Index++
	if !(t.Index < len(t.Misspellings)) {
		t.Index = 0
	}
}

func (t *TermboxUI) NextUndefined() {
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

func (t *TermboxUI) Previous() {
	t.Index--
	if t.Index < 0 {
		t.Index = len(t.Misspellings) - 1
	}
}

func (t *TermboxUI) Ignore() {
	m := t.Misspellings[t.Index]
	m.Action = Action{Type: Ignore}
	t.NextUndefined()
}

func (t *TermboxUI) Replace() {
	m := t.Misspellings[t.Index]
	m.Action = Action{
		Type:        Replace,
		Replacement: m.Suggestions[t.ReadIntegerInRange(1, len(m.Suggestions))-1]}
	t.NextUndefined()
}

func (t *TermboxUI) Edit() {
	m := t.Misspellings[t.Index]
	m.Action = Action{
		Type:        Replace,
		Replacement: t.ReadString()}
	t.NextUndefined()
}

func (t *TermboxUI) IgnoreAll() {
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

func (t *TermboxUI) ReplaceAll() {
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

func (t *TermboxUI) EditAll() {
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

func (t *TermboxUI) Apply() {
	defer termbox.PollEvent()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()
	drawRect(2, 2, w-4, h-4, 0xf7)
	t.Printer.Reset()
	t.Printer.fg = termbox.ColorGreen
	fmt.Fprint(t, "applying changes")
	termbox.Flush()

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(t.Misspellings))
	for i, m := range t.Misspellings {
		pq[i] = &Item{
			value:    m,
			priority: m.Text.Position.Offset + m.Offset,
			index:    i,
		}
	}
	heap.Init(&pq)

	// Take the items out; they arrive in decreasing priority order.
	for max := len(pq); max > 0; max-- {
		item := heap.Pop(&pq).(*Item)
		m := item.value
		if m.Action.Type == Replace {
			t.Printer.fg = termbox.ColorYellow
			fmt.Fprint(t, ".")
			pos := m.Text.Position

			b, err := ioutil.ReadFile(pos.Filename)
			if err != nil {
				panic(err)
			}
			begin := pos.Offset + m.Offset
			end := begin + len(m.Word)
			found := string(b[begin:end])
			if found == m.Word {
				replaced := replaceSlice(b, begin, end, []byte(m.Action.Replacement)...)
				ioutil.WriteFile(pos.Filename, replaced, 0644)
			} else {
				fmt.Fprintf(t, "%s != %s", found, m.Word)
			}
			termbox.Flush()
		}
	}
	fmt.Fprint(t, "done")
	termbox.Flush()
}

func replaceSlice(slice []byte, begin, end int, repl ...byte) []byte {
	total := len(slice) - (end - begin) + len(repl)
	if total > cap(slice) {
		newSlice := make([]byte, total)
		copy(newSlice, slice)
		slice = newSlice
	}
	originalLen := len(slice)
	slice = slice[:total]
	copy(slice[begin+len(repl):originalLen], slice[end:originalLen])
	copy(slice[begin:begin+len(repl)], repl)
	return slice
}

// An Item is something we manage in a priority queue.
type Item struct {
	value    *Misspelling // The value of the item; arbitrary.
	priority int          // The priority of the item in the queue.
	index    int          // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (t *TermboxUI) ReadIntegerInRange(a, b int) int {
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

func (t *TermboxUI) ReadString() string {
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

func (t *TermboxUI) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	w, h := termbox.Size()

	drawRect(2, 2, w-4, h-4, 0xf7)

	tp := t.Printer
	tp.Reset()

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

func drawRect(x, y, w, h int, bg termbox.Attribute) {
	c := ' '
	fg := termbox.ColorDefault
	for i := x; i < x+w; i++ {
		termbox.SetCell(i, y, c, fg, bg)
		termbox.SetCell(i, y+h-1, c, fg, bg)
	}
	for j := y; j < y+h; j++ {
		termbox.SetCell(x, j, c, fg, bg)
		termbox.SetCell(x+1, j, c, fg, bg)
		termbox.SetCell(x+w-2, j, c, fg, bg)
		termbox.SetCell(x+w-1, j, c, fg, bg)
	}
}

func IFix(misspellings []*Misspelling) {
	if len(misspellings) == 0 {
		return
	}
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.HideCursor()
	termbox.SetOutputMode(termbox.Output256)

	ui := NewTermboxUI(misspellings)
	ui.Draw()

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
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
					break mainloop
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		ui.Draw()
	}
}
