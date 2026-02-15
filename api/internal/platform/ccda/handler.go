package ccda

import (
	"context"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// DataFetcher retrieves all clinical data for a patient.
type DataFetcher interface {
	FetchPatientData(ctx context.Context, patientID string) (*PatientData, error)
}

// Handler provides HTTP endpoints for C-CDA generation and parsing.
type Handler struct {
	generator *Generator
	parser    *Parser
	fetcher   DataFetcher
}

// NewHandler creates a new C-CDA handler.
func NewHandler(generator *Generator, parser *Parser, fetcher DataFetcher) *Handler {
	return &Handler{
		generator: generator,
		parser:    parser,
		fetcher:   fetcher,
	}
}

// RegisterRoutes registers C-CDA endpoints on the provided route group.
//
//	GET  /api/v1/patients/:id/ccd  - Generate CCD for a patient
//	POST /api/v1/ccda/parse        - Parse an incoming C-CDA document
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/patients/:id/ccd", h.GenerateCCD)
	g.POST("/ccda/parse", h.ParseCCDA)
}

// GenerateCCD handles GET /api/v1/patients/:id/ccd.
// It fetches all patient data and returns a CCD XML document.
func (h *Handler) GenerateCCD(c echo.Context) error {
	patientID := c.Param("id")
	if patientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "patient ID is required",
		})
	}

	data, err := h.fetcher.FetchPatientData(c.Request().Context(), patientID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "patient not found: " + err.Error(),
		})
	}

	xmlData, err := h.generator.GenerateCCD(data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate CCD: " + err.Error(),
		})
	}

	return c.Blob(http.StatusOK, "application/xml", xmlData)
}

// ParseCCDA handles POST /api/v1/ccda/parse.
// It accepts an XML body and returns parsed sections as JSON.
func (h *Handler) ParseCCDA(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to read request body",
		})
	}

	parsed, err := h.parser.Parse(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to parse C-CDA: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, parsed)
}
