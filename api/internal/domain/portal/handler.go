package portal

import (
	"net/http"

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
	// Read endpoints – admin, physician, nurse, patient
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "patient"))
	readGroup.GET("/portal-accounts", h.ListPortalAccounts)
	readGroup.GET("/portal-accounts/:id", h.GetPortalAccount)
	readGroup.GET("/portal-messages", h.ListPortalMessages)
	readGroup.GET("/portal-messages/:id", h.GetPortalMessage)
	readGroup.GET("/questionnaires", h.ListQuestionnaires)
	readGroup.GET("/questionnaires/:id", h.GetQuestionnaire)
	readGroup.GET("/questionnaires/:id/items", h.GetQuestionnaireItems)
	readGroup.GET("/questionnaire-responses", h.ListQuestionnaireResponses)
	readGroup.GET("/questionnaire-responses/:id", h.GetQuestionnaireResponse)
	readGroup.GET("/questionnaire-responses/:id/items", h.GetQuestionnaireResponseItems)
	readGroup.GET("/patient-checkins", h.ListPatientCheckins)
	readGroup.GET("/patient-checkins/:id", h.GetPatientCheckin)

	// Write endpoints – admin, physician, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	writeGroup.POST("/portal-accounts", h.CreatePortalAccount)
	writeGroup.PUT("/portal-accounts/:id", h.UpdatePortalAccount)
	writeGroup.DELETE("/portal-accounts/:id", h.DeletePortalAccount)
	writeGroup.POST("/portal-messages", h.CreatePortalMessage)
	writeGroup.PUT("/portal-messages/:id", h.UpdatePortalMessage)
	writeGroup.DELETE("/portal-messages/:id", h.DeletePortalMessage)
	writeGroup.POST("/questionnaires", h.CreateQuestionnaire)
	writeGroup.PUT("/questionnaires/:id", h.UpdateQuestionnaire)
	writeGroup.DELETE("/questionnaires/:id", h.DeleteQuestionnaire)
	writeGroup.POST("/questionnaires/:id/items", h.AddQuestionnaireItem)
	writeGroup.POST("/questionnaire-responses", h.CreateQuestionnaireResponse)
	writeGroup.PUT("/questionnaire-responses/:id", h.UpdateQuestionnaireResponse)
	writeGroup.DELETE("/questionnaire-responses/:id", h.DeleteQuestionnaireResponse)
	writeGroup.POST("/questionnaire-responses/:id/items", h.AddQuestionnaireResponseItem)
	writeGroup.POST("/patient-checkins", h.CreatePatientCheckin)
	writeGroup.PUT("/patient-checkins/:id", h.UpdatePatientCheckin)
	writeGroup.DELETE("/patient-checkins/:id", h.DeletePatientCheckin)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "patient"))
	fhirRead.GET("/Questionnaire", h.SearchQuestionnairesFHIR)
	fhirRead.GET("/Questionnaire/:id", h.GetQuestionnaireFHIR)
	fhirRead.GET("/QuestionnaireResponse", h.SearchQuestionnaireResponsesFHIR)
	fhirRead.GET("/QuestionnaireResponse/:id", h.GetQuestionnaireResponseFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fhirWrite.POST("/Questionnaire", h.CreateQuestionnaireFHIR)
	fhirWrite.POST("/QuestionnaireResponse", h.CreateQuestionnaireResponseFHIR)
}

// -- Portal Account Handlers --

func (h *Handler) CreatePortalAccount(c echo.Context) error {
	var a PortalAccount
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePortalAccount(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetPortalAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetPortalAccount(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "portal account not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListPortalAccounts(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListPortalAccountsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListPortalAccounts(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePortalAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a PortalAccount
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdatePortalAccount(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeletePortalAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePortalAccount(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Portal Message Handlers --

func (h *Handler) CreatePortalMessage(c echo.Context) error {
	var m PortalMessage
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePortalMessage(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetPortalMessage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetPortalMessage(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "portal message not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListPortalMessages(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListPortalMessagesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListPortalMessages(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePortalMessage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m PortalMessage
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdatePortalMessage(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeletePortalMessage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePortalMessage(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Questionnaire Handlers --

func (h *Handler) CreateQuestionnaire(c echo.Context) error {
	var q Questionnaire
	if err := c.Bind(&q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateQuestionnaire(c.Request().Context(), &q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, q)
}

func (h *Handler) GetQuestionnaire(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	q, err := h.svc.GetQuestionnaire(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "questionnaire not found")
	}
	return c.JSON(http.StatusOK, q)
}

func (h *Handler) ListQuestionnaires(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.SearchQuestionnaires(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateQuestionnaire(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var q Questionnaire
	if err := c.Bind(&q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	q.ID = id
	if err := h.svc.UpdateQuestionnaire(c.Request().Context(), &q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, q)
}

func (h *Handler) DeleteQuestionnaire(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteQuestionnaire(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddQuestionnaireItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var item QuestionnaireItem
	if err := c.Bind(&item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	item.QuestionnaireID = id
	if err := h.svc.AddQuestionnaireItem(c.Request().Context(), &item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetQuestionnaireItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetQuestionnaireItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Questionnaire Response Handlers --

func (h *Handler) CreateQuestionnaireResponse(c echo.Context) error {
	var qr QuestionnaireResponse
	if err := c.Bind(&qr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateQuestionnaireResponse(c.Request().Context(), &qr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, qr)
}

func (h *Handler) GetQuestionnaireResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	qr, err := h.svc.GetQuestionnaireResponse(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "questionnaire response not found")
	}
	return c.JSON(http.StatusOK, qr)
}

func (h *Handler) ListQuestionnaireResponses(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListQuestionnaireResponsesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListQuestionnaireResponses(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateQuestionnaireResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var qr QuestionnaireResponse
	if err := c.Bind(&qr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	qr.ID = id
	if err := h.svc.UpdateQuestionnaireResponse(c.Request().Context(), &qr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, qr)
}

func (h *Handler) DeleteQuestionnaireResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteQuestionnaireResponse(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddQuestionnaireResponseItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var item QuestionnaireResponseItem
	if err := c.Bind(&item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	item.ResponseID = id
	if err := h.svc.AddQuestionnaireResponseItem(c.Request().Context(), &item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetQuestionnaireResponseItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetQuestionnaireResponseItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Patient Checkin Handlers --

func (h *Handler) CreatePatientCheckin(c echo.Context) error {
	var ch PatientCheckin
	if err := c.Bind(&ch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePatientCheckin(c.Request().Context(), &ch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ch)
}

func (h *Handler) GetPatientCheckin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ch, err := h.svc.GetPatientCheckin(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "patient checkin not found")
	}
	return c.JSON(http.StatusOK, ch)
}

func (h *Handler) ListPatientCheckins(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListPatientCheckinsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListPatientCheckins(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePatientCheckin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ch PatientCheckin
	if err := c.Bind(&ch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ch.ID = id
	if err := h.svc.UpdatePatientCheckin(c.Request().Context(), &ch); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ch)
}

func (h *Handler) DeletePatientCheckin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePatientCheckin(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchQuestionnairesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"name", "status", "publisher"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchQuestionnaires(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Questionnaire"))
}

func (h *Handler) GetQuestionnaireFHIR(c echo.Context) error {
	q, err := h.svc.GetQuestionnaireByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Questionnaire", c.Param("id")))
	}
	return c.JSON(http.StatusOK, q.ToFHIR())
}

func (h *Handler) CreateQuestionnaireFHIR(c echo.Context) error {
	var q Questionnaire
	if err := c.Bind(&q); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateQuestionnaire(c.Request().Context(), &q); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Questionnaire/"+q.FHIRID)
	return c.JSON(http.StatusCreated, q.ToFHIR())
}

func (h *Handler) SearchQuestionnaireResponsesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "questionnaire", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchQuestionnaireResponses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/QuestionnaireResponse"))
}

func (h *Handler) GetQuestionnaireResponseFHIR(c echo.Context) error {
	qr, err := h.svc.GetQuestionnaireResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("QuestionnaireResponse", c.Param("id")))
	}
	return c.JSON(http.StatusOK, qr.ToFHIR())
}

func (h *Handler) CreateQuestionnaireResponseFHIR(c echo.Context) error {
	var qr QuestionnaireResponse
	if err := c.Bind(&qr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateQuestionnaireResponse(c.Request().Context(), &qr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/QuestionnaireResponse/"+qr.FHIRID)
	return c.JSON(http.StatusCreated, qr.ToFHIR())
}
