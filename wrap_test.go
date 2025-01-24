package flexwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	assert.Equal(t, []string{"abc", "def", "gh"}, wrap("abcdefgh", 3))
}

func TestMinContent(t *testing.T) {
	assert.Equal(t, 0, minContent(""))
	assert.Equal(t, 1, minContent("a b c d"))
	assert.Equal(t, 28, minContent("a long word in the English language is antidisestablishmentarianism"))
	assert.Equal(t, 34, minContent("supercalifragilisticexpialidocious is even longer"))
	assert.Equal(t, 2, minContent("私はフライドポテトです。"))
	assert.Equal(t, 6, minContent("私はフライドpotatoです。"))
}
