package biologicallyderivedproduct

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
	read.GET("/biologically-derived-products", h.ListBiologicallyDerivedProducts)
	read.GET("/biologically-derived-products/:id", h.GetBiologicallyDerivedProduct)

	write := api.Group("", role)
	write.POST("/biologically-derived-products", h.CreateBiologicallyDerivedProduct)
	write.PUT("/biologically-derived-products/:id", h.UpdateBiologicallyDerivedProduct)
	write.DELETE("/biologically-derived-products/:id", h.DeleteBiologicallyDerivedProduct)

	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/BiologicallyDerivedProduct", h.SearchBiologicallyDerivedProductsFHIR)
	fhirRead.GET("/BiologicallyDerivedProduct/:id", h.GetBiologicallyDerivedProductFHIR)

	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/BiologicallyDerivedProduct", h.CreateBiologicallyDerivedProductFHIR)
	fhirWrite.PUT("/BiologicallyDerivedProduct/:id", h.UpdateBiologicallyDerivedProductFHIR)
	fhirWrite.DELETE("/BiologicallyDerivedProduct/:id", h.DeleteBiologicallyDerivedProductFHIR)
	fhirWrite.PATCH("/BiologicallyDerivedProduct/:id", h.PatchBiologicallyDerivedProductFHIR)

	fhirRead.POST("/BiologicallyDerivedProduct/_search", h.SearchBiologicallyDerivedProductsFHIR)
	fhirRead.GET("/BiologicallyDerivedProduct/:id/_history/:vid", h.VreadBiologicallyDerivedProductFHIR)
	fhirRead.GET("/BiologicallyDerivedProduct/:id/_history", h.HistoryBiologicallyDerivedProductFHIR)
}

// -- REST Endpoints --

func (h *Handler) CreateBiologicallyDerivedProduct(c echo.Context) error {
	var b BiologicallyDerivedProduct
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateBiologicallyDerivedProduct(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, b)
}

func (h *Handler) GetBiologicallyDerivedProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	b, err := h.svc.GetBiologicallyDerivedProduct(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "biologically derived product not found")
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) ListBiologicallyDerivedProducts(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchBiologicallyDerivedProducts(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateBiologicallyDerivedProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var b BiologicallyDerivedProduct
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	b.ID = id
	if err := h.svc.UpdateBiologicallyDerivedProduct(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) DeleteBiologicallyDerivedProduct(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteBiologicallyDerivedProduct(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchBiologicallyDerivedProductsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchBiologicallyDerivedProducts(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/BiologicallyDerivedProduct"))
}

func (h *Handler) GetBiologicallyDerivedProductFHIR(c echo.Context) error {
	b, err := h.svc.GetBiologicallyDerivedProductByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BiologicallyDerivedProduct", c.Param("id")))
	}
	return c.JSON(http.StatusOK, b.ToFHIR())
}

func (h *Handler) CreateBiologicallyDerivedProductFHIR(c echo.Context) error {
	var b BiologicallyDerivedProduct
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateBiologicallyDerivedProduct(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/BiologicallyDerivedProduct/"+b.FHIRID)
	return c.JSON(http.StatusCreated, b.ToFHIR())
}

func (h *Handler) UpdateBiologicallyDerivedProductFHIR(c echo.Context) error {
	var b BiologicallyDerivedProduct
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetBiologicallyDerivedProductByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BiologicallyDerivedProduct", c.Param("id")))
	}
	b.ID = existing.ID
	b.FHIRID = existing.FHIRID
	if err := h.svc.UpdateBiologicallyDerivedProduct(c.Request().Context(), &b); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, b.ToFHIR())
}

func (h *Handler) DeleteBiologicallyDerivedProductFHIR(c echo.Context) error {
	existing, err := h.svc.GetBiologicallyDerivedProductByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BiologicallyDerivedProduct", c.Param("id")))
	}
	if err := h.svc.DeleteBiologicallyDerivedProduct(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchBiologicallyDerivedProductFHIR(c echo.Context) error {
	return h.handlePatch(c, c.Param("id"))
}

func (h *Handler) VreadBiologicallyDerivedProductFHIR(c echo.Context) error {
	b, err := h.svc.GetBiologicallyDerivedProductByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BiologicallyDerivedProduct", c.Param("id")))
	}
	result := b.ToFHIR()
	fhir.SetVersionHeaders(c, 1, b.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryBiologicallyDerivedProductFHIR(c echo.Context) error {
	b, err := h.svc.GetBiologicallyDerivedProductByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BiologicallyDerivedProduct", c.Param("id")))
	}
	result := b.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "BiologicallyDerivedProduct", ResourceID: b.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: b.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) handlePatch(c echo.Context, fhirID string) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetBiologicallyDerivedProductByFHIRID(c.Request().Context(), fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("BiologicallyDerivedProduct", fhirID))
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
		existing.Status = &v
	}
	if err := h.svc.UpdateBiologicallyDerivedProduct(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}
