package fhir

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// :missing modifier
// ---------------------------------------------------------------------------

// MissingSearchClause generates SQL for the :missing modifier.
// When missing is true, generates "column IS NULL".
// When missing is false, generates "column IS NOT NULL".
// No bind parameters are needed, so nextIdx remains unchanged.
func MissingSearchClause(column string, missing bool, startIdx int) (string, []interface{}, int) {
	if missing {
		return fmt.Sprintf("%s IS NULL", column), nil, startIdx
	}
	return fmt.Sprintf("%s IS NOT NULL", column), nil, startIdx
}

// ParseMissingModifier checks whether paramName ends with ":missing".
// Returns the base parameter name, whether the value represents a missing=true
// search, and whether the :missing modifier was present at all.
func ParseMissingModifier(paramName string) (baseName string, isMissing bool, hasMissing bool) {
	const suffix = ":missing"
	if strings.HasSuffix(paramName, suffix) {
		baseName = paramName[:len(paramName)-len(suffix)]
		return baseName, true, true
	}
	return paramName, false, false
}

// ---------------------------------------------------------------------------
// :type modifier for polymorphic references
// ---------------------------------------------------------------------------

// TypedReferenceSearchClause generates SQL for a reference search filtered by
// target resource type. For example, subject:Patient=123 produces:
//
//	ref_column = $1 AND type_column = $2
//
// Both the reference value and the type value are bound as positional args.
func TypedReferenceSearchClause(refColumn, typeColumn, refValue, typeValue string, startIdx int) (string, []interface{}, int) {
	// Strip ResourceType/ prefix from the reference value if present
	if idx := strings.LastIndex(refValue, "/"); idx >= 0 {
		refValue = refValue[idx+1:]
	}
	clause := fmt.Sprintf("(%s = $%d AND %s = $%d)", refColumn, startIdx, typeColumn, startIdx+1)
	return clause, []interface{}{refValue, typeValue}, startIdx + 2
}

// ParseTypeModifier checks whether paramName contains a :ResourceType suffix
// (e.g. "subject:Patient"). It validates the resource type against known FHIR
// types. Returns the base name, the resource type, and whether the modifier
// was present and valid.
func ParseTypeModifier(paramName string) (baseName string, resourceType string, hasType bool) {
	idx := strings.Index(paramName, ":")
	if idx < 0 {
		return paramName, "", false
	}
	base := paramName[:idx]
	candidate := paramName[idx+1:]

	// The candidate must start with an uppercase letter to be a resource type,
	// and must be a known FHIR resource type.
	if candidate == "" || candidate[0] < 'A' || candidate[0] > 'Z' {
		return paramName, "", false
	}
	if !IsKnownResourceType(candidate) {
		return paramName, "", false
	}
	return base, candidate, true
}

// ---------------------------------------------------------------------------
// _total parameter
// ---------------------------------------------------------------------------

// TotalMode controls whether the server includes a total count in search
// result Bundles.
type TotalMode string

const (
	TotalNone     TotalMode = "none"     // Do not include total (best performance)
	TotalEstimate TotalMode = "estimate" // Include an estimated total
	TotalAccurate TotalMode = "accurate" // Include an exact total (may be slow)
)

// ParseTotalParam converts a raw _total query parameter value to a TotalMode.
// Unrecognised values default to TotalNone.
func ParseTotalParam(value string) TotalMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none":
		return TotalNone
	case "estimate":
		return TotalEstimate
	case "accurate":
		return TotalAccurate
	default:
		return TotalNone
	}
}

// ShouldIncludeTotal returns true when the TotalMode indicates that a count
// query should be executed.
func ShouldIncludeTotal(mode TotalMode) bool {
	return mode == TotalEstimate || mode == TotalAccurate
}

// ---------------------------------------------------------------------------
// :of-type modifier for token search
// ---------------------------------------------------------------------------

// OfTypeSearchClause generates SQL for the :of-type token modifier.
// The value is expected in the format "system|type|value" (three pipe-separated
// parts). Produces: sys_column = $N AND type_column = $N+1 AND code_column = $N+2
func OfTypeSearchClause(sysColumn, codeColumn, typeColumn, value string, startIdx int) (string, []interface{}, int) {
	parts := strings.SplitN(value, "|", 3)
	if len(parts) != 3 {
		// Malformed value — return a clause that will never match.
		return "1=0", nil, startIdx
	}
	system := parts[0]
	typeVal := parts[1]
	codeVal := parts[2]

	clause := fmt.Sprintf("(%s = $%d AND %s = $%d AND %s = $%d)",
		sysColumn, startIdx, typeColumn, startIdx+1, codeColumn, startIdx+2)
	return clause, []interface{}{system, typeVal, codeVal}, startIdx + 3
}

// ---------------------------------------------------------------------------
// :not modifier for token search
// ---------------------------------------------------------------------------

// NotTokenSearchClause generates a negated token search clause.
// For "system|code" it produces NOT (sys_column = $N AND code_column = $N+1).
// For code-only it produces NOT (code_column = $N).
func NotTokenSearchClause(sysColumn, codeColumn, value string, startIdx int) (string, []interface{}, int) {
	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		system := parts[0]
		code := parts[1]

		if system != "" && code != "" {
			clause := fmt.Sprintf("NOT (%s = $%d AND %s = $%d)", sysColumn, startIdx, codeColumn, startIdx+1)
			return clause, []interface{}{system, code}, startIdx + 2
		} else if system != "" {
			clause := fmt.Sprintf("NOT (%s = $%d)", sysColumn, startIdx)
			return clause, []interface{}{system}, startIdx + 1
		} else if code != "" {
			clause := fmt.Sprintf("NOT (%s = $%d)", codeColumn, startIdx)
			return clause, []interface{}{code}, startIdx + 1
		}
	}

	// Code-only
	clause := fmt.Sprintf("NOT (%s = $%d)", codeColumn, startIdx)
	return clause, []interface{}{value}, startIdx + 1
}

// ---------------------------------------------------------------------------
// :above / :below modifiers for token hierarchy
// ---------------------------------------------------------------------------

// AboveTokenSearchClause generates SQL that matches the given code and its
// ancestors. As a simple approximation without a full terminology service,
// this uses a prefix match where the given code must be a prefix of (or equal
// to) the stored code — i.e. the stored value "starts with" the search code.
func AboveTokenSearchClause(sysColumn, codeColumn, value string, startIdx int) (string, []interface{}, int) {
	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		system := parts[0]
		code := parts[1]

		if system != "" && code != "" {
			clause := fmt.Sprintf("(%s = $%d AND %s LIKE $%d)", sysColumn, startIdx, codeColumn, startIdx+1)
			return clause, []interface{}{system, code + "%"}, startIdx + 2
		} else if code != "" {
			clause := fmt.Sprintf("%s LIKE $%d", codeColumn, startIdx)
			return clause, []interface{}{code + "%"}, startIdx + 1
		} else if system != "" {
			clause := fmt.Sprintf("%s = $%d", sysColumn, startIdx)
			return clause, []interface{}{system}, startIdx + 1
		}
	}

	clause := fmt.Sprintf("%s LIKE $%d", codeColumn, startIdx)
	return clause, []interface{}{value + "%"}, startIdx + 1
}

// BelowTokenSearchClause generates SQL that matches the given code and its
// descendants. Uses a prefix match where the stored code starts with the
// search value.
func BelowTokenSearchClause(sysColumn, codeColumn, value string, startIdx int) (string, []interface{}, int) {
	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		system := parts[0]
		code := parts[1]

		if system != "" && code != "" {
			clause := fmt.Sprintf("(%s = $%d AND %s LIKE $%d)", sysColumn, startIdx, codeColumn, startIdx+1)
			return clause, []interface{}{system, code + "%"}, startIdx + 2
		} else if code != "" {
			clause := fmt.Sprintf("%s LIKE $%d", codeColumn, startIdx)
			return clause, []interface{}{code + "%"}, startIdx + 1
		} else if system != "" {
			clause := fmt.Sprintf("%s = $%d", sysColumn, startIdx)
			return clause, []interface{}{system}, startIdx + 1
		}
	}

	clause := fmt.Sprintf("%s LIKE $%d", codeColumn, startIdx)
	return clause, []interface{}{value + "%"}, startIdx + 1
}

// ---------------------------------------------------------------------------
// :in / :not-in modifiers — ValueSet expansion
// ---------------------------------------------------------------------------

// ValueSetResolver resolves a ValueSet URL into the set of codes it contains.
// Implementations typically call a terminology service or local cache.
type ValueSetResolver interface {
	Expand(url string) ([]string, error)
}

// InValueSetClause generates a SQL placeholder clause for the :in modifier.
// It produces a single-parameter clause: code_column = ANY($N) with the
// ValueSet URL as the argument. This is intended to be resolved at execution
// time by a middleware or the repository layer that expands the ValueSet.
func InValueSetClause(codeColumn, valueSetURL string, startIdx int) (string, []interface{}, int) {
	if valueSetURL == "" {
		return "1=0", nil, startIdx
	}
	clause := fmt.Sprintf("%s = ANY($%d)", codeColumn, startIdx)
	return clause, []interface{}{valueSetURL}, startIdx + 1
}

// NotInValueSetClause generates a SQL placeholder clause for the :not-in modifier.
func NotInValueSetClause(codeColumn, valueSetURL string, startIdx int) (string, []interface{}, int) {
	if valueSetURL == "" {
		return "1=1", nil, startIdx
	}
	clause := fmt.Sprintf("NOT (%s = ANY($%d))", codeColumn, startIdx)
	return clause, []interface{}{valueSetURL}, startIdx + 1
}

// ---------------------------------------------------------------------------
// ApplySearchModifiers — unified entry point
// ---------------------------------------------------------------------------

// ApplySearchModifiers inspects paramName for known FHIR search modifiers and
// applies the corresponding clause to the SearchQuery. It returns true if a
// modifier was detected and applied, false otherwise (the caller should then
// handle the parameter with normal search logic).
func ApplySearchModifiers(q *SearchQuery, paramName, paramValue string, config SearchParamConfig) bool {
	// 1. Check for :missing modifier
	baseName, _, hasMissing := ParseMissingModifier(paramName)
	if hasMissing {
		_ = baseName // baseName is used only for config lookup by caller
		missing := strings.ToLower(strings.TrimSpace(paramValue)) == "true"
		clause, args, nextIdx := MissingSearchClause(config.Column, missing, q.Idx())
		q.Add(clause, args...)
		// MissingSearchClause does not consume args, but Add already handles idx
		// advancement via len(args). Since args is nil, idx stays the same.
		// However, we need to ensure q.idx matches nextIdx.
		_ = nextIdx
		return true
	}

	// 2. Check for :ResourceType modifier (typed reference)
	_, resourceType, hasType := ParseTypeModifier(paramName)
	if hasType && config.Type == SearchParamReference {
		typeColumn := config.Column + "_type"
		if config.SysColumn != "" {
			typeColumn = config.SysColumn
		}
		clause, args, nextIdx := TypedReferenceSearchClause(config.Column, typeColumn, paramValue, resourceType, q.Idx())
		q.where += " AND " + clause
		q.args = append(q.args, args...)
		q.idx = nextIdx
		return true
	}

	// 3. Check for other known modifiers via ParseParamModifier
	_, modifier := ParseParamModifier(paramName)
	switch modifier {
	case "not":
		if config.Type == SearchParamToken {
			clause, args, nextIdx := NotTokenSearchClause(config.SysColumn, config.Column, paramValue, q.Idx())
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx = nextIdx
			return true
		}
	case "above":
		if config.Type == SearchParamToken && config.SysColumn != "" {
			clause, args, nextIdx := AboveTokenSearchClause(config.SysColumn, config.Column, paramValue, q.Idx())
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx = nextIdx
			return true
		}
	case "below":
		if config.Type == SearchParamToken && config.SysColumn != "" {
			clause, args, nextIdx := BelowTokenSearchClause(config.SysColumn, config.Column, paramValue, q.Idx())
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx = nextIdx
			return true
		}
	case "of-type":
		if config.Type == SearchParamToken && config.SysColumn != "" {
			typeColumn := config.Column + "_type"
			clause, args, nextIdx := OfTypeSearchClause(config.SysColumn, config.Column, typeColumn, paramValue, q.Idx())
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx = nextIdx
			return true
		}
	case "in":
		if config.Type == SearchParamToken {
			clause, args, nextIdx := InValueSetClause(config.Column, paramValue, q.Idx())
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx = nextIdx
			return true
		}
	case "not-in":
		if config.Type == SearchParamToken {
			clause, args, nextIdx := NotInValueSetClause(config.Column, paramValue, q.Idx())
			q.where += " AND " + clause
			q.args = append(q.args, args...)
			q.idx = nextIdx
			return true
		}
	}

	return false
}
