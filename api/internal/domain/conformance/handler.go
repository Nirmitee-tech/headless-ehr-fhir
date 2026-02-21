package conformance

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
	role := auth.RequireRole("admin")

	// REST endpoints
	read := api.Group("", role)
	read.GET("/naming-systems", h.ListNamingSystems)
	read.GET("/naming-systems/:id", h.GetNamingSystem)
	read.GET("/naming-systems/:id/unique-ids", h.GetNamingSystemUniqueIDs)
	read.GET("/operation-definitions", h.ListOperationDefinitions)
	read.GET("/operation-definitions/:id", h.GetOperationDefinition)
	read.GET("/operation-definitions/:id/parameters", h.GetOperationDefinitionParameters)
	read.GET("/message-definitions", h.ListMessageDefinitions)
	read.GET("/message-definitions/:id", h.GetMessageDefinition)
	read.GET("/message-headers", h.ListMessageHeaders)
	read.GET("/message-headers/:id", h.GetMessageHeader)

	write := api.Group("", role)
	write.POST("/naming-systems", h.CreateNamingSystem)
	write.PUT("/naming-systems/:id", h.UpdateNamingSystem)
	write.DELETE("/naming-systems/:id", h.DeleteNamingSystem)
	write.POST("/naming-systems/:id/unique-ids", h.AddNamingSystemUniqueID)
	write.POST("/operation-definitions", h.CreateOperationDefinition)
	write.PUT("/operation-definitions/:id", h.UpdateOperationDefinition)
	write.DELETE("/operation-definitions/:id", h.DeleteOperationDefinition)
	write.POST("/operation-definitions/:id/parameters", h.AddOperationDefinitionParameter)
	write.POST("/message-definitions", h.CreateMessageDefinition)
	write.PUT("/message-definitions/:id", h.UpdateMessageDefinition)
	write.DELETE("/message-definitions/:id", h.DeleteMessageDefinition)
	write.POST("/message-headers", h.CreateMessageHeader)
	write.PUT("/message-headers/:id", h.UpdateMessageHeader)
	write.DELETE("/message-headers/:id", h.DeleteMessageHeader)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", role)
	fhirRead.GET("/NamingSystem", h.SearchNamingSystemsFHIR)
	fhirRead.GET("/NamingSystem/:id", h.GetNamingSystemFHIR)
	fhirRead.GET("/OperationDefinition", h.SearchOperationDefinitionsFHIR)
	fhirRead.GET("/OperationDefinition/:id", h.GetOperationDefinitionFHIR)
	fhirRead.GET("/MessageDefinition", h.SearchMessageDefinitionsFHIR)
	fhirRead.GET("/MessageDefinition/:id", h.GetMessageDefinitionFHIR)
	fhirRead.GET("/MessageHeader", h.SearchMessageHeadersFHIR)
	fhirRead.GET("/MessageHeader/:id", h.GetMessageHeaderFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", role)
	fhirWrite.POST("/NamingSystem", h.CreateNamingSystemFHIR)
	fhirWrite.PUT("/NamingSystem/:id", h.UpdateNamingSystemFHIR)
	fhirWrite.DELETE("/NamingSystem/:id", h.DeleteNamingSystemFHIR)
	fhirWrite.PATCH("/NamingSystem/:id", h.PatchNamingSystemFHIR)
	fhirWrite.POST("/OperationDefinition", h.CreateOperationDefinitionFHIR)
	fhirWrite.PUT("/OperationDefinition/:id", h.UpdateOperationDefinitionFHIR)
	fhirWrite.DELETE("/OperationDefinition/:id", h.DeleteOperationDefinitionFHIR)
	fhirWrite.PATCH("/OperationDefinition/:id", h.PatchOperationDefinitionFHIR)
	fhirWrite.POST("/MessageDefinition", h.CreateMessageDefinitionFHIR)
	fhirWrite.PUT("/MessageDefinition/:id", h.UpdateMessageDefinitionFHIR)
	fhirWrite.DELETE("/MessageDefinition/:id", h.DeleteMessageDefinitionFHIR)
	fhirWrite.PATCH("/MessageDefinition/:id", h.PatchMessageDefinitionFHIR)
	fhirWrite.POST("/MessageHeader", h.CreateMessageHeaderFHIR)
	fhirWrite.PUT("/MessageHeader/:id", h.UpdateMessageHeaderFHIR)
	fhirWrite.DELETE("/MessageHeader/:id", h.DeleteMessageHeaderFHIR)
	fhirWrite.PATCH("/MessageHeader/:id", h.PatchMessageHeaderFHIR)

	// FHIR _search endpoints
	fhirRead.POST("/NamingSystem/_search", h.SearchNamingSystemsFHIR)
	fhirRead.POST("/OperationDefinition/_search", h.SearchOperationDefinitionsFHIR)
	fhirRead.POST("/MessageDefinition/_search", h.SearchMessageDefinitionsFHIR)
	fhirRead.POST("/MessageHeader/_search", h.SearchMessageHeadersFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/NamingSystem/:id/_history/:vid", h.VreadNamingSystemFHIR)
	fhirRead.GET("/NamingSystem/:id/_history", h.HistoryNamingSystemFHIR)
	fhirRead.GET("/OperationDefinition/:id/_history/:vid", h.VreadOperationDefinitionFHIR)
	fhirRead.GET("/OperationDefinition/:id/_history", h.HistoryOperationDefinitionFHIR)
	fhirRead.GET("/MessageDefinition/:id/_history/:vid", h.VreadMessageDefinitionFHIR)
	fhirRead.GET("/MessageDefinition/:id/_history", h.HistoryMessageDefinitionFHIR)
	fhirRead.GET("/MessageHeader/:id/_history/:vid", h.VreadMessageHeaderFHIR)
	fhirRead.GET("/MessageHeader/:id/_history", h.HistoryMessageHeaderFHIR)
}

// ==================== NamingSystem REST ====================

func (h *Handler) CreateNamingSystem(c echo.Context) error {
	var ns NamingSystem
	if err := c.Bind(&ns); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateNamingSystem(c.Request().Context(), &ns); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ns)
}

func (h *Handler) GetNamingSystem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ns, err := h.svc.GetNamingSystem(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "naming system not found")
	}
	return c.JSON(http.StatusOK, ns)
}

func (h *Handler) ListNamingSystems(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListNamingSystems(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateNamingSystem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ns NamingSystem
	if err := c.Bind(&ns); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ns.ID = id
	if err := h.svc.UpdateNamingSystem(c.Request().Context(), &ns); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ns)
}

func (h *Handler) DeleteNamingSystem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteNamingSystem(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddNamingSystemUniqueID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var uid NamingSystemUniqueID
	if err := c.Bind(&uid); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	uid.NamingSystemID = id
	if err := h.svc.AddNamingSystemUniqueID(c.Request().Context(), &uid); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, uid)
}

func (h *Handler) GetNamingSystemUniqueIDs(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetNamingSystemUniqueIDs(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// ==================== OperationDefinition REST ====================

func (h *Handler) CreateOperationDefinition(c echo.Context) error {
	var od OperationDefinition
	if err := c.Bind(&od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateOperationDefinition(c.Request().Context(), &od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, od)
}

func (h *Handler) GetOperationDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	od, err := h.svc.GetOperationDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "operation definition not found")
	}
	return c.JSON(http.StatusOK, od)
}

func (h *Handler) ListOperationDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListOperationDefinitions(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateOperationDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var od OperationDefinition
	if err := c.Bind(&od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	od.ID = id
	if err := h.svc.UpdateOperationDefinition(c.Request().Context(), &od); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, od)
}

func (h *Handler) DeleteOperationDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteOperationDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddOperationDefinitionParameter(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p OperationDefinitionParameter
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.OperationDefinitionID = id
	if err := h.svc.AddOperationDefinitionParameter(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetOperationDefinitionParameters(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetOperationDefinitionParameters(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// ==================== MessageDefinition REST ====================

func (h *Handler) CreateMessageDefinition(c echo.Context) error {
	var md MessageDefinition
	if err := c.Bind(&md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMessageDefinition(c.Request().Context(), &md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, md)
}

func (h *Handler) GetMessageDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	md, err := h.svc.GetMessageDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "message definition not found")
	}
	return c.JSON(http.StatusOK, md)
}

func (h *Handler) ListMessageDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListMessageDefinitions(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMessageDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var md MessageDefinition
	if err := c.Bind(&md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	md.ID = id
	if err := h.svc.UpdateMessageDefinition(c.Request().Context(), &md); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, md)
}

func (h *Handler) DeleteMessageDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMessageDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ==================== MessageHeader REST ====================

func (h *Handler) CreateMessageHeader(c echo.Context) error {
	var mh MessageHeader
	if err := c.Bind(&mh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMessageHeader(c.Request().Context(), &mh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, mh)
}

func (h *Handler) GetMessageHeader(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	mh, err := h.svc.GetMessageHeader(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "message header not found")
	}
	return c.JSON(http.StatusOK, mh)
}

func (h *Handler) ListMessageHeaders(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListMessageHeaders(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMessageHeader(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var mh MessageHeader
	if err := c.Bind(&mh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	mh.ID = id
	if err := h.svc.UpdateMessageHeader(c.Request().Context(), &mh); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, mh)
}

func (h *Handler) DeleteMessageHeader(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMessageHeader(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ==================== FHIR NamingSystem ====================

func (h *Handler) SearchNamingSystemsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchNamingSystems(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/NamingSystem"))
}

func (h *Handler) GetNamingSystemFHIR(c echo.Context) error {
	ns, err := h.svc.GetNamingSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NamingSystem", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, ns.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, ns.ToFHIR())
}

func (h *Handler) CreateNamingSystemFHIR(c echo.Context) error {
	var ns NamingSystem
	if err := c.Bind(&ns); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateNamingSystem(c.Request().Context(), &ns); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/NamingSystem/"+ns.FHIRID)
	return c.JSON(http.StatusCreated, ns.ToFHIR())
}

func (h *Handler) UpdateNamingSystemFHIR(c echo.Context) error {
	var ns NamingSystem
	if err := c.Bind(&ns); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetNamingSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NamingSystem", c.Param("id")))
	}
	ns.ID = existing.ID
	ns.FHIRID = existing.FHIRID
	if err := h.svc.UpdateNamingSystem(c.Request().Context(), &ns); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ns.ToFHIR())
}

func (h *Handler) DeleteNamingSystemFHIR(c echo.Context) error {
	existing, err := h.svc.GetNamingSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NamingSystem", c.Param("id")))
	}
	if err := h.svc.DeleteNamingSystem(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchNamingSystemFHIR(c echo.Context) error {
	return h.handlePatch(c, "NamingSystem", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetNamingSystemByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NamingSystem", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if v, ok := resource["name"].(string); ok {
			existing.Name = v
		}
		if err := h.svc.UpdateNamingSystem(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadNamingSystemFHIR(c echo.Context) error {
	ns, err := h.svc.GetNamingSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NamingSystem", c.Param("id")))
	}
	result := ns.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ns.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryNamingSystemFHIR(c echo.Context) error {
	ns, err := h.svc.GetNamingSystemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NamingSystem", c.Param("id")))
	}
	result := ns.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "NamingSystem", ResourceID: ns.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ns.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ==================== FHIR OperationDefinition ====================

func (h *Handler) SearchOperationDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchOperationDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/OperationDefinition"))
}

func (h *Handler) GetOperationDefinitionFHIR(c echo.Context) error {
	od, err := h.svc.GetOperationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OperationDefinition", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, od.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, od.ToFHIR())
}

func (h *Handler) CreateOperationDefinitionFHIR(c echo.Context) error {
	var od OperationDefinition
	if err := c.Bind(&od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateOperationDefinition(c.Request().Context(), &od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/OperationDefinition/"+od.FHIRID)
	return c.JSON(http.StatusCreated, od.ToFHIR())
}

func (h *Handler) UpdateOperationDefinitionFHIR(c echo.Context) error {
	var od OperationDefinition
	if err := c.Bind(&od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetOperationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OperationDefinition", c.Param("id")))
	}
	od.ID = existing.ID
	od.FHIRID = existing.FHIRID
	if err := h.svc.UpdateOperationDefinition(c.Request().Context(), &od); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, od.ToFHIR())
}

func (h *Handler) DeleteOperationDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetOperationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OperationDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteOperationDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchOperationDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, "OperationDefinition", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetOperationDefinitionByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OperationDefinition", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if v, ok := resource["name"].(string); ok {
			existing.Name = v
		}
		if err := h.svc.UpdateOperationDefinition(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadOperationDefinitionFHIR(c echo.Context) error {
	od, err := h.svc.GetOperationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OperationDefinition", c.Param("id")))
	}
	result := od.ToFHIR()
	fhir.SetVersionHeaders(c, 1, od.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryOperationDefinitionFHIR(c echo.Context) error {
	od, err := h.svc.GetOperationDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("OperationDefinition", c.Param("id")))
	}
	result := od.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "OperationDefinition", ResourceID: od.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: od.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ==================== FHIR MessageDefinition ====================

func (h *Handler) SearchMessageDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMessageDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MessageDefinition"))
}

func (h *Handler) GetMessageDefinitionFHIR(c echo.Context) error {
	md, err := h.svc.GetMessageDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageDefinition", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, md.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, md.ToFHIR())
}

func (h *Handler) CreateMessageDefinitionFHIR(c echo.Context) error {
	var md MessageDefinition
	if err := c.Bind(&md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMessageDefinition(c.Request().Context(), &md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MessageDefinition/"+md.FHIRID)
	return c.JSON(http.StatusCreated, md.ToFHIR())
}

func (h *Handler) UpdateMessageDefinitionFHIR(c echo.Context) error {
	var md MessageDefinition
	if err := c.Bind(&md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMessageDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageDefinition", c.Param("id")))
	}
	md.ID = existing.ID
	md.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMessageDefinition(c.Request().Context(), &md); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, md.ToFHIR())
}

func (h *Handler) DeleteMessageDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetMessageDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteMessageDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchMessageDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, "MessageDefinition", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetMessageDefinitionByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageDefinition", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateMessageDefinition(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadMessageDefinitionFHIR(c echo.Context) error {
	md, err := h.svc.GetMessageDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageDefinition", c.Param("id")))
	}
	result := md.ToFHIR()
	fhir.SetVersionHeaders(c, 1, md.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMessageDefinitionFHIR(c echo.Context) error {
	md, err := h.svc.GetMessageDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageDefinition", c.Param("id")))
	}
	result := md.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MessageDefinition", ResourceID: md.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: md.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ==================== FHIR MessageHeader ====================

func (h *Handler) SearchMessageHeadersFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchMessageHeaders(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/MessageHeader"))
}

func (h *Handler) GetMessageHeaderFHIR(c echo.Context) error {
	mh, err := h.svc.GetMessageHeaderByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageHeader", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, mh.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, mh.ToFHIR())
}

func (h *Handler) CreateMessageHeaderFHIR(c echo.Context) error {
	var mh MessageHeader
	if err := c.Bind(&mh); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateMessageHeader(c.Request().Context(), &mh); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/MessageHeader/"+mh.FHIRID)
	return c.JSON(http.StatusCreated, mh.ToFHIR())
}

func (h *Handler) UpdateMessageHeaderFHIR(c echo.Context) error {
	var mh MessageHeader
	if err := c.Bind(&mh); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetMessageHeaderByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageHeader", c.Param("id")))
	}
	mh.ID = existing.ID
	mh.FHIRID = existing.FHIRID
	if err := h.svc.UpdateMessageHeader(c.Request().Context(), &mh); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, mh.ToFHIR())
}

func (h *Handler) DeleteMessageHeaderFHIR(c echo.Context) error {
	existing, err := h.svc.GetMessageHeaderByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageHeader", c.Param("id")))
	}
	if err := h.svc.DeleteMessageHeader(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchMessageHeaderFHIR(c echo.Context) error {
	return h.handlePatch(c, "MessageHeader", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetMessageHeaderByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageHeader", ctx.Param("id")))
		}
		if v, ok := resource["eventCoding"].(map[string]interface{}); ok {
			if code, ok := v["code"].(string); ok {
				existing.EventCodingCode = code
			}
		}
		if err := h.svc.UpdateMessageHeader(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadMessageHeaderFHIR(c echo.Context) error {
	mh, err := h.svc.GetMessageHeaderByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageHeader", c.Param("id")))
	}
	result := mh.ToFHIR()
	fhir.SetVersionHeaders(c, 1, mh.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryMessageHeaderFHIR(c echo.Context) error {
	mh, err := h.svc.GetMessageHeaderByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("MessageHeader", c.Param("id")))
	}
	result := mh.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "MessageHeader", ResourceID: mh.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: mh.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ==================== handlePatch helper ====================

func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	var currentResource map[string]interface{}
	switch resourceType {
	case "NamingSystem":
		existing, err := h.svc.GetNamingSystemByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "OperationDefinition":
		existing, err := h.svc.GetOperationDefinitionByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "MessageDefinition":
		existing, err := h.svc.GetMessageDefinitionByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "MessageHeader":
		existing, err := h.svc.GetMessageHeaderByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	default:
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("unsupported resource type for PATCH"))
	}

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

	return applyFn(c, patched)
}
