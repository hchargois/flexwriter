# Flexwriter

[![Go Reference](https://pkg.go.dev/badge/github.com/hchargois/flexwriter.svg)](https://pkg.go.dev/github.com/hchargois/flexwriter)

```
go get github.com/hchargois/flexwriter
```

Flexwriter arranges rows of data into columns with configurable widths and alignments.

As the name suggests, it implements the CSS flexbox model to define column widths.

If the contents are too long, flexwriter automatically wraps the text over multiple lines.
Text containing escape sequences (e.g. color codes) is correctly wrapped.

The output can be decorated with simple column separators or to look like tables.

![demo screenshot showing features such as flexed columns, alignments and color support](demo.png)

# Basic usage

```go
import "github.com/hchargois/flexwriter"

// by default, the flexwriter will output to standard output; and all
// columns will default to being "shrinkable" columns (i.e. they will match
// their content size if it fits within the output width, but will shrink to
// match the output width if the content is too big to fit on a single line);
// and all columns will be separated by two spaces
writer := flexwriter.New()

// write some data (any non-string will pass through fmt.Sprint)
writer.WriteRow("deep", "thought", "says", ":")
writer.WriteRow("the", "answer", "is", 42)
writer.WriteRow(true, "or", false, "?")

// calling Flush() is required to actually output the rows
writer.Flush()
```

This will output:

```
deep  thought  says   :
the   answer   is     42
true  or       false  ?
```

Here's another example showing how to configure the columns and set a table
decorator, and that shows how the Shrinkable columns shrinks to fit in the
configured width of the output (70 columns wide):

```go
writer := flexwriter.New()
writer.SetColumns(
    // first column, a Rigid, will not shrink and wrap
    flexwriter.Rigid{},
    // second column will
    flexwriter.Shrinkable{})
writer.SetDecorator(flexwriter.AsciiTableDecorator())
writer.SetWidth(70)

lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, "+
"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim "+
"ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip"

writer.WriteRow("lorem ipsum says:", lorem)

writer.Flush()
```

This outputs:

```
+-------------------+------------------------------------------------+
| lorem ipsum says: | Lorem ipsum dolor sit amet, consectetur        |
|                   | adipiscing elit, sed do eiusmod tempor         |
|                   | incididunt ut labore et dolore magna aliqua.   |
|                   | Ut enim ad minim veniam, quis nostrud          |
|                   | exercitation ullamco laboris nisi ut aliquip   |
+-------------------+------------------------------------------------+
```

Many more examples can be found in the [godoc](https://pkg.go.dev/github.com/hchargois/flexwriter).

# Alternatives

 - the OG, standard library's [text/tabwriter](https://pkg.go.dev/text/tabwriter)
 - for a more full-fledged table writer, see [github.com/olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)
   but note that as of Jan 2025 it doesn't correctly handle wrapping text
   with escape strings.

# Thanks

 - [Gio UI](https://gioui.org) for the initial inspiration to use the [flex model](https://pkg.go.dev/gioui.org/layout#Flex)
 - [github.com/MichaelMure/go-term-text](https://github.com/MichaelMure/go-term-text) for the escape-sequence aware text wrapping
