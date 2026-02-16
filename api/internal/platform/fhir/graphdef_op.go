package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// GraphDefinition represents a FHIR GraphDefinition resource for traversal.
type GraphDefinition struct {
	ResourceType string      `json:"resourceType"`
	ID           string      `json:"id,omitempty"`
	Name         string      `json:"name"`
	Start        string      `json:"start"` // Starting resource type
	Link         []GraphLink `json:"link,omitempty"`
}

// GraphLink describes a link in the graph from a source resource to one or
// more targets via a FHIRPath-like path expression.
type GraphLink struct {
	Path        string        `json:"path,omitempty"`
	SliceName   string        `json:"sliceName,omitempty"`
	Min         *int          `json:"min,omitempty"`
	Max         string        `json:"max,omitempty"`
	Description string        `json:"description,omitempty"`
	Target      []GraphTarget `json:"target,omitempty"`
}

// GraphTarget describes a target resource type and optional constraints for
// a graph link.
type GraphTarget struct {
	Type        string             `json:"type"`
	Params      string             `json:"params,omitempty"`
	Profile     string             `json:"profile,omitempty"`
	Compartment []GraphCompartment `json:"compartment,omitempty"`
	Link        []GraphLink        `json:"link,omitempty"` // Recursive
}

// GraphCompartment describes a compartment constraint on a graph target.
type GraphCompartment struct {
	Use        string `json:"use"`                  // "condition" or "requirement"
	Code       string `json:"code"`                 // compartment code
	Rule       string `json:"rule"`                 // "identical", "matching", "different", or "custom"
	Expression string `json:"expression,omitempty"` // FHIRPath expression for custom rules
}

// ParseGraphDefinition parses a GraphDefinition from raw JSON bytes.
func ParseGraphDefinition(data []byte) (*GraphDefinition, error) {
	var gd GraphDefinition
	if err := json.Unmarshal(data, &gd); err != nil {
		return nil, fmt.Errorf("invalid GraphDefinition JSON: %w", err)
	}
	return &gd, nil
}

// ValidateGraphDefinition performs basic structural validation on a
// GraphDefinition resource and returns any issues found.
func ValidateGraphDefinition(gd *GraphDefinition) []ValidationIssue {
	var issues []ValidationIssue

	if gd == nil {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityFatal,
			Code:        VIssueTypeStructure,
			Diagnostics: "GraphDefinition is nil",
		})
		return issues
	}

	if gd.ResourceType != "" && gd.ResourceType != "GraphDefinition" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeStructure,
			Location:    "GraphDefinition.resourceType",
			Diagnostics: fmt.Sprintf("resourceType must be 'GraphDefinition', got '%s'", gd.ResourceType),
		})
	}

	if gd.Name == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "GraphDefinition.name",
			Diagnostics: "name is required",
		})
	}

	if gd.Start == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityError,
			Code:        VIssueTypeRequired,
			Location:    "GraphDefinition.start",
			Diagnostics: "start resource type is required",
		})
	}

	// Validate links recursively.
	for i, link := range gd.Link {
		prefix := fmt.Sprintf("GraphDefinition.link[%d]", i)
		issues = append(issues, validateGraphLink(link, prefix)...)
	}

	return issues
}

// validateGraphLink validates a single GraphLink and its nested targets/links.
func validateGraphLink(link GraphLink, prefix string) []ValidationIssue {
	var issues []ValidationIssue

	if link.Path == "" {
		issues = append(issues, ValidationIssue{
			Severity:    SeverityWarning,
			Code:        VIssueTypeValue,
			Location:    prefix + ".path",
			Diagnostics: "link path is empty",
		})
	}

	for i, target := range link.Target {
		tPrefix := fmt.Sprintf("%s.target[%d]", prefix, i)
		if target.Type == "" {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeRequired,
				Location:    tPrefix + ".type",
				Diagnostics: "target type is required",
			})
		}

		// Validate compartments.
		for j, comp := range target.Compartment {
			cPrefix := fmt.Sprintf("%s.compartment[%d]", tPrefix, j)
			if comp.Use == "" {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    cPrefix + ".use",
					Diagnostics: "compartment use is required",
				})
			} else if comp.Use != "condition" && comp.Use != "requirement" {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    cPrefix + ".use",
					Diagnostics: fmt.Sprintf("compartment use must be 'condition' or 'requirement', got '%s'", comp.Use),
				})
			}

			if comp.Code == "" {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    cPrefix + ".code",
					Diagnostics: "compartment code is required",
				})
			}

			validRules := map[string]bool{
				"identical": true,
				"matching":  true,
				"different": true,
				"custom":    true,
			}
			if comp.Rule == "" {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeRequired,
					Location:    cPrefix + ".rule",
					Diagnostics: "compartment rule is required",
				})
			} else if !validRules[comp.Rule] {
				issues = append(issues, ValidationIssue{
					Severity:    SeverityError,
					Code:        VIssueTypeValue,
					Location:    cPrefix + ".rule",
					Diagnostics: fmt.Sprintf("compartment rule must be 'identical', 'matching', 'different', or 'custom', got '%s'", comp.Rule),
				})
			}
		}

		// Recurse into nested links.
		for k, nestedLink := range target.Link {
			nPrefix := fmt.Sprintf("%s.link[%d]", tPrefix, k)
			issues = append(issues, validateGraphLink(nestedLink, nPrefix)...)
		}
	}

	return issues
}

// =========================================================================
// Graph traversal
// =========================================================================

// graphTraverser holds state for a single $graph traversal.
type graphTraverser struct {
	registry *IncludeRegistry
	seen     map[string]bool
	entries  []BundleEntry
}

// newGraphTraverser creates a new traverser.
func newGraphTraverser(registry *IncludeRegistry) *graphTraverser {
	return &graphTraverser{
		registry: registry,
		seen:     make(map[string]bool),
	}
}

// addResource records a resource in the result set, returning false if it was
// already present.
func (t *graphTraverser) addResource(resource map[string]interface{}) bool {
	rt, _ := resource["resourceType"].(string)
	id, _ := resource["id"].(string)
	key := rt + "/" + id
	if t.seen[key] {
		return false
	}
	t.seen[key] = true

	raw, _ := json.Marshal(resource)
	t.entries = append(t.entries, BundleEntry{
		FullURL:  key,
		Resource: raw,
		Search: &BundleSearch{
			Mode: "match",
		},
	})
	return true
}

// traverse follows the graph links starting from the given resource.
func (t *graphTraverser) traverse(ctx context.Context, resource map[string]interface{}, links []GraphLink) {
	for _, link := range links {
		refs := extractRefsFromPath(resource, link.Path)
		for _, target := range link.Target {
			for _, ref := range refs {
				refType, refID := parseReference(ref)
				if refType == "" || refID == "" {
					continue
				}
				// Only follow if the reference matches the declared target type.
				if target.Type != "" && refType != target.Type {
					continue
				}

				t.registry.mu.RLock()
				fetcher, ok := t.registry.fetchers[refType]
				t.registry.mu.RUnlock()
				if !ok {
					continue
				}

				fetched, err := fetcher(ctx, refID)
				if err != nil || fetched == nil {
					continue
				}

				if t.addResource(fetched) {
					// Recurse into nested links.
					if len(target.Link) > 0 {
						t.traverse(ctx, fetched, target.Link)
					}
				}
			}
		}
	}
}

// extractRefsFromPath extracts reference strings from a resource at the given
// path. It supports simple dotted paths (e.g. "subject", "participant.individual")
// and handles both single references and arrays.
func extractRefsFromPath(resource map[string]interface{}, path string) []string {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	return extractRefsRecursive(resource, parts)
}

// extractRefsRecursive walks into the resource map following the path parts.
func extractRefsRecursive(obj map[string]interface{}, parts []string) []string {
	if len(parts) == 0 {
		// We've reached the target element; try to extract a reference.
		ref, _ := obj["reference"].(string)
		if ref != "" {
			return []string{ref}
		}
		return nil
	}

	field := parts[0]
	remaining := parts[1:]

	val, ok := obj[field]
	if !ok {
		return nil
	}

	switch v := val.(type) {
	case map[string]interface{}:
		if len(remaining) == 0 {
			// This element is the reference object itself.
			ref, _ := v["reference"].(string)
			if ref != "" {
				return []string{ref}
			}
			return nil
		}
		return extractRefsRecursive(v, remaining)

	case []interface{}:
		var refs []string
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if len(remaining) == 0 {
				ref, _ := m["reference"].(string)
				if ref != "" {
					refs = append(refs, ref)
				}
			} else {
				refs = append(refs, extractRefsRecursive(m, remaining)...)
			}
		}
		return refs
	}

	return nil
}

// parseReference splits a FHIR reference string ("Type/id") into its parts.
func parseReference(ref string) (string, string) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// =========================================================================
// HTTP Handler
// =========================================================================

// graphApplyRequest is the expected JSON body for the $graph operation.
type graphApplyRequest struct {
	GraphDefinition *GraphDefinition `json:"graphDefinition"`
	ResourceID      string           `json:"resourceId"`
	ResourceType    string           `json:"resourceType"`
}

// GraphApplyHandler creates a handler for POST /fhir/$graph.
//
// The handler accepts a JSON body containing a GraphDefinition and a starting
// resource identifier. It uses the IncludeRegistry to traverse references
// according to the graph definition and returns a Bundle containing all
// traversed resources.
func GraphApplyHandler(registry *IncludeRegistry) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "structure", "Failed to read request body"))
		}
		if len(body) == 0 {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "required", "Request body is empty"))
		}

		var req graphApplyRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
		}

		gd := req.GraphDefinition
		if gd == nil {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "required", "graphDefinition is required"))
		}

		// Validate the GraphDefinition.
		issues := ValidateGraphDefinition(gd)
		for _, issue := range issues {
			if issue.Severity == SeverityError || issue.Severity == SeverityFatal {
				return c.JSON(http.StatusBadRequest,
					operationOutcome("error", "structure",
						"Invalid GraphDefinition: "+issue.Diagnostics))
			}
		}

		// Determine start resource type and ID.
		startType := gd.Start
		if req.ResourceType != "" {
			startType = req.ResourceType
		}
		startID := req.ResourceID
		if startID == "" {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "required", "resourceId is required"))
		}

		// Fetch the starting resource.
		registry.mu.RLock()
		fetcher, ok := registry.fetchers[startType]
		registry.mu.RUnlock()
		if !ok {
			return c.JSON(http.StatusNotFound,
				operationOutcome("error", "not-found",
					fmt.Sprintf("No fetcher registered for resource type '%s'", startType)))
		}

		ctx := c.Request().Context()
		startResource, err := fetcher(ctx, startID)
		if err != nil || startResource == nil {
			return c.JSON(http.StatusNotFound,
				operationOutcome("error", "not-found",
					fmt.Sprintf("%s/%s not found", startType, startID)))
		}

		// Traverse the graph.
		traverser := newGraphTraverser(registry)
		traverser.addResource(startResource)
		traverser.traverse(ctx, startResource, gd.Link)

		// Build the result bundle.
		total := len(traverser.entries)
		bundle := &Bundle{
			ResourceType: "Bundle",
			Type:         "collection",
			Total:        &total,
			Entry:        traverser.entries,
		}

		c.Response().Header().Set(echo.HeaderContentType, FHIRContentType)
		return c.JSON(http.StatusOK, bundle)
	}
}
