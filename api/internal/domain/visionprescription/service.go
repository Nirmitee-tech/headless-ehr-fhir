package visionprescription

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	prescriptions VisionPrescriptionRepository
	vt            *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(prescriptions VisionPrescriptionRepository) *Service {
	return &Service{prescriptions: prescriptions}
}

var validStatuses = map[string]bool{
	"active": true, "cancelled": true, "draft": true, "entered-in-error": true,
}

func (s *Service) CreateVisionPrescription(ctx context.Context, v *VisionPrescription) error {
	if v.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if v.Status == "" {
		v.Status = "active"
	}
	if !validStatuses[v.Status] {
		return fmt.Errorf("invalid status: %s", v.Status)
	}
	if err := s.prescriptions.Create(ctx, v); err != nil {
		return err
	}
	v.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "VisionPrescription", v.FHIRID, v.ToFHIR())
	}
	return nil
}

func (s *Service) GetVisionPrescription(ctx context.Context, id uuid.UUID) (*VisionPrescription, error) {
	return s.prescriptions.GetByID(ctx, id)
}

func (s *Service) GetVisionPrescriptionByFHIRID(ctx context.Context, fhirID string) (*VisionPrescription, error) {
	return s.prescriptions.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateVisionPrescription(ctx context.Context, v *VisionPrescription) error {
	if v.Status != "" && !validStatuses[v.Status] {
		return fmt.Errorf("invalid status: %s", v.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "VisionPrescription", v.FHIRID, v.VersionID, v.ToFHIR())
		if err == nil {
			v.VersionID = newVer
		}
	}
	return s.prescriptions.Update(ctx, v)
}

func (s *Service) DeleteVisionPrescription(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		v, err := s.prescriptions.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "VisionPrescription", v.FHIRID, v.VersionID)
		}
	}
	return s.prescriptions.Delete(ctx, id)
}

func (s *Service) ListVisionPrescriptions(ctx context.Context, limit, offset int) ([]*VisionPrescription, int, error) {
	return s.prescriptions.List(ctx, limit, offset)
}

func (s *Service) SearchVisionPrescriptions(ctx context.Context, params map[string]string, limit, offset int) ([]*VisionPrescription, int, error) {
	return s.prescriptions.Search(ctx, params, limit, offset)
}

func (s *Service) AddLensSpec(ctx context.Context, ls *VisionPrescriptionLensSpec) error {
	if err := ls.Validate(); err != nil {
		return err
	}
	return s.prescriptions.AddLensSpec(ctx, ls)
}

func (s *Service) GetLensSpecs(ctx context.Context, prescriptionID uuid.UUID) ([]*VisionPrescriptionLensSpec, error) {
	return s.prescriptions.GetLensSpecs(ctx, prescriptionID)
}
