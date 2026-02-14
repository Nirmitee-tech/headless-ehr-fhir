package fhir

import (
	"fmt"
	"strings"
)

// SortSpec represents a single sort directive.
type SortSpec struct {
	Field      string
	Descending bool
}

// ParseSort parses the _sort query parameter value.
// Format: "-date,status" means date DESC, status ASC.
// A leading "-" indicates descending order.
func ParseSort(sortParam string) []SortSpec {
	if sortParam == "" {
		return nil
	}

	parts := strings.Split(sortParam, ",")
	specs := make([]SortSpec, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		spec := SortSpec{}
		if strings.HasPrefix(part, "-") {
			spec.Descending = true
			spec.Field = part[1:]
		} else {
			spec.Field = part
		}

		if spec.Field != "" {
			specs = append(specs, spec)
		}
	}

	return specs
}

// BuildOrderClause generates an ORDER BY clause from sort specs using a field mapping.
// The fieldMap maps FHIR search parameter names to SQL column names.
// Returns empty string if no valid sort fields are found.
// defaultOrder is appended if no valid sort fields match (e.g., "created_at DESC").
func BuildOrderClause(specs []SortSpec, fieldMap map[string]string, defaultOrder string) string {
	if len(specs) == 0 {
		if defaultOrder != "" {
			return " ORDER BY " + defaultOrder
		}
		return ""
	}

	var parts []string
	for _, spec := range specs {
		col, ok := fieldMap[spec.Field]
		if !ok {
			continue
		}

		dir := "ASC"
		if spec.Descending {
			dir = "DESC"
		}
		parts = append(parts, fmt.Sprintf("%s %s", col, dir))
	}

	if len(parts) == 0 {
		if defaultOrder != "" {
			return " ORDER BY " + defaultOrder
		}
		return ""
	}

	return " ORDER BY " + strings.Join(parts, ", ")
}

// BuildOrderClauseNullsLast is like BuildOrderClause but appends NULLS LAST to DESC columns.
func BuildOrderClauseNullsLast(specs []SortSpec, fieldMap map[string]string, defaultOrder string) string {
	if len(specs) == 0 {
		if defaultOrder != "" {
			return " ORDER BY " + defaultOrder
		}
		return ""
	}

	var parts []string
	for _, spec := range specs {
		col, ok := fieldMap[spec.Field]
		if !ok {
			continue
		}

		if spec.Descending {
			parts = append(parts, fmt.Sprintf("%s DESC NULLS LAST", col))
		} else {
			parts = append(parts, fmt.Sprintf("%s ASC", col))
		}
	}

	if len(parts) == 0 {
		if defaultOrder != "" {
			return " ORDER BY " + defaultOrder
		}
		return ""
	}

	return " ORDER BY " + strings.Join(parts, ", ")
}
