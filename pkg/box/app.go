package box

import "image"

// App is a termbox-based app.
type App struct {
	termboxer
}

// NewApp creates a new App. Call its Init method before using it and Close when
// done.
func NewApp() App {
	return App{termboxer: tb{}}
}

// Render fits f to the screen dimensions, and updates the visible buffer.
func (app App) Render(f Fitter) {
	w, h := app.Size()
	screen := image.Rect(0, 0, w, h)
	copy(app.CellBuffer(), f.Fit(screen).CellBuffer())
	app.Flush()
}
