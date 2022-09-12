package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsStringSortedList(t *testing.T) {
	var sortedStrs = []string{"a", "b", "c"}

	assert.EqualValues(t, true, Contains(sortedStrs, "a"))
	assert.EqualValues(t, true, Contains(sortedStrs, "b"))
	assert.EqualValues(t, true, Contains(sortedStrs, "c"))
}

func TestContainsStringUnsortedList(t *testing.T) {
	var sortedStrs = []string{"c", "b", "a"}

	assert.EqualValues(t, true, Contains(sortedStrs, "a"))
	assert.EqualValues(t, true, Contains(sortedStrs, "b"))
	assert.EqualValues(t, true, Contains(sortedStrs, "c"))
}
