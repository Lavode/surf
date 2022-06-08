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
	it.EdgeIndex = 0
	assert.Nil(t, it.GoToChild('f'))
	assert.Equal(t, 1, it.NodeIndex)
	assert.Equal(t, byte(0), it.EdgeIndex)

	// root -> h non-existant
	it.NodeIndex = 0
	it.EdgeIndex = 0
	assert.ErrorIs(t, it.GoToChild('h'), ErrNoSuchEdge)

	// root -> s leaf
	it.NodeIndex = 0
	it.EdgeIndex = 0
	assert.ErrorIs(t, it.GoToChild('s'), ErrIsLeaf)

	// root -> t node
	it.NodeIndex = 0
	it.EdgeIndex = 0
	assert.Nil(t, it.GoToChild('t'))
	assert.Equal(t, 2, it.NodeIndex)
	assert.Equal(t, byte(0), it.EdgeIndex)

	// t -> r node
	assert.Nil(t, it.GoToChild('r'))
	assert.Equal(t, 5, it.NodeIndex)
	assert.Equal(t, byte(0), it.EdgeIndex)

	// r -> i
	assert.Nil(t, it.GoToChild('i'))
	assert.Equal(t, 7, it.NodeIndex)
	assert.Equal(t, byte(0), it.EdgeIndex)
}
