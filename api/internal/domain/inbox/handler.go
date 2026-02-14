package inbox

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// Read endpoints – admin, physician, nurse, pharmacist
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "pharmacist"))
	readGroup.GET("/message-pools", h.ListMessagePools)
	readGroup.GET("/message-pools/:id", h.GetMessagePool)
	readGroup.GET("/message-pools/:id/members", h.GetPoolMembers)
	readGroup.GET("/inbox-messages", h.ListInboxMessages)
	readGroup.GET("/inbox-messages/:id", h.GetInboxMessage)
	readGroup.GET("/cosign-requests", h.ListCosignRequests)
	readGroup.GET("/cosign-requests/:id", h.GetCosignRequest)
	readGroup.GET("/patient-lists", h.ListPatientLists)
	readGroup.GET("/patient-lists/:id", h.GetPatientList)
	readGroup.GET("/patient-lists/:id/members", h.GetPatientListMembers)
	readGroup.GET("/handoffs", h.ListHandoffs)
	readGroup.GET("/handoffs/:id", h.GetHandoff)

	// Write endpoints – admin, physician, nurse, pharmacist
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "pharmacist"))
	writeGroup.POST("/message-pools", h.CreateMessagePool)
	writeGroup.PUT("/message-pools/:id", h.UpdateMessagePool)
	writeGroup.DELETE("/message-pools/:id", h.DeleteMessagePool)
	writeGroup.POST("/message-pools/:id/members", h.AddPoolMember)
	writeGroup.DELETE("/message-pools/:id/members/:memberID", h.RemovePoolMember)
	writeGroup.POST("/inbox-messages", h.CreateInboxMessage)
	writeGroup.PUT("/inbox-messages/:id", h.UpdateInboxMessage)
	writeGroup.DELETE("/inbox-messages/:id", h.DeleteInboxMessage)
	writeGroup.POST("/cosign-requests", h.CreateCosignRequest)
	writeGroup.PUT("/cosign-requests/:id", h.UpdateCosignRequest)
	writeGroup.POST("/patient-lists", h.CreatePatientList)
	writeGroup.PUT("/patient-lists/:id", h.UpdatePatientList)
	writeGroup.DELETE("/patient-lists/:id", h.DeletePatientList)
	writeGroup.POST("/patient-lists/:id/members", h.AddPatientListMember)
	writeGroup.PUT("/patient-lists/:id/members/:memberID", h.UpdatePatientListMember)
	writeGroup.DELETE("/patient-lists/:id/members/:memberID", h.RemovePatientListMember)
	writeGroup.POST("/handoffs", h.CreateHandoff)
	writeGroup.PUT("/handoffs/:id", h.UpdateHandoff)

	// fhirGroup is accepted but not used; these are operational-only resources.
}

// -- Message Pool Handlers --

func (h *Handler) CreateMessagePool(c echo.Context) error {
	var p MessagePool
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateMessagePool(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetMessagePool(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetMessagePool(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "message pool not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListMessagePools(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListMessagePools(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateMessagePool(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p MessagePool
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdateMessagePool(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteMessagePool(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteMessagePool(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPoolMember(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m MessagePoolMember
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.PoolID = id
	if err := h.svc.AddPoolMember(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetPoolMembers(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	members, err := h.svc.GetPoolMembers(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, members)
}

func (h *Handler) RemovePoolMember(c echo.Context) error {
	memberID, err := uuid.Parse(c.Param("memberID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid member id")
	}
	if err := h.svc.RemovePoolMember(c.Request().Context(), memberID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Inbox Message Handlers --

func (h *Handler) CreateInboxMessage(c echo.Context) error {
	var m InboxMessage
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateInboxMessage(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetInboxMessage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetInboxMessage(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "inbox message not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListInboxMessages(c echo.Context) error {
	pg := pagination.FromContext(c)
	if recipientID := c.QueryParam("recipient_id"); recipientID != "" {
		rid, err := uuid.Parse(recipientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid recipient_id")
		}
		items, total, err := h.svc.ListInboxMessagesByRecipient(c.Request().Context(), rid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListInboxMessagesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchInboxMessages(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateInboxMessage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m InboxMessage
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdateInboxMessage(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteInboxMessage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteInboxMessage(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Cosign Request Handlers --

func (h *Handler) CreateCosignRequest(c echo.Context) error {
	var r CosignRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCosignRequest(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetCosignRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetCosignRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cosign request not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListCosignRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	if cosignerID := c.QueryParam("cosigner_id"); cosignerID != "" {
		cid, err := uuid.Parse(cosignerID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cosigner_id")
		}
		items, total, err := h.svc.ListCosignRequestsByCosigner(c.Request().Context(), cid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if requesterID := c.QueryParam("requester_id"); requesterID != "" {
		rid, err := uuid.Parse(requesterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid requester_id")
		}
		items, total, err := h.svc.ListCosignRequestsByRequester(c.Request().Context(), rid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "cosigner_id or requester_id query parameter is required")
}

func (h *Handler) UpdateCosignRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r CosignRequest
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateCosignRequest(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

// -- Patient List Handlers --

func (h *Handler) CreatePatientList(c echo.Context) error {
	var l PatientList
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePatientList(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, l)
}

func (h *Handler) GetPatientList(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	l, err := h.svc.GetPatientList(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "patient list not found")
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) ListPatientLists(c echo.Context) error {
	pg := pagination.FromContext(c)
	ownerID := c.QueryParam("owner_id")
	if ownerID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "owner_id query parameter is required")
	}
	oid, err := uuid.Parse(ownerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid owner_id")
	}
	items, total, err := h.svc.ListPatientListsByOwner(c.Request().Context(), oid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePatientList(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var l PatientList
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	l.ID = id
	if err := h.svc.UpdatePatientList(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) DeletePatientList(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePatientList(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPatientListMember(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m PatientListMember
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ListID = id
	if err := h.svc.AddPatientListMember(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetPatientListMembers(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pg := pagination.FromContext(c)
	members, total, err := h.svc.GetPatientListMembers(c.Request().Context(), id, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(members, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePatientListMember(c echo.Context) error {
	memberID, err := uuid.Parse(c.Param("memberID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid member id")
	}
	var m PatientListMember
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = memberID
	if err := h.svc.UpdatePatientListMember(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) RemovePatientListMember(c echo.Context) error {
	memberID, err := uuid.Parse(c.Param("memberID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid member id")
	}
	if err := h.svc.RemovePatientListMember(c.Request().Context(), memberID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Handoff Handlers --

func (h *Handler) CreateHandoff(c echo.Context) error {
	var ho HandoffRecord
	if err := c.Bind(&ho); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateHandoff(c.Request().Context(), &ho); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ho)
}

func (h *Handler) GetHandoff(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ho, err := h.svc.GetHandoff(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "handoff not found")
	}
	return c.JSON(http.StatusOK, ho)
}

func (h *Handler) ListHandoffs(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListHandoffsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if providerID := c.QueryParam("provider_id"); providerID != "" {
		prid, err := uuid.Parse(providerID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid provider_id")
		}
		items, total, err := h.svc.ListHandoffsByProvider(c.Request().Context(), prid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "patient_id or provider_id query parameter is required")
}

func (h *Handler) UpdateHandoff(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ho HandoffRecord
	if err := c.Bind(&ho); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ho.ID = id
	if err := h.svc.UpdateHandoff(c.Request().Context(), &ho); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ho)
}
