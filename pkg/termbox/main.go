package main

import (
	"crypto/sha1"
	"fmt"

	"github.com/nsf/termbox-go"
	"github.com/rhcarvalho/typokiller/pkg/termbox/ui"
	"github.com/rhcarvalho/typokiller/pkg/termbox/widgets"
)

func main() {
	tab1 := widgets.NewParagraph("this is tab 1")
	tab2 := widgets.NewParagraph("and this is tab 2").
		Bind(func(w widgets.Widget, e termbox.Event) (widgets.Widget, error) {
		switch w := w.(type) {
		case widgets.Paragraph:
			switch e.Type {
			case termbox.EventKey:
				switch e.Ch {
				case 'e':
					w.Text = fmt.Sprintf("%s\t%x", w.Text, sha1.Sum([]byte(w.Text)))
					return w, nil
				}
			}
		}
		return w, nil
	})
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
					return w.SelectTab(1), nil
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
