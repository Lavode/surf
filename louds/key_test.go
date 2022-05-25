package louds

import (
	"bytes"
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestLess(t *testing.T) {
	a := Key{0x00, 0x01}
	b := Key{0x00, 0x01, 0x02}
	c := Key{0x00, 0x02}

	// Irreflexivity
	assert.False(t, a.Less(a))

	// If shared prefix differs, first differing byte governs which is
	// lesser
	assert.True(t, b.Less(c))
	assert.False(t, c.Less(b))

	// If shared prefix equal, shorter key is lesser
	assert.True(t, a.Less(b))
	assert.False(t, b.Less(a))

	// Transitivity
	assert.True(t, a.Less(c))
	assert.False(t, c.Less(a))
}

func TestTruncate(t *testing.T) {
	keys := []Key{
		[]byte("f"),
		[]byte("farther"),
		[]byte("fas"),
		[]byte("fasten"),
		[]byte("fat"),
		[]byte("splice"),
		[]byte("topper"),
		[]byte("toy"),
		[]byte("tries"),
		[]byte("tripper"),
		[]byte("trying"),
	}

	expected := []Key{
		[]byte("f"),
		[]byte("far"),
		[]byte("fas"),
		[]byte("fast"),
		[]byte("fat"),
		[]byte("s"),
		[]byte("top"),
		[]byte("toy"),
		[]byte("trie"),
		[]byte("trip"),
		[]byte("try"),
	}

	truncated := Truncate(keys)
	assert.Equal(t, expected, truncated)
}

func TestTruncateRandom(t *testing.T) {
	rand.Seed(42)

	numKeys := 1_000_000
	maxKeyLength := 128

	// Generate set of random keys
	keys := make([]Key, numKeys)
	for i := 0; i < numKeys; i++ {
		key := make([]byte, rand.Intn(maxKeyLength)+1)
		rand.Read(key)
		keys[i] = key
	}

	sort := func(x, y Key) bool {
		return x.Less(y)
	}
	slices.SortFunc(keys, sort)

	log.Printf("Before compaction: %d", len(keys))
	// We can't have any duplicates to start with
	equal := func(x, y Key) bool {
		return bytes.Equal(x, y)
	}
	keys = slices.CompactFunc(keys, equal)
	log.Printf("After compaction: %d", len(keys))

	truncated := Truncate(keys)
	for i := 0; i < len(truncated)-1; i++ {
		// Keys' prefixes must be preserved
		assert.True(t, bytes.HasPrefix(keys[i], truncated[i]))

		// Keys must still be unique
		assert.NotEqual(t, truncated[i], truncated[i+1])
	}
}

func TestFirstDifferenceAt(t *testing.T) {
	a := []byte{0x00, 0x01}
	b := []byte{0x01, 0x00}
	differs, idx := FirstDifferenceAt(a, b)
	assert.True(t, differs)
	assert.Equal(t, 0, idx)

	a = []byte{0x00, 0x01, 0x02, 0x00}
	b = []byte{0x00, 0x01, 0x03, 0x00}
	differs, idx = FirstDifferenceAt(a, b)
	assert.True(t, differs)
	assert.Equal(t, 2, idx)

	a = []byte{0x00, 0x01, 0x02}
	b = []byte{0x00, 0x01, 0x02, 0x03}
	differs, idx = FirstDifferenceAt(a, b)
	assert.True(t, differs)
	assert.Equal(t, 3, idx)

	a = []byte{0x00, 0x01, 0x02, 0x03}
	b = []byte{0x00, 0x01, 0x02}
	differs, idx = FirstDifferenceAt(a, b)
	assert.True(t, differs)
	assert.Equal(t, 3, idx)

	a = []byte{0x00, 0x01, 0x02, 0x03}
	b = []byte{0x00, 0x01, 0x02, 0x03}
	differs, idx = FirstDifferenceAt(a, b)
	assert.False(t, differs)
}
