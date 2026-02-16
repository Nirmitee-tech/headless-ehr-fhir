package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// SearchParameterResource
// ---------------------------------------------------------------------------

// SearchParameterResource represents a FHIR SearchParameter resource that
// defines a search parameter and its properties for use in search operations.
type SearchParameterResource struct {
	ResourceType string   `json:"resourceType"`
	ID           string   `json:"id,omitempty"`
	URL          string   `json:"url"`
	Name         string   `json:"name"`
	Status       string   `json:"status"` // draft, active, retired
	Description  string   `json:"description,omitempty"`
	Code         string   `json:"code"` // name used in search URL
	Base         []string `json:"base"` // resource types this applies to
	Type         string   `json:"type"` // number, date, string, token, reference, composite, quantity, uri, special
	Expression   string   `json:"expression,omitempty"` // FHIRPath expression
	XPath        string   `json:"xpath,omitempty"`
	Target       []string `json:"target,omitempty"` // for reference type params
	Comparator   []string `json:"comparator,omitempty"` // eq, ne, gt, lt, ge, le, sa, eb, ap
	Modifier     []string `json:"modifier,omitempty"` // missing, exact, contains, text, etc.
	MultipleOr   *bool    `json:"multipleOr,omitempty"`
	MultipleAnd  *bool    `json:"multipleAnd,omitempty"`
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool {
	return &b
}

// validSearchParamTypes enumerates the allowed SearchParameter.type values.
var validSearchParamTypes = map[string]bool{
	"number":    true,
	"date":      true,
	"string":    true,
	"token":     true,
	"reference": true,
	"composite": true,
	"quantity":  true,
	"uri":       true,
	"special":   true,
}

// validSearchParamStatuses enumerates the allowed SearchParameter.status values.
var validSearchParamStatuses = map[string]bool{
	"draft":   true,
	"active":  true,
	"retired": true,
}

// ---------------------------------------------------------------------------
// SearchParameterStore
// ---------------------------------------------------------------------------

// SearchParameterStore is a thread-safe in-memory store for FHIR
// SearchParameter resources, keyed by ID.
type SearchParameterStore struct {
	mu     sync.RWMutex
	params map[string]*SearchParameterResource // keyed by ID
}

// NewSearchParameterStore creates an empty SearchParameterStore.
func NewSearchParameterStore() *SearchParameterStore {
	return &SearchParameterStore{
		params: make(map[string]*SearchParameterResource),
	}
}

// Create adds a new SearchParameterResource to the store. It returns an error
// if the resource has no ID or if a resource with the same ID already exists.
func (s *SearchParameterStore) Create(sp *SearchParameterResource) error {
	if sp.ID == "" {
		return fmt.Errorf("SearchParameter ID is required")
	}
	if sp.ResourceType == "" {
		sp.ResourceType = "SearchParameter"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.params[sp.ID]; exists {
		return fmt.Errorf("SearchParameter with ID %q already exists", sp.ID)
	}

	// Store a copy to avoid external mutation.
	stored := *sp
	s.params[sp.ID] = &stored
	return nil
}

// Get retrieves a SearchParameterResource by ID. It returns an error if the
// resource is not found.
func (s *SearchParameterStore) Get(id string) (*SearchParameterResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sp, ok := s.params[id]
	if !ok {
		return nil, fmt.Errorf("SearchParameter/%s not found", id)
	}

	// Return a copy.
	result := *sp
	return &result, nil
}

// Update replaces an existing SearchParameterResource in the store. It returns
// an error if the resource does not exist.
func (s *SearchParameterStore) Update(id string, sp *SearchParameterResource) error {
	if sp.ResourceType == "" {
		sp.ResourceType = "SearchParameter"
	}
	sp.ID = id

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.params[id]; !exists {
		return fmt.Errorf("SearchParameter/%s not found", id)
	}

	stored := *sp
	s.params[id] = &stored
	return nil
}

// Delete removes a SearchParameterResource from the store. It returns an
// error if the resource does not exist.
func (s *SearchParameterStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.params[id]; !exists {
		return fmt.Errorf("SearchParameter/%s not found", id)
	}

	delete(s.params, id)
	return nil
}

// Search returns all SearchParameterResources that match the given filter
// parameters. Supported filter keys: "name", "code", "url", "status",
// "type", "base". An empty filter map returns all resources. Results are
// sorted alphabetically by ID for deterministic output.
func (s *SearchParameterStore) Search(params map[string]string) []*SearchParameterResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect and sort keys for deterministic iteration.
	ids := make([]string, 0, len(s.params))
	for id := range s.params {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	results := make([]*SearchParameterResource, 0, len(s.params))
	for _, id := range ids {
		sp := s.params[id]
		if !matchSearchParam(sp, params) {
			continue
		}
		// Return copies.
		cp := *sp
		results = append(results, &cp)
	}
	return results
}

// List returns the total number of SearchParameterResources in the store.
func (s *SearchParameterStore) List() []*SearchParameterResource {
	return s.Search(nil)
}

// matchSearchParam checks whether a SearchParameterResource matches the
// given filter criteria.
func matchSearchParam(sp *SearchParameterResource, params map[string]string) bool {
	if params == nil {
		return true
	}

	if v, ok := params["name"]; ok && !strings.EqualFold(sp.Name, v) {
		return false
	}
	if v, ok := params["code"]; ok && sp.Code != v {
		return false
	}
	if v, ok := params["url"]; ok && sp.URL != v {
		return false
	}
	if v, ok := params["status"]; ok && sp.Status != v {
		return false
	}
	if v, ok := params["type"]; ok && sp.Type != v {
		return false
	}
	if v, ok := params["base"]; ok {
		found := false
		for _, b := range sp.Base {
			if strings.EqualFold(b, v) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// DefaultSearchParameters — common FHIR R4 search parameters
// ---------------------------------------------------------------------------

// DefaultSearchParameters returns a set of common FHIR R4 search parameters
// that are pre-registered on server startup. This includes cross-resource
// parameters (prefixed with _) and resource-specific parameters for commonly
// used resource types.
func DefaultSearchParameters() []*SearchParameterResource {
	t := true
	return []*SearchParameterResource{
		// ---------------------------------------------------------------
		// Cross-resource (Resource / DomainResource) search parameters
		// ---------------------------------------------------------------
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-id",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-id",
			Name:         "ResourceId",
			Status:       "active",
			Description:  "Logical id of this artifact",
			Code:         "_id",
			Base:         []string{"Resource"},
			Type:         "token",
			Expression:   "Resource.id",
			MultipleOr:   &t,
			MultipleAnd:  &t,
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-lastUpdated",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-lastUpdated",
			Name:         "ResourceLastUpdated",
			Status:       "active",
			Description:  "When the resource version last changed",
			Code:         "_lastUpdated",
			Base:         []string{"Resource"},
			Type:         "date",
			Expression:   "Resource.meta.lastUpdated",
			Comparator:   []string{"eq", "ne", "gt", "lt", "ge", "le", "sa", "eb", "ap"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-tag",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-tag",
			Name:         "ResourceTag",
			Status:       "active",
			Description:  "Tags applied to this resource",
			Code:         "_tag",
			Base:         []string{"Resource"},
			Type:         "token",
			Expression:   "Resource.meta.tag",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-security",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-security",
			Name:         "ResourceSecurity",
			Status:       "active",
			Description:  "Security Labels applied to this resource",
			Code:         "_security",
			Base:         []string{"Resource"},
			Type:         "token",
			Expression:   "Resource.meta.security",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-profile",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-profile",
			Name:         "ResourceProfile",
			Status:       "active",
			Description:  "Profiles this resource claims to conform to",
			Code:         "_profile",
			Base:         []string{"Resource"},
			Type:         "uri",
			Expression:   "Resource.meta.profile",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-text",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-text",
			Name:         "ResourceText",
			Status:       "active",
			Description:  "Search on the narrative text of the resource",
			Code:         "_text",
			Base:         []string{"DomainResource"},
			Type:         "string",
			Modifier:     []string{"missing", "exact", "contains"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-content",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-content",
			Name:         "ResourceContent",
			Status:       "active",
			Description:  "Search on the entire content of the resource",
			Code:         "_content",
			Base:         []string{"Resource"},
			Type:         "string",
			Modifier:     []string{"missing", "exact", "contains"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-has",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-has",
			Name:         "ResourceHas",
			Status:       "active",
			Description:  "Provides limited support for reverse chaining",
			Code:         "_has",
			Base:         []string{"Resource"},
			Type:         "special",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-list",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-list",
			Name:         "ResourceList",
			Status:       "active",
			Description:  "Search for resources in a particular list",
			Code:         "_list",
			Base:         []string{"Resource"},
			Type:         "special",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Resource-source",
			URL:          "http://hl7.org/fhir/SearchParameter/Resource-source",
			Name:         "ResourceSource",
			Status:       "active",
			Description:  "Identifies where the resource comes from",
			Code:         "_source",
			Base:         []string{"Resource"},
			Type:         "uri",
			Expression:   "Resource.meta.source",
		},

		// ---------------------------------------------------------------
		// Patient search parameters
		// ---------------------------------------------------------------
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-name",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-name",
			Name:         "PatientName",
			Status:       "active",
			Description:  "A server defined search that may match any of the string fields in the HumanName",
			Code:         "name",
			Base:         []string{"Patient"},
			Type:         "string",
			Expression:   "Patient.name",
			Modifier:     []string{"missing", "exact", "contains"},
			MultipleOr:   &t,
			MultipleAnd:  &t,
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-family",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-family",
			Name:         "PatientFamily",
			Status:       "active",
			Description:  "A portion of the family name of the patient",
			Code:         "family",
			Base:         []string{"Patient"},
			Type:         "string",
			Expression:   "Patient.name.family",
			Modifier:     []string{"missing", "exact", "contains"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-given",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-given",
			Name:         "PatientGiven",
			Status:       "active",
			Description:  "A portion of the given name of the patient",
			Code:         "given",
			Base:         []string{"Patient"},
			Type:         "string",
			Expression:   "Patient.name.given",
			Modifier:     []string{"missing", "exact", "contains"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-birthdate",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-birthdate",
			Name:         "PatientBirthdate",
			Status:       "active",
			Description:  "The patient's date of birth",
			Code:         "birthdate",
			Base:         []string{"Patient"},
			Type:         "date",
			Expression:   "Patient.birthDate",
			Comparator:   []string{"eq", "ne", "gt", "lt", "ge", "le", "sa", "eb", "ap"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-gender",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-gender",
			Name:         "PatientGender",
			Status:       "active",
			Description:  "Gender of the patient",
			Code:         "gender",
			Base:         []string{"Patient"},
			Type:         "token",
			Expression:   "Patient.gender",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-identifier",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-identifier",
			Name:         "PatientIdentifier",
			Status:       "active",
			Description:  "A patient identifier",
			Code:         "identifier",
			Base:         []string{"Patient"},
			Type:         "token",
			Expression:   "Patient.identifier",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Patient-general-practitioner",
			URL:          "http://hl7.org/fhir/SearchParameter/Patient-general-practitioner",
			Name:         "PatientGeneralPractitioner",
			Status:       "active",
			Description:  "Patient's nominated general practitioner",
			Code:         "general-practitioner",
			Base:         []string{"Patient"},
			Type:         "reference",
			Expression:   "Patient.generalPractitioner",
			Target:       []string{"Organization", "Practitioner", "PractitionerRole"},
		},

		// ---------------------------------------------------------------
		// Observation search parameters
		// ---------------------------------------------------------------
		{
			ResourceType: "SearchParameter",
			ID:           "Observation-code",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-code",
			Name:         "ObservationCode",
			Status:       "active",
			Description:  "The code of the observation type",
			Code:         "code",
			Base:         []string{"Observation"},
			Type:         "token",
			Expression:   "Observation.code",
			MultipleOr:   &t,
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Observation-patient",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-patient",
			Name:         "ObservationPatient",
			Status:       "active",
			Description:  "The subject that the observation is about (if patient)",
			Code:         "patient",
			Base:         []string{"Observation"},
			Type:         "reference",
			Expression:   "Observation.subject.where(resolve() is Patient)",
			Target:       []string{"Patient"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Observation-category",
			URL:          "http://hl7.org/fhir/SearchParameter/Observation-category",
			Name:         "ObservationCategory",
			Status:       "active",
			Description:  "The classification of the type of observation",
			Code:         "category",
			Base:         []string{"Observation"},
			Type:         "token",
			Expression:   "Observation.category",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Observation-date",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-date",
			Name:         "ObservationDate",
			Status:       "active",
			Description:  "Obtained date/time",
			Code:         "date",
			Base:         []string{"Observation"},
			Type:         "date",
			Expression:   "Observation.effective",
			Comparator:   []string{"eq", "ne", "gt", "lt", "ge", "le", "sa", "eb", "ap"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Observation-status",
			URL:          "http://hl7.org/fhir/SearchParameter/Observation-status",
			Name:         "ObservationStatus",
			Status:       "active",
			Description:  "The status of the observation",
			Code:         "status",
			Base:         []string{"Observation"},
			Type:         "token",
			Expression:   "Observation.status",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Observation-value-quantity",
			URL:          "http://hl7.org/fhir/SearchParameter/Observation-value-quantity",
			Name:         "ObservationValueQuantity",
			Status:       "active",
			Description:  "The value of the observation, if the value is a Quantity",
			Code:         "value-quantity",
			Base:         []string{"Observation"},
			Type:         "quantity",
			Expression:   "(Observation.value as Quantity)",
			Comparator:   []string{"eq", "ne", "gt", "lt", "ge", "le", "sa", "eb", "ap"},
		},

		// ---------------------------------------------------------------
		// Encounter search parameters
		// ---------------------------------------------------------------
		{
			ResourceType: "SearchParameter",
			ID:           "Encounter-patient",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-patient",
			Name:         "EncounterPatient",
			Status:       "active",
			Description:  "The patient or group present at the encounter",
			Code:         "patient",
			Base:         []string{"Encounter"},
			Type:         "reference",
			Expression:   "Encounter.subject.where(resolve() is Patient)",
			Target:       []string{"Patient"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Encounter-status",
			URL:          "http://hl7.org/fhir/SearchParameter/Encounter-status",
			Name:         "EncounterStatus",
			Status:       "active",
			Description:  "Status of the encounter",
			Code:         "status",
			Base:         []string{"Encounter"},
			Type:         "token",
			Expression:   "Encounter.status",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Encounter-class",
			URL:          "http://hl7.org/fhir/SearchParameter/Encounter-class",
			Name:         "EncounterClass",
			Status:       "active",
			Description:  "Classification of patient encounter",
			Code:         "class",
			Base:         []string{"Encounter"},
			Type:         "token",
			Expression:   "Encounter.class",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Encounter-date",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-date",
			Name:         "EncounterDate",
			Status:       "active",
			Description:  "A date within the period the Encounter lasted",
			Code:         "date",
			Base:         []string{"Encounter"},
			Type:         "date",
			Expression:   "Encounter.period",
			Comparator:   []string{"eq", "ne", "gt", "lt", "ge", "le", "sa", "eb", "ap"},
		},

		// ---------------------------------------------------------------
		// Condition search parameters
		// ---------------------------------------------------------------
		{
			ResourceType: "SearchParameter",
			ID:           "Condition-code",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-code",
			Name:         "ConditionCode",
			Status:       "active",
			Description:  "Code for the condition",
			Code:         "code",
			Base:         []string{"Condition"},
			Type:         "token",
			Expression:   "Condition.code",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Condition-clinical-status",
			URL:          "http://hl7.org/fhir/SearchParameter/Condition-clinical-status",
			Name:         "ConditionClinicalStatus",
			Status:       "active",
			Description:  "The clinical status of the condition",
			Code:         "clinical-status",
			Base:         []string{"Condition"},
			Type:         "token",
			Expression:   "Condition.clinicalStatus",
		},
		{
			ResourceType: "SearchParameter",
			ID:           "Condition-patient",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-patient",
			Name:         "ConditionPatient",
			Status:       "active",
			Description:  "Who has the condition",
			Code:         "patient",
			Base:         []string{"Condition"},
			Type:         "reference",
			Expression:   "Condition.subject.where(resolve() is Patient)",
			Target:       []string{"Patient"},
		},

		// ---------------------------------------------------------------
		// MedicationRequest search parameters
		// ---------------------------------------------------------------
		{
			ResourceType: "SearchParameter",
			ID:           "MedicationRequest-patient",
			URL:          "http://hl7.org/fhir/SearchParameter/clinical-patient",
			Name:         "MedicationRequestPatient",
			Status:       "active",
			Description:  "Returns prescriptions for a specific patient",
			Code:         "patient",
			Base:         []string{"MedicationRequest"},
			Type:         "reference",
			Expression:   "MedicationRequest.subject.where(resolve() is Patient)",
			Target:       []string{"Patient"},
		},
		{
			ResourceType: "SearchParameter",
			ID:           "MedicationRequest-status",
			URL:          "http://hl7.org/fhir/SearchParameter/medications-status",
			Name:         "MedicationRequestStatus",
			Status:       "active",
			Description:  "Status of the prescription",
			Code:         "status",
			Base:         []string{"MedicationRequest"},
			Type:         "token",
			Expression:   "MedicationRequest.status",
		},
	}
}

// ---------------------------------------------------------------------------
// SearchParameterHandler — HTTP handler for SearchParameter endpoints
// ---------------------------------------------------------------------------

// SearchParameterHandler serves FHIR SearchParameter endpoints backed by a
// SearchParameterStore.
type SearchParameterHandler struct {
	store *SearchParameterStore
}

// NewSearchParameterHandler creates a handler backed by the given store.
func NewSearchParameterHandler(store *SearchParameterStore) *SearchParameterHandler {
	return &SearchParameterHandler{store: store}
}

// RegisterRoutes registers SearchParameter CRUD endpoints on the provided
// Echo group. The group is typically mounted at /fhir.
func (h *SearchParameterHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/SearchParameter", h.Search)
	g.GET("/SearchParameter/:id", h.Read)
	g.POST("/SearchParameter", h.Create)
	g.PUT("/SearchParameter/:id", h.Update)
	g.DELETE("/SearchParameter/:id", h.Delete)
}

// Search handles GET /fhir/SearchParameter, returning a searchset Bundle of
// matching SearchParameter resources. Supported query parameters: name, code,
// url, status, type, base.
func (h *SearchParameterHandler) Search(c echo.Context) error {
	filters := make(map[string]string)
	for _, key := range []string{"name", "code", "url", "status", "type", "base"} {
		if v := c.QueryParam(key); v != "" {
			filters[key] = v
		}
	}

	results := h.store.Search(filters)

	entries := make([]map[string]interface{}, 0, len(results))
	for _, sp := range results {
		entries = append(entries, map[string]interface{}{
			"resource": sp,
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

// Read handles GET /fhir/SearchParameter/:id, returning a single
// SearchParameter resource or a 404 OperationOutcome.
func (h *SearchParameterHandler) Read(c echo.Context) error {
	id := c.Param("id")

	sp, err := h.store.Get(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("SearchParameter", id))
	}

	return c.JSON(http.StatusOK, sp)
}

// Create handles POST /fhir/SearchParameter, creating a new SearchParameter
// resource. It validates required fields and returns the created resource with
// a 201 status.
func (h *SearchParameterHandler) Create(c echo.Context) error {
	var sp SearchParameterResource
	if err := json.NewDecoder(c.Request().Body).Decode(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	// Validate required fields.
	if err := validateSearchParameter(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	// Default resource type.
	sp.ResourceType = "SearchParameter"

	if err := h.store.Create(&sp); err != nil {
		return c.JSON(http.StatusConflict, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusCreated, &sp)
}

// Update handles PUT /fhir/SearchParameter/:id, updating an existing
// SearchParameter resource.
func (h *SearchParameterHandler) Update(c echo.Context) error {
	id := c.Param("id")

	var sp SearchParameterResource
	if err := json.NewDecoder(c.Request().Body).Decode(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	// Validate required fields.
	if err := validateSearchParameter(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	sp.ResourceType = "SearchParameter"

	if err := h.store.Update(id, &sp); err != nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("SearchParameter", id))
	}

	return c.JSON(http.StatusOK, &sp)
}

// Delete handles DELETE /fhir/SearchParameter/:id, removing a SearchParameter
// resource from the store.
func (h *SearchParameterHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	if err := h.store.Delete(id); err != nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("SearchParameter", id))
	}

	return c.NoContent(http.StatusNoContent)
}

// validateSearchParameter checks that a SearchParameterResource has the
// minimum required fields and valid enum values.
func validateSearchParameter(sp *SearchParameterResource) error {
	if sp.URL == "" {
		return fmt.Errorf("SearchParameter.url is required")
	}
	if sp.Name == "" {
		return fmt.Errorf("SearchParameter.name is required")
	}
	if sp.Status == "" {
		return fmt.Errorf("SearchParameter.status is required")
	}
	if !validSearchParamStatuses[sp.Status] {
		return fmt.Errorf("SearchParameter.status must be one of: draft, active, retired; got %q", sp.Status)
	}
	if sp.Code == "" {
		return fmt.Errorf("SearchParameter.code is required")
	}
	if len(sp.Base) == 0 {
		return fmt.Errorf("SearchParameter.base is required (at least one resource type)")
	}
	if sp.Type == "" {
		return fmt.Errorf("SearchParameter.type is required")
	}
	if !validSearchParamTypes[sp.Type] {
		return fmt.Errorf("SearchParameter.type must be one of: number, date, string, token, reference, composite, quantity, uri, special; got %q", sp.Type)
	}
	return nil
}

// ---------------------------------------------------------------------------
// NewDefaultSearchParameterStore — convenience constructor
// ---------------------------------------------------------------------------

// NewDefaultSearchParameterStore returns a SearchParameterStore pre-populated
// with common FHIR R4 search parameters from DefaultSearchParameters.
func NewDefaultSearchParameterStore() *SearchParameterStore {
	store := NewSearchParameterStore()
	for _, sp := range DefaultSearchParameters() {
		// Ignore errors from default params since they are well-formed.
		_ = store.Create(sp)
	}
	return store
}
