package main

import (
	"github.com/nsf/termbox-go"
	"github.com/rhcarvalho/typokiller/pkg/termbox/ui"
	"github.com/rhcarvalho/typokiller/pkg/termbox/widgets"
)

func main() {
	tab1 := widgets.NewParagraph("this is tab 1")
	tab2 := widgets.NewParagraph("and this is tab 2")
	mainView := widgets.NewTabGroup().
		AddTab("TAB1", tab1).
		AddTab("TAB2", tab2).
		Bind(func(w widgets.Widget, e termbox.Event) (widgets.Widget, error) {
		switch w := w.(type) {
		case widgets.TabGroup:
			switch e.Type {
			case termbox.EventKey:
				switch e.Ch {
				case '1':
					return w.SelectTab(0), nil
				case '2':
					return tab2, nil
				}
			}
		}
		return w, nil
	})
	ui.NewLoop(mainView).
		// Global event binding.
		Bind(func(loop ui.Loop, e termbox.Event) bool {
		switch e.Type {
		case termbox.EventKey:
			if e.Key == termbox.KeyEsc {
				loop.Stop()
				return false
			}
		}
		return true
	}).
		Start()
}
