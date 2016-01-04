package box

import (
	"image"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/nsf/termbox-go"
)

func TestCellBufferer(t *testing.T) {
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
		{
			CellBufferers{
				NewBuffer("ABC"),
				NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("XYZ")),
			},
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'}, {Ch: 'C'},
				{Ch: 'X'}, {Ch: 'Y'}, {Ch: 'Z'}, {}},
		},
		// Colored buffer tests.
		{
			NewBuffer("Blue").Fg(termbox.ColorRed),
			[]termbox.Cell{
				{Ch: 'B', Fg: termbox.ColorRed},
				{Ch: 'l', Fg: termbox.ColorRed},
				{Ch: 'u', Fg: termbox.ColorRed},
				{Ch: 'e', Fg: termbox.ColorRed},
			},
		},
		{
			NewBuffer("Blue").Bg(termbox.ColorRed),
			[]termbox.Cell{
				{Ch: 'B', Bg: termbox.ColorRed},
				{Ch: 'l', Bg: termbox.ColorRed},
				{Ch: 'u', Bg: termbox.ColorRed},
				{Ch: 'e', Bg: termbox.ColorRed},
			},
		},
		{
			NewBuffer("B").
				Fg(termbox.ColorRed).
				Bg(termbox.ColorGreen),
			[]termbox.Cell{{Ch: 'B',
				Fg: termbox.ColorRed,
				Bg: termbox.ColorGreen}},
		},
	}
	for _, test := range tests {
		got := test.in.CellBuffer()
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("%v.CellBuffer() = %v, want %v", test.in, got, test.want)
		}
	}
}

func TestBoundedCellBufferer(t *testing.T) {
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

func TestFitter(t *testing.T) {
	tests := []struct {
		fitter Fitter
		rect   image.Rectangle
		want   []termbox.Cell
	}{
		// Buffer, Block and a mix of those should implement Fitter.
		{
			NewBuffer("OK"),
			image.Rect(0, 0, 1, 2),
			[]termbox.Cell{{Ch: 'O'}, {Ch: 'K'}},
		},
		{
			NewBlock(image.Rect(0, 0, 1, 2), NewBuffer("OK")),
			image.Rect(0, 0, 2, 2),
			[]termbox.Cell{{Ch: 'O'}, {}, {Ch: 'K'}, {}},
		},
		{
			CellBufferers{NewBuffer("X"), NewBuffer("Y")},
			image.Rect(0, 0, 1, 2),
			[]termbox.Cell{{Ch: 'X'}, {Ch: 'Y'}},
		},
		{
			BoundedCellBufferers{
				NewBlock(image.Rect(4, 0, 5, 2), NewBuffer("OK")),
				NewBlock(image.Rect(0, 0, 4, 1), NewBuffer("Very")),
			},
			image.Rect(0, 0, 5, 2),
			[]termbox.Cell{
				{Ch: 'V'}, {Ch: 'e'}, {Ch: 'r'}, {Ch: 'y'}, {Ch: 'O'},
				{}, {}, {}, {}, {Ch: 'K'},
			},
		},
		// Empty Grid tests.
		{
			Grid{},
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{},
			image.Rect(42, 42, 42, 42),
			[]termbox.Cell{},
		},
		{
			Grid{},
			image.Rect(-1, -1, 3, 3),
			make([]termbox.Cell, 16),
		},
		// Single Col tests.
		{
			Grid{}.Col(0, nil),
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{}.Col(1, nil),
			image.Rect(0, 0, 3, 3),
			make([]termbox.Cell, 9),
		},
		{
			Grid{}.Col(1, NewBuffer("ABCD")),
			image.Rect(0, 0, 3, 3),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'}, {Ch: 'C'},
				{Ch: 'D'}, {}, {},
				{}, {}, {},
			},
		},
		{
			Grid{}.Col(100, NewBuffer("ABCDE")),
			image.Rect(1, 1, 3, 3),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'},
				{Ch: 'C'}, {Ch: 'D'},
			},
		},
		{
			// The Block should be positioned relative to the bounds
			// we're fitting it in.
			Grid{}.Col(1, NewBlock(image.Rect(1, 1, 3, 3), NewBuffer("ABCDE"))),
			image.Rect(2, 2, 6, 6),
			[]termbox.Cell{
				{}, {}, {}, {},
				{}, {Ch: 'A'}, {Ch: 'B'}, {},
				{}, {Ch: 'C'}, {Ch: 'D'}, {},
				{}, {}, {}, {},
			},
		},
		{
			// The Block should be positioned relative to the bounds
			// we're fitting it in.
			Grid{}.Col(1, NewBlock(image.Rect(-1, -1, 1, 1), NewBuffer("ABCDE"))),
			image.Rect(2, 2, 6, 4),
			[]termbox.Cell{
				{Ch: 'D'}, {}, {}, {},
				{}, {}, {}, {},
			},
		},
		{
			// The Block is wider than the column width, it will be
			// truncated.
			Grid{}.Col(1, NewBlock(image.Rect(0, 0, 5, 1), NewBuffer("ABCDE"))),
			image.Rect(2, 2, 6, 4),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'}, {Ch: 'C'}, {Ch: 'D'},
				{}, {}, {}, {},
			},
		},
		{
			// The Block is taller than the row height, it will be
			// truncated.
			Grid{}.Col(1, NewBlock(image.Rect(0, 0, 1, 5), NewBuffer("ABCDE"))),
			image.Rect(2, 2, 4, 5),
			[]termbox.Cell{
				{Ch: 'A'}, {},
				{Ch: 'B'}, {},
				{Ch: 'C'}, {},
			},
		},
		// Multiple Col tests.
		{
			Grid{}.Col(0, nil).Col(0, nil).Col(0, nil),
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{}.Col(1, nil).Col(1, nil).Col(1, nil),
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{}.Col(1, nil).Col(1, nil).Col(1, nil),
			image.Rect(2, 2, 3, 3),
			make([]termbox.Cell, 1),
		},
		{
			Grid{}.Col(1, NewBuffer("XYZ")).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 10, 1),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'}, {Ch: 'Z'}, {}, {}, // col #1
				{Ch: 'M'}, {Ch: 'N'}, {Ch: 'O'}, {}, {}, // col #2
			},
		},
		{
			Grid{}.Col(1, NewBuffer("XYZ")).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 4, 1),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'}, // col #1
				{Ch: 'M'}, {Ch: 'N'}, // col #2
			},
		},
		{
			Grid{}.Col(1, NewBuffer("XYZ")).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 4, 2),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'}, {Ch: 'M'}, {Ch: 'N'}, // row #1
				{Ch: 'Z'}, {}, {Ch: 'O'}, {}, // row #2
			},
		},
		{
			// The columns do not fit perfectly, there's an extra
			// empty column in the far left that could not be split
			// between the two columns.
			Grid{}.Col(1, NewBuffer("XYZ")).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 5, 2),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'}, {Ch: 'M'}, {Ch: 'N'}, {}, // row #1
				{Ch: 'Z'}, {}, {Ch: 'O'}, {}, {}, // row #2
			},
		},
		{
			Grid{}.
				Col(1, NewBuffer("A")).
				Col(0, NewBuffer("B")). // Hidden column.
				Col(1, NewBuffer("C")),
			image.Rect(0, 0, 3, 1),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'C'}, {},
			},
		},
		// Single Row tests.
		{
			Grid{}.Row(0),
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{}.Row(1),
			image.Rect(0, 0, 3, 3),
			make([]termbox.Cell, 9),
		},
		{
			Grid{}.Row(1).Col(1, NewBuffer("ABCD")),
			image.Rect(0, 0, 3, 3),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'}, {Ch: 'C'},
				{Ch: 'D'}, {}, {},
				{}, {}, {},
			},
		},
		{
			Grid{}.Row(100).Col(100, NewBuffer("ABCDE")),
			image.Rect(1, 1, 3, 3),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'B'},
				{Ch: 'C'}, {Ch: 'D'},
			},
		},
		// Multiple Row tests.
		{
			Grid{}.Row(0).Row(0).Row(0),
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{}.Row(1).Row(1).Row(1),
			image.ZR,
			[]termbox.Cell{},
		},
		{
			Grid{}.Row(1).Row(1).Row(1),
			image.Rect(2, 2, 3, 3),
			make([]termbox.Cell, 1),
		},
		{
			Grid{}.Row(1).Col(1, NewBuffer("XYZ")).
				Row(1).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 2, 6),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'},
				{Ch: 'Z'}, {},
				{}, {},
				{Ch: 'M'}, {Ch: 'N'},
				{Ch: 'O'}, {},
				{}, {},
			},
		},
		{
			Grid{}.Row(1).Col(1, NewBuffer("XYZ")).
				Row(1).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 1, 4),
			[]termbox.Cell{
				{Ch: 'X'},
				{Ch: 'Y'},
				{Ch: 'M'},
				{Ch: 'N'},
			},
		},
		{
			Grid{}.Row(1).Col(1, NewBuffer("XYZ")).
				Row(1).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 2, 4),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'},
				{Ch: 'Z'}, {},
				{Ch: 'M'}, {Ch: 'N'},
				{Ch: 'O'}, {},
			},
		},
		{
			// The rows do not fit perfectly, there's an extra
			// empty row in the far bottom that could not be split
			// between the two rows.
			Grid{}.Row(1).Col(1, NewBuffer("XYZ")).
				Row(1).Col(1, NewBuffer("MNO")),
			image.Rect(0, 0, 2, 3),
			[]termbox.Cell{
				{Ch: 'X'}, {Ch: 'Y'},
				{Ch: 'M'}, {Ch: 'N'},
				{}, {},
			},
		},
		{
			Grid{}.Row(1).Col(1, NewBuffer("A")).
				Row(0).Col(1, NewBuffer("B")). // Hidden row.
				Row(1).Col(1, NewBuffer("C")),
			image.Rect(0, 0, 1, 3),
			[]termbox.Cell{
				{Ch: 'A'}, {Ch: 'C'}, {},
			},
		},
		// Complex layout tests.
		{
			Grid{}.
				Row(1).
				Col(1, NewBuffer("A")).
				Col(2, NewBuffer("B")).
				Col(1, NewBuffer("C")).
				Row(2).
				Col(3, NewBuffer("D")).
				Col(2, NewBuffer("E")).
				Col(1, NewBuffer("F")).
				Row(3).
				Col(1, NewBuffer("G")).
				Col(1, NewBuffer("H")),
			image.Rect(-2, -2, 10, 6),
			[]termbox.Cell{
				{Ch: 'A'}, {}, {}, {Ch: 'B'}, {}, {}, {}, {}, {}, {Ch: 'C'}, {}, {},
				{Ch: 'D'}, {}, {}, {}, {}, {}, {Ch: 'E'}, {}, {}, {}, {Ch: 'F'}, {},
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
				{Ch: 'G'}, {}, {}, {}, {}, {}, {Ch: 'H'}, {}, {}, {}, {}, {},
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
			},
		},
	}
	for _, test := range tests {
		got := test.fitter.Fit(test.rect)
		gotBounds := got.Bounds()
		if !reflect.DeepEqual(gotBounds, test.rect) {
			t.Fatalf("%v.Fit(%v).Bounds() = %v, want %v", test.fitter, test.rect, gotBounds, test.rect)
		}
		gotCells := got.CellBuffer()
		if !reflect.DeepEqual(gotCells, test.want) {
			t.Fatalf("%v.Fit(%v).CellBuffer() = %v, want %v", test.fitter, test.rect, gotCells, test.want)
		}
	}
}
