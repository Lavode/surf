package main

import (
	"fmt"
	"log"

	"github.com/Lavode/surf/store"
)

const BITMAP_CAPACITY = 80_000_000 // 10 MB

func main() {
	keys := [][]byte{
		[]byte("f"),
		[]byte("farther"),
		[]byte("fas"),
		[]byte("fasten"),
		[]byte("fat"),
		[]byte("splice"),
		[]byte("topper"),
		[]byte("toy"),
		[]byte("tries"),
		[]byte("tripper"),
		[]byte("trying"),
	}

	surf, err := store.New(keys, store.SURFOptions{})
	if err != nil {
		log.Fatalf("Error instantinating SuRF: %v", err)
	}

	fmt.Printf("D-Labels:\n%s\n", surf.DenseLabels)
	fmt.Printf("D-HasChild:\n%s\n", surf.DenseHasChild)
	fmt.Printf("D-IsPrefixKey:\n%s\n", surf.DenseHasChild)

	lookupKeys := [][]byte{
		[]byte("fasten"), // True positive
		[]byte("faster"), // False positive
		[]byte("fasi"),   // True negative
		[]byte("f"),      // True positive (and prefix key)
		[]byte("s"),      // True positive (and prefix key)
		[]byte("fa"),     // True negative (and not a prefix key)
	}

	for _, key := range lookupKeys {
		exists, err := surf.Lookup(key)

		if err != nil {
			log.Fatalf("Error looking up key %s: %v", key, err)
		}

		log.Printf("Looked up key %s: %t\n\n", key, exists)
	}
}
