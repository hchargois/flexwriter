package flexwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoratorWidth(t *testing.T) {
	deco := GapDecorator{
		Left:  " ",
		Gap:   "  ",
		Right: "   ",
	}

	assert.Equal(t, 4, decoratorWidth(deco, 0))
	assert.Equal(t, 4, decoratorWidth(deco, 1))
	assert.Equal(t, 6, decoratorWidth(deco, 2))
	assert.Equal(t, 8, decoratorWidth(deco, 3))
}
