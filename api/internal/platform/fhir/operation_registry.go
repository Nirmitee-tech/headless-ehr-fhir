package fhir

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// OperationDefinition types
// ---------------------------------------------------------------------------

// OperationParam describes a single input or output parameter for an
// OperationDefinition.
type OperationParam struct {
	Name          string `json:"name"`
	Use           string `json:"use"` // "in" or "out"
	Min           int    `json:"min"`
	Max           string `json:"max"` // "1", "*"
	Type          string `json:"type,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

// OperationDefinitionResource is the FHIR OperationDefinition resource
// representation used by the operation registry.
type OperationDefinitionResource struct {
	ResourceType string           `json:"resourceType"`
	ID           string           `json:"id,omitempty"`
	URL          string           `json:"url"`
	Name         string           `json:"name"`
	Title        string           `json:"title,omitempty"`
	Status       string           `json:"status"`
	Kind         string           `json:"kind"` // "operation" or "query"
	Code         string           `json:"code"` // e.g., "validate", "meta"
	System       bool             `json:"system"`
	Type         bool             `json:"type"`
	Instance     bool             `json:"instance"`
	Resource     []string         `json:"resource,omitempty"`
	Parameter    []OperationParam `json:"parameter,omitempty"`
	Description  string           `json:"description,omitempty"`
}

// ---------------------------------------------------------------------------
// OperationRegistry
// ---------------------------------------------------------------------------

// OperationRegistry is a thread-safe registry of FHIR OperationDefinition
// resources supported by the server.
type OperationRegistry struct {
	mu         sync.RWMutex
	operations map[string]*OperationDefinitionResource
}

// NewOperationRegistry creates an empty OperationRegistry.
func NewOperationRegistry() *OperationRegistry {
	return &OperationRegistry{
		operations: make(map[string]*OperationDefinitionResource),
	}
}

// Register adds an OperationDefinitionResource to the registry, keyed by its
// Code field.
func (r *OperationRegistry) Register(op *OperationDefinitionResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.operations[op.Code] = op
}

// Get retrieves an operation by code, returning nil if not found.
func (r *OperationRegistry) Get(code string) *OperationDefinitionResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.operations[code]
}

// List returns all registered operations sorted alphabetically by code.
func (r *OperationRegistry) List() []*OperationDefinitionResource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	codes := make([]string, 0, len(r.operations))
	for code := range r.operations {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	result := make([]*OperationDefinitionResource, 0, len(codes))
	for _, code := range codes {
		result = append(result, r.operations[code])
	}
	return result
}

// ---------------------------------------------------------------------------
// DefaultOperationRegistry
// ---------------------------------------------------------------------------

// DefaultOperationRegistry returns an OperationRegistry pre-populated with
// all standard FHIR operations supported by the server.
func DefaultOperationRegistry() *OperationRegistry {
	reg := NewOperationRegistry()

	// $validate — system, type, instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "validate",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-validate",
		Name:         "Validate",
		Title:        "Validate a resource",
		Status:       "active",
		Kind:         "operation",
		Code:         "validate",
		System:       true,
		Type:         true,
		Instance:     true,
		Description:  "Validate a resource against its structure definition and business rules",
		Parameter: []OperationParam{
			{Name: "resource", Use: "in", Min: 1, Max: "1", Type: "Resource", Documentation: "The resource to validate"},
			{Name: "mode", Use: "in", Min: 0, Max: "1", Type: "code", Documentation: "Validation mode: create, update, delete"},
			{Name: "profile", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "Profile to validate against"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "OperationOutcome", Documentation: "Validation results"},
		},
	})

	// $everything — Patient instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "everything",
		URL:          "http://hl7.org/fhir/OperationDefinition/Patient-everything",
		Name:         "Everything",
		Title:        "Fetch Patient Record",
		Status:       "active",
		Kind:         "operation",
		Code:         "everything",
		System:       false,
		Type:         false,
		Instance:     true,
		Resource:     []string{"Patient"},
		Description:  "Return all resources related to the patient",
		Parameter: []OperationParam{
			{Name: "start", Use: "in", Min: 0, Max: "1", Type: "date", Documentation: "Start date for filtering"},
			{Name: "end", Use: "in", Min: 0, Max: "1", Type: "date", Documentation: "End date for filtering"},
			{Name: "_type", Use: "in", Min: 0, Max: "*", Type: "code", Documentation: "Resource types to include"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Bundle", Documentation: "Bundle of matching resources"},
		},
	})

	// $export — system, Patient type, Group type
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "export",
		URL:          "http://hl7.org/fhir/uv/bulkdata/OperationDefinition/export",
		Name:         "Export",
		Title:        "Bulk Data Export",
		Status:       "active",
		Kind:         "operation",
		Code:         "export",
		System:       true,
		Type:         true,
		Instance:     false,
		Resource:     []string{"Patient", "Group"},
		Description:  "Export data from the server in bulk using the FHIR Bulk Data specification",
		Parameter: []OperationParam{
			{Name: "_outputFormat", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Output format (default: application/fhir+ndjson)"},
			{Name: "_since", Use: "in", Min: 0, Max: "1", Type: "instant", Documentation: "Only include resources modified after this instant"},
			{Name: "_type", Use: "in", Min: 0, Max: "*", Type: "string", Documentation: "Resource types to include"},
		},
	})

	// $expand — ValueSet type/instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "expand",
		URL:          "http://hl7.org/fhir/OperationDefinition/ValueSet-expand",
		Name:         "Expand",
		Title:        "Value Set Expansion",
		Status:       "active",
		Kind:         "operation",
		Code:         "expand",
		System:       false,
		Type:         true,
		Instance:     true,
		Resource:     []string{"ValueSet"},
		Description:  "Expand a value set to produce a list of codes",
		Parameter: []OperationParam{
			{Name: "url", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "Canonical URL of the value set"},
			{Name: "filter", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Text filter for expansion"},
			{Name: "count", Use: "in", Min: 0, Max: "1", Type: "integer", Documentation: "Number of codes to return"},
			{Name: "offset", Use: "in", Min: 0, Max: "1", Type: "integer", Documentation: "Offset for paging"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "ValueSet", Documentation: "Expanded value set"},
		},
	})

	// $lookup — CodeSystem type
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "lookup",
		URL:          "http://hl7.org/fhir/OperationDefinition/CodeSystem-lookup",
		Name:         "Lookup",
		Title:        "Concept Lookup",
		Status:       "active",
		Kind:         "operation",
		Code:         "lookup",
		System:       false,
		Type:         true,
		Instance:     false,
		Resource:     []string{"CodeSystem"},
		Description:  "Look up properties and designations for a code in a code system",
		Parameter: []OperationParam{
			{Name: "code", Use: "in", Min: 1, Max: "1", Type: "code", Documentation: "The code to look up"},
			{Name: "system", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "The code system URI"},
			{Name: "version", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "The code system version"},
			{Name: "name", Use: "out", Min: 1, Max: "1", Type: "string", Documentation: "Name of the code system"},
			{Name: "display", Use: "out", Min: 0, Max: "1", Type: "string", Documentation: "Display text for the code"},
		},
	})

	// $validate-code — ValueSet type/instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "validate-code",
		URL:          "http://hl7.org/fhir/OperationDefinition/ValueSet-validate-code",
		Name:         "ValidateCode",
		Title:        "Value Set Code Validation",
		Status:       "active",
		Kind:         "operation",
		Code:         "validate-code",
		System:       false,
		Type:         true,
		Instance:     true,
		Resource:     []string{"ValueSet"},
		Description:  "Validate that a code is in a value set",
		Parameter: []OperationParam{
			{Name: "url", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "Canonical URL of the value set"},
			{Name: "code", Use: "in", Min: 0, Max: "1", Type: "code", Documentation: "The code to validate"},
			{Name: "system", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "The code system URI"},
			{Name: "display", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "The display text"},
			{Name: "result", Use: "out", Min: 1, Max: "1", Type: "boolean", Documentation: "Whether the code is valid"},
			{Name: "message", Use: "out", Min: 0, Max: "1", Type: "string", Documentation: "Error message if invalid"},
		},
	})

	// $translate — ConceptMap type/instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "translate",
		URL:          "http://hl7.org/fhir/OperationDefinition/ConceptMap-translate",
		Name:         "Translate",
		Title:        "Concept Translation",
		Status:       "active",
		Kind:         "operation",
		Code:         "translate",
		System:       false,
		Type:         true,
		Instance:     true,
		Resource:     []string{"ConceptMap"},
		Description:  "Translate a code from one value set to another using a concept map",
		Parameter: []OperationParam{
			{Name: "url", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "Canonical URL of the concept map"},
			{Name: "code", Use: "in", Min: 0, Max: "1", Type: "code", Documentation: "The code to translate"},
			{Name: "system", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "The source code system URI"},
			{Name: "targetsystem", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "The target code system URI"},
			{Name: "result", Use: "out", Min: 1, Max: "1", Type: "boolean", Documentation: "Whether a translation was found"},
		},
	})

	// $subsumes — CodeSystem type
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "subsumes",
		URL:          "http://hl7.org/fhir/OperationDefinition/CodeSystem-subsumes",
		Name:         "Subsumes",
		Title:        "Subsumption Testing",
		Status:       "active",
		Kind:         "operation",
		Code:         "subsumes",
		System:       false,
		Type:         true,
		Instance:     false,
		Resource:     []string{"CodeSystem"},
		Description:  "Test the subsumption relationship between two codes",
		Parameter: []OperationParam{
			{Name: "codeA", Use: "in", Min: 1, Max: "1", Type: "code", Documentation: "The A code to test"},
			{Name: "codeB", Use: "in", Min: 1, Max: "1", Type: "code", Documentation: "The B code to test"},
			{Name: "system", Use: "in", Min: 0, Max: "1", Type: "uri", Documentation: "The code system URI"},
			{Name: "outcome", Use: "out", Min: 1, Max: "1", Type: "code", Documentation: "Subsumption outcome: equivalent, subsumes, subsumed-by, not-subsumed"},
		},
	})

	// $match — Patient type
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "match",
		URL:          "http://hl7.org/fhir/OperationDefinition/Patient-match",
		Name:         "Match",
		Title:        "Patient Matching",
		Status:       "active",
		Kind:         "operation",
		Code:         "match",
		System:       false,
		Type:         true,
		Instance:     false,
		Resource:     []string{"Patient"},
		Description:  "Find patient records matching the supplied demographics",
		Parameter: []OperationParam{
			{Name: "resource", Use: "in", Min: 1, Max: "1", Type: "Resource", Documentation: "Patient resource with demographics to match"},
			{Name: "onlyCertainMatches", Use: "in", Min: 0, Max: "1", Type: "boolean", Documentation: "Only return certain matches"},
			{Name: "count", Use: "in", Min: 0, Max: "1", Type: "integer", Documentation: "Maximum number of matches"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Bundle", Documentation: "Bundle of matching patients with search scores"},
		},
	})

	// $meta — instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "meta",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-meta",
		Name:         "Meta",
		Title:        "Access Meta Information",
		Status:       "active",
		Kind:         "operation",
		Code:         "meta",
		System:       false,
		Type:         false,
		Instance:     true,
		Description:  "Retrieve the meta information for a resource instance",
		Parameter: []OperationParam{
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Meta", Documentation: "The meta information"},
		},
	})

	// $meta-add — instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "meta-add",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-meta-add",
		Name:         "MetaAdd",
		Title:        "Add Metadata Tags",
		Status:       "active",
		Kind:         "operation",
		Code:         "meta-add",
		System:       false,
		Type:         false,
		Instance:     true,
		Description:  "Add tags, security labels, and profiles to a resource",
		Parameter: []OperationParam{
			{Name: "meta", Use: "in", Min: 1, Max: "1", Type: "Meta", Documentation: "Meta information to add"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Meta", Documentation: "Updated meta information"},
		},
	})

	// $meta-delete — instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "meta-delete",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-meta-delete",
		Name:         "MetaDelete",
		Title:        "Remove Metadata Tags",
		Status:       "active",
		Kind:         "operation",
		Code:         "meta-delete",
		System:       false,
		Type:         false,
		Instance:     true,
		Description:  "Remove tags, security labels, and profiles from a resource",
		Parameter: []OperationParam{
			{Name: "meta", Use: "in", Min: 1, Max: "1", Type: "Meta", Documentation: "Meta information to remove"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Meta", Documentation: "Updated meta information"},
		},
	})

	// $diff — instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "diff",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-diff",
		Name:         "Diff",
		Title:        "Resource Diff",
		Status:       "active",
		Kind:         "operation",
		Code:         "diff",
		System:       false,
		Type:         false,
		Instance:     true,
		Description:  "Generate a diff between two versions of a resource",
		Parameter: []OperationParam{
			{Name: "from", Use: "in", Min: 0, Max: "1", Type: "id", Documentation: "Version ID to diff from"},
			{Name: "to", Use: "in", Min: 0, Max: "1", Type: "id", Documentation: "Version ID to diff to"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Parameters", Documentation: "Diff result as a Parameters resource"},
		},
	})

	// $lastn — Observation type
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "lastn",
		URL:          "http://hl7.org/fhir/OperationDefinition/Observation-lastn",
		Name:         "Lastn",
		Title:        "Last N Observations",
		Status:       "active",
		Kind:         "operation",
		Code:         "lastn",
		System:       false,
		Type:         true,
		Instance:     false,
		Resource:     []string{"Observation"},
		Description:  "Return the last N observations matching specified criteria",
		Parameter: []OperationParam{
			{Name: "max", Use: "in", Min: 0, Max: "1", Type: "positiveInt", Documentation: "Maximum number of observations per group (default 1)"},
			{Name: "subject", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Subject reference"},
			{Name: "category", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Category filter"},
			{Name: "code", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Code filter"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Bundle", Documentation: "Bundle of matching observations"},
		},
	})

	// $stats — Observation type
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "stats",
		URL:          "http://hl7.org/fhir/OperationDefinition/Observation-stats",
		Name:         "Stats",
		Title:        "Observation Statistics",
		Status:       "active",
		Kind:         "operation",
		Code:         "stats",
		System:       false,
		Type:         true,
		Instance:     false,
		Resource:     []string{"Observation"},
		Description:  "Compute statistical summaries for observation data",
		Parameter: []OperationParam{
			{Name: "subject", Use: "in", Min: 1, Max: "1", Type: "uri", Documentation: "Subject reference"},
			{Name: "code", Use: "in", Min: 1, Max: "*", Type: "string", Documentation: "Observation code(s)"},
			{Name: "period", Use: "in", Min: 1, Max: "1", Type: "Period", Documentation: "Time period for statistics"},
			{Name: "statistic", Use: "in", Min: 1, Max: "*", Type: "code", Documentation: "Statistics to compute: average, min, max, count"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Observation", Documentation: "Statistical observation result"},
		},
	})

	// $convert — system
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "convert",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-convert",
		Name:         "Convert",
		Title:        "Resource Conversion",
		Status:       "active",
		Kind:         "operation",
		Code:         "convert",
		System:       true,
		Type:         false,
		Instance:     false,
		Description:  "Convert a resource between FHIR versions or formats",
		Parameter: []OperationParam{
			{Name: "input", Use: "in", Min: 1, Max: "1", Type: "Resource", Documentation: "The resource to convert"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Resource", Documentation: "The converted resource"},
		},
	})

	// $graph — system
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "graph",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-graph",
		Name:         "Graph",
		Title:        "Return a Graph of Resources",
		Status:       "active",
		Kind:         "operation",
		Code:         "graph",
		System:       true,
		Type:         false,
		Instance:     false,
		Description:  "Return a graph of resources based on a GraphDefinition",
		Parameter: []OperationParam{
			{Name: "graph", Use: "in", Min: 1, Max: "1", Type: "uri", Documentation: "Canonical reference to a GraphDefinition"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Bundle", Documentation: "Bundle containing the graph of resources"},
		},
	})

	// $batch-validate — system
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "batch-validate",
		URL:          "http://hl7.org/fhir/OperationDefinition/Resource-batch-validate",
		Name:         "BatchValidate",
		Title:        "Batch Validate Resources",
		Status:       "active",
		Kind:         "operation",
		Code:         "batch-validate",
		System:       true,
		Type:         false,
		Instance:     false,
		Description:  "Validate multiple resources in a single request",
		Parameter: []OperationParam{
			{Name: "resource", Use: "in", Min: 1, Max: "*", Type: "Resource", Documentation: "Resources to validate"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Bundle", Documentation: "Bundle of OperationOutcome resources"},
		},
	})

	// $document — Composition instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "document",
		URL:          "http://hl7.org/fhir/OperationDefinition/Composition-document",
		Name:         "Document",
		Title:        "Generate Document",
		Status:       "active",
		Kind:         "operation",
		Code:         "document",
		System:       false,
		Type:         false,
		Instance:     true,
		Resource:     []string{"Composition"},
		Description:  "Generate a FHIR document bundle from a Composition",
		Parameter: []OperationParam{
			{Name: "persist", Use: "in", Min: 0, Max: "1", Type: "boolean", Documentation: "Whether to persist the generated document"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "Bundle", Documentation: "The generated document bundle"},
		},
	})

	// $closure — system
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "closure",
		URL:          "http://hl7.org/fhir/OperationDefinition/ConceptMap-closure",
		Name:         "Closure",
		Title:        "Closure Table Maintenance",
		Status:       "active",
		Kind:         "operation",
		Code:         "closure",
		System:       true,
		Type:         false,
		Instance:     false,
		Description:  "Maintain a client-side closure table for a code system",
		Parameter: []OperationParam{
			{Name: "name", Use: "in", Min: 1, Max: "1", Type: "string", Documentation: "Name of the closure table"},
			{Name: "concept", Use: "in", Min: 0, Max: "*", Type: "Coding", Documentation: "Concepts to add to the closure table"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "ConceptMap", Documentation: "Updated closure table as a ConceptMap"},
		},
	})

	// $apply — PlanDefinition instance
	reg.Register(&OperationDefinitionResource{
		ResourceType: "OperationDefinition",
		ID:           "apply",
		URL:          "http://hl7.org/fhir/OperationDefinition/PlanDefinition-apply",
		Name:         "Apply",
		Title:        "Apply PlanDefinition",
		Status:       "active",
		Kind:         "operation",
		Code:         "apply",
		System:       false,
		Type:         false,
		Instance:     true,
		Resource:     []string{"PlanDefinition"},
		Description:  "Apply a PlanDefinition to generate a CarePlan or RequestGroup",
		Parameter: []OperationParam{
			{Name: "subject", Use: "in", Min: 1, Max: "*", Type: "string", Documentation: "Subject(s) to apply the plan to"},
			{Name: "encounter", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Encounter context for the plan"},
			{Name: "practitioner", Use: "in", Min: 0, Max: "1", Type: "string", Documentation: "Practitioner applying the plan"},
			{Name: "return", Use: "out", Min: 1, Max: "1", Type: "CarePlan", Documentation: "Generated CarePlan or RequestGroup"},
		},
	})

	return reg
}

// ---------------------------------------------------------------------------
// OperationRegistryHandler — HTTP handler for OperationDefinition endpoints
// ---------------------------------------------------------------------------

// OperationRegistryHandler serves FHIR OperationDefinition endpoints backed
// by an OperationRegistry.
type OperationRegistryHandler struct {
	registry *OperationRegistry
}

// NewOperationRegistryHandler creates a handler backed by the given registry.
func NewOperationRegistryHandler(registry *OperationRegistry) *OperationRegistryHandler {
	return &OperationRegistryHandler{registry: registry}
}

// RegisterRoutes registers OperationDefinition endpoints on the provided
// Echo group. The group is typically mounted at /fhir.
func (h *OperationRegistryHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/OperationDefinition", h.Search)
	g.GET("/OperationDefinition/:id", h.Read)
}

// Search handles GET /fhir/OperationDefinition, returning a searchset Bundle
// of all registered operation definitions. The optional query parameters
// "name", "code", "system", "type", "instance", and "resource" can be used
// to filter results.
func (h *OperationRegistryHandler) Search(c echo.Context) error {
	ops := h.registry.List()

	// Apply query-parameter filters.
	nameFilter := c.QueryParam("name")
	codeFilter := c.QueryParam("code")
	systemFilter := c.QueryParam("system")
	typeFilter := c.QueryParam("type")
	instanceFilter := c.QueryParam("instance")
	resourceFilter := c.QueryParam("resource")

	filtered := make([]*OperationDefinitionResource, 0, len(ops))
	for _, op := range ops {
		if nameFilter != "" && !strings.EqualFold(op.Name, nameFilter) {
			continue
		}
		if codeFilter != "" && op.Code != codeFilter {
			continue
		}
		if systemFilter == "true" && !op.System {
			continue
		}
		if systemFilter == "false" && op.System {
			continue
		}
		if typeFilter == "true" && !op.Type {
			continue
		}
		if typeFilter == "false" && op.Type {
			continue
		}
		if instanceFilter == "true" && !op.Instance {
			continue
		}
		if instanceFilter == "false" && op.Instance {
			continue
		}
		if resourceFilter != "" {
			found := false
			for _, r := range op.Resource {
				if strings.EqualFold(r, resourceFilter) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, op)
	}

	// Build a searchset Bundle.
	entries := make([]map[string]interface{}, 0, len(filtered))
	for _, op := range filtered {
		entries = append(entries, map[string]interface{}{
			"resource": op,
			"search": map[string]string{
				"mode": "match",
			},
		})
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(entries),
		"entry":        entries,
	}

	return c.JSON(http.StatusOK, bundle)
}

// Read handles GET /fhir/OperationDefinition/:id, returning a single
// OperationDefinitionResource or a 404 OperationOutcome.
func (h *OperationRegistryHandler) Read(c echo.Context) error {
	id := c.Param("id")

	op := h.registry.Get(id)
	if op == nil {
		return c.JSON(http.StatusNotFound, ErrorOutcome("OperationDefinition/"+id+" not found"))
	}

	return c.JSON(http.StatusOK, op)
}
