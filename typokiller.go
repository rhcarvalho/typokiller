package main

import (
	"bufio"
	"container/heap"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	docopt "github.com/docopt/docopt-go"
	termbox "github.com/nsf/termbox-go"
)

// Package holds the comments of a Go package and a list of identifiers.
// The identifiers are useful to avoid false positives when spellchecking the
// comments.
type Package struct {
	Name        string `json:"PackageName"`
	Identifiers []string
	Comments    []*Comment
}

// Comment holds the text of a comment and its detailed position.
type Comment struct {
	Text           string
	Position       token.Position
	SpellingErrors []*SpellingError
	Package        *Package
}

type SpellingError struct {
	Word        string
	Offset      int
	Suggestions []string
	Action      *Action
	Comment     *Comment
}

type Action struct {
	Type        ActionType
	Replacement string
}

type ActionType int

const (
	Undefined ActionType = iota
	Ignore
	Replace
)

// ReadDir extracts comments of Go files.
func ReadDir(path string) []*Package {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatalln(err)
	}
	var r []*Package
	for name, pkg := range pkgs {
		r = append(r, ReadPackage(name, pkg, fset))
	}
	return r
}

// ReadPackage extracts comments of a Go package.
func ReadPackage(name string, pkg *ast.Package, fset *token.FileSet) *Package {
	p := &Package{Name: name}
	for _, f := range pkg.Files {
		// Collect comments
		for _, c := range f.Comments {
			begin := fset.Position(c.Pos())
			end := fset.Position(c.End())
			b, err := ioutil.ReadFile(begin.Filename)
			if err != nil {
				panic(err)
			}
			text := string(b[begin.Offset:end.Offset])
			p.Comments = append(p.Comments, &Comment{Text: text, Position: begin})
		}

		// Collect identifiers
		ast.Inspect(pkg, func(n ast.Node) bool {
			if ident, isIdent := n.(*ast.Ident); isIdent {
				p.Identifiers = append(p.Identifiers, ident.String())
			}
			return true
		})
	}
	return p
}

type TermboxUI struct {
	SpellingErrors []*SpellingError
	Index          int
	Printer        *TermboxPrinter
}

func NewTermboxUI(spellingErrors []*SpellingError) *TermboxUI {
	ui := &TermboxUI{
		SpellingErrors: spellingErrors,
		Printer:        &TermboxPrinter{left: 5, right: 5, top: 5, bottom: 5},
	}
	return ui
}

func (t *TermboxUI) Next() {
	t.Index++
	if !(t.Index < len(t.SpellingErrors)) {
		t.Index = 0
	}
}

func (t *TermboxUI) NextUndefined() {
	start := t.Index
	for {
		t.Index++
		if !(t.Index < len(t.SpellingErrors)) {
			t.Index = 0
		}
		se := t.SpellingErrors[t.Index]
		if se.Action == nil || se.Action.Type == Undefined {
			break
		}
		if t.Index == start {
			t.Printer.fg = termbox.ColorGreen
			t.Printer.Println("all done")
			break
		}
	}
}

func (t *TermboxUI) Previous() {
	t.Index--
	if t.Index < 0 {
		t.Index = len(t.SpellingErrors) - 1
	}
}

func (t *TermboxUI) Ignore() {
	se := t.SpellingErrors[t.Index]
	se.Action = &Action{Type: Ignore}
	t.NextUndefined()
}

func (t *TermboxUI) Replace() {
	se := t.SpellingErrors[t.Index]
	se.Action = &Action{
		Type:        Replace,
		Replacement: se.Suggestions[t.ReadIntegerInRange(1, len(se.Suggestions))-1]}
	t.NextUndefined()
}

func (t *TermboxUI) Edit() {
	se := t.SpellingErrors[t.Index]
	se.Action = &Action{
		Type:        Replace,
		Replacement: t.ReadString()}
	t.NextUndefined()
}

func (t *TermboxUI) IgnoreAll() {
	se := t.SpellingErrors[t.Index]
	word := se.Word
	for i := t.Index; i < len(t.SpellingErrors); i++ {
		se = t.SpellingErrors[i]
		if se.Word == word && (se.Action == nil || se.Action.Type == Undefined) {
			se.Action = &Action{Type: Ignore}
		}
	}
	t.NextUndefined()
}

func (t *TermboxUI) ReplaceAll() {
	se := t.SpellingErrors[t.Index]
	word := se.Word
	replacement := se.Suggestions[t.ReadIntegerInRange(1, len(se.Suggestions))-1]
	for i := t.Index; i < len(t.SpellingErrors); i++ {
		se = t.SpellingErrors[i]
		if se.Word == word && (se.Action == nil || se.Action.Type == Undefined) {
			se.Action = &Action{Type: Replace, Replacement: replacement}
		}
	}
	t.NextUndefined()
}

func (t *TermboxUI) EditAll() {
	se := t.SpellingErrors[t.Index]
	word := se.Word
	replacement := t.ReadString()
	for i := t.Index; i < len(t.SpellingErrors); i++ {
		se = t.SpellingErrors[i]
		if se.Word == word && (se.Action == nil || se.Action.Type == Undefined) {
			se.Action = &Action{Type: Replace, Replacement: replacement}
		}
	}
	t.NextUndefined()
}

func (t *TermboxUI) Apply() {
	defer termbox.PollEvent()
	t.Printer.fg = termbox.ColorGreen
	t.Printer.Println("applying changes...")
	termbox.Flush()
	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(t.SpellingErrors))
	for i, se := range t.SpellingErrors {
		pq[i] = &Item{
			value:    se,
			priority: se.Comment.Position.Offset + se.Offset,
			index:    i,
		}
	}
	heap.Init(&pq)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		se := item.value
		if se.Action != nil && se.Action.Type == Replace {
			pos := se.Comment.Position
			t.Printer.Printf("%s:%d:%d '%s' -> '%s'\n",
				pos.Filename, pos.Line, pos.Column,
				se.Word, se.Action.Replacement)
			termbox.Flush()
		}
	}
}

// An Item is something we manage in a priority queue.
type Item struct {
	value    *SpellingError // The value of the item; arbitrary.
	priority int            // The priority of the item in the queue.
	index    int            // The index of the item in the heap.
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
	t.Printer.Printf("\nenter number in range [%d, %d]: ", a, b)
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
				t.Printer.Print(string(ev.Ch))
				termbox.Flush()
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
	t.Printer.ResetColors()
	i, err := strconv.Atoi(string(v))
	if err != nil || !(a <= i && i <= b) {
		t.Printer.Println(" → invalid number, try again")
		termbox.Flush()
		v = nil
		goto start
	}
	return i
}

func (t *TermboxUI) ReadString() string {
	t.Printer.fg |= termbox.AttrBold
	t.Printer.Print("\nreplace with: ")
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
					t.Printer.Print(" ")
					t.Printer.x--
					termbox.Flush()
				}
			default:
				if unicode.IsPrint(ev.Ch) {
					v = append(v, ev.Ch)
					t.Printer.Print(string(ev.Ch))
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

	tp.Print("Spelling error ")
	tp.Bold()
	tp.Printf("%d", t.Index+1)
	tp.ResetColors()
	tp.Print(" of ")
	tp.Bold()
	tp.Printf("%d", len(t.SpellingErrors))
	tp.ResetColors()

	se := t.SpellingErrors[t.Index]
	comment := se.Comment

	tp.SkipLines(2)
	tp.Underline()
	tp.Printf("%s:%d:%d\n", comment.Position.Filename, comment.Position.Line, comment.Position.Column)

	tp.SkipLines(1)
	tmp := tp.y
	tp.fg = 0xf0
	tp.Println(comment.Text)
	tmp, tp.y = tp.y, tmp

	tp.x += se.Offset - strings.LastIndex(comment.Text[:se.Offset], "\n") - 1
	tp.y += strings.Count(comment.Text[:se.Offset], "\n")
	tp.fg = termbox.ColorRed | termbox.AttrBold
	tp.Print(se.Word)
	tp.ResetColors()
	tp.y = tmp

	tp.SkipLines(1)
	tp.Print("Suggestions: ")
	for i, suggestion := range se.Suggestions {
		if i > 0 {
			tp.Print(", ")
		}
		tp.Printf("[%d] %s", i+1, suggestion)
	}
	tp.SkipLines(2)

	tp.Print("Actions: ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("r")
	tp.ResetColors()
	tp.Print("eplace, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("R")
	tp.ResetColors()
	tp.Print("eplace all, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("i")
	tp.ResetColors()
	tp.Print("gnore, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("I")
	tp.ResetColors()
	tp.Print("gnore all, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("e")
	tp.ResetColors()
	tp.Print("dit, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("E")
	tp.ResetColors()
	tp.Print("dit all, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("n")
	tp.ResetColors()
	tp.Print("ext undefined, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("a")
	tp.ResetColors()
	tp.Print("pply, ")
	tp.fg |= termbox.AttrUnderline
	tp.Print("q")
	tp.ResetColors()
	tp.Println("uit")

	if se.Action != nil {
		tp.SkipLines(1)
		tp.fg = termbox.ColorBlue
		switch se.Action.Type {
		case Ignore:
			tp.Println("ignored")
		case Replace:
			tp.Printf("replace with '%s'\n", se.Action.Replacement)
		}
		tp.ResetColors()
	}
}

func IFix(spellingErrors []*SpellingError) {
	if len(spellingErrors) == 0 {
		return
	}
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.HideCursor()
	termbox.SetOutputMode(termbox.Output256)

	ui := NewTermboxUI(spellingErrors)
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

type TermboxPrinter struct {
	x, y        int               // current cursor position (column, line)
	left, right int               // left and right margins
	top, bottom int               // top and bottom margins
	fg, bg      termbox.Attribute // foreground and background colors
}

func (tbp *TermboxPrinter) Reset() {
	tbp.x = 0
	tbp.y = 0
	tbp.ResetColors()
}

func (tbp *TermboxPrinter) Print(text string) {
	for _, c := range text {
		if c == '\n' {
			tbp.x = 0
			tbp.y++
			continue
		}
		termbox.SetCell(tbp.left+tbp.x, tbp.top+tbp.y, c, tbp.fg, tbp.bg)
		tbp.x++
		w, _ := termbox.Size()
		if tbp.x >= w-tbp.right-tbp.left-2 {
			termbox.SetCell(tbp.left+tbp.x+1, tbp.top+tbp.y, '⏎', termbox.ColorWhite, termbox.ColorRed)
			tbp.SkipLines(1)
		}
	}
}

func (tbp *TermboxPrinter) Println(text string) {
	tbp.Print(text)
	tbp.Print("\n")
}

func (tbp *TermboxPrinter) Printf(format string, a ...interface{}) {
	tbp.Print(fmt.Sprintf(format, a...))
}

func (tbp *TermboxPrinter) Bold() {
	tbp.fg |= termbox.AttrBold
}

func (tbp *TermboxPrinter) Underline() {
	tbp.fg |= termbox.AttrUnderline
}

func (tbp *TermboxPrinter) ResetColors() {
	tbp.fg = termbox.ColorDefault
	tbp.bg = termbox.ColorDefault
}

func (tbp *TermboxPrinter) SkipLines(n int) {
	tbp.y += n
	tbp.x = 0
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

func main() {
	usage := `Usage:
  typokiller scan PATH ...
  typokiller fix

Find comments in Go source files and interactively fix typos.

Options:
  -h --help  Show this usage help
  --version  Show version

Commands:
  scan       Outputs comments for the packages found
  fix        Reads spelling error information from STDIN and allows for interative patching`
	arguments, _ := docopt.Parse(usage, nil, true, "typokiller 0.1", false)

	// fix typos mode
	if arguments["fix"].(bool) {
		reader := bufio.NewReaderSize(os.Stdin, 64*1024*1024) // 64 MB
		var spellingErrors []*SpellingError
		var err error
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			var pkg *Package
			if err = json.Unmarshal(line, &pkg); err != nil {
				log.Fatalf("error: %v\nline: %s\n", err, line)
			}

			for _, c := range pkg.Comments {
				c.Package = pkg
				for _, s := range c.SpellingErrors {
					s.Comment = c
					spellingErrors = append(spellingErrors, s)
				}
			}
		}
		if err != nil && err != io.EOF {
			log.Fatalln(err)
		}
		IFix(spellingErrors)
	}

	// scan comments mode
	enc := json.NewEncoder(os.Stdout)
	for _, path := range arguments["PATH"].([]string) {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Fatalln(err)
		}
		for _, pkg := range ReadDir(path) {
			enc.Encode(pkg)
		}
	}
}
