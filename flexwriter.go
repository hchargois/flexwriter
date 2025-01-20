package flexwriter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	text "github.com/MichaelMure/go-term-text"
	"golang.org/x/term"
)

// Column holds the configuration of a column for the flex writer. This
// interface is sealed, use either [Rigid] or [Flexed] to create Columns.
type Column interface {
	// validated returns a copy of the column where impossible or zero values
	// have been replaced by correct defaults; doing it this way instead of
	// having a validate() method that modifies the column in place has two
	// advantages:
	// - a validate() method would need a pointer receiver, and the columns
	//   would need to be declared as pointers everywhere which is annoying;
	// - using a copy makes sure the user cannot keep and modify the column
	//   after it's used to configure the writer.
	validated() Column
	align() Alignment
}

// Rigid columns try to match the size of their content, as long
// as it is between Min and Max, regardless of the width of the output.
//
// By setting Min and Max to the same value, you can create a column of a fixed
// width.
type Rigid struct {
	// Min is the minimum width of the column.
	Min int
	// Max is the maximum width of the column, if the content is longer it will
	// be wrapped. If Max is 0, then there is no maximum width.
	Max int
	// Align is the alignment of the content within the column; default is left.
	Align Alignment
}

func (r Rigid) validated() Column {
	if r.Max != 0 && r.Min > r.Max {
		r.Min = r.Max
	}
	return r
}

func (r Rigid) width(dataWidth int) int {
	if dataWidth < r.Min {
		return r.Min
	}
	if r.Max != 0 && dataWidth > r.Max {
		return r.Max
	}
	return dataWidth
}

func (r Rigid) align() Alignment {
	return r.Align
}

// Omit columns will not appear in the output.
//
// This can be very useful in cases where it's simpler to modify the columns
// configuration than the row data.
type Omit struct{}

func (o Omit) validated() Column {
	return o
}

func (o Omit) align() Alignment {
	return Left
}

// Flexed columns take a size proportional to their weight (vs all other flexed
// columns weights), within the width remaining after rigid columns are placed,
// and regardless of the size of their content.
type Flexed struct {
	// Weight is the flex weight of the column; if 0 or less, it defaults to 1.
	Weight int
	// Min is the minimum width of the column. Flexed columns need to have a
	// min width because we need to ensure they have a sufficient width even if
	// the output (e.g. terminal) width is too small to fit all the columns.
	// If 0 or less, it defaults to 10.
	Min int
	// Align is the alignment of the content within the column; default is left.
	Align Alignment
}

func (f Flexed) validated() Column {
	if f.Weight < 1 {
		f.Weight = 1
	}
	if f.Min < 1 {
		f.Min = 10
	}
	return f
}

func (f Flexed) width(remainingWidth int, totalWeights int) int {
	w := remainingWidth * f.Weight / totalWeights
	if w < f.Min {
		return f.Min
	}
	return w
}
func (f Flexed) align() Alignment {
	return f.Align
}

type Writer struct {
	width           int
	output          io.Writer
	defaultCol      Column
	columns         []Column // all column defs including Omits
	filteredColumns []Column // column defs without Omits
	deco            Decorator

	mu        sync.Mutex
	buffer    []byte
	colBuffer [][]string
}

// SetColumns sets the configuration for the first len(cols) columns.
func (w *Writer) SetColumns(cols ...Column) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.columns = transform(cols, func(col Column) Column {
		return col.validated()
	})
	for _, col := range w.columns {
		if _, ok := col.(Omit); ok {
			continue
		}
		w.filteredColumns = append(w.filteredColumns, col)
	}
}

// SetDefaultColumn sets the default column configuration. This configuration is
// used when more columns are written than are configured with
// [Writer.SetColumns].
func (w *Writer) SetDefaultColumn(col Column) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.defaultCol = col.validated()
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
//   - a default column setting of an unlimited width, left-aligned rigid column
func New() *Writer {
	var writer Writer
	writer.SetWidth(80)
	writer.SetOutput(os.Stdout)
	writer.SetDefaultColumn(Rigid{})
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
		if _, ok := w.getColumnDef(i).(Omit); ok {
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

func (w *Writer) getColumnDef(i int) Column {
	if i < len(w.columns) {
		return w.columns[i]
	}
	return w.defaultCol
}

func (w *Writer) getFilteredColumnDef(i int) Column {
	if i < len(w.filteredColumns) {
		return w.filteredColumns[i]
	}
	return w.defaultCol
}

func (w *Writer) computeWidths() []int {
	rowColLengths := transform(w.colBuffer, func(rows []string) []int {
		return transform(rows, text.Len)
	})
	colRowLengths := transpose(rowColLengths)
	colLengths := transform(colRowLengths, max)

	nColumns := len(colLengths)

	widths := make([]int, nColumns)

	var sumRigidWidths int
	var totalWeights int
	for i := range nColumns {
		col := w.getFilteredColumnDef(i)
		switch tcol := col.(type) {
		case Rigid:
			w := tcol.width(colLengths[i])
			widths[i] = w
			sumRigidWidths += w
		case Flexed:
			totalWeights += tcol.Weight
		}
	}

	remainingWidth := w.width - sumRigidWidths
	remainingWidth -= decoratorWidth(w.deco, nColumns)
	for i := range nColumns {
		col := w.getFilteredColumnDef(i)
		if flexed, ok := col.(Flexed); ok {
			w := flexed.width(remainingWidth, totalWeights)
			remainingWidth -= w
			totalWeights -= flexed.Weight
			widths[i] = w
		}
	}
	return widths
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
				colAlign := w.getFilteredColumnDef(ci).align()
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
