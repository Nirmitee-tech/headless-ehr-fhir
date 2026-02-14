package oncology

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

func (h *Handler) RegisterRoutes(api *echo.Group) {
	// Read endpoints – admin, physician, nurse
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	readGroup.GET("/cancer-diagnoses", h.ListCancerDiagnoses)
	readGroup.GET("/cancer-diagnoses/:id", h.GetCancerDiagnosis)
	readGroup.GET("/treatment-protocols", h.ListTreatmentProtocols)
	readGroup.GET("/treatment-protocols/:id", h.GetTreatmentProtocol)
	readGroup.GET("/treatment-protocols/:id/drugs", h.GetProtocolDrugs)
	readGroup.GET("/chemo-cycles", h.ListChemoCycles)
	readGroup.GET("/chemo-cycles/:id", h.GetChemoCycle)
	readGroup.GET("/chemo-cycles/:id/administrations", h.GetChemoAdministrations)
	readGroup.GET("/radiation-therapies", h.ListRadiationTherapies)
	readGroup.GET("/radiation-therapies/:id", h.GetRadiationTherapy)
	readGroup.GET("/radiation-therapies/:id/sessions", h.GetRadiationSessions)
	readGroup.GET("/tumor-markers", h.ListTumorMarkers)
	readGroup.GET("/tumor-markers/:id", h.GetTumorMarker)
	readGroup.GET("/tumor-board-reviews", h.ListTumorBoardReviews)
	readGroup.GET("/tumor-board-reviews/:id", h.GetTumorBoardReview)

	// Write endpoints – admin, physician, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	writeGroup.POST("/cancer-diagnoses", h.CreateCancerDiagnosis)
	writeGroup.PUT("/cancer-diagnoses/:id", h.UpdateCancerDiagnosis)
	writeGroup.DELETE("/cancer-diagnoses/:id", h.DeleteCancerDiagnosis)
	writeGroup.POST("/treatment-protocols", h.CreateTreatmentProtocol)
	writeGroup.PUT("/treatment-protocols/:id", h.UpdateTreatmentProtocol)
	writeGroup.DELETE("/treatment-protocols/:id", h.DeleteTreatmentProtocol)
	writeGroup.POST("/treatment-protocols/:id/drugs", h.AddProtocolDrug)
	writeGroup.POST("/chemo-cycles", h.CreateChemoCycle)
	writeGroup.PUT("/chemo-cycles/:id", h.UpdateChemoCycle)
	writeGroup.DELETE("/chemo-cycles/:id", h.DeleteChemoCycle)
	writeGroup.POST("/chemo-cycles/:id/administrations", h.AddChemoAdministration)
	writeGroup.POST("/radiation-therapies", h.CreateRadiationTherapy)
	writeGroup.PUT("/radiation-therapies/:id", h.UpdateRadiationTherapy)
	writeGroup.DELETE("/radiation-therapies/:id", h.DeleteRadiationTherapy)
	writeGroup.POST("/radiation-therapies/:id/sessions", h.AddRadiationSession)
	writeGroup.POST("/tumor-markers", h.CreateTumorMarker)
	writeGroup.PUT("/tumor-markers/:id", h.UpdateTumorMarker)
	writeGroup.DELETE("/tumor-markers/:id", h.DeleteTumorMarker)
	writeGroup.POST("/tumor-board-reviews", h.CreateTumorBoardReview)
	writeGroup.PUT("/tumor-board-reviews/:id", h.UpdateTumorBoardReview)
	writeGroup.DELETE("/tumor-board-reviews/:id", h.DeleteTumorBoardReview)
}

// -- Cancer Diagnosis Handlers --

func (h *Handler) CreateCancerDiagnosis(c echo.Context) error {
	var d CancerDiagnosis
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCancerDiagnosis(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetCancerDiagnosis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetCancerDiagnosis(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cancer diagnosis not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) ListCancerDiagnoses(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListCancerDiagnosesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListCancerDiagnoses(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCancerDiagnosis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d CancerDiagnosis
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateCancerDiagnosis(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteCancerDiagnosis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCancerDiagnosis(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Treatment Protocol Handlers --

func (h *Handler) CreateTreatmentProtocol(c echo.Context) error {
	var p TreatmentProtocol
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTreatmentProtocol(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetTreatmentProtocol(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetTreatmentProtocol(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "treatment protocol not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListTreatmentProtocols(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListTreatmentProtocols(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTreatmentProtocol(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p TreatmentProtocol
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdateTreatmentProtocol(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeleteTreatmentProtocol(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTreatmentProtocol(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddProtocolDrug(c echo.Context) error {
	protoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d TreatmentProtocolDrug
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ProtocolID = protoID
	if err := h.svc.AddProtocolDrug(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetProtocolDrugs(c echo.Context) error {
	protoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	drugs, err := h.svc.GetProtocolDrugs(c.Request().Context(), protoID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, drugs)
}

// -- Chemo Cycle Handlers --

func (h *Handler) CreateChemoCycle(c echo.Context) error {
	var cycle ChemoCycle
	if err := c.Bind(&cycle); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateChemoCycle(c.Request().Context(), &cycle); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cycle)
}

func (h *Handler) GetChemoCycle(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cycle, err := h.svc.GetChemoCycle(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "chemo cycle not found")
	}
	return c.JSON(http.StatusOK, cycle)
}

func (h *Handler) ListChemoCycles(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListChemoCycles(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateChemoCycle(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cycle ChemoCycle
	if err := c.Bind(&cycle); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cycle.ID = id
	if err := h.svc.UpdateChemoCycle(c.Request().Context(), &cycle); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cycle)
}

func (h *Handler) DeleteChemoCycle(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteChemoCycle(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddChemoAdministration(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a ChemoAdministration
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.CycleID = cycleID
	if err := h.svc.AddChemoAdministration(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetChemoAdministrations(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetChemoAdministrations(c.Request().Context(), cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Radiation Therapy Handlers --

func (h *Handler) CreateRadiationTherapy(c echo.Context) error {
	var r RadiationTherapy
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateRadiationTherapy(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetRadiationTherapy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetRadiationTherapy(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "radiation therapy not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListRadiationTherapies(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListRadiationTherapies(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateRadiationTherapy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r RadiationTherapy
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateRadiationTherapy(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteRadiationTherapy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteRadiationTherapy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddRadiationSession(c echo.Context) error {
	radID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var s RadiationSession
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	s.RadiationTherapyID = radID
	if err := h.svc.AddRadiationSession(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, s)
}

func (h *Handler) GetRadiationSessions(c echo.Context) error {
	radID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sessions, err := h.svc.GetRadiationSessions(c.Request().Context(), radID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sessions)
}

// -- Tumor Marker Handlers --

func (h *Handler) CreateTumorMarker(c echo.Context) error {
	var m TumorMarker
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTumorMarker(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetTumorMarker(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	m, err := h.svc.GetTumorMarker(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tumor marker not found")
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) ListTumorMarkers(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListTumorMarkers(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTumorMarker(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var m TumorMarker
	if err := c.Bind(&m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	m.ID = id
	if err := h.svc.UpdateTumorMarker(c.Request().Context(), &m); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteTumorMarker(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTumorMarker(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Tumor Board Review Handlers --

func (h *Handler) CreateTumorBoardReview(c echo.Context) error {
	var r TumorBoardReview
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateTumorBoardReview(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetTumorBoardReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, err := h.svc.GetTumorBoardReview(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tumor board review not found")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) ListTumorBoardReviews(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListTumorBoardReviews(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateTumorBoardReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var r TumorBoardReview
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	r.ID = id
	if err := h.svc.UpdateTumorBoardReview(c.Request().Context(), &r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteTumorBoardReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTumorBoardReview(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
