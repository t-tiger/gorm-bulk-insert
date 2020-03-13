package internal

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_splitObjects(t *testing.T) {
	var objArr []interface{}
	for i := 0; i < 100; i++ {
		objArr = append(objArr, i)
	}

	objSet := SplitObjects(objArr, 30)

	assert.Len(t, objSet, 4)
	assert.Len(t, objSet[len(objSet)-1], 10)
}

func Test_sortedKeys(t *testing.T) {
	value := map[string]interface{}{}
	for i := 1; i <= 9; i++ {
		value[strconv.Itoa(i)] = i
	}

	keys := SortedKeys(value)

	assert.Len(t, keys, 9)
	assert.Equal(t, keys[0], "1")
	assert.Equal(t, keys[len(keys)-1], "9")
}

func Test_containString(t *testing.T) {
	sliceVal := []string{"a", "b", "c"}

	assert.True(t, ContainString(sliceVal, "a"))
	assert.False(t, ContainString(sliceVal, "d"))
}
