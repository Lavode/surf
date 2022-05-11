package louds

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
