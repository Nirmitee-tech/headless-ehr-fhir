package fhir

import (
	"fmt"
	"strings"
)

// SearchParamComposite is the SearchParamType for composite search parameters.
// Composite parameters combine two or more existing search parameters into one.
const SearchParamComposite SearchParamType = 100

// CompositeComponent describes one component within a composite search parameter.
// Each component maps to an underlying search parameter type and its database columns.
type CompositeComponent struct {
	Name      string          // Logical name of this component (e.g., "code", "value")
	Type      SearchParamType // The underlying search type (token, date, quantity, string, etc.)
	Column    string          // Primary DB column for this component
	SysColumn string          // System column (for token components with system|code)
}

// CompositeSearchConfig defines the structure of a composite search parameter.
// Components are ordered: the first component matches the value before the first $,
// the second component matches the value between the first and second $, and so on.
type CompositeSearchConfig struct {
	Components []CompositeComponent
}

// CompositeSearchClause generates a SQL WHERE clause for a composite search parameter.
// The value is split by "$" and each part is matched against the corresponding component.
// Components are combined with AND. Returns the combined clause, all arguments, and
// the next available parameter index.
//
// Example: value "http://loinc.org|8480-6$gt5.4" with token+quantity components produces:
//
//	(code_system = $1 AND code_value = $2) AND value_quantity > $3
func CompositeSearchClause(config CompositeSearchConfig, value string, startIdx int) (string, []interface{}, int) {
	parts := strings.Split(value, "$")

	if len(config.Components) == 0 {
		return "1=1", nil, startIdx
	}

	var clauses []string
	var allArgs []interface{}
	idx := startIdx

	// Process each component up to min(len(parts), len(components))
	limit := len(parts)
	if len(config.Components) < limit {
		limit = len(config.Components)
	}

	for i := 0; i < limit; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		comp := config.Components[i]
		var clause string
		var args []interface{}
		var nextIdx int

		switch comp.Type {
		case SearchParamToken:
			if comp.SysColumn != "" {
				clause, args, nextIdx = TokenSearchClause(comp.SysColumn, comp.Column, part, idx)
			} else {
				clause = fmt.Sprintf("%s = $%d", comp.Column, idx)
				args = []interface{}{part}
				nextIdx = idx + 1
			}
		case SearchParamDate:
			clause, args, nextIdx = DateSearchClause(comp.Column, part, idx)
		case SearchParamNumber, SearchParamQuantity:
			// Quantity values may contain unit info: number|system|code
			// Parse the quantity value and use the numeric portion for comparison
			clause, args, nextIdx = quantityComponentClause(comp.Column, part, idx)
		case SearchParamString:
			clause, args, nextIdx = StringSearchClause(comp.Column, part, "", idx)
		case SearchParamReference:
			clause, args, nextIdx = ReferenceSearchClause(comp.Column, part, idx)
		case SearchParamURI:
			clause = fmt.Sprintf("%s = $%d", comp.Column, idx)
			args = []interface{}{part}
			nextIdx = idx + 1
		default:
			// Fallback: exact match
			clause = fmt.Sprintf("%s = $%d", comp.Column, idx)
			args = []interface{}{part}
			nextIdx = idx + 1
		}

		clauses = append(clauses, clause)
		allArgs = append(allArgs, args...)
		idx = nextIdx
	}

	if len(clauses) == 0 {
		return "1=1", nil, idx
	}

	combined := strings.Join(clauses, " AND ")
	if len(clauses) > 1 {
		combined = "(" + combined + ")"
	}

	return combined, allArgs, idx
}

// quantityComponentClause handles the quantity portion of a composite value.
// FHIR quantity values in composites use the pipe separator: number|system|code
// (e.g., "5.4|http://unitsofmeasure.org|mmol" or "gt5.4|http://unitsofmeasure.org|mmol").
// If no pipe is present, it delegates to NumberSearchClause directly.
func quantityComponentClause(column string, value string, argIdx int) (string, []interface{}, int) {
	if !strings.Contains(value, "|") {
		return NumberSearchClause(column, value, argIdx)
	}

	// Split on pipe: number|system|code
	qParts := strings.SplitN(value, "|", 3)
	numPart := qParts[0]

	// Use NumberSearchClause for the numeric portion (handles prefixes like gt, le, etc.)
	return NumberSearchClause(column, numPart, argIdx)
}

// AddComposite adds a composite search clause to the SearchQuery.
// The value should contain dollar-sign-separated component values
// (e.g., "http://loinc.org|8480-6$gt100").
func (q *SearchQuery) AddComposite(config CompositeSearchConfig, value string) {
	clause, args, nextIdx := CompositeSearchClause(config, value, q.idx)
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
}

// ---------------------------------------------------------------------------
// Default composite search configurations for common FHIR R4 resources
// ---------------------------------------------------------------------------

// DefaultCompositeConfigs returns a map of standard FHIR composite search
// parameter names to their configurations. These cover the most common
// composite parameters defined in the FHIR R4 specification, primarily
// for the Observation resource.
func DefaultCompositeConfigs() map[string]CompositeSearchConfig {
	return map[string]CompositeSearchConfig{
		// code-value-quantity: Observation code + value as Quantity
		// Usage: code-value-quantity=http://loinc.org|8480-6$gt5.4|http://unitsofmeasure.org|mmol
		"code-value-quantity": {
			Components: []CompositeComponent{
				{
					Name:      "code",
					Type:      SearchParamToken,
					Column:    "code_value",
					SysColumn: "code_system",
				},
				{
					Name:   "value",
					Type:   SearchParamQuantity,
					Column: "value_quantity",
				},
			},
		},

		// code-value-concept: Observation code + value as CodeableConcept
		// Usage: code-value-concept=http://loinc.org|8480-6$http://snomed.info/sct|227507002
		"code-value-concept": {
			Components: []CompositeComponent{
				{
					Name:      "code",
					Type:      SearchParamToken,
					Column:    "code_value",
					SysColumn: "code_system",
				},
				{
					Name:      "value",
					Type:      SearchParamToken,
					Column:    "value_code",
					SysColumn: "value_system",
				},
			},
		},

		// code-value-date: Observation code + value as date/dateTime
		// Usage: code-value-date=http://loinc.org|8480-6$ge2023-01-01
		"code-value-date": {
			Components: []CompositeComponent{
				{
					Name:      "code",
					Type:      SearchParamToken,
					Column:    "code_value",
					SysColumn: "code_system",
				},
				{
					Name:   "value",
					Type:   SearchParamDate,
					Column: "value_date",
				},
			},
		},

		// code-value-string: Observation code + value as string
		// Usage: code-value-string=http://loinc.org|8480-6$positive
		"code-value-string": {
			Components: []CompositeComponent{
				{
					Name:      "code",
					Type:      SearchParamToken,
					Column:    "code_value",
					SysColumn: "code_system",
				},
				{
					Name:   "value",
					Type:   SearchParamString,
					Column: "value_string",
				},
			},
		},

		// combo-code-value-quantity: Observation code + value across components
		// Searches both Observation.code/value and Observation.component.code/value
		// Usage: combo-code-value-quantity=http://loinc.org|8480-6$gt5.4
		"combo-code-value-quantity": {
			Components: []CompositeComponent{
				{
					Name:      "code",
					Type:      SearchParamToken,
					Column:    "combo_code_value",
					SysColumn: "combo_code_system",
				},
				{
					Name:   "value",
					Type:   SearchParamQuantity,
					Column: "combo_value_quantity",
				},
			},
		},

		// combo-code-value-concept: Observation code + value as CodeableConcept across components
		// Usage: combo-code-value-concept=http://loinc.org|8480-6$http://snomed.info/sct|227507002
		"combo-code-value-concept": {
			Components: []CompositeComponent{
				{
					Name:      "code",
					Type:      SearchParamToken,
					Column:    "combo_code_value",
					SysColumn: "combo_code_system",
				},
				{
					Name:      "value",
					Type:      SearchParamToken,
					Column:    "combo_value_code",
					SysColumn: "combo_value_system",
				},
			},
		},
	}
}
