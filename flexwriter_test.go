package flexwriter

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

const complexTest = "\x1b[1mThis is text in bold\x1b[0m and this is normal text. " +
	"\x1b[3mThis text is in italics\x1b[0m " +
	"and \x1b[9mthis one is strikethrough\x1b[0m. " +
	"\x1b[31mThis is in a red font color.\x1b[0m " +
	"This is double-width text: 私はフライドポテトです。"

func lorem(n int) string {
	const l = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim " +
		"ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip " +
		"ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate " +
		"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat " +
		"cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	return strings.Join(strings.Split(l, " ")[:n], " ")
}

var update = flag.Bool("update", false, "update golden files")

func assertGolden(t *testing.T, actual []byte, filename string) {
	golden := filepath.Join("testdata", filename)
	if *update {
		if err := os.WriteFile(golden, actual, 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}

func TestSimple(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetWidth(40)
	writer.SetColumns(
		Rigid{Max: 10},
		Flexed{},
		Rigid{Max: 10},
	)

	writer.WriteRow("hello", lorem(10), "woooooooooooorld")
	writer.WriteRow("helloooooooooo", lorem(10), "world")
	writer.Flush()

	assertGolden(t, buf.Bytes(), "simple.txt")
}

func TestMultipleFlushes(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)

	writer.WriteRow("hello", "world")
	writer.Flush()
	writer.WriteRow("how", "are", "you")
	writer.Flush()

	assertGolden(t, buf.Bytes(), "multiflush.txt")
}

func TestAligments(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetColumns(
		Rigid{Max: 20, Align: Left},
		Rigid{Max: 20, Align: Center},
		Rigid{Max: 20, Align: Right},
	)

	writer.WriteRow(lorem(10), lorem(10), lorem(10))
	writer.Flush()

	assertGolden(t, buf.Bytes(), "alignments.txt")
}

func TestFlexed(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetWidth(80)
	writer.SetColumns(
		Flexed{Weight: 1},
		Flexed{Weight: 2},
		Flexed{}, // same as Weight: 1
	)

	writer.WriteRow(lorem(10), lorem(20), lorem(12))
	writer.WriteRow(lorem(15), lorem(30), lorem(8))
	writer.Flush()

	assertGolden(t, buf.Bytes(), "flexed.txt")
}

func TestDefaultColumn(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)

	writer.WriteRow("helloooo", "world")
	writer.WriteRow("hello", "wooooorld")
	writer.WriteRow("oh", "hi", "mark")
	writer.Flush()

	assertGolden(t, buf.Bytes(), "defaultcol.txt")
}

func TestFlushWithoutWrite(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.Flush()

	assert.Equal(t, "", buf.String())
}

func TestEmptyWriteRow(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.WriteRow()
	writer.Flush()

	assert.Equal(t, "", buf.String())
}

func TestComplexText(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetColumns(
		Rigid{Max: 15},
		Rigid{Max: 15},
		Rigid{Max: 15},
	)

	writer.WriteRow(lorem(30), complexTest, lorem(30))
	writer.WriteRow(complexTest, lorem(30), complexTest)
	writer.Flush()

	assertGolden(t, buf.Bytes(), "complex.txt")
}

func TestTabWriterCompat(t *testing.T) {
	writeRows := func(w io.Writer) {
		fmt.Fprintf(w, "%s\t%s\n", "hello", "world")
		fmt.Fprintf(w, "%s\t%s\n", "helloooooooo", "world")
		fmt.Fprintf(w, "%s\t%s\n", "hello", "woooooooooooorld")
	}

	var bufFlex bytes.Buffer
	fwriter := New()
	fwriter.SetOutput(&bufFlex)
	writeRows(fwriter)
	fwriter.Flush()

	var bufTab bytes.Buffer
	twriter := tabwriter.NewWriter(&bufTab, 0, 0, 2, ' ', 0)
	writeRows(twriter)
	twriter.Flush()

	assert.Equal(t, bufFlex.String(), bufTab.String())
}

func TestTableDecorator(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetDefaultColumn(Rigid{Max: 15})
	writer.SetDecorator(BoxDrawingTableDecorator())

	writer.WriteRow(lorem(30), lorem(20), lorem(10))
	writer.WriteRow(lorem(10), lorem(30), lorem(20))
	writer.Flush()

	assertGolden(t, buf.Bytes(), "table.txt")
}

func TestTableDecoratorColor(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetDefaultColumn(Rigid{Max: 15})
	writer.SetDecorator(ColorizeDecorator(
		BoxDrawingTableDecorator(),
		color.New(color.FgYellow),
	))

	writer.WriteRow(lorem(30), lorem(20), lorem(10))
	writer.WriteRow(lorem(10), lorem(30), lorem(20))
	writer.Flush()

	assertGolden(t, buf.Bytes(), "tablecolor.txt")
}

func BenchmarkFlexwriter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		writer := New()
		writer.SetOutput(&buf)
		writer.SetDecorator(ColorizeDecorator(
			BoxDrawingTableDecorator(),
			color.New(color.FgYellow),
		))

		for i := 0; i < 50; i++ {
			writer.WriteRow("hello", "world", "how are you")
			writer.WriteRow("oh", "hi", "mark", 42)
		}
		writer.Flush()
	}
}

func TestDecoratorUnequalRowLens(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetDecorator(BoxDrawingTableDecorator())

	writer.WriteRow("ab", "cde")
	writer.WriteRow(1, 2, 3, 4)
	writer.WriteRow(1)
	writer.WriteRow()
	writer.WriteRow(1, 2, "fghij")
	writer.Flush()

	assertGolden(t, buf.Bytes(), "unequalrowlens.txt")
}

func TestDecoratorIndices(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetDecorator(debugDecorator{})

	writer.WriteRow("A", "B", "C")
	writer.WriteRow("D", "E", "F")
	writer.WriteRow("G", "H", "I")
	writer.WriteRow("J", "K", "L")

	writer.Flush()
	assertGolden(t, buf.Bytes(), "decorator.txt")
}

type debugDecorator struct{}

func (debugDecorator) RowSeparator(rowIdx int, widths []int) string {
	return fmt.Sprintf("--- %d ---", rowIdx)
}

func (debugDecorator) ColumnSeparator(rowIdx, colIdx int) string {
	return fmt.Sprintf(" %d/%d ", rowIdx, colIdx)
}

func TestOmit(t *testing.T) {
	var buf bytes.Buffer
	writer := New()
	writer.SetOutput(&buf)
	writer.SetColumns(
		Rigid{},
		Omit{},
		Rigid{},
	)
	writer.SetDefaultColumn(Omit{})
	writer.SetDecorator(BoxDrawingTableDecorator())

	writer.WriteRow("A", "B", "C", "D", "E")
	writer.WriteRow("F", "G", "H", "I", "J")
	writer.Flush()

	// same thing with the Writer interface
	fmt.Fprintln(writer, "A\tB\tC\tD\tE")
	fmt.Fprintln(writer, "F\tG\tH\tI\tJ")
	writer.Flush()

	assertGolden(t, buf.Bytes(), "omit.txt")
}
