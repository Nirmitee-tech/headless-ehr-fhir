package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Core types
// ---------------------------------------------------------------------------

// SearchParam describes a search parameter for use with the CapabilityBuilder.
type SearchParam struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Documentation string `json:"documentation,omitempty"`
}

// CapabilityConfig holds top-level server metadata for the CapabilityStatement.
type CapabilityConfig struct {
	ServerName        string   `json:"serverName"`
	ServerVersion     string   `json:"serverVersion"`
	FHIRVersion       string   `json:"fhirVersion"`
	Publisher         string   `json:"publisher"`
	Description       string   `json:"description"`
	BaseURL           string   `json:"baseURL"`
	SupportedFormats  []string `json:"supportedFormats"`
	SupportedVersions []string `json:"supportedVersions"`
}

// SearchParamCapability describes a search parameter in a ResourceCapability.
type SearchParamCapability struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Definition    string `json:"definition,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

// OperationCapability describes an operation (resource-level or system-level).
type OperationCapability struct {
	Name          string `json:"name"`
	Definition    string `json:"definition"`
	Documentation string `json:"documentation,omitempty"`
}

// ResourceCapability is the high-level description of a FHIR resource type
// that can be registered with the builder via AddResourceCapability.
type ResourceCapabilityDef struct {
	Type              string                  `json:"type"`
	Profile           string                  `json:"profile,omitempty"`
	Interactions      []string                `json:"interactions"`
	SearchParams      []SearchParamCapability  `json:"searchParams,omitempty"`
	Operations        []OperationCapability   `json:"operations,omitempty"`
	Versioning        string                  `json:"versioning"`
	ConditionalCreate bool                    `json:"conditionalCreate,omitempty"`
	ConditionalUpdate bool                    `json:"conditionalUpdate,omitempty"`
	ConditionalDelete bool                    `json:"conditionalDelete,omitempty"`
	SearchInclude     []string                `json:"searchInclude,omitempty"`
	SearchRevInclude  []string                `json:"searchRevInclude,omitempty"`
}

// CustomSearchParam defines a user-defined search parameter.
type CustomSearchParam struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	ResourceType string `json:"resourceType"`
	Expression   string `json:"expression"`
	Description  string `json:"description"`
}

// ---------------------------------------------------------------------------
// resourceEntry – internal accumulator for legacy AddResource API
// ---------------------------------------------------------------------------

type resourceEntry struct {
	resourceType      string
	interactions      []string
	searchParams      []SearchParam
	profiles          []string
	searchInclude     []string
	searchRevInclude  []string
	conditionalCreate bool
	conditionalUpdate bool
	conditionalDelete string // "not-supported", "single", "multiple"
	readHistory       bool
	updateCreate      bool
	patchFormats      []string

	// Extended fields from AddResourceCapability
	profile    string // canonical profile URL
	versioning string
	operations []OperationCapability
}

// ---------------------------------------------------------------------------
// CapabilityBuilder
// ---------------------------------------------------------------------------

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

	// Extended config (set via NewCapabilityBuilderFromConfig)
	config CapabilityConfig

	// Server-level operations
	serverOperations []OperationCapability

	// System-level interactions
	systemInteractions []string

	// Custom search parameters keyed by resource type.
	customSearchParams map[string][]CustomSearchParam
}

// NewCapabilityBuilder creates a new builder. The baseURL is the FHIR server
// base URL (e.g., "http://localhost:8000/fhir"), and version is the server
// software version.
func NewCapabilityBuilder(baseURL, version string) *CapabilityBuilder {
	return &CapabilityBuilder{
		resources: make(map[string]*resourceEntry),
		config: CapabilityConfig{
			ServerName:    "Headless EHR",
			ServerVersion: version,
			BaseURL:       baseURL,
			FHIRVersion:   "4.0.1",
		},
		ServerVersion:      version,
		BaseURL:            baseURL,
		customSearchParams: make(map[string][]CustomSearchParam),
	}
}

// NewCapabilityBuilderFromConfig creates a CapabilityBuilder from a
// CapabilityConfig, applying defaults for any empty fields.
func NewCapabilityBuilderFromConfig(cfg CapabilityConfig) *CapabilityBuilder {
	if cfg.ServerName == "" {
		cfg.ServerName = "Headless EHR FHIR Server"
	}
	if cfg.FHIRVersion == "" {
		cfg.FHIRVersion = "4.0.1"
	}
	if len(cfg.SupportedFormats) == 0 {
		cfg.SupportedFormats = []string{"application/fhir+json"}
	}
	if len(cfg.SupportedVersions) == 0 {
		cfg.SupportedVersions = []string{"4.0.1"}
	}

	return &CapabilityBuilder{
		resources:          make(map[string]*resourceEntry),
		config:             cfg,
		ServerVersion:      cfg.ServerVersion,
		BaseURL:            cfg.BaseURL,
		customSearchParams: make(map[string][]CustomSearchParam),
	}
}

// ---------------------------------------------------------------------------
// OAuth helpers
// ---------------------------------------------------------------------------

// SetOAuthURIs configures the SMART on FHIR OAuth URIs included in the
// security section of the CapabilityStatement.
func (b *CapabilityBuilder) SetOAuthURIs(authorizeURL, tokenURL string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.AuthorizeURL = authorizeURL
	b.TokenURL = tokenURL
}

// ---------------------------------------------------------------------------
// Legacy AddResource API (backward compatible)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Enhanced AddResourceCapability API
// ---------------------------------------------------------------------------

// AddResourceCapability registers a resource using the richer ResourceCapability struct.
func (b *CapabilityBuilder) AddResourceCapability(cap ResourceCapabilityDef) {
	// Convert SearchParamCapability to SearchParam for merging
	sp := make([]SearchParam, len(cap.SearchParams))
	for i, p := range cap.SearchParams {
		sp[i] = SearchParam{
			Name:          p.Name,
			Type:          p.Type,
			Documentation: p.Documentation,
		}
	}
	b.AddResource(cap.Type, cap.Interactions, sp)

	b.mu.Lock()
	defer b.mu.Unlock()
	entry := b.resources[cap.Type]
	if cap.Profile != "" {
		entry.profile = cap.Profile
	}
	if cap.Versioning != "" {
		entry.versioning = cap.Versioning
	}
	if len(cap.Operations) > 0 {
		entry.operations = append(entry.operations, cap.Operations...)
	}
	if cap.ConditionalCreate {
		entry.conditionalCreate = true
	}
	if cap.ConditionalUpdate {
		entry.conditionalUpdate = true
	}
	if cap.ConditionalDelete {
		entry.conditionalDelete = "single"
	}
	if len(cap.SearchInclude) > 0 {
		entry.searchInclude = cap.SearchInclude
	}
	if len(cap.SearchRevInclude) > 0 {
		entry.searchRevInclude = cap.SearchRevInclude
	}
}

// ---------------------------------------------------------------------------
// Server operations & system interactions
// ---------------------------------------------------------------------------

// AddServerOperation adds a system-level operation (e.g., $export, $graphql).
func (b *CapabilityBuilder) AddServerOperation(op OperationCapability) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.serverOperations = append(b.serverOperations, op)
}

// SetSystemInteractions sets the system-level interaction codes
// (transaction, batch, search-system, history-system).
func (b *CapabilityBuilder) SetSystemInteractions(codes []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.systemInteractions = codes
}

// ---------------------------------------------------------------------------
// Custom search parameters
// ---------------------------------------------------------------------------

// AddCustomSearchParam registers a custom search parameter.
func (b *CapabilityBuilder) AddCustomSearchParam(param CustomSearchParam) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.customSearchParams[param.ResourceType] = append(b.customSearchParams[param.ResourceType], param)
}

// ListCustomSearchParams returns all custom search parameters for a resource type.
func (b *CapabilityBuilder) ListCustomSearchParams(resourceType string) []CustomSearchParam {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]CustomSearchParam, len(b.customSearchParams[resourceType]))
	copy(result, b.customSearchParams[resourceType])
	return result
}

// ListAllCustomSearchParams returns every custom search parameter across all
// resource types.
func (b *CapabilityBuilder) ListAllCustomSearchParams() []CustomSearchParam {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var all []CustomSearchParam
	for _, params := range b.customSearchParams {
		all = append(all, params...)
	}
	return all
}

// DeleteCustomSearchParam removes a custom search parameter. Returns an error
// if the parameter is not found.
func (b *CapabilityBuilder) DeleteCustomSearchParam(resourceType, name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	params, ok := b.customSearchParams[resourceType]
	if !ok {
		return fmt.Errorf("no custom search params for resource type %s", resourceType)
	}
	for i, p := range params {
		if p.Name == name {
			b.customSearchParams[resourceType] = append(params[:i], params[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("custom search param %q not found for resource type %s", name, resourceType)
}

// ---------------------------------------------------------------------------
// Build
// ---------------------------------------------------------------------------

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
		res := b.buildResourceEntry(entry, rt)
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

	// System-level operations
	if len(b.serverOperations) > 0 {
		ops := make([]map[string]interface{}, len(b.serverOperations))
		for i, op := range b.serverOperations {
			o := map[string]interface{}{
				"name":       op.Name,
				"definition": op.Definition,
			}
			if op.Documentation != "" {
				o["documentation"] = op.Documentation
			}
			ops[i] = o
		}
		rest["operation"] = ops
	}

	// System-level interactions
	if len(b.systemInteractions) > 0 {
		ia := make([]map[string]string, len(b.systemInteractions))
		for i, code := range b.systemInteractions {
			ia[i] = map[string]string{"code": code}
		}
		rest["interaction"] = ia
	}

	// Determine format list
	formats := b.config.SupportedFormats
	if len(formats) == 0 {
		formats = []string{"json"}
	}

	// Determine server name
	serverName := b.config.ServerName
	if serverName == "" {
		serverName = "Headless EHR"
	}

	// Determine description
	description := b.config.Description
	if description == "" {
		description = "Headless EHR FHIR R4 Server"
	}

	// Determine FHIR version
	fhirVersion := b.config.FHIRVersion
	if fhirVersion == "" {
		fhirVersion = "4.0.1"
	}

	cs := map[string]interface{}{
		"resourceType": "CapabilityStatement",
		"status":       "active",
		"date":         time.Now().UTC().Format("2006-01-02"),
		"kind":         "instance",
		"fhirVersion":  fhirVersion,
		"format":       formats,
		"instantiates": []string{
			"http://hl7.org/fhir/uv/bulkdata/CapabilityStatement/bulk-data",
		},
		"software": map[string]string{
			"name":    serverName,
			"version": b.config.ServerVersion,
		},
		"implementation": map[string]string{
			"description": description,
			"url":         b.config.BaseURL,
		},
		"rest": []map[string]interface{}{rest},
	}

	if b.config.Publisher != "" {
		cs["publisher"] = b.config.Publisher
	}

	return cs
}

// buildResourceEntry constructs the map for a single resource type.
func (b *CapabilityBuilder) buildResourceEntry(entry *resourceEntry, rt string) map[string]interface{} {
	versioning := entry.versioning
	if versioning == "" {
		versioning = "versioned"
	}

	res := map[string]interface{}{
		"type":         entry.resourceType,
		"versioning":   versioning,
		"readHistory":  entry.readHistory,
		"updateCreate": entry.updateCreate,
	}

	// Profile
	if entry.profile != "" {
		res["profile"] = entry.profile
	}

	// Conditional operations
	if entry.conditionalCreate {
		res["conditionalCreate"] = true
	}
	if entry.conditionalUpdate {
		res["conditionalUpdate"] = true
	}
	if entry.conditionalDelete != "" {
		res["conditionalDelete"] = entry.conditionalDelete
	}

	// Patch formats
	if len(entry.patchFormats) > 0 {
		res["patchFormats"] = entry.patchFormats
	}

	// Interactions
	if len(entry.interactions) > 0 {
		interactions := make([]map[string]string, len(entry.interactions))
		for i, code := range entry.interactions {
			interactions[i] = map[string]string{"code": code}
		}
		res["interaction"] = interactions
	}

	// Search parameters (built-in + custom)
	allParams := make([]SearchParam, len(entry.searchParams))
	copy(allParams, entry.searchParams)

	// Append custom search params for this resource type
	if customs, ok := b.customSearchParams[rt]; ok {
		for _, cp := range customs {
			allParams = append(allParams, SearchParam{
				Name:          cp.Name,
				Type:          cp.Type,
				Documentation: cp.Description,
			})
		}
	}

	if len(allParams) > 0 {
		params := make([]map[string]string, len(allParams))
		for i, sp := range allParams {
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

	// Search include/revinclude
	if len(entry.searchInclude) > 0 {
		res["searchInclude"] = entry.searchInclude
	}
	if len(entry.searchRevInclude) > 0 {
		res["searchRevInclude"] = entry.searchRevInclude
	}

	// Supported profiles (legacy)
	if len(entry.profiles) > 0 {
		res["supportedProfile"] = entry.profiles
	}

	// Resource-level operations
	if len(entry.operations) > 0 {
		ops := make([]map[string]interface{}, len(entry.operations))
		for i, op := range entry.operations {
			o := map[string]interface{}{
				"name":       op.Name,
				"definition": op.Definition,
			}
			if op.Documentation != "" {
				o["documentation"] = op.Documentation
			}
			ops[i] = o
		}
		res["operation"] = ops
	}

	return res
}

// buildSecurity creates the SMART on FHIR security section with OAuth extension
// and SMART capabilities per the FHIR spec.
func (b *CapabilityBuilder) buildSecurity() map[string]interface{} {
	service := map[string]interface{}{
		"coding": []map[string]string{
			{
				"system":  "http://terminology.hl7.org/CodeSystem/restful-security-service",
				"code":    "SMART-on-FHIR",
				"display": "SMART on FHIR",
			},
		},
	}

	security := map[string]interface{}{
		"cors":        true,
		"service":     []map[string]interface{}{service},
		"description": "OAuth2 using SMART on FHIR profile (see http://docs.smarthealthit.org)",
	}

	// Build extensions list
	var extensions []map[string]interface{}

	// Add SMART on FHIR OAuth URI extension if URIs are configured
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

		extensions = append(extensions, smartExtension)
	}

	// SMART capabilities extension — declares supported SMART on FHIR features
	smartCapabilities := []string{
		"launch-ehr",
		"launch-standalone",
		"client-public",
		"client-confidential-symmetric",
		"sso-openid-connect",
		"permission-patient",
		"permission-user",
		"context-ehr-patient",
	}
	for _, cap := range smartCapabilities {
		extensions = append(extensions, map[string]interface{}{
			"url":       "http://fhir-registry.smarthealthit.org/StructureDefinition/capabilities",
			"valueCode": cap,
		})
	}

	if len(extensions) > 0 {
		security["extension"] = extensions
	}

	return security
}

// ResourceCount returns the number of registered resource types.
func (b *CapabilityBuilder) ResourceCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.resources)
}

// GetResourceTypes returns a sorted list of registered resource type names.
func (b *CapabilityBuilder) GetResourceTypes() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	types := make([]string, 0, len(b.resources))
	for rt := range b.resources {
		types = append(types, rt)
	}
	sort.Strings(types)
	return types
}

// GetResourceEntry returns the built map for a single resource type, or nil
// if the type is not registered.
func (b *CapabilityBuilder) GetResourceEntry(resourceType string) map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	entry, ok := b.resources[resourceType]
	if !ok {
		return nil
	}
	return b.buildResourceEntry(entry, resourceType)
}

// GetServerOperations returns a copy of server-level operations.
func (b *CapabilityBuilder) GetServerOperations() []OperationCapability {
	b.mu.RLock()
	defer b.mu.RUnlock()
	ops := make([]OperationCapability, len(b.serverOperations))
	copy(ops, b.serverOperations)
	return ops
}

// ---------------------------------------------------------------------------
// Convenience interaction sets
// ---------------------------------------------------------------------------

// DefaultInteractions returns the standard set of CRUD interactions for a
// resource type. This is a convenience for domain modules that support all
// standard operations.
func DefaultInteractions() []string {
	return []string{"read", "vread", "search-type", "create", "update", "patch", "delete", "history-instance", "history-type"}
}

// ReadOnlyInteractions returns interactions for read-only resources.
func ReadOnlyInteractions() []string {
	return []string{"read", "vread", "search-type"}
}

// ---------------------------------------------------------------------------
// ResourceCapabilityOptions (legacy advanced flags)
// ---------------------------------------------------------------------------

// SetResourceCapabilities sets advanced capability flags for a resource type.
func (b *CapabilityBuilder) SetResourceCapabilities(resourceType string, opts ResourceCapabilityOptions) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry, ok := b.resources[resourceType]
	if !ok {
		return
	}
	entry.conditionalCreate = opts.ConditionalCreate
	entry.conditionalUpdate = opts.ConditionalUpdate
	entry.conditionalDelete = opts.ConditionalDelete
	entry.readHistory = opts.ReadHistory
	entry.updateCreate = opts.UpdateCreate
	entry.patchFormats = opts.PatchFormats
	entry.searchInclude = opts.SearchInclude
	entry.searchRevInclude = opts.SearchRevInclude
}

// ResourceCapabilityOptions defines the advanced capability flags for a resource type.
type ResourceCapabilityOptions struct {
	ConditionalCreate bool
	ConditionalUpdate bool
	ConditionalDelete string
	ReadHistory       bool
	UpdateCreate      bool
	PatchFormats      []string
	SearchInclude     []string
	SearchRevInclude  []string
}

// DefaultCapabilityOptions returns the standard capability options for a fully
// FHIR R4 conformant resource.
func DefaultCapabilityOptions() ResourceCapabilityOptions {
	return ResourceCapabilityOptions{
		ConditionalCreate: true,
		ConditionalUpdate: true,
		ConditionalDelete: "single",
		ReadHistory:       true,
		UpdateCreate:      false,
		PatchFormats: []string{
			"application/json-patch+json",
			"application/merge-patch+json",
		},
	}
}

// ---------------------------------------------------------------------------
// DefaultCapabilityBuilder
// ---------------------------------------------------------------------------

// DefaultCapabilityBuilder returns a CapabilityBuilder pre-populated with 35+
// resource types, realistic search parameters, and 12 server-level operations.
func DefaultCapabilityBuilder() *CapabilityBuilder {
	cfg := CapabilityConfig{
		ServerName:        "Headless EHR FHIR Server",
		ServerVersion:     "0.1.0",
		FHIRVersion:       "4.0.1",
		Publisher:         "Headless EHR",
		Description:       "Headless EHR FHIR R4 Server",
		BaseURL:           "http://localhost:8000/fhir",
		SupportedFormats:  []string{"application/fhir+json"},
		SupportedVersions: []string{"4.0.1"},
	}
	b := NewCapabilityBuilderFromConfig(cfg)

	// -- Resources --

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Patient",
		Profile: "http://hl7.org/fhir/StructureDefinition/Patient",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		ConditionalCreate: true,
		ConditionalUpdate: true,
		ConditionalDelete: true,
		SearchParams: []SearchParamCapability{
			{Name: "name", Type: "string", Documentation: "A server defined search that may match any of the string fields in HumanName"},
			{Name: "family", Type: "string", Documentation: "A portion of the family name of the patient"},
			{Name: "given", Type: "string", Documentation: "A portion of the given name of the patient"},
			{Name: "birthdate", Type: "date", Documentation: "The patient's date of birth"},
			{Name: "gender", Type: "token", Documentation: "Gender of the patient"},
			{Name: "identifier", Type: "token", Documentation: "A patient identifier"},
			{Name: "address", Type: "string", Documentation: "A server defined search that may match any of the string fields in Address"},
			{Name: "phone", Type: "token", Documentation: "A value in a phone contact"},
			{Name: "email", Type: "token", Documentation: "A value in an email contact"},
			{Name: "_id", Type: "token", Documentation: "The ID of the resource"},
			{Name: "_lastUpdated", Type: "date", Documentation: "When the resource version last changed"},
			{Name: "general-practitioner", Type: "reference", Documentation: "Patient's nominated general practitioner"},
			{Name: "organization", Type: "reference", Documentation: "The organization that is the custodian of the patient record"},
			{Name: "active", Type: "token", Documentation: "Whether the patient record is active"},
			{Name: "deceased", Type: "token", Documentation: "This patient has been marked as deceased, or has a death date"},
			{Name: "death-date", Type: "date", Documentation: "The date of death has been provided and satisfies this search value"},
			{Name: "language", Type: "token", Documentation: "Language code"},
			{Name: "address-city", Type: "string", Documentation: "A city specified in an address"},
			{Name: "address-state", Type: "string", Documentation: "A state specified in an address"},
			{Name: "address-postalcode", Type: "string", Documentation: "A postalCode specified in an address"},
		},
		Operations: []OperationCapability{
			{Name: "$everything", Definition: "http://hl7.org/fhir/OperationDefinition/Patient-everything"},
			{Name: "$match", Definition: "http://hl7.org/fhir/OperationDefinition/Patient-match"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Observation",
		Profile: "http://hl7.org/fhir/StructureDefinition/Observation",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "category", Type: "token"},
			{Name: "code", Type: "token"},
			{Name: "date", Type: "date"},
			{Name: "status", Type: "token"},
			{Name: "value-quantity", Type: "quantity"},
			{Name: "combo-code", Type: "token"},
			{Name: "component-code", Type: "token"},
			{Name: "encounter", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Encounter",
		Profile: "http://hl7.org/fhir/StructureDefinition/Encounter",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "date", Type: "date"},
			{Name: "status", Type: "token"},
			{Name: "class", Type: "token"},
			{Name: "type", Type: "token"},
			{Name: "participant", Type: "reference"},
			{Name: "location", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Condition",
		Profile: "http://hl7.org/fhir/StructureDefinition/Condition",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "code", Type: "token"},
			{Name: "clinical-status", Type: "token"},
			{Name: "verification-status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "onset-date", Type: "date"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "MedicationRequest",
		Profile: "http://hl7.org/fhir/StructureDefinition/MedicationRequest",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "intent", Type: "token"},
			{Name: "authoredon", Type: "date"},
			{Name: "medication", Type: "reference"},
			{Name: "requester", Type: "reference"},
			{Name: "encounter", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "DiagnosticReport",
		Profile: "http://hl7.org/fhir/StructureDefinition/DiagnosticReport",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "code", Type: "token"},
			{Name: "date", Type: "date"},
			{Name: "status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "encounter", Type: "reference"},
			{Name: "result", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Procedure",
		Profile: "http://hl7.org/fhir/StructureDefinition/Procedure",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "code", Type: "token"},
			{Name: "date", Type: "date"},
			{Name: "status", Type: "token"},
			{Name: "encounter", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "AllergyIntolerance",
		Profile: "http://hl7.org/fhir/StructureDefinition/AllergyIntolerance",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "clinical-status", Type: "token"},
			{Name: "type", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "criticality", Type: "token"},
			{Name: "code", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Immunization",
		Profile: "http://hl7.org/fhir/StructureDefinition/Immunization",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "date", Type: "date"},
			{Name: "status", Type: "token"},
			{Name: "vaccine-code", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "DocumentReference",
		Profile: "http://hl7.org/fhir/StructureDefinition/DocumentReference",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "subject", Type: "reference"},
			{Name: "type", Type: "token"},
			{Name: "date", Type: "date"},
			{Name: "status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "author", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Practitioner",
		Profile: "http://hl7.org/fhir/StructureDefinition/Practitioner",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "name", Type: "string"},
			{Name: "identifier", Type: "token"},
			{Name: "active", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Organization",
		Profile: "http://hl7.org/fhir/StructureDefinition/Organization",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "name", Type: "string"},
			{Name: "identifier", Type: "token"},
			{Name: "type", Type: "token"},
			{Name: "active", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Location",
		Profile: "http://hl7.org/fhir/StructureDefinition/Location",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "name", Type: "string"},
			{Name: "address", Type: "string"},
			{Name: "type", Type: "token"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	// Additional resource types (14-35)

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Medication",
		Profile: "http://hl7.org/fhir/StructureDefinition/Medication",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "code", Type: "token"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "MedicationAdministration",
		Profile: "http://hl7.org/fhir/StructureDefinition/MedicationAdministration",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "effective-time", Type: "date"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "MedicationDispense",
		Profile: "http://hl7.org/fhir/StructureDefinition/MedicationDispense",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "ServiceRequest",
		Profile: "http://hl7.org/fhir/StructureDefinition/ServiceRequest",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "code", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "CarePlan",
		Profile: "http://hl7.org/fhir/StructureDefinition/CarePlan",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "CareTeam",
		Profile: "http://hl7.org/fhir/StructureDefinition/CareTeam",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Goal",
		Profile: "http://hl7.org/fhir/StructureDefinition/Goal",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "lifecycle-status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Appointment",
		Profile: "http://hl7.org/fhir/StructureDefinition/Appointment",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "date", Type: "date"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Schedule",
		Profile: "http://hl7.org/fhir/StructureDefinition/Schedule",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "actor", Type: "reference"},
			{Name: "active", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Slot",
		Profile: "http://hl7.org/fhir/StructureDefinition/Slot",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "schedule", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "start", Type: "date"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Coverage",
		Profile: "http://hl7.org/fhir/StructureDefinition/Coverage",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Claim",
		Profile: "http://hl7.org/fhir/StructureDefinition/Claim",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Consent",
		Profile: "http://hl7.org/fhir/StructureDefinition/Consent",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "category", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Composition",
		Profile: "http://hl7.org/fhir/StructureDefinition/Composition",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "type", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Communication",
		Profile: "http://hl7.org/fhir/StructureDefinition/Communication",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Questionnaire",
		Profile: "http://hl7.org/fhir/StructureDefinition/Questionnaire",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "name", Type: "string"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "QuestionnaireResponse",
		Profile: "http://hl7.org/fhir/StructureDefinition/QuestionnaireResponse",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "questionnaire", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "ImmunizationRecommendation",
		Profile: "http://hl7.org/fhir/StructureDefinition/ImmunizationRecommendation",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "vaccine-type", Type: "token"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "FamilyMemberHistory",
		Profile: "http://hl7.org/fhir/StructureDefinition/FamilyMemberHistory",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "relationship", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "RelatedPerson",
		Profile: "http://hl7.org/fhir/StructureDefinition/RelatedPerson",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "relationship", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Provenance",
		Profile: "http://hl7.org/fhir/StructureDefinition/Provenance",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "target", Type: "reference"},
			{Name: "agent", Type: "reference"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Task",
		Profile: "http://hl7.org/fhir/StructureDefinition/Task",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "owner", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "intent", Type: "token"},
			{Name: "priority", Type: "token"},
			{Name: "code", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Device",
		Profile: "http://hl7.org/fhir/StructureDefinition/Device",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "type", Type: "token"},
			{Name: "manufacturer", Type: "string"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "ResearchStudy",
		Profile: "http://hl7.org/fhir/StructureDefinition/ResearchStudy",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "status", Type: "token"},
			{Name: "title", Type: "string"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Subscription",
		Profile: "http://hl7.org/fhir/StructureDefinition/Subscription",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "status", Type: "token"},
			{Name: "type", Type: "token"},
			{Name: "criteria", Type: "string"},
			{Name: "url", Type: "uri"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "ImagingStudy",
		Profile: "http://hl7.org/fhir/StructureDefinition/ImagingStudy",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	b.AddResourceCapability(ResourceCapabilityDef{
		Type:    "Specimen",
		Profile: "http://hl7.org/fhir/StructureDefinition/Specimen",
		Interactions: DefaultInteractions(),
		Versioning:   "versioned",
		SearchParams: []SearchParamCapability{
			{Name: "patient", Type: "reference"},
			{Name: "status", Type: "token"},
			{Name: "_id", Type: "token"},
		},
	})

	// -- Server-level operations --

	b.AddServerOperation(OperationCapability{
		Name:       "$export",
		Definition: "http://hl7.org/fhir/uv/bulkdata/OperationDefinition/export",
		Documentation: "FHIR Bulk Data Export",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$everything",
		Definition: "http://hl7.org/fhir/OperationDefinition/Patient-everything",
		Documentation: "Fetch a patient's complete record",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$validate",
		Definition: "http://hl7.org/fhir/OperationDefinition/Resource-validate",
		Documentation: "Validate a resource",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$match",
		Definition: "http://hl7.org/fhir/OperationDefinition/Patient-match",
		Documentation: "Patient matching",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$translate",
		Definition: "http://hl7.org/fhir/OperationDefinition/ConceptMap-translate",
		Documentation: "Concept translation",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$subsumes",
		Definition: "http://hl7.org/fhir/OperationDefinition/CodeSystem-subsumes",
		Documentation: "Subsumption testing",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$validate-code",
		Definition: "http://hl7.org/fhir/OperationDefinition/ValueSet-validate-code",
		Documentation: "Value set code validation",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$document",
		Definition: "http://hl7.org/fhir/OperationDefinition/Composition-document",
		Documentation: "Generate a document",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$process-message",
		Definition: "http://hl7.org/fhir/OperationDefinition/MessageHeader-process-message",
		Documentation: "Process a FHIR message",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$graphql",
		Definition: "http://hl7.org/fhir/OperationDefinition/Resource-graphql",
		Documentation: "Execute a GraphQL query",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$closure",
		Definition: "http://hl7.org/fhir/OperationDefinition/ConceptMap-closure",
		Documentation: "Closure table maintenance",
	})
	b.AddServerOperation(OperationCapability{
		Name:       "$import",
		Definition: "http://hl7.org/fhir/uv/bulkdata/OperationDefinition/import",
		Documentation: "Bulk data import",
	})

	// -- System-level interactions --
	b.SetSystemInteractions([]string{"transaction", "batch", "search-system", "history-system"})

	return b
}

// ---------------------------------------------------------------------------
// CapabilityHandler — Echo HTTP handler for metadata endpoints
// ---------------------------------------------------------------------------

// CapabilityHandler serves CapabilityStatement and related metadata endpoints.
type CapabilityHandler struct {
	builder *CapabilityBuilder
}

// NewCapabilityHandler creates a handler backed by the given builder.
func NewCapabilityHandler(builder *CapabilityBuilder) *CapabilityHandler {
	return &CapabilityHandler{builder: builder}
}

// RegisterRoutes registers all metadata endpoints on the provided Echo group.
func (h *CapabilityHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/metadata", h.GetMetadata)
	g.GET("/metadata/resources", h.ListResources)
	g.GET("/metadata/resources/:type", h.GetResourceCapability)
	g.GET("/metadata/operations", h.ListOperations)
	g.POST("/metadata/search-params", h.RegisterCustomSearchParam)
	g.GET("/metadata/search-params", h.ListCustomSearchParams)
	g.DELETE("/metadata/search-params/:type/:name", h.DeleteCustomSearchParam)
}

// GetMetadata returns the full CapabilityStatement.
func (h *CapabilityHandler) GetMetadata(c echo.Context) error {
	return c.JSON(http.StatusOK, h.builder.Build())
}

// ListResources returns a list of supported resource type names.
func (h *CapabilityHandler) ListResources(c echo.Context) error {
	types := h.builder.GetResourceTypes()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"resourceTypes": types,
	})
}

// GetResourceCapability returns capability details for a single resource type.
func (h *CapabilityHandler) GetResourceCapability(c echo.Context) error {
	rt := c.Param("type")
	entry := h.builder.GetResourceEntry(rt)
	if entry == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("resource type %q not found", rt),
		})
	}
	return c.JSON(http.StatusOK, entry)
}

// ListOperations returns all server-level operations.
func (h *CapabilityHandler) ListOperations(c echo.Context) error {
	ops := h.builder.GetServerOperations()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"operations": ops,
	})
}

// RegisterCustomSearchParam registers a new custom search parameter.
func (h *CapabilityHandler) RegisterCustomSearchParam(c echo.Context) error {
	var param CustomSearchParam
	if err := json.NewDecoder(c.Request().Body).Decode(&param); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		})
	}
	if param.Name == "" || param.Type == "" || param.ResourceType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name, type, and resourceType are required",
		})
	}
	h.builder.AddCustomSearchParam(param)
	return c.JSON(http.StatusCreated, param)
}

// ListCustomSearchParams returns all custom search parameters.
func (h *CapabilityHandler) ListCustomSearchParams(c echo.Context) error {
	all := h.builder.ListAllCustomSearchParams()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"customSearchParams": all,
	})
}

// DeleteCustomSearchParam removes a custom search parameter.
func (h *CapabilityHandler) DeleteCustomSearchParam(c echo.Context) error {
	rt := c.Param("type")
	name := c.Param("name")
	if err := h.builder.DeleteCustomSearchParam(rt, name); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}
	return c.NoContent(http.StatusNoContent)
}
