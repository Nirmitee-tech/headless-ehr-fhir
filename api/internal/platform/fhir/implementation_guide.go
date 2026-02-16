package fhir

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ============================================================================
// ImplementationGuide Resource
// ============================================================================

// ImplementationGuideResource represents a FHIR ImplementationGuide resource.
type ImplementationGuideResource struct {
	ResourceType string         `json:"resourceType"`
	ID           string         `json:"id,omitempty"`
	URL          string         `json:"url"`
	Version      string         `json:"version,omitempty"`
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
	Status       string         `json:"status"`
	FHIRVersion  []string       `json:"fhirVersion,omitempty"`
	Description  string         `json:"description,omitempty"`
	PackageID    string         `json:"packageId,omitempty"`
	DependsOn    []IGDependency `json:"dependsOn,omitempty"`
	Global       []IGGlobal     `json:"global,omitempty"`
	Definition   *IGDefinition  `json:"definition,omitempty"`
}

// IGDependency describes a dependency on another ImplementationGuide.
type IGDependency struct {
	URI     string `json:"uri"`
	Version string `json:"version,omitempty"`
}

// IGGlobal describes a global profile constraint applied to a resource type.
type IGGlobal struct {
	Type    string `json:"type"`
	Profile string `json:"profile"`
}

// IGDefinition holds the definition section of an ImplementationGuide.
type IGDefinition struct {
	Resource  []IGResource  `json:"resource,omitempty"`
	Page      *IGPage       `json:"page,omitempty"`
	Parameter []IGParameter `json:"parameter,omitempty"`
}

// IGResource describes a resource included in the ImplementationGuide.
type IGResource struct {
	Reference        map[string]string `json:"reference"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
	ExampleCanonical string            `json:"exampleCanonical,omitempty"`
}

// IGPage represents a page in the IG narrative. Pages can be nested recursively.
type IGPage struct {
	NameURL    string   `json:"nameUrl"`
	Title      string   `json:"title"`
	Generation string   `json:"generation"` // html, markdown, xml, generated
	Page       []IGPage `json:"page,omitempty"`
}

// IGParameter defines a build parameter for the IG publisher.
type IGParameter struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

// ============================================================================
// TerminologyCapabilities Resource
// ============================================================================

// TerminologyCapabilitiesResource represents a FHIR TerminologyCapabilities resource.
type TerminologyCapabilitiesResource struct {
	ResourceType string           `json:"resourceType"`
	ID           string           `json:"id,omitempty"`
	URL          string           `json:"url,omitempty"`
	Status       string           `json:"status"`
	Kind         string           `json:"kind"`
	Date         string           `json:"date"`
	Description  string           `json:"description,omitempty"`
	CodeSystem   []TCCodeSystem   `json:"codeSystem,omitempty"`
	Expansion    *TCExpansion     `json:"expansion,omitempty"`
	ValidateCode *TCValidateCode  `json:"validateCode,omitempty"`
	Translation  *TCTranslation   `json:"translation,omitempty"`
	Closure      *TCClosure       `json:"closure,omitempty"`
}

// TCCodeSystem describes a code system supported by the terminology server.
type TCCodeSystem struct {
	URI     string `json:"uri"`
	Version []struct {
		Code string `json:"code"`
	} `json:"version,omitempty"`
}

// TCExpansion describes the expansion capabilities.
type TCExpansion struct {
	Hierarchical bool `json:"hierarchical"`
	Paging       bool `json:"paging"`
	Incomplete   bool `json:"incomplete"`
}

// TCValidateCode describes the code validation capabilities.
type TCValidateCode struct {
	Translations bool `json:"translations"`
}

// TCTranslation describes the concept translation capabilities.
type TCTranslation struct {
	NeedsMap bool `json:"needsMap"`
}

// TCClosure describes the closure table maintenance capabilities.
type TCClosure struct {
	Translation bool `json:"translation"`
}

// ============================================================================
// Default Instances
// ============================================================================

// DefaultImplementationGuide returns a pre-built ImplementationGuide describing
// the Headless EHR server's conformance profile, including US Core 6.1.0 and
// FHIR R4 4.0.1 dependencies, key supported profiles, and server capabilities.
func DefaultImplementationGuide() *ImplementationGuideResource {
	return &ImplementationGuideResource{
		ResourceType: "ImplementationGuide",
		ID:           "headless-ehr-ig",
		URL:          "http://headless-ehr.example.org/ImplementationGuide/headless-ehr",
		Version:      "0.1.0",
		Name:         "HeadlessEHRImplementationGuide",
		Title:        "Headless EHR Server Implementation Guide",
		Status:       "active",
		FHIRVersion:  []string{"4.0.1"},
		Description:  "Implementation guide for the Headless EHR FHIR R4 server describing supported profiles, operations, and conformance requirements.",
		PackageID:    "org.example.headless-ehr",
		DependsOn: []IGDependency{
			{
				URI:     "http://hl7.org/fhir/us/core/ImplementationGuide/hl7.fhir.us.core",
				Version: "6.1.0",
			},
			{
				URI:     "http://hl7.org/fhir/ImplementationGuide/hl7.fhir.r4.core",
				Version: "4.0.1",
			},
		},
		Global: []IGGlobal{
			{Type: "Patient", Profile: "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"},
			{Type: "Condition", Profile: "http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition-problems-health-concerns"},
			{Type: "Observation", Profile: "http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-lab"},
		},
		Definition: &IGDefinition{
			Resource: []IGResource{
				{
					Reference:   map[string]string{"reference": "StructureDefinition/us-core-patient"},
					Name:        "US Core Patient Profile",
					Description: "Defines constraints on the Patient resource for use in US healthcare settings.",
				},
				{
					Reference:   map[string]string{"reference": "StructureDefinition/us-core-condition-problems-health-concerns"},
					Name:        "US Core Condition Profile",
					Description: "Defines constraints on the Condition resource for problems and health concerns.",
				},
				{
					Reference:   map[string]string{"reference": "StructureDefinition/us-core-observation-lab"},
					Name:        "US Core Laboratory Result Observation Profile",
					Description: "Defines constraints on the Observation resource for laboratory results.",
				},
				{
					Reference:   map[string]string{"reference": "StructureDefinition/us-core-encounter"},
					Name:        "US Core Encounter Profile",
					Description: "Defines constraints on the Encounter resource for US healthcare settings.",
				},
				{
					Reference:   map[string]string{"reference": "CapabilityStatement/headless-ehr"},
					Name:        "Headless EHR CapabilityStatement",
					Description: "Server capability statement describing supported resources and operations.",
				},
			},
			Page: &IGPage{
				NameURL:    "index.html",
				Title:      "Headless EHR Implementation Guide",
				Generation: "html",
				Page: []IGPage{
					{
						NameURL:    "profiles.html",
						Title:      "Profiles",
						Generation: "markdown",
					},
					{
						NameURL:    "operations.html",
						Title:      "Operations",
						Generation: "markdown",
					},
					{
						NameURL:    "capability-statement.html",
						Title:      "Capability Statement",
						Generation: "html",
					},
				},
			},
			Parameter: []IGParameter{
				{Code: "copyrightyear", Value: "2024+"},
				{Code: "releaselabel", Value: "CI Build"},
				{Code: "path-resource", Value: "input/resources"},
				{Code: "path-pages", Value: "input/pagecontent"},
				{Code: "path-tx-cache", Value: "input-cache/txcache"},
			},
		},
	}
}

// DefaultTerminologyCapabilities returns a TerminologyCapabilities resource
// describing the terminology services supported by the Headless EHR server.
func DefaultTerminologyCapabilities() *TerminologyCapabilitiesResource {
	return &TerminologyCapabilitiesResource{
		ResourceType: "TerminologyCapabilities",
		ID:           "headless-ehr-terminology",
		URL:          "http://headless-ehr.example.org/TerminologyCapabilities/headless-ehr",
		Status:       "active",
		Kind:         "instance",
		Date:         time.Now().UTC().Format("2006-01-02"),
		Description:  "Terminology capabilities for the Headless EHR FHIR R4 server including code system support, expansion, validation, translation, and closure operations.",
		CodeSystem: []TCCodeSystem{
			{URI: "http://snomed.info/sct"},
			{URI: "http://loinc.org"},
			{URI: "http://www.nlm.nih.gov/research/umls/rxnorm"},
			{URI: "http://hl7.org/fhir/sid/icd-10-cm"},
			{URI: "http://terminology.hl7.org/CodeSystem/v3-ActCode"},
			{URI: "http://terminology.hl7.org/CodeSystem/observation-category"},
			{URI: "http://terminology.hl7.org/CodeSystem/condition-clinical"},
			{URI: "http://terminology.hl7.org/CodeSystem/condition-ver-status"},
		},
		Expansion: &TCExpansion{
			Hierarchical: true,
			Paging:       true,
			Incomplete:   false,
		},
		ValidateCode: &TCValidateCode{
			Translations: true,
		},
		Translation: &TCTranslation{
			NeedsMap: true,
		},
		Closure: &TCClosure{
			Translation: true,
		},
	}
}

// ============================================================================
// ImplementationGuide Handler
// ============================================================================

// ImplementationGuideHandler serves ImplementationGuide resources via HTTP.
type ImplementationGuideHandler struct {
	mu     sync.RWMutex
	guides map[string]*ImplementationGuideResource
}

// NewImplementationGuideHandler creates a handler pre-loaded with the default IG.
func NewImplementationGuideHandler() *ImplementationGuideHandler {
	h := &ImplementationGuideHandler{
		guides: make(map[string]*ImplementationGuideResource),
	}
	defaultIG := DefaultImplementationGuide()
	h.guides[defaultIG.ID] = defaultIG
	return h
}

// RegisterRoutes registers ImplementationGuide routes on the provided Echo group.
func (h *ImplementationGuideHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/ImplementationGuide", h.List)
	g.GET("/ImplementationGuide/:id", h.Read)
}

// List handles GET /fhir/ImplementationGuide and returns a searchset Bundle.
func (h *ImplementationGuideHandler) List(c echo.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := make([]map[string]interface{}, 0, len(h.guides))
	for _, ig := range h.guides {
		entry := map[string]interface{}{
			"fullUrl":  ig.URL,
			"resource": ig,
		}
		entries = append(entries, entry)
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(entries),
		"entry":        entries,
	}
	return c.JSON(http.StatusOK, bundle)
}

// Read handles GET /fhir/ImplementationGuide/:id.
func (h *ImplementationGuideHandler) Read(c echo.Context) error {
	id := c.Param("id")

	h.mu.RLock()
	ig, ok := h.guides[id]
	h.mu.RUnlock()

	if !ok {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("ImplementationGuide", id))
	}
	return c.JSON(http.StatusOK, ig)
}

// AddGuide registers an additional ImplementationGuide with the handler.
func (h *ImplementationGuideHandler) AddGuide(ig *ImplementationGuideResource) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.guides[ig.ID] = ig
}

// ============================================================================
// TerminologyCapabilities Handler
// ============================================================================

// TerminologyCapabilitiesHandler serves the TerminologyCapabilities resource.
type TerminologyCapabilitiesHandler struct {
	capabilities *TerminologyCapabilitiesResource
}

// NewTerminologyCapabilitiesHandler creates a handler with the default
// terminology capabilities.
func NewTerminologyCapabilitiesHandler() *TerminologyCapabilitiesHandler {
	return &TerminologyCapabilitiesHandler{
		capabilities: DefaultTerminologyCapabilities(),
	}
}

// RegisterRoutes registers the TerminologyCapabilities endpoint.
func (h *TerminologyCapabilitiesHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/TerminologyCapabilities", h.Get)
}

// Get handles GET /fhir/TerminologyCapabilities.
func (h *TerminologyCapabilitiesHandler) Get(c echo.Context) error {
	return c.JSON(http.StatusOK, h.capabilities)
}
