package flexwriter

import (
	"strings"

	text "github.com/MichaelMure/go-term-text"
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
