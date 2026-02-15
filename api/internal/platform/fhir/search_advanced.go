package fhir

import (
	"fmt"
	"strings"
)

// resourceTableMap maps FHIR resource types to their database table names.
var resourceTableMap = map[string]string{
	"Patient":              "patients",
	"Observation":          "observations",
	"Condition":            "conditions",
	"Encounter":            "encounters",
	"Procedure":            "procedures",
	"MedicationRequest":    "medication_requests",
	"AllergyIntolerance":   "allergy_intolerances",
	"DiagnosticReport":     "diagnostic_reports",
	"Immunization":         "immunizations",
	"CarePlan":             "care_plans",
	"Claim":                "claims",
	"ServiceRequest":       "service_requests",
}

// referenceColumnMap maps FHIR reference search parameters to database column names.
var referenceColumnMap = map[string]string{
	"patient":   "patient_id",
	"subject":   "patient_id",
	"encounter": "encounter_id",
	"performer": "performer_id",
	"requester": "requester_id",
	"author":    "author_id",
}

// searchColumnMap maps FHIR search parameters to database column names.
var searchColumnMap = map[string]string{
	"code":            "code",
	"status":          "status",
	"category":        "category",
	"date":            "effective_date",
	"type":            "type",
	"clinical-status": "clinical_status",
}

// filterFieldMap maps FHIR _filter field names to database column names.
var filterFieldMap = map[string]string{
	"name":      "last_name",
	"family":    "last_name",
	"given":     "first_name",
	"birthdate": "birth_date",
	"gender":    "gender",
	"status":    "status",
	"date":      "effective_date",
	"code":      "code",
	"category":  "category",
	"active":    "active",
}

// ---------------------------------------------------------------------------
// _has: reverse chaining
// ---------------------------------------------------------------------------

// ParseHasParamFromQuery parses a _has parameter from a query key and value.
// The key format is: _has:TargetResource:referenceParam:searchParam
// Returns an error for malformed parameters.
func ParseHasParamFromQuery(key, value string) (*HasParam, error) {
	if value == "" {
		return nil, fmt.Errorf("_has parameter value must not be empty")
	}

	if !strings.HasPrefix(key, "_has:") {
		return nil, fmt.Errorf("_has parameter key must start with '_has:', got %q", key)
	}

	rest := key[len("_has:"):]
	if rest == "" {
		return nil, fmt.Errorf("_has parameter key missing resource type, reference param, and search param")
	}

	parts := strings.Split(rest, ":")
	if len(parts) < 3 {
		return nil, fmt.Errorf("_has parameter key must have format _has:ResourceType:referenceParam:searchParam, got %q", key)
	}
	if len(parts) > 3 {
		return nil, fmt.Errorf("_has parameter key has too many parts: expected 3 after _has, got %d in %q", len(parts), key)
	}

	resourceType := parts[0]
	referenceParam := parts[1]
	searchParam := parts[2]

	if resourceType == "" || referenceParam == "" || searchParam == "" {
		return nil, fmt.Errorf("_has parameter parts must not be empty in %q", key)
	}

	// Validate that the resource type is known.
	if _, ok := resourceTableMap[resourceType]; !ok {
		return nil, fmt.Errorf("unknown FHIR resource type %q in _has parameter", resourceType)
	}

	return &HasParam{
		TargetType:  resourceType,
		TargetParam: referenceParam,
		SearchParam: searchParam,
		Value:       value,
	}, nil
}

// HasQueryBuilder converts _has parameters to SQL WHERE EXISTS subqueries.
type HasQueryBuilder struct {
	knownTypes map[string]bool
}

// NewHasQueryBuilder creates a new HasQueryBuilder.
func NewHasQueryBuilder() *HasQueryBuilder {
	known := make(map[string]bool, len(resourceTableMap))
	for k := range resourceTableMap {
		known[k] = true
	}
	return &HasQueryBuilder{knownTypes: known}
}

// BuildSQL returns a SQL fragment and bind parameters for a _has parameter.
// The baseTable is the table name of the resource being searched (e.g. "patients").
// Example output: "EXISTS (SELECT 1 FROM observations WHERE observations.patient_id = patients.id AND observations.code = $1)"
func (b *HasQueryBuilder) BuildSQL(baseTable string, has *HasParam) (string, []interface{}, error) {
	if has == nil {
		return "", nil, fmt.Errorf("has parameter must not be nil")
	}

	targetTable, ok := resourceTableMap[has.TargetType]
	if !ok {
		return "", nil, fmt.Errorf("unknown resource type %q for _has query", has.TargetType)
	}

	refCol, ok := referenceColumnMap[has.TargetParam]
	if !ok {
		return "", nil, fmt.Errorf("unknown reference parameter %q for _has query", has.TargetParam)
	}

	searchCol, ok := searchColumnMap[has.SearchParam]
	if !ok {
		return "", nil, fmt.Errorf("unknown search parameter %q for _has query", has.SearchParam)
	}

	sql := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.id AND %s.%s = $1)",
		targetTable, targetTable, refCol, baseTable, targetTable, searchCol,
	)

	return sql, []interface{}{has.Value}, nil
}

// ---------------------------------------------------------------------------
// _filter: advanced filtering
// ---------------------------------------------------------------------------

// FilterOp represents a comparison operator in a _filter expression.
type FilterOp string

const (
	FilterOpEq FilterOp = "eq" // equals
	FilterOpNe FilterOp = "ne" // not equals
	FilterOpGt FilterOp = "gt" // greater than
	FilterOpLt FilterOp = "lt" // less than
	FilterOpGe FilterOp = "ge" // greater or equal
	FilterOpLe FilterOp = "le" // less or equal
	FilterOpCo FilterOp = "co" // contains
	FilterOpSw FilterOp = "sw" // starts with
	FilterOpEw FilterOp = "ew" // ends with
	FilterOpSa FilterOp = "sa" // starts after
	FilterOpEb FilterOp = "eb" // ends before
)

// validFilterOps contains all valid filter operators for validation.
var validFilterOps = map[FilterOp]bool{
	FilterOpEq: true, FilterOpNe: true,
	FilterOpGt: true, FilterOpLt: true,
	FilterOpGe: true, FilterOpLe: true,
	FilterOpCo: true, FilterOpSw: true, FilterOpEw: true,
	FilterOpSa: true, FilterOpEb: true,
}

// FilterLogic represents a logical combiner between filter expressions.
type FilterLogic string

const (
	FilterLogicAnd  FilterLogic = "and"
	FilterLogicOr   FilterLogic = "or"
	FilterLogicNone FilterLogic = ""
)

// FilterExpression represents a parsed _filter expression.
type FilterExpression struct {
	Field    string             // e.g. "name", "birthdate"
	Op       FilterOp           // e.g. "eq", "ne", "gt"
	Value    string             // e.g. "Smith", "1990-01-01"
	Logic    FilterLogic        // "and", "or" (for combining with next expression)
	Children []*FilterExpression // for grouped sub-expressions (reserved for future use)
}

// ParseFilter parses a _filter expression string into a slice of FilterExpressions.
// Example: `name eq "Smith" and birthdate ge 1990-01-01`
// Returns an error if the expression is empty or malformed.
func ParseFilter(expr string) ([]*FilterExpression, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("_filter expression must not be empty")
	}

	// Tokenize the expression, splitting on " and " / " or " while respecting quoted strings.
	clauses, logics, err := splitFilterClauses(expr)
	if err != nil {
		return nil, err
	}

	results := make([]*FilterExpression, 0, len(clauses))
	for i, clause := range clauses {
		fe, err := parseFilterClause(clause)
		if err != nil {
			return nil, fmt.Errorf("error parsing filter clause %q: %w", clause, err)
		}
		if i < len(logics) {
			fe.Logic = logics[i]
		}
		results = append(results, fe)
	}

	return results, nil
}

// splitFilterClauses splits a filter expression into individual clauses and the logic
// operators between them, respecting quoted strings.
func splitFilterClauses(expr string) ([]string, []FilterLogic, error) {
	var clauses []string
	var logics []FilterLogic

	// We scan through the expression looking for top-level " and " or " or " delimiters,
	// skipping over content inside double quotes.
	inQuote := false
	start := 0

	for i := 0; i < len(expr); i++ {
		if expr[i] == '"' {
			inQuote = !inQuote
			continue
		}
		if inQuote {
			continue
		}

		// Check for " and " (space-and-space).
		if i+5 <= len(expr) && expr[i:i+5] == " and " {
			clause := strings.TrimSpace(expr[start:i])
			if clause == "" {
				return nil, nil, fmt.Errorf("empty clause before 'and' in filter expression")
			}
			clauses = append(clauses, clause)
			logics = append(logics, FilterLogicAnd)
			i += 4 // skip " and", loop will increment past the trailing space
			start = i + 1
			continue
		}

		// Check for " or " (space-or-space).
		if i+4 <= len(expr) && expr[i:i+4] == " or " {
			clause := strings.TrimSpace(expr[start:i])
			if clause == "" {
				return nil, nil, fmt.Errorf("empty clause before 'or' in filter expression")
			}
			clauses = append(clauses, clause)
			logics = append(logics, FilterLogicOr)
			i += 3
			start = i + 1
			continue
		}
	}

	// Final clause after the last logic operator.
	last := strings.TrimSpace(expr[start:])
	if last == "" {
		return nil, nil, fmt.Errorf("empty clause at end of filter expression")
	}
	clauses = append(clauses, last)

	return clauses, logics, nil
}

// parseFilterClause parses a single "field op value" clause.
func parseFilterClause(clause string) (*FilterExpression, error) {
	clause = strings.TrimSpace(clause)

	// Find the first space to separate the field.
	idx1 := strings.IndexByte(clause, ' ')
	if idx1 < 0 {
		return nil, fmt.Errorf("expected 'field op value', got %q", clause)
	}
	field := clause[:idx1]
	rest := strings.TrimSpace(clause[idx1+1:])

	// Find the second space to separate the operator from the value.
	idx2 := strings.IndexByte(rest, ' ')
	if idx2 < 0 {
		return nil, fmt.Errorf("expected 'field op value', got %q", clause)
	}
	opStr := rest[:idx2]
	rawValue := strings.TrimSpace(rest[idx2+1:])

	op := FilterOp(opStr)
	if !validFilterOps[op] {
		return nil, fmt.Errorf("unknown filter operator %q", opStr)
	}

	// Strip surrounding double quotes from the value if present.
	value := stripQuotes(rawValue)

	return &FilterExpression{
		Field: field,
		Op:    op,
		Value: value,
	}, nil
}

// stripQuotes removes surrounding double quotes from a string.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// FilterToSQL converts parsed filter expressions to a SQL WHERE clause fragment
// and corresponding bind parameters. The resourceType is currently unused but
// reserved for resource-specific column resolution.
func FilterToSQL(filters []*FilterExpression, resourceType string) (string, []interface{}, error) {
	if len(filters) == 0 {
		return "", nil, fmt.Errorf("no filter expressions provided")
	}

	var clauses []string
	var args []interface{}
	argIdx := 1

	for i, f := range filters {
		col, ok := filterFieldMap[f.Field]
		if !ok {
			return "", nil, fmt.Errorf("unknown filter field %q", f.Field)
		}

		sqlFrag, val, err := filterOpToSQL(col, f.Op, f.Value, argIdx)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, sqlFrag)
		args = append(args, val)
		argIdx++

		// Add the logic combiner if this is not the last expression.
		if i < len(filters)-1 && f.Logic != FilterLogicNone {
			clauses = append(clauses, strings.ToUpper(string(f.Logic)))
		}
	}

	return strings.Join(clauses, " "), args, nil
}

// filterOpToSQL converts a single filter operator to a SQL fragment.
func filterOpToSQL(column string, op FilterOp, value string, argIdx int) (string, interface{}, error) {
	placeholder := fmt.Sprintf("$%d", argIdx)

	switch op {
	case FilterOpEq:
		return fmt.Sprintf("%s = %s", column, placeholder), value, nil
	case FilterOpNe:
		return fmt.Sprintf("%s != %s", column, placeholder), value, nil
	case FilterOpGt, FilterOpSa:
		return fmt.Sprintf("%s > %s", column, placeholder), value, nil
	case FilterOpLt, FilterOpEb:
		return fmt.Sprintf("%s < %s", column, placeholder), value, nil
	case FilterOpGe:
		return fmt.Sprintf("%s >= %s", column, placeholder), value, nil
	case FilterOpLe:
		return fmt.Sprintf("%s <= %s", column, placeholder), value, nil
	case FilterOpCo:
		return fmt.Sprintf("%s ILIKE %s", column, placeholder), "%" + value + "%", nil
	case FilterOpSw:
		return fmt.Sprintf("%s ILIKE %s", column, placeholder), value + "%", nil
	case FilterOpEw:
		return fmt.Sprintf("%s ILIKE %s", column, placeholder), "%" + value, nil
	default:
		return "", nil, fmt.Errorf("unsupported filter operator %q", op)
	}
}
