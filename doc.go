/*
Package flexwriter arranges rows of data into columns with configurable
widths and alignments.

Flexwriter tries to match a given output size, which by default is the width of
the terminal (if the output is the standard output, which is the default, and it
is a terminal). Depending on the content and the configuration of the columns,
flexwriter will wrap text in multiple lines in order to fit the target output
width.

The size of each column is determined by the flexbox algorithm.

The Basis, Grow, and Shrink properties of the flexbox model can be configured
for each column by using the [Flexbox] column type.

Moreover, a few "preconfigured" column types are provided for common use cases:

  - a [Rigid] is an inflexible column, it is sized depending on its content,
    regardless of the width of the output;
  - a [Flexed] is a "proportional" or "absolute" flexed column, expanding to
    take a specified share of the output width, regardless of the size of its
    content;
  - a [Shrinkable] is a column that is the same size as its content if it's
    small enough, but it can shrink as needed to fit the width of the output if
    the content is bigger.

These 3 column types are similar to the "flex: none", "flex: N", and
"flex: initial" CSS shorthand values, respectively.

A column can also be omitted from the output by using the special [Omit] column
type.

The output can be decorated with simple column separators or to look like a
table, and any decorator can be colorized.

Flexwriter correctly supports aligning and wrapping text even if it contains
ANSI escape sequences, such as color codes.
*/
package flexwriter
