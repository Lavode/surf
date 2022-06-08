package store

import (
	"errors"
	"fmt"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
	"github.com/Lavode/surf/stack"
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
	// nodes is the stack of node indices we have visited along our path
	nodes stack.Stack[int]

	// nextEdge is the value of the next edge we will (try to) visit in
	// this node, if it exists.
	nextEdge int
	// edges is the stack of edges we passed through to get to the current
	// node.
	edges stack.Stack[int]

	// keyPrefix is the sequence of bytes defining the key leading up to
	// the current node.
	keyPrefix stack.Stack[byte]
}

func NewIterator(labels, hasChild, isPrefixKey *bitmap.Bitmap) Iterator {
	it := Iterator{
		Labels:      labels,
		HasChild:    hasChild,
		IsPrefixKey: isPrefixKey,
	}

	it.keyPrefix = stack.Stack[byte]{}
	it.nodes = stack.Stack[int]{}
	it.edges = stack.Stack[int]{}

	return it
}

// ErrNoSuchEdge indicates that the requested edge does not exist.
var ErrNoSuchEdge = errors.New("Cannot move to non-existant edge")

// ErrIsLeaf indicates that the requested edge points to a leaf node, which
// cannot be travelled to.
var ErrIsLeaf = errors.New("Cannot move to leaf node")

// ErrEndOfTrie indicates that trie traversal reached the end of the trie.
var ErrEndOfTrie = errors.New("Reached end of trie")

// GoToChild attempts to move down the edge with the specified value.
//
// It will keep track of the current node it is at, as well as of the key
// defined by the edge along the path.
//
// If the edge does not exist, ErrNoSuchEdge is returned.
// If the edge points to a leaf node (which cannot be travelled to), ErrIsLeaf
// is returned.
// If an error occurs in the underlying data structure, a generic error is
// returned.
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

	// Update path we took to get here
	it.keyPrefix.Push(edge)
	it.nodes.Push(it.NodeIndex)
	it.edges.Push(int(edge))

	it.NodeIndex = nextNode
	it.nextEdge = 0 // We'll start at the first edge of the new node

	return nil
}

// Next moves to and returns the next key in lexicographic order.
//
// Once the end of the trie is reached, ErrEndOfTrie is returned.
func (it *Iterator) Next() (louds.Key, error) {
	for {
		// We first attempt to go depth-first down the first available edge.
		// TODO This is not a very elegant solution for LOUDS-DENSE, but will
		// do for now.
		for ; it.nextEdge < 256; it.nextEdge++ {
			err := it.GoToChild(byte(it.nextEdge))
			if errors.Is(err, ErrNoSuchEdge) {
				continue
			} else if errors.Is(err, ErrIsLeaf) {
				// While we can't traverse to a leaf, it's certainly a
				// value we can yield.
				key := it.keyPrefix.Data()
				key = append(key, byte(it.nextEdge))
				it.nextEdge++
				return key, nil
			} else if err != nil {
				// Something went awry
				return nil, err
			} else {
				// We actually managed to dive down one level.

				isPrefixKey, err := it.IsPrefixKey.Get(it.NodeIndex)
				if err != nil {
					return nil, err
				}

				if isPrefixKey == 1 {
					// The node we dove to is a prefix key, so is
					// the next key in the sequence.
					key := it.keyPrefix.Data()

					return key, nil
				}

			}
		}

		// We exhausted all possible edges in the current node, so must go up
		// one level.
		if it.NodeIndex == 0 {
			// But if we're at the root node, there's no way to go
			// up, we traversed the whole trie.
			return nil, ErrEndOfTrie
		} else {
			it.NodeIndex = it.nodes.Pop()
			it.nextEdge = it.edges.Pop() + 1 // Don't want to dive down the same edge again
			it.keyPrefix.Pop()
		}
	}
}
