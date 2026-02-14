package scheduling

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
	// Read endpoints – admin, physician, nurse, registrar
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	readGroup.GET("/schedules", h.ListSchedules)
	readGroup.GET("/schedules/:id", h.GetSchedule)
	readGroup.GET("/slots", h.ListSlots)
	readGroup.GET("/slots/:id", h.GetSlot)
	readGroup.GET("/appointments", h.ListAppointments)
	readGroup.GET("/appointments/:id", h.GetAppointment)
	readGroup.GET("/appointments/:id/participants", h.GetParticipants)
	readGroup.GET("/waitlist", h.ListWaitlist)
	readGroup.GET("/waitlist/:id", h.GetWaitlistEntry)

	// Write endpoints – admin, physician, nurse, registrar
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	writeGroup.POST("/schedules", h.CreateSchedule)
	writeGroup.PUT("/schedules/:id", h.UpdateSchedule)
	writeGroup.DELETE("/schedules/:id", h.DeleteSchedule)
	writeGroup.POST("/slots", h.CreateSlot)
	writeGroup.PUT("/slots/:id", h.UpdateSlot)
	writeGroup.DELETE("/slots/:id", h.DeleteSlot)
	writeGroup.POST("/appointments", h.CreateAppointment)
	writeGroup.PUT("/appointments/:id", h.UpdateAppointment)
	writeGroup.DELETE("/appointments/:id", h.DeleteAppointment)
	writeGroup.POST("/appointments/:id/participants", h.AddParticipant)
	writeGroup.POST("/waitlist", h.CreateWaitlistEntry)
	writeGroup.PUT("/waitlist/:id", h.UpdateWaitlistEntry)
	writeGroup.DELETE("/waitlist/:id", h.DeleteWaitlistEntry)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	fhirRead.GET("/Schedule", h.SearchSchedulesFHIR)
	fhirRead.GET("/Schedule/:id", h.GetScheduleFHIR)
	fhirRead.GET("/Slot", h.SearchSlotsFHIR)
	fhirRead.GET("/Slot/:id", h.GetSlotFHIR)
	fhirRead.GET("/Appointment", h.SearchAppointmentsFHIR)
	fhirRead.GET("/Appointment/:id", h.GetAppointmentFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	fhirWrite.POST("/Schedule", h.CreateScheduleFHIR)
	fhirWrite.POST("/Slot", h.CreateSlotFHIR)
	fhirWrite.POST("/Appointment", h.CreateAppointmentFHIR)
}

// -- Schedule Handlers --

func (h *Handler) CreateSchedule(c echo.Context) error {
	var sched Schedule
	if err := c.Bind(&sched); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSchedule(c.Request().Context(), &sched); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sched)
}

func (h *Handler) GetSchedule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sched, err := h.svc.GetSchedule(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "schedule not found")
	}
	return c.JSON(http.StatusOK, sched)
}

func (h *Handler) ListSchedules(c echo.Context) error {
	pg := pagination.FromContext(c)
	if practID := c.QueryParam("practitioner_id"); practID != "" {
		pid, err := uuid.Parse(practID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid practitioner_id")
		}
		items, total, err := h.svc.ListSchedulesByPractitioner(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchSchedules(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSchedule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sched Schedule
	if err := c.Bind(&sched); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sched.ID = id
	if err := h.svc.UpdateSchedule(c.Request().Context(), &sched); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sched)
}

func (h *Handler) DeleteSchedule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSchedule(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Slot Handlers --

func (h *Handler) CreateSlot(c echo.Context) error {
	var sl Slot
	if err := c.Bind(&sl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSlot(c.Request().Context(), &sl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sl)
}

func (h *Handler) GetSlot(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sl, err := h.svc.GetSlot(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "slot not found")
	}
	return c.JSON(http.StatusOK, sl)
}

func (h *Handler) ListSlots(c echo.Context) error {
	pg := pagination.FromContext(c)
	if schedID := c.QueryParam("schedule_id"); schedID != "" {
		sid, err := uuid.Parse(schedID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule_id")
		}
		items, total, err := h.svc.ListSlotsBySchedule(c.Request().Context(), sid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchAvailableSlots(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSlot(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sl Slot
	if err := c.Bind(&sl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sl.ID = id
	if err := h.svc.UpdateSlot(c.Request().Context(), &sl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sl)
}

func (h *Handler) DeleteSlot(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSlot(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Appointment Handlers --

func (h *Handler) CreateAppointment(c echo.Context) error {
	var a Appointment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateAppointment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetAppointment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetAppointment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "appointment not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListAppointments(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListAppointmentsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if practID := c.QueryParam("practitioner_id"); practID != "" {
		pid, err := uuid.Parse(practID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid practitioner_id")
		}
		items, total, err := h.svc.ListAppointmentsByPractitioner(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchAppointments(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateAppointment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a Appointment
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateAppointment(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeleteAppointment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteAppointment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddParticipant(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p AppointmentParticipant
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.AppointmentID = id
	if err := h.svc.AddAppointmentParticipant(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetParticipants(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	participants, err := h.svc.GetAppointmentParticipants(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, participants)
}

// -- Waitlist Handlers --

func (h *Handler) CreateWaitlistEntry(c echo.Context) error {
	var w Waitlist
	if err := c.Bind(&w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateWaitlistEntry(c.Request().Context(), &w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, w)
}

func (h *Handler) GetWaitlistEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	w, err := h.svc.GetWaitlistEntry(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "waitlist entry not found")
	}
	return c.JSON(http.StatusOK, w)
}

func (h *Handler) ListWaitlist(c echo.Context) error {
	pg := pagination.FromContext(c)
	if dept := c.QueryParam("department"); dept != "" {
		items, total, err := h.svc.ListWaitlistByDepartment(c.Request().Context(), dept, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	if practID := c.QueryParam("practitioner_id"); practID != "" {
		pid, err := uuid.Parse(practID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid practitioner_id")
		}
		items, total, err := h.svc.ListWaitlistByPractitioner(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	return echo.NewHTTPError(http.StatusBadRequest, "department or practitioner_id query parameter required")
}

func (h *Handler) UpdateWaitlistEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var w Waitlist
	if err := c.Bind(&w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	w.ID = id
	if err := h.svc.UpdateWaitlistEntry(c.Request().Context(), &w); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, w)
}

func (h *Handler) DeleteWaitlistEntry(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteWaitlistEntry(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchSchedulesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"practitioner", "active", "service-type"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchSchedules(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Schedule"))
}

func (h *Handler) GetScheduleFHIR(c echo.Context) error {
	sched, err := h.svc.GetScheduleByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Schedule", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sched.ToFHIR())
}

func (h *Handler) CreateScheduleFHIR(c echo.Context) error {
	var sched Schedule
	if err := c.Bind(&sched); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSchedule(c.Request().Context(), &sched); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Schedule/"+sched.FHIRID)
	return c.JSON(http.StatusCreated, sched.ToFHIR())
}

func (h *Handler) SearchSlotsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"schedule", "status", "start", "service-type"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchAvailableSlots(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Slot"))
}

func (h *Handler) GetSlotFHIR(c echo.Context) error {
	sl, err := h.svc.GetSlotByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Slot", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sl.ToFHIR())
}

func (h *Handler) CreateSlotFHIR(c echo.Context) error {
	var sl Slot
	if err := c.Bind(&sl); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSlot(c.Request().Context(), &sl); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Slot/"+sl.FHIRID)
	return c.JSON(http.StatusCreated, sl.ToFHIR())
}

func (h *Handler) SearchAppointmentsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "practitioner", "status", "date", "service-type"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchAppointments(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Appointment"))
}

func (h *Handler) GetAppointmentFHIR(c echo.Context) error {
	a, err := h.svc.GetAppointmentByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Appointment", c.Param("id")))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) CreateAppointmentFHIR(c echo.Context) error {
	var a Appointment
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateAppointment(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Appointment/"+a.FHIRID)
	return c.JSON(http.StatusCreated, a.ToFHIR())
}
