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
}
