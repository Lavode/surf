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

func TestBuildOneLevel(t *testing.T) {
	builder := NewBuilder(BUILDER_MEMORY_LIMIT)

	keys := []louds.Key{[]byte{0x00}, []byte{0x17}, []byte{0x42}, []byte{0x60}, []byte{0xF9}}

	expectedLabels := bitmap.New(256, 256)
	expectedHasChild := bitmap.New(256, 256)
	expectedIsPrefixKey := bitmap.New(1, 256)

	for _, k := range keys {
		assert.Nil(t, expectedLabels.Set(int(k[0])))
	}

	assert.Nil(t, builder.Build(keys))

	assert.True(t, expectedLabels.Equal(builder.Labels), "Expected Labels:\n%s\nGot:\n%s", expectedLabels, builder.Labels)
	assert.True(t, expectedHasChild.Equal(builder.HasChild), "Expected HasChild:\n%s\nGot:\n%s", expectedHasChild, builder.HasChild)
	assert.True(t, expectedIsPrefixKey.Equal(builder.IsPrefixKey), "Expected IsPrefixKey:\n%s\nGot:\n%s", expectedIsPrefixKey, builder.IsPrefixKey)
}

func TestBuildTwoLevels(t *testing.T) {
	builder := NewBuilder(BUILDER_MEMORY_LIMIT)
	keys := []louds.Key{[]byte("ai"), []byte("ao"), []byte("f"), []byte("fa"), []byte("fe")}

	expectedLabels := bitmap.New(768, 768)
	expectedHasChild := bitmap.New(768, 768)
	expectedIsPrefixKey := bitmap.New(3, 256)

	// First node: Edges f, a
	k := 0
	expectedLabels.Set(k*256 + 102) // f
	expectedLabels.Set(k*256 + 97)  // a
	// And both have FST sub-trees
	expectedHasChild.Set(k*256 + 102) // f
	expectedHasChild.Set(k*256 + 97)  // a

	// Second node: Edges i, o
	k++
	expectedLabels.Set(k*256 + 105) // i
	expectedLabels.Set(k*256 + 111) // o

	// Third node: Edges a, e
	k++
	expectedLabels.Set(k*256 + 97)  // a
	expectedLabels.Set(k*256 + 101) // e
	// And 'f' is prefix key
	expectedIsPrefixKey.Set(k)

	// Let's test it :)
	assert.Nil(t, builder.Build(keys))

	assert.True(t, expectedLabels.Equal(builder.Labels), "Expected Labels:\n%s\nGot:\n%s", expectedLabels, builder.Labels)
	assert.True(t, expectedHasChild.Equal(builder.HasChild), "Expected HasChild:\n%s\nGot:\n%s", expectedHasChild, builder.HasChild)
	assert.True(t, expectedIsPrefixKey.Equal(builder.IsPrefixKey), "Expected IsPrefixKey:\n%s\nGot:\n%s", expectedIsPrefixKey, builder.IsPrefixKey)

}
