package encounter

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc  *Service
	pool *pgxpool.Pool
}

func NewHandler(svc *Service, pool *pgxpool.Pool) *Handler {
	return &Handler{svc: svc, pool: pool}
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// Read endpoints – admin, physician, nurse, registrar
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	readGroup.GET("/encounters", h.ListEncounters)
	readGroup.GET("/encounters/:id", h.GetEncounter)
	readGroup.GET("/encounters/:id/participants", h.GetParticipants)
	readGroup.GET("/encounters/:id/diagnoses", h.GetDiagnoses)
	readGroup.GET("/encounters/:id/status-history", h.GetStatusHistory)

	// Write endpoints – admin, physician, nurse, registrar
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	writeGroup.POST("/encounters", h.CreateEncounter)
	writeGroup.PUT("/encounters/:id", h.UpdateEncounter)
	writeGroup.DELETE("/encounters/:id", h.DeleteEncounter)
	writeGroup.PATCH("/encounters/:id/status", h.UpdateEncounterStatus)
	writeGroup.POST("/encounters/:id/participants", h.AddParticipant)
	writeGroup.POST("/encounters/:id/diagnoses", h.AddDiagnosis)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	fhirRead.GET("/Encounter", h.SearchEncountersFHIR)
	fhirRead.GET("/Encounter/:id", h.GetEncounterFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	fhirWrite.POST("/Encounter", h.CreateEncounterFHIR)
	fhirWrite.PUT("/Encounter/:id", h.UpdateEncounterFHIR)
	fhirWrite.DELETE("/Encounter/:id", h.DeleteEncounterFHIR)
	fhirWrite.PATCH("/Encounter/:id", h.PatchEncounterFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/Encounter/_search", h.SearchEncountersFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Encounter/:id/_history/:vid", h.VreadEncounterFHIR)
	fhirRead.GET("/Encounter/:id/_history", h.HistoryEncounterFHIR)
}

// -- Operational Handlers --

func (h *Handler) CreateEncounter(c echo.Context) error {
	var enc Encounter
	if err := c.Bind(&enc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEncounter(c.Request().Context(), &enc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, enc)
}

func (h *Handler) GetEncounter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	enc, err := h.svc.GetEncounter(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "encounter not found")
	}
	return c.JSON(http.StatusOK, enc)
}

func (h *Handler) ListEncounters(c echo.Context) error {
	pg := pagination.FromContext(c)

	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		encs, total, err := h.svc.ListEncountersByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(encs, total, pg.Limit, pg.Offset))
	}

	encs, total, err := h.svc.ListEncounters(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(encs, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateEncounter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var enc Encounter
	if err := c.Bind(&enc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	enc.ID = id
	if err := h.svc.UpdateEncounter(c.Request().Context(), &enc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, enc)
}

func (h *Handler) DeleteEncounter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEncounter(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UpdateEncounterStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.UpdateEncounterStatus(c.Request().Context(), id, body.Status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": body.Status})
}

func (h *Handler) AddParticipant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p EncounterParticipant
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.EncounterID = id
	if err := h.svc.AddParticipant(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetParticipants(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	parts, err := h.svc.GetParticipants(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, parts)
}

func (h *Handler) AddDiagnosis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d EncounterDiagnosis
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.EncounterID = id
	if err := h.svc.AddDiagnosis(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDiagnoses(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	diags, err := h.svc.GetDiagnoses(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, diags)
}

func (h *Handler) GetStatusHistory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	history, err := h.svc.GetStatusHistory(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, history)
}

// -- FHIR Encounter Handlers --

func (h *Handler) SearchEncountersFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)

	encs, total, err := h.svc.SearchEncounters(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}

	resources := make([]interface{}, len(encs))
	for i, e := range encs {
		resources[i] = e.ToFHIR()
	}
	bundle := fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/Encounter",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	})
	fhir.HandleProvenanceRevInclude(bundle, c, h.pool)
	return c.JSON(http.StatusOK, bundle)
}

func (h *Handler) GetEncounterFHIR(c echo.Context) error {
	enc, err := h.svc.GetEncounterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Encounter", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, enc.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, enc.ToFHIR())
}

func (h *Handler) CreateEncounterFHIR(c echo.Context) error {
	var enc Encounter
	if err := c.Bind(&enc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateEncounter(c.Request().Context(), &enc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Encounter/"+enc.FHIRID)
	return c.JSON(http.StatusCreated, enc.ToFHIR())
}

func (h *Handler) UpdateEncounterFHIR(c echo.Context) error {
	existing, err := h.svc.GetEncounterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Encounter", c.Param("id")))
	}
	var enc Encounter
	if err := c.Bind(&enc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	enc.ID = existing.ID
	enc.FHIRID = existing.FHIRID
	enc.PatientID = existing.PatientID
	if err := h.svc.UpdateEncounter(c.Request().Context(), &enc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, enc.ToFHIR())
}

func (h *Handler) DeleteEncounterFHIR(c echo.Context) error {
	enc, err := h.svc.GetEncounterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Encounter", c.Param("id")))
	}
	if err := h.svc.DeleteEncounter(c.Request().Context(), enc.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR PATCH Endpoints --

func (h *Handler) PatchEncounterFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetEncounterByFHIRID(ctx, c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Encounter", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome("PATCH requires application/json-patch+json or application/merge-patch+json"))
	}
	applyEncounterPatch(existing, patched)
	if err := h.svc.UpdateEncounter(ctx, existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

// -- FHIR vread and history endpoints --

func (h *Handler) VreadEncounterFHIR(c echo.Context) error {
	enc, err := h.svc.GetEncounterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Encounter", c.Param("id")))
	}
	result := enc.ToFHIR()
	fhir.SetVersionHeaders(c, 1, enc.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryEncounterFHIR(c echo.Context) error {
	enc, err := h.svc.GetEncounterByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Encounter", c.Param("id")))
	}
	result := enc.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Encounter", ResourceID: enc.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: enc.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// -- FHIR PATCH helpers --

func applyEncounterPatch(e *Encounter, patched map[string]interface{}) {
	if v, ok := patched["status"].(string); ok {
		e.Status = v
	}
	// class
	if v, ok := patched["class"]; ok {
		if cls, ok := v.(map[string]interface{}); ok {
			if code, ok := cls["code"].(string); ok {
				e.ClassCode = code
			}
			if display, ok := cls["display"].(string); ok {
				e.ClassDisplay = &display
			}
		}
	}
	// type
	if v, ok := patched["type"]; ok {
		if types, ok := v.([]interface{}); ok && len(types) > 0 {
			if t, ok := types[0].(map[string]interface{}); ok {
				if coding, ok := t["coding"].([]interface{}); ok && len(coding) > 0 {
					if c, ok := coding[0].(map[string]interface{}); ok {
						if code, ok := c["code"].(string); ok {
							e.TypeCode = &code
						}
						if display, ok := c["display"].(string); ok {
							e.TypeDisplay = &display
						}
					}
				}
			}
		}
	}
	// serviceType
	if v, ok := patched["serviceType"]; ok {
		if st, ok := v.(map[string]interface{}); ok {
			if coding, ok := st["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if code, ok := c["code"].(string); ok {
						e.ServiceTypeCode = &code
					}
					if display, ok := c["display"].(string); ok {
						e.ServiceTypeDisplay = &display
					}
				}
			}
		}
	}
	// priority
	if v, ok := patched["priority"]; ok {
		if pri, ok := v.(map[string]interface{}); ok {
			if coding, ok := pri["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if code, ok := c["code"].(string); ok {
						e.PriorityCode = &code
					}
				}
			}
		}
	}
	// period
	if v, ok := patched["period"]; ok {
		if period, ok := v.(map[string]interface{}); ok {
			if start, ok := period["start"].(string); ok {
				if t, err := time.Parse(time.RFC3339, start); err == nil {
					e.PeriodStart = t
				}
			}
			if end, ok := period["end"].(string); ok {
				if t, err := time.Parse(time.RFC3339, end); err == nil {
					e.PeriodEnd = &t
				}
			}
		}
	}
	// reasonCode
	if v, ok := patched["reasonCode"]; ok {
		if reasons, ok := v.([]interface{}); ok && len(reasons) > 0 {
			if reason, ok := reasons[0].(map[string]interface{}); ok {
				if text, ok := reason["text"].(string); ok {
					e.ReasonText = &text
				}
			}
		}
	}
}
