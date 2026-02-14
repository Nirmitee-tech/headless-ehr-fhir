package inbox

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func newTestHandler() (*Handler, *echo.Echo) {
	svc := newTestService()
	h := NewHandler(svc)
	e := echo.New()
	return h, e
}

// -- MessagePool Handler Tests --

func TestHandler_CreateMessagePool(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pool_name":"Cardiology","pool_type":"department"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMessagePool(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateMessagePool_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"pool_type":"department"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateMessagePool(c)
	if err == nil {
		t.Error("expected error for missing pool_name")
	}
}

func TestHandler_GetMessagePool(t *testing.T) {
	h, e := newTestHandler()
	p := &MessagePool{PoolName: "Test", PoolType: "shared"}
	h.svc.CreateMessagePool(nil, p)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.GetMessagePool(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetMessagePool_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetMessagePool(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteMessagePool(t *testing.T) {
	h, e := newTestHandler()
	p := &MessagePool{PoolName: "Test", PoolType: "shared"}
	h.svc.CreateMessagePool(nil, p)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.DeleteMessagePool(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddPoolMember(t *testing.T) {
	h, e := newTestHandler()
	p := &MessagePool{PoolName: "Test", PoolType: "shared"}
	h.svc.CreateMessagePool(nil, p)

	body := `{"user_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID.String())

	err := h.AddPoolMember(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result MessagePoolMember
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.PoolID != p.ID {
		t.Error("expected pool_id to match")
	}
}

// -- InboxMessage Handler Tests --

func TestHandler_CreateInboxMessage(t *testing.T) {
	h, e := newTestHandler()
	body := `{"message_type":"result","subject":"Lab Results"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateInboxMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateInboxMessage_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"subject":"Lab Results"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateInboxMessage(c)
	if err == nil {
		t.Error("expected error for missing message_type")
	}
}

func TestHandler_GetInboxMessage(t *testing.T) {
	h, e := newTestHandler()
	m := &InboxMessage{MessageType: "result", Subject: "Lab"}
	h.svc.CreateInboxMessage(nil, m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.GetInboxMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetInboxMessage_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetInboxMessage(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestHandler_DeleteInboxMessage(t *testing.T) {
	h, e := newTestHandler()
	m := &InboxMessage{MessageType: "result", Subject: "Lab"}
	h.svc.CreateInboxMessage(nil, m)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	err := h.DeleteInboxMessage(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

// -- CosignRequest Handler Tests --

func TestHandler_CreateCosignRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"document_type":"progress_note","requester_id":"` + uuid.New().String() + `","cosigner_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCosignRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateCosignRequest_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"requester_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCosignRequest(c)
	if err == nil {
		t.Error("expected error for missing document_type")
	}
}

func TestHandler_GetCosignRequest(t *testing.T) {
	h, e := newTestHandler()
	r := &CosignRequest{DocumentType: "progress_note", RequesterID: uuid.New(), CosignerID: uuid.New()}
	h.svc.CreateCosignRequest(nil, r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(r.ID.String())

	err := h.GetCosignRequest(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -- PatientList Handler Tests --

func TestHandler_CreatePatientList(t *testing.T) {
	h, e := newTestHandler()
	body := `{"list_name":"My Patients","list_type":"personal","owner_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePatientList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreatePatientList_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"list_type":"personal"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePatientList(c)
	if err == nil {
		t.Error("expected error for missing list_name")
	}
}

func TestHandler_GetPatientList(t *testing.T) {
	h, e := newTestHandler()
	l := &PatientList{ListName: "Test", ListType: "personal", OwnerID: uuid.New()}
	h.svc.CreatePatientList(nil, l)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.GetPatientList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_DeletePatientList(t *testing.T) {
	h, e := newTestHandler()
	l := &PatientList{ListName: "Test", ListType: "personal", OwnerID: uuid.New()}
	h.svc.CreatePatientList(nil, l)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.DeletePatientList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_AddPatientListMember(t *testing.T) {
	h, e := newTestHandler()
	l := &PatientList{ListName: "Test", ListType: "personal", OwnerID: uuid.New()}
	h.svc.CreatePatientList(nil, l)

	body := `{"patient_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(l.ID.String())

	err := h.AddPatientListMember(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result PatientListMember
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result.ListID != l.ID {
		t.Error("expected list_id to match")
	}
}

// -- Handoff Handler Tests --

func TestHandler_CreateHandoff(t *testing.T) {
	h, e := newTestHandler()
	body := `{"patient_id":"` + uuid.New().String() + `","from_provider_id":"` + uuid.New().String() + `","to_provider_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateHandoff(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_CreateHandoff_BadRequest(t *testing.T) {
	h, e := newTestHandler()
	body := `{"from_provider_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateHandoff(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestHandler_GetHandoff(t *testing.T) {
	h, e := newTestHandler()
	ho := &HandoffRecord{PatientID: uuid.New(), FromProviderID: uuid.New(), ToProviderID: uuid.New()}
	h.svc.CreateHandoff(nil, ho)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(ho.ID.String())

	err := h.GetHandoff(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetHandoff_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetHandoff(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

// -- Route Registration --

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhir := e.Group("/fhir")
	h.RegisterRoutes(api, fhir)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/message-pools",
		"GET:/api/v1/message-pools/:id",
		"POST:/api/v1/inbox-messages",
		"GET:/api/v1/inbox-messages/:id",
		"POST:/api/v1/cosign-requests",
		"GET:/api/v1/cosign-requests/:id",
		"POST:/api/v1/patient-lists",
		"GET:/api/v1/patient-lists/:id",
		"POST:/api/v1/handoffs",
		"GET:/api/v1/handoffs/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}
