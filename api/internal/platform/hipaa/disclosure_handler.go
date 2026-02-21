package hipaa

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
)

// DisclosureHandler provides Echo HTTP handlers for accounting of disclosures.
type DisclosureHandler struct {
	store *DisclosureStore
}

// NewDisclosureHandler creates a new handler backed by the given store.
func NewDisclosureHandler(store *DisclosureStore) *DisclosureHandler {
	return &DisclosureHandler{store: store}
}

// RegisterDisclosureRoutes registers disclosure routes on the API and FHIR groups.
func RegisterDisclosureRoutes(apiV1 *echo.Group, fhirGroup *echo.Group, store *DisclosureStore) {
	h := NewDisclosureHandler(store)

	// POST /api/v1/disclosures - Record a disclosure (admin, physician)
	apiV1.POST("/disclosures", h.HandleRecordDisclosure, auth.RequireRole("admin", "physician"))

	// GET /api/v1/disclosures - List all disclosures (admin only, with pagination)
	apiV1.GET("/disclosures", h.HandleListDisclosures, auth.RequireRole("admin"))

	// GET /api/v1/patients/:patientId/disclosures - List disclosures for a patient
	apiV1.GET("/patients/:patientId/disclosures", h.HandleListPatientDisclosures, auth.RequireRole("admin", "physician", "patient"))

	// GET /fhir/Patient/:id/$accounting-of-disclosures - FHIR-style endpoint
	fhirGroup.GET("/Patient/:id/$accounting-of-disclosures", h.HandleFHIRAccountingOfDisclosures)
}

// CreateDisclosureRequest is the request body for recording a disclosure.
type CreateDisclosureRequest struct {
	PatientID       string   `json:"patient_id"`
	DisclosedTo     string   `json:"disclosed_to"`
	DisclosedToType string   `json:"disclosed_to_type"`
	Purpose         string   `json:"purpose"`
	ResourceTypes   []string `json:"resource_types"`
	ResourceIDs     []string `json:"resource_ids,omitempty"`
	DateDisclosed   string   `json:"date_disclosed,omitempty"` // RFC3339
	DisclosedBy     string   `json:"disclosed_by"`
	Method          string   `json:"method"`
	Description     string   `json:"description"`
}

// HandleRecordDisclosure handles POST /api/v1/disclosures.
func (h *DisclosureHandler) HandleRecordDisclosure(c echo.Context) error {
	var req CreateDisclosureRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if req.PatientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "patient_id is required"})
	}
	if req.DisclosedTo == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "disclosed_to is required"})
	}
	if req.Purpose == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "purpose is required"})
	}
	if !IsValidDisclosurePurpose(req.Purpose) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid purpose: must be one of: public-health, research, law-enforcement, judicial, workers-comp, decedent, organ-donation, health-oversight, other",
		})
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid patient_id: " + err.Error()})
	}

	var dateDisclosed time.Time
	if req.DateDisclosed != "" {
		dateDisclosed, err = time.Parse(time.RFC3339, req.DateDisclosed)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid date_disclosed: " + err.Error()})
		}
	}

	// Use the authenticated user ID if disclosed_by is not provided
	disclosedBy := req.DisclosedBy
	if disclosedBy == "" {
		disclosedBy = auth.UserIDFromContext(c.Request().Context())
	}

	disclosure := &Disclosure{
		PatientID:       patientID,
		DisclosedTo:     req.DisclosedTo,
		DisclosedToType: req.DisclosedToType,
		Purpose:         req.Purpose,
		ResourceTypes:   req.ResourceTypes,
		ResourceIDs:     req.ResourceIDs,
		DateDisclosed:   dateDisclosed,
		DisclosedBy:     disclosedBy,
		Method:          req.Method,
		Description:     req.Description,
	}

	if err := h.store.Record(disclosure); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, disclosure)
}

// HandleListDisclosures handles GET /api/v1/disclosures (admin only, paginated).
func (h *DisclosureHandler) HandleListDisclosures(c echo.Context) error {
	limit := 20
	offset := 0

	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if limit > 100 {
		limit = 100
	}

	disclosures, total, err := h.store.ListAll(limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":     disclosures,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": offset+limit < total,
	})
}

// HandleListPatientDisclosures handles GET /api/v1/patients/:patientId/disclosures.
func (h *DisclosureHandler) HandleListPatientDisclosures(c echo.Context) error {
	patientIDStr := c.Param("patientId")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid patient ID"})
	}

	// Default to 6-year window (HIPAA requirement)
	to := time.Now().UTC()
	from := to.AddDate(-6, 0, 0)

	if v := c.QueryParam("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}

	disclosures, err := h.store.ListByPatient(patientID, from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":       disclosures,
		"patient_id": patientID.String(),
		"from":       from.Format(time.RFC3339),
		"to":         to.Format(time.RFC3339),
		"total":      len(disclosures),
	})
}

// HandleFHIRAccountingOfDisclosures handles GET /fhir/Patient/:id/$accounting-of-disclosures.
// Returns disclosures in a FHIR Bundle-like structure per HIPAA Section 164.528.
func (h *DisclosureHandler) HandleFHIRAccountingOfDisclosures(c echo.Context) error {
	idStr := c.Param("id")
	patientID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"resourceType": "OperationOutcome",
			"issue": []map[string]interface{}{
				{
					"severity":    "error",
					"code":        "invalid",
					"diagnostics": "Invalid patient ID format",
				},
			},
		})
	}

	// Default 6-year window
	to := time.Now().UTC()
	from := to.AddDate(-6, 0, 0)

	if v := c.QueryParam("start"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := c.QueryParam("end"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}

	disclosures, err := h.store.ListByPatient(patientID, from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"resourceType": "OperationOutcome",
			"issue": []map[string]interface{}{
				{
					"severity":    "error",
					"code":        "exception",
					"diagnostics": err.Error(),
				},
			},
		})
	}

	// Build FHIR Bundle entries
	entries := make([]map[string]interface{}, 0, len(disclosures))
	for _, d := range disclosures {
		entry := map[string]interface{}{
			"fullUrl": "urn:uuid:" + d.ID.String(),
			"resource": map[string]interface{}{
				"resourceType": "AuditEvent",
				"id":           d.ID.String(),
				"type": map[string]interface{}{
					"system":  "http://dicom.nema.org/resources/ontology/DCM",
					"code":    "110106",
					"display": "Export",
				},
				"subtype": []map[string]interface{}{
					{
						"system":  "http://hl7.org/fhir/audit-event-sub-type",
						"code":    "disclosure",
						"display": "Disclosure",
					},
				},
				"recorded": d.DateDisclosed.Format(time.RFC3339),
				"agent": []map[string]interface{}{
					{
						"who": map[string]interface{}{
							"display": d.DisclosedBy,
						},
						"requestor": true,
					},
				},
				"entity": []map[string]interface{}{
					{
						"what": map[string]interface{}{
							"reference": "Patient/" + d.PatientID.String(),
						},
						"type": map[string]interface{}{
							"system":  "http://terminology.hl7.org/CodeSystem/audit-entity-type",
							"code":    "1",
							"display": "Person",
						},
					},
				},
				"purposeOfEvent": []map[string]interface{}{
					{
						"coding": []map[string]interface{}{
							{
								"system":  "http://terminology.hl7.org/CodeSystem/v3-ActReason",
								"code":    d.Purpose,
								"display": d.Purpose,
							},
						},
					},
				},
				"extension": []map[string]interface{}{
					{
						"url":         "http://ehr.example.com/fhir/StructureDefinition/disclosure-recipient",
						"valueString": d.DisclosedTo,
					},
					{
						"url":         "http://ehr.example.com/fhir/StructureDefinition/disclosure-method",
						"valueString": d.Method,
					},
					{
						"url":         "http://ehr.example.com/fhir/StructureDefinition/disclosure-description",
						"valueString": d.Description,
					},
				},
			},
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
