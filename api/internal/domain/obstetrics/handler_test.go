package obstetrics

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

func TestHandler_CreatePregnancy(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePregnancy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePregnancy_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePregnancy(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetPregnancy(t *testing.T) {
	h, e := newTestHandler()
	p := &Pregnancy{PatientID: uuid.New()}
	h.svc.CreatePregnancy(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetPregnancy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetPregnancy_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetPregnancy(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeletePregnancy(t *testing.T) {
	h, e := newTestHandler()
	p := &Pregnancy{PatientID: uuid.New()}
	h.svc.CreatePregnancy(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.DeletePregnancy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_CreatePrenatalVisit(t *testing.T) {
	h, e := newTestHandler()
	pregID := uuid.New()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pregID.String())

	err := h.CreatePrenatalVisit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateLaborRecord(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pregnancy_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateLaborRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateLaborRecord_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateLaborRecord(c)
	if err == nil {
		t.Error("expected error for missing pregnancy_id")
	}
}

func TestHandler_AddCervicalExam(t *testing.T) {
	h, e := newTestHandler()
	laborID := uuid.New()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(laborID.String())

	err := h.AddCervicalExam(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_AddFetalMonitoring(t *testing.T) {
	h, e := newTestHandler()
	laborID := uuid.New()
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(laborID.String())

	err := h.AddFetalMonitoring(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateDelivery(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pregnancy_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `","delivery_datetime":"2025-06-01T10:00:00Z","delivery_method":"vaginal","delivering_provider_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDelivery(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateDelivery_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pregnancy_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateDelivery(c)
	if err == nil {
		t.Error("expected error for missing required fields")
	}
}

func TestHandler_CreateNewborn(t *testing.T) {
	h, e := newTestHandler()
	body := `{"delivery_id":"` + uuid.New().String() + `","birth_datetime":"2025-06-01T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateNewborn(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateNewborn_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"delivery_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateNewborn(c)
	if err == nil {
		t.Error("expected error for missing birth_datetime")
	}
}

func TestHandler_CreatePostpartum(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pregnancy_id":"` + uuid.New().String() + `","patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePostpartum(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePostpartum_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pregnancy_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePostpartum(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetDelivery(t *testing.T) {
	h, e := newTestHandler()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	h.svc.CreateDelivery(nil, d)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())

	err := h.GetDelivery(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetNewborn(t *testing.T) {
	h, e := newTestHandler()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	h.svc.CreateNewborn(nil, n)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(n.ID.String())

	err := h.GetNewborn(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- List/Update Tests for Pregnancy --

func TestHandler_ListPregnancies(t *testing.T) {
	h, e := newTestHandler()
	p := &Pregnancy{PatientID: uuid.New()}
	h.svc.CreatePregnancy(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListPregnancies(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePregnancy(t *testing.T) {
	h, e := newTestHandler()
	p := &Pregnancy{PatientID: uuid.New()}
	h.svc.CreatePregnancy(nil, p)

	body := `{"patient_id":"` + p.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.UpdatePregnancy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for PrenatalVisit --

func TestHandler_GetPrenatalVisit(t *testing.T) {
	h, e := newTestHandler()
	pregID := uuid.New()
	v := &PrenatalVisit{PregnancyID: pregID}
	h.svc.CreatePrenatalVisit(nil, v)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(v.ID.String())

	err := h.GetPrenatalVisit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListPrenatalVisits(t *testing.T) {
	h, e := newTestHandler()
	pregID := uuid.New()
	v := &PrenatalVisit{PregnancyID: pregID}
	h.svc.CreatePrenatalVisit(nil, v)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(pregID.String())

	err := h.ListPrenatalVisits(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePrenatalVisit(t *testing.T) {
	h, e := newTestHandler()
	v := &PrenatalVisit{PregnancyID: uuid.New()}
	h.svc.CreatePrenatalVisit(nil, v)

	body := `{}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(v.ID.String())

	err := h.UpdatePrenatalVisit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePrenatalVisit(t *testing.T) {
	h, e := newTestHandler()
	v := &PrenatalVisit{PregnancyID: uuid.New()}
	h.svc.CreatePrenatalVisit(nil, v)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(v.ID.String())

	err := h.DeletePrenatalVisit(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for LaborRecord --

func TestHandler_GetLaborRecord(t *testing.T) {
	h, e := newTestHandler()
	l := &LaborRecord{PregnancyID: uuid.New()}
	h.svc.CreateLaborRecord(nil, l)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.GetLaborRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListLaborRecords(t *testing.T) {
	h, e := newTestHandler()
	l := &LaborRecord{PregnancyID: uuid.New()}
	h.svc.CreateLaborRecord(nil, l)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListLaborRecords(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateLaborRecord(t *testing.T) {
	h, e := newTestHandler()
	l := &LaborRecord{PregnancyID: uuid.New()}
	h.svc.CreateLaborRecord(nil, l)

	body := `{"pregnancy_id":"` + l.PregnancyID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.UpdateLaborRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteLaborRecord(t *testing.T) {
	h, e := newTestHandler()
	l := &LaborRecord{PregnancyID: uuid.New()}
	h.svc.CreateLaborRecord(nil, l)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.DeleteLaborRecord(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- Sub-resource Tests for LaborRecord --

func TestHandler_GetCervicalExams(t *testing.T) {
	h, e := newTestHandler()
	laborID := uuid.New()
	exam := &LaborCervicalExam{LaborRecordID: laborID}
	h.svc.AddCervicalExam(nil, exam)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(laborID.String())

	err := h.GetCervicalExams(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetFetalMonitoring(t *testing.T) {
	h, e := newTestHandler()
	laborID := uuid.New()
	fm := &FetalMonitoring{LaborRecordID: laborID}
	h.svc.AddFetalMonitoring(nil, fm)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(laborID.String())

	err := h.GetFetalMonitoring(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- Update/Delete Tests for Delivery --

func TestHandler_UpdateDelivery(t *testing.T) {
	h, e := newTestHandler()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	h.svc.CreateDelivery(nil, d)

	body := `{"pregnancy_id":"` + d.PregnancyID.String() + `","patient_id":"` + d.PatientID.String() + `","delivery_datetime":"2025-06-01T10:00:00Z","delivery_method":"cesarean","delivering_provider_id":"` + d.DeliveringProviderID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())

	err := h.UpdateDelivery(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteDelivery(t *testing.T) {
	h, e := newTestHandler()
	d := &DeliveryRecord{
		PregnancyID:          uuid.New(),
		PatientID:            uuid.New(),
		DeliveryDatetime:     time.Now(),
		DeliveryMethod:       "vaginal",
		DeliveringProviderID: uuid.New(),
	}
	h.svc.CreateDelivery(nil, d)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(d.ID.String())

	err := h.DeleteDelivery(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- List/Update/Delete Tests for Newborn --

func TestHandler_ListNewborns(t *testing.T) {
	h, e := newTestHandler()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	h.svc.CreateNewborn(nil, n)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListNewborns(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdateNewborn(t *testing.T) {
	h, e := newTestHandler()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	h.svc.CreateNewborn(nil, n)

	body := `{"delivery_id":"` + n.DeliveryID.String() + `","birth_datetime":"2025-06-01T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(n.ID.String())

	err := h.UpdateNewborn(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeleteNewborn(t *testing.T) {
	h, e := newTestHandler()
	n := &NewbornRecord{DeliveryID: uuid.New(), BirthDatetime: time.Now()}
	h.svc.CreateNewborn(nil, n)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(n.ID.String())

	err := h.DeleteNewborn(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- Get/List/Update/Delete Tests for Postpartum --

func TestHandler_GetPostpartum(t *testing.T) {
	h, e := newTestHandler()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreatePostpartum(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetPostpartum(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_ListPostpartumRecords(t *testing.T) {
	h, e := newTestHandler()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreatePostpartum(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListPostpartumRecords(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_UpdatePostpartum(t *testing.T) {
	h, e := newTestHandler()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreatePostpartum(nil, p)

	body := `{"pregnancy_id":"` + p.PregnancyID.String() + `","patient_id":"` + p.PatientID.String() + `"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.UpdatePostpartum(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePostpartum(t *testing.T) {
	h, e := newTestHandler()
	p := &PostpartumRecord{PregnancyID: uuid.New(), PatientID: uuid.New()}
	h.svc.CreatePostpartum(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.DeletePostpartum(c)
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
		"POST:/api/v1/pregnancies",
		"GET:/api/v1/pregnancies",
		"GET:/api/v1/pregnancies/:id",
		"PUT:/api/v1/pregnancies/:id",
		"DELETE:/api/v1/pregnancies/:id",
		"POST:/api/v1/pregnancies/:id/prenatal-visits",
		"GET:/api/v1/pregnancies/:id/prenatal-visits",
		"GET:/api/v1/prenatal-visits/:id",
		"POST:/api/v1/labor-records",
		"GET:/api/v1/labor-records",
		"GET:/api/v1/labor-records/:id",
		"POST:/api/v1/deliveries",
		"GET:/api/v1/deliveries/:id",
		"POST:/api/v1/newborns",
		"GET:/api/v1/newborns",
		"GET:/api/v1/newborns/:id",
		"POST:/api/v1/postpartum-records",
		"GET:/api/v1/postpartum-records",
		"GET:/api/v1/postpartum-records/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
