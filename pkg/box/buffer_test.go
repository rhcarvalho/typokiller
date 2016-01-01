package box

import (
	"image"
	"reflect"
	"testing"
	"testing/quick"

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
		// []Block tests.
		{
			BoundedCellBufferers{},
			image.ZR,
			[]termbox.Cell{},
		},
		{
			BoundedCellBufferers{
				NewBlock(image.Rect(10, 20, 30, 40), nil),
			},
			image.Rect(10, 20, 30, 40),
			make([]termbox.Cell, 400),
		},
		{
			BoundedCellBufferers{
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("MNOP")),
			},
			image.Rect(1, 1, 3, 3),
			[]termbox.Cell{
				{Ch: 'M'}, {Ch: 'N'},
				{Ch: 'O'}, {Ch: 'P'},
			},
		},
		{
			BoundedCellBufferers{
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("MNOP")),
				NewBlock(image.Rect(2, 1, 4, 3), NewBuffer("1234")),
			},
			image.Rect(1, 1, 4, 3),
			[]termbox.Cell{
				{Ch: 'M'}, {Ch: '1'}, {Ch: '2'},
				{Ch: 'O'}, {Ch: '3'}, {Ch: '4'},
			},
		},
		{
			BoundedCellBufferers{
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("MNOP")),
				NewBlock(image.Rect(4, 0, 5, 4), NewBuffer("1234")),
			},
			image.Rect(1, 0, 5, 4),
			[]termbox.Cell{
				{}, {}, {}, {Ch: '1'},
				{Ch: 'M'}, {Ch: 'N'}, {}, {Ch: '2'},
				{Ch: 'O'}, {Ch: 'P'}, {}, {Ch: '3'},
				{}, {}, {}, {Ch: '4'},
			},
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

func TestBoundedCellBufferersQuick(t *testing.T) {
	var b BoundedCellBufferer
	var p image.Point
	f := func(x0, y0, x1, y1, x2, y2, x3, y3 int8) bool {
		r1 := image.Rect(int(x0), int(y0), int(x1), int(y1))
		r2 := image.Rect(int(x2), int(y2), int(x3), int(y3))
		b = BoundedCellBufferers{
			NewBlock(r1, nil),
			NewBlock(r2, nil),
		}
		p = r1.Union(r2).Size()
		return len(b.CellBuffer()) == p.X*p.Y
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("%v: len(b.CellBuffer())=%v, p=%v", err, len(b.CellBuffer()), p)
	}
}
