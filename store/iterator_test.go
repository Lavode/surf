package store

import (
	"testing"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
	"github.com/Lavode/surf/louds/dense"
	"github.com/stretchr/testify/assert"
)

// Initialize LOUDS-DENSE encoded FST with toy-sized dataset from paper. This
// relies on LOUDS-DENSE builder being functional and correct, but seems more
// sane than maintaining the bitmaps by hand.
func buildFST(t *testing.T) (labels, hasChild, isPrefixKey *bitmap.Bitmap) {
	// These are already truncated and sorted, so can be fed to builder
	// directly.
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

	builder := dense.NewBuilder(1_000_000) // 1 MiB memory limit is plenty
	assert.Nil(t, builder.Build(keys))

	labels = builder.Labels
	hasChild = builder.HasChild
	isPrefixKey = builder.IsPrefixKey

	return
}

func TestGoToChild(t *testing.T) {
	labels, hasChild, isPrefixKey := buildFST(t)
	it := Iterator{
		Labels:      labels,
		HasChild:    hasChild,
		IsPrefixKey: isPrefixKey,
	}

	// root -> f child
	it.NodeIndex = 0
	assert.Nil(t, it.GoToChild('f'))
	assert.Equal(t, 1, it.NodeIndex)

	// root -> h non-existant
	it.NodeIndex = 0
	assert.ErrorIs(t, it.GoToChild('h'), ErrNoSuchEdge)

	// root -> s leaf
	it.NodeIndex = 0
	assert.ErrorIs(t, it.GoToChild('s'), ErrIsLeaf)

	// root -> t node
	it.NodeIndex = 0
	assert.Nil(t, it.GoToChild('t'))
	assert.Equal(t, 2, it.NodeIndex)

	// t -> r node
	assert.Nil(t, it.GoToChild('r'))
	assert.Equal(t, 5, it.NodeIndex)

	// r -> i
	assert.Nil(t, it.GoToChild('i'))
	assert.Equal(t, 7, it.NodeIndex)
}

func TestNext(t *testing.T) {
	labels, hasChild, isPrefixKey := buildFST(t)
	it := Iterator{
		Labels:      labels,
		HasChild:    hasChild,
		IsPrefixKey: isPrefixKey,
	}

	// We'll iterate through keys until we hit the end of the tree,
	// expecting them to be yielded in lexicographic order.
	key, err := it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("f"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("far"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("fas"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("fast"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("fat"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("s"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("top"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("toy"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("trie"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("trip"), key)

	key, err = it.Next()
	assert.Nil(t, err)
	assert.Equal(t, louds.Key("try"), key)

	_, err = it.Next()
	assert.ErrorIs(t, err, ErrEndOfTrie)
}

func TestNextAndGotoChild(t *testing.T) {
	labels, hasChild, isPrefixKey := buildFST(t)
	it := Iterator{
		Labels:      labels,
		HasChild:    hasChild,
		IsPrefixKey: isPrefixKey,
	}

	// This test aims to ensure that various combinations of "go to child"
	// and "next" work as they should. This is worth testing as both
	// maintain and modify the iterator's internal state.

	err := it.GoToChild('f')
	// Now at 'f'
	assert.Nil(t, err)

	key, err := it.Next()
	// Now at 'fa' with key 'far'
	assert.Nil(t, err)
	assert.Equal(t, louds.Key{'f', 'a', 'r'}, key)

	err = it.GoToChild('s')
	// Now at 'fas'
	assert.Nil(t, err)

	key, err = it.Next()
	// Now at 'fas' with key 'fast'
	assert.Equal(t, louds.Key{'f', 'a', 's', 't'}, key)

	key, err = it.Next()
	// Now at 'fa' with key 'fat'
	assert.Equal(t, louds.Key{'f', 'a', 't'}, key)

	key, err = it.Next()
	// Now at '' with key 't'
	assert.Equal(t, louds.Key{'s'}, key)

	err = it.GoToChild('t')
	// Now at 't'
	assert.Nil(t, err)

	err = it.GoToChild('o')
	// Now at 'to'
	assert.Nil(t, err)

	key, err = it.Next()
	// Now at 'to' with key 'top'
	assert.Nil(t, err)
	assert.Equal(t, louds.Key{'t', 'o', 'p'}, key)

	key, err = it.Next()
	// Now at 'to' with key 'toy'
	assert.Nil(t, err)
	assert.Equal(t, louds.Key{'t', 'o', 'y'}, key)

	key, err = it.Next()
	// Now at 'tri' with key 'trie'
	assert.Nil(t, err)
	assert.Equal(t, louds.Key{'t', 'r', 'i', 'e'}, key)

	// Which has a leaf node 'trip'
	err = it.GoToChild('p')
	assert.ErrorIs(t, err, ErrIsLeaf)
}
