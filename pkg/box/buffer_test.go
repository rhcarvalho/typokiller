package box

import (
	"image"
	"reflect"
	"testing"

	"github.com/nsf/termbox-go"
)

func TestBuffer(t *testing.T) {
	tests := []struct {
		in   CellBufferer
		want []termbox.Cell
	}{
		// Buffer tests.
		{
			Buffer{},
			[]termbox.Cell{},
		},
		{
			NewBuffer("123"),
			[]termbox.Cell{{Ch: '1'}, {Ch: '2'}, {Ch: '3'}},
		},
		// []Buffer tests.
		{
			CellBufferers{},
			[]termbox.Cell{},
		},
		{
			CellBufferers{
				Buffer{},
			},
			[]termbox.Cell{},
		},
		{
			CellBufferers{
				Buffer{},
				NewBuffer("ABC"),
			},
			[]termbox.Cell{{Ch: 'A'}, {Ch: 'B'}, {Ch: 'C'}},
		},
		{
			CellBufferers{
				NewBuffer("XYZ"),
				Buffer{},
			},
			[]termbox.Cell{{Ch: 'X'}, {Ch: 'Y'}, {Ch: 'Z'}},
		},
		{
			CellBufferers{
				NewBuffer("ABC"),
				NewBuffer("XYZ"),
			},
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'}, {Ch: 'C'},
				{Ch: 'X'}, {Ch: 'Y'}, {Ch: 'Z'}},
		},
	}
	for _, test := range tests {
		got := test.in.CellBuffer()
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("%v.CellBuffer() = %v, want %v", test.in, got, test.want)
		}
	}
}

func TestBlock(t *testing.T) {
	tests := []struct {
		in         BoundedCellBufferer
		wantBounds image.Rectangle
		wantCells  []termbox.Cell
	}{
		// Empty Block tests.
		{
			Block{},
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Block{
				Rectangle: image.Rectangle{
					image.Point{4, 8},
					image.Point{0, 0}},
			},
			image.Rect(0, 0, 4, 8),
			make([]termbox.Cell, 32),
		},
		// Buffer-in-a-Block tests.
		{
			// Buffer will be truncated to fit in the rectangle.
			NewBlock(image.Rect(1, 2, 2, 3), NewBuffer("MNO")),
			image.Rect(1, 2, 2, 3),
			[]termbox.Cell{{Ch: 77}},
		},
		{
			NewBlock(image.Rect(0, 0, 3, 2), NewBuffer("MNO")),
			image.Rect(0, 0, 3, 2),
			[]termbox.Cell{{Ch: 77}, {Ch: 78}, {Ch: 79}, {}, {}, {}},
		},
		{
			// Bounds can be anywhere in the 2D space.
			NewBlock(image.Rect(-42, -43, -41, -39), NewBuffer("MNO")),
			image.Rect(-42, -43, -41, -39),
			[]termbox.Cell{{Ch: 77}, {Ch: 78}, {Ch: 79}, {}},
		},
		// Block-in-a-Block tests.
		{
			NewBlock(image.Rect(0, 0, 4, 4),
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("MNO"))),
			image.Rect(0, 0, 4, 4),
			[]termbox.Cell{
				{}, {}, {}, {},
				{}, {Ch: 'M'}, {Ch: 'N'}, {},
				{}, {Ch: 'O'}, {}, {},
				{}, {}, {}, {},
			},
		},
		{
			// Inner blocks bounds are considered to be relative. In
			// this example, the final rectangle is 3x3 and the
			// first non-empty cell is at (1, 1).
			NewBlock(image.Rect(1, 1, 4, 4),
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("MNO"))),
			image.Rect(1, 1, 4, 4),
			[]termbox.Cell{
				{}, {}, {},
				{}, {Ch: 'M'}, {Ch: 'N'},
				{}, {Ch: 'O'}, {},
			},
		},
		{
			// Inner block is truncated from bottom-right.
			NewBlock(image.Rect(2, 2, 4, 4),
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("MNO"))),
			image.Rect(2, 2, 4, 4),
			[]termbox.Cell{
				{}, {},
				{}, {Ch: 'M'},
			},
		},
		{
			// Inner block is truncated from top-left.
			NewBlock(image.Rect(0, 0, 2, 2),
				NewBlock(image.Rect(-1, -1, 1, 1), NewBuffer("MNOP"))),
			image.Rect(0, 0, 2, 2),
			[]termbox.Cell{
				{Ch: 'P'}, {},
				{}, {},
			},
		},
		// Block-in-a-Block-in-a-Block tests.
		{
			NewBlock(image.Rect(0, 0, 1, 1),
				NewBlock(image.Rect(0, 0, 2, 2),
					NewBlock(image.Rect(-1, -1, 1, 1), NewBuffer("MNOP")))),
			image.Rect(0, 0, 1, 1),
			[]termbox.Cell{{Ch: 'P'}},
		},
	}
	for _, test := range tests {
		gotBounds := test.in.Bounds()
		if !reflect.DeepEqual(gotBounds, test.wantBounds) {
			t.Fatalf("%v.Bounds() = %v, want %v", test.in, gotBounds, test.wantBounds)
		}
		gotCells := test.in.CellBuffer()
		if !reflect.DeepEqual(gotCells, test.wantCells) {
			t.Fatalf("%v.CellBuffer() = %v, want %v", test.in, gotCells, test.wantCells)
		}
	}
}
