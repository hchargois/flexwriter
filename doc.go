/*
Package flexwriter arranges rows of data into columns with configurable
widths and alignments.

Flexwriter tries to match a given output size, which by default is the width of
the terminal (if the output is the standard output, which is the default, and it
is a terminal). Depending on the content and the configuration of the columns,
flexwriter will wrap text in multiple lines in order to fit the target output
width.

The columns can be configured to be one of two types, which affects how they're
sized:

  - a [Rigid] is sized depending on its content, regardless of the width
    of the output;
  - a [Flexed] is sized depending on the width of the output, regardless
    of the size of its content.

A column can also be omitted from the output by using the [Omit] column type.

The size of the columns is determined in this way:

  - the widths of Rigid columns are computed first, depending on the size
    of their contents and their configured min and max widths;
  - then, the widths of the Flexed are computed to fill the remaining
    space, proportional to their weights.

The output can be decorated with simple column separators or to look like a
table, and any decorator can be colorized.

Flexwriter correctly supports aligning and wrapping text even if it contains
ANSI escape sequences, such as color codes.
*/
package flexwriter
