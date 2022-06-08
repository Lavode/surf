package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushPop(t *testing.T) {
	stack := Stack[byte]{}

	stack.Push(3)
	stack.Push(17)
	stack.Push(255)

	assert.Equal(t, byte(255), stack.Pop())
	assert.Equal(t, byte(17), stack.Pop())

	stack.Push(20)
	assert.Equal(t, byte(20), stack.Pop())
	assert.Equal(t, byte(3), stack.Pop())
}
