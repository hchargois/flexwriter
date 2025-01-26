package flexwriter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	text "github.com/MichaelMure/go-term-text"
	"github.com/hchargois/flexwriter/flex"
	"golang.org/x/term"
)

// Column holds the configuration of a column for the flex writer. This
// interface is sealed, use one of the provided implementations:
//   - [Flexbox]
//   - [Rigid]
//   - [Flexed]
//   - [Shrinkable]
//   - [Omit]
type Column interface {
	flex() flexItem
}

type flexItem struct {
	flex.Item
	Alignment
}

// Rigid columns try to match the size of their content, as long
// as it is between Min and Max, regardless of the width of the output.
//
// By setting Min and Max to the same value, you can create a column of a fixed
// width.
//
// A Rigid is actually just a shortcut for a [Flexbox] with an Auto Basis, and
// Grow and Shrink factors of 0. This is similar to a "flex: none" in CSS.
type Rigid struct {
	// Min is the minimum width of the column. If the content is smaller, the
	// column will be padded.
	Min int
	// Max is the maximum width of the column, if the content is longer it will
	// be wrapped. If Max is 0, then there is no maximum width.
	Max int
	// Align is the alignment of the content within the column; default is left.
	Align Alignment
}

func (r Rigid) flex() flexItem {
	if r.Max != 0 && r.Min > r.Max {
		r.Min = r.Max
	}
	return flexItem{
		Item: flex.Item{
			Basis: Auto,
			Min:   r.Min,
			Max:   r.Max,
		},
		Alignment: r.Align,
	}
}

// Shrinkable columns try to match the size of their content, but if the width of
// the output is too small, they can shrink up to their Min width.
//
// A Shrinkable is actually just a shortcut for a [Flexbox] with an Auto Basis, a
// Grow of 0 and a Shrinkable of the given Weight, or 1 if 0/unset. This is similar
// to a "flex: initial" in CSS.
type Shrinkable struct {
	// Weight is the shrink weight of the column; if 0 or less, it defaults to 1.
	Weight int
	// Min is the minimum width of the column. If the content is smaller, the
	// column will be padded.
	Min int
	// Max is the maximum width of the column, if the content is longer it will
	// be wrapped. If Max is 0, then there is no maximum width.
	Max int
	// Align is the alignment of the content within the column; default is left.
	Align Alignment
}

func (s Shrinkable) flex() flexItem {
	if s.Max != 0 && s.Min > s.Max {
		s.Min = s.Max
	}
	if s.Weight < 1 {
		s.Weight = 1
	}
	return flexItem{
		Item: flex.Item{
			Basis:  Auto,
			Shrink: s.Weight,
			Min:    s.Min,
			Max:    s.Max,
		},
		Alignment: s.Align,
	}
}

// Omit columns will not appear in the output.
//
// This can be very useful in cases where it's simpler to modify the columns
// configuration than the row data.
type Omit struct{}

func (o Omit) flex() flexItem {
	panic("Omit.flexed() should not be called")
}

// Auto can be set as the Basis of a [Flexbox] column to make the basis as large
// as the content.
const Auto = -1

// Flexed columns take a size proportional to their weight (vs all other flexed
// columns weights) within the available width, regardless of the size of their
// content.
//
// A Flexed column is similar to a "flex: N" column in CSS.
type Flexed struct {
	// Weight is the grow weight of the column; if 0 or less, it defaults to 1.
	Weight int
	// Min is the minimum width of the column. If 0, it defaults to the
	// "min content" size, i.e. the size of the longest word in the content.
	Min int
	// Max is the maximum width of the column, if the content is longer it will
	// be wrapped. If Max is 0, then there is no maximum width.
	Max int
	// Align is the alignment of the content within the column; default is left.
	Align Alignment
}

func (f Flexed) flex() flexItem {
	if f.Weight < 1 {
		f.Weight = 1
	}
	if f.Max != 0 && f.Min > f.Max {
		f.Min = f.Max
	}
	return flexItem{
		Item: flex.Item{
			Grow:   f.Weight,
			Shrink: 1,
			Basis:  0,
			Min:    f.Min,
			Max:    f.Max,
		},
		Alignment: f.Align,
	}
}

// Flexbox columns allow you to specify the exact flex attributes as in CSS
// flexbox; however note that default values are all zero, there are no "smart"
// defaults as when using the "flex: ..." CSS syntax.
type Flexbox struct {
	// Basis is the flexbox basis, i.e. the initial size of the column before
	// it grows or shrinks. Use the constant Auto (or -1) to make the basis
	// equal to the content size.
	Basis int
	// Grow is the flexbox grow weight.
	Grow int
	// Shrink is the flexbox shrink weight.
	Shrink int
	// Min is the minimum width of the column. If 0, it defaults to the
	// "min content" size, i.e. the size of the longest word in the content.
	Min int
	// Max is the maximum width of the column, if the content is longer it will
	// be wrapped. If Max is 0, then there is no maximum width.
	Max int
	// Align is the alignment of the content within the column; default is left.
	Align Alignment
}

func (f Flexbox) flex() flexItem {
	if f.Max != 0 && f.Min > f.Max {
		f.Min = f.Max
	}
	return flexItem{
		Item: flex.Item{
			Basis:  f.Basis,
			Grow:   f.Grow,
			Shrink: f.Shrink,
			Min:    f.Min,
			Max:    f.Max,
		},
		Alignment: f.Align,
	}
}

type Writer struct {
	width       int
	output      io.Writer
	omittedCols []bool     // whether each configured column is omitted
	omitDefault bool       // whether unconfigured columns are omitted
	columns     []flexItem // only non-omitted columns
	defaultCol  flexItem
	deco        Decorator

	mu        sync.Mutex
	buffer    []byte
	colBuffer [][]string
}

// SetColumns sets the configuration for the first len(cols) columns.
func (w *Writer) SetColumns(cols ...Column) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.omittedCols = make([]bool, len(cols))
	w.columns = nil
	for i, col := range cols {
		if _, ok := col.(Omit); ok {
			w.omittedCols[i] = true
			continue
		}
		w.columns = append(w.columns, col.flex())
	}
}

// SetDefaultColumn sets the default column configuration. This configuration is
// used when more columns are written than are configured with
// [Writer.SetColumns].
func (w *Writer) SetDefaultColumn(col Column) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := col.(Omit); ok {
		w.omitDefault = true
		return
	}
	w.omitDefault = false
	w.defaultCol = col.flex()
}

// SetOutput sets the output writer for this flex writer. If the output is a
// terminal, the width of the flex writer is automatically configured to be the
// width of the terminal. If auto-detection is not desired, call
// [Writer.SetWidth] after SetOutput.
func (w *Writer) SetOutput(out io.Writer) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if f, ok := out.(*os.File); ok {
		if term.IsTerminal(int(f.Fd())) {
			width, _, err := term.GetSize(int(f.Fd()))
			if err == nil && width > 0 {
				w.width = width
			}
		}
	}
	w.output = out
}

// SetWidth sets the target width of the output; note however that depending
// on the columns min width constraints, this may not be honored.
// The width is also set when the output is set with [Writer.SetOutput] and the
// output is a terminal. If you want to force a width even if the output is a
// terminal, call SetWidth after [Writer.SetOutput].
func (w *Writer) SetWidth(width int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.width = width
}

// SetDecorator sets the decorator for this flex writer.
func (w *Writer) SetDecorator(deco Decorator) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.deco = deco
}

// New creates a new flex writer with the default configuration:
//   - write to standard output
//   - a target width equal to the width of the standard output if it's a
//     terminal, otherwise 80
//   - a gap of 2 spaces between columns, none on the sides
//   - a default column setting of a left-aligned Shrinkable column
func New() *Writer {
	var writer Writer
	writer.SetWidth(80)
	writer.SetOutput(os.Stdout)
	writer.SetDefaultColumn(Shrinkable{})
	writer.SetDecorator(GapDecorator{Gap: "  "})
	return &writer
}

// Write writes row(s) to the flex writer; rows are delimited by a newline
// (`\n`) and within a row the columns are delimited by a tab (`\t`).
// This is mostly compatible with the [text/tabwriter] package.
// If possible, use [Writer.WriteRow] instead.
//
// This method only appends to an internal buffer and thus never returns an
// error. Call [Writer.Flush] to write the buffer to the output.
//
// You should not alternate between calls to Write and [Writer.WriteRow],
// unless you call [Writer.Flush] in between.
func (w *Writer) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer = append(w.buffer, b...)
	return len(b), nil
}

// WriteRow writes a single row of cells to the flex writer. If the
// cells are not strings, they are converted to strings using [fmt.Sprint].
//
// This is the recommended method for writing data to the flex writer.
//
// This method only appends to an internal buffer; call [Writer.Flush] to write
// the buffer to the output.
//
// You should not alternate between calls to [Writer.Write] and WriteRow,
// unless you call [Writer.Flush] in between.
func (w *Writer) WriteRow(cells ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writeRow(cells...)
}

func (w *Writer) writeRow(cells ...any) {
	var filteredCells []any
	for i, cell := range cells {
		if w.isOmitted(i) {
			continue
		}
		filteredCells = append(filteredCells, cell)
	}

	toString := func(a any) string {
		if s, ok := a.(string); ok {
			return s
		}
		return fmt.Sprint(a)
	}
	scells := transform(filteredCells, toString)
	w.colBuffer = append(w.colBuffer, scells)
}

func (w *Writer) isOmitted(i int) bool {
	if i < len(w.omittedCols) {
		return w.omittedCols[i]
	}
	return w.omitDefault
}

func (w *Writer) getColumnDef(i int) flexItem {
	if i < len(w.columns) {
		return w.columns[i]
	}
	return w.defaultCol
}

func (w *Writer) colMinContent(colIdx int) int {
	return max(transform(w.colBuffer, func(row []string) int {
		if colIdx >= len(row) {
			return 0
		}
		return minContent(row[colIdx])
	}))
}

func (w *Writer) computeWidths() []int {
	rowColLengths := transform(w.colBuffer, func(rows []string) []int {
		return transform(rows, text.Len)
	})
	colRowLengths := transpose(rowColLengths)
	colLengths := transform(colRowLengths, max)

	nColumns := len(colLengths)

	flexItems := make([]flex.Item, nColumns)
	for i := 0; i < nColumns; i++ {
		col := w.getColumnDef(i)

		var minSize int
		if col.Min > 0 {
			minSize = col.Min
		} else {
			minSize = w.colMinContent(i)
		}
		if col.Max > 0 && minSize > col.Max {
			minSize = col.Max
		}
		it := col.Item
		it.Min = minSize
		it.Size = colLengths[i]

		flexItems[i] = it
	}

	freeSpace := w.width - decoratorWidth(w.deco, nColumns)

	return flex.ResolveFlexLengths(flexItems, freeSpace)
}

func (w *Writer) flushBuffer() {
	rows := strings.Split(string(w.buffer), "\n")
	// remove trailing empty line
	if len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	for _, row := range rows {
		var cells []any
		for _, cell := range strings.Split(row, "\t") {
			cells = append(cells, cell)
		}
		w.writeRow(cells...)
	}
	w.buffer = nil
}

// Flush writes the contents of the internal buffer to the output. This also
// resets the internal buffer and the associated column widths.
func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.flushBuffer()
	widths := w.computeWidths()

	var out bytes.Buffer

	if hdr := w.deco.RowSeparator(0, widths); hdr != "" {
		out.WriteString(hdr + "\n")
	}
	for ri, row := range w.colBuffer {
		if ri == len(w.colBuffer)-1 {
			ri = -1
		} else {
			ri += 1
		}
		if len(row) < len(widths) {
			// pad rows with missing columns
			row = append(row, make([]string, len(widths)-len(row))...)
		}

		wrappedCols := make([][]string, len(row))
		for ci, col := range row {
			wrappedCols[ci] = wrap(col, widths[ci])
		}
		transposed := transpose(wrappedCols)
		for _, line := range transposed {
			out.WriteString(w.deco.ColumnSeparator(ri, 0))
			for ci, col := range line {
				colAlign := w.getColumnDef(ci).Alignment
				if ci != len(line)-1 {
					out.WriteString(align(col, widths[ci], colAlign, true))
					out.WriteString(w.deco.ColumnSeparator(ri, ci+1))
				} else {
					// last column is right-padded with spaces only if there is
					// a right separator, otherwise we avoid adding the extra
					// trailing spaces
					rightSep := w.deco.ColumnSeparator(ri, -1)
					if rightSep != "" {
						out.WriteString(align(col, widths[ci], colAlign, true))
						out.WriteString(rightSep)
					} else {
						out.WriteString(align(col, widths[ci], colAlign, false))
					}
				}
			}
			out.WriteByte('\n')
		}

		if sep := w.deco.RowSeparator(ri, widths); sep != "" {
			out.WriteString(sep + "\n")
		}
	}

	_, err := w.output.Write(out.Bytes())
	if err != nil {
		return err
	}

	w.colBuffer = nil
	return nil
}
