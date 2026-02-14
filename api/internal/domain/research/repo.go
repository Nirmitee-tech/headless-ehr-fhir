package research

import (
	"context"

	"github.com/google/uuid"
)

type ResearchStudyRepository interface {
	Create(ctx context.Context, s *ResearchStudy) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchStudy, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ResearchStudy, error)
	Update(ctx context.Context, s *ResearchStudy) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ResearchStudy, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchStudy, int, error)
	// Arms
	AddArm(ctx context.Context, a *ResearchArm) error
	GetArms(ctx context.Context, studyID uuid.UUID) ([]*ResearchArm, error)
}

type EnrollmentRepository interface {
	Create(ctx context.Context, e *ResearchEnrollment) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchEnrollment, error)
	Update(ctx context.Context, e *ResearchEnrollment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByStudy(ctx context.Context, studyID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error)
}

type AdverseEventRepository interface {
	Create(ctx context.Context, ae *ResearchAdverseEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchAdverseEvent, error)
	Update(ctx context.Context, ae *ResearchAdverseEvent) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByEnrollment(ctx context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchAdverseEvent, int, error)
}

type DeviationRepository interface {
	Create(ctx context.Context, d *ResearchProtocolDeviation) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchProtocolDeviation, error)
	Update(ctx context.Context, d *ResearchProtocolDeviation) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByEnrollment(ctx context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchProtocolDeviation, int, error)
}
