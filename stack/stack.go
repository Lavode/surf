package stack

// Stack implements a very basic stack.
//
// Very little is provided in terms of safety or convenience. It is
// specifically not thread-safe.
type Stack[T any] struct {
	data []T
}

// Push adds a new element to the top of the stack.
func (stack *Stack[T]) Push(x T) {
	stack.data = append(stack.data, x)
}

// Pop removes and returns the element on top of the stack.
//
// Popping from an empty stack will cause a panic.
func (stack *Stack[T]) Pop() T {
	idx := len(stack.data) - 1

	x := stack.data[idx]
	stack.data = stack.data[:idx]

	return x
}
