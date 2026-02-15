package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Template Engine Tests
// ---------------------------------------------------------------------------

func TestTemplateEngine_RegisterAndRender(t *testing.T) {
	eng := NewTemplateEngine()
	eng.RegisterTemplate(Template{
		ID:      "test-tpl",
		Name:    "Test Template",
		Subject: "Hello {{name}}",
		Body:    "Dear {{name}}, your code is {{code}}.",
		Type:    TypeEmail,
	})

	subject, body, err := eng.Render("test-tpl", map[string]string{
		"name": "Alice",
		"code": "1234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject != "Hello Alice" {
		t.Errorf("subject = %q, want %q", subject, "Hello Alice")
	}
	if body != "Dear Alice, your code is 1234." {
		t.Errorf("body = %q, want %q", body, "Dear Alice, your code is 1234.")
	}
}

func TestTemplateEngine_RenderMissing(t *testing.T) {
	eng := NewTemplateEngine()
	_, _, err := eng.Render("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
}

func TestTemplateEngine_BuiltInTemplates(t *testing.T) {
	eng := NewTemplateEngine()
	builtIn := []string{
		"appointment-reminder",
		"lab-result-ready",
		"prescription-filled",
		"password-reset",
		"visit-summary",
	}
	for _, id := range builtIn {
		_, _, err := eng.Render(id, map[string]string{
			"patient_name": "Test",
			"date":         "2026-01-01",
			"time":         "10:00",
			"provider":     "Dr. Smith",
			"lab_type":     "CBC",
			"medication":   "Aspirin",
			"pharmacy":     "CVS",
			"reset_link":   "https://example.com/reset",
			"visit_date":   "2026-01-01",
			"summary":      "All good",
		})
		if err != nil {
			t.Errorf("built-in template %q not found: %v", id, err)
		}
	}
}

func TestTemplateEngine_RenderWithData(t *testing.T) {
	eng := NewTemplateEngine()
	eng.RegisterTemplate(Template{
		ID:      "data-tpl",
		Name:    "Data Template",
		Subject: "Order {{order_id}}",
		Body:    "Item: {{item}}, Qty: {{qty}}",
		Type:    TypeEmail,
	})

	subject, body, err := eng.Render("data-tpl", map[string]string{
		"order_id": "ORD-999",
		"item":     "Widget",
		"qty":      "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject != "Order ORD-999" {
		t.Errorf("subject = %q, want %q", subject, "Order ORD-999")
	}
	if body != "Item: Widget, Qty: 5" {
		t.Errorf("body = %q, want %q", body, "Item: Widget, Qty: 5")
	}
}

func TestTemplateEngine_RenderMissingKey(t *testing.T) {
	eng := NewTemplateEngine()
	eng.RegisterTemplate(Template{
		ID:      "partial-tpl",
		Name:    "Partial",
		Subject: "Hi {{name}}",
		Body:    "Your code is {{code}} and token is {{token}}.",
		Type:    TypeEmail,
	})

	subject, body, err := eng.Render("partial-tpl", map[string]string{
		"name": "Bob",
		"code": "5678",
		// "token" deliberately missing
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject != "Hi Bob" {
		t.Errorf("subject = %q, want %q", subject, "Hi Bob")
	}
	// unreplaced keys left as-is
	expected := "Your code is 5678 and token is {{token}}."
	if body != expected {
		t.Errorf("body = %q, want %q", body, expected)
	}
}

// ---------------------------------------------------------------------------
// Notification Manager Tests
// ---------------------------------------------------------------------------

func TestNotificationManager_SendEmail(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	n := &Notification{
		Type:      TypeEmail,
		Recipient: "alice@example.com",
		Subject:   "Test Subject",
		Body:      "Test Body",
		Priority:  "normal",
	}

	err := mgr.Send(context.Background(), n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != "sent" {
		t.Errorf("status = %q, want %q", n.Status, "sent")
	}
	if n.SentAt == nil {
		t.Error("SentAt should be set after successful send")
	}
	if len(emailMock.Calls()) != 1 {
		t.Errorf("expected 1 email call, got %d", len(emailMock.Calls()))
	}
	call := emailMock.Calls()[0]
	if call.To != "alice@example.com" || call.Subject != "Test Subject" || call.Body != "Test Body" {
		t.Errorf("unexpected email call: %+v", call)
	}
}

func TestNotificationManager_SendSMS(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	n := &Notification{
		Type:      TypeSMS,
		Recipient: "+15551234567",
		Body:      "Your code is 1234",
		Priority:  "high",
	}

	err := mgr.Send(context.Background(), n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != "sent" {
		t.Errorf("status = %q, want %q", n.Status, "sent")
	}
	if len(smsMock.Calls()) != 1 {
		t.Errorf("expected 1 sms call, got %d", len(smsMock.Calls()))
	}
	call := smsMock.Calls()[0]
	if call.To != "+15551234567" || call.Body != "Your code is 1234" {
		t.Errorf("unexpected sms call: %+v", call)
	}
}

func TestNotificationManager_SendFailed(t *testing.T) {
	emailMock := &MockEmailSender{ShouldFail: true, FailError: "SMTP connection refused"}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	n := &Notification{
		Type:      TypeEmail,
		Recipient: "fail@example.com",
		Subject:   "Will Fail",
		Body:      "This should fail",
		Priority:  "normal",
	}

	err := mgr.Send(context.Background(), n)
	if err == nil {
		t.Fatal("expected error from failed send")
	}
	if n.Status != "failed" {
		t.Errorf("status = %q, want %q", n.Status, "failed")
	}
	if n.Error != "SMTP connection refused" {
		t.Errorf("error = %q, want %q", n.Error, "SMTP connection refused")
	}
}

func TestNotificationManager_SendFromTemplate(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	eng := NewTemplateEngine()
	mgr := NewNotificationManager(emailMock, smsMock, eng)

	n, err := mgr.SendFromTemplate(context.Background(), "appointment-reminder", map[string]string{
		"patient_name": "Alice",
		"date":         "2026-03-01",
		"time":         "14:00",
		"provider":     "Dr. Smith",
	}, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != "sent" {
		t.Errorf("status = %q, want %q", n.Status, "sent")
	}
	if n.TemplateID != "appointment-reminder" {
		t.Errorf("templateID = %q, want %q", n.TemplateID, "appointment-reminder")
	}
	if !strings.Contains(n.Body, "Alice") {
		t.Errorf("body should contain patient name, got %q", n.Body)
	}
}

func TestNotificationManager_GetNotification(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	n := &Notification{
		Type:      TypeEmail,
		Recipient: "get@example.com",
		Subject:   "Get Test",
		Body:      "Body",
		Priority:  "normal",
	}
	_ = mgr.Send(context.Background(), n)

	got, err := mgr.GetNotification(context.Background(), n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != n.ID {
		t.Errorf("ID = %q, want %q", got.ID, n.ID)
	}
}

func TestNotificationManager_GetNotFound(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	_, err := mgr.GetNotification(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent notification")
	}
}

func TestNotificationManager_ListByRecipient(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	for i := 0; i < 5; i++ {
		_ = mgr.Send(context.Background(), &Notification{
			Type:      TypeEmail,
			Recipient: "list@example.com",
			Subject:   "List Test",
			Body:      "Body",
			Priority:  "normal",
		})
	}
	// different recipient
	_ = mgr.Send(context.Background(), &Notification{
		Type:      TypeEmail,
		Recipient: "other@example.com",
		Subject:   "Other",
		Body:      "Other Body",
		Priority:  "normal",
	})

	list, err := mgr.ListByRecipient(context.Background(), "list@example.com", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 5 {
		t.Errorf("len = %d, want 5", len(list))
	}

	// test limit
	list2, err := mgr.ListByRecipient(context.Background(), "list@example.com", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list2) != 3 {
		t.Errorf("len = %d, want 3", len(list2))
	}
}

func TestNotificationManager_Retry(t *testing.T) {
	emailMock := &MockEmailSender{ShouldFail: true, FailError: "temporary failure"}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	n := &Notification{
		Type:      TypeEmail,
		Recipient: "retry@example.com",
		Subject:   "Retry Test",
		Body:      "Retry Body",
		Priority:  "normal",
	}
	_ = mgr.Send(context.Background(), n)
	if n.Status != "failed" {
		t.Fatalf("expected failed status, got %q", n.Status)
	}

	// Fix the mock so retry succeeds
	emailMock.ShouldFail = false

	err := mgr.Retry(context.Background(), n.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := mgr.GetNotification(context.Background(), n.ID)
	if got.Status != "sent" {
		t.Errorf("status = %q, want %q after retry", got.Status, "sent")
	}
	if got.SentAt == nil {
		t.Error("SentAt should be set after successful retry")
	}
	if got.Error != "" {
		t.Errorf("error should be cleared after retry, got %q", got.Error)
	}
}

func TestNotificationManager_RetryNonFailed(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	n := &Notification{
		Type:      TypeEmail,
		Recipient: "ok@example.com",
		Subject:   "OK",
		Body:      "OK Body",
		Priority:  "normal",
	}
	_ = mgr.Send(context.Background(), n)
	if n.Status != "sent" {
		t.Fatalf("expected sent status, got %q", n.Status)
	}

	err := mgr.Retry(context.Background(), n.ID)
	if err == nil {
		t.Fatal("expected error when retrying non-failed notification")
	}
}

func TestNotificationManager_Stats(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	// Send 3 successful emails
	for i := 0; i < 3; i++ {
		_ = mgr.Send(context.Background(), &Notification{
			Type:      TypeEmail,
			Recipient: "stats@example.com",
			Subject:   "Stats",
			Body:      "Stats Body",
			Priority:  "normal",
		})
	}

	// Send 2 failed emails
	emailMock.ShouldFail = true
	emailMock.FailError = "fail"
	for i := 0; i < 2; i++ {
		_ = mgr.Send(context.Background(), &Notification{
			Type:      TypeEmail,
			Recipient: "stats@example.com",
			Subject:   "Stats Fail",
			Body:      "Fail Body",
			Priority:  "normal",
		})
	}

	stats := mgr.NotificationStats(context.Background())
	if stats["sent"] != 3 {
		t.Errorf("sent = %d, want 3", stats["sent"])
	}
	if stats["failed"] != 2 {
		t.Errorf("failed = %d, want 2", stats["failed"])
	}
}

func TestNotificationManager_ConcurrentSend(t *testing.T) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	mgr := NewNotificationManager(emailMock, smsMock, NewTemplateEngine())

	var wg sync.WaitGroup
	count := 50
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			_ = mgr.Send(context.Background(), &Notification{
				Type:      TypeEmail,
				Recipient: "concurrent@example.com",
				Subject:   "Concurrent",
				Body:      "Concurrent Body",
				Priority:  "normal",
			})
		}()
	}
	wg.Wait()

	stats := mgr.NotificationStats(context.Background())
	if stats["sent"] != count {
		t.Errorf("sent = %d, want %d", stats["sent"], count)
	}
}

// ---------------------------------------------------------------------------
// HTTP Handler Tests
// ---------------------------------------------------------------------------

func setupHandler() (*NotificationHandler, *echo.Echo) {
	emailMock := &MockEmailSender{}
	smsMock := &MockSMSSender{}
	eng := NewTemplateEngine()
	mgr := NewNotificationManager(emailMock, smsMock, eng)
	h := NewNotificationHandler(mgr)
	e := echo.New()
	return h, e
}

func TestNotificationHandler_SendEmail(t *testing.T) {
	h, e := setupHandler()

	body := `{"type":"email","recipient":"handler@example.com","subject":"Handler Test","body":"Handler Body","priority":"normal"}`
	req := httptest.NewRequest(http.MethodPost, "/notifications/send", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/notifications/send")

	err := h.HandleSend(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "sent" {
		t.Errorf("response status = %v, want %q", resp["status"], "sent")
	}
}

func TestNotificationHandler_SendTemplate(t *testing.T) {
	h, e := setupHandler()

	body := `{"template_id":"appointment-reminder","recipient":"tpl@example.com","data":{"patient_name":"Alice","date":"2026-03-01","time":"14:00","provider":"Dr. Smith"}}`
	req := httptest.NewRequest(http.MethodPost, "/notifications/send-template", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/notifications/send-template")

	err := h.HandleSendTemplate(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestNotificationHandler_GetNotification(t *testing.T) {
	h, e := setupHandler()

	// First send one to have something to retrieve
	sendBody := `{"type":"email","recipient":"gethandler@example.com","subject":"Get","body":"Get Body","priority":"normal"}`
	sendReq := httptest.NewRequest(http.MethodPost, "/notifications/send", strings.NewReader(sendBody))
	sendReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	sendRec := httptest.NewRecorder()
	sendCtx := e.NewContext(sendReq, sendRec)
	sendCtx.SetPath("/notifications/send")
	_ = h.HandleSend(sendCtx)

	var sendResp map[string]interface{}
	_ = json.Unmarshal(sendRec.Body.Bytes(), &sendResp)
	id := sendResp["id"].(string)

	// Now GET it
	req := httptest.NewRequest(http.MethodGet, "/notifications/"+id, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/notifications/:id")
	c.SetParamNames("id")
	c.SetParamValues(id)

	err := h.HandleGet(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var getResp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &getResp)
	if getResp["id"] != id {
		t.Errorf("id = %v, want %v", getResp["id"], id)
	}
}

func TestNotificationHandler_ListByRecipient(t *testing.T) {
	h, e := setupHandler()

	// Send two notifications
	for i := 0; i < 2; i++ {
		body := `{"type":"email","recipient":"listhandler@example.com","subject":"List","body":"List Body","priority":"normal"}`
		req := httptest.NewRequest(http.MethodPost, "/notifications/send", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/notifications/send")
		_ = h.HandleSend(c)
	}

	// List them
	req := httptest.NewRequest(http.MethodGet, "/notifications?recipient=listhandler@example.com", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/notifications")

	err := h.HandleList(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var list []map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Errorf("len = %d, want 2", len(list))
	}
}

func TestNotificationHandler_RetryNotification(t *testing.T) {
	emailMock := &MockEmailSender{ShouldFail: true, FailError: "temp error"}
	smsMock := &MockSMSSender{}
	eng := NewTemplateEngine()
	mgr := NewNotificationManager(emailMock, smsMock, eng)
	h := NewNotificationHandler(mgr)
	e := echo.New()

	// Send a failing notification
	sendBody := `{"type":"email","recipient":"retry@example.com","subject":"Retry","body":"Retry Body","priority":"normal"}`
	sendReq := httptest.NewRequest(http.MethodPost, "/notifications/send", strings.NewReader(sendBody))
	sendReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	sendRec := httptest.NewRecorder()
	sendCtx := e.NewContext(sendReq, sendRec)
	sendCtx.SetPath("/notifications/send")
	_ = h.HandleSend(sendCtx)

	var sendResp map[string]interface{}
	_ = json.Unmarshal(sendRec.Body.Bytes(), &sendResp)
	id := sendResp["id"].(string)

	// Fix the mock
	emailMock.ShouldFail = false

	// Retry
	req := httptest.NewRequest(http.MethodPost, "/notifications/"+id+"/retry", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/notifications/:id/retry")
	c.SetParamNames("id")
	c.SetParamValues(id)

	err := h.HandleRetry(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNotificationHandler_Stats(t *testing.T) {
	h, e := setupHandler()

	// Send a couple of notifications first
	for i := 0; i < 3; i++ {
		body := `{"type":"email","recipient":"stats@example.com","subject":"Stats","body":"Stats Body","priority":"normal"}`
		req := httptest.NewRequest(http.MethodPost, "/notifications/send", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/notifications/send")
		_ = h.HandleSend(c)
	}

	req := httptest.NewRequest(http.MethodGet, "/notifications/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/notifications/stats")

	err := h.HandleStats(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var stats map[string]int
	_ = json.Unmarshal(rec.Body.Bytes(), &stats)
	if stats["sent"] != 3 {
		t.Errorf("sent = %d, want 3", stats["sent"])
	}
}
