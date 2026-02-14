package terminology

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
)

// Handler provides REST endpoints for terminology services.
type Handler struct {
	svc *Service
}

// NewHandler creates a new terminology handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers terminology routes on the API and FHIR groups.
func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// All authenticated users can search terminology
	termGroup := api.Group("/terminology", auth.RequireRole("admin", "physician", "nurse", "pharmacist", "lab-tech"))
	termGroup.GET("/loinc", h.SearchLOINC)
	termGroup.GET("/icd10", h.SearchICD10)
	termGroup.GET("/snomed", h.SearchSNOMED)
	termGroup.GET("/rxnorm", h.SearchRxNorm)
	termGroup.GET("/cpt", h.SearchCPT)

	// FHIR terminology operations
	fhirTerm := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "pharmacist", "lab-tech"))
	fhirTerm.POST("/CodeSystem/$lookup", h.FHIRLookup)
	fhirTerm.POST("/CodeSystem/$validate-code", h.FHIRValidateCode)
	fhirTerm.GET("/ValueSet/$expand", h.ExpandValueSet)
	fhirTerm.POST("/ValueSet/$expand", h.ExpandValueSet)
}

func getLimit(c echo.Context) int {
	limit, _ := strconv.Atoi(c.QueryParam("_count"))
	if limit <= 0 {
		limit, _ = strconv.Atoi(c.QueryParam("limit"))
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit
}

// SearchLOINC handles GET /api/v1/terminology/loinc?q=...
func (h *Handler) SearchLOINC(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}
	results, err := h.svc.SearchLOINC(c.Request().Context(), query, getLimit(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, results)
}

// SearchICD10 handles GET /api/v1/terminology/icd10?q=...
func (h *Handler) SearchICD10(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}
	results, err := h.svc.SearchICD10(c.Request().Context(), query, getLimit(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, results)
}

// SearchSNOMED handles GET /api/v1/terminology/snomed?q=...
func (h *Handler) SearchSNOMED(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}
	results, err := h.svc.SearchSNOMED(c.Request().Context(), query, getLimit(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, results)
}

// SearchRxNorm handles GET /api/v1/terminology/rxnorm?q=...
func (h *Handler) SearchRxNorm(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}
	results, err := h.svc.SearchRxNorm(c.Request().Context(), query, getLimit(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, results)
}

// SearchCPT handles GET /api/v1/terminology/cpt?q=...
func (h *Handler) SearchCPT(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}
	results, err := h.svc.SearchCPT(c.Request().Context(), query, getLimit(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, results)
}

// FHIRLookup handles POST /fhir/CodeSystem/$lookup
func (h *Handler) FHIRLookup(c echo.Context) error {
	var req LookupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	resp, err := h.svc.Lookup(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, resp)
}

// FHIRValidateCode handles POST /fhir/CodeSystem/$validate-code
func (h *Handler) FHIRValidateCode(c echo.Context) error {
	var req ValidateCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	resp, err := h.svc.ValidateCode(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, resp)
}

// ExpandValueSet handles GET/POST /fhir/ValueSet/$expand
func (h *Handler) ExpandValueSet(c echo.Context) error {
	url := c.QueryParam("url")
	filter := c.QueryParam("filter")
	countStr := c.QueryParam("count")
	offsetStr := c.QueryParam("offset")

	count := 100
	if countStr != "" {
		if v, err := strconv.Atoi(countStr); err == nil && v > 0 {
			count = v
		}
	}
	offset := 0
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	// Map well-known ValueSet URLs to code system lookups
	var systemURI string
	switch {
	case strings.Contains(url, "loinc"):
		systemURI = SystemLOINC
	case strings.Contains(url, "icd10") || strings.Contains(url, "icd-10"):
		systemURI = SystemICD10
	case strings.Contains(url, "snomed") || strings.Contains(url, "sct"):
		systemURI = SystemSNOMED
	case strings.Contains(url, "rxnorm"):
		systemURI = SystemRxNorm
	case strings.Contains(url, "cpt"):
		systemURI = SystemCPT
	}

	var contains []map[string]interface{}

	if systemURI != "" && filter != "" {
		results, err := h.svc.SearchCodes(c.Request().Context(), systemURI, filter, count, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
		}
		for _, r := range results {
			contains = append(contains, map[string]interface{}{
				"system":  r.SystemURI,
				"code":    r.Code,
				"display": r.Display,
			})
		}
	}

	if contains == nil {
		contains = []map[string]interface{}{}
	}

	result := map[string]interface{}{
		"resourceType": "ValueSet",
		"expansion": map[string]interface{}{
			"identifier": uuid.New().String(),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"total":      len(contains),
			"offset":     offset,
			"contains":   contains,
		},
	}
	return c.JSON(http.StatusOK, result)
}
