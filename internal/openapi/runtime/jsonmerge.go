// Code generated from github.com/oapi-codegen/runtime; DO NOT EDIT.

package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// JSONMerge merges two JSON representations into a single object. data is the
// existing representation and patch is the new data to be merged in.
//
// This is a dependency-free reimplementation of the merge behavior the
// generated client previously obtained from github.com/apapsch/go-jsonmerge/v2
// (configured with CopyNonexistent: true). It mirrors that semantics exactly:
//
//   - both values are objects: keys are merged recursively; keys present only in
//     data are kept; keys present only in patch are added.
//   - data value is an object but the patch value is not: data wins (type
//     mismatch is preserved rather than overwritten).
//   - data value is an array and the patch value is an object: array elements are
//     merged by their string index.
//   - otherwise: the patch value wins.
//
// Numbers are decoded with json.Number to preserve their original formatting and
// precision, matching the previous implementation.
func JSONMerge(data, patch json.RawMessage) (json.RawMessage, error) {
	if len(data) == 0 {
		data = []byte("{}")
	}
	if len(patch) == 0 {
		patch = []byte("{}")
	}

	var dataVal, patchVal interface{}
	if err := decodeUseNumber(data, &dataVal); err != nil {
		return nil, fmt.Errorf("error in data JSON: %w", err)
	}
	if err := decodeUseNumber(patch, &patchVal); err != nil {
		return nil, fmt.Errorf("error in patch JSON: %w", err)
	}

	merged, err := json.Marshal(mergeObjects(dataVal, patchVal))
	if err != nil {
		return nil, fmt.Errorf("error writing merged JSON: %w", err)
	}
	return merged, nil
}

// decodeUseNumber unmarshals JSON while preserving numbers as json.Number.
func decodeUseNumber(buf []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.UseNumber()
	return dec.Decode(v)
}

// mergeObjects merges patch into data. patch must be an object for any merging to
// occur; otherwise data is returned unchanged.
func mergeObjects(data, patch interface{}) interface{} {
	patchObj, ok := patch.(map[string]interface{})
	if !ok {
		return data
	}

	switch d := data.(type) {
	case []interface{}:
		ret := make([]interface{}, len(d))
		for i, v := range d {
			ret[i] = mergeValue(patchObj, strconv.Itoa(i), v)
		}
		return ret
	case map[string]interface{}:
		ret := make(map[string]interface{}, len(d))
		for k, v := range d {
			ret[k] = mergeValue(patchObj, k, v)
		}
		for k, v := range patchObj {
			if _, exists := d[k]; !exists {
				ret[k] = v
			}
		}
		return ret
	default:
		return data
	}
}

// mergeValue merges a single keyed value from patch into the data value.
func mergeValue(patch map[string]interface{}, key string, value interface{}) interface{} {
	patchValue, ok := patch[key]
	if !ok {
		return value
	}

	_, patchValueIsObject := patchValue.(map[string]interface{})

	if _, ok := value.(map[string]interface{}); ok {
		if !patchValueIsObject {
			return value
		}
		return mergeObjects(value, patchValue)
	}

	if _, ok := value.([]interface{}); ok && patchValueIsObject {
		return mergeObjects(value, patchValue)
	}

	return patchValue
}
