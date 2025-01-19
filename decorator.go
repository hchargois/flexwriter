package flexwriter

import (
	"strings"

	text "github.com/MichaelMure/go-term-text"
	"github.com/fatih/color"
)

// Decorator is used to decorate the output. It can be used to add spacing
// between columns, or create borders in order to create a table.
type Decorator interface {
	// RowSeparator defines the horizontal separator between rows, as well as
	// before the first row and after the last one. rowIdx will be 0 for the
	// separator before the first row, -1 for the separator after the last row.
	// It will be N for the separator just below the Nth row.
	// If the empty string is returned, rows will not be separated.
	// widths are the widths of the column without any padding.
	RowSeparator(rowIdx int, widths []int) string

	// ColumnSeparator defines the vertical separator between columns.
	// For a given colIdx, the length of the string returned should stay
	// constant for all rows, otherwise the result will not be properly aligned.
	// colIdx starts at 0 for the separator left of the first column, and ends
	// at -1 for the separator right of the last column; otherwise it is N for
	// the separator just to the right of the Nth column.
	// rowIdx starts at 1 for the first row, and ends at -1 for the last row.
	ColumnSeparator(rowIdx, colIdx int) string
}

// GapDecorator is a simple decorator that adds a fixed gap between each column,
// as well as a left gap (before the left-most column) and a right gap (after the
// right-most column).
type GapDecorator struct {
	Gap   string
	Left  string
	Right string
}

func (d GapDecorator) RowSeparator(rowIdx int, widths []int) string {
	return ""
}

func (d GapDecorator) ColumnSeparator(_, colIdx int) string {
	switch colIdx {
	case 0:
		return d.Left
	default:
		return d.Gap
	case -1:
		return d.Right
	}
}

// TableDecorator is a decorator that creates a table with configurable borders
// and intersections.
type TableDecorator struct {
	TopIntersections    [3]string // (left, middle, right) top intersections
	MiddleIntersections [3]string // etc.
	BottomIntersections [3]string
	VertBorders         [3]string
	HorizBorders        [3]string // (top, middle, bottom), must be of width 1, will be repeated as needed
}

func (d TableDecorator) rowSep(intersects [3]string, horiz string, widths []int) string {
	borders := transform(widths, func(w int) string {
		return strings.Repeat(horiz, w)
	})
	return intersects[0] +
		strings.Join(borders, intersects[1]) +
		intersects[2]
}

func (d TableDecorator) RowSeparator(rowIdx int, widths []int) string {
	switch rowIdx {
	case 0:
		return d.rowSep(d.TopIntersections, d.HorizBorders[0], widths)
	default:
		return d.rowSep(d.MiddleIntersections, d.HorizBorders[1], widths)
	case -1:
		return d.rowSep(d.BottomIntersections, d.HorizBorders[2], widths)
	}
}

func (d TableDecorator) ColumnSeparator(_, colIdx int) string {
	switch colIdx {
	case 0:
		return d.VertBorders[0]
	default:
		return d.VertBorders[1]
	case -1:
		return d.VertBorders[2]
	}
}

// AsciiTableDecorator creates a table with ASCII characters + and - for an
// old-school look.
func AsciiTableDecorator() Decorator {
	return &TableDecorator{
		TopIntersections:    [3]string{"+-", "-+-", "-+"},
		MiddleIntersections: [3]string{"+-", "-+-", "-+"},
		BottomIntersections: [3]string{"+-", "-+-", "-+"},
		VertBorders:         [3]string{"| ", " | ", " |"},
		HorizBorders:        [3]string{"-", "-", "-"},
	}
}

// BoxDrawingTableDecorator creates a table with Unicode box drawing characters.
func BoxDrawingTableDecorator() Decorator {
	return &TableDecorator{
		TopIntersections:    [3]string{"┌─", "─┬─", "─┐"},
		MiddleIntersections: [3]string{"├─", "─┼─", "─┤"},
		BottomIntersections: [3]string{"└─", "─┴─", "─┘"},
		VertBorders:         [3]string{"│ ", " │ ", " │"},
		HorizBorders:        [3]string{"─", "─", "─"},
	}
}

type colorDecorator struct {
	parent Decorator
	in     string
	out    string
}

// ColorizeDecorator wraps a decorator to make it colorful.
func ColorizeDecorator(parent Decorator, color *color.Color) Decorator {
	// fatih/color is very badly designed and is extremely inefficient, but we
	// can improve the situation by first making it colorize a string, use it
	// to extract the in and out escape strings, and then use those with simple
	// concatenation.
	cut := "__CUT_HERE__"
	colored := color.Sprint(cut)
	in, out, _ := strings.Cut(colored, cut)
	return colorDecorator{
		parent: parent,
		in:     in,
		out:    out,
	}
}

func (d colorDecorator) RowSeparator(rowIdx int, widths []int) string {
	return d.in + d.parent.RowSeparator(rowIdx, widths) + d.out
}

func (d colorDecorator) ColumnSeparator(rowIdx, colIdx int) string {
	return d.in + d.parent.ColumnSeparator(rowIdx, colIdx) + d.out
}

func decoratorWidth(deco Decorator, cols int) int {
	rlen := text.Len
	var w int
	w += rlen(deco.ColumnSeparator(0, 0))
	w += rlen(deco.ColumnSeparator(0, -1))
	for i := 1; i < cols; i++ {
		w += rlen(deco.ColumnSeparator(0, i))
	}
	return w
}
