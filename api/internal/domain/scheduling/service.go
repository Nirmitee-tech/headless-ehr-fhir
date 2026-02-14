package scheduling

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	schedules    ScheduleRepository
	slots        SlotRepository
	appointments AppointmentRepository
	waitlist     WaitlistRepository
}

func NewService(sched ScheduleRepository, slot SlotRepository, appt AppointmentRepository, wl WaitlistRepository) *Service {
	return &Service{schedules: sched, slots: slot, appointments: appt, waitlist: wl}
}

// -- Schedule --

func (s *Service) CreateSchedule(ctx context.Context, sched *Schedule) error {
	if sched.PractitionerID == uuid.Nil {
		return fmt.Errorf("practitioner_id is required")
	}
	if sched.Active == nil {
		active := true
		sched.Active = &active
	}
	return s.schedules.Create(ctx, sched)
}

func (s *Service) GetSchedule(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	return s.schedules.GetByID(ctx, id)
}

func (s *Service) GetScheduleByFHIRID(ctx context.Context, fhirID string) (*Schedule, error) {
	return s.schedules.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSchedule(ctx context.Context, sched *Schedule) error {
	return s.schedules.Update(ctx, sched)
}

func (s *Service) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	return s.schedules.Delete(ctx, id)
}

func (s *Service) ListSchedulesByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Schedule, int, error) {
	return s.schedules.ListByPractitioner(ctx, practitionerID, limit, offset)
}

func (s *Service) SearchSchedules(ctx context.Context, params map[string]string, limit, offset int) ([]*Schedule, int, error) {
	return s.schedules.Search(ctx, params, limit, offset)
}

// -- Slot --

var validSlotStatuses = map[string]bool{
	"busy": true, "free": true, "busy-unavailable": true,
	"busy-tentative": true, "entered-in-error": true,
}

func (s *Service) CreateSlot(ctx context.Context, sl *Slot) error {
	if sl.ScheduleID == uuid.Nil {
		return fmt.Errorf("schedule_id is required")
	}
	if sl.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}
	if sl.EndTime.IsZero() {
		return fmt.Errorf("end_time is required")
	}
	if sl.Status == "" {
		sl.Status = "free"
	}
	if !validSlotStatuses[sl.Status] {
		return fmt.Errorf("invalid slot status: %s", sl.Status)
	}
	return s.slots.Create(ctx, sl)
}

func (s *Service) GetSlot(ctx context.Context, id uuid.UUID) (*Slot, error) {
	return s.slots.GetByID(ctx, id)
}

func (s *Service) GetSlotByFHIRID(ctx context.Context, fhirID string) (*Slot, error) {
	return s.slots.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSlot(ctx context.Context, sl *Slot) error {
	if sl.Status != "" && !validSlotStatuses[sl.Status] {
		return fmt.Errorf("invalid slot status: %s", sl.Status)
	}
	return s.slots.Update(ctx, sl)
}

func (s *Service) DeleteSlot(ctx context.Context, id uuid.UUID) error {
	return s.slots.Delete(ctx, id)
}

func (s *Service) ListSlotsBySchedule(ctx context.Context, scheduleID uuid.UUID, limit, offset int) ([]*Slot, int, error) {
	return s.slots.ListBySchedule(ctx, scheduleID, limit, offset)
}

func (s *Service) SearchAvailableSlots(ctx context.Context, params map[string]string, limit, offset int) ([]*Slot, int, error) {
	return s.slots.SearchAvailable(ctx, params, limit, offset)
}

// -- Appointment --

var validAppointmentStatuses = map[string]bool{
	"proposed": true, "pending": true, "booked": true, "arrived": true,
	"fulfilled": true, "cancelled": true, "noshow": true,
	"entered-in-error": true, "checked-in": true, "waitlist": true,
}

func (s *Service) CreateAppointment(ctx context.Context, a *Appointment) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.StartTime == nil || a.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}
	if a.Status == "" {
		a.Status = "proposed"
	}
	if !validAppointmentStatuses[a.Status] {
		return fmt.Errorf("invalid appointment status: %s", a.Status)
	}
	return s.appointments.Create(ctx, a)
}

func (s *Service) GetAppointment(ctx context.Context, id uuid.UUID) (*Appointment, error) {
	return s.appointments.GetByID(ctx, id)
}

func (s *Service) GetAppointmentByFHIRID(ctx context.Context, fhirID string) (*Appointment, error) {
	return s.appointments.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateAppointment(ctx context.Context, a *Appointment) error {
	if a.Status != "" && !validAppointmentStatuses[a.Status] {
		return fmt.Errorf("invalid appointment status: %s", a.Status)
	}
	return s.appointments.Update(ctx, a)
}

func (s *Service) DeleteAppointment(ctx context.Context, id uuid.UUID) error {
	return s.appointments.Delete(ctx, id)
}

func (s *Service) ListAppointmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Appointment, int, error) {
	return s.appointments.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListAppointmentsByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Appointment, int, error) {
	return s.appointments.ListByPractitioner(ctx, practitionerID, limit, offset)
}

func (s *Service) SearchAppointments(ctx context.Context, params map[string]string, limit, offset int) ([]*Appointment, int, error) {
	return s.appointments.Search(ctx, params, limit, offset)
}

func (s *Service) AddAppointmentParticipant(ctx context.Context, p *AppointmentParticipant) error {
	if p.AppointmentID == uuid.Nil {
		return fmt.Errorf("appointment_id is required")
	}
	if p.ActorType == "" {
		return fmt.Errorf("actor_type is required")
	}
	if p.ActorID == uuid.Nil {
		return fmt.Errorf("actor_id is required")
	}
	if p.Status == "" {
		p.Status = "needs-action"
	}
	return s.appointments.AddParticipant(ctx, p)
}

func (s *Service) GetAppointmentParticipants(ctx context.Context, appointmentID uuid.UUID) ([]*AppointmentParticipant, error) {
	return s.appointments.GetParticipants(ctx, appointmentID)
}

func (s *Service) RemoveAppointmentParticipant(ctx context.Context, id uuid.UUID) error {
	return s.appointments.RemoveParticipant(ctx, id)
}

// -- Waitlist --

var validWaitlistStatuses = map[string]bool{
	"waiting": true, "called": true, "in-consult": true,
	"completed": true, "cancelled": true, "no-show": true,
}

func (s *Service) CreateWaitlistEntry(ctx context.Context, w *Waitlist) error {
	if w.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if w.Status == "" {
		w.Status = "waiting"
	}
	if !validWaitlistStatuses[w.Status] {
		return fmt.Errorf("invalid waitlist status: %s", w.Status)
	}
	return s.waitlist.Create(ctx, w)
}

func (s *Service) GetWaitlistEntry(ctx context.Context, id uuid.UUID) (*Waitlist, error) {
	return s.waitlist.GetByID(ctx, id)
}

func (s *Service) UpdateWaitlistEntry(ctx context.Context, w *Waitlist) error {
	if w.Status != "" && !validWaitlistStatuses[w.Status] {
		return fmt.Errorf("invalid waitlist status: %s", w.Status)
	}
	return s.waitlist.Update(ctx, w)
}

func (s *Service) DeleteWaitlistEntry(ctx context.Context, id uuid.UUID) error {
	return s.waitlist.Delete(ctx, id)
}

func (s *Service) ListWaitlistByDepartment(ctx context.Context, department string, limit, offset int) ([]*Waitlist, int, error) {
	return s.waitlist.ListByDepartment(ctx, department, limit, offset)
}

func (s *Service) ListWaitlistByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Waitlist, int, error) {
	return s.waitlist.ListByPractitioner(ctx, practitionerID, limit, offset)
}
