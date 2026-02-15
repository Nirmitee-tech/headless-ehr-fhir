package clinical

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// ClinicalSafetyService groups Flag, DetectedIssue, AdverseEvent, ClinicalImpression, RiskAssessment.
type ClinicalSafetyService struct {
	flags               FlagRepository
	detectedIssues      DetectedIssueRepository
	adverseEvents       AdverseEventRepository
	clinicalImpressions ClinicalImpressionRepository
	riskAssessments     RiskAssessmentRepository
	vt                  *fhir.VersionTracker
}

func NewClinicalSafetyService(
	flags FlagRepository,
	detectedIssues DetectedIssueRepository,
	adverseEvents AdverseEventRepository,
	clinicalImpressions ClinicalImpressionRepository,
	riskAssessments RiskAssessmentRepository,
) *ClinicalSafetyService {
	return &ClinicalSafetyService{
		flags:               flags,
		detectedIssues:      detectedIssues,
		adverseEvents:       adverseEvents,
		clinicalImpressions: clinicalImpressions,
		riskAssessments:     riskAssessments,
	}
}

func (s *ClinicalSafetyService) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *ClinicalSafetyService) VersionTracker() *fhir.VersionTracker      { return s.vt }

// -- Flag --

var validFlagStatuses = map[string]bool{
	"active": true, "inactive": true, "entered-in-error": true,
}

func (s *ClinicalSafetyService) CreateFlag(ctx context.Context, f *Flag) error {
	if f.CodeCode == "" {
		return fmt.Errorf("code_code is required")
	}
	if f.Status == "" {
		f.Status = "active"
	}
	if !validFlagStatuses[f.Status] {
		return fmt.Errorf("invalid flag status: %s", f.Status)
	}
	if err := s.flags.Create(ctx, f); err != nil {
		return err
	}
	f.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Flag", f.FHIRID, f.ToFHIR())
	}
	return nil
}

func (s *ClinicalSafetyService) GetFlag(ctx context.Context, id uuid.UUID) (*Flag, error) {
	return s.flags.GetByID(ctx, id)
}

func (s *ClinicalSafetyService) GetFlagByFHIRID(ctx context.Context, fhirID string) (*Flag, error) {
	return s.flags.GetByFHIRID(ctx, fhirID)
}

func (s *ClinicalSafetyService) UpdateFlag(ctx context.Context, f *Flag) error {
	if f.Status != "" && !validFlagStatuses[f.Status] {
		return fmt.Errorf("invalid flag status: %s", f.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Flag", f.FHIRID, f.VersionID, f.ToFHIR())
		if err == nil {
			f.VersionID = newVer
		}
	}
	return s.flags.Update(ctx, f)
}

func (s *ClinicalSafetyService) DeleteFlag(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		f, err := s.flags.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Flag", f.FHIRID, f.VersionID)
		}
	}
	return s.flags.Delete(ctx, id)
}

func (s *ClinicalSafetyService) ListFlags(ctx context.Context, limit, offset int) ([]*Flag, int, error) {
	return s.flags.List(ctx, limit, offset)
}

func (s *ClinicalSafetyService) SearchFlags(ctx context.Context, params map[string]string, limit, offset int) ([]*Flag, int, error) {
	return s.flags.Search(ctx, params, limit, offset)
}

// -- DetectedIssue --

var validDetectedIssueStatuses = map[string]bool{
	"registered": true, "preliminary": true, "final": true, "amended": true,
	"corrected": true, "cancelled": true, "entered-in-error": true, "unknown": true,
}

func (s *ClinicalSafetyService) CreateDetectedIssue(ctx context.Context, d *DetectedIssue) error {
	if d.Status == "" {
		d.Status = "final"
	}
	if !validDetectedIssueStatuses[d.Status] {
		return fmt.Errorf("invalid detected issue status: %s", d.Status)
	}
	if err := s.detectedIssues.Create(ctx, d); err != nil {
		return err
	}
	d.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DetectedIssue", d.FHIRID, d.ToFHIR())
	}
	return nil
}

func (s *ClinicalSafetyService) GetDetectedIssue(ctx context.Context, id uuid.UUID) (*DetectedIssue, error) {
	return s.detectedIssues.GetByID(ctx, id)
}

func (s *ClinicalSafetyService) GetDetectedIssueByFHIRID(ctx context.Context, fhirID string) (*DetectedIssue, error) {
	return s.detectedIssues.GetByFHIRID(ctx, fhirID)
}

func (s *ClinicalSafetyService) UpdateDetectedIssue(ctx context.Context, d *DetectedIssue) error {
	if d.Status != "" && !validDetectedIssueStatuses[d.Status] {
		return fmt.Errorf("invalid detected issue status: %s", d.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DetectedIssue", d.FHIRID, d.VersionID, d.ToFHIR())
		if err == nil {
			d.VersionID = newVer
		}
	}
	return s.detectedIssues.Update(ctx, d)
}

func (s *ClinicalSafetyService) DeleteDetectedIssue(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		d, err := s.detectedIssues.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DetectedIssue", d.FHIRID, d.VersionID)
		}
	}
	return s.detectedIssues.Delete(ctx, id)
}

func (s *ClinicalSafetyService) SearchDetectedIssues(ctx context.Context, params map[string]string, limit, offset int) ([]*DetectedIssue, int, error) {
	return s.detectedIssues.Search(ctx, params, limit, offset)
}

// -- AdverseEvent --

var validAdverseEventActuality = map[string]bool{
	"actual": true, "potential": true,
}

func (s *ClinicalSafetyService) CreateAdverseEvent(ctx context.Context, a *AdverseEvent) error {
	if a.Actuality == "" {
		a.Actuality = "actual"
	}
	if !validAdverseEventActuality[a.Actuality] {
		return fmt.Errorf("invalid adverse event actuality: %s", a.Actuality)
	}
	if err := s.adverseEvents.Create(ctx, a); err != nil {
		return err
	}
	a.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "AdverseEvent", a.FHIRID, a.ToFHIR())
	}
	return nil
}

func (s *ClinicalSafetyService) GetAdverseEvent(ctx context.Context, id uuid.UUID) (*AdverseEvent, error) {
	return s.adverseEvents.GetByID(ctx, id)
}

func (s *ClinicalSafetyService) GetAdverseEventByFHIRID(ctx context.Context, fhirID string) (*AdverseEvent, error) {
	return s.adverseEvents.GetByFHIRID(ctx, fhirID)
}

func (s *ClinicalSafetyService) UpdateAdverseEvent(ctx context.Context, a *AdverseEvent) error {
	if a.Actuality != "" && !validAdverseEventActuality[a.Actuality] {
		return fmt.Errorf("invalid adverse event actuality: %s", a.Actuality)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "AdverseEvent", a.FHIRID, a.VersionID, a.ToFHIR())
		if err == nil {
			a.VersionID = newVer
		}
	}
	return s.adverseEvents.Update(ctx, a)
}

func (s *ClinicalSafetyService) DeleteAdverseEvent(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		a, err := s.adverseEvents.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "AdverseEvent", a.FHIRID, a.VersionID)
		}
	}
	return s.adverseEvents.Delete(ctx, id)
}

func (s *ClinicalSafetyService) SearchAdverseEvents(ctx context.Context, params map[string]string, limit, offset int) ([]*AdverseEvent, int, error) {
	return s.adverseEvents.Search(ctx, params, limit, offset)
}

// -- ClinicalImpression --

var validCIStatuses = map[string]bool{
	"in-progress": true, "completed": true, "entered-in-error": true,
}

func (s *ClinicalSafetyService) CreateClinicalImpression(ctx context.Context, ci *ClinicalImpression) error {
	if ci.Status == "" {
		ci.Status = "in-progress"
	}
	if !validCIStatuses[ci.Status] {
		return fmt.Errorf("invalid clinical impression status: %s", ci.Status)
	}
	if err := s.clinicalImpressions.Create(ctx, ci); err != nil {
		return err
	}
	ci.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ClinicalImpression", ci.FHIRID, ci.ToFHIR())
	}
	return nil
}

func (s *ClinicalSafetyService) GetClinicalImpression(ctx context.Context, id uuid.UUID) (*ClinicalImpression, error) {
	return s.clinicalImpressions.GetByID(ctx, id)
}

func (s *ClinicalSafetyService) GetClinicalImpressionByFHIRID(ctx context.Context, fhirID string) (*ClinicalImpression, error) {
	return s.clinicalImpressions.GetByFHIRID(ctx, fhirID)
}

func (s *ClinicalSafetyService) UpdateClinicalImpression(ctx context.Context, ci *ClinicalImpression) error {
	if ci.Status != "" && !validCIStatuses[ci.Status] {
		return fmt.Errorf("invalid clinical impression status: %s", ci.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ClinicalImpression", ci.FHIRID, ci.VersionID, ci.ToFHIR())
		if err == nil {
			ci.VersionID = newVer
		}
	}
	return s.clinicalImpressions.Update(ctx, ci)
}

func (s *ClinicalSafetyService) DeleteClinicalImpression(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ci, err := s.clinicalImpressions.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ClinicalImpression", ci.FHIRID, ci.VersionID)
		}
	}
	return s.clinicalImpressions.Delete(ctx, id)
}

func (s *ClinicalSafetyService) SearchClinicalImpressions(ctx context.Context, params map[string]string, limit, offset int) ([]*ClinicalImpression, int, error) {
	return s.clinicalImpressions.Search(ctx, params, limit, offset)
}

// -- RiskAssessment --

var validRAStatuses = map[string]bool{
	"registered": true, "preliminary": true, "final": true, "amended": true,
	"corrected": true, "cancelled": true, "entered-in-error": true, "unknown": true,
}

func (s *ClinicalSafetyService) CreateRiskAssessment(ctx context.Context, ra *RiskAssessment) error {
	if ra.Status == "" {
		ra.Status = "final"
	}
	if !validRAStatuses[ra.Status] {
		return fmt.Errorf("invalid risk assessment status: %s", ra.Status)
	}
	if err := s.riskAssessments.Create(ctx, ra); err != nil {
		return err
	}
	ra.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "RiskAssessment", ra.FHIRID, ra.ToFHIR())
	}
	return nil
}

func (s *ClinicalSafetyService) GetRiskAssessment(ctx context.Context, id uuid.UUID) (*RiskAssessment, error) {
	return s.riskAssessments.GetByID(ctx, id)
}

func (s *ClinicalSafetyService) GetRiskAssessmentByFHIRID(ctx context.Context, fhirID string) (*RiskAssessment, error) {
	return s.riskAssessments.GetByFHIRID(ctx, fhirID)
}

func (s *ClinicalSafetyService) UpdateRiskAssessment(ctx context.Context, ra *RiskAssessment) error {
	if ra.Status != "" && !validRAStatuses[ra.Status] {
		return fmt.Errorf("invalid risk assessment status: %s", ra.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "RiskAssessment", ra.FHIRID, ra.VersionID, ra.ToFHIR())
		if err == nil {
			ra.VersionID = newVer
		}
	}
	return s.riskAssessments.Update(ctx, ra)
}

func (s *ClinicalSafetyService) DeleteRiskAssessment(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ra, err := s.riskAssessments.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "RiskAssessment", ra.FHIRID, ra.VersionID)
		}
	}
	return s.riskAssessments.Delete(ctx, id)
}

func (s *ClinicalSafetyService) SearchRiskAssessments(ctx context.Context, params map[string]string, limit, offset int) ([]*RiskAssessment, int, error) {
	return s.riskAssessments.Search(ctx, params, limit, offset)
}
