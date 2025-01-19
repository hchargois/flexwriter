package flexwriter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransposeDense(t *testing.T) {
	out := transpose([][]int{
		{1, 2},
		{3, 4},
		{5, 6},
	})
	assert.Equal(t, [][]int{
		{1, 3, 5},
		{2, 4, 6},
	}, out)
}

func TestTransposeSparse(t *testing.T) {
	out := transpose([][]int{
		{1, 2},
		{3},
		{4, 5, 6},
	})
	assert.Equal(t, [][]int{
		{1, 3, 4},
		{2, 0, 5},
		{0, 0, 6},
	}, out)
}
