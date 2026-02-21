package visionprescription

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
	role := auth.RequireRole("admin", "physician")

	// REST endpoints
	rest := api.Group("", role)
	rest.GET("/vision-prescriptions", h.ListVisionPrescriptions)
	rest.POST("/vision-prescriptions", h.CreateVisionPrescription)
	rest.GET("/vision-prescriptions/:id", h.GetVisionPrescription)
	rest.PUT("/vision-prescriptions/:id", h.UpdateVisionPrescription)
	rest.DELETE("/vision-prescriptions/:id", h.DeleteVisionPrescription)
	rest.POST("/vision-prescriptions/:id/lens-specs", h.AddLensSpec)
	rest.GET("/vision-prescriptions/:id/lens-specs", h.ListLensSpecs)

	// FHIR endpoints
	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/VisionPrescription", h.SearchVisionPrescriptionsFHIR)
	fhirRead.GET("/VisionPrescription/:id", h.GetVisionPrescriptionFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/VisionPrescription", h.CreateVisionPrescriptionFHIR)
	fhirWrite.PUT("/VisionPrescription/:id", h.UpdateVisionPrescriptionFHIR)
	fhirWrite.DELETE("/VisionPrescription/:id", h.DeleteVisionPrescriptionFHIR)
	fhirWrite.PATCH("/VisionPrescription/:id", h.PatchVisionPrescriptionFHIR)

	fhirRead.POST("/VisionPrescription/_search", h.SearchVisionPrescriptionsFHIR)

	fhirRead.GET("/VisionPrescription/:id/_history/:vid", h.VreadVisionPrescriptionFHIR)
	fhirRead.GET("/VisionPrescription/:id/_history", h.HistoryVisionPrescriptionFHIR)
}

// -- REST --

func (h *Handler) CreateVisionPrescription(c echo.Context) error {
	var v VisionPrescription
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateVisionPrescription(c.Request().Context(), &v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, v)
}

func (h *Handler) GetVisionPrescription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	v, err := h.svc.GetVisionPrescription(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "vision prescription not found")
	}
	return c.JSON(http.StatusOK, v)
}

func (h *Handler) ListVisionPrescriptions(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	if p := c.QueryParam("patient_id"); p != "" {
		params["patient"] = p
	}
	if s := c.QueryParam("status"); s != "" {
		params["status"] = s
	}
	items, total, err := h.svc.SearchVisionPrescriptions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateVisionPrescription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var v VisionPrescription
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	v.ID = id
	if err := h.svc.UpdateVisionPrescription(c.Request().Context(), &v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, v)
}

func (h *Handler) DeleteVisionPrescription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteVisionPrescription(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddLensSpec(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ls VisionPrescriptionLensSpec
	if err := c.Bind(&ls); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ls.PrescriptionID = id
	if err := h.svc.AddLensSpec(c.Request().Context(), &ls); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ls)
}

func (h *Handler) ListLensSpecs(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	specs, err := h.svc.GetLensSpecs(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, specs)
}

// -- FHIR --

func (h *Handler) SearchVisionPrescriptionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchVisionPrescriptions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/VisionPrescription"))
}

func (h *Handler) GetVisionPrescriptionFHIR(c echo.Context) error {
	v, err := h.svc.GetVisionPrescriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("VisionPrescription", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, v.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, v.ToFHIR())
}

func (h *Handler) CreateVisionPrescriptionFHIR(c echo.Context) error {
	var v VisionPrescription
	if err := c.Bind(&v); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateVisionPrescription(c.Request().Context(), &v); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/VisionPrescription/"+v.FHIRID)
	return c.JSON(http.StatusCreated, v.ToFHIR())
}

func (h *Handler) UpdateVisionPrescriptionFHIR(c echo.Context) error {
	var v VisionPrescription
	if err := c.Bind(&v); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetVisionPrescriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("VisionPrescription", c.Param("id")))
	}
	v.ID = existing.ID
	v.FHIRID = existing.FHIRID
	if err := h.svc.UpdateVisionPrescription(c.Request().Context(), &v); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, v.ToFHIR())
}

func (h *Handler) DeleteVisionPrescriptionFHIR(c echo.Context) error {
	existing, err := h.svc.GetVisionPrescriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("VisionPrescription", c.Param("id")))
	}
	if err := h.svc.DeleteVisionPrescription(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchVisionPrescriptionFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadVisionPrescriptionFHIR(c echo.Context) error {
	v, err := h.svc.GetVisionPrescriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("VisionPrescription", c.Param("id")))
	}
	result := v.ToFHIR()
	fhir.SetVersionHeaders(c, 1, v.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryVisionPrescriptionFHIR(c echo.Context) error {
	v, err := h.svc.GetVisionPrescriptionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("VisionPrescription", c.Param("id")))
	}
	result := v.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "VisionPrescription", ResourceID: v.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: v.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetVisionPrescriptionByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("VisionPrescription", fhirID))
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
	if err := h.svc.UpdateVisionPrescription(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
