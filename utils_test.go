package main

import (
	"strconv"
	"testing"
)

func Test_splitObjects(t *testing.T) {
	var objArr []interface{}

	for i := 0; i < 100; i++ {
		objArr = append(objArr, i)
	}

	objSet := splitObjects(objArr, 30)

	if len(objSet) != 4 {
		t.Error("split size must be 4")
	}
	if len(objSet[len(objSet)-1]) != 10 {
		t.Error("last chunk size must be 10")
	}
}

func Test_sortedKeys(t *testing.T) {
	value := map[string]interface{}{}
	for i := 1; i <= 9; i++ {
		value[strconv.Itoa(i)] = i
	}

	keys := sortedKeys(value)

	if len(keys) != 9 {
		t.Error("key size must be 9")
	}
	if keys[0] != "1" {
		t.Error("first key must be 9")
	}
	if keys[len(keys)-1] != "9" {
		t.Error("first key must be 9")
	}
}
