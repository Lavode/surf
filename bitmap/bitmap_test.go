package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	bm := New(128, 256)
	assert.Equal(t, 2, len(bm.data))
	assert.Equal(t, 4, cap(bm.data))
	assert.Equal(t, 256, bm.capacity)

	// This one will trigger rounding up
	bm = New(129, 257)
	assert.Equal(t, 3, len(bm.data))
	assert.Equal(t, 5, cap(bm.data))
	assert.Equal(t, 257, bm.capacity)
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

func TestInvalidIndex(t *testing.T) {
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
	bm := New(0, 1024)

	idxs := []int{0, 10, 100, 500, 1023}
	for _, k := range idxs {
		err := bm.Set(k)
		assert.Nil(t, err)
	}

	for _, k := range idxs {
		val, err := bm.Get(k)
		assert.Nil(t, err)
		assert.Equal(t, byte(1), val)
	}
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
