package documents

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	// Document template endpoints
	readGroup.GET("/document-templates", h.ListTemplates)
	readGroup.GET("/document-templates/:id", h.GetTemplate)
	writeGroup.POST("/document-templates", h.CreateTemplate)
	writeGroup.PUT("/document-templates/:id", h.UpdateTemplate)
	writeGroup.DELETE("/document-templates/:id", h.DeleteTemplate)
	writeGroup.POST("/document-templates/:id/render", h.RenderTemplate)

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
	fhirWrite.PUT("/Consent/:id", h.UpdateConsentFHIR)
	fhirWrite.DELETE("/Consent/:id", h.DeleteConsentFHIR)
	fhirWrite.PATCH("/Consent/:id", h.PatchConsentFHIR)
	fhirWrite.POST("/DocumentReference", h.CreateDocumentReferenceFHIR)
	fhirWrite.PUT("/DocumentReference/:id", h.UpdateDocumentReferenceFHIR)
	fhirWrite.DELETE("/DocumentReference/:id", h.DeleteDocumentReferenceFHIR)
	fhirWrite.PATCH("/DocumentReference/:id", h.PatchDocumentReferenceFHIR)
	fhirWrite.POST("/Composition", h.CreateCompositionFHIR)
	fhirWrite.PUT("/Composition/:id", h.UpdateCompositionFHIR)
	fhirWrite.DELETE("/Composition/:id", h.DeleteCompositionFHIR)
	fhirWrite.PATCH("/Composition/:id", h.PatchCompositionFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/Consent/_search", h.SearchConsentsFHIR)
	fhirRead.POST("/DocumentReference/_search", h.SearchDocumentReferencesFHIR)
	fhirRead.POST("/Composition/_search", h.SearchCompositionsFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Consent/:id/_history/:vid", h.VreadConsentFHIR)
	fhirRead.GET("/Consent/:id/_history", h.HistoryConsentFHIR)
	fhirRead.GET("/DocumentReference/:id/_history/:vid", h.VreadDocumentReferenceFHIR)
	fhirRead.GET("/DocumentReference/:id/_history", h.HistoryDocumentReferenceFHIR)
	fhirRead.GET("/Composition/:id/_history/:vid", h.VreadCompositionFHIR)
	fhirRead.GET("/Composition/:id/_history", h.HistoryCompositionFHIR)
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

// -- DocumentTemplate Handlers --

func (h *Handler) CreateTemplate(c echo.Context) error {
	var t DocumentTemplate
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTemplate(c.Request().Context(), &t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *Handler) GetTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	t, err := h.svc.GetTemplate(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handler) ListTemplates(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListTemplates(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var t DocumentTemplate
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	t.ID = id
	if err := h.svc.UpdateTemplate(c.Request().Context(), &t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handler) DeleteTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTemplate(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RenderTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var body struct {
		Variables map[string]string `json:"variables"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	rendered, err := h.svc.RenderTemplate(c.Request().Context(), id, body.Variables)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, rendered)
}

// -- FHIR Endpoints --

func (h *Handler) SearchConsentsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchConsents(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/Consent",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	}))
}

func (h *Handler) GetConsentFHIR(c echo.Context) error {
	consent, err := h.svc.GetConsentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, consent.VersionID, consent.UpdatedAt.Format("2006-01-02T15:04:05Z"))
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
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchDocumentReferences(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	bundle := fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/DocumentReference",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	})
	fhir.HandleProvenanceRevInclude(bundle, c, h.pool)
	return c.JSON(http.StatusOK, bundle)
}

func (h *Handler) GetDocumentReferenceFHIR(c echo.Context) error {
	doc, err := h.svc.GetDocumentReferenceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, doc.VersionID, doc.UpdatedAt.Format("2006-01-02T15:04:05Z"))
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
		return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
			ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
			BaseURL:  "/fhir/Composition",
			QueryStr: c.QueryString(),
			Count:    pg.Limit,
			Offset:   pg.Offset,
			Total:    total,
		}))
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundleWithLinks(nil, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/Composition",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    0,
	}))
}

func (h *Handler) GetCompositionFHIR(c echo.Context) error {
	comp, err := h.svc.GetCompositionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, comp.VersionID, comp.UpdatedAt.Format("2006-01-02T15:04:05Z"))
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

// -- FHIR Update Endpoints --

func (h *Handler) UpdateConsentFHIR(c echo.Context) error {
	var consent Consent
	if err := c.Bind(&consent); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetConsentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", c.Param("id")))
	}
	consent.ID = existing.ID
	consent.FHIRID = existing.FHIRID
	if err := h.svc.UpdateConsent(c.Request().Context(), &consent); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, consent.ToFHIR())
}

func (h *Handler) UpdateDocumentReferenceFHIR(c echo.Context) error {
	var doc DocumentReference
	if err := c.Bind(&doc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetDocumentReferenceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", c.Param("id")))
	}
	doc.ID = existing.ID
	doc.FHIRID = existing.FHIRID
	if err := h.svc.UpdateDocumentReference(c.Request().Context(), &doc); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, doc.ToFHIR())
}

func (h *Handler) UpdateCompositionFHIR(c echo.Context) error {
	var comp Composition
	if err := c.Bind(&comp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetCompositionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", c.Param("id")))
	}
	comp.ID = existing.ID
	comp.FHIRID = existing.FHIRID
	if err := h.svc.UpdateComposition(c.Request().Context(), &comp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, comp.ToFHIR())
}

// -- FHIR Delete Endpoints --

func (h *Handler) DeleteConsentFHIR(c echo.Context) error {
	existing, err := h.svc.GetConsentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", c.Param("id")))
	}
	if err := h.svc.DeleteConsent(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteDocumentReferenceFHIR(c echo.Context) error {
	existing, err := h.svc.GetDocumentReferenceByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", c.Param("id")))
	}
	if err := h.svc.DeleteDocumentReference(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteCompositionFHIR(c echo.Context) error {
	existing, err := h.svc.GetCompositionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", c.Param("id")))
	}
	if err := h.svc.DeleteComposition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR PATCH Endpoints --

func (h *Handler) PatchConsentFHIR(c echo.Context) error {
	return h.handlePatch(c, "Consent", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetConsentByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateConsent(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) PatchDocumentReferenceFHIR(c echo.Context) error {
	return h.handlePatch(c, "DocumentReference", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetDocumentReferenceByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateDocumentReference(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) PatchCompositionFHIR(c echo.Context) error {
	return h.handlePatch(c, "Composition", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetCompositionByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateComposition(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

// handlePatch dispatches to JSON Patch or Merge Patch based on Content-Type.
func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	// Get current resource as FHIR map
	var currentResource map[string]interface{}
	switch resourceType {
	case "Consent":
		existing, err := h.svc.GetConsentByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "DocumentReference":
		existing, err := h.svc.GetDocumentReferenceByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "Composition":
		existing, err := h.svc.GetCompositionByFHIRID(c.Request().Context(), fhirID)
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

// -- FHIR vread and history endpoints --

func (h *Handler) VreadConsentFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	fhirID := c.Param("id")
	vidStr := c.Param("vid")

	if vt := h.svc.VersionTracker(); vt != nil {
		var vid int
		if _, err := fmt.Sscanf(vidStr, "%d", &vid); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid version id"))
		}
		entry, err := vt.GetVersion(ctx, "Consent", fhirID, vid)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", fhirID+"/_history/"+vidStr))
		}
		var resource map[string]interface{}
		if err := json.Unmarshal(entry.Resource, &resource); err != nil {
			return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome("failed to parse versioned resource"))
		}
		fhir.SetVersionHeaders(c, entry.VersionID, entry.Timestamp.Format("2006-01-02T15:04:05Z"))
		return c.JSON(http.StatusOK, resource)
	}

	consent, err := h.svc.GetConsentByFHIRID(ctx, fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", fhirID))
	}
	result := consent.ToFHIR()
	fhir.SetVersionHeaders(c, consent.VersionID, consent.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryConsentFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	fhirID := c.Param("id")

	if vt := h.svc.VersionTracker(); vt != nil {
		entries, total, err := vt.ListVersions(ctx, "Consent", fhirID, 100, 0)
		if err != nil || total == 0 {
			consent, ferr := h.svc.GetConsentByFHIRID(ctx, fhirID)
			if ferr != nil {
				return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", fhirID))
			}
			result := consent.ToFHIR()
			raw, _ := json.Marshal(result)
			entry := &fhir.HistoryEntry{
				ResourceType: "Consent", ResourceID: consent.FHIRID, VersionID: consent.VersionID,
				Resource: raw, Action: "create", Timestamp: consent.CreatedAt,
			}
			return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
		}
		return c.JSON(http.StatusOK, fhir.NewHistoryBundle(entries, total, "/fhir"))
	}

	consent, err := h.svc.GetConsentByFHIRID(ctx, fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Consent", fhirID))
	}
	result := consent.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Consent", ResourceID: consent.FHIRID, VersionID: consent.VersionID,
		Resource: raw, Action: "create", Timestamp: consent.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadDocumentReferenceFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	fhirID := c.Param("id")
	vidStr := c.Param("vid")

	if vt := h.svc.VersionTracker(); vt != nil {
		var vid int
		if _, err := fmt.Sscanf(vidStr, "%d", &vid); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid version id"))
		}
		entry, err := vt.GetVersion(ctx, "DocumentReference", fhirID, vid)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", fhirID+"/_history/"+vidStr))
		}
		var resource map[string]interface{}
		if err := json.Unmarshal(entry.Resource, &resource); err != nil {
			return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome("failed to parse versioned resource"))
		}
		fhir.SetVersionHeaders(c, entry.VersionID, entry.Timestamp.Format("2006-01-02T15:04:05Z"))
		return c.JSON(http.StatusOK, resource)
	}

	doc, err := h.svc.GetDocumentReferenceByFHIRID(ctx, fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", fhirID))
	}
	result := doc.ToFHIR()
	fhir.SetVersionHeaders(c, doc.VersionID, doc.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryDocumentReferenceFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	fhirID := c.Param("id")

	if vt := h.svc.VersionTracker(); vt != nil {
		entries, total, err := vt.ListVersions(ctx, "DocumentReference", fhirID, 100, 0)
		if err != nil || total == 0 {
			doc, ferr := h.svc.GetDocumentReferenceByFHIRID(ctx, fhirID)
			if ferr != nil {
				return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", fhirID))
			}
			result := doc.ToFHIR()
			raw, _ := json.Marshal(result)
			entry := &fhir.HistoryEntry{
				ResourceType: "DocumentReference", ResourceID: doc.FHIRID, VersionID: doc.VersionID,
				Resource: raw, Action: "create", Timestamp: doc.CreatedAt,
			}
			return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
		}
		return c.JSON(http.StatusOK, fhir.NewHistoryBundle(entries, total, "/fhir"))
	}

	doc, err := h.svc.GetDocumentReferenceByFHIRID(ctx, fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DocumentReference", fhirID))
	}
	result := doc.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "DocumentReference", ResourceID: doc.FHIRID, VersionID: doc.VersionID,
		Resource: raw, Action: "create", Timestamp: doc.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadCompositionFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	fhirID := c.Param("id")
	vidStr := c.Param("vid")

	if vt := h.svc.VersionTracker(); vt != nil {
		var vid int
		if _, err := fmt.Sscanf(vidStr, "%d", &vid); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid version id"))
		}
		entry, err := vt.GetVersion(ctx, "Composition", fhirID, vid)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", fhirID+"/_history/"+vidStr))
		}
		var resource map[string]interface{}
		if err := json.Unmarshal(entry.Resource, &resource); err != nil {
			return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome("failed to parse versioned resource"))
		}
		fhir.SetVersionHeaders(c, entry.VersionID, entry.Timestamp.Format("2006-01-02T15:04:05Z"))
		return c.JSON(http.StatusOK, resource)
	}

	comp, err := h.svc.GetCompositionByFHIRID(ctx, fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", fhirID))
	}
	result := comp.ToFHIR()
	fhir.SetVersionHeaders(c, comp.VersionID, comp.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryCompositionFHIR(c echo.Context) error {
	ctx := c.Request().Context()
	fhirID := c.Param("id")

	if vt := h.svc.VersionTracker(); vt != nil {
		entries, total, err := vt.ListVersions(ctx, "Composition", fhirID, 100, 0)
		if err != nil || total == 0 {
			comp, ferr := h.svc.GetCompositionByFHIRID(ctx, fhirID)
			if ferr != nil {
				return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", fhirID))
			}
			result := comp.ToFHIR()
			raw, _ := json.Marshal(result)
			entry := &fhir.HistoryEntry{
				ResourceType: "Composition", ResourceID: comp.FHIRID, VersionID: comp.VersionID,
				Resource: raw, Action: "create", Timestamp: comp.CreatedAt,
			}
			return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
		}
		return c.JSON(http.StatusOK, fhir.NewHistoryBundle(entries, total, "/fhir"))
	}

	comp, err := h.svc.GetCompositionByFHIRID(ctx, fhirID)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Composition", fhirID))
	}
	result := comp.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Composition", ResourceID: comp.FHIRID, VersionID: comp.VersionID,
		Resource: raw, Action: "create", Timestamp: comp.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
