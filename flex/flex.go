// Package flex implements a simplified flexbox layout algorithm.
//
// It tries to stick to the flexbox algorithm and namings as specified:
// https://www.w3.org/TR/css-flexbox-1/#layout-algorithm
// but is geared towards terminal output and thus takes many shortcuts and makes
// some assumptions:
//   - only horizontal layout, and thus "width" and "size" are the same thing
//   - only integer sizes, corresponding to terminal columns
//   - no infinite container size (container size is the width of the terminal
//     or a fallback size, and in both cases non-zero)
//   - no wrapping (in the sense of flex-wrap) of the items on multiple lines
//   - items have no padding so inner width is the same as outer width (any
//     padding between items is subtracted from the container size prior to
//     running the algorithm)
package flex

type Item struct {
	Basis  int // -1 = auto
	Grow   int
	Shrink int

	// "natural" size, e.g. size of the content, or a desired fixed size
	Size int
	// minimum size for the item, e.g. the minimum size the content can wrap to
	// (e.g. length of longest word); must be > 0
	Min int
	// maximum size for the item; if <=0, no maximum is enforced
	Max int

	flexBaseSize   int
	hypoMainSize   int
	frozen         bool
	targetMainSize int
}

func (it *Item) Validate() {
	if it.Min < 1 {
		it.Min = 1
	}
	if it.Grow < 0 {
		it.Grow = 0
	}
	if it.Shrink < 0 {
		it.Shrink = 0
	}
	if it.Size < 1 {
		it.Size = 1
	}
	if it.Max < 0 {
		it.Max = 0
	}
	if it.Size < it.Min {
		it.Size = it.Min
	}
	if it.Max != 0 {
		if it.Max < it.Min {
			it.Max = it.Min
		}
		if it.Size > it.Max {
			it.Size = it.Max
		}
	}
}

func (it *Item) determineFlexBaseSize() {
	if it.Basis >= 0 {
		it.flexBaseSize = it.Basis
	} else {
		it.flexBaseSize = it.Size
	}
}

func (it *Item) determineHypoMainSize() {
	it.hypoMainSize = it.flexBaseSize
	if it.hypoMainSize < it.Min {
		it.hypoMainSize = it.Min
	}
	if it.Max != 0 && it.hypoMainSize > it.Max {
		it.hypoMainSize = it.Max
	}
}

func (it *Item) sizeIfInflexible(useGrow bool) {
	if useGrow {
		if it.Grow == 0 || it.flexBaseSize > it.hypoMainSize {
			it.targetMainSize = it.hypoMainSize
			it.frozen = true
		}
	} else {
		if it.Shrink == 0 || it.flexBaseSize < it.hypoMainSize {
			it.targetMainSize = it.hypoMainSize
			it.frozen = true
		}
	}
}

func ResolveFlexLengths(items []Item, containerSize int) []int {
	var mutItems []*Item
	for i := range items {
		mutItems = append(mutItems, &items[i])
	}

	if containerSize < 0 {
		containerSize = 0
	}
	var sumHypoMainSize int
	for _, it := range mutItems {
		it.Validate()

		// spec 9.2 / 3.
		it.determineFlexBaseSize()
		it.determineHypoMainSize()

		// spec 9.7 / 1.
		sumHypoMainSize += it.hypoMainSize
	}

	// Not in algorithm. But required? or at least a sensible optimization?
	if sumHypoMainSize == containerSize {
		lens := make([]int, len(mutItems))
		for i, it := range mutItems {
			lens[i] = it.hypoMainSize
		}
		return lens
	}

	// still spec 9.7 / 1.
	// true if using useGrow factor, false if using shrink factor
	useGrow := sumHypoMainSize < containerSize

	// spec 9.7 / 2.
	for _, it := range mutItems {
		it.sizeIfInflexible(useGrow)
	}

	// spec 9.7 / 3.
	initialFreeSpace := containerSize
	for _, it := range mutItems {
		if it.frozen {
			initialFreeSpace -= it.targetMainSize
		} else {
			initialFreeSpace -= it.flexBaseSize
		}
	}

	var iterations int
	for {
		iterations++
		if iterations > len(mutItems)+1 {
			// avoid infinite looping at all cost, but shouldn't happen
			panic("no solution found")
		}

		// spec 9.7 / 4.a
		if allFrozen(mutItems) {
			break
		}

		// spec 9.7 / 4.b
		remainingFreeSpace := containerSize
		for _, it := range mutItems {
			if it.frozen {
				remainingFreeSpace -= it.targetMainSize
			} else {
				remainingFreeSpace -= it.flexBaseSize
			}
		}

		// spec 9.7 / 4.c
		if remainingFreeSpace != 0 {
			if useGrow {
				sumGF := sumUnfrozenGrowFactors(mutItems)
				for _, it := range mutItems {
					if it.frozen {
						continue
					}

					growSpace := remainingFreeSpace * it.Grow / sumGF
					it.targetMainSize = it.flexBaseSize + growSpace
					remainingFreeSpace -= growSpace
					sumGF -= it.Grow
				}
			} else {
				scaledShrinks := make([]int, len(mutItems))
				var sumScaledShrinks int
				for i, it := range mutItems {
					if it.frozen {
						continue
					}
					scaledShrinks[i] = it.Shrink * it.flexBaseSize
					sumScaledShrinks += scaledShrinks[i]
				}

				for i, it := range mutItems {
					if it.frozen {
						continue
					}

					shrinkSpace := remainingFreeSpace * scaledShrinks[i] / sumScaledShrinks
					// remember shrinkSpace is negative so when the spec says
					// remove the absolute value, we instead add the value
					it.targetMainSize = it.flexBaseSize + shrinkSpace
					remainingFreeSpace -= shrinkSpace
					sumScaledShrinks -= scaledShrinks[i]
				}
			}
		}

		// spec 9.7 / 4.d

		// violation = clamped-unclamped, so will be >0 for min violations, <0 for max violations
		violations := make([]int, len(mutItems))
		var totalViolation int
		for i, it := range mutItems {
			if it.frozen {
				continue
			}
			if it.targetMainSize < it.Min {
				violations[i] = it.Min - it.targetMainSize
				it.targetMainSize = it.Min
				totalViolation += violations[i]
				continue
			}
			if it.Max != 0 && it.targetMainSize > it.Max {
				violations[i] = it.Max - it.targetMainSize
				it.targetMainSize = it.Max
				totalViolation += violations[i]
			}
		}

		if totalViolation == 0 {
			for _, it := range mutItems {
				it.frozen = true
			}
		} else if totalViolation > 0 {
			for i, it := range mutItems {
				if violations[i] > 0 {
					it.frozen = true
				}
			}
		} else /* totalViolation < 0 */ {
			for i, it := range mutItems {
				if violations[i] < 0 {
					it.frozen = true
				}
			}
		}
	}

	lens := make([]int, len(mutItems))
	for i, it := range mutItems {
		lens[i] = it.targetMainSize
	}
	return lens
}

func allFrozen(items []*Item) bool {
	for _, it := range items {
		if !it.frozen {
			return false
		}
	}
	return true
}

func sumUnfrozenGrowFactors(items []*Item) int {
	var sum int
	for _, it := range items {
		if !it.frozen {
			sum += it.Grow
		}
	}
	return sum
}
