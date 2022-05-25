package main

import (
	"fmt"
	"log"

	"github.com/Lavode/surf/louds"
	"github.com/Lavode/surf/louds/dense"
	"golang.org/x/exp/slices"
)

const BITMAP_CAPACITY = 80_000_000 // 10 MB

func main() {
	keys := []louds.Key{
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

	// Sort & truncate keys

	sort := func(x, y louds.Key) bool {
		return x.Less(y)
	}
	slices.SortFunc(keys, sort)

	keys = louds.Truncate(keys)

	fmt.Println("Sorted & truncated keys:")
	for _, k := range keys {
		fmt.Printf("\t%s\n", k)
	}

	builder := dense.NewBuilder(BITMAP_CAPACITY)
	fmt.Printf("Node capacity: %d\n", builder.NodeCapacity())

	err := builder.Build(keys)
	if err != nil {
		log.Panicf("Error building tree: %v", err)
	}

	fmt.Printf("Labels:\n%s\n", builder.Labels)
	fmt.Printf("HasChild:\n%s\n", builder.HasChild)
	fmt.Printf("IsPrefixKey:\n%s\n", builder.HasChild)
}
