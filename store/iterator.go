package store

import (
	"errors"
	"fmt"

	"github.com/Lavode/surf/bitmap"
)

// Iterator implements an iterator through a LOUDS-DENSE encoded FST.
//
// It allows to navigate up, down, and along the tree.
type Iterator struct {
	// Labels is the D-Labels bitmap of the LOUDS-DENSE encoding.
	Labels *bitmap.Bitmap
	// HasChild is the D-HasChild bitmap of the LOUDS-DENSE encoding.
	HasChild *bitmap.Bitmap
	// IsPrefixKey is the D-IsPrefixKey bitmap of the LOUDS-DENSE encoding.
	IsPrefixKey *bitmap.Bitmap

	// NodeIndex is the level-order index of the node the iterator
	// currently points to.
	// The node with index 0 is the root node.
	NodeIndex int
	// EdgeIndex is the index of the edge of the node the iterator
	// currently points to.
	EdgeIndex byte
}

var ErrNoSuchEdge = errors.New("Cannot move to non-existant edge")
var ErrIsLeaf = errors.New("Cannot move to leaf node")

func (it *Iterator) GoToChild(edge byte) error {
	offset := 256*it.NodeIndex + int(edge)

	hasLabel, err := it.Labels.Get(offset)
	if err != nil {
		return fmt.Errorf("Error accessing bit %d of Labels: %v", offset, err)
	}

	if hasLabel != 1 {
		return fmt.Errorf("%w: %d", ErrNoSuchEdge, edge)
	}

	hasChild, err := it.HasChild.Get(offset)
	if err != nil {
		return fmt.Errorf("Error accessing bit %d of HasChild: %v", offset, err)
	}

	if hasChild != 1 {
		return fmt.Errorf("%w: %d", ErrIsLeaf, edge)
	}

	// Index of node this edge points to is given by:
	// rank_1(D-HasChild, offset)
	nextNode, err := it.HasChild.Rank(1, offset)
	if err != nil {
		return fmt.Errorf("Error calculating rank_1(%d) over HasChild: %v", offset, err)
	}

	it.NodeIndex = nextNode
	it.EdgeIndex = 0 // We'll start at the first edge of the new node

	return nil
}
