package store

import (
	"errors"
	"fmt"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
	"github.com/Lavode/surf/louds/dense"
	"golang.org/x/exp/slices"
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

	// Keys must be sorted for truncation to work
	sort := func(x, y louds.Key) bool {
		return x.Less(y)
	}
	slices.SortFunc(keys, sort)

	// Truncate keys
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
	exists, _, _, err := surf.lookup(key)

	return exists, err
}

// lookup checks existence of a key in the SuRF store (see documentation of
// Lookup).
//
// In addition it will yield the actual key which was found as well as the
// iterator in whatever state it was then lookup terminated.
// This can be used to e.g. implement find-next-greater-than.
//
// As this leaks internal structure this method is not part of the public API.
func (surf *SURF) lookup(key []byte) (bool, []byte, Iterator, error) {
	it := Iterator{
		Labels:      surf.DenseLabels,
		HasChild:    surf.DenseHasChild,
		IsPrefixKey: surf.DenseIsPrefixKey,
	}

	for i := 0; i < len(key); i++ {
		keyByte := key[i]

		err := it.GoToChild(keyByte)
		if err != nil {
			if errors.Is(err, ErrNoSuchEdge) {
				// No edge with this value, so the key doesn't exist.
				return false, []byte{}, it, nil
			} else if errors.Is(err, ErrIsLeaf) {
				// We attempted to enter a leaf node, so the key exists
				return true, key[:i+1], it, nil
			} else {
				// Non-specific error, e.g. issue with bitmap access
				return false, []byte{}, it, err
			}
		}
	}

	// If we get until here, then we traversed the whole key. To determine
	// whether the key exists, we now must check if our current node has
	// IsPrefixKey set to true.
	isPrefixKey, err := surf.DenseIsPrefixKey.Get(it.NodeIndex)
	if err != nil {
		return false, []byte{}, it, fmt.Errorf("Error accessing bit %d in D-IsPrefixKey: %v", it.NodeIndex, err)
	}

	if isPrefixKey == 1 {
		return true, key, it, nil
	} else {
		return false, []byte{}, it, nil
	}
}

// LookupOrGreater checks existence of a key in the SuRF store.
//
// If it is found it is returned. If not, then the next greater key is
// returned.
//
// If no greater key is found, ErrEndOfTrie is returned.
func (surf *SURF) LookupOrGreater(key []byte) ([]byte, error) {
	exists, matchedKey, it, err := surf.lookup(key)
	if err != nil {
		return []byte{}, err
	}

	if exists {
		// We won't return `key` but rather the (potentially truncated)
		// key stored in the FST.
		return matchedKey, nil
	} else {
		// We can easily find the next larger key by telling the
		// iterator to find the next key from where it is at currently.
		largerKey, err := it.Next()
		if errors.Is(err, ErrEndOfTrie) {
			return []byte{}, err
		} else if err != nil {
			return []byte{}, err
		}

		return largerKey, nil
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
