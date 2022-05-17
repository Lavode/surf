package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	bm := New(128, 256)
	assert.Equal(t, 2, len(bm.data))
	assert.Equal(t, 4, cap(bm.data))
	assert.Equal(t, 256, bm.Capacity)
	assert.Equal(t, 128, bm.length)

	// This one will trigger rounding up
	bm = New(129, 257)
	assert.Equal(t, 3, len(bm.data))
	assert.Equal(t, 5, cap(bm.data))
	assert.Equal(t, 257, bm.Capacity)
	assert.Equal(t, 192, bm.length)
}

func TestSet(t *testing.T) {
	// A brief implementation-aware test, to ensure we actually set bits
	// the way we think we do.
	bm := New(64, 64)

	keys := []int{0, 7, 16, 63}
	for _, k := range keys {
		err := bm.Set(k)
		assert.Nil(t, err)
	}

	// Thus we expect bits 0, 7, 16 and 63 to be set, counting from the left.
	assert.Equal(t, uint64(0x8100800000000001), bm.data[0])
}

func TestSetGetAndUnset(t *testing.T) {
	bm := New(256, 256)

	for i := 0; i < 256; i++ {
		val, err := bm.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, byte(0), val)

		err = bm.Set(i)
		assert.Nil(t, err)

		val, err = bm.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, byte(1), val)
	}

	for i := 0; i < 256; i++ {
		err := bm.Unset(i)
		assert.Nil(t, err)

		val, err := bm.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, byte(0), val)
	}
}

func TestAccessInvalidIndex(t *testing.T) {
	bm := New(200, 300)

	invalidIdxs := []int{-17, -1, 300, 1024}
	for _, i := range invalidIdxs {
		_, err := bm.Get(i)
		assert.Error(t, err)

		err = bm.Set(i)
		assert.Error(t, err)

		err = bm.Unset(i)
		assert.Error(t, err)
	}
}

func TestResize(t *testing.T) {
	idxs := []int{0, 10, 100, 500, 1023}

	// Ensure Set resizes
	bm := New(0, 1024)
	for _, k := range idxs {
		err := bm.Set(k)
		assert.Nil(t, err)
	}

	// Ensure Unset resizes
	bm = New(0, 1024)
	for _, k := range idxs {
		err := bm.Unset(k)
		assert.Nil(t, err)
	}

	// Ensure Get resizes
	bm = New(0, 1024)
	for _, k := range idxs {
		_, err := bm.Get(k)
		assert.Nil(t, err)
	}
}

func TestRank(t *testing.T) {
	bm := New(128, 128)
	bm.data = []uint64{
		0x84d5f768e45d7022,
		0x1feeb05a21aeb691,
	}
	// 0b 10000100 11010101 11110111 01101000
	//    11100100 01011101 01110000 00100010
	//    00011111 11101110 10110000 01011010
	//    00100001 10101110 10110110 10010001

	tests := []struct {
		idx    int
		ones   int
		zeroes int
	}{
		{0, 1, 0},
		{7, 2, 6},
		{17, 9, 9},
		{31, 17, 15},
		{55, 29, 27},
		{64, 31, 34},
		{120, 62, 59},
		{127, 64, 64},
	}

	for _, test := range tests {
		rankZero, err := bm.Rank(0, test.idx)
		assert.Nil(t, err)
		assert.Equal(t, test.zeroes, rankZero)

		rankOne, err := bm.Rank(1, test.idx)
		assert.Nil(t, err)
		assert.Equal(t, test.ones, rankOne)
	}
}

func TestRankInvalidArguments(t *testing.T) {
	bm := New(128, 128)

	// Invalid lookup value
	_, err := bm.Rank(3, 27)
	assert.Error(t, err)

	// Invalid indices
	_, err = bm.Rank(0, -3)
	assert.Error(t, err)

	_, err = bm.Rank(0, 128)
	assert.Error(t, err)
}

func TestSelect(t *testing.T) {
	bm := New(128, 128)
	bm.data = []uint64{
		0x84d5f768e45d7022,
		0x1feeb05a21aeb691,
	}
	// 0b 10000100 11010101 11110111 01101000
	//    11100100 01011101 01110000 00100010
	//    00011111 11101110 10110000 01011010
	//    00100001 10101110 10110110 10010001

	tests := []struct {
		n       int
		oneIdx  int
		zeroIdx int
	}{
		{7, 15, 10},
		{17, 28, 36},
		{31, 62, 60},
		{55, 109, 107},
		{64, 127, 126},
	}

	for _, test := range tests {
		selectZero, err := bm.Select(0, test.n)
		assert.Nil(t, err)
		assert.Equal(t, test.zeroIdx, selectZero)

		selectOne, err := bm.Select(1, test.n)
		assert.Nil(t, err)
		assert.Equal(t, test.oneIdx, selectOne)
	}
}

func TestSelectInvalidArguments(t *testing.T) {
	bm := New(128, 128)

	// Invalid lookup value
	_, err := bm.Select(3, 27)
	assert.Error(t, err)

	// Invalid n-th value
	_, err = bm.Select(0, -3)
	assert.Error(t, err)

	_, err = bm.Select(0, 0)
	assert.Error(t, err)

	_, err = bm.Select(0, 129)
	assert.Error(t, err)
}

func TestString(t *testing.T) {
	bm := New(256, 512)

	bits := []int{0, 3, 17, 20, 45, 62, 63, 101, 117, 156, 184, 212, 255}
	for _, b := range bits {
		assert.Nil(t, bm.Set(b))
	}

	str := bm.String()
	expected :=
		`000 | 10010000 00000000 01001000 00000000 00000000 00000100 00000000 00000011
064 | 00000000 00000000 00000000 00000000 00000100 00000000 00000100 00000000
128 | 00000000 00000000 00000000 00001000 00000000 00000000 00000000 10000000
192 | 00000000 00000000 00001000 00000000 00000000 00000000 00000000 00000001
`
	assert.Equal(t, expected, str)
}

func BenchmarkSet(b *testing.B) {
	bm := New(0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := bm.Set(i)
		if err != nil {
			// The error check might be a few splits of a
			// nanosecond, but we can probably live with that.
			b.Errorf("Error while setting bit %d: %v", i, err)
		}
	}
}
