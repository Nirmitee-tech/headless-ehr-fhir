package scheduling

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Common errors returned by the self-scheduling API.
var (
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotAlreadyBooked  = errors.New("slot is already booked")
	ErrSlotNotFree        = errors.New("slot is not available")
	ErrMissingSlotID      = errors.New("slot_id is required")
	ErrMissingPatientID   = errors.New("patient_id is required")
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrWrongPatient       = errors.New("patient is not authorized to access this appointment")
)

// SlotFinder searches for available appointment slots.
type SlotFinder interface {
	FindAvailableSlots(ctx context.Context, params SlotSearchParams) ([]AvailableSlot, error)
}

// BookingService creates and manages patient appointments.
type BookingService interface {
	BookAppointment(ctx context.Context, req BookingRequest) (*BookingConfirmation, error)
	CancelAppointment(ctx context.Context, appointmentID, patientID, reason string) error
	GetAppointment(ctx context.Context, appointmentID, patientID string) (*BookingConfirmation, error)
	ListPatientAppointments(ctx context.Context, patientID string, status string, limit int) ([]BookingConfirmation, error)
}

// SlotSearchParams defines the search criteria for available slots.
type SlotSearchParams struct {
	ScheduleID  string    // optional: filter by specific schedule/provider
	ServiceType string    // optional: e.g. "general", "specialist", "lab"
	StartDate   time.Time // required: earliest date
	EndDate     time.Time // required: latest date
	Duration    int       // optional: desired duration in minutes (default 30)
}

// AvailableSlot represents a bookable time slot.
type AvailableSlot struct {
	ID           string    `json:"id"`
	ScheduleID   string    `json:"schedule_id"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	Duration     int       `json:"duration_minutes"`
	ServiceType  string    `json:"service_type,omitempty"`
	ProviderName string    `json:"provider_name,omitempty"`
	LocationName string    `json:"location_name,omitempty"`
	Status       string    `json:"status"`
}

// BookingRequest is the patient's appointment booking request.
type BookingRequest struct {
	SlotID       string `json:"slot_id"`
	PatientID    string `json:"patient_id"`
	Reason       string `json:"reason,omitempty"`
	Comment      string `json:"comment,omitempty"`
	ContactPhone string `json:"contact_phone,omitempty"`
	ContactEmail string `json:"contact_email,omitempty"`
}

// BookingConfirmation is returned after a successful booking.
type BookingConfirmation struct {
	AppointmentID string    `json:"appointment_id"`
	Status        string    `json:"status"`
	SlotID        string    `json:"slot_id"`
	PatientID     string    `json:"patient_id"`
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	ProviderName  string    `json:"provider_name,omitempty"`
	LocationName  string    `json:"location_name,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// SelfScheduleManager is an in-memory implementation of both SlotFinder and BookingService.
type SelfScheduleManager struct {
	mu           sync.RWMutex
	slots        map[string]*AvailableSlot      // slot ID -> slot
	appointments map[string]*BookingConfirmation // appointment ID -> booking
	slotBookings map[string]string               // slot ID -> appointment ID (prevents double-booking)
}

// NewSelfScheduleManager creates a new SelfScheduleManager.
func NewSelfScheduleManager() *SelfScheduleManager {
	return &SelfScheduleManager{
		slots:        make(map[string]*AvailableSlot),
		appointments: make(map[string]*BookingConfirmation),
		slotBookings: make(map[string]string),
	}
}

// AddSlot adds a slot to the manager for test setup and initialization.
func (m *SelfScheduleManager) AddSlot(slot AvailableSlot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := slot
	m.slots[slot.ID] = &s
}

// FindAvailableSlots returns slots matching the search criteria.
func (m *SelfScheduleManager) FindAvailableSlots(_ context.Context, params SlotSearchParams) ([]AvailableSlot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []AvailableSlot
	for _, slot := range m.slots {
		// Only return free slots that are not already booked.
		if slot.Status != "free" {
			continue
		}
		if _, booked := m.slotBookings[slot.ID]; booked {
			continue
		}

		// Filter by date range: slot start must be >= StartDate and < EndDate.
		if !params.StartDate.IsZero() && slot.Start.Before(params.StartDate) {
			continue
		}
		if !params.EndDate.IsZero() && !slot.Start.Before(params.EndDate) {
			continue
		}

		// Filter by ScheduleID if provided.
		if params.ScheduleID != "" && slot.ScheduleID != params.ScheduleID {
			continue
		}

		// Filter by ServiceType if provided.
		if params.ServiceType != "" && slot.ServiceType != params.ServiceType {
			continue
		}

		results = append(results, *slot)
	}

	// Sort by start time ascending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Start.Before(results[j].Start)
	})

	return results, nil
}

// BookAppointment books a slot for a patient.
func (m *SelfScheduleManager) BookAppointment(_ context.Context, req BookingRequest) (*BookingConfirmation, error) {
	if req.SlotID == "" {
		return nil, ErrMissingSlotID
	}
	if req.PatientID == "" {
		return nil, ErrMissingPatientID
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	slot, exists := m.slots[req.SlotID]
	if !exists {
		return nil, ErrSlotNotFound
	}
	if slot.Status != "free" {
		return nil, ErrSlotNotFree
	}
	if _, booked := m.slotBookings[req.SlotID]; booked {
		return nil, ErrSlotAlreadyBooked
	}

	appointmentID := uuid.New().String()
	now := time.Now()

	confirmation := &BookingConfirmation{
		AppointmentID: appointmentID,
		Status:        "booked",
		SlotID:        req.SlotID,
		PatientID:     req.PatientID,
		Start:         slot.Start,
		End:           slot.End,
		ProviderName:  slot.ProviderName,
		LocationName:  slot.LocationName,
		Reason:        req.Reason,
		CreatedAt:     now,
	}

	m.appointments[appointmentID] = confirmation
	m.slotBookings[req.SlotID] = appointmentID

	return confirmation, nil
}

// CancelAppointment cancels an existing appointment.
func (m *SelfScheduleManager) CancelAppointment(_ context.Context, appointmentID, patientID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	appt, exists := m.appointments[appointmentID]
	if !exists {
		return ErrAppointmentNotFound
	}
	if appt.PatientID != patientID {
		return ErrWrongPatient
	}

	appt.Status = "cancelled"

	// Free the slot by removing it from slotBookings.
	delete(m.slotBookings, appt.SlotID)

	return nil
}

// GetAppointment retrieves a single appointment, verifying patient ownership.
func (m *SelfScheduleManager) GetAppointment(_ context.Context, appointmentID, patientID string) (*BookingConfirmation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	appt, exists := m.appointments[appointmentID]
	if !exists {
		return nil, ErrAppointmentNotFound
	}
	if appt.PatientID != patientID {
		return nil, ErrWrongPatient
	}

	return appt, nil
}

// ListPatientAppointments lists appointments for a specific patient.
func (m *SelfScheduleManager) ListPatientAppointments(_ context.Context, patientID string, status string, limit int) ([]BookingConfirmation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []BookingConfirmation
	for _, appt := range m.appointments {
		if appt.PatientID != patientID {
			continue
		}
		if status != "" && appt.Status != status {
			continue
		}
		results = append(results, *appt)
	}

	// Sort by start time ascending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Start.Before(results[j].Start)
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SelfScheduleHandler provides HTTP handlers for the patient self-scheduling API.
type SelfScheduleHandler struct {
	finder SlotFinder
	booker BookingService
}

// NewSelfScheduleHandler creates a new SelfScheduleHandler.
func NewSelfScheduleHandler(finder SlotFinder, booker BookingService) *SelfScheduleHandler {
	return &SelfScheduleHandler{
		finder: finder,
		booker: booker,
	}
}

// RegisterRoutes registers the self-scheduling API routes on the given Echo group.
func (h *SelfScheduleHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/scheduling/slots", h.SearchSlots)
	g.POST("/scheduling/book", h.BookAppointment)
	g.POST("/scheduling/cancel/:id", h.CancelAppointment)
	g.GET("/scheduling/appointments", h.ListAppointments)
	g.GET("/scheduling/appointments/:id", h.GetAppointment)
}

// SearchSlots handles GET /scheduling/slots.
func (h *SelfScheduleHandler) SearchSlots(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	if startStr == "" || endStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "start and end query parameters are required")
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid start date format: %s", err.Error()))
	}
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid end date format: %s", err.Error()))
	}
	// End date is exclusive: add one day so slots on the end date are included.
	endDate = endDate.AddDate(0, 0, 1)

	params := SlotSearchParams{
		ScheduleID:  c.QueryParam("schedule_id"),
		ServiceType: c.QueryParam("service_type"),
		StartDate:   startDate,
		EndDate:     endDate,
	}

	slots, err := h.finder.FindAvailableSlots(c.Request().Context(), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, slots)
}

// cancelRequest is the JSON body for the cancel endpoint.
type cancelRequest struct {
	PatientID string `json:"patient_id"`
	Reason    string `json:"reason"`
}

// BookAppointment handles POST /scheduling/book.
func (h *SelfScheduleHandler) BookAppointment(c echo.Context) error {
	var req BookingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	confirmation, err := h.booker.BookAppointment(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, ErrMissingSlotID) || errors.Is(err, ErrMissingPatientID) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, ErrSlotNotFound) || errors.Is(err, ErrSlotNotFree) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, ErrSlotAlreadyBooked) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, confirmation)
}

// CancelAppointment handles POST /scheduling/cancel/:id.
func (h *SelfScheduleHandler) CancelAppointment(c echo.Context) error {
	appointmentID := c.Param("id")

	var req cancelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.PatientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}

	err := h.booker.CancelAppointment(c.Request().Context(), appointmentID, req.PatientID, req.Reason)
	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, ErrWrongPatient) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Return the updated appointment.
	appt, err := h.booker.GetAppointment(c.Request().Context(), appointmentID, req.PatientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, appt)
}

// ListAppointments handles GET /scheduling/appointments.
func (h *SelfScheduleHandler) ListAppointments(c echo.Context) error {
	patientID := c.QueryParam("patient_id")
	if patientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}

	status := c.QueryParam("status")
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	appointments, err := h.booker.ListPatientAppointments(c.Request().Context(), patientID, status, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, appointments)
}

// GetAppointment handles GET /scheduling/appointments/:id.
func (h *SelfScheduleHandler) GetAppointment(c echo.Context) error {
	appointmentID := c.Param("id")
	patientID := c.QueryParam("patient_id")

	if patientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}

	appt, err := h.booker.GetAppointment(c.Request().Context(), appointmentID, patientID)
	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, ErrWrongPatient) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, appt)
}
