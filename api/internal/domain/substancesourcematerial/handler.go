package substancesourcematerial

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

type Handler struct{ svc *Service }
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	role := auth.RequireRole("admin", "physician", "pharmacist")
	read := api.Group("", role)
	read.GET("/substance-source-materials", h.List)
	read.GET("/substance-source-materials/:id", h.Get)
	write := api.Group("", role)
	write.POST("/substance-source-materials", h.Create)
	write.PUT("/substance-source-materials/:id", h.Update)
	write.DELETE("/substance-source-materials/:id", h.Delete)
	fr := fhirGroup.Group("", role)
	fr.GET("/SubstanceSourceMaterial", h.SearchFHIR)
	fr.GET("/SubstanceSourceMaterial/:id", h.GetFHIR)
	fr.POST("/SubstanceSourceMaterial", h.CreateFHIR)
	fr.PUT("/SubstanceSourceMaterial/:id", h.UpdateFHIR)
	fr.DELETE("/SubstanceSourceMaterial/:id", h.DeleteFHIR)
	fr.PATCH("/SubstanceSourceMaterial/:id", h.PatchFHIR)
	fr.POST("/SubstanceSourceMaterial/_search", h.SearchFHIR)
	fr.GET("/SubstanceSourceMaterial/:id/_history/:vid", h.VreadFHIR)
	fr.GET("/SubstanceSourceMaterial/:id/_history", h.HistoryFHIR)
}

func (h *Handler) Create(c echo.Context) error { var m SubstanceSourceMaterial; if err := c.Bind(&m); err != nil { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }; if err := h.svc.Create(c.Request().Context(), &m); err != nil { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }; return c.JSON(http.StatusCreated, m) }
func (h *Handler) Get(c echo.Context) error { id, err := uuid.Parse(c.Param("id")); if err != nil { return echo.NewHTTPError(http.StatusBadRequest, "invalid id") }; m, err := h.svc.GetByID(c.Request().Context(), id); if err != nil { return echo.NewHTTPError(http.StatusNotFound, "not found") }; return c.JSON(http.StatusOK, m) }
func (h *Handler) List(c echo.Context) error { pg := pagination.FromContext(c); items, total, err := h.svc.Search(c.Request().Context(), nil, pg.Limit, pg.Offset); if err != nil { return echo.NewHTTPError(http.StatusInternalServerError, err.Error()) }; return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset)) }
func (h *Handler) Update(c echo.Context) error { id, err := uuid.Parse(c.Param("id")); if err != nil { return echo.NewHTTPError(http.StatusBadRequest, "invalid id") }; var m SubstanceSourceMaterial; if err := c.Bind(&m); err != nil { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }; m.ID = id; if err := h.svc.Update(c.Request().Context(), &m); err != nil { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }; return c.JSON(http.StatusOK, m) }
func (h *Handler) Delete(c echo.Context) error { id, err := uuid.Parse(c.Param("id")); if err != nil { return echo.NewHTTPError(http.StatusBadRequest, "invalid id") }; if err := h.svc.Delete(c.Request().Context(), id); err != nil { return echo.NewHTTPError(http.StatusInternalServerError, err.Error()) }; return c.NoContent(http.StatusNoContent) }

func (h *Handler) SearchFHIR(c echo.Context) error { pg := pagination.FromContext(c); params := map[string]string{}; items, total, err := h.svc.Search(c.Request().Context(), params, pg.Limit, pg.Offset); if err != nil { return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error())) }; resources := make([]interface{}, len(items)); for i, item := range items { resources[i] = item.ToFHIR() }; return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/SubstanceSourceMaterial")) }
func (h *Handler) GetFHIR(c echo.Context) error { m, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id")); if err != nil { return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSourceMaterial", c.Param("id"))) }; return c.JSON(http.StatusOK, m.ToFHIR()) }
func (h *Handler) CreateFHIR(c echo.Context) error { var m SubstanceSourceMaterial; if err := c.Bind(&m); err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error())) }; if err := h.svc.Create(c.Request().Context(), &m); err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error())) }; c.Response().Header().Set("Location", "/fhir/SubstanceSourceMaterial/"+m.FHIRID); return c.JSON(http.StatusCreated, m.ToFHIR()) }
func (h *Handler) UpdateFHIR(c echo.Context) error { var m SubstanceSourceMaterial; if err := c.Bind(&m); err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error())) }; existing, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id")); if err != nil { return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSourceMaterial", c.Param("id"))) }; m.ID = existing.ID; m.FHIRID = existing.FHIRID; if err := h.svc.Update(c.Request().Context(), &m); err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error())) }; return c.JSON(http.StatusOK, m.ToFHIR()) }
func (h *Handler) DeleteFHIR(c echo.Context) error { existing, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id")); if err != nil { return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSourceMaterial", c.Param("id"))) }; if err := h.svc.Delete(c.Request().Context(), existing.ID); err != nil { return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error())) }; return c.NoContent(http.StatusNoContent) }
func (h *Handler) PatchFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type"); body, err := io.ReadAll(c.Request().Body); if err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body")) }
	existing, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id")); if err != nil { return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSourceMaterial", c.Param("id"))) }
	currentResource := existing.ToFHIR(); var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") { ops, err := fhir.ParseJSONPatch(body); if err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error())) }; patched, err = fhir.ApplyJSONPatch(currentResource, ops); if err != nil { return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error())) }
	} else if strings.Contains(contentType, "merge-patch+json") { var mp map[string]interface{}; if err := json.Unmarshal(body, &mp); err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error())) }; patched, err = fhir.ApplyMergePatch(currentResource, mp); if err != nil { return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error())) }
	} else { return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome("PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json")) }
	_ = patched; if err := h.svc.Update(c.Request().Context(), existing); err != nil { return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error())) }; return c.JSON(http.StatusOK, existing.ToFHIR())
}
func (h *Handler) VreadFHIR(c echo.Context) error { m, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id")); if err != nil { return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSourceMaterial", c.Param("id"))) }; fhir.SetVersionHeaders(c, 1, m.UpdatedAt.Format("2006-01-02T15:04:05Z")); return c.JSON(http.StatusOK, m.ToFHIR()) }
func (h *Handler) HistoryFHIR(c echo.Context) error { m, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id")); if err != nil { return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("SubstanceSourceMaterial", c.Param("id"))) }; raw, _ := json.Marshal(m.ToFHIR()); entry := &fhir.HistoryEntry{ResourceType: "SubstanceSourceMaterial", ResourceID: m.FHIRID, VersionID: 1, Resource: raw, Action: "create", Timestamp: m.CreatedAt}; return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir")) }
