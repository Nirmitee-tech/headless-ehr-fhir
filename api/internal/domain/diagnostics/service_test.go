package diagnostics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockSRRepo struct {
	srs map[uuid.UUID]*ServiceRequest
}

func newMockSRRepo() *mockSRRepo {
	return &mockSRRepo{srs: make(map[uuid.UUID]*ServiceRequest)}
}

func (m *mockSRRepo) Create(_ context.Context, sr *ServiceRequest) error {
	sr.ID = uuid.New()
	if sr.FHIRID == "" {
		sr.FHIRID = sr.ID.String()
	}
	sr.CreatedAt = time.Now()
	sr.UpdatedAt = time.Now()
	m.srs[sr.ID] = sr
	return nil
}

func (m *mockSRRepo) GetByID(_ context.Context, id uuid.UUID) (*ServiceRequest, error) {
	sr, ok := m.srs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return sr, nil
}

func (m *mockSRRepo) GetByFHIRID(_ context.Context, fhirID string) (*ServiceRequest, error) {
	for _, sr := range m.srs {
		if sr.FHIRID == fhirID {
			return sr, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockSRRepo) Update(_ context.Context, sr *ServiceRequest) error {
	m.srs[sr.ID] = sr
	return nil
}

func (m *mockSRRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.srs, id)
	return nil
}

func (m *mockSRRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ServiceRequest, int, error) {
	var result []*ServiceRequest
	for _, sr := range m.srs {
		if sr.PatientID == patientID {
			result = append(result, sr)
		}
	}
	return result, len(result), nil
}

func (m *mockSRRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*ServiceRequest, int, error) {
	var result []*ServiceRequest
	for _, sr := range m.srs {
		result = append(result, sr)
	}
	return result, len(result), nil
}

type mockSpecimenRepo struct {
	specs map[uuid.UUID]*Specimen
}

func newMockSpecimenRepo() *mockSpecimenRepo {
	return &mockSpecimenRepo{specs: make(map[uuid.UUID]*Specimen)}
}

func (m *mockSpecimenRepo) Create(_ context.Context, sp *Specimen) error {
	sp.ID = uuid.New()
	if sp.FHIRID == "" {
		sp.FHIRID = sp.ID.String()
	}
	sp.CreatedAt = time.Now()
	sp.UpdatedAt = time.Now()
	m.specs[sp.ID] = sp
	return nil
}

func (m *mockSpecimenRepo) GetByID(_ context.Context, id uuid.UUID) (*Specimen, error) {
	sp, ok := m.specs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return sp, nil
}

func (m *mockSpecimenRepo) GetByFHIRID(_ context.Context, fhirID string) (*Specimen, error) {
	for _, sp := range m.specs {
		if sp.FHIRID == fhirID {
			return sp, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockSpecimenRepo) Update(_ context.Context, sp *Specimen) error {
	m.specs[sp.ID] = sp
	return nil
}

func (m *mockSpecimenRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.specs, id)
	return nil
}

func (m *mockSpecimenRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Specimen, int, error) {
	var result []*Specimen
	for _, sp := range m.specs {
		if sp.PatientID == patientID {
			result = append(result, sp)
		}
	}
	return result, len(result), nil
}

func (m *mockSpecimenRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*Specimen, int, error) {
	var result []*Specimen
	for _, sp := range m.specs {
		result = append(result, sp)
	}
	return result, len(result), nil
}

type mockDRRepo struct {
	reports map[uuid.UUID]*DiagnosticReport
	results map[uuid.UUID][]uuid.UUID
}

func newMockDRRepo() *mockDRRepo {
	return &mockDRRepo{
		reports: make(map[uuid.UUID]*DiagnosticReport),
		results: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockDRRepo) Create(_ context.Context, dr *DiagnosticReport) error {
	dr.ID = uuid.New()
	if dr.FHIRID == "" {
		dr.FHIRID = dr.ID.String()
	}
	dr.CreatedAt = time.Now()
	dr.UpdatedAt = time.Now()
	m.reports[dr.ID] = dr
	return nil
}

func (m *mockDRRepo) GetByID(_ context.Context, id uuid.UUID) (*DiagnosticReport, error) {
	dr, ok := m.reports[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return dr, nil
}

func (m *mockDRRepo) GetByFHIRID(_ context.Context, fhirID string) (*DiagnosticReport, error) {
	for _, dr := range m.reports {
		if dr.FHIRID == fhirID {
			return dr, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDRRepo) Update(_ context.Context, dr *DiagnosticReport) error {
	m.reports[dr.ID] = dr
	return nil
}

func (m *mockDRRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.reports, id)
	return nil
}

func (m *mockDRRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*DiagnosticReport, int, error) {
	var result []*DiagnosticReport
	for _, dr := range m.reports {
		if dr.PatientID == patientID {
			result = append(result, dr)
		}
	}
	return result, len(result), nil
}

func (m *mockDRRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*DiagnosticReport, int, error) {
	var result []*DiagnosticReport
	for _, dr := range m.reports {
		result = append(result, dr)
	}
	return result, len(result), nil
}

func (m *mockDRRepo) AddResult(_ context.Context, reportID uuid.UUID, observationID uuid.UUID) error {
	m.results[reportID] = append(m.results[reportID], observationID)
	return nil
}

func (m *mockDRRepo) GetResults(_ context.Context, reportID uuid.UUID) ([]uuid.UUID, error) {
	return m.results[reportID], nil
}

func (m *mockDRRepo) RemoveResult(_ context.Context, reportID uuid.UUID, observationID uuid.UUID) error {
	ids := m.results[reportID]
	for i, id := range ids {
		if id == observationID {
			m.results[reportID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	return nil
}

type mockISRepo struct {
	studies map[uuid.UUID]*ImagingStudy
}

func newMockISRepo() *mockISRepo {
	return &mockISRepo{studies: make(map[uuid.UUID]*ImagingStudy)}
}

func (m *mockISRepo) Create(_ context.Context, is *ImagingStudy) error {
	is.ID = uuid.New()
	if is.FHIRID == "" {
		is.FHIRID = is.ID.String()
	}
	is.CreatedAt = time.Now()
	is.UpdatedAt = time.Now()
	m.studies[is.ID] = is
	return nil
}

func (m *mockISRepo) GetByID(_ context.Context, id uuid.UUID) (*ImagingStudy, error) {
	is, ok := m.studies[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return is, nil
}

func (m *mockISRepo) GetByFHIRID(_ context.Context, fhirID string) (*ImagingStudy, error) {
	for _, is := range m.studies {
		if is.FHIRID == fhirID {
			return is, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockISRepo) Update(_ context.Context, is *ImagingStudy) error {
	m.studies[is.ID] = is
	return nil
}

func (m *mockISRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.studies, id)
	return nil
}

func (m *mockISRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ImagingStudy, int, error) {
	var result []*ImagingStudy
	for _, is := range m.studies {
		if is.PatientID == patientID {
			result = append(result, is)
		}
	}
	return result, len(result), nil
}

func (m *mockISRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*ImagingStudy, int, error) {
	var result []*ImagingStudy
	for _, is := range m.studies {
		result = append(result, is)
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockSRRepo(), newMockSpecimenRepo(), newMockDRRepo(), newMockISRepo())
}

func TestCreateServiceRequest(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	err := svc.CreateServiceRequest(context.Background(), sr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sr.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", sr.Status)
	}
	if sr.Intent != "order" {
		t.Errorf("expected default intent 'order', got %s", sr.Intent)
	}
}

func TestCreateServiceRequest_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{RequesterID: uuid.New(), CodeValue: "CBC"}
	err := svc.CreateServiceRequest(context.Background(), sr)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateServiceRequest_RequesterIDRequired(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), CodeValue: "CBC"}
	err := svc.CreateServiceRequest(context.Background(), sr)
	if err == nil {
		t.Error("expected error for missing requester_id")
	}
}

func TestCreateServiceRequest_CodeValueRequired(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New()}
	err := svc.CreateServiceRequest(context.Background(), sr)
	if err == nil {
		t.Error("expected error for missing code_value")
	}
}

func TestGetServiceRequest(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	svc.CreateServiceRequest(context.Background(), sr)

	fetched, err := svc.GetServiceRequest(context.Background(), sr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.CodeValue != "CBC" {
		t.Errorf("expected 'CBC', got %s", fetched.CodeValue)
	}
}

func TestDeleteServiceRequest(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	svc.CreateServiceRequest(context.Background(), sr)
	err := svc.DeleteServiceRequest(context.Background(), sr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetServiceRequest(context.Background(), sr.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCreateSpecimen(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{PatientID: uuid.New()}
	err := svc.CreateSpecimen(context.Background(), sp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sp.Status != "available" {
		t.Errorf("expected default status 'available', got %s", sp.Status)
	}
}

func TestCreateSpecimen_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{}
	err := svc.CreateSpecimen(context.Background(), sp)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateDiagnosticReport(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	err := svc.CreateDiagnosticReport(context.Background(), dr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dr.Status != "registered" {
		t.Errorf("expected default status 'registered', got %s", dr.Status)
	}
}

func TestCreateDiagnosticReport_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{CodeValue: "CBC"}
	err := svc.CreateDiagnosticReport(context.Background(), dr)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateDiagnosticReport_CodeValueRequired(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New()}
	err := svc.CreateDiagnosticReport(context.Background(), dr)
	if err == nil {
		t.Error("expected error for missing code_value")
	}
}

func TestDiagnosticReportResults(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)

	obsID := uuid.New()
	err := svc.AddDiagnosticReportResult(context.Background(), dr.ID, obsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := svc.GetDiagnosticReportResults(context.Background(), dr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCreateImagingStudy(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{PatientID: uuid.New()}
	err := svc.CreateImagingStudy(context.Background(), is)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if is.Status != "registered" {
		t.Errorf("expected default status 'registered', got %s", is.Status)
	}
}

func TestCreateImagingStudy_PatientIDRequired(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{}
	err := svc.CreateImagingStudy(context.Background(), is)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestGetServiceRequestByFHIRID(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	svc.CreateServiceRequest(context.Background(), sr)

	fetched, err := svc.GetServiceRequestByFHIRID(context.Background(), sr.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != sr.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetServiceRequestByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetServiceRequestByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateServiceRequest(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	svc.CreateServiceRequest(context.Background(), sr)

	sr.Status = "active"
	err := svc.UpdateServiceRequest(context.Background(), sr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateServiceRequest_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sr := &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"}
	svc.CreateServiceRequest(context.Background(), sr)

	sr.Status = "bogus"
	err := svc.UpdateServiceRequest(context.Background(), sr)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListServiceRequestsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateServiceRequest(context.Background(), &ServiceRequest{PatientID: patientID, RequesterID: uuid.New(), CodeValue: "CBC"})
	svc.CreateServiceRequest(context.Background(), &ServiceRequest{PatientID: patientID, RequesterID: uuid.New(), CodeValue: "BMP"})
	svc.CreateServiceRequest(context.Background(), &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CMP"})

	result, total, err := svc.ListServiceRequestsByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestSearchServiceRequests(t *testing.T) {
	svc := newTestService()
	svc.CreateServiceRequest(context.Background(), &ServiceRequest{PatientID: uuid.New(), RequesterID: uuid.New(), CodeValue: "CBC"})

	result, total, err := svc.SearchServiceRequests(context.Background(), map[string]string{"code": "CBC"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(result) < 1 {
		t.Error("expected results")
	}
}

func TestGetSpecimen(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{PatientID: uuid.New()}
	svc.CreateSpecimen(context.Background(), sp)

	fetched, err := svc.GetSpecimen(context.Background(), sp.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != sp.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetSpecimen_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSpecimen(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetSpecimenByFHIRID(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{PatientID: uuid.New()}
	svc.CreateSpecimen(context.Background(), sp)

	fetched, err := svc.GetSpecimenByFHIRID(context.Background(), sp.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != sp.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetSpecimenByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSpecimenByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateSpecimen(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{PatientID: uuid.New()}
	svc.CreateSpecimen(context.Background(), sp)

	sp.Status = "unavailable"
	err := svc.UpdateSpecimen(context.Background(), sp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateSpecimen_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{PatientID: uuid.New()}
	svc.CreateSpecimen(context.Background(), sp)

	sp.Status = "bogus"
	err := svc.UpdateSpecimen(context.Background(), sp)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteSpecimen(t *testing.T) {
	svc := newTestService()
	sp := &Specimen{PatientID: uuid.New()}
	svc.CreateSpecimen(context.Background(), sp)
	err := svc.DeleteSpecimen(context.Background(), sp.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetSpecimen(context.Background(), sp.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListSpecimensByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateSpecimen(context.Background(), &Specimen{PatientID: patientID})
	svc.CreateSpecimen(context.Background(), &Specimen{PatientID: uuid.New()})

	result, total, err := svc.ListSpecimensByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestSearchSpecimens(t *testing.T) {
	svc := newTestService()
	svc.CreateSpecimen(context.Background(), &Specimen{PatientID: uuid.New()})

	result, total, err := svc.SearchSpecimens(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(result) < 1 {
		t.Error("expected results")
	}
}

func TestGetDiagnosticReport(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)

	fetched, err := svc.GetDiagnosticReport(context.Background(), dr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != dr.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetDiagnosticReport_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetDiagnosticReport(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetDiagnosticReportByFHIRID(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)

	fetched, err := svc.GetDiagnosticReportByFHIRID(context.Background(), dr.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != dr.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetDiagnosticReportByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetDiagnosticReportByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateDiagnosticReport(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)

	dr.Status = "final"
	err := svc.UpdateDiagnosticReport(context.Background(), dr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDiagnosticReport_InvalidStatus(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)

	dr.Status = "bogus"
	err := svc.UpdateDiagnosticReport(context.Background(), dr)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteDiagnosticReport(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)
	err := svc.DeleteDiagnosticReport(context.Background(), dr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetDiagnosticReport(context.Background(), dr.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListDiagnosticReportsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateDiagnosticReport(context.Background(), &DiagnosticReport{PatientID: patientID, CodeValue: "CBC"})
	svc.CreateDiagnosticReport(context.Background(), &DiagnosticReport{PatientID: uuid.New(), CodeValue: "BMP"})

	result, total, err := svc.ListDiagnosticReportsByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestSearchDiagnosticReports(t *testing.T) {
	svc := newTestService()
	svc.CreateDiagnosticReport(context.Background(), &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"})

	result, total, err := svc.SearchDiagnosticReports(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(result) < 1 {
		t.Error("expected results")
	}
}

func TestRemoveDiagnosticReportResult(t *testing.T) {
	svc := newTestService()
	dr := &DiagnosticReport{PatientID: uuid.New(), CodeValue: "CBC"}
	svc.CreateDiagnosticReport(context.Background(), dr)

	obsID := uuid.New()
	svc.AddDiagnosticReportResult(context.Background(), dr.ID, obsID)

	err := svc.RemoveDiagnosticReportResult(context.Background(), dr.ID, obsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, _ := svc.GetDiagnosticReportResults(context.Background(), dr.ID)
	if len(results) != 0 {
		t.Errorf("expected 0 results after removal, got %d", len(results))
	}
}

func TestGetImagingStudy(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{PatientID: uuid.New()}
	svc.CreateImagingStudy(context.Background(), is)

	fetched, err := svc.GetImagingStudy(context.Background(), is.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != is.ID {
		t.Error("unexpected ID mismatch")
	}
}

func TestGetImagingStudy_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetImagingStudy(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGetImagingStudyByFHIRID(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{PatientID: uuid.New()}
	svc.CreateImagingStudy(context.Background(), is)

	fetched, err := svc.GetImagingStudyByFHIRID(context.Background(), is.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != is.ID {
		t.Errorf("expected same ID")
	}
}

func TestGetImagingStudyByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetImagingStudyByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateImagingStudy(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{PatientID: uuid.New()}
	svc.CreateImagingStudy(context.Background(), is)

	is.Status = "available"
	err := svc.UpdateImagingStudy(context.Background(), is)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateImagingStudy_InvalidStatus(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{PatientID: uuid.New()}
	svc.CreateImagingStudy(context.Background(), is)

	is.Status = "bogus"
	err := svc.UpdateImagingStudy(context.Background(), is)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteImagingStudy(t *testing.T) {
	svc := newTestService()
	is := &ImagingStudy{PatientID: uuid.New()}
	svc.CreateImagingStudy(context.Background(), is)
	err := svc.DeleteImagingStudy(context.Background(), is.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetImagingStudy(context.Background(), is.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListImagingStudiesByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateImagingStudy(context.Background(), &ImagingStudy{PatientID: patientID})
	svc.CreateImagingStudy(context.Background(), &ImagingStudy{PatientID: uuid.New()})

	result, total, err := svc.ListImagingStudiesByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestSearchImagingStudies(t *testing.T) {
	svc := newTestService()
	svc.CreateImagingStudy(context.Background(), &ImagingStudy{PatientID: uuid.New()})

	result, total, err := svc.SearchImagingStudies(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(result) < 1 {
		t.Error("expected results")
	}
}

func TestServiceRequestToFHIR(t *testing.T) {
	sr := &ServiceRequest{
		FHIRID:      "sr-123",
		Status:      "active",
		Intent:      "order",
		PatientID:   uuid.New(),
		RequesterID: uuid.New(),
		CodeValue:   "CBC",
		CodeDisplay: "Complete Blood Count",
		UpdatedAt:   time.Now(),
	}
	fhirRes := sr.ToFHIR()
	if fhirRes["resourceType"] != "ServiceRequest" {
		t.Errorf("expected ServiceRequest, got %v", fhirRes["resourceType"])
	}
	if fhirRes["id"] != "sr-123" {
		t.Errorf("expected sr-123, got %v", fhirRes["id"])
	}
}

func TestDiagnosticReportToFHIR(t *testing.T) {
	dr := &DiagnosticReport{
		FHIRID:      "dr-123",
		Status:      "final",
		PatientID:   uuid.New(),
		CodeValue:   "CBC",
		CodeDisplay: "Complete Blood Count",
		UpdatedAt:   time.Now(),
	}
	fhirRes := dr.ToFHIR()
	if fhirRes["resourceType"] != "DiagnosticReport" {
		t.Errorf("expected DiagnosticReport, got %v", fhirRes["resourceType"])
	}
}

func TestSpecimenToFHIR(t *testing.T) {
	sp := &Specimen{
		FHIRID:    "sp-123",
		Status:    "available",
		PatientID: uuid.New(),
		UpdatedAt: time.Now(),
	}
	fhirRes := sp.ToFHIR()
	if fhirRes["resourceType"] != "Specimen" {
		t.Errorf("expected Specimen, got %v", fhirRes["resourceType"])
	}
}

func TestImagingStudyToFHIR(t *testing.T) {
	is := &ImagingStudy{
		FHIRID:    "is-123",
		Status:    "available",
		PatientID: uuid.New(),
		UpdatedAt: time.Now(),
	}
	fhirRes := is.ToFHIR()
	if fhirRes["resourceType"] != "ImagingStudy" {
		t.Errorf("expected ImagingStudy, got %v", fhirRes["resourceType"])
	}
}
