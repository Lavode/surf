package bitops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirst(t *testing.T) {
	var b uint64 = 0b0001101111001100000111111010100110101111111011110101000010100001

	tests := []struct {
		n        int
		expected uint64
	}{
		{1, 0b0000000000000000000000000000000000000000000000000000000000000000},
		{3, 0b0000000000000000000000000000000000000000000000000000000000000000},
		{17, 0b0001101111001100000000000000000000000000000000000000000000000000},
		{62, 0b0001101111001100000111111010100110101111111011110101000010100000},
		{64, 0b0001101111001100000111111010100110101111111011110101000010100001},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, FirstBits(test.n, b))
	}
}

func TestLast(t *testing.T) {
	var b uint64 = 0b0001101111001100000111111010100110101111111011110101000010100001

	tests := []struct {
		n        int
		expected uint64
	}{
		{1, 0b0000000000000000000000000000000000000000000000000000000000000001},
		{3, 0b0000000000000000000000000000000000000000000000000000000000000001},
		{17, 0b0000000000000000000000000000000000000000000000010101000010100001},
		{62, 0b0001101111001100000111111010100110101111111011110101000010100001},
		{64, 0b0001101111001100000111111010100110101111111011110101000010100001},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, LastBits(test.n, b))
	}

}

func TestLeadingOnesMask(t *testing.T) {
	tests := []struct {
		n        int
		expected uint64
	}{
		{-5, 0x0000000000000000},
		{0, 0x0000000000000000},
		{1, 0x8000000000000000},
		{2, 0xC000000000000000},
		{3, 0xE000000000000000},
		{4, 0xF000000000000000},
		{8, 0xFF00000000000000},
		{32, 0xFFFFFFFF00000000},
		{62, 0xFFFFFFFFFFFFFFFC},
		{64, 0xFFFFFFFFFFFFFFFF},
		{70, 0xFFFFFFFFFFFFFFFF},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, LeadingOnesMask(test.n))
	}
}

func TestTrailingOnesMask(t *testing.T) {
	tests := []struct {
		n        int
		expected uint64
	}{
		{-5, 0x0000000000000000},
		{0, 0x0000000000000000},
		{1, 0x0000000000000001},
		{2, 0x0000000000000003},
		{3, 0x0000000000000007},
		{4, 0x000000000000000F},
		{8, 0x00000000000000FF},
		{32, 0x00000000FFFFFFFF},
		{62, 0x3FFFFFFFFFFFFFFF},
		{64, 0xFFFFFFFFFFFFFFFF},
		{70, 0xFFFFFFFFFFFFFFFF},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, TrailingOnesMask(test.n))
	}
}

func TestOnesMask(t *testing.T) {
	tests := []struct {
		leading  int
		trailing int
		expected uint64
	}{
		{0, 0, 0x0000000000000000},
		{3, 17, 0xE00000000001FFFF},
		{24, 32, 0xFFFFFF00FFFFFFFF},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, OnesMask(test.leading, test.trailing))
	}
}

func TestSingleOneMask(t *testing.T) {
	tests := []struct {
		n        int
		expected uint64
	}{
		{-5, 0x8000000000000000},
		{0, 0x8000000000000000},
		{1, 0x4000000000000000},
		{2, 0x2000000000000000},
		{3, 0x1000000000000000},
		{4, 0x0800000000000000},
		{8, 0x0080000000000000},
		{32, 0x0000000080000000},
		{62, 0x0000000000000002},
		{63, 0x0000000000000001},
		{70, 0x0000000000000001},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, SingleOneMask(test.n))
	}
}
