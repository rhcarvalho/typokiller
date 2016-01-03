package box

import (
	"reflect"
	"testing"

	"github.com/nsf/termbox-go"
)

type TestApp struct {
	App
}

func NewTestApp() TestApp {
	app := TestApp{App: App{termboxer: &testBox{w: 10, h: 3}}}
	app.Init()
	return app
}

func (app TestApp) FrontBuffer() []termbox.Cell {
	return app.termboxer.(*testBox).frontBuf
}

func (app TestApp) Log() []string {
	return app.termboxer.(*testBox).log
}

func TestRender(t *testing.T) {
	app := NewTestApp()
	buf := NewBuffer("This Is A Test!")
	app.Render(buf)
	got := app.FrontBuffer()
	want := []termbox.Cell{
		// row #1
		{Ch: 'T'}, {Ch: 'h'}, {Ch: 'i'}, {Ch: 's'}, {Ch: ' '},
		{Ch: 'I'}, {Ch: 's'}, {Ch: ' '}, {Ch: 'A'}, {Ch: ' '},
		// row #2
		{Ch: 'T'}, {Ch: 'e'}, {Ch: 's'}, {Ch: 't'}, {Ch: '!'},
		{}, {}, {}, {}, {},
		// row #3
		{}, {}, {}, {}, {},
		{}, {}, {}, {}, {},
	}
	if !reflect.DeepEqual(got, want) {
		t.Log(app.Log())
		t.Fatalf("app.Render(%v): got %v, want %v", buf, got, want)
	}
}
