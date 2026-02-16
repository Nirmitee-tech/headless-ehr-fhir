package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// ValidationProfile represents a StructureDefinition profile for validation.
type ValidationProfile struct {
	URL          string                        // Canonical URL of the profile
	Name         string                        // Human-readable name
	ResourceType string                        // Base resource type
	Required     bool                          // Whether validation against this profile is required
	Constraints  []VProfileConstraint          // Profile-level constraints (FHIRPath expressions)
	Elements     map[string]*ElementConstraint // path -> constraint (field name -> constraint)
	Extensions   []ExtensionDefinition         // Extension definitions
	Version      string
}

// VProfileConstraint represents a single validation constraint with a FHIRPath expression.
type VProfileConstraint struct {
	Key        string // Constraint key (e.g., "us-core-1")
	Severity   string // error | warning
	Human      string // Human-readable description
	Expression string // FHIRPath expression
	XPath      string // XPath expression (legacy)
	Source     string // Profile URL that defines this constraint
}

// ElementConstraint describes constraints on a specific element.
type ElementConstraint struct {
	Path        string
	Min         int          // Minimum cardinality
	Max         string       // Maximum cardinality ("*" = unbounded)
	Types       []string     // Allowed types
	MustSupport bool
	Binding     *ValueSetBinding // Value set binding
	Fixed       interface{}      // Fixed value (if any)
	Pattern     interface{}      // Pattern value (if any)
	Slicing     *SlicingRules    // Slicing rules (if any)
}

// ValueSetBinding describes a value set binding on an element.
type ValueSetBinding struct {
	Strength string // required | extensible | preferred | example
	ValueSet string // Canonical URL of the value set
}

// SlicingRules describes how a list element is sliced.
type SlicingRules struct {
	Discriminator []SlicingDiscriminator
	Rules         string // closed | open | openAtEnd
	Ordered       bool
}

// SlicingDiscriminator identifies how slices are distinguished.
type SlicingDiscriminator struct {
	Type string // value | exists | pattern | type | profile
	Path string
}

// ExtensionDefinition describes an allowed/required extension.
type ExtensionDefinition struct {
	URL       string
	Required  bool
	ValueType string // Type of the extension value (e.g., "valueString", "valueBoolean")
}

// ProfileValidationResult contains the result of profile validation.
type ProfileValidationResult struct {
	Valid       bool
	Issues      []ValidationIssue
	ProfileURL  string
	ProfileName string
}

// ProfileValidationConfig configures validation behavior.
type ProfileValidationConfig struct {
	ValidateOnCreate bool     // Validate on resource creation
	ValidateOnUpdate bool     // Validate on resource update
	StrictMode       bool     // Treat warnings as errors
	RequiredProfiles []string // Profiles that must always be satisfied
	IgnoreProfiles   []string // Profiles to skip validation for
}

// ---------------------------------------------------------------------------
// ValidationProfileRegistry
// ---------------------------------------------------------------------------

// ValidationProfileRegistry manages validation profiles.
type ValidationProfileRegistry struct {
	mu       sync.RWMutex
	profiles map[string]*ValidationProfile // URL -> profile
	byType   map[string][]*ValidationProfile
	defaults map[string][]string // resourceType -> default profile URLs
}

// NewValidationProfileRegistry creates a new registry.
func NewValidationProfileRegistry() *ValidationProfileRegistry {
	return &ValidationProfileRegistry{
		profiles: make(map[string]*ValidationProfile),
		byType:   make(map[string][]*ValidationProfile),
		defaults: make(map[string][]string),
	}
}

// RegisterValidationProfile adds a validation profile to the registry.
func (r *ValidationProfileRegistry) RegisterValidationProfile(profile *ValidationProfile) error {
	if profile.URL == "" {
		return fmt.Errorf("profile URL is required")
	}
	if profile.ResourceType == "" {
		return fmt.Errorf("profile ResourceType is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// If a profile with this URL already exists, remove from type index.
	if existing, ok := r.profiles[profile.URL]; ok {
		r.removeFromTypeIndex(existing)
	}

	r.profiles[profile.URL] = profile
	r.byType[profile.ResourceType] = append(r.byType[profile.ResourceType], profile)

	return nil
}

// removeFromTypeIndex removes a profile from the type index. Must be called with mu held.
func (r *ValidationProfileRegistry) removeFromTypeIndex(p *ValidationProfile) {
	profiles := r.byType[p.ResourceType]
	for i, pp := range profiles {
		if pp.URL == p.URL {
			r.byType[p.ResourceType] = append(profiles[:i], profiles[i+1:]...)
			break
		}
	}
}

// UnregisterValidationProfile removes a validation profile.
func (r *ValidationProfileRegistry) UnregisterValidationProfile(url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.profiles[url]
	if !ok {
		return fmt.Errorf("profile '%s' not found", url)
	}

	r.removeFromTypeIndex(existing)
	delete(r.profiles, url)

	return nil
}

// GetValidationProfile returns a profile by URL.
func (r *ValidationProfileRegistry) GetValidationProfile(url string) (*ValidationProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.profiles[url]
	return p, ok
}

// ListValidationProfiles returns all registered profiles.
func (r *ValidationProfileRegistry) ListValidationProfiles() []*ValidationProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ValidationProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		result = append(result, p)
	}
	return result
}

// ListValidationProfilesForResourceType returns profiles for a specific resource type.
func (r *ValidationProfileRegistry) ListValidationProfilesForResourceType(resourceType string) []*ValidationProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	profiles := r.byType[resourceType]
	result := make([]*ValidationProfile, len(profiles))
	copy(result, profiles)
	return result
}

// SetDefaultValidationProfiles sets default profiles for a resource type.
func (r *ValidationProfileRegistry) SetDefaultValidationProfiles(resourceType string, profileURLs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.defaults[resourceType] = profileURLs
}

// GetDefaultValidationProfiles returns default profiles for a resource type.
func (r *ValidationProfileRegistry) GetDefaultValidationProfiles(resourceType string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	urls, ok := r.defaults[resourceType]
	if !ok {
		return nil
	}
	return urls
}

// ---------------------------------------------------------------------------
// Validation Functions
// ---------------------------------------------------------------------------

// ValidateAgainstProfile validates a resource against a specific profile.
func ValidateAgainstProfile(resource map[string]interface{}, profile *ValidationProfile) *ProfileValidationResult {
	result := &ProfileValidationResult{Valid: true}

	if profile == nil {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Diagnostics: "profile is nil",
		})
		return result
	}

	result.ProfileURL = profile.URL
	result.ProfileName = profile.Name

	if resource == nil {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Diagnostics: "resource is nil",
		})
		return result
	}

	// Check resource type matches
	rt, _ := resource["resourceType"].(string)
	if rt != profile.ResourceType {
		result.Valid = false
		result.Issues = append(result.Issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Diagnostics: fmt.Sprintf("resource type '%s' does not match profile type '%s'", rt, profile.ResourceType),
		})
		return result
	}

	// Validate elements
	for fieldName, constraint := range profile.Elements {
		// Cardinality checks
		cardIssues := ValidateCardinality(resource, fieldName, constraint)
		for _, ci := range cardIssues {
			result.Issues = append(result.Issues, ci)
			if ci.Severity == SeverityError || ci.Severity == SeverityFatal {
				result.Valid = false
			}
		}

		// Binding checks
		if constraint.Binding != nil {
			val, present := resource[fieldName]
			if present {
				bindIssues := ValidateBinding(val, constraint.Binding)
				for _, bi := range bindIssues {
					result.Issues = append(result.Issues, bi)
					if bi.Severity == SeverityError || bi.Severity == SeverityFatal {
						result.Valid = false
					}
				}
			}
		}

		// Fixed value checks
		if constraint.Fixed != nil {
			val, present := resource[fieldName]
			if present && !ValidateFixed(val, constraint.Fixed) {
				result.Valid = false
				result.Issues = append(result.Issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    constraint.Path,
					Diagnostics: fmt.Sprintf("value does not match fixed value for '%s'", constraint.Path),
				})
			}
		}

		// Pattern value checks
		if constraint.Pattern != nil {
			val, present := resource[fieldName]
			if present && !ValidatePattern(val, constraint.Pattern) {
				result.Valid = false
				result.Issues = append(result.Issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    constraint.Path,
					Diagnostics: fmt.Sprintf("value does not match pattern for '%s'", constraint.Path),
				})
			}
		}

		// Slicing checks
		if constraint.Slicing != nil {
			val, present := resource[fieldName]
			if present {
				if arr, ok := val.([]interface{}); ok {
					sliceIssues := ValidateSlicing(arr, constraint.Slicing)
					for _, si := range sliceIssues {
						result.Issues = append(result.Issues, si)
						if si.Severity == SeverityError || si.Severity == SeverityFatal {
							result.Valid = false
						}
					}
				}
			}
		}
	}

	// Must-support checks (warnings only, never make result invalid)
	msIssues := ValidateMustSupport(resource, profile.Elements)
	result.Issues = append(result.Issues, msIssues...)

	// Extension checks
	if len(profile.Extensions) > 0 {
		extIssues := ValidateExtensions(resource, profile.Extensions)
		for _, ei := range extIssues {
			result.Issues = append(result.Issues, ei)
			if ei.Severity == SeverityError || ei.Severity == SeverityFatal {
				result.Valid = false
			}
		}
	}

	return result
}

// ValidateCardinality checks min/max cardinality constraints.
func ValidateCardinality(resource map[string]interface{}, path string, constraint *ElementConstraint) []ValidationIssue {
	var issues []ValidationIssue

	val, present := resource[path]

	// Count the number of values
	count := 0
	if present {
		switch v := val.(type) {
		case []interface{}:
			count = len(v)
		case nil:
			count = 0
		default:
			count = 1
		}
	}

	// Check minimum cardinality
	if constraint.Min > 0 && count < constraint.Min {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    constraint.Path,
			Diagnostics: fmt.Sprintf("element '%s' requires at least %d value(s), found %d", constraint.Path, constraint.Min, count),
		})
	}

	// Check maximum cardinality
	if constraint.Max != "*" && constraint.Max != "" {
		maxVal, err := strconv.Atoi(constraint.Max)
		if err == nil {
			if maxVal == 0 && count > 0 {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeStructure,
					Location:    constraint.Path,
					Diagnostics: fmt.Sprintf("element '%s' is prohibited (max=0) but has %d value(s)", constraint.Path, count),
				})
			} else if maxVal > 0 && count > maxVal {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeStructure,
					Location:    constraint.Path,
					Diagnostics: fmt.Sprintf("element '%s' allows at most %d value(s), found %d", constraint.Path, maxVal, count),
				})
			}
		}
	}

	return issues
}

// ValidateBinding checks value set binding constraints.
func ValidateBinding(value interface{}, binding *ValueSetBinding) []ValidationIssue {
	if binding == nil {
		return nil
	}

	// Example binding produces no issues
	if binding.Strength == "example" {
		return nil
	}

	// Preferred binding produces no issues (just informational)
	if binding.Strength == "preferred" {
		return nil
	}

	// For known value sets, we can do actual validation
	switch binding.ValueSet {
	case "http://hl7.org/fhir/ValueSet/administrative-gender":
		return validateGenderBinding(value, binding)
	case "http://hl7.org/fhir/ValueSet/observation-status":
		return validateObservationStatusBinding(value, binding)
	}

	// For extensible binding with unknown value sets, produce a warning at most
	if binding.Strength == "extensible" {
		return nil
	}

	return nil
}

// validateGenderBinding validates against the administrative-gender value set.
func validateGenderBinding(value interface{}, binding *ValueSetBinding) []ValidationIssue {
	validGenders := map[string]bool{"male": true, "female": true, "other": true, "unknown": true}

	code := extractCodeFromValue(value)
	if code == "" {
		return nil
	}

	if !validGenders[code] {
		severity := SeverityError
		if binding.Strength == "extensible" {
			severity = SeverityWarning
		}
		return []ValidationIssue{{
			Severity:    severity,
			Code:        VIssueTypeValue,
			Diagnostics: fmt.Sprintf("value '%s' is not in value set '%s'", code, binding.ValueSet),
		}}
	}

	return nil
}

// validateObservationStatusBinding validates against the observation-status value set.
func validateObservationStatusBinding(value interface{}, binding *ValueSetBinding) []ValidationIssue {
	validStatuses := map[string]bool{
		"registered": true, "preliminary": true, "final": true, "amended": true,
		"corrected": true, "cancelled": true, "entered-in-error": true, "unknown": true,
	}

	code := extractCodeFromValue(value)
	if code == "" {
		return nil
	}

	if !validStatuses[code] {
		severity := SeverityError
		if binding.Strength == "extensible" {
			severity = SeverityWarning
		}
		return []ValidationIssue{{
			Severity:    severity,
			Code:        VIssueTypeValue,
			Diagnostics: fmt.Sprintf("value '%s' is not in value set '%s'", code, binding.ValueSet),
		}}
	}

	return nil
}

// extractCodeFromValue extracts a code string from various FHIR value types.
func extractCodeFromValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]interface{}:
		// CodeableConcept or Coding
		if codings, ok := v["coding"].([]interface{}); ok && len(codings) > 0 {
			if first, ok := codings[0].(map[string]interface{}); ok {
				if code, ok := first["code"].(string); ok {
					return code
				}
			}
		}
		if code, ok := v["code"].(string); ok {
			return code
		}
	}
	return ""
}

// ValidateFixed checks fixed value constraints.
func ValidateFixed(value interface{}, expected interface{}) bool {
	if value == nil && expected == nil {
		return true
	}
	if value == nil || expected == nil {
		return false
	}

	// Use JSON marshal for deep comparison
	valJSON, err1 := json.Marshal(value)
	expJSON, err2 := json.Marshal(expected)
	if err1 != nil || err2 != nil {
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", expected)
	}

	return string(valJSON) == string(expJSON)
}

// ValidatePattern checks pattern matching constraints.
// A pattern match means the value contains at least all the fields specified in the pattern.
func ValidatePattern(value interface{}, pattern interface{}) bool {
	if value == nil && pattern == nil {
		return true
	}
	if value == nil || pattern == nil {
		return false
	}

	switch pv := pattern.(type) {
	case map[string]interface{}:
		// Value must be a map that contains all pattern keys
		switch vv := value.(type) {
		case map[string]interface{}:
			return mapMatchesPattern(vv, pv)
		case []interface{}:
			// If value is an array, check if any element matches
			for _, item := range vv {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if mapMatchesPattern(itemMap, pv) {
						return true
					}
				}
			}
			return false
		default:
			return false
		}
	case []interface{}:
		// Pattern is an array - check that value (as array) contains matching elements
		valueArr, ok := value.([]interface{})
		if !ok {
			// If value is a map, check it against each pattern element
			if valueMap, ok := value.(map[string]interface{}); ok {
				for _, pItem := range pv {
					if pMap, ok := pItem.(map[string]interface{}); ok {
						if !mapContainsArrayPattern(valueMap, pMap) {
							return false
						}
					}
				}
				return true
			}
			return false
		}
		// Each pattern element must be found in the value array
		for _, pItem := range pv {
			found := false
			for _, vItem := range valueArr {
				if ValidatePattern(vItem, pItem) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		// Primitive comparison
		return ValidateFixed(value, pattern)
	}
}

// mapMatchesPattern checks that a map contains all the key-value pairs in the pattern.
func mapMatchesPattern(value, pattern map[string]interface{}) bool {
	for key, pVal := range pattern {
		vVal, ok := value[key]
		if !ok {
			return false
		}
		if !ValidatePattern(vVal, pVal) {
			return false
		}
	}
	return true
}

// mapContainsArrayPattern checks if a map's field contains an array matching the pattern.
func mapContainsArrayPattern(value map[string]interface{}, pattern map[string]interface{}) bool {
	for key, pVal := range pattern {
		vVal, ok := value[key]
		if !ok {
			return false
		}
		if !ValidatePattern(vVal, pVal) {
			return false
		}
	}
	return true
}

// ValidateMustSupport checks that must-support elements are present.
func ValidateMustSupport(resource map[string]interface{}, elements map[string]*ElementConstraint) []ValidationIssue {
	var issues []ValidationIssue

	for fieldName, constraint := range elements {
		if !constraint.MustSupport {
			continue
		}

		val, present := resource[fieldName]
		if !present || isEmptyVal(val) {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityWarning,
				Code:        VIssueTypeInvariant,
				Location:    constraint.Path,
				Diagnostics: fmt.Sprintf("MustSupport element '%s' is not present", constraint.Path),
			})
		}
	}

	return issues
}

// isEmptyVal checks if a value is considered empty for validation purposes.
func isEmptyVal(val interface{}) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

// ValidateExtensions validates extensions against profile definitions.
func ValidateExtensions(resource map[string]interface{}, extensions []ExtensionDefinition) []ValidationIssue {
	var issues []ValidationIssue

	// Build a map of present extensions
	presentExtensions := make(map[string]map[string]interface{})
	if extVal, ok := resource["extension"]; ok {
		if extArr, ok := extVal.([]interface{}); ok {
			for _, ext := range extArr {
				if extMap, ok := ext.(map[string]interface{}); ok {
					if extURL, ok := extMap["url"].(string); ok {
						presentExtensions[extURL] = extMap
					}
				}
			}
		}
	}

	for _, extDef := range extensions {
		ext, present := presentExtensions[extDef.URL]

		if extDef.Required && !present {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    "extension",
				Diagnostics: fmt.Sprintf("required extension '%s' is not present", extDef.URL),
			})
			continue
		}

		if !present {
			continue
		}

		// Check value type if specified
		if extDef.ValueType != "" {
			if _, hasType := ext[extDef.ValueType]; !hasType {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    "extension",
					Diagnostics: fmt.Sprintf("extension '%s' should have value type '%s'", extDef.URL, extDef.ValueType),
				})
			}
		}
	}

	return issues
}

// ValidateSlicing validates slicing rules on a list of values.
func ValidateSlicing(values []interface{}, rules *SlicingRules) []ValidationIssue {
	if rules == nil {
		return nil
	}
	if len(values) == 0 {
		return nil
	}

	var issues []ValidationIssue

	// Extract discriminator values for each element
	type discValue struct {
		key   string
		index int
	}

	for _, disc := range rules.Discriminator {
		var discValues []discValue

		for i, val := range values {
			valMap, ok := val.(map[string]interface{})
			if !ok {
				continue
			}

			var dv string
			switch disc.Type {
			case "value":
				dv = extractDiscriminatorValue(valMap, disc.Path)
			case "exists":
				_, exists := resolveNestedPath(valMap, disc.Path)
				if exists {
					dv = "true"
				} else {
					dv = "false"
				}
			case "pattern":
				dv = extractDiscriminatorValue(valMap, disc.Path)
			case "type":
				dv = extractDiscriminatorValue(valMap, disc.Path)
			default:
				dv = extractDiscriminatorValue(valMap, disc.Path)
			}

			discValues = append(discValues, discValue{key: dv, index: i})
		}

		// For closed slicing, check for duplicate discriminator values
		if rules.Rules == "closed" {
			seen := make(map[string]bool)
			for _, dv := range discValues {
				if dv.key == "" {
					continue
				}
				if seen[dv.key] {
					issues = append(issues, ValidationIssue{
						Severity:    SeverityError,
						Code:        VIssueTypeStructure,
						Diagnostics: fmt.Sprintf("duplicate discriminator value '%s' in closed slicing at path '%s'", dv.key, disc.Path),
					})
					break
				}
				seen[dv.key] = true
			}
		}

		// For ordered slicing, check that discriminator values are in sorted order
		if rules.Ordered && len(discValues) > 1 {
			for i := 1; i < len(discValues); i++ {
				if discValues[i].key < discValues[i-1].key {
					issues = append(issues, ValidationIssue{
						Severity:    SeverityError,
						Code:        VIssueTypeStructure,
						Diagnostics: fmt.Sprintf("slicing order violation: '%s' should come before '%s'", discValues[i].key, discValues[i-1].key),
					})
					break
				}
			}
		}
	}

	return issues
}

// extractDiscriminatorValue extracts a discriminator value from a map at the given path.
func extractDiscriminatorValue(m map[string]interface{}, path string) string {
	val, ok := resolveNestedPath(m, path)
	if !ok {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// resolveNestedPath resolves a dotted path within a map (e.g., "coding.system").
func resolveNestedPath(m map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = m

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		case []interface{}:
			// For arrays, try the first element
			if len(v) == 0 {
				return nil, false
			}
			if first, ok := v[0].(map[string]interface{}); ok {
				val, ok := first[part]
				if !ok {
					return nil, false
				}
				current = val
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// ExtractProfiles extracts profile URLs from resource meta.profile.
func ExtractProfiles(resource map[string]interface{}) []string {
	if resource == nil {
		return nil
	}

	metaVal, ok := resource["meta"]
	if !ok {
		return nil
	}

	metaMap, ok := metaVal.(map[string]interface{})
	if !ok {
		return nil
	}

	profileVal, ok := metaMap["profile"]
	if !ok {
		return nil
	}

	profileArr, ok := profileVal.([]interface{})
	if !ok {
		return nil
	}

	var profiles []string
	for _, p := range profileArr {
		if pStr, ok := p.(string); ok {
			profiles = append(profiles, pStr)
		}
	}

	return profiles
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// ProfileValidationMiddleware returns middleware that validates resources on write operations.
func ProfileValidationMiddleware(registry *ValidationProfileRegistry, config *ProfileValidationConfig) echo.MiddlewareFunc {
	// Build ignored set
	ignoredSet := make(map[string]bool)
	for _, url := range config.IgnoreProfiles {
		ignoredSet[url] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method

			// Skip non-write operations
			if method != http.MethodPost && method != http.MethodPut {
				return next(c)
			}

			// Check if we should validate for this operation
			if method == http.MethodPost && !config.ValidateOnCreate {
				return next(c)
			}
			if method == http.MethodPut && !config.ValidateOnUpdate {
				return next(c)
			}

			// Read and buffer the body
			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return next(c)
			}

			if len(body) == 0 {
				return next(c)
			}

			var resource map[string]interface{}
			if err := json.Unmarshal(body, &resource); err != nil {
				return next(c)
			}

			// Re-set the body for downstream handlers
			c.Request().Body = io.NopCloser(strings.NewReader(string(body)))

			// Determine which profiles to validate against
			rt, _ := resource["resourceType"].(string)
			if rt == "" {
				return next(c)
			}

			// Collect profile URLs
			var profileURLs []string

			// Add defaults for this resource type
			defaults := registry.GetDefaultValidationProfiles(rt)
			profileURLs = append(profileURLs, defaults...)

			// Add required profiles from config
			profileURLs = append(profileURLs, config.RequiredProfiles...)

			// Add profiles claimed in meta.profile
			metaProfiles := ExtractProfiles(resource)
			profileURLs = append(profileURLs, metaProfiles...)

			// Deduplicate
			seen := make(map[string]bool)
			var uniqueURLs []string
			for _, url := range profileURLs {
				if !seen[url] && !ignoredSet[url] {
					seen[url] = true
					uniqueURLs = append(uniqueURLs, url)
				}
			}

			// Validate against each profile
			var allResults []*ProfileValidationResult
			hasErrors := false

			for _, profileURL := range uniqueURLs {
				profile, ok := registry.GetValidationProfile(profileURL)
				if !ok {
					continue
				}

				result := ValidateAgainstProfile(resource, profile)
				allResults = append(allResults, result)

				if !result.Valid {
					hasErrors = true
				}

				// In strict mode, warnings become errors
				if config.StrictMode {
					for _, issue := range result.Issues {
						if issue.Severity == SeverityWarning {
							hasErrors = true
						}
					}
				}
			}

			if hasErrors {
				outcome := BuildProfileOperationOutcome(allResults)
				return c.JSON(http.StatusUnprocessableEntity, outcome)
			}

			return next(c)
		}
	}
}

// ---------------------------------------------------------------------------
// HTTP Handler
// ---------------------------------------------------------------------------

// ProfileValidationHandler returns a handler for POST /fhir/{type}/$validate?profile={url}.
func ProfileValidationHandler(registry *ValidationProfileRegistry) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, buildVPOutcome([]ValidationIssue{{
				Severity:    SeverityFatal,
				Code:        VIssueTypeStructure,
				Diagnostics: "failed to read request body",
			}}))
		}

		if len(body) == 0 {
			return c.JSON(http.StatusBadRequest, buildVPOutcome([]ValidationIssue{{
				Severity:    SeverityFatal,
				Code:        VIssueTypeStructure,
				Diagnostics: "request body is empty",
			}}))
		}

		var resource map[string]interface{}
		if err := json.Unmarshal(body, &resource); err != nil {
			return c.JSON(http.StatusBadRequest, buildVPOutcome([]ValidationIssue{{
				Severity:    SeverityFatal,
				Code:        VIssueTypeStructure,
				Diagnostics: "invalid JSON: " + err.Error(),
			}}))
		}

		// Parse profile parameters
		profileURLs := ParseValidateProfileParams(c.QueryParams())

		if len(profileURLs) == 0 {
			// No profiles specified - return success
			return c.JSON(http.StatusOK, buildVPOutcome([]ValidationIssue{{
				Severity:    SeverityInformation,
				Code:        VIssueTypeInvariant,
				Diagnostics: "no profile specified for validation; resource parsed successfully",
			}}))
		}

		var allResults []*ProfileValidationResult

		for _, profileURL := range profileURLs {
			profile, ok := registry.GetValidationProfile(profileURL)
			if !ok {
				return c.JSON(http.StatusNotFound, buildVPOutcome([]ValidationIssue{{
					Severity:    SeverityError,
					Code:        VIssueTypeNotFound,
					Diagnostics: fmt.Sprintf("profile '%s' not found", profileURL),
				}}))
			}

			result := ValidateAgainstProfile(resource, profile)
			allResults = append(allResults, result)
		}

		outcome := BuildProfileOperationOutcome(allResults)
		return c.JSON(http.StatusOK, outcome)
	}
}

// ParseValidateProfileParams parses the ?profile= query parameter.
func ParseValidateProfileParams(params url.Values) []string {
	values := params["profile"]
	var result []string
	for _, v := range values {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// US Core Default Profiles
// ---------------------------------------------------------------------------

// DefaultUSCoreValidationProfiles returns US Core v3 profile definitions.
func DefaultUSCoreValidationProfiles() []*ValidationProfile {
	return []*ValidationProfile{
		defaultUSCorePatientValidationProfile(),
		defaultUSCoreObservationValidationProfile(),
		defaultUSCoreConditionValidationProfile(),
	}
}

func defaultUSCorePatientValidationProfile() *ValidationProfile {
	return &ValidationProfile{
		URL:          "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
		Name:         "USCorePatientProfile",
		ResourceType: "Patient",
		Required:     true,
		Version:      "3.1.1",
		Elements: map[string]*ElementConstraint{
			"identifier": {
				Path:        "Patient.identifier",
				Min:         1,
				Max:         "*",
				MustSupport: true,
			},
			"name": {
				Path:        "Patient.name",
				Min:         1,
				Max:         "*",
				MustSupport: true,
			},
			"gender": {
				Path:        "Patient.gender",
				Min:         1,
				Max:         "1",
				MustSupport: true,
				Binding: &ValueSetBinding{
					Strength: "required",
					ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
				},
			},
			"birthDate": {
				Path:        "Patient.birthDate",
				Min:         0,
				Max:         "1",
				MustSupport: true,
			},
			"address": {
				Path:        "Patient.address",
				Min:         0,
				Max:         "*",
				MustSupport: true,
			},
			"telecom": {
				Path:        "Patient.telecom",
				Min:         0,
				Max:         "*",
				MustSupport: true,
			},
		},
		Extensions: []ExtensionDefinition{
			{
				URL:       "http://hl7.org/fhir/us/core/StructureDefinition/us-core-race",
				Required:  false,
				ValueType: "valueString",
			},
			{
				URL:       "http://hl7.org/fhir/us/core/StructureDefinition/us-core-ethnicity",
				Required:  false,
				ValueType: "valueString",
			},
		},
	}
}

func defaultUSCoreObservationValidationProfile() *ValidationProfile {
	return &ValidationProfile{
		URL:          "http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-lab",
		Name:         "USCoreObservationLabProfile",
		ResourceType: "Observation",
		Required:     true,
		Version:      "3.1.1",
		Elements: map[string]*ElementConstraint{
			"status": {
				Path: "Observation.status",
				Min:  1,
				Max:  "1",
				Binding: &ValueSetBinding{
					Strength: "required",
					ValueSet: "http://hl7.org/fhir/ValueSet/observation-status",
				},
			},
			"category": {
				Path: "Observation.category",
				Min:  1,
				Max:  "*",
			},
			"code": {
				Path: "Observation.code",
				Min:  1,
				Max:  "1",
			},
			"subject": {
				Path: "Observation.subject",
				Min:  1,
				Max:  "1",
			},
		},
	}
}

func defaultUSCoreConditionValidationProfile() *ValidationProfile {
	return &ValidationProfile{
		URL:          "http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition",
		Name:         "USCoreConditionProfile",
		ResourceType: "Condition",
		Required:     true,
		Version:      "3.1.1",
		Elements: map[string]*ElementConstraint{
			"clinicalStatus": {
				Path:        "Condition.clinicalStatus",
				Min:         0,
				Max:         "1",
				MustSupport: true,
			},
			"verificationStatus": {
				Path:        "Condition.verificationStatus",
				Min:         0,
				Max:         "1",
				MustSupport: true,
			},
			"category": {
				Path: "Condition.category",
				Min:  1,
				Max:  "*",
			},
			"code": {
				Path:        "Condition.code",
				Min:         1,
				Max:         "1",
				MustSupport: true,
			},
			"subject": {
				Path: "Condition.subject",
				Min:  1,
				Max:  "1",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// OperationOutcome Generation
// ---------------------------------------------------------------------------

// BuildProfileOperationOutcome converts validation results to FHIR OperationOutcome.
func BuildProfileOperationOutcome(results []*ProfileValidationResult) map[string]interface{} {
	if len(results) == 0 {
		return buildVPOutcome([]ValidationIssue{{
			Severity:    SeverityInformation,
			Code:        VIssueTypeInvariant,
			Diagnostics: "no validation results",
		}})
	}

	var allIssues []ValidationIssue
	allValid := true

	for _, result := range results {
		if !result.Valid {
			allValid = false
		}
		allIssues = append(allIssues, result.Issues...)
	}

	if allValid && len(allIssues) == 0 {
		return buildVPOutcome([]ValidationIssue{{
			Severity:    SeverityInformation,
			Code:        VIssueTypeInvariant,
			Diagnostics: "validation successful",
		}})
	}

	if len(allIssues) == 0 {
		return buildVPOutcome([]ValidationIssue{{
			Severity:    SeverityInformation,
			Code:        VIssueTypeInvariant,
			Diagnostics: "validation successful",
		}})
	}

	return buildVPOutcome(allIssues)
}

// buildVPOutcome builds a raw OperationOutcome map from validation issues.
func buildVPOutcome(issues []ValidationIssue) map[string]interface{} {
	issueList := make([]map[string]interface{}, 0, len(issues))
	for _, issue := range issues {
		entry := map[string]interface{}{
			"severity":    string(issue.Severity),
			"code":        string(issue.Code),
			"diagnostics": issue.Diagnostics,
		}
		if issue.Location != "" {
			entry["location"] = []string{issue.Location}
		}
		issueList = append(issueList, entry)
	}

	return map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue":        issueList,
	}
}
