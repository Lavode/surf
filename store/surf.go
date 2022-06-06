package store

import (
	"fmt"
	"log"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
	"github.com/Lavode/surf/louds/dense"
)

type SURF struct {
	// R defines the ratio between sparse and dense LOUDS encodings.
	// Check SURFOption for details.
	R uint

	// HashBits defines the number of additional bits to use for storing a
	// partial hash value of keys.
	// Check SURFOption for details.
	HashBits uint

	// RealBits defines the number of additional bits to use for storing
	// parts of the key.
	// Check SURFOption for details.
	RealBits uint

	// LOUDS-DENSE encoding

	// DenseLabels is the D-Labels bitmap of the LOUDS-DENSE encoding.
	// Bits are set corresponding to the outbound edges of a node.
	DenseLabels *bitmap.Bitmap
	// DenseHasChild is the D-HasChild bitmap of the LOUDS-DENSE encoding.
	// Bits are set if the thing pointed to by the edge is an FST
	// sub-component.
	DenseHasChild *bitmap.Bitmap
	// DenseIsPrefixKey is the D-IsPrefixKey bitmap of the LOUDS-DENSE
	// encoding.
	// Bits are set if the prefix leading up to a node is also a stored
	// key.
	DenseIsPrefixKey *bitmap.Bitmap
}

func New(rawKeys [][]byte, options SURFOptions) (*SURF, error) {
	surf := SURF{}

	options.setDefaults()

	surf.R = *options.R
	surf.HashBits = *options.HashBits
	surf.RealBits = *options.RealBits

	// TODO this can't be the proper way, surely :)
	keys := make([]louds.Key, len(rawKeys))
	for i := 0; i < len(rawKeys); i++ {
		keys[i] = louds.Key(rawKeys[i])
	}

	keys = louds.Truncate(keys)

	// TODO once LOUDS-SPARSE support added, memory limit must be split
	// appropriately (based on options.R) between DENSE and SPARSE builder.
	denseBuilder := dense.NewBuilder(*options.MemoryLimit)
	err := denseBuilder.Build(keys)
	if err != nil {
		return nil, fmt.Errorf("Error building LOUDS-DENSE representation: %v", err)
	}

	surf.DenseLabels = denseBuilder.Labels
	surf.DenseHasChild = denseBuilder.HasChild
	surf.DenseIsPrefixKey = denseBuilder.IsPrefixKey

	return &surf, nil
}

// Lookup checks existence of a key in the SuRF store.
//
// While there can be no false negatives, there is the possibility of a false
// positive. The chance of a false positive depends on the distribution of
// keys, as well as the number of additional bits used to store full keys
// (RealBits) and the hash value of keys (KeyBits).
func (surf *SURF) Lookup(key []byte) (bool, error) {
	// The level-order index of the node we are currently in.
	// As such, currentNode is also the offset in D-IsPrefixKey, and
	// currentNode * 256 is the offset in D-Labels and D-HasChild.
	currentNode := 0

	for i := 0; i < len(key); i++ {
		keyByte := key[i]
		labelOffset := currentNode*256 + int(keyByte)

		log.Printf("Currently checking node %d, looking for edge %x (offset = %d)", currentNode, keyByte, labelOffset)

		hasLabel, err := surf.DenseLabels.Get(labelOffset)
		if err != nil {
			return false, fmt.Errorf("Error accessing bit %d in D-Labels: %v", labelOffset, err)
		}

		if hasLabel == 0 {
			// There's no outbound edge with the value we are
			// looking for, so the key doesn't exist.
			log.Printf("Edge not found, key does not exist")
			return false, nil
		}

		// Outbound edge exists. There's three cases now:
		// 1) It points to a value => Key exists
		// 2) It points to a node
		//    a) Which is a prefix key => Key exists
		//    b) Which is not a prefix key => Check further
		hasChild, err := surf.DenseHasChild.Get(labelOffset)
		if err != nil {
			return false, fmt.Errorf("Error accessing bit %d in D-HasChild: %v", labelOffset, err)
		}

		if hasChild == 0 {
			log.Printf("Edge found with HasChild == 0, key exists")
			return true, nil
		} else {
			log.Printf("Edge found with HasChild == 1, must dive deeper")

			// Index of node this edge points to is given by:
			// rank_1(D-HasChild, offset)
			currentNode, err = surf.DenseHasChild.Rank(1, labelOffset)
			if err != nil {
				return false, fmt.Errorf("Error calculating rank_1(%d) over D-HasChild: %v", labelOffset, err)
			}
		}
	}

	// If we get until here, then we traversed the whole key. To determine
	// whether the key exists, we now must check if our current node has
	// IsPrefixKey set to true.
	isPrefixKey, err := surf.DenseIsPrefixKey.Get(currentNode)
	if err != nil {
		return false, fmt.Errorf("Error accessing bit %d in D-IsPrefixKey: %v", currentNode, err)
	}

	if isPrefixKey == 1 {
		log.Printf("Reached end of key on node with IsPrefixKey = 1, key found")
		return true, nil
	} else {
		log.Printf("Reached end of key on node with IsPrefixKey = 0, key not found")
		return false, nil
	}
}

// RangeLookup checks the existence of a key in the [low, high] range,
// including boundaries.
//
// As with Lookup, there is the possibility of false positives.
func (surf *SURF) RangeLookup(low, high []byte) (bool, error) {
	return true, nil
}

// Count returns an approximate count of the number of keys in [low, high],
// including the boundaries.
//
// The count is exact, except for the two boundary cases. As such there is the
// possibility to overcount by up to two.
func (surf *SURF) Count(low, high []byte) (int, error) {
	return 0, nil
}
