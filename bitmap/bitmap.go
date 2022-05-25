package bitmap

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/Lavode/surf/bitops"
)

// Bitmap provides a size-limited continuous binary structure, allowing access
// to individiual bits.
//
// It further provides methods to count the number of 0s and 1s up to a given
// position, respectively find the position of the i-th 0 and 1.
//
// Under the hood it is implemented as a slice of int64 which grows as
// required.
type Bitmap struct {
	// Capacity is the number of bits the bitmap allows to access
	Capacity int
	// length is the number of bits accessible without a resize
	length int
	data   []uint64
}

// New initializes a new bitmap.
//
// Capacity specifies the maximum size of the bitmap in bits. Thus the
// addressable bits will be in the closed interval [0, capacity - 1].
//
// size specifies the size with which the bitmap will be initialized, in bits.
func New(size, capacity int) *Bitmap {
	// We can fit 64 bits into each uint64.
	dataSize := size / 64
	dataCapacity := capacity / 64

	// But have to round up, in case it's not a multiple of 64
	if size%64 != 0 {
		dataSize++
	}
	if capacity%64 != 0 {
		dataCapacity++
	}

	data := make([]uint64, dataSize, dataCapacity)
	bm := Bitmap{Capacity: capacity, length: dataSize * 64, data: data}

	return &bm
}

// Set sets the bit at a given index to 1.
//
// An error is returned if the index is invalid.
func (bm *Bitmap) Set(bit int) error {
	if bit < 0 || bit >= bm.Capacity {
		return fmt.Errorf("Invalid index %d. Must be in range [0, %d]", bit, bm.Capacity-1)
	}

	if bit >= bm.length {
		bm.resize(bit + 1)
	}

	idx := bit / 64
	offset := bit % 64

	mask := bitops.SingleOneMask(offset)
	bm.data[idx] = bm.data[idx] | mask

	return nil
}

// Unset sets the bit at a given index to 0.
//
// An error is returned if the index is invalid.
func (bm *Bitmap) Unset(bit int) error {
	if bit < 0 || bit >= bm.Capacity {
		return fmt.Errorf("Invalid index %d. Must be in range [0, %d]", bit, bm.Capacity-1)
	}

	if bit >= bm.length {
		bm.resize(bit + 1)
	}

	idx := bit / 64
	offset := bit % 64

	mask := bitops.OnesMask(offset, 64-offset-1) // 64 1s, except for one 0 at offset
	bm.data[idx] = bm.data[idx] & mask

	return nil
}

// Get retrieves the value at a given index.
//
// While the returned value is a byte, it will always be either 0 or 1.
//
// An error is returned if the index is invalid.
func (bm *Bitmap) Get(bit int) (byte, error) {
	if bit < 0 || bit >= bm.Capacity {
		return 0, fmt.Errorf("Invalid index %d. Must be in range [0, %d]", bit, bm.Capacity-1)
	}

	if bit >= bm.length {
		bm.resize(bit + 1)
	}

	idx := bit / 64
	offset := bit % 64
	mask := bitops.SingleOneMask(offset)
	// Shifting right will ensure we've got a 0 or a 1, so we can safely
	// convert to byte.
	val := byte(bm.data[idx] & mask >> (64 - offset - 1))

	return val, nil
}

// Select returns the index of the nth bit value val.
//
// An error is returned if there is no nth bit of value val in the bitmap, or
// if val is neither 0 nor 1.
func (bm *Bitmap) Select(val, nth int) (int, error) {
	if !(val == 0 || val == 1) {
		return 0, fmt.Errorf("Val must be one of 0, 1. Was %d", val)
	}

	if nth <= 0 || nth > bm.length {
		return 0, fmt.Errorf("Nth must be in [0, %d]. Was %d", bm.length, nth)
	}

	checkOnes := val == 1

	count := 0
	var idx int
	for idx = 0; idx < bm.length; idx += 64 {
		var additionalCount int
		if checkOnes {
			additionalCount = bits.OnesCount64(bm.data[idx/64])
		} else {
			additionalCount = 64 - bits.OnesCount64(bm.data[idx/64])
		}

		if count+additionalCount >= nth {
			// We'd overshoot, which would require tricky
			// backtracking. So instead we'll bail out.
			// We also bail out if it's an exact hit, to simplify
			// follow-up code at the cost of checking the uint64
			// twice.
			break
		}

		// Otherwise we proceed on our merry way
		count += additionalCount
	}

	// At this point we either a) ran out of data or b) need to consider
	// the most recent uint64 bit by bit.

	// Case 1: We ran out of data.
	if idx > bm.length {
		return 0, fmt.Errorf("Bitmap only contained %d bits of value %d", count, val)
	}

	// Case 2: We bailed out as we'd have overshot. idx currently points to
	// the first bit of the most-recently-considered uint64.
	// We'll go through the current uint64 bit by bit until we get to the
	// desired count.
	for count < nth {
		val, err := bm.Get(idx)
		if err != nil {
			return 0, err
		}

		if (val == 1 && checkOnes) || (val == 0 && !checkOnes) {
			count += 1
		}

		idx += 1
	}

	// idx is one too high as we increased it one last time in the last
	// iteration of the loop
	return idx - 1, nil
}

// Rank returns the number of bits with value val, up to and including
// position idx.
//
// An error is returned if the index is outside the range of the bitmap, or if
// val is neither 0 nor 1.
func (bm *Bitmap) Rank(val, idx int) (int, error) {
	if idx < 0 || idx > bm.length-1 {
		return 0, fmt.Errorf("Index must be in range [%d, %d]. Was %d]", 0, bm.length-1, idx)
	}

	if !(val == 0 || val == 1) {
		return 0, fmt.Errorf("Val must be one of 0, 1. Was %d", val)
	}
	checkOnes := val == 1

	cnt := 0
	for i := 0; i <= idx; i += 64 {
		var onesCount int

		// We can count ones/zeroes in the full next uint64
		fullBlock := i+63 < idx
		if fullBlock {
			// Here we can consider the full uint64
			onesCount = bits.OnesCount64(bm.data[i/64])
		} else {
			// Wheras here we only care about the first (idx-i)+1 bits
			mask := bitops.LeadingOnesMask(idx - i + 1)
			onesCount = bits.OnesCount64(bm.data[i/64] & mask)
		}

		if checkOnes {
			cnt += onesCount
		} else {
			if fullBlock {
				cnt += 64 - onesCount
			} else {
				// We only checked the first (idx - i + 1) bits
				// for ones.
				cnt += idx - i + 1 - onesCount
			}
		}
	}

	return cnt, nil
}

// String returs a representation of the bitmap's contents as a string of bits.
//
// This means that each stored bit will be encoded as either a 0 or 1 ASCII
// character. Eight bits are grouped together, with 8 bytes per line. The
// output also contains decimal bit offsets per line.
func (bitmap Bitmap) String() string {
	out := ""

	// What a monstrosity :)
	maxOffsetLength := int(math.Ceil(math.Log10(float64(bitmap.length))))
	// Decimal (%d)
	// Padded to maxOffsetLength
	// With 0s rather than spaces (0)
	pattern := fmt.Sprintf("%%0%dd |", maxOffsetLength)

	// As we work with int64s internally, we know that the bitmap's length
	// is guaranteed to be a multiple of 64.
	for i := 0; i < len(bitmap.data); i++ {
		// We want the bit as address
		out += fmt.Sprintf(pattern, i*64)

		// We'll print eight bytes at a time
		bitsString := fmt.Sprintf("%064b", bitmap.data[i])

		// But we want a space every eight digits
		for j := 0; j < 64; j += 8 {
			b := bitsString[j : j+8]
			out += fmt.Sprintf(" %s", b)
		}

		out += "\n"
	}

	return out
}

// Equal checks whether the length and content of two bitmaps is equal.
//
// It does not compare their capacities, however.
func (bm *Bitmap) Equal(other *Bitmap) bool {
	if bm.length != other.length {
		return false
	}

	for i := 0; i < bm.length/64; i++ {
		if bm.data[i] != other.data[i] {
			return false
		}
	}

	return true
}

// resize will increase the bitmap's internal memory such that it can accomodate
// a given number of bits.
//
// bit specifies the number of bits which are guaranteed to be accessible after
// the resize. That is the maximum index guaranteed to be accessible will be
// bit - 1.
func (bm *Bitmap) resize(bit int) {
	if bit < bm.length {
		return
	}

	if bit > bm.Capacity {
		bit = bm.Capacity
	}

	// Amount of int64s we need in total to accomodate new amount of bits
	newLength := bit / 64
	if bit%64 != 0 {
		newLength++
	}

	additionalUints := newLength - len(bm.data)
	for i := 0; i < additionalUints; i++ {
		bm.data = append(bm.data, uint64(0))
	}

	bm.length = newLength * 64
}
