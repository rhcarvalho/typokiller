package box

import (
	"bytes"
	"fmt"
	"image"

	"github.com/nsf/termbox-go"
)

// A CellBufferer is an interface implemented by types that can be used to fill
// in termbox's cell buffer.
type CellBufferer interface {
	CellBuffer() []termbox.Cell
}

// Buffer is a basic implementation of a CellBufferer.
type Buffer struct {
	bytes.Buffer
	fg, bg termbox.Attribute
}

// NewBuffer creates a new Buffer with the string s and default foreground and
// background colors.
func NewBuffer(s string) Buffer {
	return Buffer{
		Buffer: *bytes.NewBufferString(s),
	}
}

func (b Buffer) String() string {
	return fmt.Sprintf("Buffer{bytes:%v, fg:%v, bg:%v}", b.Bytes(), b.fg, b.bg)
}

// CellBuffer implements CellBufferer.
func (b Buffer) CellBuffer() []termbox.Cell {
	runes := bytes.Runes(b.Bytes())
	cellBuf := make([]termbox.Cell, len(runes))
	for i, r := range runes {
		cellBuf[i].Ch = r
		cellBuf[i].Fg = b.fg
		cellBuf[i].Bg = b.bg
	}
	return cellBuf
}

// Fit implements Fitter.
func (b Buffer) Fit(r image.Rectangle) BoundedCellBufferer {
	return NewBlock(r, b)
}

// CellBufferers groups together cell bufferers and implements the CellBufferer
// interface itself.
type CellBufferers []CellBufferer

// CellBuffer implements CellBufferer.
func (bs CellBufferers) CellBuffer() []termbox.Cell {
	cellBuf := []termbox.Cell{}
	for _, b := range bs {
		cellBuf = append(cellBuf, b.CellBuffer()...)
	}
	return cellBuf
}

// A BoundedCellBufferer is a CellBufferer with rectangular bounds. Bounds make
// the cell buffer fit a certain number of lines and columns. The number of
// cells returned by CellBuffer() is the area of the rectangle.
type BoundedCellBufferer interface {
	Bounds() image.Rectangle
	CellBufferer
}

// Block is a bounded cell bufferer.
type Block struct {
	image.Rectangle
	CellBufferer
}

// NewBlock creates a new block bounded by r and with the cell buffer given by
// cb.
func NewBlock(r image.Rectangle, cb CellBufferer) Block {
	return Block{
		Rectangle:    r,
		CellBufferer: cb,
	}
}

func (b Block) String() string {
	return fmt.Sprintf("Block{bounds:%v, buffer:%v}", b.Rectangle, b.CellBufferer)
}

// Bounds implements BoundedCellBufferer.
func (b Block) Bounds() image.Rectangle {
	return b.Canon()
}

// CellBuffer implements BoundedCellBufferer.
func (b Block) CellBuffer() []termbox.Cell {
	cellBuf := make([]termbox.Cell, b.Dx()*b.Dy())
	if bcb, ok := b.CellBufferer.(BoundedCellBufferer); ok {
		cb := bcb.CellBuffer()
		// An inner block is considered relative to the outer block, so
		// we need to translate the inner block.
		r := bcb.Bounds().Add(b.Min)
		// Only the intersection between b and r needs to be copied to
		// cellBuf.
		ri := b.Intersect(r)
		for j := 0; j < ri.Dy(); j++ {
			k1 := ri.Min.X - b.Min.X + (ri.Min.Y-b.Min.Y+j)*b.Dx()
			k2 := ri.Min.X - r.Min.X + (ri.Min.Y-r.Min.Y+j)*r.Dx()
			copy(cellBuf[k1:k1+ri.Dx()], cb[k2:k2+ri.Dx()])
		}
	} else if b.CellBufferer != nil {
		copy(cellBuf, b.CellBufferer.CellBuffer())
	}
	return cellBuf
}

// Fit implements Fitter.
func (b Block) Fit(r image.Rectangle) BoundedCellBufferer {
	return NewBlock(r, b)
}

// BoundedCellBufferers is a group of bounded cell bufferers. It implements
// BoundedCellBufferer itself.
type BoundedCellBufferers []BoundedCellBufferer

// Bounds implements BoundedCellBufferer.
func (bs BoundedCellBufferers) Bounds() image.Rectangle {
	if len(bs) == 0 {
		return image.ZR
	}
	r := bs[0].Bounds()
	for _, b := range bs[1:] {
		r = r.Union(b.Bounds())
	}
	return r
}

// CellBuffer implements BoundedCellBufferer.
func (bs BoundedCellBufferers) CellBuffer() []termbox.Cell {
	p := bs.Bounds().Size()
	cellBuf := make([]termbox.Cell, p.X*p.Y)
	m := bs.Bounds().Min
	for _, b := range bs {
		r := b.Bounds()
		cb := b.CellBuffer()
		for j := 0; j < r.Dy(); j++ {
			k1 := r.Min.X - m.X + (r.Min.Y-m.Y+j)*p.X
			k2 := j * r.Dx()
			copy(cellBuf[k1:k1+r.Dx()], cb[k2:k2+r.Dx()])
		}
	}
	return cellBuf
}

// A Grid groups CellBufferers horizontally and/or vertically, stacking them
// side by side.
type Grid struct {
	rows []row
}

// Col adds a new column with a certain weight and returns a new Grid. If called
// before a call to Row, the new column is inserted in a new row with weight 1.
func (g Grid) Col(weight uint8, cb CellBufferer) Grid {
	if len(g.rows) == 0 {
		g.rows = append(g.rows, row{1, nil})
	}
	row := &g.rows[len(g.rows)-1]
	row.cols = append(row.cols, col{weight, cb})
	return g
}

// Row adds a new row with a certain weight and returns a new Grid.
func (g Grid) Row(weight uint8) Grid {
	g.rows = append(g.rows, row{weight: weight})
	return g
}

// Fit returns a BoundedCellBufferer in which column widths and row heights are
// proportional to their weights.
func (g Grid) Fit(r image.Rectangle) BoundedCellBufferer {
	return fitRows(r.Canon(), g.rows)
}

func fitRows(r image.Rectangle, rows []row) BoundedCellBufferer {
	b := append(BoundedCellBufferers(nil), NewBlock(r, nil))
	var sum int
	for _, row := range rows {
		sum += int(row.weight)
	}
	if sum == 0 {
		// No row with weight > 0, return early.
		return b
	}
	y0 := r.Min.Y
	for _, row := range rows {
		y1 := y0 + int(row.weight)*r.Dy()/sum
		b = append(b, NewBlock(
			image.Rect(r.Min.X, y0, r.Max.X, y1),
			fitCols(r.Sub(r.Min), row.cols)))
		y0 = y1
	}
	return b
}

func fitCols(r image.Rectangle, cols []col) BoundedCellBufferer {
	b := append(BoundedCellBufferers(nil), NewBlock(r, nil))
	var sum int
	for _, c := range cols {
		sum += int(c.weight)
	}
	if sum == 0 {
		// No column with weight > 0, return early.
		return b
	}
	x0 := r.Min.X
	for _, c := range cols {
		x1 := x0 + int(c.weight)*r.Dx()/sum
		b = append(b, NewBlock(
			image.Rect(x0, r.Min.Y, x1, r.Max.Y),
			c.cb))
		x0 = x1
	}
	return b
}

type row struct {
	weight uint8
	cols   []col
}

type col struct {
	weight uint8
	cb     CellBufferer
}

// Fitter is an interface implemented by types that can produce a
// BoundedCellBufferer fitting a given rectangle.
type Fitter interface {
	Fit(r image.Rectangle) BoundedCellBufferer
}
