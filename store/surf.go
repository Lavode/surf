package store

import (
	"fmt"

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
	return true, nil
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
