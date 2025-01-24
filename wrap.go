package flexwriter

import (
	"strings"

	text "github.com/MichaelMure/go-term-text"
	"github.com/mattn/go-runewidth"
)

func wrap(s string, width int) []string {
	if width <= 0 {
		panic("width must be > 0")
	}

	// strangely, text.Wrap doesn't return early if there is no need to wrap,
	// and is quite inefficient to "wrap" something that doesn't need to be;
	// so we check ourselves
	if text.Len(s) <= width {
		return []string{s}
	}

	wrapped, _ := text.Wrap(s, width)
	lines := strings.Split(wrapped, "\n")

	var state text.EscapeState
	for i, line := range lines {
		line = state.FormatString() + line
		state.Witness(line)
		line = line + state.ResetString()
		lines[i] = line
	}

	return lines
}

type Alignment int

const (
	Left Alignment = iota
	Center
	Right
)

func align(s string, width int, align Alignment, padRight bool) string {
	s = text.TrimSpace(s)

	padLen := width - text.Len(s)
	if padLen <= 0 {
		return s
	}

	switch align {
	case Center:
		padLeft := padLen / 2
		padLen -= padLeft
		s = strings.Repeat(" ", padLeft) + s
	case Right:
		return strings.Repeat(" ", padLen) + s
	}
	if !padRight {
		return s
	}
	return s + strings.Repeat(" ", padLen)
}

func minContent(s string) int {
	// adapted from go-term-text.segmentLine
	escaped, _ := text.ExtractTermEscapes(s)

	var max int

	var wordLen int
	wordType := none
	flushWord := func() {
		if wordLen > max {
			max = wordLen
		}
		wordLen = 0
		wordType = none
	}

	for _, r := range escaped {
		// A WIDE_CHAR itself constitutes a chunk.
		thisType, rw := runeTypeOf(r)
		if thisType == wideChar {
			if wordType != none {
				flushWord()
			}
			wordLen = rw
			flushWord()
			continue
		}
		// Other type of chunks starts with a char of that type, and ends with a
		// char with different type or end of string.
		if thisType != wordType {
			if wordType != none {
				flushWord()
			}
			wordLen = rw
			wordType = thisType
		} else {
			wordLen += rw
		}
	}
	if wordLen != 0 {
		flushWord()
	}

	return max
}

type runeType int

// Rune categories
//
// These categories are so defined that each category forms a non-breakable
// chunk. It IS NOT the same as unicode code point categories.
const (
	none runeType = iota
	wideChar
	invisible
	shortUnicode
	space
	visibleAscii
)

// Determine the category of a rune.
func runeTypeOf(r rune) (runeType, int) {
	rw := runewidth.RuneWidth(r)
	if rw > 1 {
		return wideChar, rw
	} else if rw == 0 {
		return invisible, rw
	} else if r > 127 {
		return shortUnicode, rw
	} else if r == ' ' {
		return space, rw
	} else {
		return visibleAscii, rw
	}
}
