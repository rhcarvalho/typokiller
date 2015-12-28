package main

import (
	"crypto/sha1"
	"fmt"

	"github.com/nsf/termbox-go"
	"github.com/rhcarvalho/typokiller/pkg/termbox/ui"
	"github.com/rhcarvalho/typokiller/pkg/termbox/widgets"
)

func main() {
	tab1 := widgets.NewParagraph("this is tab 1").
		Bind(func(w widgets.Widget, e termbox.Event) (widgets.Widget, bool) {
		switch tab1 := w.(type) {
		case widgets.Paragraph:
			if e.Type == termbox.EventKey && e.Ch == 'e' {
				return widgets.NewInput("", tab1.Text).
					Bind(func(w widgets.Widget, e termbox.Event) (widgets.Widget, bool) {
					if e.Type == termbox.EventKey {
						switch e.Key {
						case termbox.KeyEnter:
							termbox.HideCursor()
							tab1.Text = w.(widgets.Input).Value()
							return tab1, true
						case termbox.KeyEsc:
							termbox.HideCursor()
							return tab1, true
						}

					}
					return w, false
				}), true
			}
		}
		return w, false
	})
	tab2 := widgets.NewParagraph("and this is tab 2").
		Bind(func(w widgets.Widget, e termbox.Event) (widgets.Widget, bool) {
		switch w := w.(type) {
		case widgets.Paragraph:
			switch e.Type {
			case termbox.EventKey:
				switch e.Ch {
				case 'e':
					w.Text = fmt.Sprintf("%s\t%x", w.Text, sha1.Sum([]byte(w.Text)))
					return w, true
				}
			}
		}
		return w, false
	})
	mainView := widgets.NewTabGroup().
		AddTab("TAB1", tab1).
		AddTab("TAB2", tab2).
		Bind(func(w widgets.Widget, e termbox.Event) (widgets.Widget, bool) {
		switch w := w.(type) {
		case widgets.TabGroup:
			switch e.Type {
			case termbox.EventKey:
				switch e.Ch {
				case '1':
					return w.SelectTab(0), false
				case '2':
					return w.SelectTab(1), false
				}
			}
		}
		return w, false
	})
	ui.NewLoop(mainView).
		// Global event binding.
		Bind(func(loop ui.Loop, e termbox.Event) ui.Loop {
		switch e.Type {
		case termbox.EventKey:
			if e.Key == termbox.KeyEsc {
				loop.Stop()
			}
		}
		return loop
	}).
		Start()
}
