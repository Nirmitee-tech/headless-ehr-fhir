package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Types
// ===========================================================================

// SearchParamExpression defines a search parameter with a FHIRPath expression.
// It extends the static SearchParameter system to support user-defined search
// parameters that can be dynamically registered and evaluated against resources.
type SearchParamExpression struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`           // token, string, date, reference, quantity, number, uri, composite
	Expression    string   `json:"expression"`      // FHIRPath expression
	ResourceTypes []string `json:"resourceTypes"`   // Which resource types this applies to
	Description   string   `json:"description,omitempty"`
	XPath         string   `json:"xpath,omitempty"` // Legacy XPath (informational only)
	Target        []string `json:"target,omitempty"`
	Modifier      []string `json:"modifier,omitempty"`
	Comparator    []string `json:"comparator,omitempty"`
	MultipleOr    bool     `json:"multipleOr,omitempty"`
	MultipleAnd   bool     `json:"multipleAnd,omitempty"`
}

// SearchIndexValue represents an extracted, indexed value from a resource.
type SearchIndexValue struct {
	ParamName      string
	ResourceType   string
	ResourceID     string
	ValueType      string // string, token, date, reference, quantity, number, uri
	StringValue    *string
	TokenSystem    *string
	TokenCode      *string
	DateValue      *time.Time
	NumberValue    *float64
	QuantityValue  *float64
	QuantityUnit   *string
	ReferenceValue *string
	URIValue       *string
}

// SearchExpressionRegistry manages dynamic search parameter definitions.
type SearchExpressionRegistry struct {
	mu          sync.RWMutex
	expressions map[string]map[string]*SearchParamExpression // resourceType -> paramName -> expression
}

// SearchExpressionIndex stores extracted index values for fast search.
type SearchExpressionIndex struct {
	mu     sync.RWMutex
	values map[string][]SearchIndexValue // "resourceType/resourceID" -> values
}

// ExpressionEvaluator evaluates FHIRPath expressions against resources using
// a simplified subset evaluator suitable for search parameter extraction.
type ExpressionEvaluator struct {
	engine *FHIRPathEngine
}

// ExpressionResult holds the result of evaluating a FHIRPath expression.
type ExpressionResult struct {
	Values []interface{}
	Type   string // string, boolean, integer, decimal, date, coding, reference, etc.
}

// ===========================================================================
// SearchExpressionRegistry
// ===========================================================================

// NewSearchExpressionRegistry creates a new registry.
func NewSearchExpressionRegistry() *SearchExpressionRegistry {
	return &SearchExpressionRegistry{
		expressions: make(map[string]map[string]*SearchParamExpression),
	}
}

// Register adds a search parameter expression to the registry. It registers
// the expression for each resource type specified. Returns an error if an
// expression with the same name is already registered for any resource type.
func (r *SearchExpressionRegistry) Register(expr *SearchParamExpression) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicates.
	for _, rt := range expr.ResourceTypes {
		if byName, ok := r.expressions[rt]; ok {
			if _, exists := byName[expr.Name]; exists {
				return fmt.Errorf("expression %q already registered for resource type %q", expr.Name, rt)
			}
		}
	}

	// Register for each resource type.
	for _, rt := range expr.ResourceTypes {
		if r.expressions[rt] == nil {
			r.expressions[rt] = make(map[string]*SearchParamExpression)
		}
		// Store a copy.
		stored := *expr
		r.expressions[rt][expr.Name] = &stored
	}
	return nil
}

// Unregister removes a search parameter expression from the registry.
func (r *SearchExpressionRegistry) Unregister(resourceType, paramName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	byName, ok := r.expressions[resourceType]
	if !ok {
		return fmt.Errorf("no expressions registered for resource type %q", resourceType)
	}
	if _, exists := byName[paramName]; !exists {
		return fmt.Errorf("expression %q not found for resource type %q", paramName, resourceType)
	}
	delete(byName, paramName)
	if len(byName) == 0 {
		delete(r.expressions, resourceType)
	}
	return nil
}

// Get returns an expression by resource type and param name.
func (r *SearchExpressionRegistry) Get(resourceType, paramName string) (*SearchParamExpression, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	byName, ok := r.expressions[resourceType]
	if !ok {
		return nil, false
	}
	expr, exists := byName[paramName]
	if !exists {
		return nil, false
	}
	// Return a copy.
	result := *expr
	return &result, true
}

// ListForResourceType returns all expressions for a resource type, sorted by name.
func (r *SearchExpressionRegistry) ListForResourceType(resourceType string) []*SearchParamExpression {
	r.mu.RLock()
	defer r.mu.RUnlock()

	byName, ok := r.expressions[resourceType]
	if !ok {
		return nil
	}

	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]*SearchParamExpression, 0, len(byName))
	for _, name := range names {
		cp := *byName[name]
		result = append(result, &cp)
	}
	return result
}

// listAll returns all expressions across all resource types. Used for handler listing.
func (r *SearchExpressionRegistry) listAll() []*SearchParamExpression {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var result []*SearchParamExpression
	// Sorted by resource type for deterministic output.
	rtKeys := make([]string, 0, len(r.expressions))
	for rt := range r.expressions {
		rtKeys = append(rtKeys, rt)
	}
	sort.Strings(rtKeys)

	for _, rt := range rtKeys {
		byName := r.expressions[rt]
		names := make([]string, 0, len(byName))
		for name := range byName {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			key := rt + "/" + name
			if seen[key] {
				continue
			}
			seen[key] = true
			cp := *byName[name]
			result = append(result, &cp)
		}
	}
	return result
}

// ===========================================================================
// Validation
// ===========================================================================

// validExprParamTypes enumerates the allowed search parameter type values.
var validExprParamTypes = map[string]bool{
	"number":    true,
	"date":      true,
	"string":    true,
	"token":     true,
	"reference": true,
	"composite": true,
	"quantity":  true,
	"uri":       true,
}

// ValidateSearchParamExpression validates an expression definition and returns
// any issues found. Issues with SeverityError indicate the expression is invalid.
func ValidateSearchParamExpression(expr *SearchParamExpression) []ValidationIssue {
	var issues []ValidationIssue

	if expr.Name == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "SearchParamExpression.name",
			Diagnostics: "name is required",
		})
	}

	if expr.Expression == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "SearchParamExpression.expression",
			Diagnostics: "expression is required",
		})
	}

	if expr.Type == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "SearchParamExpression.type",
			Diagnostics: "type is required",
		})
	} else if !validExprParamTypes[expr.Type] {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeValue,
			Location:    "SearchParamExpression.type",
			Diagnostics: fmt.Sprintf("type must be one of: number, date, string, token, reference, composite, quantity, uri; got %q", expr.Type),
		})
	}

	if len(expr.ResourceTypes) == 0 {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "SearchParamExpression.resourceTypes",
			Diagnostics: "at least one resource type is required",
		})
	}

	return issues
}

// ===========================================================================
// ExpressionEvaluator
// ===========================================================================

// NewExpressionEvaluator creates a new evaluator backed by the FHIRPathEngine.
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{
		engine: NewFHIRPathEngine(),
	}
}

// Evaluate evaluates a FHIRPath expression against a resource and returns
// the result wrapped in an ExpressionResult.
func (e *ExpressionEvaluator) Evaluate(expression string, resource map[string]interface{}) (*ExpressionResult, error) {
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}
	if resource == nil {
		return &ExpressionResult{Values: []interface{}{}}, nil
	}

	values, err := e.engine.Evaluate(resource, expression)
	if err != nil {
		return nil, fmt.Errorf("expression evaluation failed: %w", err)
	}

	// Flatten any nested arrays.
	flat := flattenValues(values)

	resultType := inferType(flat)
	return &ExpressionResult{
		Values: flat,
		Type:   resultType,
	}, nil
}

// flattenValues flattens nested slices in a result collection.
func flattenValues(values []interface{}) []interface{} {
	var out []interface{}
	for _, v := range values {
		if arr, ok := v.([]interface{}); ok {
			out = append(out, flattenValues(arr)...)
		} else {
			out = append(out, v)
		}
	}
	return out
}

// inferType infers the FHIRPath type of a result collection.
func inferType(values []interface{}) string {
	if len(values) == 0 {
		return ""
	}
	v := values[0]
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case int, int64:
		return "integer"
	case float64:
		return "decimal"
	case time.Time:
		return "date"
	case map[string]interface{}:
		m := v.(map[string]interface{})
		if _, ok := m["system"]; ok {
			if _, ok2 := m["code"]; ok2 {
				return "coding"
			}
		}
		if _, ok := m["reference"]; ok {
			return "reference"
		}
		return "object"
	default:
		return "unknown"
	}
}

// ===========================================================================
// Value extraction
// ===========================================================================

// ExtractSearchValues extracts indexed values from a resource based on a
// search parameter expression definition.
func ExtractSearchValues(expr *SearchParamExpression, resource map[string]interface{}, evaluator *ExpressionEvaluator) ([]SearchIndexValue, error) {
	if resource == nil {
		return nil, nil
	}

	resourceID, _ := resource["id"].(string)
	resourceType, _ := resource["resourceType"].(string)

	result, err := evaluator.Evaluate(expr.Expression, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression %q: %w", expr.Expression, err)
	}

	if len(result.Values) == 0 {
		return nil, nil
	}

	var values []SearchIndexValue
	for _, val := range result.Values {
		sv, err := ConvertToSearchValue(expr.Name, resourceType, resourceID, val, expr.Type)
		if err != nil {
			continue // skip values that can't be converted
		}
		if sv != nil {
			values = append(values, *sv)
		}
	}
	return values, nil
}

// ConvertToSearchValue converts an evaluated result to a SearchIndexValue.
func ConvertToSearchValue(paramName, resourceType, resourceID string, value interface{}, valueType string) (*SearchIndexValue, error) {
	sv := &SearchIndexValue{
		ParamName:    paramName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ValueType:    valueType,
	}

	switch valueType {
	case "string":
		s := fmt.Sprintf("%v", value)
		sv.StringValue = &s

	case "token":
		switch v := value.(type) {
		case map[string]interface{}:
			if sys, ok := v["system"].(string); ok {
				sv.TokenSystem = &sys
			}
			if code, ok := v["code"].(string); ok {
				sv.TokenCode = &code
			}
		case string:
			sv.TokenCode = &v
		default:
			s := fmt.Sprintf("%v", value)
			sv.TokenCode = &s
		}

	case "date":
		switch v := value.(type) {
		case time.Time:
			sv.DateValue = &v
		case string:
			t, err := parseFlexDate(v)
			if err != nil {
				return nil, fmt.Errorf("cannot parse date value %q: %w", v, err)
			}
			sv.DateValue = &t
		default:
			return nil, fmt.Errorf("unexpected date value type: %T", value)
		}

	case "number":
		switch v := value.(type) {
		case float64:
			sv.NumberValue = &v
		case int:
			f := float64(v)
			sv.NumberValue = &f
		case int64:
			f := float64(v)
			sv.NumberValue = &f
		case string:
			// Keep raw string; caller handles parsing
			sv.StringValue = &v
		default:
			return nil, fmt.Errorf("unexpected number value type: %T", value)
		}

	case "quantity":
		switch v := value.(type) {
		case map[string]interface{}:
			if val, ok := v["value"].(float64); ok {
				sv.QuantityValue = &val
			}
			if unit, ok := v["unit"].(string); ok {
				sv.QuantityUnit = &unit
			}
		default:
			return nil, fmt.Errorf("unexpected quantity value type: %T", value)
		}

	case "reference":
		switch v := value.(type) {
		case map[string]interface{}:
			if ref, ok := v["reference"].(string); ok {
				sv.ReferenceValue = &ref
			}
		case string:
			sv.ReferenceValue = &v
		default:
			return nil, fmt.Errorf("unexpected reference value type: %T", value)
		}

	case "uri":
		s := fmt.Sprintf("%v", value)
		sv.URIValue = &s

	default:
		s := fmt.Sprintf("%v", value)
		sv.StringValue = &s
	}

	return sv, nil
}

// ===========================================================================
// SearchExpressionIndex
// ===========================================================================

// NewSearchExpressionIndex creates a new in-memory index.
func NewSearchExpressionIndex() *SearchExpressionIndex {
	return &SearchExpressionIndex{
		values: make(map[string][]SearchIndexValue),
	}
}

// indexKey returns the key for a resource in the index.
func indexKey(resourceType, resourceID string) string {
	return resourceType + "/" + resourceID
}

// Index stores extracted values for a resource, replacing any previous values.
func (idx *SearchExpressionIndex) Index(resourceType, resourceID string, values []SearchIndexValue) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	key := indexKey(resourceType, resourceID)
	idx.values[key] = values
}

// Remove removes index entries for a resource.
func (idx *SearchExpressionIndex) Remove(resourceType, resourceID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	key := indexKey(resourceType, resourceID)
	delete(idx.values, key)
}

// Search searches the index for matching resources, returning matching resource IDs.
func (idx *SearchExpressionIndex) Search(resourceType, paramName string, operator string, value interface{}) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	seen := make(map[string]bool)
	var results []string

	for key, svList := range idx.values {
		if !strings.HasPrefix(key, resourceType+"/") {
			continue
		}
		for _, sv := range svList {
			if sv.ParamName != paramName {
				continue
			}
			if matchIndexValue(sv, operator, value) {
				if !seen[sv.ResourceID] {
					seen[sv.ResourceID] = true
					results = append(results, sv.ResourceID)
				}
			}
		}
	}

	sort.Strings(results)
	return results
}

// matchIndexValue checks whether an indexed value matches the search criteria.
func matchIndexValue(sv SearchIndexValue, operator string, value interface{}) bool {
	switch sv.ValueType {
	case "string":
		if sv.StringValue == nil {
			return false
		}
		target := fmt.Sprintf("%v", value)
		return compareString(*sv.StringValue, target, operator)

	case "token":
		target := fmt.Sprintf("%v", value)
		return matchToken(sv, target, operator)

	case "number":
		if sv.NumberValue == nil {
			return false
		}
		targetNum, ok := toFloat64Ok(value)
		if !ok {
			return false
		}
		return compareFloat(*sv.NumberValue, targetNum, operator)

	case "date":
		if sv.DateValue == nil {
			return false
		}
		switch tv := value.(type) {
		case time.Time:
			return compareDate(*sv.DateValue, tv, operator)
		case string:
			t, err := parseFlexDate(tv)
			if err != nil {
				return false
			}
			return compareDate(*sv.DateValue, t, operator)
		}
		return false

	case "reference":
		if sv.ReferenceValue == nil {
			return false
		}
		target := fmt.Sprintf("%v", value)
		return compareString(*sv.ReferenceValue, target, operator)

	case "uri":
		if sv.URIValue == nil {
			return false
		}
		target := fmt.Sprintf("%v", value)
		return compareString(*sv.URIValue, target, operator)

	case "quantity":
		if sv.QuantityValue == nil {
			return false
		}
		targetNum, ok := toFloat64Ok(value)
		if !ok {
			return false
		}
		return compareFloat(*sv.QuantityValue, targetNum, operator)
	}
	return false
}

// matchToken matches a token search value against a target string.
// Supports "system|code", "|code", "system|", or just "code".
func matchToken(sv SearchIndexValue, target string, operator string) bool {
	if strings.Contains(target, "|") {
		parts := strings.SplitN(target, "|", 2)
		sys := parts[0]
		code := parts[1]

		sysMatch := true
		codeMatch := true

		if sys != "" {
			if sv.TokenSystem == nil || *sv.TokenSystem != sys {
				sysMatch = false
			}
		}
		if code != "" {
			if sv.TokenCode == nil || *sv.TokenCode != code {
				codeMatch = false
			}
		}

		matched := sysMatch && codeMatch
		if operator == "ne" {
			return !matched
		}
		return matched
	}

	// Just code matching.
	if sv.TokenCode == nil {
		if operator == "ne" {
			return true
		}
		return false
	}
	matched := *sv.TokenCode == target
	if operator == "ne" {
		return !matched
	}
	return matched
}

// compareString compares two strings using the given operator.
func compareString(a, b string, operator string) bool {
	switch operator {
	case "eq":
		return a == b
	case "ne":
		return a != b
	case "gt":
		return a > b
	case "lt":
		return a < b
	case "ge":
		return a >= b
	case "le":
		return a <= b
	default:
		return a == b
	}
}

// compareFloat compares two float64 values using the given operator.
func compareFloat(a, b float64, operator string) bool {
	switch operator {
	case "eq":
		return a == b
	case "ne":
		return a != b
	case "gt":
		return a > b
	case "lt":
		return a < b
	case "ge":
		return a >= b
	case "le":
		return a <= b
	default:
		return a == b
	}
}

// compareDate compares two time.Time values using the given operator.
func compareDate(a, b time.Time, operator string) bool {
	switch operator {
	case "eq":
		return a.Equal(b)
	case "ne":
		return !a.Equal(b)
	case "gt":
		return a.After(b)
	case "lt":
		return a.Before(b)
	case "ge":
		return a.Equal(b) || a.After(b)
	case "le":
		return a.Equal(b) || a.Before(b)
	default:
		return a.Equal(b)
	}
}

// toFloat64Ok attempts to convert an interface value to float64,
// returning a boolean indicating success. This wraps the package-level
// toFloat64 with an additional ok flag for types that cannot convert.
func toFloat64Ok(v interface{}) (float64, bool) {
	switch v.(type) {
	case float64, float32, int, int64, int32:
		return toFloat64(v), true
	}
	return 0, false
}

// ===========================================================================
// SQL generation
// ===========================================================================

// expressionToJSONBPath converts a FHIRPath expression to a JSONB extraction path.
// For example, "Patient.name.family" -> "resource->'name'->>'family'"
func expressionToJSONBPath(expression string) (string, string) {
	parts := strings.Split(expression, ".")
	if len(parts) <= 1 {
		return "resource", ""
	}

	// Skip the resource type prefix.
	fields := parts[1:]

	// Strip .where() and other function calls from the path.
	var cleanFields []string
	for _, f := range fields {
		if strings.Contains(f, "(") {
			break
		}
		cleanFields = append(cleanFields, f)
	}

	if len(cleanFields) == 0 {
		return "resource", ""
	}

	// Build JSONB path: resource->'field1'->'field2'->>'lastField'
	if len(cleanFields) == 1 {
		return fmt.Sprintf("resource->>'%s'", cleanFields[0]), cleanFields[len(cleanFields)-1]
	}

	var path strings.Builder
	path.WriteString("resource")
	for i, f := range cleanFields {
		if i == len(cleanFields)-1 {
			path.WriteString(fmt.Sprintf("->>'%s'", f))
		} else {
			path.WriteString(fmt.Sprintf("->'%s'", f))
		}
	}
	return path.String(), cleanFields[len(cleanFields)-1]
}

// GenerateSearchSQL generates SQL for a dynamic search parameter query.
// It produces a WHERE clause fragment using JSONB operators.
func GenerateSearchSQL(expr *SearchParamExpression, operator, value string, startIdx int) (string, []interface{}, error) {
	jsonbPath, _ := expressionToJSONBPath(expr.Expression)
	idx := startIdx

	switch expr.Type {
	case "string":
		switch operator {
		case "exact":
			return fmt.Sprintf("%s = $%d", jsonbPath, idx), []interface{}{value}, nil
		case "contains":
			return fmt.Sprintf("%s ILIKE $%d", jsonbPath, idx), []interface{}{"%" + value + "%"}, nil
		default: // eq or default prefix match
			return fmt.Sprintf("%s ILIKE $%d", jsonbPath, idx), []interface{}{value + "%"}, nil
		}

	case "token":
		if strings.Contains(value, "|") {
			parts := strings.SplitN(value, "|", 2)
			sys := parts[0]
			code := parts[1]
			if sys != "" && code != "" {
				// Match both system and code in JSONB.
				clause := fmt.Sprintf("resource @> $%d::jsonb", idx)
				jsonVal := buildTokenJSONB(expr.Expression, sys, code)
				return clause, []interface{}{jsonVal}, nil
			} else if code != "" {
				clause := fmt.Sprintf("%s = $%d", jsonbPath, idx)
				return clause, []interface{}{code}, nil
			} else if sys != "" {
				clause := fmt.Sprintf("%s = $%d", jsonbPath, idx)
				return clause, []interface{}{sys}, nil
			}
		}
		return fmt.Sprintf("%s = $%d", jsonbPath, idx), []interface{}{value}, nil

	case "date":
		switch operator {
		case "gt", "sa":
			return fmt.Sprintf("(%s)::timestamp > $%d", jsonbPath, idx), []interface{}{value}, nil
		case "lt", "eb":
			return fmt.Sprintf("(%s)::timestamp < $%d", jsonbPath, idx), []interface{}{value}, nil
		case "ge":
			return fmt.Sprintf("(%s)::timestamp >= $%d", jsonbPath, idx), []interface{}{value}, nil
		case "le":
			return fmt.Sprintf("(%s)::timestamp <= $%d", jsonbPath, idx), []interface{}{value}, nil
		case "ne":
			return fmt.Sprintf("(%s)::timestamp != $%d", jsonbPath, idx), []interface{}{value}, nil
		default:
			return fmt.Sprintf("(%s)::timestamp = $%d", jsonbPath, idx), []interface{}{value}, nil
		}

	case "number":
		castPath := fmt.Sprintf("(%s)::numeric", jsonbPath)
		switch operator {
		case "gt", "sa":
			return fmt.Sprintf("%s > $%d", castPath, idx), []interface{}{value}, nil
		case "lt", "eb":
			return fmt.Sprintf("%s < $%d", castPath, idx), []interface{}{value}, nil
		case "ge":
			return fmt.Sprintf("%s >= $%d", castPath, idx), []interface{}{value}, nil
		case "le":
			return fmt.Sprintf("%s <= $%d", castPath, idx), []interface{}{value}, nil
		case "ne":
			return fmt.Sprintf("%s != $%d", castPath, idx), []interface{}{value}, nil
		default:
			return fmt.Sprintf("%s = $%d", castPath, idx), []interface{}{value}, nil
		}

	case "quantity":
		// Quantity searches match on value; unit matching is done separately.
		valuePath := fmt.Sprintf("(resource->'%s'->>'value')::numeric", extractFieldName(expr.Expression))
		switch operator {
		case "gt", "sa":
			return fmt.Sprintf("%s > $%d", valuePath, idx), []interface{}{value}, nil
		case "lt", "eb":
			return fmt.Sprintf("%s < $%d", valuePath, idx), []interface{}{value}, nil
		case "ge":
			return fmt.Sprintf("%s >= $%d", valuePath, idx), []interface{}{value}, nil
		case "le":
			return fmt.Sprintf("%s <= $%d", valuePath, idx), []interface{}{value}, nil
		case "ne":
			return fmt.Sprintf("%s != $%d", valuePath, idx), []interface{}{value}, nil
		default:
			return fmt.Sprintf("%s = $%d", valuePath, idx), []interface{}{value}, nil
		}

	case "reference":
		return fmt.Sprintf("%s = $%d", jsonbPath, idx), []interface{}{value}, nil

	case "uri":
		return fmt.Sprintf("%s = $%d", jsonbPath, idx), []interface{}{value}, nil

	default:
		return fmt.Sprintf("%s = $%d", jsonbPath, idx), []interface{}{value}, nil
	}
}

// buildTokenJSONB builds a JSON structure for JSONB @> matching on token values.
func buildTokenJSONB(expression, system, code string) string {
	parts := strings.Split(expression, ".")
	if len(parts) <= 1 {
		return fmt.Sprintf(`{"system":"%s","code":"%s"}`, system, code)
	}

	// Build nested JSON from expression path, skipping resource type.
	fields := parts[1:]
	inner := fmt.Sprintf(`{"system":"%s","code":"%s"}`, system, code)

	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		if strings.Contains(field, "(") {
			continue
		}
		inner = fmt.Sprintf(`{"%s":[%s]}`, field, inner)
	}
	return inner
}

// extractFieldName extracts the last simple field name from an expression.
func extractFieldName(expression string) string {
	parts := strings.Split(expression, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if !strings.Contains(parts[i], "(") {
			return parts[i]
		}
	}
	return ""
}

// ===========================================================================
// Default expressions
// ===========================================================================

// DefaultSearchExpressions returns pre-built expressions for common
// extension-based searches from US Core profiles.
func DefaultSearchExpressions() []*SearchParamExpression {
	return []*SearchParamExpression{
		{
			Name:          "race",
			Type:          "token",
			Expression:    "Patient.extension.where(url='http://hl7.org/fhir/us/core/StructureDefinition/us-core-race').extension.where(url='ombCategory').valueCoding",
			ResourceTypes: []string{"Patient"},
			Description:   "US Core Race extension search",
		},
		{
			Name:          "ethnicity",
			Type:          "token",
			Expression:    "Patient.extension.where(url='http://hl7.org/fhir/us/core/StructureDefinition/us-core-ethnicity').extension.where(url='ombCategory').valueCoding",
			ResourceTypes: []string{"Patient"},
			Description:   "US Core Ethnicity extension search",
		},
		{
			Name:          "birthsex",
			Type:          "token",
			Expression:    "Patient.extension.where(url='http://hl7.org/fhir/us/core/StructureDefinition/us-core-birthsex').valueCode",
			ResourceTypes: []string{"Patient"},
			Description:   "US Core Birth Sex extension search",
		},
	}
}

// ===========================================================================
// ToSearchParameter conversion
// ===========================================================================

// ToSearchParameter converts a SearchParamExpression to a FHIR SearchParameter
// resource representation.
func (expr *SearchParamExpression) ToSearchParameter() map[string]interface{} {
	sp := map[string]interface{}{
		"resourceType": "SearchParameter",
		"name":         expr.Name,
		"code":         expr.Name,
		"status":       "active",
		"type":         expr.Type,
		"expression":   expr.Expression,
		"base":         expr.ResourceTypes,
	}

	if expr.Description != "" {
		sp["description"] = expr.Description
	}
	if expr.XPath != "" {
		sp["xpath"] = expr.XPath
	}
	if len(expr.Target) > 0 {
		sp["target"] = expr.Target
	}
	if len(expr.Modifier) > 0 {
		sp["modifier"] = expr.Modifier
	}
	if len(expr.Comparator) > 0 {
		sp["comparator"] = expr.Comparator
	}
	if expr.MultipleOr {
		sp["multipleOr"] = true
	}
	if expr.MultipleAnd {
		sp["multipleAnd"] = true
	}

	return sp
}

// ===========================================================================
// HTTP handler
// ===========================================================================

// RegisterSearchParamHandler returns an echo.HandlerFunc that manages search
// parameter expressions via HTTP. Supports:
//   - GET: list expressions (optional ?resourceType= filter)
//   - POST: create a new expression
//   - DELETE: remove an expression (?resourceType= and ?name= required)
func RegisterSearchParamHandler(registry *SearchExpressionRegistry) echo.HandlerFunc {
	return func(c echo.Context) error {
		switch c.Request().Method {
		case http.MethodGet:
			return handleSearchExprGet(c, registry)
		case http.MethodPost:
			return handleSearchExprPost(c, registry)
		case http.MethodDelete:
			return handleSearchExprDelete(c, registry)
		default:
			return c.JSON(http.StatusMethodNotAllowed, ErrorOutcome("method not allowed"))
		}
	}
}

func handleSearchExprGet(c echo.Context, registry *SearchExpressionRegistry) error {
	resourceType := c.QueryParam("resourceType")

	var exprs []*SearchParamExpression
	if resourceType != "" {
		exprs = registry.ListForResourceType(resourceType)
	} else {
		exprs = registry.listAll()
	}

	entries := make([]map[string]interface{}, 0, len(exprs))
	for _, expr := range exprs {
		entries = append(entries, map[string]interface{}{
			"resource": expr.ToSearchParameter(),
			"search": map[string]string{
				"mode": "match",
			},
		})
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(entries),
		"entry":        entries,
	}
	return c.JSON(http.StatusOK, bundle)
}

func handleSearchExprPost(c echo.Context, registry *SearchExpressionRegistry) error {
	var expr SearchParamExpression
	if err := json.NewDecoder(c.Request().Body).Decode(&expr); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	issues := ValidateSearchParamExpression(&expr)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return c.JSON(http.StatusBadRequest, ErrorOutcome(issue.Diagnostics))
		}
	}

	if err := registry.Register(&expr); err != nil {
		return c.JSON(http.StatusConflict, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusCreated, expr.ToSearchParameter())
}

func handleSearchExprDelete(c echo.Context, registry *SearchExpressionRegistry) error {
	resourceType := c.QueryParam("resourceType")
	name := c.QueryParam("name")

	if resourceType == "" || name == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType and name query parameters are required"))
	}

	if err := registry.Unregister(resourceType, name); err != nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome(err.Error()))
	}

	return c.NoContent(http.StatusNoContent)
}
