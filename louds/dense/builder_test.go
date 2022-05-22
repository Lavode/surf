package dense

/*
This set of unit tests is very much implementation-aware, with a tight coupling
to the builder's internals.

This feels a tad inelegant, and produces brittle tests. However I have not managed to
nicely break up the builder's logic into separate components, meaning it is
complicated and keeps track of a lot of state.

These tests thus served as guidelines during the implementation of the builder,
and act as safeguards against the introduction of new bugs if there ever is
some refactoring.
*/

import (
	"testing"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
	"github.com/stretchr/testify/assert"
)

const BUILDER_MEMORY_LIMIT = 80_000_000 // 10 MB

var keys []louds.Key = []louds.Key{
	[]byte("f"),    // 0
	[]byte("far"),  // 1
	[]byte("fas"),  // 2
	[]byte("fast"), // 3
	[]byte("fat"),  // 4
	[]byte("s"),    // 5
	[]byte("top"),  // 6
	[]byte("toy"),  // 7
	[]byte("trie"), // 8
	[]byte("trip"), // 9
	[]byte("try"),  // 10
}

func TestAddEdge(t *testing.T) {
	b := NewBuilder(BUILDER_MEMORY_LIMIT)

	// As our first edge we'll use what is the default value of
	// builder.currentEdge, to make sure this is handled properly.
	// A new edge should:
	assert.Nil(t, b.addEdge(0x00))
	// - Be kept track of
	assert.Equal(t, uint8(0x00), b.currentEdge)
	// - Be added to the bitmap
	val, err := b.Labels.Get(0x00)
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)

	// An existing edge should:
	assert.Nil(t, b.addEdge(0x42))
	// - Not change the current edge
	assert.Equal(t, uint8(0x42), b.currentEdge)
	// - Not affect the bitmap
	val, err = b.Labels.Get(0x42)
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)

	// A second new edge for flavour
	assert.Nil(t, b.addEdge(0xF7))
	// - Be kept track of
	assert.Equal(t, uint8(0xF7), b.currentEdge)
	// - Be added to the bitmap
	val, err = b.Labels.Get(0xF7)
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)
}

func TestBeginNode(t *testing.T) {
	b := NewBuilder(BUILDER_MEMORY_LIMIT)

	// Starting a new node should:
	b.beginNode()
	// - Mark that the current node has no edges defined yet
	assert.Equal(t, false, b.currentNodeHasEdges)
}

func TestBuildOneLevel(t *testing.T) {
	builder := NewBuilder(BUILDER_MEMORY_LIMIT)

	keys := []louds.Key{[]byte{0x00}, []byte{0x17}, []byte{0x42}, []byte{0x60}, []byte{0xF9}}

	expectedLabels := bitmap.New(256, 256)
	expectedHasChild := bitmap.New(256, 256)
	expectedIsPrefixKey := bitmap.New(1, 256)

	for _, k := range keys {
		assert.Nil(t, expectedLabels.Set(int(k[0])))
	}
	assert.Nil(t, expectedIsPrefixKey.Set(0))

	assert.Nil(t, builder.Build(keys))

	assert.True(t, expectedLabels.Equal(builder.Labels), "Expected Labels:\n%s\nGot:\n%s", expectedLabels, builder.Labels)
	assert.True(t, expectedHasChild.Equal(builder.HasChild), "Expected HasChild:\n%s\nGot:\n%s", expectedHasChild, builder.HasChild)
}
