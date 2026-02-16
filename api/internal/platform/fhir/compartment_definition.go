package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// FHIR CompartmentDefinition resource types
// ---------------------------------------------------------------------------

// FHIRCompartmentDefinition represents the FHIR CompartmentDefinition resource.
// It describes how a compartment is defined: which resource types belong and
// what search parameters link them into the compartment.
type FHIRCompartmentDefinition struct {
	ResourceType string                `json:"resourceType"`
	ID           string                `json:"id,omitempty"`
	URL          string                `json:"url"`
	Name         string                `json:"name"`
	Status       string                `json:"status"`
	Code         string                `json:"code"` // Patient, Encounter, Practitioner, RelatedPerson, Device
	Search       bool                  `json:"search"`
	Resource     []CompartmentResource `json:"resource,omitempty"`
}

// CompartmentResource describes a single resource type's membership in a compartment,
// including the search parameters that link it to the compartment subject.
type CompartmentResource struct {
	Code  string   `json:"code"`  // Resource type (e.g. "Observation")
	Param []string `json:"param"` // Search parameters linking to compartment
}

// ---------------------------------------------------------------------------
// Standard FHIR R4 Compartment Definitions
// ---------------------------------------------------------------------------

// PatientCompartmentDef returns the FHIR R4 Patient compartment definition
// with all resource type memberships per the specification.
func PatientCompartmentDef() *FHIRCompartmentDefinition {
	return &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		ID:           "patient",
		URL:          "http://hl7.org/fhir/CompartmentDefinition/patient",
		Name:         "Patient",
		Status:       "active",
		Code:         "Patient",
		Search:       true,
		Resource: []CompartmentResource{
			{Code: "Account", Param: []string{"subject"}},
			{Code: "AllergyIntolerance", Param: []string{"patient", "recorder", "asserter"}},
			{Code: "Appointment", Param: []string{"actor"}},
			{Code: "AuditEvent", Param: []string{"patient"}},
			{Code: "CarePlan", Param: []string{"patient", "performer"}},
			{Code: "CareTeam", Param: []string{"patient", "participant"}},
			{Code: "Claim", Param: []string{"patient", "payee"}},
			{Code: "ClinicalImpression", Param: []string{"subject"}},
			{Code: "Communication", Param: []string{"subject", "sender", "recipient"}},
			{Code: "Condition", Param: []string{"patient", "asserter"}},
			{Code: "Consent", Param: []string{"patient"}},
			{Code: "Coverage", Param: []string{"patient", "subscriber", "beneficiary", "payor"}},
			{Code: "DetectedIssue", Param: []string{"patient"}},
			{Code: "DeviceRequest", Param: []string{"subject", "performer"}},
			{Code: "DiagnosticReport", Param: []string{"subject"}},
			{Code: "DocumentReference", Param: []string{"subject", "author"}},
			{Code: "Encounter", Param: []string{"patient"}},
			{Code: "EpisodeOfCare", Param: []string{"patient"}},
			{Code: "ExplanationOfBenefit", Param: []string{"patient", "payee"}},
			{Code: "FamilyMemberHistory", Param: []string{"patient"}},
			{Code: "Goal", Param: []string{"patient"}},
			{Code: "ImagingStudy", Param: []string{"patient"}},
			{Code: "Immunization", Param: []string{"patient"}},
			{Code: "List", Param: []string{"subject", "source"}},
			{Code: "MedicationAdministration", Param: []string{"patient", "performer", "subject"}},
			{Code: "MedicationDispense", Param: []string{"subject", "patient", "receiver"}},
			{Code: "MedicationRequest", Param: []string{"subject"}},
			{Code: "MedicationStatement", Param: []string{"subject"}},
			{Code: "NutritionOrder", Param: []string{"patient"}},
			{Code: "Observation", Param: []string{"subject", "performer"}},
			{Code: "Procedure", Param: []string{"patient", "performer"}},
			{Code: "Provenance", Param: []string{"patient"}},
			{Code: "QuestionnaireResponse", Param: []string{"subject", "author"}},
			{Code: "RelatedPerson", Param: []string{"patient"}},
			{Code: "RiskAssessment", Param: []string{"subject"}},
			{Code: "Schedule", Param: []string{"actor"}},
			{Code: "ServiceRequest", Param: []string{"subject", "performer"}},
			{Code: "Specimen", Param: []string{"subject"}},
		},
	}
}

// EncounterCompartmentDef returns the FHIR R4 Encounter compartment definition.
func EncounterCompartmentDef() *FHIRCompartmentDefinition {
	return &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		ID:           "encounter",
		URL:          "http://hl7.org/fhir/CompartmentDefinition/encounter",
		Name:         "Encounter",
		Status:       "active",
		Code:         "Encounter",
		Search:       true,
		Resource: []CompartmentResource{
			{Code: "CarePlan", Param: []string{"encounter"}},
			{Code: "CareTeam", Param: []string{"encounter"}},
			{Code: "Claim", Param: []string{"encounter"}},
			{Code: "Communication", Param: []string{"encounter"}},
			{Code: "Composition", Param: []string{"encounter"}},
			{Code: "Condition", Param: []string{"encounter"}},
			{Code: "DiagnosticReport", Param: []string{"encounter"}},
			{Code: "DocumentReference", Param: []string{"encounter"}},
			{Code: "Encounter", Param: []string{"_id"}},
			{Code: "ExplanationOfBenefit", Param: []string{"encounter"}},
			{Code: "MedicationAdministration", Param: []string{"context"}},
			{Code: "MedicationDispense", Param: []string{"context"}},
			{Code: "MedicationRequest", Param: []string{"encounter"}},
			{Code: "NutritionOrder", Param: []string{"encounter"}},
			{Code: "Observation", Param: []string{"encounter"}},
			{Code: "Procedure", Param: []string{"encounter"}},
			{Code: "QuestionnaireResponse", Param: []string{"encounter"}},
			{Code: "RiskAssessment", Param: []string{"encounter"}},
			{Code: "ServiceRequest", Param: []string{"encounter"}},
		},
	}
}

// PractitionerCompartmentDef returns the FHIR R4 Practitioner compartment definition.
func PractitionerCompartmentDef() *FHIRCompartmentDefinition {
	return &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		ID:           "practitioner",
		URL:          "http://hl7.org/fhir/CompartmentDefinition/practitioner",
		Name:         "Practitioner",
		Status:       "active",
		Code:         "Practitioner",
		Search:       true,
		Resource: []CompartmentResource{
			{Code: "Account", Param: []string{"subject"}},
			{Code: "Appointment", Param: []string{"actor"}},
			{Code: "AuditEvent", Param: []string{"agent"}},
			{Code: "CarePlan", Param: []string{"performer"}},
			{Code: "CareTeam", Param: []string{"participant"}},
			{Code: "Claim", Param: []string{"enterer", "provider", "payee"}},
			{Code: "Communication", Param: []string{"sender", "recipient"}},
			{Code: "CommunicationRequest", Param: []string{"sender", "recipient", "requester"}},
			{Code: "Composition", Param: []string{"subject", "author", "attester"}},
			{Code: "Condition", Param: []string{"asserter"}},
			{Code: "DiagnosticReport", Param: []string{"performer"}},
			{Code: "DocumentReference", Param: []string{"author"}},
			{Code: "Encounter", Param: []string{"practitioner", "participant"}},
			{Code: "EpisodeOfCare", Param: []string{"care-manager"}},
			{Code: "ExplanationOfBenefit", Param: []string{"enterer", "provider", "payee"}},
			{Code: "ImagingStudy", Param: []string{"performer"}},
			{Code: "Immunization", Param: []string{"performer"}},
			{Code: "MedicationAdministration", Param: []string{"performer"}},
			{Code: "MedicationDispense", Param: []string{"performer", "receiver"}},
			{Code: "MedicationRequest", Param: []string{"requester"}},
			{Code: "Observation", Param: []string{"performer"}},
			{Code: "Procedure", Param: []string{"performer"}},
			{Code: "Provenance", Param: []string{"agent"}},
			{Code: "Schedule", Param: []string{"actor"}},
			{Code: "ServiceRequest", Param: []string{"performer", "requester"}},
			{Code: "Task", Param: []string{"owner", "requester"}},
		},
	}
}

// RelatedPersonCompartmentDef returns the FHIR R4 RelatedPerson compartment definition.
func RelatedPersonCompartmentDef() *FHIRCompartmentDefinition {
	return &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		ID:           "relatedperson",
		URL:          "http://hl7.org/fhir/CompartmentDefinition/relatedperson",
		Name:         "RelatedPerson",
		Status:       "active",
		Code:         "RelatedPerson",
		Search:       true,
		Resource: []CompartmentResource{
			{Code: "AllergyIntolerance", Param: []string{"asserter"}},
			{Code: "Appointment", Param: []string{"actor"}},
			{Code: "CarePlan", Param: []string{"performer"}},
			{Code: "CareTeam", Param: []string{"participant"}},
			{Code: "Claim", Param: []string{"payee"}},
			{Code: "Communication", Param: []string{"sender", "recipient"}},
			{Code: "CommunicationRequest", Param: []string{"sender", "recipient", "requester"}},
			{Code: "Composition", Param: []string{"author"}},
			{Code: "Condition", Param: []string{"asserter"}},
			{Code: "Consent", Param: []string{"actor"}},
			{Code: "Encounter", Param: []string{"participant"}},
			{Code: "MedicationAdministration", Param: []string{"performer"}},
			{Code: "MedicationDispense", Param: []string{"performer"}},
			{Code: "Observation", Param: []string{"performer"}},
			{Code: "Patient", Param: []string{"link"}},
			{Code: "Procedure", Param: []string{"performer"}},
			{Code: "Provenance", Param: []string{"agent"}},
			{Code: "ServiceRequest", Param: []string{"performer"}},
		},
	}
}

// DeviceCompartmentDef returns the FHIR R4 Device compartment definition.
func DeviceCompartmentDef() *FHIRCompartmentDefinition {
	return &FHIRCompartmentDefinition{
		ResourceType: "CompartmentDefinition",
		ID:           "device",
		URL:          "http://hl7.org/fhir/CompartmentDefinition/device",
		Name:         "Device",
		Status:       "active",
		Code:         "Device",
		Search:       true,
		Resource: []CompartmentResource{
			{Code: "Account", Param: []string{"subject"}},
			{Code: "Appointment", Param: []string{"actor"}},
			{Code: "AuditEvent", Param: []string{"agent"}},
			{Code: "Communication", Param: []string{"sender"}},
			{Code: "CommunicationRequest", Param: []string{"sender"}},
			{Code: "DeviceMetric", Param: []string{"source"}},
			{Code: "DeviceRequest", Param: []string{"device", "subject"}},
			{Code: "DeviceUseStatement", Param: []string{"device"}},
			{Code: "DiagnosticReport", Param: []string{"subject"}},
			{Code: "DocumentReference", Param: []string{"author"}},
			{Code: "Media", Param: []string{"subject"}},
			{Code: "MedicationAdministration", Param: []string{"device"}},
			{Code: "Observation", Param: []string{"subject", "device"}},
			{Code: "Procedure", Param: []string{"performer"}},
			{Code: "Schedule", Param: []string{"actor"}},
			{Code: "ServiceRequest", Param: []string{"performer", "requester"}},
			{Code: "Specimen", Param: []string{"subject"}},
			{Code: "Task", Param: []string{"owner", "requester"}},
		},
	}
}

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

// allCompartmentDefinitions returns all five standard compartment definitions
// keyed by their ID.
func allCompartmentDefinitions() map[string]*FHIRCompartmentDefinition {
	return map[string]*FHIRCompartmentDefinition{
		"patient":       PatientCompartmentDef(),
		"encounter":     EncounterCompartmentDef(),
		"practitioner":  PractitionerCompartmentDef(),
		"relatedperson": RelatedPersonCompartmentDef(),
		"device":        DeviceCompartmentDef(),
	}
}

// GetCompartmentDefinitionByCode returns the compartment definition matching
// the given code (e.g. "Patient"). The match is case-insensitive.
func GetCompartmentDefinitionByCode(code string) *FHIRCompartmentDefinition {
	lower := strings.ToLower(code)
	for _, def := range allCompartmentDefinitions() {
		if strings.ToLower(def.Code) == lower {
			return def
		}
	}
	return nil
}

// CompartmentResourceParams returns the search parameters that link a given
// resource type into a compartment definition. Returns nil if the resource
// type is not a member of the compartment.
func CompartmentResourceParams(def *FHIRCompartmentDefinition, resourceType string) []string {
	for _, r := range def.Resource {
		if r.Code == resourceType {
			return r.Param
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// CompartmentDefinitionHandler — HTTP handler for CompartmentDefinition
// ---------------------------------------------------------------------------

// CompartmentDefinitionHandler serves FHIR CompartmentDefinition resources.
type CompartmentDefinitionHandler struct {
	definitions map[string]*FHIRCompartmentDefinition
}

// NewCompartmentDefinitionHandler creates a handler pre-loaded with the five
// standard FHIR R4 compartment definitions.
func NewCompartmentDefinitionHandler() *CompartmentDefinitionHandler {
	return &CompartmentDefinitionHandler{
		definitions: allCompartmentDefinitions(),
	}
}

// RegisterRoutes registers CompartmentDefinition endpoints on the provided
// Echo group.
func (h *CompartmentDefinitionHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/CompartmentDefinition/:id", h.GetDefinition)
	g.GET("/CompartmentDefinition", h.SearchDefinitions)
}

// GetDefinition handles GET /fhir/CompartmentDefinition/:id.
// It returns the compartment definition with the matching ID.
func (h *CompartmentDefinitionHandler) GetDefinition(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("compartment definition ID is required"))
	}

	def, ok := h.definitions[id]
	if !ok {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("CompartmentDefinition", id))
	}

	return c.JSON(http.StatusOK, def)
}

// SearchDefinitions handles GET /fhir/CompartmentDefinition.
// It supports optional query parameters:
//   - code: filter by compartment code (e.g. "Patient")
//   - url:  filter by canonical URL
//   - name: filter by name
func (h *CompartmentDefinitionHandler) SearchDefinitions(c echo.Context) error {
	codeFilter := c.QueryParam("code")
	urlFilter := c.QueryParam("url")
	nameFilter := c.QueryParam("name")

	var matches []interface{}
	for _, def := range h.definitions {
		if codeFilter != "" && !strings.EqualFold(def.Code, codeFilter) {
			continue
		}
		if urlFilter != "" && def.URL != urlFilter {
			continue
		}
		if nameFilter != "" && !strings.EqualFold(def.Name, nameFilter) {
			continue
		}
		matches = append(matches, def)
	}

	total := len(matches)

	// Build bundle entries
	entries := make([]BundleEntry, len(matches))
	for i, m := range matches {
		raw, _ := json.Marshal(m)
		def := m.(*FHIRCompartmentDefinition)
		entries[i] = BundleEntry{
			FullURL:  fmt.Sprintf("CompartmentDefinition/%s", def.ID),
			Resource: raw,
			Search: &BundleSearch{
				Mode: "match",
			},
		}
	}

	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &total,
		Entry:        entries,
	}

	return c.JSON(http.StatusOK, bundle)
}
