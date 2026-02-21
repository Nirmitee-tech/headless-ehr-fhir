package clinical

import (
	"encoding/json"
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
	readGroup.GET("/conditions", h.ListConditions)
	readGroup.GET("/conditions/:id", h.GetCondition)
	readGroup.GET("/observations", h.ListObservations)
	readGroup.GET("/observations/:id", h.GetObservation)
	readGroup.GET("/observations/:id/components", h.GetObservationComponents)
	readGroup.GET("/allergies", h.ListAllergies)
	readGroup.GET("/allergies/:id", h.GetAllergy)
	readGroup.GET("/allergies/:id/reactions", h.GetReactions)
	readGroup.GET("/procedures", h.ListProcedures)
	readGroup.GET("/procedures/:id", h.GetProcedure)
	readGroup.GET("/procedures/:id/performers", h.GetPerformers)

	// Write endpoints – admin, physician, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	writeGroup.POST("/conditions", h.CreateCondition)
	writeGroup.PUT("/conditions/:id", h.UpdateCondition)
	writeGroup.DELETE("/conditions/:id", h.DeleteCondition)
	writeGroup.POST("/observations", h.CreateObservation)
	writeGroup.PUT("/observations/:id", h.UpdateObservation)
	writeGroup.DELETE("/observations/:id", h.DeleteObservation)
	writeGroup.POST("/observations/:id/components", h.AddObservationComponent)
	writeGroup.POST("/allergies", h.CreateAllergy)
	writeGroup.PUT("/allergies/:id", h.UpdateAllergy)
	writeGroup.DELETE("/allergies/:id", h.DeleteAllergy)
	writeGroup.POST("/allergies/:id/reactions", h.AddReaction)
	writeGroup.POST("/procedures", h.CreateProcedure)
	writeGroup.PUT("/procedures/:id", h.UpdateProcedure)
	writeGroup.DELETE("/procedures/:id", h.DeleteProcedure)
	writeGroup.POST("/procedures/:id/performers", h.AddPerformer)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fhirRead.GET("/Condition", h.SearchConditionsFHIR)
	fhirRead.GET("/Condition/:id", h.GetConditionFHIR)
	fhirRead.GET("/Observation", h.SearchObservationsFHIR)
	fhirRead.GET("/Observation/:id", h.GetObservationFHIR)
	fhirRead.GET("/AllergyIntolerance", h.SearchAllergiesFHIR)
	fhirRead.GET("/AllergyIntolerance/:id", h.GetAllergyFHIR)
	fhirRead.GET("/Procedure", h.SearchProceduresFHIR)
	fhirRead.GET("/Procedure/:id", h.GetProcedureFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fhirWrite.POST("/Condition", h.CreateConditionFHIR)
	fhirWrite.PUT("/Condition/:id", h.UpdateConditionFHIR)
	fhirWrite.DELETE("/Condition/:id", h.DeleteConditionFHIR)
	fhirWrite.PATCH("/Condition/:id", h.PatchConditionFHIR)
	fhirWrite.POST("/Observation", h.CreateObservationFHIR)
	fhirWrite.PUT("/Observation/:id", h.UpdateObservationFHIR)
	fhirWrite.DELETE("/Observation/:id", h.DeleteObservationFHIR)
	fhirWrite.PATCH("/Observation/:id", h.PatchObservationFHIR)
	fhirWrite.POST("/AllergyIntolerance", h.CreateAllergyFHIR)
	fhirWrite.PUT("/AllergyIntolerance/:id", h.UpdateAllergyFHIR)
	fhirWrite.DELETE("/AllergyIntolerance/:id", h.DeleteAllergyFHIR)
	fhirWrite.PATCH("/AllergyIntolerance/:id", h.PatchAllergyFHIR)
	fhirWrite.POST("/Procedure", h.CreateProcedureFHIR)
	fhirWrite.PUT("/Procedure/:id", h.UpdateProcedureFHIR)
	fhirWrite.DELETE("/Procedure/:id", h.DeleteProcedureFHIR)
	fhirWrite.PATCH("/Procedure/:id", h.PatchProcedureFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/Condition/_search", h.SearchConditionsFHIR)
	fhirRead.POST("/Observation/_search", h.SearchObservationsFHIR)
	fhirRead.POST("/AllergyIntolerance/_search", h.SearchAllergiesFHIR)
	fhirRead.POST("/Procedure/_search", h.SearchProceduresFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Condition/:id/_history/:vid", h.VreadConditionFHIR)
	fhirRead.GET("/Condition/:id/_history", h.HistoryConditionFHIR)
	fhirRead.GET("/Observation/:id/_history/:vid", h.VreadObservationFHIR)
	fhirRead.GET("/Observation/:id/_history", h.HistoryObservationFHIR)
	fhirRead.GET("/AllergyIntolerance/:id/_history/:vid", h.VreadAllergyFHIR)
	fhirRead.GET("/AllergyIntolerance/:id/_history", h.HistoryAllergyFHIR)
	fhirRead.GET("/Procedure/:id/_history/:vid", h.VreadProcedureFHIR)
	fhirRead.GET("/Procedure/:id/_history", h.HistoryProcedureFHIR)
}

// -- Condition Handlers --

func (h *Handler) CreateCondition(c echo.Context) error {
	var cond Condition
	if err := c.Bind(&cond); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCondition(c.Request().Context(), &cond); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cond)
}

func (h *Handler) GetCondition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cond, err := h.svc.GetCondition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "condition not found")
	}
	return c.JSON(http.StatusOK, cond)
}

func (h *Handler) ListConditions(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListConditionsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchConditions(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCondition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cond Condition
	if err := c.Bind(&cond); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cond.ID = id
	if err := h.svc.UpdateCondition(c.Request().Context(), &cond); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cond)
}

func (h *Handler) DeleteCondition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCondition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Observation Handlers --

func (h *Handler) CreateObservation(c echo.Context) error {
	var obs Observation
	if err := c.Bind(&obs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateObservation(c.Request().Context(), &obs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, obs)
}

func (h *Handler) GetObservation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	obs, err := h.svc.GetObservation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "observation not found")
	}
	return c.JSON(http.StatusOK, obs)
}

func (h *Handler) ListObservations(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListObservationsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchObservations(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateObservation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var obs Observation
	if err := c.Bind(&obs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	obs.ID = id
	if err := h.svc.UpdateObservation(c.Request().Context(), &obs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, obs)
}

func (h *Handler) DeleteObservation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteObservation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddObservationComponent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var comp ObservationComponent
	if err := c.Bind(&comp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	comp.ObservationID = id
	if err := h.svc.AddObservationComponent(c.Request().Context(), &comp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, comp)
}

func (h *Handler) GetObservationComponents(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	comps, err := h.svc.GetObservationComponents(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, comps)
}

// -- Allergy Handlers --

func (h *Handler) CreateAllergy(c echo.Context) error {
	var a AllergyIntolerance
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateAllergy(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetAllergy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetAllergy(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "allergy not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListAllergies(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListAllergiesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchAllergies(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateAllergy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a AllergyIntolerance
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateAllergy(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeleteAllergy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteAllergy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddReaction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r AllergyReaction
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.AllergyID = id
	if err := h.svc.AddAllergyReaction(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetReactions(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	reactions, err := h.svc.GetAllergyReactions(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, reactions)
}

// -- Procedure Handlers --

func (h *Handler) CreateProcedure(c echo.Context) error {
	var p ProcedureRecord
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateProcedure(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetProcedure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetProcedure(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "procedure not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListProcedures(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListProceduresByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchProcedures(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateProcedure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p ProcedureRecord
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdateProcedure(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteProcedure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteProcedure(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPerformer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var pf ProcedurePerformer
	if err := c.Bind(&pf); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	pf.ProcedureID = id
	if err := h.svc.AddProcedurePerformer(c.Request().Context(), &pf); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, pf)
}

func (h *Handler) GetPerformers(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	performers, err := h.svc.GetProcedurePerformers(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, performers)
}

// -- FHIR Endpoints --

func (h *Handler) SearchConditionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchConditions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	bundle := fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/Condition",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	})
	fhir.HandleProvenanceRevInclude(bundle, c, h.pool)
	return c.JSON(http.StatusOK, bundle)
}

func (h *Handler) GetConditionFHIR(c echo.Context) error {
	cond, err := h.svc.GetConditionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Condition", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, cond.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, cond.ToFHIR())
}

func (h *Handler) CreateConditionFHIR(c echo.Context) error {
	var cond Condition
	if err := c.Bind(&cond); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCondition(c.Request().Context(), &cond); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Condition/"+cond.FHIRID)
	return c.JSON(http.StatusCreated, cond.ToFHIR())
}

func (h *Handler) SearchObservationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchObservations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	bundle := fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/Observation",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	})
	fhir.HandleProvenanceRevInclude(bundle, c, h.pool)
	return c.JSON(http.StatusOK, bundle)
}

func (h *Handler) GetObservationFHIR(c echo.Context) error {
	obs, err := h.svc.GetObservationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Observation", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, obs.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, obs.ToFHIR())
}

func (h *Handler) CreateObservationFHIR(c echo.Context) error {
	var obs Observation
	if err := c.Bind(&obs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateObservation(c.Request().Context(), &obs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Observation/"+obs.FHIRID)
	return c.JSON(http.StatusCreated, obs.ToFHIR())
}

func (h *Handler) SearchAllergiesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchAllergies(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	bundle := fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/AllergyIntolerance",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	})
	fhir.HandleProvenanceRevInclude(bundle, c, h.pool)
	return c.JSON(http.StatusOK, bundle)
}

func (h *Handler) GetAllergyFHIR(c echo.Context) error {
	a, err := h.svc.GetAllergyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AllergyIntolerance", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, a.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) CreateAllergyFHIR(c echo.Context) error {
	var a AllergyIntolerance
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateAllergy(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/AllergyIntolerance/"+a.FHIRID)
	return c.JSON(http.StatusCreated, a.ToFHIR())
}

func (h *Handler) SearchProceduresFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := fhir.ExtractSearchParams(c)
	items, total, err := h.svc.SearchProcedures(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	bundle := fhir.NewSearchBundleWithLinks(resources, fhir.SearchBundleParams{
		ServerBaseURL: fhir.ServerBaseURLFromRequest(c),
		BaseURL:  "/fhir/Procedure",
		QueryStr: c.QueryString(),
		Count:    pg.Limit,
		Offset:   pg.Offset,
		Total:    total,
	})
	fhir.HandleProvenanceRevInclude(bundle, c, h.pool)
	return c.JSON(http.StatusOK, bundle)
}

func (h *Handler) GetProcedureFHIR(c echo.Context) error {
	p, err := h.svc.GetProcedureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Procedure", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, p.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) CreateProcedureFHIR(c echo.Context) error {
	var p ProcedureRecord
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateProcedure(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Procedure/"+p.FHIRID)
	return c.JSON(http.StatusCreated, p.ToFHIR())
}

// -- FHIR Update Endpoints --

func (h *Handler) UpdateConditionFHIR(c echo.Context) error {
	var cond Condition
	if err := c.Bind(&cond); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetConditionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Condition", c.Param("id")))
	}
	cond.ID = existing.ID
	cond.FHIRID = existing.FHIRID
	if err := h.svc.UpdateCondition(c.Request().Context(), &cond); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cond.ToFHIR())
}

func (h *Handler) UpdateObservationFHIR(c echo.Context) error {
	var obs Observation
	if err := c.Bind(&obs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetObservationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Observation", c.Param("id")))
	}
	obs.ID = existing.ID
	obs.FHIRID = existing.FHIRID
	if err := h.svc.UpdateObservation(c.Request().Context(), &obs); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, obs.ToFHIR())
}

func (h *Handler) UpdateAllergyFHIR(c echo.Context) error {
	var a AllergyIntolerance
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetAllergyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AllergyIntolerance", c.Param("id")))
	}
	a.ID = existing.ID
	a.FHIRID = existing.FHIRID
	if err := h.svc.UpdateAllergy(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) UpdateProcedureFHIR(c echo.Context) error {
	var p ProcedureRecord
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetProcedureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Procedure", c.Param("id")))
	}
	p.ID = existing.ID
	p.FHIRID = existing.FHIRID
	if err := h.svc.UpdateProcedure(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

// -- FHIR Delete Endpoints --

func (h *Handler) DeleteConditionFHIR(c echo.Context) error {
	existing, err := h.svc.GetConditionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Condition", c.Param("id")))
	}
	if err := h.svc.DeleteCondition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteObservationFHIR(c echo.Context) error {
	existing, err := h.svc.GetObservationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Observation", c.Param("id")))
	}
	if err := h.svc.DeleteObservation(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteAllergyFHIR(c echo.Context) error {
	existing, err := h.svc.GetAllergyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AllergyIntolerance", c.Param("id")))
	}
	if err := h.svc.DeleteAllergy(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteProcedureFHIR(c echo.Context) error {
	existing, err := h.svc.GetProcedureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Procedure", c.Param("id")))
	}
	if err := h.svc.DeleteProcedure(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR PATCH Endpoints --

func (h *Handler) PatchConditionFHIR(c echo.Context) error {
	return h.handlePatch(c, "Condition", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetConditionByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Condition", ctx.Param("id")))
		}
		// Apply patch result back to the model (simplified: update key fields)
		if v, ok := resource["clinicalStatus"].(map[string]interface{}); ok {
			if coding, ok := v["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if code, ok := c["code"].(string); ok {
						existing.ClinicalStatus = code
					}
				}
			}
		}
		if v, ok := resource["note"].([]interface{}); ok && len(v) > 0 {
			if n, ok := v[0].(map[string]interface{}); ok {
				if text, ok := n["text"].(string); ok {
					existing.Note = &text
				}
			}
		}
		if err := h.svc.UpdateCondition(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) PatchObservationFHIR(c echo.Context) error {
	return h.handlePatch(c, "Observation", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetObservationByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Observation", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateObservation(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) PatchAllergyFHIR(c echo.Context) error {
	return h.handlePatch(c, "AllergyIntolerance", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetAllergyByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AllergyIntolerance", ctx.Param("id")))
		}
		if err := h.svc.UpdateAllergy(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) PatchProcedureFHIR(c echo.Context) error {
	return h.handlePatch(c, "Procedure", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetProcedureByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Procedure", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateProcedure(ctx.Request().Context(), existing); err != nil {
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
	case "Condition":
		existing, err := h.svc.GetConditionByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "Observation":
		existing, err := h.svc.GetObservationByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "AllergyIntolerance":
		existing, err := h.svc.GetAllergyByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "Procedure":
		existing, err := h.svc.GetProcedureByFHIRID(c.Request().Context(), fhirID)
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

func (h *Handler) VreadConditionFHIR(c echo.Context) error {
	// vread returns the current version (simplified - full impl uses history table)
	cond, err := h.svc.GetConditionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Condition", c.Param("id")))
	}
	result := cond.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cond.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryConditionFHIR(c echo.Context) error {
	cond, err := h.svc.GetConditionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Condition", c.Param("id")))
	}
	result := cond.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Condition", ResourceID: cond.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cond.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadObservationFHIR(c echo.Context) error {
	obs, err := h.svc.GetObservationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Observation", c.Param("id")))
	}
	result := obs.ToFHIR()
	fhir.SetVersionHeaders(c, 1, obs.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryObservationFHIR(c echo.Context) error {
	obs, err := h.svc.GetObservationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Observation", c.Param("id")))
	}
	result := obs.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Observation", ResourceID: obs.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: obs.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadAllergyFHIR(c echo.Context) error {
	a, err := h.svc.GetAllergyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AllergyIntolerance", c.Param("id")))
	}
	result := a.ToFHIR()
	fhir.SetVersionHeaders(c, 1, a.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryAllergyFHIR(c echo.Context) error {
	a, err := h.svc.GetAllergyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("AllergyIntolerance", c.Param("id")))
	}
	result := a.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "AllergyIntolerance", ResourceID: a.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: a.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadProcedureFHIR(c echo.Context) error {
	p, err := h.svc.GetProcedureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Procedure", c.Param("id")))
	}
	result := p.ToFHIR()
	fhir.SetVersionHeaders(c, 1, p.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryProcedureFHIR(c echo.Context) error {
	p, err := h.svc.GetProcedureByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Procedure", c.Param("id")))
	}
	result := p.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Procedure", ResourceID: p.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: p.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
