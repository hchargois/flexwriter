package flexwriter_test

import "github.com/hchargois/flexwriter"

func Example() {
	// by default, the flexwriter will output to standard output; and all
	// columns will default to being left-aligned rigids with no maximum width
	// (i.e. they will be exactly as wide as needed to fit their content);
	// and all columns will be separated by two spaces
	writer := flexwriter.New()

	// write some data (any non-string will pass through fmt.Sprint)
	writer.WriteRow("deep", "thought", "says", ":")
	writer.WriteRow("the", "answer", "is", 42)
	writer.WriteRow(true, "or", false, "?")

	// calling Flush() is required to actually output the rows
	writer.Flush()
	// Output:
	// deep  thought  says   :
	// the   answer   is     42
	// true  or       false  ?
}

func Example_configuration() {
	writer := flexwriter.New()
	writer.SetColumns(
		// the first column will have a width equal to the width of its content,
		// same as the default for a new flexwriter as shown above
		flexwriter.Rigid{},

		// the second column will have a width equal to the width of its content
		// but only if it's less than 10 characters, otherwise it will be 10
		// characters wide and longer content will be wrapped
		flexwriter.Rigid{Max: 10},

		// the third column is a flexed with an implicit weight of 1, so it
		// will take one third of the remaining space
		flexwriter.Flexed{},

		// the fourth column is a right-aligned flexed that will take two thirds
		// of the remaining space
		flexwriter.Flexed{Weight: 2, Align: flexwriter.Right},
	)
	// any additional column will be exactly 5 characters wide and centered
	writer.SetDefaultColumn(flexwriter.Rigid{Min: 5, Max: 5, Align: flexwriter.Center})

	// use a decorator to make columns stand out better
	writer.SetDecorator(flexwriter.GapDecorator{Left: "| ", Gap: " | ", Right: " |"})

	writer.WriteRow(1, "hello", "world", "what's up", "A")
	writer.WriteRow(2, "this text is quite long", "so", "it will wrap", "B")
	writer.WriteRow(3, "I", "like", "bunnies", "C")

	writer.Flush()
	// Output:
	// | 1 | hello      | world            |                        what's up |   A   |
	// | 2 | this text  | so               |                     it will wrap |   B   |
	// |   | is quite   |                  |                                  |       |
	// |   | long       |                  |                                  |       |
	// | 3 | I          | like             |                          bunnies |   C   |
}

func ExampleRigid() {
	writer := flexwriter.New()
	writer.SetColumns(
		flexwriter.Rigid{},
		flexwriter.Rigid{Min: 20},
		flexwriter.Rigid{Max: 20},
	)
	writer.SetDecorator(flexwriter.GapDecorator{Left: "| ", Gap: " | ", Right: " |"})

	writer.WriteRow("sized to content", "min 20 wide", "maximum of 20 characters, longer content will wrap")

	writer.Flush()
	// Output:
	// | sized to content | min 20 wide          | maximum of 20        |
	// |                  |                      | characters, longer   |
	// |                  |                      | content will wrap    |
}

func ExampleFlexed() {
	writer := flexwriter.New()
	writer.SetColumns(
		flexwriter.Flexed{},
		flexwriter.Flexed{Weight: 2},
		flexwriter.Flexed{Weight: 3},
	)
	writer.SetDecorator(flexwriter.GapDecorator{Left: "| ", Gap: " | ", Right: " |"})

	writer.WriteRow("one sixth,", "one third,", "and half of the output width")

	writer.Flush()
	// Output:
	// | one sixth,  | one third,              | and half of the output width         |
}

func ExampleWriter_SetDefaultColumn() {
	writer := flexwriter.New()
	writer.SetColumns(flexwriter.Rigid{})
	writer.SetDefaultColumn(flexwriter.Flexed{})

	writer.WriteRow(
		"first column is sized to content",
		"all other columns",
		"will share the rest of the output width",
		"equally and wrap as needed.")

	writer.Flush()
	// Output:
	// first column is sized to content  all other       will share the  equally and
	//                                   columns         rest of the     wrap as
	//                                                   output width    needed.
}

func ExampleAsciiTableDecorator() {
	writer := flexwriter.New()
	writer.SetDecorator(flexwriter.AsciiTableDecorator())

	writer.WriteRow("a", "nice", "table")
	writer.WriteRow("with", "a classic", "look")

	writer.Flush()
	// Output:
	// +------+-----------+-------+
	// | a    | nice      | table |
	// +------+-----------+-------+
	// | with | a classic | look  |
	// +------+-----------+-------+
}

func ExampleBoxDrawingTableDecorator() {
	writer := flexwriter.New()
	writer.SetDecorator(flexwriter.BoxDrawingTableDecorator())

	writer.WriteRow("a", "nice", "table")
	writer.WriteRow("with", "a modern", "look")

	writer.Flush()
	// Output:
	// ┌──────┬──────────┬───────┐
	// │ a    │ nice     │ table │
	// ├──────┼──────────┼───────┤
	// │ with │ a modern │ look  │
	// └──────┴──────────┴───────┘
}
