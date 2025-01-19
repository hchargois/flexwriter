package flexwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	assert.Equal(t, []string{"abc", "def", "gh"}, wrap("abcdefgh", 3))
}
