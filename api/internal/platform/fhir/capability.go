package fhir

import (
	"sort"
	"sync"
	"time"
)

// SearchParam describes a search parameter for use with the CapabilityBuilder.
type SearchParam struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Documentation string `json:"documentation,omitempty"`
}

// CapabilityBuilder accumulates resource registrations from domain modules and
// builds a dynamic FHIR CapabilityStatement. Domains call AddResource during
// server initialization so the /fhir/metadata response reflects only what is
// actually available.
type CapabilityBuilder struct {
	mu        sync.RWMutex
	resources map[string]*resourceEntry

	// Server metadata
	ServerVersion string
	BaseURL       string
	AuthorizeURL  string
	TokenURL      string
}

// resourceEntry holds the accumulated information for a single FHIR resource type.
type resourceEntry struct {
	resourceType string
	interactions []string
	searchParams []SearchParam
	profiles     []string
}

// NewCapabilityBuilder creates a new builder. The baseURL is the FHIR server
// base URL (e.g., "http://localhost:8000/fhir"), and version is the server
// software version.
func NewCapabilityBuilder(baseURL, version string) *CapabilityBuilder {
	return &CapabilityBuilder{
		resources:     make(map[string]*resourceEntry),
		ServerVersion: version,
		BaseURL:       baseURL,
	}
}

// SetOAuthURIs configures the SMART on FHIR OAuth URIs included in the
// security section of the CapabilityStatement.
func (b *CapabilityBuilder) SetOAuthURIs(authorizeURL, tokenURL string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.AuthorizeURL = authorizeURL
	b.TokenURL = tokenURL
}

// AddResource registers a FHIR resource type with the given interactions and
// search parameters. If the resource type was already registered, the new
// interactions and search parameters are merged with the existing ones.
func (b *CapabilityBuilder) AddResource(resourceType string, interactions []string, searchParams []SearchParam) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry, ok := b.resources[resourceType]
	if !ok {
		entry = &resourceEntry{
			resourceType: resourceType,
		}
		b.resources[resourceType] = entry
	}

	// Merge interactions (deduplicate)
	existing := make(map[string]bool, len(entry.interactions))
	for _, i := range entry.interactions {
		existing[i] = true
	}
	for _, i := range interactions {
		if !existing[i] {
			entry.interactions = append(entry.interactions, i)
			existing[i] = true
		}
	}

	// Merge search params (deduplicate by name)
	existingParams := make(map[string]bool, len(entry.searchParams))
	for _, p := range entry.searchParams {
		existingParams[p.Name] = true
	}
	for _, p := range searchParams {
		if !existingParams[p.Name] {
			entry.searchParams = append(entry.searchParams, p)
			existingParams[p.Name] = true
		}
	}
}

// AddResourceWithProfile registers a FHIR resource type with a supported profile URI.
func (b *CapabilityBuilder) AddResourceWithProfile(resourceType string, interactions []string, searchParams []SearchParam, profiles []string) {
	b.AddResource(resourceType, interactions, searchParams)

	b.mu.Lock()
	defer b.mu.Unlock()

	entry := b.resources[resourceType]
	existing := make(map[string]bool, len(entry.profiles))
	for _, p := range entry.profiles {
		existing[p] = true
	}
	for _, p := range profiles {
		if !existing[p] {
			entry.profiles = append(entry.profiles, p)
		}
	}
}

// Build constructs the full CapabilityStatement as a map suitable for JSON
// serialization. Resources are sorted alphabetically by type.
func (b *CapabilityBuilder) Build() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Sort resource types alphabetically for deterministic output
	types := make([]string, 0, len(b.resources))
	for rt := range b.resources {
		types = append(types, rt)
	}
	sort.Strings(types)

	// Build resource entries
	resources := make([]map[string]interface{}, 0, len(types))
	for _, rt := range types {
		entry := b.resources[rt]
		res := map[string]interface{}{
			"type":       entry.resourceType,
			"versioning": "versioned",
		}

		// Interactions
		if len(entry.interactions) > 0 {
			interactions := make([]map[string]string, len(entry.interactions))
			for i, code := range entry.interactions {
				interactions[i] = map[string]string{"code": code}
			}
			res["interaction"] = interactions
		}

		// Search parameters
		if len(entry.searchParams) > 0 {
			params := make([]map[string]string, len(entry.searchParams))
			for i, sp := range entry.searchParams {
				p := map[string]string{
					"name": sp.Name,
					"type": sp.Type,
				}
				if sp.Documentation != "" {
					p["documentation"] = sp.Documentation
				}
				params[i] = p
			}
			res["searchParam"] = params
		}

		// Supported profiles
		if len(entry.profiles) > 0 {
			res["supportedProfile"] = entry.profiles
		}

		resources = append(resources, res)
	}

	// Build security section
	security := b.buildSecurity()

	// Build rest entry
	rest := map[string]interface{}{
		"mode":     "server",
		"resource": resources,
	}
	if security != nil {
		rest["security"] = security
	}

	cs := map[string]interface{}{
		"resourceType": "CapabilityStatement",
		"status":       "active",
		"date":         time.Now().UTC().Format("2006-01-02"),
		"kind":         "instance",
		"fhirVersion":  "4.0.1",
		"format":       []string{"json"},
		"software": map[string]string{
			"name":    "Headless EHR",
			"version": b.ServerVersion,
		},
		"implementation": map[string]string{
			"description": "Headless EHR FHIR R4 Server",
			"url":         b.BaseURL,
		},
		"rest": []map[string]interface{}{rest},
	}

	return cs
}

// buildSecurity creates the SMART on FHIR security section with OAuth extension.
func (b *CapabilityBuilder) buildSecurity() map[string]interface{} {
	service := map[string]interface{}{
		"coding": []map[string]string{
			{
				"system": "http://hl7.org/fhir/restful-security-service",
				"code":   "SMART-on-FHIR",
			},
		},
	}

	security := map[string]interface{}{
		"cors":    true,
		"service": []map[string]interface{}{service},
	}

	// Add SMART on FHIR OAuth extension if URIs are configured
	if b.AuthorizeURL != "" || b.TokenURL != "" {
		oauthExtensions := make([]map[string]string, 0, 2)
		if b.AuthorizeURL != "" {
			oauthExtensions = append(oauthExtensions, map[string]string{
				"url":      "authorize",
				"valueUri": b.AuthorizeURL,
			})
		}
		if b.TokenURL != "" {
			oauthExtensions = append(oauthExtensions, map[string]string{
				"url":      "token",
				"valueUri": b.TokenURL,
			})
		}

		smartExtension := map[string]interface{}{
			"url":       "http://fhir-registry.smarthealthit.org/StructureDefinition/oauth-uris",
			"extension": oauthExtensions,
		}

		security["extension"] = []map[string]interface{}{smartExtension}
	}

	return security
}

// ResourceCount returns the number of registered resource types.
func (b *CapabilityBuilder) ResourceCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.resources)
}

// DefaultInteractions returns the standard set of CRUD interactions for a
// resource type. This is a convenience for domain modules that support all
// standard operations.
func DefaultInteractions() []string {
	return []string{"read", "vread", "search-type", "create", "update", "delete"}
}

// ReadOnlyInteractions returns interactions for read-only resources.
func ReadOnlyInteractions() []string {
	return []string{"read", "vread", "search-type"}
}
