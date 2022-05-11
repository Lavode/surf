package louds

// Key specifies a single key which can be stored in a LOUDS-encoded FST tree.c
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
