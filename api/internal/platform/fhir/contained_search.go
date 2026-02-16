package fhir

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ---------------------------------------------------------------------------
// Contained resource search mode
// ---------------------------------------------------------------------------

// ContainedSearchMode controls how contained resources are handled in search.
type ContainedSearchMode string

const (
	ContainedModeNone ContainedSearchMode = "false" // Don't search contained (default)
	ContainedModeTrue ContainedSearchMode = "true"  // Search contained only
	ContainedModeBoth ContainedSearchMode = "both"  // Search both regular and contained
)

// ---------------------------------------------------------------------------
// Configuration types
// ---------------------------------------------------------------------------

// ContainedSearchConfig describes how to search within contained resources
// for a given parent resource type.
type ContainedSearchConfig struct {
	ResourceType    string                            // Parent resource type (e.g., "MedicationRequest")
	ContainedColumn string                            // JSON column storing the resource (typically "resource_json")
	ContainedPath   string                            // JSON path to contained array (typically "contained")
	IndexedFields   map[string]ContainedFieldConfig   // Searchable fields within contained resources
}

// ContainedFieldConfig describes a searchable field within a contained resource.
type ContainedFieldConfig struct {
	FHIRPath   string // FHIR path (e.g., "Medication.code.coding.code")
	JSONPath   string // JSON path within contained resource
	SearchType string // token, string, reference, date, etc.
}

// ContainedSearchResult represents a match within contained resources.
type ContainedSearchResult struct {
	ParentID       string                 // ID of the containing resource
	ContainedIndex int                    // Index in the contained array
	ContainedID    string                 // ID of the contained resource (e.g., "#med1")
	ResourceType   string                 // Type of the contained resource
	Resource       map[string]interface{} // The contained resource itself
}

// ---------------------------------------------------------------------------
// Parsing functions
// ---------------------------------------------------------------------------

// ParseContainedParam parses the _contained query parameter.
// Valid values: "false" (default), "true", "both". Empty defaults to "false".
func ParseContainedParam(value string) (ContainedSearchMode, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "", "false":
		return ContainedModeNone, nil
	case "true":
		return ContainedModeTrue, nil
	case "both":
		return ContainedModeBoth, nil
	default:
		return "", fmt.Errorf("invalid _contained value: %q (must be false, true, or both)", value)
	}
}

// ParseContainedTypeParam parses the _containedType query parameter.
// Value is a comma-separated list of resource types. Empty returns an empty slice.
func ParseContainedTypeParam(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}, nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// SQL clause generation — contained search
// ---------------------------------------------------------------------------

// ContainedSearchClause generates SQL for searching within contained resources
// using PostgreSQL jsonb_array_elements and JSON path queries.
// Returns the SQL clause and bind arguments. When mode is ContainedModeNone,
// returns a no-op clause ("1=1").
func ContainedSearchClause(config *ContainedSearchConfig, searchParam string, searchValue string, mode ContainedSearchMode, startIdx int) (string, []interface{}) {
	if mode == ContainedModeNone {
		return "1=1", nil
	}

	field, ok := config.IndexedFields[searchParam]
	if !ok {
		return "1=0", nil
	}

	switch field.SearchType {
	case "token":
		system, code := splitContainedTokenValue(searchValue)
		return ContainedTokenSearchClause(config, field, system, code, startIdx)
	case "string":
		return ContainedStringSearchClause(config, field, searchValue, false, startIdx)
	case "date":
		parsed := ParseSearchValue(searchValue)
		return ContainedDateSearchClause(config, field, string(parsed.Prefix), parsed.Value, startIdx)
	default:
		// Fallback: exact match on extracted text
		return ContainedStringSearchClause(config, field, searchValue, true, startIdx)
	}
}

// splitContainedTokenValue splits a FHIR token value into system and code parts.
func splitContainedTokenValue(value string) (system, code string) {
	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		return parts[0], parts[1]
	}
	return "", value
}

// ContainedTokenSearchClause generates SQL for token search within contained resources.
// It uses jsonb_array_elements to expand the contained array and searches for
// matching system/code within the coding array at the specified JSON path.
func ContainedTokenSearchClause(config *ContainedSearchConfig, field ContainedFieldConfig, system, code string, startIdx int) (string, []interface{}) {
	// Build the EXISTS subquery that unnests the contained array and searches
	// coding elements for matching system/code values.
	//
	// SQL pattern:
	// EXISTS (
	//   SELECT 1 FROM jsonb_array_elements(col->'contained') AS ce,
	//   jsonb_array_elements(ce->'code'->'coding') AS coding
	//   WHERE coding->>'system' = $N AND coding->>'code' = $N+1
	// )
	pathParts := strings.Split(field.JSONPath, ".")
	jsonPathExpr := buildJSONPathNav(pathParts)

	var conditions []string
	var args []interface{}
	idx := startIdx

	if system != "" && code != "" {
		conditions = append(conditions, fmt.Sprintf("coding->>'system' = $%d", idx))
		args = append(args, system)
		idx++
		conditions = append(conditions, fmt.Sprintf("coding->>'code' = $%d", idx))
		args = append(args, code)
	} else if system != "" {
		conditions = append(conditions, fmt.Sprintf("coding->>'system' = $%d", idx))
		args = append(args, system)
	} else if code != "" {
		conditions = append(conditions, fmt.Sprintf("coding->>'code' = $%d", idx))
		args = append(args, code)
	}

	whereClause := strings.Join(conditions, " AND ")

	clause := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM jsonb_array_elements(%s->'%s') AS ce, jsonb_array_elements(ce%s) AS coding WHERE %s)",
		config.ContainedColumn,
		config.ContainedPath,
		jsonPathExpr,
		whereClause,
	)

	return clause, args
}

// buildJSONPathNav converts a dot-separated JSON path into chained -> operators.
// For example: ["code", "coding"] becomes "->'code'->'coding'"
func buildJSONPathNav(parts []string) string {
	var sb strings.Builder
	for _, p := range parts {
		sb.WriteString("->'" + p + "'")
	}
	return sb.String()
}

// ContainedStringSearchClause generates SQL for string search within contained
// resources. Uses ILIKE for non-exact matches (prefix match by default) and
// exact equality for exact mode.
func ContainedStringSearchClause(config *ContainedSearchConfig, field ContainedFieldConfig, value string, exact bool, startIdx int) (string, []interface{}) {
	if value == "" {
		return "1=1", nil
	}

	pathParts := strings.Split(field.JSONPath, ".")
	jsonPathExpr := buildJSONPathTextNav(pathParts)

	if exact {
		clause := fmt.Sprintf(
			"EXISTS (SELECT 1 FROM jsonb_array_elements(%s->'%s') AS ce WHERE ce%s = $%d)",
			config.ContainedColumn,
			config.ContainedPath,
			jsonPathExpr,
			startIdx,
		)
		return clause, []interface{}{value}
	}

	clause := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM jsonb_array_elements(%s->'%s') AS ce WHERE ce%s ILIKE $%d)",
		config.ContainedColumn,
		config.ContainedPath,
		jsonPathExpr,
		startIdx,
	)
	return clause, []interface{}{value + "%"}
}

// buildJSONPathTextNav converts a dot-separated JSON path into chained -> operators
// with the last segment using ->> (text extraction).
// For example: ["code", "text"] becomes "->'code'->>'text'"
func buildJSONPathTextNav(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, p := range parts {
		if i == len(parts)-1 {
			sb.WriteString("->>'" + p + "'")
		} else {
			sb.WriteString("->'" + p + "'")
		}
	}
	return sb.String()
}

// ContainedDateSearchClause generates SQL for date search within contained
// resources. Extracts the date from the JSONB field as text and compares it
// using the specified prefix operator.
func ContainedDateSearchClause(config *ContainedSearchConfig, field ContainedFieldConfig, prefix, value string, startIdx int) (string, []interface{}) {
	pathParts := strings.Split(field.JSONPath, ".")
	jsonPathExpr := buildJSONPathTextNav(pathParts)

	var op string
	switch SearchPrefix(prefix) {
	case PrefixGt, PrefixSa:
		op = ">"
	case PrefixLt, PrefixEb:
		op = "<"
	case PrefixGe:
		op = ">="
	case PrefixLe:
		op = "<="
	case PrefixNe:
		op = "!="
	default: // eq
		op = "="
	}

	clause := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM jsonb_array_elements(%s->'%s') AS ce WHERE ce%s %s $%d)",
		config.ContainedColumn,
		config.ContainedPath,
		jsonPathExpr,
		op,
		startIdx,
	)
	return clause, []interface{}{value}
}

// ContainedTypeFilterClause generates SQL to filter contained resources by
// resource type. Produces an EXISTS clause checking the resourceType field
// within each contained element.
func ContainedTypeFilterClause(config *ContainedSearchConfig, resourceTypes []string, startIdx int) (string, []interface{}) {
	if len(resourceTypes) == 0 {
		return "1=1", nil
	}

	if len(resourceTypes) == 1 {
		clause := fmt.Sprintf(
			"EXISTS (SELECT 1 FROM jsonb_array_elements(%s->'%s') AS ce WHERE ce->>'resourceType' = $%d)",
			config.ContainedColumn,
			config.ContainedPath,
			startIdx,
		)
		return clause, []interface{}{resourceTypes[0]}
	}

	// Multiple types: use IN with positional params
	placeholders := make([]string, len(resourceTypes))
	args := make([]interface{}, len(resourceTypes))
	for i, rt := range resourceTypes {
		placeholders[i] = fmt.Sprintf("$%d", startIdx+i)
		args[i] = rt
	}

	clause := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM jsonb_array_elements(%s->'%s') AS ce WHERE ce->>'resourceType' IN (%s))",
		config.ContainedColumn,
		config.ContainedPath,
		strings.Join(placeholders, ", "),
	)
	return clause, args
}

// ---------------------------------------------------------------------------
// Resource extraction and resolution
// ---------------------------------------------------------------------------

// ExtractContainedResources extracts all contained resources from a FHIR resource.
// Returns an empty slice if no contained resources exist or the contained field
// is not a valid array.
func ExtractContainedResources(resource map[string]interface{}) []map[string]interface{} {
	contained, ok := resource["contained"]
	if !ok {
		return nil
	}

	arr, ok := contained.([]interface{})
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

// ResolveContainedReference resolves a local reference (#id) to a contained
// resource within the parent. The localRef must start with '#'.
func ResolveContainedReference(resource map[string]interface{}, localRef string) (map[string]interface{}, error) {
	if localRef == "" {
		return nil, fmt.Errorf("empty reference")
	}
	if !strings.HasPrefix(localRef, "#") {
		return nil, fmt.Errorf("local reference must start with '#', got %q", localRef)
	}

	targetID := localRef[1:] // strip '#'

	contained := ExtractContainedResources(resource)
	if len(contained) == 0 {
		return nil, fmt.Errorf("resource has no contained resources")
	}

	for _, c := range contained {
		id, _ := c["id"].(string)
		// Match against the raw id or the id with # prefix stripped
		cleanID := strings.TrimPrefix(id, "#")
		if cleanID == targetID {
			return c, nil
		}
	}

	return nil, fmt.Errorf("contained resource %q not found", localRef)
}

// IndexContainedResources creates a lookup map of contained resources keyed
// by their local ID (without '#' prefix). Resources without an id are skipped.
func IndexContainedResources(resource map[string]interface{}) map[string]map[string]interface{} {
	contained := ExtractContainedResources(resource)
	index := make(map[string]map[string]interface{}, len(contained))

	for _, c := range contained {
		id, ok := c["id"].(string)
		if !ok || id == "" {
			continue
		}
		// Strip '#' prefix if present for consistent keying
		cleanID := strings.TrimPrefix(id, "#")
		index[cleanID] = c
	}

	return index
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// ValidateContainedResources validates that contained resources follow FHIR rules:
//   - Must have an id field
//   - Cannot themselves contain resources (no nesting)
//   - Must be referenced by the parent resource
func ValidateContainedResources(resource map[string]interface{}) []ValidationIssue {
	contained := ExtractContainedResources(resource)
	if len(contained) == 0 {
		return nil
	}

	var issues []ValidationIssue

	// Collect all local references from the parent resource
	refs := collectLocalReferences(resource)

	for i, c := range contained {
		location := fmt.Sprintf("contained[%d]", i)

		// Check for id
		id, hasID := c["id"].(string)
		if !hasID || id == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    location,
				Diagnostics: "contained resource must have an id",
			})
		}

		// Check for nested contained (not allowed)
		if nested, ok := c["contained"]; ok {
			if arr, ok := nested.([]interface{}); ok && len(arr) > 0 {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeBusinessRule,
					Location:    location,
					Diagnostics: "contained resources must not nest other contained resources",
				})
			}
		}

		// Check that the contained resource is referenced
		if hasID && id != "" {
			cleanID := strings.TrimPrefix(id, "#")
			refID := "#" + cleanID
			if !refs[refID] {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityWarning,
					Code:        VIssueTypeBusinessRule,
					Location:    location,
					Diagnostics: fmt.Sprintf("contained resource %q is not referenced by the parent resource", refID),
				})
			}
		}
	}

	return issues
}

// collectLocalReferences walks the resource (excluding the "contained" key)
// and collects all local references (#id) found in "reference" fields.
func collectLocalReferences(resource map[string]interface{}) map[string]bool {
	refs := make(map[string]bool)
	walkForReferences(resource, refs, true)
	return refs
}

// walkForReferences recursively walks JSON data collecting local references.
// skipContained controls whether the "contained" key is skipped (true at the
// top level to avoid finding self-references).
func walkForReferences(data interface{}, refs map[string]bool, skipContained bool) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if skipContained && key == "contained" {
				continue
			}
			if key == "reference" {
				if s, ok := val.(string); ok && strings.HasPrefix(s, "#") {
					refs[s] = true
				}
			}
			walkForReferences(val, refs, false)
		}
	case []interface{}:
		for _, item := range v {
			walkForReferences(item, refs, false)
		}
	}
}

// ---------------------------------------------------------------------------
// Default configurations
// ---------------------------------------------------------------------------

// DefaultContainedSearchConfigs returns search configs for common resource
// types that frequently use contained resources.
func DefaultContainedSearchConfigs() map[string]*ContainedSearchConfig {
	return map[string]*ContainedSearchConfig{
		"MedicationRequest": {
			ResourceType:    "MedicationRequest",
			ContainedColumn: "resource_json",
			ContainedPath:   "contained",
			IndexedFields: map[string]ContainedFieldConfig{
				"code": {
					FHIRPath:   "Medication.code.coding.code",
					JSONPath:   "code.coding",
					SearchType: "token",
				},
				"form": {
					FHIRPath:   "Medication.form.coding.code",
					JSONPath:   "form.coding",
					SearchType: "token",
				},
			},
		},
		"MedicationAdministration": {
			ResourceType:    "MedicationAdministration",
			ContainedColumn: "resource_json",
			ContainedPath:   "contained",
			IndexedFields: map[string]ContainedFieldConfig{
				"code": {
					FHIRPath:   "Medication.code.coding.code",
					JSONPath:   "code.coding",
					SearchType: "token",
				},
			},
		},
		"MedicationDispense": {
			ResourceType:    "MedicationDispense",
			ContainedColumn: "resource_json",
			ContainedPath:   "contained",
			IndexedFields: map[string]ContainedFieldConfig{
				"code": {
					FHIRPath:   "Medication.code.coding.code",
					JSONPath:   "code.coding",
					SearchType: "token",
				},
			},
		},
		"Observation": {
			ResourceType:    "Observation",
			ContainedColumn: "resource_json",
			ContainedPath:   "contained",
			IndexedFields: map[string]ContainedFieldConfig{
				"device-name": {
					FHIRPath:   "Device.deviceName.name",
					JSONPath:   "deviceName.name",
					SearchType: "string",
				},
				"device-type": {
					FHIRPath:   "Device.type.coding.code",
					JSONPath:   "type.coding",
					SearchType: "token",
				},
			},
		},
		"DiagnosticReport": {
			ResourceType:    "DiagnosticReport",
			ContainedColumn: "resource_json",
			ContainedPath:   "contained",
			IndexedFields: map[string]ContainedFieldConfig{
				"observation-code": {
					FHIRPath:   "Observation.code.coding.code",
					JSONPath:   "code.coding",
					SearchType: "token",
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// ContainedSearchEngine
// ---------------------------------------------------------------------------

// ContainedSearchEngine manages contained resource search across resource types.
type ContainedSearchEngine struct {
	Configs map[string]*ContainedSearchConfig
}

// NewContainedSearchEngine creates an engine initialized with default configs.
func NewContainedSearchEngine() *ContainedSearchEngine {
	return &ContainedSearchEngine{
		Configs: DefaultContainedSearchConfigs(),
	}
}

// ApplyContainedSearch reads _contained and _containedType parameters from the
// URL values and adds appropriate search criteria to the SearchQuery.
func (e *ContainedSearchEngine) ApplyContainedSearch(q *SearchQuery, params url.Values) error {
	containedVal := params.Get("_contained")
	mode, err := ParseContainedParam(containedVal)
	if err != nil {
		return err
	}

	if mode == ContainedModeNone {
		return nil
	}

	// Parse _containedType filter
	typeVal := params.Get("_containedType")
	resourceTypes, err := ParseContainedTypeParam(typeVal)
	if err != nil {
		return err
	}

	// Apply type filter if specified
	if len(resourceTypes) > 0 {
		// Use a generic config for type filtering
		genericConfig := &ContainedSearchConfig{
			ContainedColumn: "resource_json",
			ContainedPath:   "contained",
		}
		clause, args := ContainedTypeFilterClause(genericConfig, resourceTypes, q.Idx())
		if clause != "1=1" {
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx += len(args)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// JSON building and manipulation
// ---------------------------------------------------------------------------

// BuildContainedJSON generates the JSON bytes for a contained resource array.
// Returns "[]" for nil or empty input.
func BuildContainedJSON(resources []map[string]interface{}) ([]byte, error) {
	if resources == nil {
		resources = []map[string]interface{}{}
	}
	return json.Marshal(resources)
}

// StripContainedResources returns a shallow copy of the resource with the
// "contained" key removed. The original resource is not modified.
func StripContainedResources(resource map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(resource))
	for k, v := range resource {
		if k == "contained" {
			continue
		}
		result[k] = v
	}
	return result
}

// MergeContainedIntoParent creates a copy of the parent resource with the
// contained resources embedded. If contained is nil or empty, returns a copy
// without a "contained" key.
func MergeContainedIntoParent(parent map[string]interface{}, contained []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(parent)+1)
	for k, v := range parent {
		if k == "contained" {
			continue
		}
		result[k] = v
	}
	if len(contained) > 0 {
		result["contained"] = contained
	}
	return result
}
