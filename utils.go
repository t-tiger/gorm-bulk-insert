package main

import "sort"

// Separate objects into several size
func splitObjects(objArr []interface{}, size int) [][]interface{} {
	var chunkSet [][]interface{}
	var chunk []interface{}

	for len(objArr) > size {
		chunk, objArr = objArr[:size], objArr[size:]
		chunkSet = append(chunkSet, chunk)
	}
	if len(objArr) > 0 {
		chunkSet = append(chunkSet, objArr[:])
	}

	return chunkSet
}

// Enable map keys to be retrieved in same order when iterating
func sortedKeys(val map[string]interface{}) []string {
	var keys []string
	for key := range val {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}