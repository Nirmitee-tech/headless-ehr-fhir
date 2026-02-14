package oncology

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

func TestHandler_CreateCancerDiagnosis(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","diagnosis_date":"2025-06-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCancerDiagnosis(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateCancerDiagnosis_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCancerDiagnosis(c)
	if err == nil {
		t.Error("expected error for missing diagnosis_date")
	}
}

func TestHandler_GetCancerDiagnosis(t *testing.T) {
	h, e := newTestHandler()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	h.svc.CreateCancerDiagnosis(nil, d)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())

	err := h.GetCancerDiagnosis(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetCancerDiagnosis_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetCancerDiagnosis(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteCancerDiagnosis(t *testing.T) {
	h, e := newTestHandler()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	h.svc.CreateCancerDiagnosis(nil, d)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())

	err := h.DeleteCancerDiagnosis(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreateTreatmentProtocol(t *testing.T) {
	h, e := newTestHandler()
	body := `{"cancer_diagnosis_id":"` + uuid.New().String() + `","protocol_name":"FOLFOX"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTreatmentProtocol(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTreatmentProtocol_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"cancer_diagnosis_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTreatmentProtocol(c)
	if err == nil {
		t.Error("expected error for missing protocol_name")
	}
}

func TestHandler_AddProtocolDrug(t *testing.T) {
	h, e := newTestHandler()
	protoID := uuid.New()
	body := `{"drug_name":"Oxaliplatin"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(protoID.String())

	err := h.AddProtocolDrug(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateChemoCycle(t *testing.T) {
	h, e := newTestHandler()
	body := `{"protocol_id":"` + uuid.New().String() + `","cycle_number":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateChemoCycle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateChemoCycle_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"protocol_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateChemoCycle(c)
	if err == nil {
		t.Error("expected error for missing cycle_number")
	}
}

func TestHandler_AddChemoAdministration(t *testing.T) {
	h, e := newTestHandler()
	cycleID := uuid.New()
	body := `{"drug_name":"Cisplatin"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cycleID.String())

	err := h.AddChemoAdministration(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateRadiationTherapy(t *testing.T) {
	h, e := newTestHandler()
	body := `{"cancer_diagnosis_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateRadiationTherapy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddRadiationSession(t *testing.T) {
	h, e := newTestHandler()
	radID := uuid.New()
	body := `{"session_number":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(radID.String())

	err := h.AddRadiationSession(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTumorMarker(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","marker_name":"PSA"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTumorMarker(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTumorBoardReview(t *testing.T) {
	h, e := newTestHandler()
	body := `{"cancer_diagnosis_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTumorBoardReview(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateTumorBoardReview_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"cancer_diagnosis_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTumorBoardReview(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

// -- List/Update Tests for CancerDiagnosis --

func TestHandler_ListCancerDiagnoses(t *testing.T) {
	h, e := newTestHandler()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	h.svc.CreateCancerDiagnosis(nil, d)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListCancerDiagnoses(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateCancerDiagnosis(t *testing.T) {
	h, e := newTestHandler()
	d := &CancerDiagnosis{PatientID: uuid.New(), DiagnosisDate: time.Now()}
	h.svc.CreateCancerDiagnosis(nil, d)

	body := `{"patient_id":"` + d.PatientID.String() + `","diagnosis_date":"2025-07-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())

	err := h.UpdateCancerDiagnosis(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for TreatmentProtocol --

func TestHandler_GetTreatmentProtocol(t *testing.T) {
	h, e := newTestHandler()
	p := &TreatmentProtocol{CancerDiagnosisID: uuid.New(), ProtocolName: "FOLFOX"}
	h.svc.CreateTreatmentProtocol(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetTreatmentProtocol(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListTreatmentProtocols(t *testing.T) {
	h, e := newTestHandler()
	p := &TreatmentProtocol{CancerDiagnosisID: uuid.New(), ProtocolName: "FOLFOX"}
	h.svc.CreateTreatmentProtocol(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListTreatmentProtocols(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateTreatmentProtocol(t *testing.T) {
	h, e := newTestHandler()
	p := &TreatmentProtocol{CancerDiagnosisID: uuid.New(), ProtocolName: "FOLFOX"}
	h.svc.CreateTreatmentProtocol(nil, p)

	body := `{"cancer_diagnosis_id":"` + p.CancerDiagnosisID.String() + `","protocol_name":"FOLFIRI"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.UpdateTreatmentProtocol(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteTreatmentProtocol(t *testing.T) {
	h, e := newTestHandler()
	p := &TreatmentProtocol{CancerDiagnosisID: uuid.New(), ProtocolName: "FOLFOX"}
	h.svc.CreateTreatmentProtocol(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.DeleteTreatmentProtocol(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetProtocolDrugs(t *testing.T) {
	h, e := newTestHandler()
	protoID := uuid.New()
	drug := &TreatmentProtocolDrug{ProtocolID: protoID, DrugName: "Oxaliplatin"}
	h.svc.AddProtocolDrug(nil, drug)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(protoID.String())

	err := h.GetProtocolDrugs(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for ChemoCycle --

func TestHandler_GetChemoCycle(t *testing.T) {
	h, e := newTestHandler()
	cycle := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 1}
	h.svc.CreateChemoCycle(nil, cycle)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cycle.ID.String())

	err := h.GetChemoCycle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListChemoCycles(t *testing.T) {
	h, e := newTestHandler()
	cycle := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 1}
	h.svc.CreateChemoCycle(nil, cycle)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListChemoCycles(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateChemoCycle(t *testing.T) {
	h, e := newTestHandler()
	cycle := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 1}
	h.svc.CreateChemoCycle(nil, cycle)

	body := `{"protocol_id":"` + cycle.ProtocolID.String() + `","cycle_number":2}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cycle.ID.String())

	err := h.UpdateChemoCycle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteChemoCycle(t *testing.T) {
	h, e := newTestHandler()
	cycle := &ChemoCycle{ProtocolID: uuid.New(), CycleNumber: 1}
	h.svc.CreateChemoCycle(nil, cycle)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cycle.ID.String())

	err := h.DeleteChemoCycle(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetChemoAdministrations(t *testing.T) {
	h, e := newTestHandler()
	cycleID := uuid.New()
	admin := &ChemoAdministration{CycleID: cycleID, DrugName: "Cisplatin"}
	h.svc.AddChemoAdministration(nil, admin)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(cycleID.String())

	err := h.GetChemoAdministrations(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for RadiationTherapy --

func TestHandler_GetRadiationTherapy(t *testing.T) {
	h, e := newTestHandler()
	r := &RadiationTherapy{CancerDiagnosisID: uuid.New()}
	h.svc.CreateRadiationTherapy(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.GetRadiationTherapy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListRadiationTherapies(t *testing.T) {
	h, e := newTestHandler()
	r := &RadiationTherapy{CancerDiagnosisID: uuid.New()}
	h.svc.CreateRadiationTherapy(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListRadiationTherapies(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateRadiationTherapy(t *testing.T) {
	h, e := newTestHandler()
	r := &RadiationTherapy{CancerDiagnosisID: uuid.New()}
	h.svc.CreateRadiationTherapy(nil, r)

	body := `{"cancer_diagnosis_id":"` + r.CancerDiagnosisID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.UpdateRadiationTherapy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteRadiationTherapy(t *testing.T) {
	h, e := newTestHandler()
	r := &RadiationTherapy{CancerDiagnosisID: uuid.New()}
	h.svc.CreateRadiationTherapy(nil, r)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.DeleteRadiationTherapy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetRadiationSessions(t *testing.T) {
	h, e := newTestHandler()
	radID := uuid.New()
	sess := &RadiationSession{RadiationTherapyID: radID, SessionNumber: 1}
	h.svc.AddRadiationSession(nil, sess)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(radID.String())

	err := h.GetRadiationSessions(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for TumorMarker --

func TestHandler_GetTumorMarker(t *testing.T) {
	h, e := newTestHandler()
	m := &TumorMarker{PatientID: uuid.New(), MarkerName: "PSA"}
	h.svc.CreateTumorMarker(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.GetTumorMarker(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListTumorMarkers(t *testing.T) {
	h, e := newTestHandler()
	m := &TumorMarker{PatientID: uuid.New(), MarkerName: "PSA"}
	h.svc.CreateTumorMarker(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListTumorMarkers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateTumorMarker(t *testing.T) {
	h, e := newTestHandler()
	m := &TumorMarker{PatientID: uuid.New(), MarkerName: "PSA"}
	h.svc.CreateTumorMarker(nil, m)

	body := `{"patient_id":"` + m.PatientID.String() + `","marker_name":"CEA"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.UpdateTumorMarker(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteTumorMarker(t *testing.T) {
	h, e := newTestHandler()
	m := &TumorMarker{PatientID: uuid.New(), MarkerName: "PSA"}
	h.svc.CreateTumorMarker(nil, m)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.DeleteTumorMarker(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for TumorBoardReview --

func TestHandler_GetTumorBoardReview(t *testing.T) {
	h, e := newTestHandler()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateTumorBoardReview(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.GetTumorBoardReview(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListTumorBoardReviews(t *testing.T) {
	h, e := newTestHandler()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateTumorBoardReview(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListTumorBoardReviews(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateTumorBoardReview(t *testing.T) {
	h, e := newTestHandler()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateTumorBoardReview(nil, r)

	body := `{"cancer_diagnosis_id":"` + r.CancerDiagnosisID.String() + `","patient_id":"` + r.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.UpdateTumorBoardReview(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteTumorBoardReview(t *testing.T) {
	h, e := newTestHandler()
	r := &TumorBoardReview{CancerDiagnosisID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreateTumorBoardReview(nil, r)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.DeleteTumorBoardReview(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- RegisterRoutes --

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	h.RegisterRoutes(api)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/cancer-diagnoses",
		"GET:/api/v1/cancer-diagnoses",
		"GET:/api/v1/cancer-diagnoses/:id",
		"PUT:/api/v1/cancer-diagnoses/:id",
		"DELETE:/api/v1/cancer-diagnoses/:id",
		"POST:/api/v1/treatment-protocols",
		"GET:/api/v1/treatment-protocols",
		"GET:/api/v1/treatment-protocols/:id",
		"GET:/api/v1/treatment-protocols/:id/drugs",
		"POST:/api/v1/chemo-cycles",
		"GET:/api/v1/chemo-cycles",
		"GET:/api/v1/chemo-cycles/:id",
		"GET:/api/v1/chemo-cycles/:id/administrations",
		"POST:/api/v1/radiation-therapies",
		"GET:/api/v1/radiation-therapies",
		"GET:/api/v1/radiation-therapies/:id",
		"GET:/api/v1/radiation-therapies/:id/sessions",
		"POST:/api/v1/tumor-markers",
		"GET:/api/v1/tumor-markers",
		"GET:/api/v1/tumor-markers/:id",
		"POST:/api/v1/tumor-board-reviews",
		"GET:/api/v1/tumor-board-reviews",
		"GET:/api/v1/tumor-board-reviews/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
