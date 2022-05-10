// bitops provides various bitwise utilities on uint64s which are not supported
// by math.bits.
package bitops

import "math"

// FirstBits returns a uint64 consisting of the first n bits of b, filled up with
// 0s.
func FirstBits(n int, b uint64) uint64 {
	return b & LeadingOnesMask(n)
}

// LastBits returns a uint64 consisting of the last n bits of b, filled up with 0s.
func LastBits(n int, b uint64) uint64 {
	return b & TrailingOnesMask(n)
}

// LeadingOnesMask returns a bitmask where the first n bits are 1, the others
// are 0.
//
// If n lies outside the valid range of [0, 64], it will be coerced to the
// nearest valid value.
func LeadingOnesMask(n int) uint64 {
	if n > 64 {
		n = 64
	} else if n < 0 {
		n = 0
	}

	return math.MaxUint64 << (64 - n)
}

// TrailingOnesMask returns a bitmask where the last n bits are 1, the others
// are 0.
//
// If n lies outside the valid range of [0, 64], it will be coerced to the
// nearest valid value.
func TrailingOnesMask(n int) uint64 {
	if n > 64 {
		n = 64
	} else if n < 0 {
		n = 0
	}

	return math.MaxUint64 >> (64 - n)
}

// OnesMask returns a bitmask where the first leading and last trailing bits
// are 1, while the bits between are 0.
//
// If either of leading or trailing lie outside the valid range of [0, 64],
// they will be coerced to the nearest valid value.
func OnesMask(leading, trailing int) uint64 {
	if leading < 0 {
		leading = 0
	} else if leading > 64 {
		leading = 64
	}

	if trailing < 0 {
		trailing = 0
	} else if trailing > 64 {
		trailing = 64
	}

	left := LeadingOnesMask(leading)
	right := TrailingOnesMask(trailing)

	return left | right
}

// SingleOneMask returns a bitmask where the single bit at position idx is 1,
// while all others are 0.
//
// If idx is outside the valid range of [0, 63] it is coerced to the nearest
// valid value.
func SingleOneMask(idx int) uint64 {
	if idx < 0 {
		idx = 0
	} else if idx > 63 {
		idx = 63
	}

	return 0x8000000000000000 >> idx
}
