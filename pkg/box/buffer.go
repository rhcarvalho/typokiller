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
