package dense

import (
	"fmt"
	"log"

	"github.com/Lavode/surf/bitmap"
	"github.com/Lavode/surf/louds"
)

// NodeTask contains things which need to be considered for building up one
// future node.
//
// This includes keys whose path contains that node, but also additional
// information such as that the node might be a prefix key.
type NodeTask struct {
	// keys is the slice of keys which will lead to edges on the future node.
	keys []louds.Key
	// isPrefixKey defines whether this node's isPrefixKey flag will have
	// to be set to true - if the node will exist at all.
	isPrefixKey bool
}

type Builder struct {
	Labels      *bitmap.Bitmap
	HasChild    *bitmap.Bitmap
	IsPrefixKey *bitmap.Bitmap

	// tasks is a slice of tasks to be taken care of to define nodes
	// further down the tree.
	// There is a 1:1 correspondence between tasks and (potential) future
	// nodes.
	tasks []*NodeTask
	// currentTask is a pointer to the most recent element of the tasks
	// slice.
	// As such it is not the task currently being worked on, but the task
	// currently being defined. A bit of a misnormer I admit.
	currentTask *NodeTask

	// currentNodeId is the 0-indexed level-order ID of the node we are
	// currently building up.
	currentNodeId int
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
		tasks:       make([]*NodeTask, 0),
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
	// For depth = 0 we'll consider all keys
	builder.appendNodeTask()
	builder.currentTask.keys = keys

	for depth := 0; depth < maxKeyLength(keys); depth++ {
		log.Printf("Current depth: %d", depth)

		// During interation we'll be adding tasks of the next tree
		// level. But we only want to consider tasks of the current
		// level.
		n := len(builder.tasks)
		for i := 0; i < n; i++ {
			task := builder.tasks[i]

			if len(task.keys) == 0 {
				log.Printf("\tSkipping empty task")
				continue
			}

			log.Printf("\tConsidering pending task with keys %s, isPrefixKey %t (now in node %d)", task.keys, task.isPrefixKey, builder.currentNodeId)
			// Each task corresponds to one node, so us starting
			// with a new task means we're populating a new node
			nodeHasEdges := false
			var mostRecentEdge byte = 0x00

			err := builder.initializeNode()
			if err != nil {
				return err
			}

			// If the node is non-empty (which is the case if we are here), and the task has
			// its isPrefixKey flag set, then that means that one key ended on this node.
			if task.isPrefixKey {
				log.Printf("\t\tSetting node's is-prefix-key flag")
				builder.setIsPrefixKey()
			}

			for _, key := range task.keys {
				edge := key[depth]
				log.Printf("\t\tConsidering key %s (edge = %c)", key, edge)

				if !nodeHasEdges || mostRecentEdge != edge {
					log.Printf("\t\t\tDefining new edge with value %c", edge)

					err := builder.addEdge(edge)
					if err != nil {
						return err
					}

					// Having added a new edge means that there will also, on the next level,
					// be a new node which future keys (if we're not at their end yet) will go into.
					builder.appendNodeTask()

					mostRecentEdge = edge
					nodeHasEdges = true
				}

				if depth == len(key)-1 {
					log.Printf("\t\t\tReached end of key")

					// Key ends at the node next to its edge, so if that node exists it will
					// have to have IsPrefixKey set to true
					builder.currentTask.isPrefixKey = true
				} else {
					err := builder.setHasChild(edge)
					if err != nil {
						return nil
					}

					log.Printf("\t\t\tKey must be considered further, adding to tasks of next level")
					builder.currentTask.keys = append(builder.currentTask.keys, key)
				}

			}

			// Reached end of the current node.
			builder.currentNodeId++
		}

		// We processed all tasks of the current level, so we'll
		// discard them.
		builder.tasks = builder.tasks[n:]
	}
	return nil
}

func (builder *Builder) addEdge(edge byte) error {
	bit := builder.labelOffset() + int(edge)
	err := builder.Labels.Set(bit)

	if err != nil {
		return fmt.Errorf(
			"LOUDS-Dense builder: Error setting bit for edge %x at %d: %v",
			edge,
			bit,
			err,
		)
	}

	return nil
}

func (builder *Builder) setHasChild(edge byte) error {
	bit := builder.hasChildOffset() + int(edge)
	err := builder.HasChild.Set(bit)

	if err != nil {
		return fmt.Errorf(
			"LOUDS-Dense builder: Error setting bit for has-child of edge %x at %d: %v",
			edge,
			bit,
			err,
		)
	}

	return nil
}

func (builder *Builder) setIsPrefixKey() error {
	bit := builder.isPrefixKeyOffset()
	err := builder.IsPrefixKey.Set(bit)

	if err != nil {
		return fmt.Errorf(
			"LOUDS-Dense builder: Error setting bit for is-prefix-key at %d: %v",
			bit,
			err,
		)
	}

	return nil
}

// labelOffset returns the offset in the D-Labels bitmap of the currently
// processed node.
func (builder *Builder) labelOffset() int {
	return builder.currentNodeId * 256
}

// hasChildOffset returns the offset in the D-HasChild bitmap of the currently
// processed node.
func (builder *Builder) hasChildOffset() int {
	return builder.currentNodeId * 256
}

// isPrefixKeyOffset returns the offset in the D-IsPrefixKey bitmap of the
// currently processed node.
func (builder *Builder) isPrefixKeyOffset() int {
	return builder.currentNodeId
}

// initialzeNode initialies the various bitmaps to accomodate the new node with
// ID builder.currentNodeId.
func (builder *Builder) initializeNode() error {
	// We'll make sure the current node's extents in the various bitmaps is
	// allocated.
	// This is not strictly needed, but makes for cleaner / easier to test
	// results.

	_, err := builder.Labels.Get(builder.labelOffset() + 255)
	if err != nil {
		return fmt.Errorf("LOUDS-Dense builder: Error initializing labels bitmap: %v", err)
	}

	_, err = builder.HasChild.Get(builder.hasChildOffset() + 255)
	if err != nil {
		return fmt.Errorf("LOUDS-Dense builder: Error initializing has-child bitmap: %v", err)
	}

	_, err = builder.IsPrefixKey.Get(builder.isPrefixKeyOffset())
	if err != nil {
		return fmt.Errorf("LOUDS-Dense builder: Error initializing is-prefix-key bitmap: %v", err)
	}

	return nil
}

func (builder *Builder) appendNodeTask() {
	task := &NodeTask{
		keys: make([]louds.Key, 0),
	}

	builder.tasks = append(builder.tasks, task)
	builder.currentTask = task
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
