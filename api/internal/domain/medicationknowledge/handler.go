package medicationknowledge

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	role := auth.RequireRole("admin", "physician", "nurse")

	read := api.Group("", role)
	read.GET("/medication-knowledge", h.ListMedicationKnowledge)
	read.GET("/medication-knowledge/:id", h.GetMedicationKnowledge)

	write := api.Group("", role)
	write.POST("/medication-knowledge", h.CreateMedicationKnowledge)
	write.PUT("/medication-knowledge/:id", h.UpdateMedicationKnowledge)
	write.DELETE("/medication-knowledge/:id", h.DeleteMedicationKnowledge)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/MedicationKnowledge", h.SearchMedicationKnowledgeFHIR)
	fhirRead.GET("/MedicationKnowledge/:id", h.GetMedicationKnowledgeFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/MedicationKnowledge", h.CreateMedicationKnowledgeFHIR)
	fhirWrite.PUT("/MedicationKnowledge/:id", h.UpdateMedicationKnowledgeFHIR)
	fhirWrite.DELETE("/MedicationKnowledge/:id", h.DeleteMedicationKnowledgeFHIR)
	fhirWrite.PATCH("/MedicationKnowledge/:id", h.PatchMedicationKnowledgeFHIR)

	fhirRead.POST("/MedicationKnowledge/_search", h.SearchMedicationKnowledgeFHIR)
	fhirRead.GET("/MedicationKnowledge/:id/_history/:vid", h.VreadMedicationKnowledgeFHIR)
	fhirRead.GET("/MedicationKnowledge/:id/_history", h.HistoryMedicationKnowledgeFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateMedicationKnowledge(c echo.Context) error {
	var m MedicationKnowledge
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMedicationKnowledge(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetMedicationKnowledge(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetMedicationKnowledge(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "medication knowledge not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListMedicationKnowledge(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchMedicationKnowledge(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMedicationKnowledge(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m MedicationKnowledge
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdateMedicationKnowledge(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteMedicationKnowledge(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMedicationKnowledge(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchMedicationKnowledgeFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "code"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchMedicationKnowledge(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MedicationKnowledge"))
}

func (h *Handler) GetMedicationKnowledgeFHIR(c echo.Context) error {
	m, err := h.svc.GetMedicationKnowledgeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationKnowledge", c.Param("id")))
	}
	return c.JSON(http.StatusOK, m.ToFHIR())
}

func (h *Handler) CreateMedicationKnowledgeFHIR(c echo.Context) error {
	var m MedicationKnowledge
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMedicationKnowledge(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MedicationKnowledge/"+m.FHIRID)
	return c.JSON(http.StatusCreated, m.ToFHIR())
}

func (h *Handler) UpdateMedicationKnowledgeFHIR(c echo.Context) error {
	var m MedicationKnowledge
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMedicationKnowledgeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationKnowledge", c.Param("id")))
	}
	m.ID = existing.ID
	m.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMedicationKnowledge(c.Request().Context(), &m); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, m.ToFHIR())
}

func (h *Handler) DeleteMedicationKnowledgeFHIR(c echo.Context) error {
	existing, err := h.svc.GetMedicationKnowledgeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationKnowledge", c.Param("id")))
	}
	if err := h.svc.DeleteMedicationKnowledge(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchMedicationKnowledgeFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadMedicationKnowledgeFHIR(c echo.Context) error {
	m, err := h.svc.GetMedicationKnowledgeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationKnowledge", c.Param("id")))
	}
	result := m.ToFHIR()
	fhir.SetVersionHeaders(c, 1, m.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMedicationKnowledgeFHIR(c echo.Context) error {
	m, err := h.svc.GetMedicationKnowledgeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationKnowledge", c.Param("id")))
	}
	result := m.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MedicationKnowledge", ResourceID: m.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: m.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetMedicationKnowledgeByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MedicationKnowledge", fhirID))
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
		var mergePatch map[string]interface{}
		if err := json.Unmarshal(body, &mergePatch); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mergePatch)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateMedicationKnowledge(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
