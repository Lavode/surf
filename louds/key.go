package louds

// Key defines a single key which can be stored in a LOUDS-encoded FST tree.
type Key []byte

// Less implements a lexicographic ordering of keys.
//
// It compares pairs of corresponding (at the same index) bytes of the two
// keys. If the two bytes differ, the key with the lesser byte is considered
// lesser.
//
// If all pairs of corresponding bytes are equal, the key with the lesser
// length is lesser.
//
// If the two keys are equal, none is considered lesser than the other.
func (key Key) Less(other Key) bool {
	minLength := len(key)
	if len(other) < minLength {
		minLength = len(other)
	}

	for i := 0; i < minLength; i++ {
		if key[i] < other[i] {
			return true
		} else if key[i] > other[i] {
			return false
		}
	}

	// Shared prefix is equal, so key is lesser if it is shorter
	if len(key) < len(other) {
		return true
	}

	// Else (shared prefix, and other keys smaller, or keys are equal, key
	// is not the lesser).
	return false
}

// Truncate truncates the list of keys such that they are still uniquely
// identifiable.
//
// The inputs must be sorted and may not contain any duplicates. Then, each key
// is truncated to as short a prefix as possible for them to still be
// unique.
//
// As an example, the following keys:
// - far
// - fast
// - john
// Would be truncated to:
// - far
// - fas
// - j
func Truncate(keys []Key) []Key {
	out := make([]Key, len(keys))

	for i, key := range keys {
		// To be able to truncate a key we must find the lowest-indexed
		// bytes where:
		// - It differs from the equivalent byte of the preceeding key
		// - It differs from the equivalent byte of the next key
		//
		// Then, the larger of the two defines the boundary of a prefix
		// of the key such that it can be distinguished from both the
		// key before as well as the one after.

		var firstDifferenceBefore, firstDifferenceAfter int
		var differ bool

		if i != 0 {
			differ, firstDifferenceBefore = FirstDifferenceAt(key, keys[i-1])
			if !differ {
				firstDifferenceBefore = len(key)
			}
		}

		if i != len(keys)-1 {
			differ, firstDifferenceAfter = FirstDifferenceAt(key, keys[i+1])
			if !differ {
				firstDifferenceAfter = len(key)
			}
		}

		var n int
		if firstDifferenceAfter > firstDifferenceBefore {
			n = firstDifferenceAfter
		} else {
			n = firstDifferenceBefore
		}

		if n < len(key) {
			// We have the lowest index such that the prefix differs, so we
			// must include this one
			// However this would be invalid if the current key had
			// be the shorter of the two, as then the first index
			// where the two differ willl already be equal to
			// len(key).
			n++
		}

		out[i] = key[:n]
	}

	return out
}

// FirstDifferenceAt compares two byte slices, finding the first byte where the
// two differ.
//
// The first returned argument indicates whether the two differ. If it is true,
// the second indicates the first index at which the two differ.
//
// If the two are of different length, with the shorter being a prefix of the
// longer, then the first byte of the longer is the one which is considered to
// differ.
func FirstDifferenceAt(a, b []byte) (bool, int) {
	var n int
	if len(a) < len(b) {
		n = len(a)
	} else {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return true, i
		}
	}

	// Either the two are equal, or one is longer, in which case the first
	// difference is the first byte of the longer of the two.
	if len(a) == len(b) {
		return false, 0
	} else {
		return true, n
	}
}
