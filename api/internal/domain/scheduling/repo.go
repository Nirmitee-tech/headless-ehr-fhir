package scheduling

import (
	"context"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	Create(ctx context.Context, s *Schedule) error
	GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Schedule, error)
	Update(ctx context.Context, s *Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Schedule, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Schedule, int, error)
}

type SlotRepository interface {
	Create(ctx context.Context, sl *Slot) error
	GetByID(ctx context.Context, id uuid.UUID) (*Slot, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Slot, error)
	Update(ctx context.Context, sl *Slot) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListBySchedule(ctx context.Context, scheduleID uuid.UUID, limit, offset int) ([]*Slot, int, error)
	SearchAvailable(ctx context.Context, params map[string]string, limit, offset int) ([]*Slot, int, error)
}

type AppointmentRepository interface {
	Create(ctx context.Context, a *Appointment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Appointment, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Appointment, error)
	Update(ctx context.Context, a *Appointment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Appointment, int, error)
	ListByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Appointment, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Appointment, int, error)
	// Participants
	AddParticipant(ctx context.Context, p *AppointmentParticipant) error
	GetParticipants(ctx context.Context, appointmentID uuid.UUID) ([]*AppointmentParticipant, error)
	RemoveParticipant(ctx context.Context, id uuid.UUID) error
}

type AppointmentResponseRepository interface {
	Create(ctx context.Context, ar *AppointmentResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*AppointmentResponse, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*AppointmentResponse, error)
	Update(ctx context.Context, ar *AppointmentResponse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*AppointmentResponse, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AppointmentResponse, int, error)
}

type WaitlistRepository interface {
	Create(ctx context.Context, w *Waitlist) error
	GetByID(ctx context.Context, id uuid.UUID) (*Waitlist, error)
	Update(ctx context.Context, w *Waitlist) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByDepartment(ctx context.Context, department string, limit, offset int) ([]*Waitlist, int, error)
	ListByPractitioner(ctx context.Context, practitionerID uuid.UUID, limit, offset int) ([]*Waitlist, int, error)
}
