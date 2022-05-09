package bitmap

import (
	"fmt"
	"math"
)

// Bitmap provides a size-limited continuous binary structure, allowing access
// to individiual bits.
//
// Under the hood it is implemented as a slice of int64 which grows as
// required.
type Bitmap struct {
	// capacity is the number of bits the bitmap allows to access
	capacity int
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
	bm := Bitmap{capacity: capacity, length: size * 64, data: data}

	return &bm
}

// Set sets the bit at a given index to 1.
//
// An error is returned if the index is invalid.
func (bm *Bitmap) Set(bit int) error {
	if bit < 0 || bit >= bm.capacity {
		return fmt.Errorf("Invalid index %d. Must be in range [0, %d]", bit, bm.capacity-1)
	}

	if bit >= bm.length {
		bm.resize(bit + 1)
	}

	idx := bit / 64
	offset := bit % 64
	// 0x8000000000000000 is a single 1, followed by 63 0s
	mask := uint64(0x8000000000000000) >> offset

	bm.data[idx] = bm.data[idx] | mask

	return nil
}

// Unset sets the bit at a given index to 0.
//
// An error is returned if the index is invalid.
func (bm *Bitmap) Unset(bit int) error {
	if bit < 0 || bit >= bm.capacity {
		return fmt.Errorf("Invalid index %d. Must be in range [0, %d]", bit, bm.capacity-1)
	}

	if bit >= bm.length {
		bm.resize(bit + 1)
	}

	idx := bit / 64
	offset := bit % 64

	// Type inference otherwise initializes them as an int, upon which the
	// below will overflow.
	var leftMask, rightMask uint64
	leftMask = math.MaxUint64 << (64 - offset) // first idx bits are 1s, followed by 0s
	rightMask = math.MaxUint64 >> (offset + 1) // last 64 - (idx + 1) bits are 1s, the rest is 0s.
	mask := leftMask | rightMask               // 64 1s, except for one 0 at offset

	bm.data[idx] = bm.data[idx] & mask

	return nil
}

// Get retrieves the value at a given index.
//
// While the returned value is a byte, it will always be either 0 or 1.
//
// An error is returned if the index is invalid.
func (bm *Bitmap) Get(bit int) (byte, error) {
	if bit < 0 || bit >= bm.capacity {
		return 0, fmt.Errorf("Invalid index %d. Must be in range [0, %d]", bit, bm.capacity-1)
	}

	if bit >= bm.length {
		bm.resize(bit + 1)
	}

	idx := bit / 64
	offset := bit % 64
	// 0x8000000000000000 is a single 1, followed by 63 0s
	mask := uint64(0x8000000000000000) >> offset

	// Shifting right will ensure we've got a 0 or a 1, so we can safely
	// convert to byte.
	val := byte(bm.data[idx] & mask >> (64 - offset - 1))

	return val, nil
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

	if bit > bm.capacity {
		bit = bm.capacity
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
