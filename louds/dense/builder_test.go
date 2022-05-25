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

	labels := []int{
		// First node: Edges a, f
		0*256 + 'a',
		0*256 + 'f',
		// Second node: Edges i, o
		1*256 + 'i',
		1*256 + 'o',
		// Third node: Edges a, e
		2*256 + 'a',
		2*256 + 'e',
	}
	for _, bit := range labels {
		assert.Nil(t, expectedLabels.Set(bit))
	}

	children := []int{
		// First node: a, f have sub-tree
		0*256 + 'a',
		0*256 + 'f',
	}
	for _, bit := range children {
		assert.Nil(t, expectedHasChild.Set(bit))
	}

	prefixKeys := []int{
		// Third node (-> 'f'): Is prefix key
		2,
	}
	for _, bit := range prefixKeys {
		assert.Nil(t, expectedIsPrefixKey.Set(bit))
	}

	// Let's test it :)
	assert.Nil(t, builder.Build(keys))

	assert.True(t, expectedLabels.Equal(builder.Labels), "Expected Labels:\n%s\nGot:\n%s", expectedLabels, builder.Labels)
	assert.True(t, expectedHasChild.Equal(builder.HasChild), "Expected HasChild:\n%s\nGot:\n%s", expectedHasChild, builder.HasChild)
	assert.True(t, expectedIsPrefixKey.Equal(builder.IsPrefixKey), "Expected IsPrefixKey:\n%s\nGot:\n%s", expectedIsPrefixKey, builder.IsPrefixKey)
}

func TestBuildMultiLevel(t *testing.T) {
	builder := NewBuilder(BUILDER_MEMORY_LIMIT)

	// These keys were chosen such as not to cause any kind of truncation.
	// (Respectively they /are/ truncated already)
	keys := []louds.Key{
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

	expectedLabels := bitmap.New(8*256, 8*256)
	expectedHasChild := bitmap.New(8*256, 8*256)
	expectedIsPrefixKey := bitmap.New(8, 256)

	labels := []int{
		// First node: Edges f, s, t
		0*256 + 'f',
		0*256 + 's',
		0*256 + 't',
		// Second node: Edges a
		1*256 + 'a',
		// Third node: Edges o, r
		2*256 + 'o',
		2*256 + 'r',
		// Fourth node: Edges r, s, t
		3*256 + 'r',
		3*256 + 's',
		3*256 + 't',
		// Fifth node: Edges p, y
		4*256 + 'p',
		4*256 + 'y',
		// Sixth node: Edges i, y
		5*256 + 'i',
		5*256 + 'y',
		// Seventh node: Edges t
		6*256 + 't',
		// Eight node: Edges e, p
		7*256 + 'e',
		7*256 + 'p',
	}
	for _, bit := range labels {
		assert.Nil(t, expectedLabels.Set(bit))
	}

	children := []int{
		// First node: f, t have sub-tree
		0*256 + 'f',
		0*256 + 't',
		// Second node: a has sub-tree
		1*256 + 'a',
		// Third node: o, r have sub-tree
		2*256 + 'o',
		2*256 + 'r',
		// Fourth node: s has sub-tree
		3*256 + 's',
		// Fifth node: No sub-trees
		// Sixth node: i has sub-tree
		5*256 + 'i',
		// Seventh node: No sub-trees
		// Eight node: No sub-trees
	}
	for _, bit := range children {
		assert.Nil(t, expectedHasChild.Set(bit))
	}

	prefixKeys := []int{
		// Second node (-> 'f'): Is prefix key
		1,
		// Seventh node (-> 'fas'): Is prefix key
		6,
	}
	for _, bit := range prefixKeys {
		assert.Nil(t, expectedIsPrefixKey.Set(bit))
	}

	// Let's test it :)
	assert.Nil(t, builder.Build(keys))

	assert.True(t, expectedLabels.Equal(builder.Labels), "Expected Labels:\n%s\nGot:\n%s", expectedLabels, builder.Labels)
	assert.True(t, expectedHasChild.Equal(builder.HasChild), "Expected HasChild:\n%s\nGot:\n%s", expectedHasChild, builder.HasChild)
	assert.True(t, expectedIsPrefixKey.Equal(builder.IsPrefixKey), "Expected IsPrefixKey:\n%s\nGot:\n%s", expectedIsPrefixKey, builder.IsPrefixKey)

	_ = keys
}
