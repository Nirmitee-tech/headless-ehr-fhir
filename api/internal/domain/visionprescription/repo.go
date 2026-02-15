package visionprescription

import (
	"context"

	"github.com/google/uuid"
)

type VisionPrescriptionRepository interface {
	Create(ctx context.Context, v *VisionPrescription) error
	GetByID(ctx context.Context, id uuid.UUID) (*VisionPrescription, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*VisionPrescription, error)
	Update(ctx context.Context, v *VisionPrescription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*VisionPrescription, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*VisionPrescription, int, error)
	AddLensSpec(ctx context.Context, ls *VisionPrescriptionLensSpec) error
	GetLensSpecs(ctx context.Context, prescriptionID uuid.UUID) ([]*VisionPrescriptionLensSpec, error)
}
