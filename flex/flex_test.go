package flex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlexGrowBasisZero(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Grow: 1},
			},
			exp: []int{60},
		},
		{
			items: []Item{
				{Grow: 1},
				{Grow: 2},
			},
			exp: []int{20, 40},
		},
		{
			items: []Item{
				{Grow: 1},
				{Grow: 2},
				{Grow: 1},
			},
			exp: []int{15, 30, 15},
		},
		{
			items: []Item{
				{Grow: 1},
				{Grow: 2},
				{Grow: 3},
			},
			exp: []int{10, 20, 30},
		},
		{
			items: []Item{
				{Grow: 3},
				{Grow: 2},
				{Grow: 1},
			},
			exp: []int{30, 20, 10},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 60)
		assert.Equal(t, tc.exp, lens)
	}
}

func TestNullContainerSize(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Basis: 10, Shrink: 1},
			},
			exp: []int{1},
		},
		{
			items: []Item{
				{Min: 10, Shrink: 1},
			},
			exp: []int{10},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 0)
		assert.Equal(t, tc.exp, lens)
	}
}

func TestFlexGrowBasisNonZero(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Grow: 1, Basis: 10},
			},
			exp: []int{60},
		},
		{
			items: []Item{
				{Grow: 1, Basis: 10},
				{Grow: 2},
			},
			exp: []int{26, 34},
		},
		{
			items: []Item{
				{Grow: 1},
				{Grow: 2, Basis: 10},
			},
			exp: []int{16, 44},
		},
		{
			items: []Item{
				{Grow: 1},
				{Grow: 2, Size: 10, Basis: -1},
			},
			exp: []int{16, 44},
		},
		{
			items: []Item{
				{Grow: 1, Basis: 10},
				{Grow: 2, Basis: 10},
			},
			exp: []int{23, 37},
		},
		{
			items: []Item{
				{Grow: 1},
				{Grow: 2, Basis: 20},
				{Grow: 1},
			},
			exp: []int{10, 40, 10},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 60)
		assert.Equal(t, tc.exp, lens)
	}
}

func TestFlexInflexible(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Basis: 10},
			},
			exp: []int{10},
		},
		{
			items: []Item{
				{Size: 10, Basis: -1},
			},
			exp: []int{10},
		},
		{
			items: []Item{
				{Basis: 10},
				{Size: 20, Basis: -1},
			},
			exp: []int{10, 20},
		},
		{
			items: []Item{
				{Basis: 100},
			},
			exp: []int{100},
		},
		{
			items: []Item{
				{Size: 100, Basis: -1},
			},
			exp: []int{100},
		},
		{
			items: []Item{
				{Basis: 100},
				{Size: 200, Basis: -1},
			},
			exp: []int{100, 200},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 60)
		assert.Equal(t, tc.exp, lens)
	}
}

func TestFlexShrink(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Size: 80, Basis: -1, Shrink: 1},
			},
			exp: []int{60},
		},
		{
			items: []Item{
				{Size: 80, Basis: -1, Shrink: 1},
				{Size: 80, Basis: -1, Shrink: 1},
			},
			exp: []int{30, 30},
		},
		{
			items: []Item{
				{Size: 160, Basis: -1, Shrink: 1},
				{Size: 80, Basis: -1, Shrink: 1},
			},
			exp: []int{40, 20},
		},
		{
			items: []Item{
				{Size: 60, Basis: -1, Shrink: 2},
				{Size: 60, Basis: -1, Shrink: 1},
			},
			exp: []int{20, 40},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 60)
		assert.Equal(t, tc.exp, lens)
	}
}

func TestFlexMixed(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Basis: 30},
				{Grow: 1},
			},
			exp: []int{30, 30},
		},
		{
			items: []Item{
				{Basis: 40},
				{Basis: 40, Shrink: 1},
			},
			exp: []int{40, 20},
		},
		{
			items: []Item{
				{Basis: 30},
				{Grow: 1},
				{Grow: 2},
			},
			exp: []int{30, 10, 20},
		},
		{
			items: []Item{
				{Basis: 40},
				{Basis: 40, Shrink: 1},
				{Basis: 40, Shrink: 1},
			},
			exp: []int{40, 10, 10},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 60)
		assert.Equal(t, tc.exp, lens)
	}
}

func TestFlexMixed2(t *testing.T) {
	for _, tc := range []struct {
		items []Item
		exp   []int
	}{
		{
			items: []Item{
				{Grow: 0, Shrink: 1, Basis: -1, Size: 20, Min: 20},
				{Grow: 0, Shrink: 1, Basis: -1, Size: 3, Min: 3},
				{Grow: 0, Shrink: 1, Basis: -1, Size: 10, Min: 5},
				{Grow: 0, Shrink: 1, Basis: -1, Size: 3, Min: 3},
				{Grow: 0, Shrink: 1, Basis: -1, Size: 3, Min: 3},
			},
			exp: []int{20, 3, 10, 3, 3},
		},
		{
			items: []Item{
				{Grow: 1, Shrink: 1, Basis: 0, Size: 20, Min: 20},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 3, Min: 3},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 10, Min: 5},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 3, Min: 3},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 3, Min: 3},
			},
			exp: []int{20, 10, 10, 10, 10},
		},
		{
			items: []Item{
				{Grow: 1, Shrink: 1, Basis: 0, Size: 20, Min: 20},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 3, Min: 3},
				{Grow: 4, Shrink: 1, Basis: 0, Size: 10, Min: 5},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 3, Min: 3},
				{Grow: 1, Shrink: 1, Basis: 0, Size: 3, Min: 3},
			},
			exp: []int{20, 5, 23, 6, 6},
		},
	} {
		lens := ResolveFlexLengths(tc.items, 60)
		assert.Equal(t, tc.exp, lens)
	}
}

func FuzzResolveFlexLengths1Item(f *testing.F) {
	f.Add(uint16(100), 1, 1, 1, 100, 1, 1000)
	f.Add(uint16(100), 1, 1, -1, 100, 1, 1000)
	f.Fuzz(func(t *testing.T, w uint16, g1, s1, b1, main1, min1, max1 int) {
		items := []Item{
			{Grow: g1, Shrink: s1, Basis: b1, Size: main1, Min: min1, Max: max1},
		}
		for _, it := range items {
			it.Validate()
		}

		widths := ResolveFlexLengths(items, int(w))

		assert.Len(t, widths, len(items), "wrong number of widths")
		for i, w := range widths {
			it := items[i]
			assert.Greater(t, w, 0, "null or negative width")
			if it.Max != 0 {
				assert.LessOrEqual(t, w, it.Max, "max size violation")
			}
			assert.GreaterOrEqual(t, w, it.Min, "min size violation")
		}
	})
}

func FuzzResolveFlexLengths2Items(f *testing.F) {
	f.Add(uint16(100), 1, 1, 1, 100, 1, 1000, 1, 1, 1, 100, 1, 1000)
	f.Add(uint16(100), 1, 1, -1, 100, 1, 1000, 1, 1, -1, 100, 1, 1000)
	f.Fuzz(func(t *testing.T, w uint16, g1, s1, b1, main1, min1, max1 int,
		g2, s2, b2, main2, min2, max2 int) {
		items := []Item{
			{Grow: g1, Shrink: s1, Basis: b1, Size: main1, Min: min1, Max: max1},
			{Grow: g2, Shrink: s2, Basis: b2, Size: main2, Min: min2, Max: max2},
		}
		for _, it := range items {
			it.Validate()
		}

		widths := ResolveFlexLengths(items, int(w))

		assert.Len(t, widths, len(items), "wrong number of widths")
		for i, w := range widths {
			it := items[i]
			assert.Greater(t, w, 0, "null or negative width")
			if it.Max != 0 {
				assert.LessOrEqual(t, w, it.Max, "max size violation")
			}
			assert.GreaterOrEqual(t, w, it.Min, "min size violation")
		}
	})
}

func FuzzResolveFlexLengths3Items(f *testing.F) {
	f.Add(uint16(100),
		1, 1, 1, 100, 1, 1000,
		1, 1, 1, 100, 1, 1000,
		1, 1, 1, 100, 1, 1000,
	)
	f.Add(uint16(100),
		1, 1, -1, 100, 1, 1000,
		1, 1, -1, 100, 1, 1000,
		1, 1, -1, 100, 1, 1000,
	)
	f.Fuzz(func(t *testing.T, w uint16, g1, s1, b1, main1, min1, max1 int,
		g2, s2, b2, main2, min2, max2 int,
		g3, s3, b3, main3, min3, max3 int) {
		items := []Item{
			{Grow: g1, Shrink: s1, Basis: b1, Size: main1, Min: min1, Max: max1},
			{Grow: g2, Shrink: s2, Basis: b2, Size: main2, Min: min2, Max: max2},
			{Grow: g3, Shrink: s3, Basis: b3, Size: main3, Min: min3, Max: max3},
		}
		for _, it := range items {
			it.Validate()
		}

		widths := ResolveFlexLengths(items, int(w))

		assert.Len(t, widths, len(items), "wrong number of widths")
		for i, w := range widths {
			it := items[i]
			assert.Greater(t, w, 0, "null or negative width")
			if it.Max != 0 {
				assert.LessOrEqual(t, w, it.Max, "max size violation")
			}
			assert.GreaterOrEqual(t, w, it.Min, "min size violation")
		}
	})
}
