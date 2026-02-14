package research

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
	// Read endpoints – admin, physician
	readGroup := api.Group("", auth.RequireRole("admin", "physician"))
	readGroup.GET("/research-studies", h.ListStudies)
	readGroup.GET("/research-studies/:id", h.GetStudy)
	readGroup.GET("/research-studies/:id/arms", h.GetArms)
	readGroup.GET("/research-enrollments", h.ListEnrollments)
	readGroup.GET("/research-enrollments/:id", h.GetEnrollment)
	readGroup.GET("/adverse-events", h.ListAdverseEvents)
	readGroup.GET("/adverse-events/:id", h.GetAdverseEvent)
	readGroup.GET("/protocol-deviations", h.ListDeviations)
	readGroup.GET("/protocol-deviations/:id", h.GetDeviation)

	// Write endpoints – admin, physician
	writeGroup := api.Group("", auth.RequireRole("admin", "physician"))
	writeGroup.POST("/research-studies", h.CreateStudy)
	writeGroup.PUT("/research-studies/:id", h.UpdateStudy)
	writeGroup.DELETE("/research-studies/:id", h.DeleteStudy)
	writeGroup.POST("/research-studies/:id/arms", h.AddArm)
	writeGroup.POST("/research-enrollments", h.CreateEnrollment)
	writeGroup.PUT("/research-enrollments/:id", h.UpdateEnrollment)
	writeGroup.DELETE("/research-enrollments/:id", h.DeleteEnrollment)
	writeGroup.POST("/adverse-events", h.CreateAdverseEvent)
	writeGroup.PUT("/adverse-events/:id", h.UpdateAdverseEvent)
	writeGroup.DELETE("/adverse-events/:id", h.DeleteAdverseEvent)
	writeGroup.POST("/protocol-deviations", h.CreateDeviation)
	writeGroup.PUT("/protocol-deviations/:id", h.UpdateDeviation)
	writeGroup.DELETE("/protocol-deviations/:id", h.DeleteDeviation)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician"))
	fhirRead.GET("/ResearchStudy", h.SearchStudiesFHIR)
	fhirRead.GET("/ResearchStudy/:id", h.GetStudyFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician"))
	fhirWrite.POST("/ResearchStudy", h.CreateStudyFHIR)
}

// -- Research Study Handlers --

func (h *Handler) CreateStudy(c echo.Context) error {
	var s ResearchStudy
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateStudy(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, s)
}

func (h *Handler) GetStudy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	s, err := h.svc.GetStudy(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "research study not found")
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) ListStudies(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListStudies(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateStudy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var s ResearchStudy
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	s.ID = id
	if err := h.svc.UpdateStudy(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) DeleteStudy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteStudy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddArm(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a ResearchArm
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.StudyID = id
	if err := h.svc.AddStudyArm(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetArms(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	arms, err := h.svc.GetStudyArms(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, arms)
}

// -- Enrollment Handlers --

func (h *Handler) CreateEnrollment(c echo.Context) error {
	var e ResearchEnrollment
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEnrollment(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetEnrollment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	e, err := h.svc.GetEnrollment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "enrollment not found")
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) ListEnrollments(c echo.Context) error {
	pg := pagination.FromContext(c)
	if studyID := c.QueryParam("study_id"); studyID != "" {
		sid, err := uuid.Parse(studyID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid study_id")
		}
		items, total, err := h.svc.ListEnrollmentsByStudy(c.Request().Context(), sid, pg.Limit, pg.Offset)
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
		items, total, err := h.svc.ListEnrollmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "study_id or patient_id query parameter is required")
}

func (h *Handler) UpdateEnrollment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e ResearchEnrollment
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.ID = id
	if err := h.svc.UpdateEnrollment(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, e)
}

func (h *Handler) DeleteEnrollment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEnrollment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Adverse Event Handlers --

func (h *Handler) CreateAdverseEvent(c echo.Context) error {
	var ae ResearchAdverseEvent
	if err := c.Bind(&ae); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateAdverseEvent(c.Request().Context(), &ae); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ae)
}

func (h *Handler) GetAdverseEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ae, err := h.svc.GetAdverseEvent(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "adverse event not found")
	}
	return c.JSON(http.StatusOK, ae)
}

func (h *Handler) ListAdverseEvents(c echo.Context) error {
	pg := pagination.FromContext(c)
	enrollmentID := c.QueryParam("enrollment_id")
	if enrollmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "enrollment_id query parameter is required")
	}
	eid, err := uuid.Parse(enrollmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid enrollment_id")
	}
	items, total, err := h.svc.ListAdverseEventsByEnrollment(c.Request().Context(), eid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateAdverseEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ae ResearchAdverseEvent
	if err := c.Bind(&ae); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ae.ID = id
	if err := h.svc.UpdateAdverseEvent(c.Request().Context(), &ae); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ae)
}

func (h *Handler) DeleteAdverseEvent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteAdverseEvent(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Protocol Deviation Handlers --

func (h *Handler) CreateDeviation(c echo.Context) error {
	var d ResearchProtocolDeviation
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDeviation(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDeviation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetDeviation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "protocol deviation not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) ListDeviations(c echo.Context) error {
	pg := pagination.FromContext(c)
	enrollmentID := c.QueryParam("enrollment_id")
	if enrollmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "enrollment_id query parameter is required")
	}
	eid, err := uuid.Parse(enrollmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid enrollment_id")
	}
	items, total, err := h.svc.ListDeviationsByEnrollment(c.Request().Context(), eid, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDeviation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d ResearchProtocolDeviation
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateDeviation(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteDeviation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDeviation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchStudiesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "title", "protocol"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchStudies(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ResearchStudy"))
}

func (h *Handler) GetStudyFHIR(c echo.Context) error {
	s, err := h.svc.GetStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ResearchStudy", c.Param("id")))
	}
	return c.JSON(http.StatusOK, s.ToFHIR())
}

func (h *Handler) CreateStudyFHIR(c echo.Context) error {
	var s ResearchStudy
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateStudy(c.Request().Context(), &s); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ResearchStudy/"+s.FHIRID)
	return c.JSON(http.StatusCreated, s.ToFHIR())
}
