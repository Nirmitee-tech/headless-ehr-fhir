package fhir

import (
	"context"
	"encoding/json"
)

// RevIncludeProvider is an interface for fetching resources referenced by _revinclude.
// Implementations can look up Provenance records (or other resource types) that
// reference a given set of target resource IDs.
type RevIncludeProvider interface {
	// FindByTargets returns resources that reference any of the given target references.
	// targetRefs are FHIR-style references like "Condition/abc-123".
	FindByTargets(ctx context.Context, targetRefs []string) ([]interface{}, error)
}

// ApplyRevInclude appends revincluded resources to a search bundle.
// It extracts target references from the bundle entries, queries the provider,
// and appends the results as "include" entries.
func ApplyRevInclude(bundle *Bundle, ctx context.Context, provider RevIncludeProvider) error {
	if provider == nil || len(bundle.Entry) == 0 {
		return nil
	}

	// Extract target references from existing entries
	var targetRefs []string
	for _, entry := range bundle.Entry {
		var resource map[string]interface{}
		if err := json.Unmarshal(entry.Resource, &resource); err != nil {
			continue
		}
		rt, _ := resource["resourceType"].(string)
		id, _ := resource["id"].(string)
		if rt != "" && id != "" {
			targetRefs = append(targetRefs, rt+"/"+id)
		}
	}

	if len(targetRefs) == 0 {
		return nil
	}

	// Fetch revincluded resources
	included, err := provider.FindByTargets(ctx, targetRefs)
	if err != nil {
		return err
	}

	// Append as include entries
	for _, r := range included {
		raw, err := json.Marshal(r)
		if err != nil {
			continue
		}
		fullURL := extractFullURL(r, "")
		bundle.Entry = append(bundle.Entry, BundleEntry{
			FullURL:  fullURL,
			Resource: raw,
			Search: &BundleSearch{
				Mode: "include",
			},
		})
	}

	return nil
}
