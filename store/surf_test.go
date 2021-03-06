package store

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	keys := [][]byte{}

	_, err := New(keys, SURFOptions{})
	if err != nil {
		t.Errorf("Error creating SuRF store with default options: %v", err)
	}

	var r uint = 20
	var hb uint = 30
	var rb uint = 40

	surf, err := New(keys, SURFOptions{R: &r, HashBits: &hb, RealBits: &rb})
	if err != nil {
		t.Errorf("Error creating SuRF store with custom options: %v", err)
	}
	if surf.R != 20 {
		t.Errorf("Expected R to be 20; was %d", surf.R)
	}
	if surf.HashBits != 30 {
		t.Errorf("Expected HashBits to be 30; was %d", surf.R)
	}
	if surf.RealBits != 40 {
		t.Errorf("Expected RealBits to be 40; was %d", surf.R)
	}

}

func TestLookup(t *testing.T) {
	keys := [][]byte{
		[]byte{0x00, 0x01},       // Key in intermediary node
		[]byte{0x00, 0x01, 0x02}, // Key in leaf node
		[]byte{0x42},
		[]byte{0xFF, 0x42, 0x70, 0x71},
	}
	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	for _, k := range keys {
		exists, err := surf.Lookup(k)
		if err != nil {
			t.Fatalf("Error looking up key: %v", err)
		}

		if !exists {
			t.Errorf("Expected key %x to exist; but did not", k)
		}
	}

	nonexistantKeys := [][]byte{
		[]byte{0x00, 0x02},
		[]byte{0x43},
	}

	for _, k := range nonexistantKeys {
		exists, err := surf.Lookup(k)
		if err != nil {
			t.Fatalf("Error looking up key: %v", err)
		}

		if exists {
			t.Errorf("Expected key %x to not exist; but did", k)
		}
	}
}

// Performs a test with the test-data-set pulled from the paper.
func TestLookupPaperTestset(t *testing.T) {
	keys := [][]byte{
		[]byte("farther"),
		[]byte("tries"),
		[]byte("fat"),
		[]byte("trying"),
		[]byte("fasten"),
		[]byte("topper"),
		[]byte("f"),
		[]byte("splice"),
		[]byte("tripper"),
		[]byte("toy"),
		[]byte("fas"),
	}

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	for _, k := range keys {
		exists, err := surf.Lookup(k)
		if err != nil {
			t.Fatalf("Error looking up key: %v", err)
		}

		if !exists {
			t.Errorf("Expected key %s to exist; but did not", k)
		}
	}

	nonexistantKeys := [][]byte{
		[]byte("x"),      // Single-length key
		[]byte("xavier"), // And an extension thereof
		[]byte("fasi"),   // Shared prefix with actual key but not causing a FP
		[]byte("fa"),     // Prefix of actual key but not causing a FP
	}

	for _, k := range nonexistantKeys {
		exists, err := surf.Lookup(k)
		if err != nil {
			t.Fatalf("Error looking up key: %v", err)
		}

		if exists {
			t.Errorf("Expected key %s to not exist; but did", k)
		}
	}

	falsePositiveKeys := [][]byte{
		[]byte("fatter"), // FP with fat,
		[]byte("faster"), // FP with fast(en)
		[]byte("sorry"),  // FP with s
	}

	for _, k := range falsePositiveKeys {
		exists, err := surf.Lookup(k)
		if err != nil {
			t.Fatalf("Error looking up key: %v", err)
		}

		if !exists {
			t.Errorf("Expected false-positive when looking up key %s; but got none", k)
		}
	}
}

func TestLookupFalsePositive(t *testing.T) {
	// This selection of keys will cause a truncated node after the path
	// 0x00 -> 0x01.
	keys := [][]byte{
		[]byte{0x00, 0x01, 0xFF},
		[]byte{0x00, 0x02},
	}

	// So this key will trigger a false positive on a lookup
	falsePositiveKey := []byte{0x00, 0x01, 0xAA}

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	exists, err := surf.Lookup(falsePositiveKey)
	if err != nil {
		t.Fatalf("Error looking up key: %v", err)
	}

	if !exists {
		t.Errorf("Expected false-positive when looking up key %x; but got none", falsePositiveKey)
	}
}

func TestLookupRandom(t *testing.T) {
	const NUM_KEYS = 100_000
	const KEY_LENGTH_MIN = 1
	const KEY_LENGTH_MAX = 50

	rand.Seed(42)

	keys := make([][]byte, NUM_KEYS)
	for i := 0; i < NUM_KEYS; i++ {
		keyLength := rand.Intn(KEY_LENGTH_MAX - KEY_LENGTH_MIN + 1) // [0, max - min + 1)
		keyLength += KEY_LENGTH_MIN                                 // [min, max + 1)

		keys[i] = make([]byte, keyLength)
		rand.Read(keys[i])
	}

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	// All added keys have to be found in there
	for _, k := range keys {
		exists, err := surf.Lookup(k)
		if err != nil {
			t.Fatalf("Error looking up key: %v", err)
		}

		if !exists {
			t.Errorf("Expected key %x to exist; but did not", k)
		}
	}
}

func TestLookupOrGreater(t *testing.T) {
	keys := [][]byte{
		[]byte("farther"),
		[]byte("tries"),
		[]byte("fat"),
		[]byte("trying"),
		[]byte("fasten"),
		[]byte("topper"),
		[]byte("f"),
		[]byte("splice"),
		[]byte("tripper"),
		[]byte("toy"),
		[]byte("fas"),
	}

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	tests := []struct {
		query  []byte
		result []byte
	}{
		{[]byte("a"), []byte("f")},
		{[]byte("fas"), []byte("fas")},
		{[]byte("fal"), []byte("far")},
		{[]byte("fasa"), []byte("fast")},
		{[]byte("t"), []byte("top")},
		{[]byte("trif"), []byte("trip")},
		// That's a potential FP match, as the untruncated key might have been e.g. tripoli
		{[]byte("tripper"), []byte("trip")},
	}

	for _, test := range tests {
		key, _, err := surf.lookupOrGreater(test.query)
		assert.Nil(t, err)
		assert.Equal(t, test.result, key)
	}

	// No next-larger key should be found
	_, _, err = surf.lookupOrGreater([]byte("trz"))
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrEndOfTrie)
}

func TestRangeLookup(t *testing.T) {
	// We'll store all keys in the range [0x0000 .. 0xFF00]
	keys := make([][]byte, 0, 65536)
	for i := 0; i < 0xFF; i++ {
		for j := 0; j <= 0xFF; j++ {
			key := []byte{byte(i), byte(j)}
			keys = append(keys, key)
		}
	}
	// We add 0xFF00 and 0xFF01 as, if we only added 0xFF, truncation would
	// cut it off after 0xFF.
	// Then, any 0xFF... query would get a (false) positive match.
	keys = append(keys, []byte{0xFF, 0x00})
	keys = append(keys, []byte{0xFF, 0x01})

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	// Lookup range fully embedded in key range
	low := []byte{0x20}
	high := []byte{0xB0, 0x27}
	exists, err := surf.RangeLookup(low, high)
	if err != nil {
		t.Fatalf("Error looking up key: %v", err)
	}
	if !exists {
		t.Errorf("Expected range lookup %x -> %x to return true; got false", low, high)
	}

	// Lookup range overlaps lastmost element of key range
	low = []byte{0xF0}
	high = []byte{0xF1}
	exists, err = surf.RangeLookup(low, high)
	if err != nil {
		t.Fatalf("Error looking up key: %v", err)
	}
	if !exists {
		t.Errorf("Expected range lookup %x -> %x to return true; got false", low, high)
	}

	// Lookup range does not overlap key range
	low = []byte{0xFF, 0x02}
	high = []byte{0xFF, 0x20}
	exists, err = surf.RangeLookup(low, high)
	if err != nil {
		t.Fatalf("Error looking up key: %v", err)
	}
	if exists {
		t.Errorf("Expected range lookup %x -> %x to return false; got true", low, high)
	}
}

func TestRangeLookupPaperDataset(t *testing.T) {
	keys := [][]byte{
		[]byte("farther"),
		[]byte("tries"),
		[]byte("fat"),
		[]byte("trying"),
		[]byte("fasten"),
		[]byte("topper"),
		[]byte("f"),
		[]byte("splice"),
		[]byte("tripper"),
		[]byte("toy"),
		[]byte("fas"),
	}

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	tests := []struct {
		lower  []byte
		upper  []byte
		hasKey bool
	}{
		{[]byte("a"), []byte("ezmatch"), false},
		{[]byte("a"), []byte("f"), true},
		{[]byte("a"), []byte("fat"), true},
		{[]byte("fal"), []byte("fat"), true},
		{[]byte("r"), []byte("s"), true},
		{[]byte("s"), []byte("s"), true},
		{[]byte("tp"), []byte("tq"), false},
		{[]byte("tp"), []byte("ts"), true},
		{[]byte("tripper"), []byte("try"), true},
		{[]byte("tripper"), []byte("zarty"), true},
		{[]byte("trz"), []byte("zarty"), false},
	}

	for _, test := range tests {
		hasKey, err := surf.RangeLookup(test.lower, test.upper)
		assert.Nil(t, err)

		assert.Equal(t, test.hasKey, hasKey, "Expected range query from %s to %s to find keys, did not", test.lower, test.upper)
	}
}

func TestCount(t *testing.T) {
	keys := [][]byte{
		[]byte{0x00, 0x01},       // Key in intermediary node
		[]byte{0x00, 0x01, 0x02}, // Key in leaf node
		[]byte{0x42},
		[]byte{0xFF, 0x42, 0x70, 0x71},
		[]byte{0xFF, 0x42, 0x70, 0x72}, // Ensure key above isn't truncated
	}
	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	// Keys fully contained in range
	low := []byte{0x00}
	high := []byte{0xA0}
	cnt, err := surf.Count(low, high)
	if err != nil {
		t.Fatalf("Error counting keys: %v", err)
	}
	if cnt != 3 {
		t.Errorf("Expected 3 keys in range %x to %x; got %d", low, high, cnt)
	}

	// Keys fully outside of range
	low = []byte{0x43}
	high = []byte{0xFF, 0x42, 0x70, 0x70}
	cnt, err = surf.Count(low, high)
	if err != nil {
		t.Fatalf("Error counting keys: %v", err)
	}
	if cnt != 0 {
		t.Errorf("Expected 0 keys in range %x to %x; got %d", low, high, cnt)
	}
}

func TestCountPaperDataset(t *testing.T) {
	keys := [][]byte{
		[]byte("farther"),
		[]byte("tries"),
		[]byte("fat"),
		[]byte("trying"),
		[]byte("fasten"),
		[]byte("topper"),
		[]byte("f"),
		[]byte("splice"),
		[]byte("tripper"),
		[]byte("toy"),
		[]byte("fas"),
	}

	surf, err := New(keys, SURFOptions{})
	if err != nil {
		t.Fatalf("Error creating SuRF store: %v", err)
	}

	tests := []struct {
		lower []byte
		upper []byte
		count int
	}{
		{[]byte("a"), []byte("ezmatch"), 0},
		{[]byte("a"), []byte("f"), 1},
		{[]byte("a"), []byte("fat"), 5},
		{[]byte("fal"), []byte("fat"), 4},
		{[]byte("s"), []byte("s"), 1},
		{[]byte("tp"), []byte("tq"), 0},
		{[]byte("tp"), []byte("ts"), 3},
		{[]byte("tripper"), []byte("try"), 2},
		{[]byte("tripper"), []byte("zarty"), 2},
		{[]byte("trz"), []byte("zarty"), 0},
	}

	for _, test := range tests {
		count, err := surf.Count(test.lower, test.upper)
		assert.Nil(t, err)

		assert.Equal(t, test.count, count, "Expected count from %s to %s = %d, got %d", test.lower, test.upper, test.count, count)
	}
}
