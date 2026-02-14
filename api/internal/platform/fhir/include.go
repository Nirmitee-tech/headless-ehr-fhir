package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ResourceFetcher retrieves a single resource by its FHIR ID and returns its FHIR map representation.
type ResourceFetcher func(ctx context.Context, fhirID string) (map[string]interface{}, error)

// IncludeRegistry maps resource types to their fetcher functions and reference definitions.
type IncludeRegistry struct {
	mu       sync.RWMutex
	fetchers map[string]ResourceFetcher
	// references maps ResourceType -> search param name -> (target resource type, db column)
	references map[string]map[string]IncludeRef
	// revRefs maps target ResourceType -> source resource type + param
	revRefs map[string][]RevIncludeRef
}

// IncludeRef defines a reference from one resource type to another.
type IncludeRef struct {
	TargetType string // e.g., "Patient"
	SearchParam string // e.g., "subject"
}

// RevIncludeRef defines a reverse include from a source to a target.
type RevIncludeRef struct {
	SourceType  string // e.g., "Observation"
	SearchParam string // e.g., "subject"
}

// NewIncludeRegistry creates a new empty IncludeRegistry.
func NewIncludeRegistry() *IncludeRegistry {
	return &IncludeRegistry{
		fetchers:   make(map[string]ResourceFetcher),
		references: make(map[string]map[string]IncludeRef),
		revRefs:    make(map[string][]RevIncludeRef),
	}
}

// RegisterFetcher registers a resource fetcher for a given resource type.
func (r *IncludeRegistry) RegisterFetcher(resourceType string, fetcher ResourceFetcher) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fetchers[resourceType] = fetcher
}

// RegisterReference registers a reference from sourceType.searchParam -> targetType.
func (r *IncludeRegistry) RegisterReference(sourceType, searchParam, targetType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.references[sourceType] == nil {
		r.references[sourceType] = make(map[string]IncludeRef)
	}
	r.references[sourceType][searchParam] = IncludeRef{TargetType: targetType, SearchParam: searchParam}
	r.revRefs[targetType] = append(r.revRefs[targetType], RevIncludeRef{SourceType: sourceType, SearchParam: searchParam})
}

// GetIncludeTargets returns the include definitions for a resource type.
func (r *IncludeRegistry) GetIncludeTargets(resourceType string) map[string]IncludeRef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.references[resourceType]
}

// GetRevIncludeTargets returns reverse include sources for a resource type.
func (r *IncludeRegistry) GetRevIncludeTargets(resourceType string) []RevIncludeRef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.revRefs[resourceType]
}

// ResolveIncludes processes _include parameters and fetches referenced resources.
// includeParam format: "ResourceType:searchParam" or "ResourceType:searchParam:targetType"
func (r *IncludeRegistry) ResolveIncludes(ctx context.Context, resources []interface{}, includeParams []string) ([]BundleEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(includeParams) == 0 || len(resources) == 0 {
		return nil, nil
	}

	// Collect reference IDs to fetch
	type fetchReq struct {
		resourceType string
		id           string
	}
	seen := make(map[string]bool)
	var toFetch []fetchReq

	for _, param := range includeParams {
		parts := strings.SplitN(param, ":", 3)
		if len(parts) < 2 {
			continue
		}
		sourceType := parts[0]
		searchParam := parts[1]

		ref, ok := r.references[sourceType]
		if !ok {
			continue
		}
		refDef, ok := ref[searchParam]
		if !ok {
			continue
		}

		targetType := refDef.TargetType
		if len(parts) == 3 {
			targetType = parts[2]
		}

		// Extract reference IDs from resources
		for _, res := range resources {
			refID := extractReferenceID(res, searchParam, targetType)
			if refID != "" {
				key := targetType + "/" + refID
				if !seen[key] {
					seen[key] = true
					toFetch = append(toFetch, fetchReq{resourceType: targetType, id: refID})
				}
			}
		}
	}

	// Fetch referenced resources
	var entries []BundleEntry
	for _, req := range toFetch {
		fetcher, ok := r.fetchers[req.resourceType]
		if !ok {
			continue
		}
		resource, err := fetcher(ctx, req.id)
		if err != nil {
			continue // Skip resources that can't be fetched
		}
		raw, _ := json.Marshal(resource)
		entries = append(entries, BundleEntry{
			FullURL:  fmt.Sprintf("%s/%s", req.resourceType, req.id),
			Resource: raw,
			Search: &BundleSearch{
				Mode: "include",
			},
		})
	}

	return entries, nil
}

// extractReferenceID extracts the ID from a FHIR reference in a resource map.
// It checks common FHIR reference patterns for the given search parameter name.
func extractReferenceID(resource interface{}, searchParam string, targetType string) string {
	m, ok := toMap(resource)
	if !ok {
		return ""
	}

	// Map FHIR search param names to possible JSON fields
	fieldNames := searchParamToFields(searchParam)

	for _, field := range fieldNames {
		val, ok := m[field]
		if !ok {
			continue
		}

		// Handle Reference type
		refMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		refStr, ok := refMap["reference"].(string)
		if !ok {
			continue
		}

		// Parse "ResourceType/id" format
		parts := strings.SplitN(refStr, "/", 2)
		if len(parts) == 2 {
			if targetType == "" || parts[0] == targetType {
				return parts[1]
			}
		}
	}

	return ""
}

// searchParamToFields maps FHIR search parameter names to JSON field names.
func searchParamToFields(param string) []string {
	// Common mappings
	mappings := map[string][]string{
		"subject":     {"subject"},
		"patient":     {"subject", "patient"},
		"encounter":   {"encounter"},
		"performer":   {"performer"},
		"requester":   {"requester"},
		"recorder":    {"recorder"},
		"asserter":    {"asserter"},
		"author":      {"author"},
		"practitioner": {"practitioner"},
		"organization": {"managingOrganization", "organization"},
		"location":    {"location"},
	}

	if fields, ok := mappings[param]; ok {
		return fields
	}
	return []string{param}
}
