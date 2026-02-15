package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// PatientResourceFetcher retrieves all resources of a given type for a patient.
// The patientID is the patient's FHIR ID (UUID string).
type PatientResourceFetcher func(ctx context.Context, patientID string) ([]map[string]interface{}, error)

// EverythingHandler implements the FHIR Patient/$everything operation.
// It aggregates data from all registered fetchers into a single searchset Bundle.
type EverythingHandler struct {
	fetchers       map[string]PatientResourceFetcher
	order          []string
	patientFetcher func(ctx context.Context, fhirID string) (map[string]interface{}, error)
}

// NewEverythingHandler creates a new EverythingHandler.
func NewEverythingHandler() *EverythingHandler {
	return &EverythingHandler{
		fetchers: make(map[string]PatientResourceFetcher),
	}
}

// SetPatientFetcher sets the function used to retrieve the Patient resource itself.
func (h *EverythingHandler) SetPatientFetcher(fn func(ctx context.Context, fhirID string) (map[string]interface{}, error)) {
	h.patientFetcher = fn
}

// RegisterFetcher registers a fetcher for the given FHIR resource type.
// Registration order determines the order of resources in the output Bundle.
func (h *EverythingHandler) RegisterFetcher(resourceType string, fn PatientResourceFetcher) {
	if _, exists := h.fetchers[resourceType]; !exists {
		h.order = append(h.order, resourceType)
	}
	h.fetchers[resourceType] = fn
}

// RegisterRoutes registers the $everything route on the FHIR group.
func (h *EverythingHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/Patient/:id/$everything", h.Handle)
}

// Handle processes GET /fhir/Patient/:id/$everything.
func (h *EverythingHandler) Handle(c echo.Context) error {
	fhirID := c.Param("id")
	if fhirID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("patient id is required"))
	}

	// Parse _type filter
	var typeFilter map[string]bool
	if typeParam := c.QueryParam("_type"); typeParam != "" {
		typeFilter = make(map[string]bool)
		for _, t := range strings.Split(typeParam, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				typeFilter[t] = true
			}
		}
	}

	// Parse _count (per-type limit)
	countLimit := 0
	if countParam := c.QueryParam("_count"); countParam != "" {
		n, err := strconv.Atoi(countParam)
		if err != nil || n < 0 {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("_count must be a non-negative integer"))
		}
		countLimit = n
	}

	ctx := c.Request().Context()

	// Fetch the Patient resource
	if h.patientFetcher == nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome("patient fetcher not configured"))
	}
	patient, err := h.patientFetcher(ctx, fhirID)
	if err != nil || patient == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("Patient", fhirID))
	}

	// Build entries starting with Patient
	var entries []BundleEntry
	patientRaw, _ := json.Marshal(patient)
	entries = append(entries, BundleEntry{
		FullURL:  fmt.Sprintf("Patient/%s", fhirID),
		Resource: patientRaw,
		Search:   &BundleSearch{Mode: "match"},
	})

	// Iterate registered fetchers in order
	for _, rt := range h.order {
		if typeFilter != nil && !typeFilter[rt] {
			continue
		}

		fn, ok := h.fetchers[rt]
		if !ok {
			continue
		}

		resources, err := fn(ctx, fhirID)
		if err != nil {
			continue
		}

		if countLimit > 0 && len(resources) > countLimit {
			resources = resources[:countLimit]
		}

		for _, r := range resources {
			raw, err := json.Marshal(r)
			if err != nil {
				continue
			}
			id, _ := r["id"].(string)
			entries = append(entries, BundleEntry{
				FullURL:  fmt.Sprintf("%s/%s", rt, id),
				Resource: raw,
				Search:   &BundleSearch{Mode: "match"},
			})
		}
	}

	total := len(entries)
	bundle := &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &total,
		Entry:        entries,
	}

	return c.JSON(http.StatusOK, bundle)
}
