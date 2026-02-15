package research

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	studies     ResearchStudyRepository
	enrollments EnrollmentRepository
	adverse     AdverseEventRepository
	deviations  DeviationRepository
	vt          *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(
	studies ResearchStudyRepository,
	enrollments EnrollmentRepository,
	adverse AdverseEventRepository,
	deviations DeviationRepository,
) *Service {
	return &Service{
		studies:     studies,
		enrollments: enrollments,
		adverse:     adverse,
		deviations:  deviations,
	}
}

// -- Research Study --

var validStudyStatuses = map[string]bool{
	"in-review": true, "approved": true, "active-recruiting": true,
	"active-not-recruiting": true, "temporarily-closed": true,
	"closed": true, "completed": true, "withdrawn": true, "suspended": true,
}

func (s *Service) CreateStudy(ctx context.Context, st *ResearchStudy) error {
	if st.ProtocolNumber == "" {
		return fmt.Errorf("protocol_number is required")
	}
	if st.Title == "" {
		return fmt.Errorf("title is required")
	}
	if st.Status == "" {
		st.Status = "in-review"
	}
	if !validStudyStatuses[st.Status] {
		return fmt.Errorf("invalid status: %s", st.Status)
	}
	if err := s.studies.Create(ctx, st); err != nil {
		return err
	}
	st.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ResearchStudy", st.FHIRID, st.ToFHIR())
	}
	return nil
}

func (s *Service) GetStudy(ctx context.Context, id uuid.UUID) (*ResearchStudy, error) {
	return s.studies.GetByID(ctx, id)
}

func (s *Service) GetStudyByFHIRID(ctx context.Context, fhirID string) (*ResearchStudy, error) {
	return s.studies.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateStudy(ctx context.Context, st *ResearchStudy) error {
	if st.Status != "" && !validStudyStatuses[st.Status] {
		return fmt.Errorf("invalid status: %s", st.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ResearchStudy", st.FHIRID, st.VersionID, st.ToFHIR())
		if err == nil {
			st.VersionID = newVer
		}
	}
	return s.studies.Update(ctx, st)
}

func (s *Service) DeleteStudy(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		st, err := s.studies.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ResearchStudy", st.FHIRID, st.VersionID)
		}
	}
	return s.studies.Delete(ctx, id)
}

func (s *Service) ListStudies(ctx context.Context, limit, offset int) ([]*ResearchStudy, int, error) {
	return s.studies.List(ctx, limit, offset)
}

func (s *Service) SearchStudies(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchStudy, int, error) {
	return s.studies.Search(ctx, params, limit, offset)
}

func (s *Service) AddStudyArm(ctx context.Context, a *ResearchArm) error {
	if a.StudyID == uuid.Nil {
		return fmt.Errorf("study_id is required")
	}
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	return s.studies.AddArm(ctx, a)
}

func (s *Service) GetStudyArms(ctx context.Context, studyID uuid.UUID) ([]*ResearchArm, error) {
	return s.studies.GetArms(ctx, studyID)
}

// -- Enrollment --

var validEnrollmentStatuses = map[string]bool{
	"pre-screening": true, "screening": true, "screen-fail": true,
	"enrolled": true, "active": true, "on-study-treatment": true,
	"follow-up": true, "completed": true, "early-termination": true,
	"withdrawn": true, "lost-to-followup": true, "deceased": true,
}

func (s *Service) CreateEnrollment(ctx context.Context, e *ResearchEnrollment) error {
	if e.StudyID == uuid.Nil {
		return fmt.Errorf("study_id is required")
	}
	if e.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if e.Status == "" {
		e.Status = "pre-screening"
	}
	if !validEnrollmentStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	return s.enrollments.Create(ctx, e)
}

func (s *Service) GetEnrollment(ctx context.Context, id uuid.UUID) (*ResearchEnrollment, error) {
	return s.enrollments.GetByID(ctx, id)
}

func (s *Service) UpdateEnrollment(ctx context.Context, e *ResearchEnrollment) error {
	if e.Status != "" && !validEnrollmentStatuses[e.Status] {
		return fmt.Errorf("invalid status: %s", e.Status)
	}
	return s.enrollments.Update(ctx, e)
}

func (s *Service) DeleteEnrollment(ctx context.Context, id uuid.UUID) error {
	return s.enrollments.Delete(ctx, id)
}

func (s *Service) ListEnrollmentsByStudy(ctx context.Context, studyID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error) {
	return s.enrollments.ListByStudy(ctx, studyID, limit, offset)
}

func (s *Service) ListEnrollmentsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error) {
	return s.enrollments.ListByPatient(ctx, patientID, limit, offset)
}

// -- Adverse Event --

func (s *Service) CreateAdverseEvent(ctx context.Context, ae *ResearchAdverseEvent) error {
	if ae.EnrollmentID == uuid.Nil {
		return fmt.Errorf("enrollment_id is required")
	}
	if ae.Description == "" {
		return fmt.Errorf("description is required")
	}
	return s.adverse.Create(ctx, ae)
}

func (s *Service) GetAdverseEvent(ctx context.Context, id uuid.UUID) (*ResearchAdverseEvent, error) {
	return s.adverse.GetByID(ctx, id)
}

func (s *Service) UpdateAdverseEvent(ctx context.Context, ae *ResearchAdverseEvent) error {
	return s.adverse.Update(ctx, ae)
}

func (s *Service) DeleteAdverseEvent(ctx context.Context, id uuid.UUID) error {
	return s.adverse.Delete(ctx, id)
}

func (s *Service) ListAdverseEventsByEnrollment(ctx context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchAdverseEvent, int, error) {
	return s.adverse.ListByEnrollment(ctx, enrollmentID, limit, offset)
}

// -- Protocol Deviation --

func (s *Service) CreateDeviation(ctx context.Context, d *ResearchProtocolDeviation) error {
	if d.EnrollmentID == uuid.Nil {
		return fmt.Errorf("enrollment_id is required")
	}
	if d.Description == "" {
		return fmt.Errorf("description is required")
	}
	return s.deviations.Create(ctx, d)
}

func (s *Service) GetDeviation(ctx context.Context, id uuid.UUID) (*ResearchProtocolDeviation, error) {
	return s.deviations.GetByID(ctx, id)
}

func (s *Service) UpdateDeviation(ctx context.Context, d *ResearchProtocolDeviation) error {
	return s.deviations.Update(ctx, d)
}

func (s *Service) DeleteDeviation(ctx context.Context, id uuid.UUID) error {
	return s.deviations.Delete(ctx, id)
}

func (s *Service) ListDeviationsByEnrollment(ctx context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchProtocolDeviation, int, error) {
	return s.deviations.ListByEnrollment(ctx, enrollmentID, limit, offset)
}
