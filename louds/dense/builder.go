package dense

import (
	"fmt"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
)

type Builder struct {
	Labels      *bitmap.Bitmap
	HasChild    *bitmap.Bitmap
	IsPrefixKey *bitmap.Bitmap

	// keys contains the keys which the builder will use to build up an FST
	// tree.
	keys []louds.Key

	// groups contains the keys, grouped by which node of the current level
	// they belong to.
	// For currentLevel = 0, it will contain one group equal to keys.
	// For currentLevel = 1, it will contain the keys grouped by their
	// first byte, and so on.
	groups [][]louds.Key

	// Node ID in level-order of node we are currently building up.
	currentNodeId int
	// currentLevel specifies the current level of the tree we are working
	// on.
	currentLevel int
	// currentEdge specifies the most recently added edge of the current node.
	currentEdge byte
	// currentNodeHasEdges specifies whether the current node has had any
	// edges added
	currentNodeHasEdges bool
}

// NewBuilder instantiates a new LOUDS-DENSE builder.
//
// memory_limit specifies the memory limits in bits.
func NewBuilder(memory_limit int) *Builder {
	// Labels and HasChild are 256 bit per node, IsPrefixKey is 1 bit per
	// node.
	memory_unit := memory_limit / (256 + 256 + 1)

	builder := Builder{
		Labels:      bitmap.New(256, 256*memory_unit),
		HasChild:    bitmap.New(256, 256*memory_unit),
		IsPrefixKey: bitmap.New(1, memory_unit),
		groups:      make([][]louds.Key, 0),
	}

	return &builder
}

// NodeCapacity returns the number of nodes which can be encoded by this
// builder in total.
func (builder *Builder) NodeCapacity() int {
	return builder.IsPrefixKey.Capacity
}

// Build instantiates a LOUDS-DENSE encoded tree using the given keys.
//
// Build may only be called on a freshly created instance. Calling Build on a
// builder more than once is not guaranteed to produce a consistent tree.
func (builder *Builder) Build(keys []louds.Key) error {
	builder.keys = keys

	for _, key := range builder.keys {
		fmt.Printf("Setting key: %x\n", key)
		err := builder.addEdge(key[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func (builder *Builder) addEdge(label byte) error {
	// We must not skip any keys if we haven't actually added any edges
	// yet, as in this case currentEdge's default value will be 0x00 which
	// might be equal to the key.
	if builder.currentNodeHasEdges && builder.currentEdge == label {
		return nil
	}

	bit := builder.labelsOffset() + int(label)
	fmt.Printf("Added edge for key = %x at bitmap offset = %x\n", label, bit)
	err := builder.Labels.Set(bit)
	if err != nil {
		return fmt.Errorf("DENSE builder: Error setting label: %v", err)
	}

	builder.currentEdge = label

	return nil
}

func (builder *Builder) beginNode() {
}

func (builder *Builder) setLabel(nodeId int, label byte) error {
	offset := 256*nodeId + int(label)
	err := builder.Labels.Set(offset)
	if err != nil {
		return fmt.Errorf("DENSE Builder: Error setting Label: %v", err)
	}

	return nil
}

func maxKeyLength(keys []louds.Key) int {
	maxKeyLength := 0
	for _, k := range keys {
		if len(k) > maxKeyLength {
			maxKeyLength = len(k)
		}
	}

	return maxKeyLength
}

// labelsOffset returns the offset where the current node starts in the Labels
// bitmap.
func (builder *Builder) labelsOffset() int {
	return 256 * builder.currentNodeId
}

// hasChildOffset returns the offset where the current node starts in the
// HasChild bitmap.
func (builder *Builder) hasChildOffset() int {
	return 256 * builder.currentNodeId
}

// isPrefixKeyOffset returns the offset where the current node starts in the
// IsPreixKey bitmap.
func (builder *Builder) isPrefixKeyOffset() int {
	return builder.currentNodeId
}
