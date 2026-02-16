package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"

	"github.com/labstack/echo/v4"
)

// DiffEntry represents a single difference between two versions of a FHIR resource.
type DiffEntry struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // "added", "removed", "changed"
	OldValue interface{} `json:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue,omitempty"`
}

// DiffResources compares two resource maps recursively and returns all differences.
// It walks through both maps, identifying added, removed, and changed values including
// nested maps and arrays.
func DiffResources(old, new map[string]interface{}) []DiffEntry {
	var diffs []DiffEntry
	diffMaps("", old, new, &diffs)
	return diffs
}

// diffMaps recursively compares two maps and appends differences to diffs.
func diffMaps(prefix string, old, new map[string]interface{}, diffs *[]DiffEntry) {
	// Collect all keys from both maps.
	keys := make(map[string]bool)
	for k := range old {
		keys[k] = true
	}
	for k := range new {
		keys[k] = true
	}

	// Sort keys for deterministic output.
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, key := range sorted {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		oldVal, inOld := old[key]
		newVal, inNew := new[key]

		if !inOld {
			*diffs = append(*diffs, DiffEntry{
				Path:     path,
				Type:     "added",
				NewValue: newVal,
			})
			continue
		}

		if !inNew {
			*diffs = append(*diffs, DiffEntry{
				Path:     path,
				Type:     "removed",
				OldValue: oldVal,
			})
			continue
		}

		// Both exist: compare values.
		diffValues(path, oldVal, newVal, diffs)
	}
}

// diffValues compares two values at the given path.
func diffValues(path string, oldVal, newVal interface{}, diffs *[]DiffEntry) {
	oldMap, oldIsMap := toMap(oldVal)
	newMap, newIsMap := toMap(newVal)

	if oldIsMap && newIsMap {
		diffMaps(path, oldMap, newMap, diffs)
		return
	}

	oldSlice, oldIsSlice := asSlice(oldVal)
	newSlice, newIsSlice := asSlice(newVal)

	if oldIsSlice && newIsSlice {
		diffSlices(path, oldSlice, newSlice, diffs)
		return
	}

	// Scalar comparison.
	if !reflect.DeepEqual(oldVal, newVal) {
		*diffs = append(*diffs, DiffEntry{
			Path:     path,
			Type:     "changed",
			OldValue: oldVal,
			NewValue: newVal,
		})
	}
}

// diffSlices compares two slices element-by-element.
func diffSlices(path string, old, new []interface{}, diffs *[]DiffEntry) {
	maxLen := len(old)
	if len(new) > maxLen {
		maxLen = len(new)
	}

	for i := 0; i < maxLen; i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)

		if i >= len(old) {
			*diffs = append(*diffs, DiffEntry{
				Path:     elemPath,
				Type:     "added",
				NewValue: new[i],
			})
			continue
		}

		if i >= len(new) {
			*diffs = append(*diffs, DiffEntry{
				Path:     elemPath,
				Type:     "removed",
				OldValue: old[i],
			})
			continue
		}

		diffValues(elemPath, old[i], new[i], diffs)
	}
}

// asSlice attempts to cast v to []interface{}, returning a boolean indicating success.
func asSlice(v interface{}) ([]interface{}, bool) {
	s, ok := v.([]interface{})
	return s, ok
}

// DiffToParameters converts a slice of DiffEntry values to a FHIR Parameters resource.
// The resulting map conforms to the FHIR Parameters resource structure with each diff
// represented as a parameter containing parts for path, type, oldValue, and newValue.
func DiffToParameters(diffs []DiffEntry) map[string]interface{} {
	params := make([]interface{}, 0, len(diffs))

	for _, d := range diffs {
		parts := []interface{}{
			map[string]interface{}{"name": "path", "valueString": d.Path},
			map[string]interface{}{"name": "type", "valueString": d.Type},
		}

		if d.OldValue != nil {
			parts = append(parts, map[string]interface{}{
				"name":        "oldValue",
				"valueString": fmt.Sprintf("%v", d.OldValue),
			})
		}

		if d.NewValue != nil {
			parts = append(parts, map[string]interface{}{
				"name":        "newValue",
				"valueString": fmt.Sprintf("%v", d.NewValue),
			})
		}

		params = append(params, map[string]interface{}{
			"name": "diff",
			"part": parts,
		})
	}

	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    params,
	}
}

// DiffHandler returns an echo.HandlerFunc that implements the FHIR $diff operation.
// It compares two versions of a resource identified by resourceType and id, using
// query parameters "from" and "to" to specify version numbers. If "to" is omitted
// it defaults to the latest version by fetching version history.
//
// Route: GET /fhir/:resourceType/:id/$diff?from=1&to=2
func DiffHandler(historyRepo *HistoryRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		resourceType := c.Param("resourceType")
		resourceID := c.Param("id")

		fromStr := c.QueryParam("from")
		if fromStr == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError,
				IssueTypeRequired,
				"query parameter 'from' is required",
			))
		}

		fromVersion, err := strconv.Atoi(fromStr)
		if err != nil || fromVersion < 1 {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError,
				IssueTypeValue,
				"'from' must be a positive integer version number",
			))
		}

		toStr := c.QueryParam("to")
		var toVersion int
		if toStr == "" {
			// Default to the latest version.
			entries, _, err := historyRepo.ListVersions(c.Request().Context(), resourceType, resourceID, 1, 0)
			if err != nil || len(entries) == 0 {
				return c.JSON(http.StatusNotFound, NewOperationOutcome(
					IssueSeverityError,
					IssueTypeNotFound,
					fmt.Sprintf("no versions found for %s/%s", resourceType, resourceID),
				))
			}
			toVersion = entries[0].VersionID
		} else {
			toVersion, err = strconv.Atoi(toStr)
			if err != nil || toVersion < 1 {
				return c.JSON(http.StatusBadRequest, NewOperationOutcome(
					IssueSeverityError,
					IssueTypeValue,
					"'to' must be a positive integer version number",
				))
			}
		}

		if fromVersion == toVersion {
			// Same version: return empty diff.
			return c.JSON(http.StatusOK, DiffToParameters(nil))
		}

		ctx := c.Request().Context()

		fromEntry, err := historyRepo.GetVersion(ctx, resourceType, resourceID, fromVersion)
		if err != nil {
			return c.JSON(http.StatusNotFound, NewOperationOutcome(
				IssueSeverityError,
				IssueTypeNotFound,
				fmt.Sprintf("version %d of %s/%s not found", fromVersion, resourceType, resourceID),
			))
		}

		toEntry, err := historyRepo.GetVersion(ctx, resourceType, resourceID, toVersion)
		if err != nil {
			return c.JSON(http.StatusNotFound, NewOperationOutcome(
				IssueSeverityError,
				IssueTypeNotFound,
				fmt.Sprintf("version %d of %s/%s not found", toVersion, resourceType, resourceID),
			))
		}

		var oldResource, newResource map[string]interface{}
		if err := json.Unmarshal(fromEntry.Resource, &oldResource); err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityFatal,
				IssueTypeException,
				fmt.Sprintf("failed to parse version %d: %v", fromVersion, err),
			))
		}
		if err := json.Unmarshal(toEntry.Resource, &newResource); err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityFatal,
				IssueTypeException,
				fmt.Sprintf("failed to parse version %d: %v", toVersion, err),
			))
		}

		diffs := DiffResources(oldResource, newResource)
		result := DiffToParameters(diffs)

		return c.JSON(http.StatusOK, result)
	}
}
