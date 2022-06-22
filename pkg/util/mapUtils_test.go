package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapMergeSuccess(t *testing.T) {
	map1 := map[string]string{"a": "test1", "b": "test2"}
	map2 := map[string]string{"c": "test1"}

	MergeMaps(map1, map2)
	assert.EqualValues(t, len(map1), 3)
	assert.EqualValues(t, map1["a"], "test1")
	assert.EqualValues(t, map1["b"], "test2")
	assert.EqualValues(t, map1["c"], "test1")
}

func TestMap1EmptyMergeSuccess(t *testing.T) {
	map1 := map[string]string{}
	map2 := map[string]string{"c": "test1"}

	MergeMaps(map1, map2)
	assert.EqualValues(t, len(map1), 1)
	assert.EqualValues(t, map1["c"], "test1")
}

func TestMap2EmptyMergeSuccess(t *testing.T) {
	map1 := map[string]string{"a": "test1", "b": "test2"}
	map2 := map[string]string{}

	MergeMaps(map1, map2)
	assert.EqualValues(t, len(map1), 2)
	assert.EqualValues(t, map1["a"], "test1")
	assert.EqualValues(t, map1["b"], "test2")
}

func TestMapOverrideMergeSuccess(t *testing.T) {
	map1 := map[string]string{"a": "test1", "b": "test2"}
	map2 := map[string]string{"a": "test3"}

	MergeMaps(map1, map2)
	assert.EqualValues(t, len(map1), 2)
	assert.EqualValues(t, map1["a"], "test3")
	assert.EqualValues(t, map1["b"], "test2")
}
