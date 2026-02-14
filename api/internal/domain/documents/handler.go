package documents

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
	// Read endpoints – admin, physician, nurse
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	readGroup.GET("/consents", h.ListConsents)
	readGroup.GET("/consents/:id", h.GetConsent)
	readGroup.GET("/document-references", h.ListDocumentReferences)
	readGroup.GET("/document-references/:id", h.GetDocumentReference)
	readGroup.GET("/clinical-notes", h.ListClinicalNotes)
	readGroup.GET("/clinical-notes/:id", h.GetClinicalNote)
	readGroup.GET("/compositions", h.ListCompositions)
	readGroup.GET("/compositions/:id", h.GetComposition)
	readGroup.GET("/compositions/:id/sections", h.GetSections)

	// Write endpoints – admin, physician, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	writeGroup.POST("/consents", h.CreateConsent)
	writeGroup.PUT("/consents/:id", h.UpdateConsent)
	writeGroup.DELETE("/consents/:id", h.DeleteConsent)
	writeGroup.POST("/document-references", h.CreateDocumentReference)
	writeGroup.PUT("/document-references/:id", h.UpdateDocumentReference)
	writeGroup.DELETE("/document-references/:id", h.DeleteDocumentReference)
	writeGroup.POST("/clinical-notes", h.CreateClinicalNote)
	writeGroup.PUT("/clinical-notes/:id", h.UpdateClinicalNote)
	writeGroup.DELETE("/clinical-notes/:id", h.DeleteClinicalNote)
	writeGroup.POST("/compositions", h.CreateComposition)
	writeGroup.PUT("/compositions/:id", h.UpdateComposition)
	writeGroup.DELETE("/compositions/:id", h.DeleteComposition)
	writeGroup.POST("/compositions/:id/sections", h.AddSection)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fhirRead.GET("/Consent", h.SearchConsentsFHIR)
	fhirRead.GET("/Consent/:id", h.GetConsentFHIR)
	fhirRead.GET("/DocumentReference", h.SearchDocumentReferencesFHIR)
	fhirRead.GET("/DocumentReference/:id", h.GetDocumentReferenceFHIR)
	fhirRead.GET("/Composition", h.SearchCompositionsFHIR)
	fhirRead.GET("/Composition/:id", h.GetCompositionFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fhirWrite.POST("/Consent", h.CreateConsentFHIR)
	fhirWrite.POST("/DocumentReference", h.CreateDocumentReferenceFHIR)
	fhirWrite.POST("/Composition", h.CreateCompositionFHIR)
}

// -- Consent Handlers --

func (h *Handler) CreateConsent(c echo.Context) error {
	var consent Consent
	if err := c.Bind(&consent); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateConsent(c.Request().Context(), &consent); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, consent)
}

func (h *Handler) GetConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	consent, err := h.svc.GetConsent(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "consent not found")
	}
	return c.JSON(http.StatusOK, consent)
}

func (h *Handler) ListConsents(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListConsentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchConsents(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var consent Consent
	if err := c.Bind(&consent); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	consent.ID = id
	if err := h.svc.UpdateConsent(c.Request().Context(), &consent); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, consent)
}

func (h *Handler) DeleteConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteConsent(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- DocumentReference Handlers --

func (h *Handler) CreateDocumentReference(c echo.Context) error {
	var doc DocumentReference
	if err := c.Bind(&doc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDocumentReference(c.Request().Context(), &doc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, doc)
}

func (h *Handler) GetDocumentReference(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	doc, err := h.svc.GetDocumentReference(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "document reference not found")
	}
	return c.JSON(http.StatusOK, doc)
}

func (h *Handler) ListDocumentReferences(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListDocumentReferencesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchDocumentReferences(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDocumentReference(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var doc DocumentReference
	if err := c.Bind(&doc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	doc.ID = id
	if err := h.svc.UpdateDocumentReference(c.Request().Context(), &doc); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, doc)
}

func (h *Handler) DeleteDocumentReference(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDocumentReference(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- ClinicalNote Handlers --

func (h *Handler) CreateClinicalNote(c echo.Context) error {
	var note ClinicalNote
	if err := c.Bind(&note); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateClinicalNote(c.Request().Context(), &note); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, note)
}

func (h *Handler) GetClinicalNote(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	note, err := h.svc.GetClinicalNote(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "clinical note not found")
	}
	return c.JSON(http.StatusOK, note)
}

func (h *Handler) ListClinicalNotes(c echo.Context) error {
	pg := pagination.FromContext(c)
	if encounterID := c.QueryParam("encounter_id"); encounterID != "" {
		eid, err := uuid.Parse(encounterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid encounter_id")
		}
		items, total, err := h.svc.ListClinicalNotesByEncounter(c.Request().Context(), eid, pg.Limit, pg.Offset)
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
		items, total, err := h.svc.ListClinicalNotesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "patient_id or encounter_id is required")
}

func (h *Handler) UpdateClinicalNote(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var note ClinicalNote
	if err := c.Bind(&note); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	note.ID = id
	if err := h.svc.UpdateClinicalNote(c.Request().Context(), &note); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, note)
}

func (h *Handler) DeleteClinicalNote(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteClinicalNote(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Composition Handlers --

func (h *Handler) CreateComposition(c echo.Context) error {
	var comp Composition
	if err := c.Bind(&comp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateComposition(c.Request().Context(), &comp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, comp)
}

func (h *Handler) GetComposition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	comp, err := h.svc.GetComposition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "composition not found")
	}
	return c.JSON(http.StatusOK, comp)
}

func (h *Handler) ListCompositions(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListCompositionsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
}

func (h *Handler) UpdateComposition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var comp Composition
	if err := c.Bind(&comp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	comp.ID = id
	if err := h.svc.UpdateComposition(c.Request().Context(), &comp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, comp)
}

func (h *Handler) DeleteComposition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteComposition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddSection(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sec CompositionSection
	if err := c.Bind(&sec); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sec.CompositionID = id
	if err := h.svc.AddCompositionSection(c.Request().Context(), &sec); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sec)
}

func (h *Handler) GetSections(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sections, err := h.svc.GetCompositionSections(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sections)
}

// -- FHIR Endpoints --

func (h *Handler) SearchConsentsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "category"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchConsents(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Consent"))
}

func (h *Handler) GetConsentFHIR(c echo.Context) error {
	consent, err := h.svc.GetConsentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", c.Param("id")))
	}
	return c.JSON(http.StatusOK, consent.ToFHIR())
}

func (h *Handler) CreateConsentFHIR(c echo.Context) error {
	var consent Consent
	if err := c.Bind(&consent); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateConsent(c.Request().Context(), &consent); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Consent/"+consent.FHIRID)
	return c.JSON(http.StatusCreated, consent.ToFHIR())
}

func (h *Handler) SearchDocumentReferencesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "type", "category"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchDocumentReferences(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/DocumentReference"))
}

func (h *Handler) GetDocumentReferenceFHIR(c echo.Context) error {
	doc, err := h.svc.GetDocumentReferenceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", c.Param("id")))
	}
	return c.JSON(http.StatusOK, doc.ToFHIR())
}

func (h *Handler) CreateDocumentReferenceFHIR(c echo.Context) error {
	var doc DocumentReference
	if err := c.Bind(&doc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateDocumentReference(c.Request().Context(), &doc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/DocumentReference/"+doc.FHIRID)
	return c.JSON(http.StatusCreated, doc.ToFHIR())
}

func (h *Handler) SearchCompositionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid patient parameter"))
		}
		items, total, err := h.svc.ListCompositionsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
		}
		resources := make([]interface{}, len(items))
		for i, item := range items {
			resources[i] = item.ToFHIR()
		}
		return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Composition"))
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(nil, 0, "/fhir/Composition"))
}

func (h *Handler) GetCompositionFHIR(c echo.Context) error {
	comp, err := h.svc.GetCompositionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, comp.ToFHIR())
}

func (h *Handler) CreateCompositionFHIR(c echo.Context) error {
	var comp Composition
	if err := c.Bind(&comp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateComposition(c.Request().Context(), &comp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Composition/"+comp.FHIRID)
	return c.JSON(http.StatusCreated, comp.ToFHIR())
}
