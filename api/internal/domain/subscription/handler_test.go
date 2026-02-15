package subscription

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

func TestGetSubscriptionFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{
		Criteria:        "Observation?code=1234",
		ChannelEndpoint: "https://example.com/webhook",
	}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.FHIRID)
	if err := h.GetSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Subscription" {
		t.Errorf("expected resourceType 'Subscription', got %v", result["resourceType"])
	}
}

func TestGetSubscriptionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.GetSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreateSubscriptionFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	body := `{"criteria":"Observation?code=1234","channel_endpoint":"https://example.com/webhook"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
	}
}

func TestSearchSubscriptionsFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateSubscription(nil, &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"})
	req := httptest.NewRequest(http.MethodGet, "/fhir/Subscription", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.SearchSubscriptionsFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle")
	}
}

func TestDeleteSubscriptionFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.FHIRID)
	if err := h.DeleteSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteSubscriptionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.DeleteSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateSubscription(t *testing.T) {
	h, e := newTestHandler()
	body := `{"criteria":"Observation","channel_endpoint":"https://example.com/webhook"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateSubscription(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetSubscription(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.ID.String())
	if err := h.GetSubscription(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_GetSubscription_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	if err := h.GetSubscription(c); err == nil {
		t.Error("expected error")
	}
}

func TestHandler_DeleteSubscription(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.ID.String())
	if err := h.DeleteSubscription(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_ListSubscriptions(t *testing.T) {
	h, e := newTestHandler()
	h.svc.CreateSubscription(nil, &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.ListSubscriptions(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h, e := newTestHandler()
	api := e.Group("/api/v1")
	fhirGroup := e.Group("/fhir")
	h.RegisterRoutes(api, fhirGroup)
	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes")
	}
	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}
	expected := []string{
		"POST:/api/v1/subscriptions",
		"GET:/api/v1/subscriptions",
		"GET:/api/v1/subscriptions/:id",
		"PUT:/api/v1/subscriptions/:id",
		"DELETE:/api/v1/subscriptions/:id",
		"GET:/fhir/Subscription",
		"GET:/fhir/Subscription/:id",
		"POST:/fhir/Subscription",
		"PUT:/fhir/Subscription/:id",
		"DELETE:/fhir/Subscription/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing route: %s", path)
		}
	}
}

func TestUpdateSubscription_REST_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	body := `{"criteria":"Patient","channel_endpoint":"https://example.com/webhook2","status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.ID.String())
	if err := h.UpdateSubscription(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateSubscription_REST_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	if err := h.UpdateSubscription(c); err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestUpdateSubscriptionFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	body := `{"criteria":"Observation","channel_endpoint":"https://example.com/webhook","status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.FHIRID)
	if err := h.UpdateSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateSubscriptionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	body := `{"criteria":"Observation","channel_endpoint":"https://example.com/webhook","status":"active"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.UpdateSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestListNotifications_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	h.svc.CreateNotification(nil, &SubscriptionNotification{
		SubscriptionID: sub.ID,
		ResourceType:   "Observation",
		ResourceID:     "obs-1",
		EventType:      "create",
		Status:         "pending",
		Payload:        json.RawMessage(`{}`),
		MaxAttempts:    5,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.ID.String())
	if err := h.ListNotifications(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestListNotifications_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	if err := h.ListNotifications(c); err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestPatchSubscriptionFHIR_MergePatch(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook", Status: "requested"}
	h.svc.CreateSubscription(nil, sub)
	body := `{"status":"active"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.FHIRID)
	if err := h.PatchSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestPatchSubscriptionFHIR_UnsupportedContentType(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.FHIRID)
	if err := h.PatchSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

func TestPatchSubscriptionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.PatchSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestVreadSubscriptionFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues(sub.FHIRID, "1")
	if err := h.VreadSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Subscription" {
		t.Errorf("expected resourceType 'Subscription', got %v", result["resourceType"])
	}
}

func TestVreadSubscriptionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "vid")
	c.SetParamValues("nonexistent", "1")
	if err := h.VreadSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHistorySubscriptionFHIR_Success(t *testing.T) {
	h, e := newTestHandler()
	sub := &Subscription{Criteria: "Observation", ChannelEndpoint: "https://example.com/webhook"}
	h.svc.CreateSubscription(nil, sub)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sub.FHIRID)
	if err := h.HistorySubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["type"] != "history" {
		t.Errorf("expected bundle type 'history', got %v", bundle["type"])
	}
}

func TestHistorySubscriptionFHIR_NotFound(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	if err := h.HistorySubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestCreateSubscriptionFHIR_ValidationError(t *testing.T) {
	h, e := newTestHandler()
	body := `{"channel_endpoint":"https://example.com/webhook"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateSubscriptionFHIR(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetSubscription_REST_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	if err := h.GetSubscription(c); err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestDeleteSubscription_REST_InvalidID(t *testing.T) {
	h, e := newTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	if err := h.DeleteSubscription(c); err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}
