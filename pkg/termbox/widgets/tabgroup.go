package widgets

import "github.com/nsf/termbox-go"

type TabGroup struct {
	titles        []string
	tabs          []Widget
	selectedIndex int
	handler       func(Widget, termbox.Event) (Widget, error)
}

func NewTabGroup() TabGroup {
	return TabGroup{}
}

func (tg TabGroup) Render(x, y, w, h int) {
	tg.render(x, y, w, h)
	if tg.selectedIndex >= 0 && tg.selectedIndex < len(tg.tabs) {
		tg.tabs[tg.selectedIndex].Render(x+1, y+1, w-2, h-2)
	}
}

func (tg TabGroup) Bind(f func(Widget, termbox.Event) (Widget, error)) Widget {
	tg.handler = f
	return tg
}

func (tb TabGroup) Handle(e termbox.Event) (Widget, error) {
	return tb.handler(tb, e)
}

func (tg TabGroup) AddTab(title string, view Widget) TabGroup {
	tg.titles = append(tg.titles, title)
	tg.tabs = append(tg.tabs, view)
	return tg
}

func (tg TabGroup) SelectTab(i int) TabGroup {
	tg.selectedIndex = i
	return tg
}

func (tg TabGroup) render(x, y, w, h int) {
	// c := ' '
	// const HORIZONTAL_LINE = '─'
	fg := termbox.ColorDefault
	bg := termbox.Attribute(0xf7)

	// Draw horizontal borders.
	for i := x; i < w-x; i++ {
		termbox.SetCell(i, y, '─', fg, bg)
		termbox.SetCell(i, h-y-1, '─', fg, bg)
	}
	// Draw vertical borders.
	for j := y; j < h-y; j++ {
		termbox.SetCell(x, j, '│', fg, bg)
		termbox.SetCell(w-x-1, j, '│', fg, bg)
	}
	// Draw corners.
	termbox.SetCell(x, y, '┌', fg, bg)
	termbox.SetCell(w-x-1, y, '┐', fg, bg)
	termbox.SetCell(x, h-y-1, '└', fg, bg)
	termbox.SetCell(w-x-1, h-y-1, '┘', fg, bg)

	if len(tg.titles) == 0 {
		return
	}

	// Draw tab titles.
	p := x + 1
	termbox.SetCell(p, y, '┤', fg, bg)

	for i, title := range tg.titles {
		p++
		termbox.SetCell(p, y, ' ', fg, fg)
		for _, c := range title {
			p++
			if i == tg.selectedIndex {
				termbox.SetCell(p, y, c, fg, termbox.ColorGreen)
			} else {
				termbox.SetCell(p, y, c, fg, fg)
			}
		}
		p++
		termbox.SetCell(p, y, ' ', fg, fg)
		p++
		termbox.SetCell(p, y, '├', fg, bg)
		p++
		termbox.SetCell(p, y, '─', fg, bg)
		p++
		termbox.SetCell(p, y, '┤', fg, bg)
	}
	termbox.SetCell(p, y, '─', fg, bg)
}
