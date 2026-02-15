package hl7v2

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler provides HTTP endpoints for HL7v2 message parsing and generation.
type Handler struct{}

// NewHandler creates a new HL7v2 handler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers HL7v2 endpoints on the provided route group.
//
//	POST /api/v1/hl7v2/parse          - Parse HL7v2 message to JSON
//	POST /api/v1/hl7v2/generate/adt   - Generate ADT message from FHIR
//	POST /api/v1/hl7v2/generate/orm   - Generate ORM message from FHIR
//	POST /api/v1/hl7v2/generate/oru   - Generate ORU message from FHIR
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/hl7v2/parse", h.ParseMessage)
	g.POST("/hl7v2/generate/adt", h.GenerateADTHandler)
	g.POST("/hl7v2/generate/orm", h.GenerateORMHandler)
	g.POST("/hl7v2/generate/oru", h.GenerateORUHandler)
}

// segmentJSON is the JSON representation of a parsed segment.
type segmentJSON struct {
	Name   string      `json:"name"`
	Fields []fieldJSON `json:"fields"`
}

// fieldJSON is the JSON representation of a parsed field.
type fieldJSON struct {
	Value      string     `json:"value"`
	Components []string   `json:"components,omitempty"`
	Repeats    [][]string `json:"repeats,omitempty"`
}

// ParseMessage handles POST /api/v1/hl7v2/parse.
// It reads raw HL7v2 from the request body and returns parsed JSON.
func (h *Handler) ParseMessage(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to read request body",
		})
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "request body is empty",
		})
	}

	msg, err := Parse(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to parse HL7v2 message: " + err.Error(),
		})
	}

	// Build JSON response
	segments := make([]segmentJSON, len(msg.Segments))
	for i, seg := range msg.Segments {
		fields := make([]fieldJSON, len(seg.Fields))
		for j, f := range seg.Fields {
			fields[j] = fieldJSON{
				Value:      f.Value,
				Components: f.Components,
				Repeats:    f.Repeats,
			}
		}
		segments[i] = segmentJSON{
			Name:   seg.Name,
			Fields: fields,
		}
	}

	result := map[string]interface{}{
		"type":         msg.Type,
		"controlId":    msg.ControlID,
		"version":      msg.Version,
		"timestamp":    msg.Timestamp.Format("2006-01-02T15:04:05Z"),
		"sendingApp":   msg.SendingApp,
		"sendingFac":   msg.SendingFac,
		"receivingApp": msg.ReceivingApp,
		"receivingFac": msg.ReceivingFac,
		"segments":     segments,
	}

	return c.JSON(http.StatusOK, result)
}

// adtRequest is the JSON request body for ADT message generation.
type adtRequest struct {
	Event     string                 `json:"event"`
	Patient   map[string]interface{} `json:"patient"`
	Encounter map[string]interface{} `json:"encounter"`
}

// GenerateADTHandler handles POST /api/v1/hl7v2/generate/adt.
// It accepts a JSON body with FHIR Patient and Encounter resources
// and returns an HL7v2 ADT message as text/plain.
func (h *Handler) GenerateADTHandler(c echo.Context) error {
	var req adtRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if req.Event == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "event is required",
		})
	}

	data, err := GenerateADT(req.Event, req.Patient, req.Encounter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate ADT message: " + err.Error(),
		})
	}

	return c.Blob(http.StatusOK, "text/plain", data)
}

// ormRequest is the JSON request body for ORM message generation.
type ormRequest struct {
	ServiceRequest map[string]interface{} `json:"serviceRequest"`
	Patient        map[string]interface{} `json:"patient"`
}

// GenerateORMHandler handles POST /api/v1/hl7v2/generate/orm.
// It accepts a JSON body with FHIR ServiceRequest and Patient resources
// and returns an HL7v2 ORM message as text/plain.
func (h *Handler) GenerateORMHandler(c echo.Context) error {
	var req ormRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
	}

	data, err := GenerateORM(req.ServiceRequest, req.Patient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate ORM message: " + err.Error(),
		})
	}

	return c.Blob(http.StatusOK, "text/plain", data)
}

// oruRequest is the JSON request body for ORU message generation.
type oruRequest struct {
	DiagnosticReport map[string]interface{}   `json:"diagnosticReport"`
	Observations     []map[string]interface{} `json:"observations"`
	Patient          map[string]interface{}   `json:"patient"`
}

// GenerateORUHandler handles POST /api/v1/hl7v2/generate/oru.
// It accepts a JSON body with FHIR DiagnosticReport, Observations, and Patient
// and returns an HL7v2 ORU message as text/plain.
func (h *Handler) GenerateORUHandler(c echo.Context) error {
	var req oruRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
	}

	data, err := GenerateORU(req.DiagnosticReport, req.Observations, req.Patient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate ORU message: " + err.Error(),
		})
	}

	return c.Blob(http.StatusOK, "text/plain", data)
}

// decodeJSONBody reads and decodes the JSON request body into the given target.
func decodeJSONBody(c echo.Context, target interface{}) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}
