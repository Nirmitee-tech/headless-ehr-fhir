package diagnostics

import (
	"context"

	"github.com/google/uuid"
)

type ServiceRequestRepository interface {
	Create(ctx context.Context, sr *ServiceRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*ServiceRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ServiceRequest, error)
	Update(ctx context.Context, sr *ServiceRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ServiceRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ServiceRequest, int, error)
}

type SpecimenRepository interface {
	Create(ctx context.Context, sp *Specimen) error
	GetByID(ctx context.Context, id uuid.UUID) (*Specimen, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Specimen, error)
	Update(ctx context.Context, sp *Specimen) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Specimen, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Specimen, int, error)
}

type DiagnosticReportRepository interface {
	Create(ctx context.Context, dr *DiagnosticReport) error
	GetByID(ctx context.Context, id uuid.UUID) (*DiagnosticReport, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DiagnosticReport, error)
	Update(ctx context.Context, dr *DiagnosticReport) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*DiagnosticReport, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DiagnosticReport, int, error)
	// Results junction
	AddResult(ctx context.Context, reportID uuid.UUID, observationID uuid.UUID) error
	GetResults(ctx context.Context, reportID uuid.UUID) ([]uuid.UUID, error)
	RemoveResult(ctx context.Context, reportID uuid.UUID, observationID uuid.UUID) error
}

type ImagingStudyRepository interface {
	Create(ctx context.Context, is *ImagingStudy) error
	GetByID(ctx context.Context, id uuid.UUID) (*ImagingStudy, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ImagingStudy, error)
	Update(ctx context.Context, is *ImagingStudy) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImagingStudy, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImagingStudy, int, error)
}
