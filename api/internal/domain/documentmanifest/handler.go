package documentmanifest

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
	read.GET("/document-manifests", h.ListDocumentManifests)
	read.GET("/document-manifests/:id", h.GetDocumentManifest)

	write := api.Group("", role)
	write.POST("/document-manifests", h.CreateDocumentManifest)
	write.PUT("/document-manifests/:id", h.UpdateDocumentManifest)
	write.DELETE("/document-manifests/:id", h.DeleteDocumentManifest)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/DocumentManifest", h.SearchDocumentManifestsFHIR)
	fhirRead.GET("/DocumentManifest/:id", h.GetDocumentManifestFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/DocumentManifest", h.CreateDocumentManifestFHIR)
	fhirWrite.PUT("/DocumentManifest/:id", h.UpdateDocumentManifestFHIR)
	fhirWrite.DELETE("/DocumentManifest/:id", h.DeleteDocumentManifestFHIR)
	fhirWrite.PATCH("/DocumentManifest/:id", h.PatchDocumentManifestFHIR)

	fhirRead.POST("/DocumentManifest/_search", h.SearchDocumentManifestsFHIR)
	fhirRead.GET("/DocumentManifest/:id/_history/:vid", h.VreadDocumentManifestFHIR)
	fhirRead.GET("/DocumentManifest/:id/_history", h.HistoryDocumentManifestFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateDocumentManifest(c echo.Context) error {
	var d DocumentManifest
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDocumentManifest(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDocumentManifest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetDocumentManifest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "document manifest not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) ListDocumentManifests(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchDocumentManifests(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDocumentManifest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d DocumentManifest
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateDocumentManifest(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteDocumentManifest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDocumentManifest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchDocumentManifestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchDocumentManifests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/DocumentManifest"))
}

func (h *Handler) GetDocumentManifestFHIR(c echo.Context) error {
	d, err := h.svc.GetDocumentManifestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentManifest", c.Param("id")))
	}
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *Handler) CreateDocumentManifestFHIR(c echo.Context) error {
	var d DocumentManifest
	if err := c.Bind(&d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateDocumentManifest(c.Request().Context(), &d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/DocumentManifest/"+d.FHIRID)
	return c.JSON(http.StatusCreated, d.ToFHIR())
}

func (h *Handler) UpdateDocumentManifestFHIR(c echo.Context) error {
	var d DocumentManifest
	if err := c.Bind(&d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetDocumentManifestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentManifest", c.Param("id")))
	}
	d.ID = existing.ID
	d.FHIRID = existing.FHIRID
	if err := h.svc.UpdateDocumentManifest(c.Request().Context(), &d); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, d.ToFHIR())
}

func (h *Handler) DeleteDocumentManifestFHIR(c echo.Context) error {
	existing, err := h.svc.GetDocumentManifestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentManifest", c.Param("id")))
	}
	if err := h.svc.DeleteDocumentManifest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchDocumentManifestFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadDocumentManifestFHIR(c echo.Context) error {
	d, err := h.svc.GetDocumentManifestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentManifest", c.Param("id")))
	}
	result := d.ToFHIR()
	fhir.SetVersionHeaders(c, 1, d.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryDocumentManifestFHIR(c echo.Context) error {
	d, err := h.svc.GetDocumentManifestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentManifest", c.Param("id")))
	}
	result := d.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "DocumentManifest", ResourceID: d.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: d.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetDocumentManifestByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentManifest", fhirID))
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
	if err := h.svc.UpdateDocumentManifest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
