package fhir

import (
	"fmt"
	"strings"
	"time"
)

// SearchPrefix represents a FHIR search prefix for ordered values.
type SearchPrefix string

const (
	PrefixEq SearchPrefix = "eq"
	PrefixNe SearchPrefix = "ne"
	PrefixGt SearchPrefix = "gt"
	PrefixLt SearchPrefix = "lt"
	PrefixGe SearchPrefix = "ge"
	PrefixLe SearchPrefix = "le"
	PrefixSa SearchPrefix = "sa" // starts after
	PrefixEb SearchPrefix = "eb" // ends before
	PrefixAp SearchPrefix = "ap" // approximately
)

// SearchModifier represents a FHIR search modifier.
type SearchModifier string

const (
	ModifierExact    SearchModifier = "exact"
	ModifierContains SearchModifier = "contains"
	ModifierText     SearchModifier = "text"
	ModifierNot      SearchModifier = "not"
	ModifierAbove    SearchModifier = "above"
	ModifierBelow    SearchModifier = "below"
	ModifierMissing  SearchModifier = "missing"
)

// ParsedSearch holds a parsed search parameter value with its prefix.
type ParsedSearch struct {
	Prefix SearchPrefix
	Value  string
}

// ParseSearchValue extracts the prefix from a FHIR search value.
// Examples: "gt2023-01-01" -> (gt, "2023-01-01"), "100" -> (eq, "100")
func ParseSearchValue(raw string) ParsedSearch {
	if len(raw) >= 2 {
		prefix := SearchPrefix(strings.ToLower(raw[:2]))
		switch prefix {
		case PrefixEq, PrefixNe, PrefixGt, PrefixLt, PrefixGe, PrefixLe, PrefixSa, PrefixEb, PrefixAp:
			return ParsedSearch{Prefix: prefix, Value: raw[2:]}
		}
	}
	return ParsedSearch{Prefix: PrefixEq, Value: raw}
}

// ParseParamModifier splits a parameter name from its modifier.
// Examples: "name:exact" -> ("name", "exact"), "code" -> ("code", "")
func ParseParamModifier(paramName string) (string, SearchModifier) {
	parts := strings.SplitN(paramName, ":", 2)
	if len(parts) == 2 {
		return parts[0], SearchModifier(parts[1])
	}
	return parts[0], ""
}

// DateSearchClause generates SQL for a date search parameter with prefix support.
// Returns the SQL clause and the arguments to bind.
// The column parameter is the SQL column name.
// The argIdx is the current parameter index for positional args.
func DateSearchClause(column string, value string, argIdx int) (string, []interface{}, int) {
	parsed := ParseSearchValue(value)

	// Try to parse the date value
	t, err := parseFlexDate(parsed.Value)
	if err != nil {
		// Fallback to exact match on the raw string
		return fmt.Sprintf("%s::text = $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	}

	switch parsed.Prefix {
	case PrefixGt, PrefixSa:
		return fmt.Sprintf("%s > $%d", column, argIdx), []interface{}{t}, argIdx + 1
	case PrefixLt, PrefixEb:
		return fmt.Sprintf("%s < $%d", column, argIdx), []interface{}{t}, argIdx + 1
	case PrefixGe:
		return fmt.Sprintf("%s >= $%d", column, argIdx), []interface{}{t}, argIdx + 1
	case PrefixLe:
		return fmt.Sprintf("%s <= $%d", column, argIdx), []interface{}{t}, argIdx + 1
	case PrefixNe:
		return fmt.Sprintf("%s != $%d", column, argIdx), []interface{}{t}, argIdx + 1
	case PrefixAp:
		// Approximate: within 10% of the duration or 1 day, whichever is greater
		oneDay := 24 * time.Hour
		low := t.Add(-oneDay)
		high := t.Add(oneDay)
		clause := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", column, argIdx, column, argIdx+1)
		return clause, []interface{}{low, high}, argIdx + 2
	default: // eq
		// For date-only values, match the entire day
		if len(parsed.Value) == 10 { // YYYY-MM-DD format
			endOfDay := t.Add(24*time.Hour - time.Nanosecond)
			clause := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", column, argIdx, column, argIdx+1)
			return clause, []interface{}{t, endOfDay}, argIdx + 2
		}
		return fmt.Sprintf("%s = $%d", column, argIdx), []interface{}{t}, argIdx + 1
	}
}

// NumberSearchClause generates SQL for a number search parameter with prefix support.
func NumberSearchClause(column string, value string, argIdx int) (string, []interface{}, int) {
	parsed := ParseSearchValue(value)

	switch parsed.Prefix {
	case PrefixGt, PrefixSa:
		return fmt.Sprintf("%s > $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	case PrefixLt, PrefixEb:
		return fmt.Sprintf("%s < $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	case PrefixGe:
		return fmt.Sprintf("%s >= $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	case PrefixLe:
		return fmt.Sprintf("%s <= $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	case PrefixNe:
		return fmt.Sprintf("%s != $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	default:
		return fmt.Sprintf("%s = $%d", column, argIdx), []interface{}{parsed.Value}, argIdx + 1
	}
}

// TokenSearchClause handles token search parameters in the format "system|code", "|code", "system|", or just "code".
func TokenSearchClause(systemCol, codeCol string, value string, argIdx int) (string, []interface{}, int) {
	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		system := parts[0]
		code := parts[1]

		if system != "" && code != "" {
			clause := fmt.Sprintf("(%s = $%d AND %s = $%d)", systemCol, argIdx, codeCol, argIdx+1)
			return clause, []interface{}{system, code}, argIdx + 2
		} else if system != "" {
			return fmt.Sprintf("%s = $%d", systemCol, argIdx), []interface{}{system}, argIdx + 1
		} else if code != "" {
			return fmt.Sprintf("%s = $%d", codeCol, argIdx), []interface{}{code}, argIdx + 1
		}
	}

	// No pipe: just match the code
	return fmt.Sprintf("%s = $%d", codeCol, argIdx), []interface{}{value}, argIdx + 1
}

// StringSearchClause handles string search parameters with modifier support.
func StringSearchClause(column string, value string, modifier SearchModifier, argIdx int) (string, []interface{}, int) {
	switch modifier {
	case ModifierExact:
		return fmt.Sprintf("%s = $%d", column, argIdx), []interface{}{value}, argIdx + 1
	case ModifierContains:
		return fmt.Sprintf("%s ILIKE $%d", column, argIdx), []interface{}{"%" + value + "%"}, argIdx + 1
	case ModifierText:
		return fmt.Sprintf("%s ILIKE $%d", column, argIdx), []interface{}{"%" + value + "%"}, argIdx + 1
	default:
		// Default string search: case-insensitive prefix match
		return fmt.Sprintf("%s ILIKE $%d", column, argIdx), []interface{}{value + "%"}, argIdx + 1
	}
}

// parseFlexDate parses a date string in multiple FHIR-supported formats.
func parseFlexDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
		"2006-01",
		"2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

// ReferenceSearchClause parses a FHIR reference value and returns a SQL clause.
// Handles formats: "Patient/fhir-id", "Patient/uuid", "fhir-id", "uuid".
// When the value is not a valid UUID, it resolves via a subquery against the
// referenced table's fhir_id column (e.g. patient_id → patient table).
func ReferenceSearchClause(column string, value string, argIdx int) (string, []interface{}, int) {
	// Extract resource type prefix if present (e.g. "Patient/abc" → resourceType="Patient", value="abc")
	var resourceType string
	if idx := strings.LastIndex(value, "/"); idx >= 0 {
		resourceType = value[:idx]
		value = value[idx+1:]
	}

	// If the value looks like a UUID, match directly.
	if isUUID(value) {
		return fmt.Sprintf("%s = $%d", column, argIdx), []interface{}{value}, argIdx + 1
	}

	// Not a UUID — resolve via fhir_id subquery.
	// Infer the target table from the column name (e.g. "patient_id" → "patient")
	// or from the resource type prefix (e.g. "Patient" → "patient").
	tableName := referenceTargetTable(column, resourceType)
	if tableName != "" {
		clause := fmt.Sprintf("%s = (SELECT id FROM %s WHERE fhir_id = $%d LIMIT 1)", column, tableName, argIdx)
		return clause, []interface{}{value}, argIdx + 1
	}

	// Fallback: direct comparison (may fail if column is UUID type).
	return fmt.Sprintf("%s = $%d", column, argIdx), []interface{}{value}, argIdx + 1
}

// isUUID checks if a string looks like a valid UUID.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// referenceTargetTable infers the PostgreSQL table name from either the column
// name (e.g. "patient_id" → "patient") or the FHIR resource type prefix
// (e.g. "Patient" → "patient").
func referenceTargetTable(column, resourceType string) string {
	// First try from resource type prefix.
	if resourceType != "" {
		// Skip URL-style references (e.g. "http://example.org/fhir/Patient") — these
		// indicate a full URL reference where the ID may not be a FHIR ID in our system.
		if strings.Contains(resourceType, "://") || strings.Contains(resourceType, ".") {
			// Fall through to column-based inference.
		} else {
			return strings.ToLower(resourceType)
		}
	}
	// Infer from column name: "patient_id" → "patient", "encounter_id" → "encounter".
	if strings.HasSuffix(column, "_id") {
		return strings.TrimSuffix(column, "_id")
	}
	return ""
}
