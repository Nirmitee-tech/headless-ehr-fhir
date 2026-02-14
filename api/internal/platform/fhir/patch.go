package fhir

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// PatchOperation represents a single JSON Patch operation (RFC 6902).
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
	From  string      `json:"from,omitempty"`
}

// ApplyJSONPatch applies a JSON Patch (RFC 6902) to a FHIR resource map.
func ApplyJSONPatch(resource map[string]interface{}, patchOps []PatchOperation) (map[string]interface{}, error) {
	result := deepCopyMap(resource)

	for i, op := range patchOps {
		var err error
		switch op.Op {
		case "add":
			err = patchAdd(result, op.Path, op.Value)
		case "remove":
			err = patchRemove(result, op.Path)
		case "replace":
			err = patchReplace(result, op.Path, op.Value)
		case "move":
			err = patchMove(result, op.From, op.Path)
		case "copy":
			err = patchCopy(result, op.From, op.Path)
		case "test":
			err = patchTest(result, op.Path, op.Value)
		default:
			err = fmt.Errorf("unknown patch operation: %s", op.Op)
		}
		if err != nil {
			return nil, fmt.Errorf("patch operation %d (%s) failed: %w", i, op.Op, err)
		}
	}

	return result, nil
}

// ApplyMergePatch applies a JSON Merge Patch (RFC 7386) to a FHIR resource map.
func ApplyMergePatch(resource map[string]interface{}, patch map[string]interface{}) (map[string]interface{}, error) {
	result := deepCopyMap(resource)
	mergePatchRecursive(result, patch)
	return result, nil
}

func mergePatchRecursive(target, patch map[string]interface{}) {
	for key, patchVal := range patch {
		if patchVal == nil {
			delete(target, key)
			continue
		}

		patchMap, patchIsMap := patchVal.(map[string]interface{})
		if patchIsMap {
			targetVal, targetExists := target[key]
			targetMap, targetIsMap := targetVal.(map[string]interface{})
			if targetExists && targetIsMap {
				mergePatchRecursive(targetMap, patchMap)
			} else {
				target[key] = deepCopyMap(patchMap)
			}
		} else {
			target[key] = patchVal
		}
	}
}

// ParseJSONPatch parses a JSON Patch document from raw JSON.
func ParseJSONPatch(data []byte) ([]PatchOperation, error) {
	var ops []PatchOperation
	if err := json.Unmarshal(data, &ops); err != nil {
		return nil, fmt.Errorf("invalid JSON Patch document: %w", err)
	}
	for i, op := range ops {
		if op.Op == "" {
			return nil, fmt.Errorf("patch operation %d: missing 'op' field", i)
		}
		if op.Path == "" && op.Op != "test" {
			return nil, fmt.Errorf("patch operation %d: missing 'path' field", i)
		}
	}
	return ops, nil
}

// ParseMergePatch parses a JSON Merge Patch document from raw JSON.
func ParseMergePatch(data []byte) (map[string]interface{}, error) {
	var patch map[string]interface{}
	if err := json.Unmarshal(data, &patch); err != nil {
		return nil, fmt.Errorf("invalid JSON Merge Patch document: %w", err)
	}
	return patch, nil
}

// --- Internal patch operations ---

func patchAdd(doc map[string]interface{}, path string, value interface{}) error {
	if path == "" || path == "/" {
		// Replace root (not typical for FHIR)
		return fmt.Errorf("cannot replace root document")
	}

	parent, lastKey, err := resolvePath(doc, path, true)
	if err != nil {
		return err
	}

	switch p := parent.(type) {
	case map[string]interface{}:
		p[lastKey] = value
	case []interface{}:
		idx, err := strconv.Atoi(lastKey)
		if err != nil {
			if lastKey == "-" {
				// Append to end
				parentMap, parentKey := resolveParentOfPath(doc, path)
				if parentMap != nil {
					parentMap[parentKey] = append(p, value)
				}
				return nil
			}
			return fmt.Errorf("invalid array index: %s", lastKey)
		}
		if idx < 0 || idx > len(p) {
			return fmt.Errorf("array index out of bounds: %d", idx)
		}
		newArr := make([]interface{}, len(p)+1)
		copy(newArr, p[:idx])
		newArr[idx] = value
		copy(newArr[idx+1:], p[idx:])
		parentMap, parentKey := resolveParentOfPath(doc, path)
		if parentMap != nil {
			parentMap[parentKey] = newArr
		}
	}
	return nil
}

func patchRemove(doc map[string]interface{}, path string) error {
	parent, lastKey, err := resolvePath(doc, path, false)
	if err != nil {
		return err
	}

	switch p := parent.(type) {
	case map[string]interface{}:
		if _, ok := p[lastKey]; !ok {
			return fmt.Errorf("path not found: %s", path)
		}
		delete(p, lastKey)
	case []interface{}:
		idx, err := strconv.Atoi(lastKey)
		if err != nil {
			return fmt.Errorf("invalid array index: %s", lastKey)
		}
		if idx < 0 || idx >= len(p) {
			return fmt.Errorf("array index out of bounds: %d", idx)
		}
		newArr := append(p[:idx], p[idx+1:]...)
		parentMap, parentKey := resolveParentOfPath(doc, path)
		if parentMap != nil {
			parentMap[parentKey] = newArr
		}
	}
	return nil
}

func patchReplace(doc map[string]interface{}, path string, value interface{}) error {
	parent, lastKey, err := resolvePath(doc, path, false)
	if err != nil {
		return err
	}

	switch p := parent.(type) {
	case map[string]interface{}:
		if _, ok := p[lastKey]; !ok {
			return fmt.Errorf("path not found: %s", path)
		}
		p[lastKey] = value
	case []interface{}:
		idx, err := strconv.Atoi(lastKey)
		if err != nil {
			return fmt.Errorf("invalid array index: %s", lastKey)
		}
		if idx < 0 || idx >= len(p) {
			return fmt.Errorf("array index out of bounds: %d", idx)
		}
		p[idx] = value
	}
	return nil
}

func patchMove(doc map[string]interface{}, from, path string) error {
	// Get value from source
	parent, lastKey, err := resolvePath(doc, from, false)
	if err != nil {
		return fmt.Errorf("move from: %w", err)
	}

	var value interface{}
	switch p := parent.(type) {
	case map[string]interface{}:
		value = p[lastKey]
	case []interface{}:
		idx, _ := strconv.Atoi(lastKey)
		value = p[idx]
	}

	// Remove from source
	if err := patchRemove(doc, from); err != nil {
		return fmt.Errorf("move remove: %w", err)
	}

	// Add to destination
	if err := patchAdd(doc, path, value); err != nil {
		return fmt.Errorf("move add: %w", err)
	}

	return nil
}

func patchCopy(doc map[string]interface{}, from, path string) error {
	parent, lastKey, err := resolvePath(doc, from, false)
	if err != nil {
		return fmt.Errorf("copy from: %w", err)
	}

	var value interface{}
	switch p := parent.(type) {
	case map[string]interface{}:
		value = p[lastKey]
	case []interface{}:
		idx, _ := strconv.Atoi(lastKey)
		value = p[idx]
	}

	return patchAdd(doc, path, value)
}

func patchTest(doc map[string]interface{}, path string, expected interface{}) error {
	parent, lastKey, err := resolvePath(doc, path, false)
	if err != nil {
		return fmt.Errorf("test path not found: %w", err)
	}

	var actual interface{}
	switch p := parent.(type) {
	case map[string]interface{}:
		actual = p[lastKey]
	case []interface{}:
		idx, _ := strconv.Atoi(lastKey)
		actual = p[idx]
	}

	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if string(actualJSON) != string(expectedJSON) {
		return fmt.Errorf("test failed: expected %s but got %s at %s", string(expectedJSON), string(actualJSON), path)
	}
	return nil
}

// resolvePath traverses the document to find the parent of the target path.
// Returns the parent container and the last key/index.
func resolvePath(doc map[string]interface{}, path string, createMissing bool) (interface{}, string, error) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, "", fmt.Errorf("empty path")
	}

	var current interface{} = doc
	for i := 0; i < len(parts)-1; i++ {
		switch c := current.(type) {
		case map[string]interface{}:
			next, ok := c[parts[i]]
			if !ok {
				if createMissing {
					newMap := make(map[string]interface{})
					c[parts[i]] = newMap
					current = newMap
					continue
				}
				return nil, "", fmt.Errorf("path not found at segment: %s", parts[i])
			}
			current = next
		case []interface{}:
			idx, err := strconv.Atoi(parts[i])
			if err != nil {
				return nil, "", fmt.Errorf("invalid array index: %s", parts[i])
			}
			if idx < 0 || idx >= len(c) {
				return nil, "", fmt.Errorf("array index out of bounds: %d", idx)
			}
			current = c[idx]
		default:
			return nil, "", fmt.Errorf("cannot traverse into non-container at: %s", parts[i])
		}
	}

	return current, parts[len(parts)-1], nil
}

func resolveParentOfPath(doc map[string]interface{}, path string) (map[string]interface{}, string) {
	parts := splitPath(path)
	if len(parts) <= 1 {
		return doc, parts[0]
	}

	parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
	parent, lastKey, err := resolvePath(doc, parentPath, false)
	if err != nil {
		return nil, ""
	}
	parentMap, ok := parent.(map[string]interface{})
	if !ok {
		return nil, ""
	}
	_ = lastKey
	return parentMap, parts[len(parts)-2]
}

func splitPath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	data, _ := json.Marshal(m)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	return result
}
