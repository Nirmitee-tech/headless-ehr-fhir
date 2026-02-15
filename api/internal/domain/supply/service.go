package supply

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	requests   SupplyRequestRepository
	deliveries SupplyDeliveryRepository
	vt         *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(requests SupplyRequestRepository, deliveries SupplyDeliveryRepository) *Service {
	return &Service{requests: requests, deliveries: deliveries}
}

var validSupplyRequestStatuses = map[string]bool{
	"draft": true, "active": true, "suspended": true, "cancelled": true,
	"completed": true, "entered-in-error": true, "unknown": true,
}

var validSupplyDeliveryStatuses = map[string]bool{
	"in-progress": true, "completed": true, "abandoned": true, "entered-in-error": true,
}

// ---- SupplyRequest ----

func (s *Service) CreateSupplyRequest(ctx context.Context, sr *SupplyRequest) error {
	if sr.ItemCode == "" {
		return fmt.Errorf("item_code is required")
	}
	if sr.QuantityValue <= 0 {
		return fmt.Errorf("quantity_value must be greater than 0")
	}
	if sr.Status == "" {
		sr.Status = "draft"
	}
	if !validSupplyRequestStatuses[sr.Status] {
		return fmt.Errorf("invalid status: %s", sr.Status)
	}
	if err := s.requests.Create(ctx, sr); err != nil {
		return err
	}
	sr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "SupplyRequest", sr.FHIRID, sr.ToFHIR())
	}
	return nil
}

func (s *Service) GetSupplyRequest(ctx context.Context, id uuid.UUID) (*SupplyRequest, error) {
	return s.requests.GetByID(ctx, id)
}

func (s *Service) GetSupplyRequestByFHIRID(ctx context.Context, fhirID string) (*SupplyRequest, error) {
	return s.requests.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSupplyRequest(ctx context.Context, sr *SupplyRequest) error {
	if sr.Status != "" && !validSupplyRequestStatuses[sr.Status] {
		return fmt.Errorf("invalid status: %s", sr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "SupplyRequest", sr.FHIRID, sr.VersionID, sr.ToFHIR())
		if err == nil {
			sr.VersionID = newVer
		}
	}
	return s.requests.Update(ctx, sr)
}

func (s *Service) DeleteSupplyRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sr, err := s.requests.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "SupplyRequest", sr.FHIRID, sr.VersionID)
		}
	}
	return s.requests.Delete(ctx, id)
}

func (s *Service) ListSupplyRequests(ctx context.Context, limit, offset int) ([]*SupplyRequest, int, error) {
	return s.requests.List(ctx, limit, offset)
}

func (s *Service) SearchSupplyRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*SupplyRequest, int, error) {
	return s.requests.Search(ctx, params, limit, offset)
}

// ---- SupplyDelivery ----

func (s *Service) CreateSupplyDelivery(ctx context.Context, sd *SupplyDelivery) error {
	if sd.Status == "" {
		sd.Status = "in-progress"
	}
	if !validSupplyDeliveryStatuses[sd.Status] {
		return fmt.Errorf("invalid status: %s", sd.Status)
	}
	if err := s.deliveries.Create(ctx, sd); err != nil {
		return err
	}
	sd.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "SupplyDelivery", sd.FHIRID, sd.ToFHIR())
	}
	return nil
}

func (s *Service) GetSupplyDelivery(ctx context.Context, id uuid.UUID) (*SupplyDelivery, error) {
	return s.deliveries.GetByID(ctx, id)
}

func (s *Service) GetSupplyDeliveryByFHIRID(ctx context.Context, fhirID string) (*SupplyDelivery, error) {
	return s.deliveries.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSupplyDelivery(ctx context.Context, sd *SupplyDelivery) error {
	if sd.Status != "" && !validSupplyDeliveryStatuses[sd.Status] {
		return fmt.Errorf("invalid status: %s", sd.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "SupplyDelivery", sd.FHIRID, sd.VersionID, sd.ToFHIR())
		if err == nil {
			sd.VersionID = newVer
		}
	}
	return s.deliveries.Update(ctx, sd)
}

func (s *Service) DeleteSupplyDelivery(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sd, err := s.deliveries.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "SupplyDelivery", sd.FHIRID, sd.VersionID)
		}
	}
	return s.deliveries.Delete(ctx, id)
}

func (s *Service) ListSupplyDeliveries(ctx context.Context, limit, offset int) ([]*SupplyDelivery, int, error) {
	return s.deliveries.List(ctx, limit, offset)
}

func (s *Service) SearchSupplyDeliveries(ctx context.Context, params map[string]string, limit, offset int) ([]*SupplyDelivery, int, error) {
	return s.deliveries.Search(ctx, params, limit, offset)
}
