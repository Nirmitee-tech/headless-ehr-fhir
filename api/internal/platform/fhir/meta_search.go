package fhir

import (
	"fmt"
	"strings"
)

// MetaSearchParam identifies the three FHIR meta search parameters that apply
// to all resource types.  These search against the resource's Meta element
// (meta.tag, meta.security, meta.profile).
const (
	MetaParamTag      = "_tag"
	MetaParamSecurity = "_security"
	MetaParamProfile  = "_profile"
)

// IsMetaSearchParam returns true if the parameter name is one of the special
// FHIR meta search parameters (_tag, _security, _profile).
func IsMetaSearchParam(name string) bool {
	switch name {
	case MetaParamTag, MetaParamSecurity, MetaParamProfile:
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// SQL-level meta search (for repositories that store meta as JSONB)
// ---------------------------------------------------------------------------

// MetaSearchClause generates a SQL WHERE clause fragment for a meta search
// parameter, assuming the resource meta is stored in a JSONB column.
//
// Parameters:
//   - paramName: one of _tag, _security, _profile
//   - value:     the search value.
//     For _tag and _security the format is "system|code" (token search).
//     For _profile the value is a canonical URI (exact match).
//   - jsonCol:   the SQL column expression that holds the JSONB resource
//     (e.g. "resource_json" or "meta_json").
//   - argIdx:    the current positional-argument index ($N).
//
// Returns the clause, bind arguments, and the next available argument index.
//
// SQL patterns:
//
//	_tag / _security  -> EXISTS (SELECT 1 FROM jsonb_array_elements(<col>->'meta'->'tag') elem
//	                      WHERE elem->>'system' = $N AND elem->>'code' = $N+1)
//	_profile          -> <col>->'meta'->'profile' ? $N
func MetaSearchClause(paramName, value, jsonCol string, argIdx int) (string, []interface{}, int) {
	switch paramName {
	case MetaParamTag:
		return metaTokenClause(jsonCol, "tag", value, argIdx)
	case MetaParamSecurity:
		return metaTokenClause(jsonCol, "security", value, argIdx)
	case MetaParamProfile:
		clause := fmt.Sprintf("%s->'meta'->'profile' ? $%d", jsonCol, argIdx)
		return clause, []interface{}{value}, argIdx + 1
	default:
		// Unknown meta param — return a no-op true clause.
		return "TRUE", nil, argIdx
	}
}

// metaTokenClause builds an EXISTS sub-query that searches a JSONB array of
// Coding objects (meta.tag or meta.security) for a matching system|code token.
func metaTokenClause(jsonCol, field, value string, argIdx int) (string, []interface{}, int) {
	path := fmt.Sprintf("%s->'meta'->'%s'", jsonCol, field)

	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		system := parts[0]
		code := parts[1]

		if system != "" && code != "" {
			clause := fmt.Sprintf(
				"EXISTS (SELECT 1 FROM jsonb_array_elements(%s) elem WHERE elem->>'system' = $%d AND elem->>'code' = $%d)",
				path, argIdx, argIdx+1,
			)
			return clause, []interface{}{system, code}, argIdx + 2
		}
		if system != "" {
			clause := fmt.Sprintf(
				"EXISTS (SELECT 1 FROM jsonb_array_elements(%s) elem WHERE elem->>'system' = $%d)",
				path, argIdx,
			)
			return clause, []interface{}{system}, argIdx + 1
		}
		if code != "" {
			clause := fmt.Sprintf(
				"EXISTS (SELECT 1 FROM jsonb_array_elements(%s) elem WHERE elem->>'code' = $%d)",
				path, argIdx,
			)
			return clause, []interface{}{code}, argIdx + 1
		}
	}

	// No pipe — match by code only.
	clause := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM jsonb_array_elements(%s) elem WHERE elem->>'code' = $%d)",
		path, argIdx,
	)
	return clause, []interface{}{value}, argIdx + 1
}

// AddMetaSearchSQL is a convenience helper that domain repositories can call
// after constructing their SearchQuery to append any _tag, _security, or
// _profile filters as SQL clauses.  It requires the table to have a JSONB
// column identified by jsonCol.
//
// Usage:
//
//	qb := fhir.NewSearchQuery("my_resource", cols)
//	qb.ApplyParams(params, mySearchConfigs)
//	fhir.AddMetaSearchSQL(qb, params, "resource_json")
func AddMetaSearchSQL(qb *SearchQuery, params map[string]string, jsonCol string) {
	for _, p := range []string{MetaParamTag, MetaParamSecurity, MetaParamProfile} {
		if v, ok := params[p]; ok && v != "" {
			clause, args, nextIdx := MetaSearchClause(p, v, jsonCol, qb.Idx())
			qb.Add(clause, args...)
			_ = nextIdx // idx is advanced by qb.Add via len(args)
		}
	}
}

// ---------------------------------------------------------------------------
// In-memory meta search filter (for repositories using individual columns)
// ---------------------------------------------------------------------------

// MetaSearchFilter holds the parsed criteria from _tag, _security, and
// _profile query parameters so they can be evaluated against FHIR resource
// maps produced by ToFHIR().
type MetaSearchFilter struct {
	// Tags is a list of (system, code) pairs parsed from _tag values.
	Tags []CodingMatch
	// SecurityLabels is a list of (system, code) pairs parsed from _security.
	SecurityLabels []CodingMatch
	// Profiles is a list of canonical URIs from _profile.
	Profiles []string
}

// CodingMatch represents a token match against a Coding element.
// Empty System or Code means "match any".
type CodingMatch struct {
	System   string
	Code     string
	MatchAny bool // true when both system and code are empty (|)
}

// NewMetaSearchFilter parses the meta search parameters from a params map.
// It returns nil if no meta search parameters are present, so callers can
// cheaply skip filtering.
func NewMetaSearchFilter(params map[string]string) *MetaSearchFilter {
	var f MetaSearchFilter
	hasAny := false

	if v, ok := params[MetaParamTag]; ok && v != "" {
		f.Tags = parseTokenList(v)
		hasAny = true
	}
	if v, ok := params[MetaParamSecurity]; ok && v != "" {
		f.SecurityLabels = parseTokenList(v)
		hasAny = true
	}
	if v, ok := params[MetaParamProfile]; ok && v != "" {
		f.Profiles = strings.Split(v, ",")
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return &f
}

// parseTokenList splits a comma-separated list of token values and parses
// each one into a CodingMatch.
func parseTokenList(raw string) []CodingMatch {
	var matches []CodingMatch
	for _, tok := range strings.Split(raw, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		matches = append(matches, parseTokenValue(tok))
	}
	return matches
}

// parseTokenValue parses a single "system|code" token value into a
// CodingMatch.
func parseTokenValue(value string) CodingMatch {
	if strings.Contains(value, "|") {
		parts := strings.SplitN(value, "|", 2)
		return CodingMatch{System: parts[0], Code: parts[1]}
	}
	return CodingMatch{Code: value}
}

// Match evaluates whether a FHIR resource map (the output of ToFHIR())
// satisfies all the meta search criteria in this filter.
//
// The resource map is expected to have a "meta" key whose value is either a
// fhir.Meta struct or a map[string]interface{} with the standard FHIR meta
// structure.
//
// Returns true if the resource matches ALL specified criteria (AND semantics
// across parameter types, OR semantics within a comma-separated list).
func (f *MetaSearchFilter) Match(resource map[string]interface{}) bool {
	if f == nil {
		return true
	}

	meta := extractMeta(resource)

	if len(f.Tags) > 0 {
		tags := extractCodingSlice(meta, "tag")
		if !matchAnyCoding(f.Tags, tags) {
			return false
		}
	}

	if len(f.SecurityLabels) > 0 {
		security := extractCodingSlice(meta, "security")
		if !matchAnyCoding(f.SecurityLabels, security) {
			return false
		}
	}

	if len(f.Profiles) > 0 {
		profiles := extractStringSlice(meta, "profile")
		if !matchAnyProfile(f.Profiles, profiles) {
			return false
		}
	}

	return true
}

// ---------------------------------------------------------------------------
// Internal helpers for in-memory matching
// ---------------------------------------------------------------------------

// metaData is a lightweight carrier for extracted meta fields.
type metaData struct {
	raw interface{}
}

// extractMeta pulls the "meta" value from a resource map.
func extractMeta(resource map[string]interface{}) metaData {
	if resource == nil {
		return metaData{}
	}
	return metaData{raw: resource["meta"]}
}

// extractCodingSlice retrieves a slice of Coding-like objects from the meta
// for the given field ("tag" or "security").
func extractCodingSlice(m metaData, field string) []Coding {
	if m.raw == nil {
		return nil
	}

	// Case 1: fhir.Meta struct (most common in this codebase).
	if meta, ok := m.raw.(Meta); ok {
		switch field {
		case "tag":
			return meta.Tag
		case "security":
			return meta.Security
		}
		return nil
	}

	// Case 2: *Meta pointer.
	if meta, ok := m.raw.(*Meta); ok && meta != nil {
		switch field {
		case "tag":
			return meta.Tag
		case "security":
			return meta.Security
		}
		return nil
	}

	// Case 3: map[string]interface{} (generic FHIR JSON).
	if meta, ok := m.raw.(map[string]interface{}); ok {
		return codingsFromSlice(meta[field])
	}

	return nil
}

// codingsFromSlice converts a generic []interface{} of coding maps to Coding.
func codingsFromSlice(v interface{}) []Coding {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	codings := make([]Coding, 0, len(arr))
	for _, elem := range arr {
		if m, ok := elem.(map[string]interface{}); ok {
			c := Coding{}
			if s, ok := m["system"].(string); ok {
				c.System = s
			}
			if s, ok := m["code"].(string); ok {
				c.Code = s
			}
			codings = append(codings, c)
		}
	}
	return codings
}

// extractStringSlice retrieves a []string from the meta for the given field
// (typically "profile").
func extractStringSlice(m metaData, field string) []string {
	if m.raw == nil {
		return nil
	}

	// fhir.Meta struct.
	if meta, ok := m.raw.(Meta); ok {
		if field == "profile" {
			return meta.Profile
		}
		return nil
	}

	// *Meta pointer.
	if meta, ok := m.raw.(*Meta); ok && meta != nil {
		if field == "profile" {
			return meta.Profile
		}
		return nil
	}

	// map[string]interface{}.
	if meta, ok := m.raw.(map[string]interface{}); ok {
		if arr, ok := meta[field].([]interface{}); ok {
			strs := make([]string, 0, len(arr))
			for _, v := range arr {
				if s, ok := v.(string); ok {
					strs = append(strs, s)
				}
			}
			return strs
		}
		if arr, ok := meta[field].([]string); ok {
			return arr
		}
	}

	return nil
}

// matchAnyCoding returns true if at least one CodingMatch in criteria matches
// at least one Coding in the resource's list (OR semantics).
func matchAnyCoding(criteria []CodingMatch, codings []Coding) bool {
	for _, cm := range criteria {
		for _, c := range codings {
			if codingMatches(cm, c) {
				return true
			}
		}
	}
	return false
}

// codingMatches checks whether a single CodingMatch criterion matches a
// Coding value.
func codingMatches(cm CodingMatch, c Coding) bool {
	if cm.System != "" && cm.System != c.System {
		return false
	}
	if cm.Code != "" && cm.Code != c.Code {
		return false
	}
	// If both system and code are empty, it matches nothing useful.
	if cm.System == "" && cm.Code == "" {
		return false
	}
	return true
}

// matchAnyProfile returns true if at least one requested profile URI appears
// in the resource's profile list.
func matchAnyProfile(requested []string, actual []string) bool {
	for _, req := range requested {
		for _, a := range actual {
			if req == a {
				return true
			}
		}
	}
	return false
}

// FilterResourceList applies the MetaSearchFilter to a slice of FHIR resource
// maps, returning only those that match.  If filter is nil, the original slice
// is returned unmodified.
func FilterResourceList(resources []map[string]interface{}, filter *MetaSearchFilter) []map[string]interface{} {
	if filter == nil {
		return resources
	}
	var result []map[string]interface{}
	for _, r := range resources {
		if filter.Match(r) {
			result = append(result, r)
		}
	}
	return result
}
