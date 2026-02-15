package fhir

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ValueSetValidator checks whether codes are members of value sets.
type ValueSetValidator struct {
	valueSets map[string]*ValueSetDef // keyed by URL
	allSets   []*ValueSetDef
}

// ValueSetDef represents a value set definition with its included codes.
type ValueSetDef struct {
	URL         string
	Name        string
	Title       string
	Status      string
	CodeSystems []ValueSetInclude
}

// ValueSetInclude defines codes from a specific code system in the value set.
type ValueSetInclude struct {
	System   string
	Concepts []ValueSetConcept
}

// ValueSetConcept is a single code in a value set.
type ValueSetConcept struct {
	Code    string
	Display string
}

// ValidateCodeResult holds the result of a validate-code check.
type ValidateCodeResult struct {
	Result  bool
	Display string
	Message string
}

// NewValueSetValidator creates a ValueSetValidator with built-in FHIR R4 value sets.
func NewValueSetValidator() *ValueSetValidator {
	v := &ValueSetValidator{
		valueSets: make(map[string]*ValueSetDef),
	}
	v.loadBuiltinValueSets()
	return v
}

// loadBuiltinValueSets registers all built-in FHIR R4 required value sets.
func (v *ValueSetValidator) loadBuiltinValueSets() {
	// 1. Observation Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/observation-status",
		Name:   "ObservationStatus",
		Title:  "Observation Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/observation-status",
				Concepts: []ValueSetConcept{
					{Code: "registered", Display: "Registered"},
					{Code: "preliminary", Display: "Preliminary"},
					{Code: "final", Display: "Final"},
					{Code: "amended", Display: "Amended"},
					{Code: "corrected", Display: "Corrected"},
					{Code: "cancelled", Display: "Cancelled"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})

	// 2. Condition Clinical Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/condition-clinical",
		Name:   "ConditionClinicalStatusCodes",
		Title:  "Condition Clinical Status Codes",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://terminology.hl7.org/CodeSystem/condition-clinical",
				Concepts: []ValueSetConcept{
					{Code: "active", Display: "Active"},
					{Code: "recurrence", Display: "Recurrence"},
					{Code: "relapse", Display: "Relapse"},
					{Code: "inactive", Display: "Inactive"},
					{Code: "remission", Display: "Remission"},
					{Code: "resolved", Display: "Resolved"},
				},
			},
		},
	})

	// 3. Allergy Intolerance Clinical Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/allergy-intolerance-clinical",
		Name:   "AllergyIntoleranceClinicalStatusCodes",
		Title:  "Allergy Intolerance Clinical Status Codes",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
				Concepts: []ValueSetConcept{
					{Code: "active", Display: "Active"},
					{Code: "inactive", Display: "Inactive"},
					{Code: "resolved", Display: "Resolved"},
				},
			},
		},
	})

	// 4. Medication Request Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/medication-request-status",
		Name:   "MedicationRequestStatus",
		Title:  "Medication Request Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/CodeSystem/medicationrequest-status",
				Concepts: []ValueSetConcept{
					{Code: "active", Display: "Active"},
					{Code: "on-hold", Display: "On Hold"},
					{Code: "cancelled", Display: "Cancelled"},
					{Code: "completed", Display: "Completed"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "stopped", Display: "Stopped"},
					{Code: "draft", Display: "Draft"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})

	// 5. Encounter Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/encounter-status",
		Name:   "EncounterStatus",
		Title:  "Encounter Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/encounter-status",
				Concepts: []ValueSetConcept{
					{Code: "planned", Display: "Planned"},
					{Code: "arrived", Display: "Arrived"},
					{Code: "triaged", Display: "Triaged"},
					{Code: "in-progress", Display: "In Progress"},
					{Code: "onleave", Display: "On Leave"},
					{Code: "finished", Display: "Finished"},
					{Code: "cancelled", Display: "Cancelled"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})

	// 6. Administrative Gender
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/administrative-gender",
		Name:   "AdministrativeGender",
		Title:  "Administrative Gender",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/administrative-gender",
				Concepts: []ValueSetConcept{
					{Code: "male", Display: "Male"},
					{Code: "female", Display: "Female"},
					{Code: "other", Display: "Other"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})

	// 7. Procedure Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/procedure-status",
		Name:   "ProcedureStatus",
		Title:  "Procedure Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/event-status",
				Concepts: []ValueSetConcept{
					{Code: "preparation", Display: "Preparation"},
					{Code: "in-progress", Display: "In Progress"},
					{Code: "not-done", Display: "Not Done"},
					{Code: "on-hold", Display: "On Hold"},
					{Code: "stopped", Display: "Stopped"},
					{Code: "completed", Display: "Completed"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})

	// 8. Diagnostic Report Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/diagnostic-report-status",
		Name:   "DiagnosticReportStatus",
		Title:  "Diagnostic Report Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/diagnostic-report-status",
				Concepts: []ValueSetConcept{
					{Code: "registered", Display: "Registered"},
					{Code: "partial", Display: "Partial"},
					{Code: "preliminary", Display: "Preliminary"},
					{Code: "final", Display: "Final"},
					{Code: "amended", Display: "Amended"},
					{Code: "corrected", Display: "Corrected"},
					{Code: "appended", Display: "Appended"},
					{Code: "cancelled", Display: "Cancelled"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})

	// 9. Immunization Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/immunization-status",
		Name:   "ImmunizationStatus",
		Title:  "Immunization Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/event-status",
				Concepts: []ValueSetConcept{
					{Code: "completed", Display: "Completed"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "not-done", Display: "Not Done"},
				},
			},
		},
	})

	// 10. Care Plan Status
	v.registerValueSet(&ValueSetDef{
		URL:    "http://hl7.org/fhir/ValueSet/care-plan-status",
		Name:   "CarePlanStatus",
		Title:  "Care Plan Status",
		Status: "active",
		CodeSystems: []ValueSetInclude{
			{
				System: "http://hl7.org/fhir/request-status",
				Concepts: []ValueSetConcept{
					{Code: "draft", Display: "Draft"},
					{Code: "active", Display: "Active"},
					{Code: "on-hold", Display: "On Hold"},
					{Code: "revoked", Display: "Revoked"},
					{Code: "completed", Display: "Completed"},
					{Code: "entered-in-error", Display: "Entered in Error"},
					{Code: "unknown", Display: "Unknown"},
				},
			},
		},
	})
}

// registerValueSet adds a ValueSetDef to the internal index.
func (v *ValueSetValidator) registerValueSet(vs *ValueSetDef) {
	v.valueSets[vs.URL] = vs
	v.allSets = append(v.allSets, vs)
}

// ValidateCode checks whether a code belongs to the specified value set.
// If system is provided, only codes from that system are matched.
func (v *ValueSetValidator) ValidateCode(url, code, system string) *ValidateCodeResult {
	vs, ok := v.valueSets[url]
	if !ok {
		return &ValidateCodeResult{
			Result:  false,
			Message: "ValueSet not found",
		}
	}

	for _, include := range vs.CodeSystems {
		if system != "" && include.System != system {
			continue
		}
		for _, concept := range include.Concepts {
			if concept.Code == code {
				return &ValidateCodeResult{
					Result:  true,
					Display: concept.Display,
					Message: "Code is valid",
				}
			}
		}
	}

	return &ValidateCodeResult{
		Result:  false,
		Message: "Code not found in ValueSet",
	}
}

// ListValueSets returns a FHIR Bundle of ValueSet resources (summary).
func (v *ValueSetValidator) ListValueSets() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(v.allSets))
	for _, vs := range v.allSets {
		result = append(result, map[string]interface{}{
			"resourceType": "ValueSet",
			"url":          vs.URL,
			"name":         vs.Name,
			"title":        vs.Title,
			"status":       vs.Status,
		})
	}
	return result
}

// ValueSetValidateHandler provides the ValueSet/$validate-code HTTP endpoints.
type ValueSetValidateHandler struct {
	validator *ValueSetValidator
}

// NewValueSetValidateHandler creates a new ValueSetValidateHandler.
func NewValueSetValidateHandler(validator *ValueSetValidator) *ValueSetValidateHandler {
	return &ValueSetValidateHandler{validator: validator}
}

// RegisterRoutes adds ValueSet/$validate-code routes to the given FHIR group.
func (h *ValueSetValidateHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/ValueSet/$validate-code", h.ValidateCode)
	g.POST("/ValueSet/$validate-code", h.ValidateCodePost)
}

// ValidateCode handles GET /fhir/ValueSet/$validate-code with query parameters.
func (h *ValueSetValidateHandler) ValidateCode(c echo.Context) error {
	url := c.QueryParam("url")
	code := c.QueryParam("code")
	system := c.QueryParam("system")

	if url == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'url' is required"))
	}
	if code == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'code' is required"))
	}

	result := h.validator.ValidateCode(url, code, system)
	return c.JSON(http.StatusOK, buildValidateCodeParametersResponse(result))
}

// ValidateCodePost handles POST /fhir/ValueSet/$validate-code with a Parameters resource body.
func (h *ValueSetValidateHandler) ValidateCodePost(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Failed to read request body"))
	}

	var params struct {
		ResourceType string `json:"resourceType"`
		Parameter    []struct {
			Name        string `json:"name"`
			ValueUri    string `json:"valueUri,omitempty"`
			ValueCode   string `json:"valueCode,omitempty"`
			ValueString string `json:"valueString,omitempty"`
		} `json:"parameter"`
	}

	if err := json.Unmarshal(body, &params); err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
	}

	var url, code, system string
	for _, p := range params.Parameter {
		switch p.Name {
		case "url":
			url = p.ValueUri
		case "code":
			code = p.ValueCode
		case "system":
			system = p.ValueUri
		}
	}

	if url == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'url' is required"))
	}
	if code == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'code' is required"))
	}

	result := h.validator.ValidateCode(url, code, system)
	return c.JSON(http.StatusOK, buildValidateCodeParametersResponse(result))
}

// buildValidateCodeParametersResponse converts a ValidateCodeResult to a FHIR Parameters resource.
func buildValidateCodeParametersResponse(result *ValidateCodeResult) map[string]interface{} {
	params := []interface{}{
		map[string]interface{}{
			"name":         "result",
			"valueBoolean": result.Result,
		},
	}

	if result.Display != "" {
		params = append(params, map[string]interface{}{
			"name":        "display",
			"valueString": result.Display,
		})
	}

	if result.Message != "" {
		params = append(params, map[string]interface{}{
			"name":        "message",
			"valueString": result.Message,
		})
	}

	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    params,
	}
}
