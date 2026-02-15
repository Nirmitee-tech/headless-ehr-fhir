package diagnostics

import (
	"context"
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	serviceRequests   ServiceRequestRepository
	specimens         SpecimenRepository
	diagnosticReports DiagnosticReportRepository
	imagingStudies    ImagingStudyRepository
	statusHistory     OrderStatusHistoryRepository
	vt                *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(sr ServiceRequestRepository, sp SpecimenRepository, dr DiagnosticReportRepository, is ImagingStudyRepository, sh OrderStatusHistoryRepository) *Service {
	return &Service{serviceRequests: sr, specimens: sp, diagnosticReports: dr, imagingStudies: is, statusHistory: sh}
}

// -- Order Workflow State Machine --

// serviceRequestTransitions defines valid status transitions for ServiceRequest.
var serviceRequestTransitions = map[string][]string{
	"draft":            {"active", "on-hold", "revoked", "entered-in-error"},
	"active":           {"on-hold", "revoked", "completed", "entered-in-error"},
	"on-hold":          {"active", "revoked", "entered-in-error"},
	"completed":        {"entered-in-error"},
	"revoked":          {"entered-in-error"},
	"entered-in-error": {},
	"unknown":          {"draft", "active", "entered-in-error"},
}

// medicationRequestTransitions defines valid status transitions for MedicationRequest.
var medicationRequestTransitions = map[string][]string{
	"draft":            {"active", "on-hold", "cancelled", "entered-in-error"},
	"active":           {"on-hold", "completed", "stopped", "cancelled", "entered-in-error"},
	"on-hold":          {"active", "cancelled", "entered-in-error"},
	"completed":        {"entered-in-error"},
	"cancelled":        {"draft", "entered-in-error"},
	"stopped":          {"entered-in-error"},
	"entered-in-error": {},
	"unknown":          {"draft", "active", "entered-in-error"},
}

// ValidateTransition checks if a status transition is valid for the given resource type.
func ValidateTransition(resourceType, from, to string) error {
	var transitions map[string][]string
	switch resourceType {
	case "ServiceRequest":
		transitions = serviceRequestTransitions
	case "MedicationRequest":
		transitions = medicationRequestTransitions
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	allowed, ok := transitions[from]
	if !ok {
		return fmt.Errorf("unknown from-status: %s", from)
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition from %s to %s for %s", from, to, resourceType)
}

// RecordStatusChange validates the transition and records it in the status history.
func (s *Service) RecordStatusChange(ctx context.Context, resourceType string, resourceID uuid.UUID, from, to, changedBy, reason string) error {
	if err := ValidateTransition(resourceType, from, to); err != nil {
		return err
	}
	h := &OrderStatusHistory{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		FromStatus:   from,
		ToStatus:     to,
		ChangedBy:    changedBy,
		ChangedAt:    time.Now(),
	}
	if reason != "" {
		h.Reason = &reason
	}
	return s.statusHistory.Create(ctx, h)
}

// GetStatusHistory returns the status change history for a resource.
func (s *Service) GetStatusHistory(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*OrderStatusHistory, error) {
	return s.statusHistory.GetByResource(ctx, resourceType, resourceID)
}

// -- ServiceRequest --

var validSRStatuses = map[string]bool{
	"draft": true, "active": true, "on-hold": true, "revoked": true,
	"completed": true, "entered-in-error": true, "unknown": true,
}

func (s *Service) CreateServiceRequest(ctx context.Context, sr *ServiceRequest) error {
	if sr.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if sr.RequesterID == uuid.Nil {
		return fmt.Errorf("requester_id is required")
	}
	if sr.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	if sr.Status == "" {
		sr.Status = "draft"
	}
	if !validSRStatuses[sr.Status] {
		return fmt.Errorf("invalid status: %s", sr.Status)
	}
	if sr.Intent == "" {
		sr.Intent = "order"
	}
	if err := s.serviceRequests.Create(ctx, sr); err != nil {
		return err
	}
	sr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ServiceRequest", sr.FHIRID, sr.ToFHIR())
	}
	return nil
}

func (s *Service) GetServiceRequest(ctx context.Context, id uuid.UUID) (*ServiceRequest, error) {
	return s.serviceRequests.GetByID(ctx, id)
}

func (s *Service) GetServiceRequestByFHIRID(ctx context.Context, fhirID string) (*ServiceRequest, error) {
	return s.serviceRequests.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateServiceRequest(ctx context.Context, sr *ServiceRequest) error {
	if sr.Status != "" && !validSRStatuses[sr.Status] {
		return fmt.Errorf("invalid status: %s", sr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ServiceRequest", sr.FHIRID, sr.VersionID, sr.ToFHIR())
		if err == nil {
			sr.VersionID = newVer
		}
	}
	return s.serviceRequests.Update(ctx, sr)
}

func (s *Service) DeleteServiceRequest(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sr, err := s.serviceRequests.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ServiceRequest", sr.FHIRID, sr.VersionID)
		}
	}
	return s.serviceRequests.Delete(ctx, id)
}

func (s *Service) ListServiceRequestsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ServiceRequest, int, error) {
	return s.serviceRequests.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchServiceRequests(ctx context.Context, params map[string]string, limit, offset int) ([]*ServiceRequest, int, error) {
	return s.serviceRequests.Search(ctx, params, limit, offset)
}

// -- Specimen --

var validSpecimenStatuses = map[string]bool{
	"available": true, "unavailable": true, "unsatisfactory": true, "entered-in-error": true,
}

func (s *Service) CreateSpecimen(ctx context.Context, sp *Specimen) error {
	if sp.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if sp.Status == "" {
		sp.Status = "available"
	}
	if !validSpecimenStatuses[sp.Status] {
		return fmt.Errorf("invalid status: %s", sp.Status)
	}
	if err := s.specimens.Create(ctx, sp); err != nil {
		return err
	}
	sp.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Specimen", sp.FHIRID, sp.ToFHIR())
	}
	return nil
}

func (s *Service) GetSpecimen(ctx context.Context, id uuid.UUID) (*Specimen, error) {
	return s.specimens.GetByID(ctx, id)
}

func (s *Service) GetSpecimenByFHIRID(ctx context.Context, fhirID string) (*Specimen, error) {
	return s.specimens.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSpecimen(ctx context.Context, sp *Specimen) error {
	if sp.Status != "" && !validSpecimenStatuses[sp.Status] {
		return fmt.Errorf("invalid status: %s", sp.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Specimen", sp.FHIRID, sp.VersionID, sp.ToFHIR())
		if err == nil {
			sp.VersionID = newVer
		}
	}
	return s.specimens.Update(ctx, sp)
}

func (s *Service) DeleteSpecimen(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sp, err := s.specimens.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Specimen", sp.FHIRID, sp.VersionID)
		}
	}
	return s.specimens.Delete(ctx, id)
}

func (s *Service) ListSpecimensByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Specimen, int, error) {
	return s.specimens.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchSpecimens(ctx context.Context, params map[string]string, limit, offset int) ([]*Specimen, int, error) {
	return s.specimens.Search(ctx, params, limit, offset)
}

// -- DiagnosticReport --

var validDRStatuses = map[string]bool{
	"registered": true, "partial": true, "preliminary": true, "final": true,
	"amended": true, "corrected": true, "appended": true,
	"cancelled": true, "entered-in-error": true, "unknown": true,
}

func (s *Service) CreateDiagnosticReport(ctx context.Context, dr *DiagnosticReport) error {
	if dr.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if dr.CodeValue == "" {
		return fmt.Errorf("code_value is required")
	}
	if dr.Status == "" {
		dr.Status = "registered"
	}
	if !validDRStatuses[dr.Status] {
		return fmt.Errorf("invalid status: %s", dr.Status)
	}
	if err := s.diagnosticReports.Create(ctx, dr); err != nil {
		return err
	}
	dr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "DiagnosticReport", dr.FHIRID, dr.ToFHIR())
	}
	return nil
}

func (s *Service) GetDiagnosticReport(ctx context.Context, id uuid.UUID) (*DiagnosticReport, error) {
	return s.diagnosticReports.GetByID(ctx, id)
}

func (s *Service) GetDiagnosticReportByFHIRID(ctx context.Context, fhirID string) (*DiagnosticReport, error) {
	return s.diagnosticReports.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateDiagnosticReport(ctx context.Context, dr *DiagnosticReport) error {
	if dr.Status != "" && !validDRStatuses[dr.Status] {
		return fmt.Errorf("invalid status: %s", dr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "DiagnosticReport", dr.FHIRID, dr.VersionID, dr.ToFHIR())
		if err == nil {
			dr.VersionID = newVer
		}
	}
	return s.diagnosticReports.Update(ctx, dr)
}

func (s *Service) DeleteDiagnosticReport(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		dr, err := s.diagnosticReports.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "DiagnosticReport", dr.FHIRID, dr.VersionID)
		}
	}
	return s.diagnosticReports.Delete(ctx, id)
}

func (s *Service) ListDiagnosticReportsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*DiagnosticReport, int, error) {
	return s.diagnosticReports.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchDiagnosticReports(ctx context.Context, params map[string]string, limit, offset int) ([]*DiagnosticReport, int, error) {
	return s.diagnosticReports.Search(ctx, params, limit, offset)
}

func (s *Service) AddDiagnosticReportResult(ctx context.Context, reportID uuid.UUID, observationID uuid.UUID) error {
	if reportID == uuid.Nil {
		return fmt.Errorf("report_id is required")
	}
	if observationID == uuid.Nil {
		return fmt.Errorf("observation_id is required")
	}
	return s.diagnosticReports.AddResult(ctx, reportID, observationID)
}

func (s *Service) GetDiagnosticReportResults(ctx context.Context, reportID uuid.UUID) ([]uuid.UUID, error) {
	return s.diagnosticReports.GetResults(ctx, reportID)
}

func (s *Service) RemoveDiagnosticReportResult(ctx context.Context, reportID uuid.UUID, observationID uuid.UUID) error {
	return s.diagnosticReports.RemoveResult(ctx, reportID, observationID)
}

// -- ImagingStudy --

var validISStatuses = map[string]bool{
	"registered": true, "available": true, "cancelled": true,
	"entered-in-error": true, "unknown": true,
}

func (s *Service) CreateImagingStudy(ctx context.Context, is *ImagingStudy) error {
	if is.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if is.Status == "" {
		is.Status = "registered"
	}
	if !validISStatuses[is.Status] {
		return fmt.Errorf("invalid status: %s", is.Status)
	}
	if err := s.imagingStudies.Create(ctx, is); err != nil {
		return err
	}
	is.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "ImagingStudy", is.FHIRID, is.ToFHIR())
	}
	return nil
}

func (s *Service) GetImagingStudy(ctx context.Context, id uuid.UUID) (*ImagingStudy, error) {
	return s.imagingStudies.GetByID(ctx, id)
}

func (s *Service) GetImagingStudyByFHIRID(ctx context.Context, fhirID string) (*ImagingStudy, error) {
	return s.imagingStudies.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateImagingStudy(ctx context.Context, is *ImagingStudy) error {
	if is.Status != "" && !validISStatuses[is.Status] {
		return fmt.Errorf("invalid status: %s", is.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "ImagingStudy", is.FHIRID, is.VersionID, is.ToFHIR())
		if err == nil {
			is.VersionID = newVer
		}
	}
	return s.imagingStudies.Update(ctx, is)
}

func (s *Service) DeleteImagingStudy(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		is, err := s.imagingStudies.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "ImagingStudy", is.FHIRID, is.VersionID)
		}
	}
	return s.imagingStudies.Delete(ctx, id)
}

func (s *Service) ListImagingStudiesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImagingStudy, int, error) {
	return s.imagingStudies.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchImagingStudies(ctx context.Context, params map[string]string, limit, offset int) ([]*ImagingStudy, int, error) {
	return s.imagingStudies.Search(ctx, params, limit, offset)
}
