// Package notification provides an Email/SMS notification system with template
// rendering, in-memory storage, retry logic, and Echo HTTP handlers.
package notification

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Notification Types
// ---------------------------------------------------------------------------

// NotificationType represents the channel used to deliver a notification.
type NotificationType string

const (
	TypeEmail NotificationType = "email"
	TypeSMS   NotificationType = "sms"
	TypePush  NotificationType = "push"
)

// ---------------------------------------------------------------------------
// Notification
// ---------------------------------------------------------------------------

// Notification represents a single outbound notification.
type Notification struct {
	ID           string            `json:"id"`
	Type         NotificationType  `json:"type"`
	Recipient    string            `json:"recipient"`
	Subject      string            `json:"subject,omitempty"`
	Body         string            `json:"body"`
	TemplateID   string            `json:"template_id,omitempty"`
	TemplateData map[string]string `json:"template_data,omitempty"`
	Priority     string            `json:"priority"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	SentAt       *time.Time        `json:"sent_at,omitempty"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ---------------------------------------------------------------------------
// Sender Interfaces
// ---------------------------------------------------------------------------

// EmailSender is the interface for sending email messages.
type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SMSSender is the interface for sending SMS messages.
type SMSSender interface {
	SendSMS(ctx context.Context, to, body string) error
}

// ---------------------------------------------------------------------------
// Template Engine
// ---------------------------------------------------------------------------

// Template defines a reusable notification template.
type Template struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Subject string           `json:"subject"`
	Body    string           `json:"body"`
	Type    NotificationType `json:"type"`
}

// TemplateEngine manages notification templates and renders them with data.
type TemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]*Template
}

// NewTemplateEngine creates a TemplateEngine with the built-in templates pre-registered.
func NewTemplateEngine() *TemplateEngine {
	e := &TemplateEngine{
		templates: make(map[string]*Template),
	}
	e.registerBuiltIn()
	return e
}

func (e *TemplateEngine) registerBuiltIn() {
	builtIn := []Template{
		{
			ID:      "appointment-reminder",
			Name:    "Appointment Reminder",
			Subject: "Appointment Reminder for {{patient_name}}",
			Body:    "Dear {{patient_name}}, this is a reminder of your appointment on {{date}} at {{time}} with {{provider}}.",
			Type:    TypeEmail,
		},
		{
			ID:      "lab-result-ready",
			Name:    "Lab Result Ready",
			Subject: "Your Lab Results Are Ready",
			Body:    "Dear {{patient_name}}, your {{lab_type}} lab results are now available. Please log in to view them.",
			Type:    TypeEmail,
		},
		{
			ID:      "prescription-filled",
			Name:    "Prescription Filled",
			Subject: "Your Prescription Has Been Filled",
			Body:    "Dear {{patient_name}}, your prescription for {{medication}} has been filled and is ready for pickup at {{pharmacy}}.",
			Type:    TypeEmail,
		},
		{
			ID:      "password-reset",
			Name:    "Password Reset",
			Subject: "Password Reset Request",
			Body:    "You requested a password reset. Click the following link to reset your password: {{reset_link}}",
			Type:    TypeEmail,
		},
		{
			ID:      "visit-summary",
			Name:    "Visit Summary",
			Subject: "Visit Summary for {{patient_name}}",
			Body:    "Dear {{patient_name}}, here is a summary of your visit on {{visit_date}}: {{summary}}",
			Type:    TypeEmail,
		},
	}
	for i := range builtIn {
		t := builtIn[i]
		e.templates[t.ID] = &t
	}
}

// RegisterTemplate adds or replaces a template in the engine.
func (e *TemplateEngine) RegisterTemplate(t Template) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.templates[t.ID] = &t
}

// Render looks up a template by ID and performs {{key}} replacement using the
// supplied data map. Keys present in the template but absent from data are left
// as-is.
func (e *TemplateEngine) Render(templateID string, data map[string]string) (subject, body string, err error) {
	e.mu.RLock()
	t, ok := e.templates[templateID]
	e.mu.RUnlock()
	if !ok {
		return "", "", fmt.Errorf("template %q not found", templateID)
	}

	subject = t.Subject
	body = t.Body
	for k, v := range data {
		placeholder := "{{" + k + "}}"
		subject = strings.ReplaceAll(subject, placeholder, v)
		body = strings.ReplaceAll(body, placeholder, v)
	}
	return subject, body, nil
}

// ---------------------------------------------------------------------------
// Mock Senders (test doubles)
// ---------------------------------------------------------------------------

// EmailCall records a single call to SendEmail.
type EmailCall struct {
	To      string
	Subject string
	Body    string
}

// MockEmailSender is a test double for EmailSender.
type MockEmailSender struct {
	mu         sync.Mutex
	calls      []EmailCall
	ShouldFail bool
	FailError  string
}

// SendEmail records the call and optionally returns an error.
func (m *MockEmailSender) SendEmail(_ context.Context, to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, EmailCall{To: to, Subject: subject, Body: body})
	if m.ShouldFail {
		return errors.New(m.FailError)
	}
	return nil
}

// Calls returns a copy of recorded email calls.
func (m *MockEmailSender) Calls() []EmailCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]EmailCall, len(m.calls))
	copy(out, m.calls)
	return out
}

// SMSCall records a single call to SendSMS.
type SMSCall struct {
	To   string
	Body string
}

// MockSMSSender is a test double for SMSSender.
type MockSMSSender struct {
	mu         sync.Mutex
	calls      []SMSCall
	ShouldFail bool
	FailError  string
}

// SendSMS records the call and optionally returns an error.
func (m *MockSMSSender) SendSMS(_ context.Context, to, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, SMSCall{To: to, Body: body})
	if m.ShouldFail {
		return errors.New(m.FailError)
	}
	return nil
}

// Calls returns a copy of recorded SMS calls.
func (m *MockSMSSender) Calls() []SMSCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]SMSCall, len(m.calls))
	copy(out, m.calls)
	return out
}

// ---------------------------------------------------------------------------
// Notification Manager
// ---------------------------------------------------------------------------

// NotificationManager orchestrates sending, storage, and retrieval of
// notifications.
type NotificationManager struct {
	emailSender   EmailSender
	smsSender     SMSSender
	templates     *TemplateEngine
	mu            sync.RWMutex
	notifications map[string]*Notification
}

// NewNotificationManager constructs a NotificationManager.
func NewNotificationManager(email EmailSender, sms SMSSender, tpl *TemplateEngine) *NotificationManager {
	return &NotificationManager{
		emailSender:   email,
		smsSender:     sms,
		templates:     tpl,
		notifications: make(map[string]*Notification),
	}
}

// Send dispatches a notification through the appropriate channel, assigns an ID
// and timestamps, and persists the result in-memory.
func (m *NotificationManager) Send(ctx context.Context, n *Notification) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	n.CreatedAt = now
	n.Status = "pending"

	var sendErr error
	switch n.Type {
	case TypeEmail:
		sendErr = m.emailSender.SendEmail(ctx, n.Recipient, n.Subject, n.Body)
	case TypeSMS:
		sendErr = m.smsSender.SendSMS(ctx, n.Recipient, n.Body)
	default:
		sendErr = fmt.Errorf("unsupported notification type: %s", n.Type)
	}

	if sendErr != nil {
		n.Status = "failed"
		n.Error = sendErr.Error()
	} else {
		n.Status = "sent"
		sentAt := time.Now().UTC()
		n.SentAt = &sentAt
	}

	m.mu.Lock()
	m.notifications[n.ID] = n
	m.mu.Unlock()

	if sendErr != nil {
		return sendErr
	}
	return nil
}

// SendFromTemplate renders a template and sends the resulting notification.
func (m *NotificationManager) SendFromTemplate(ctx context.Context, templateID string, data map[string]string, recipient string) (*Notification, error) {
	subject, body, err := m.templates.Render(templateID, data)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}

	// Determine type from template
	m.templates.mu.RLock()
	tpl := m.templates.templates[templateID]
	nType := tpl.Type
	m.templates.mu.RUnlock()

	n := &Notification{
		Type:         nType,
		Recipient:    recipient,
		Subject:      subject,
		Body:         body,
		TemplateID:   templateID,
		TemplateData: data,
		Priority:     "normal",
	}

	if err := m.Send(ctx, n); err != nil {
		return n, err
	}
	return n, nil
}

// GetNotification retrieves a notification by ID.
func (m *NotificationManager) GetNotification(_ context.Context, id string) (*Notification, error) {
	m.mu.RLock()
	n, ok := m.notifications[id]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("notification %q not found", id)
	}
	return n, nil
}

// ListByRecipient returns notifications for a given recipient, up to limit.
func (m *NotificationManager) ListByRecipient(_ context.Context, recipient string, limit int) ([]*Notification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Notification
	for _, n := range m.notifications {
		if n.Recipient == recipient {
			result = append(result, n)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// Retry re-sends a failed notification. Returns an error if the notification is
// not in "failed" status.
func (m *NotificationManager) Retry(ctx context.Context, id string) error {
	m.mu.RLock()
	n, ok := m.notifications[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("notification %q not found", id)
	}
	if n.Status != "failed" {
		return fmt.Errorf("notification %q is not in failed status (current: %s)", id, n.Status)
	}

	var sendErr error
	switch n.Type {
	case TypeEmail:
		sendErr = m.emailSender.SendEmail(ctx, n.Recipient, n.Subject, n.Body)
	case TypeSMS:
		sendErr = m.smsSender.SendSMS(ctx, n.Recipient, n.Body)
	default:
		sendErr = fmt.Errorf("unsupported notification type: %s", n.Type)
	}

	m.mu.Lock()
	if sendErr != nil {
		n.Status = "failed"
		n.Error = sendErr.Error()
	} else {
		n.Status = "sent"
		sentAt := time.Now().UTC()
		n.SentAt = &sentAt
		n.Error = ""
	}
	m.mu.Unlock()

	return sendErr
}

// NotificationStats returns counts of notifications grouped by status.
func (m *NotificationManager) NotificationStats(_ context.Context) map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]int)
	for _, n := range m.notifications {
		stats[n.Status]++
	}
	return stats
}

// ---------------------------------------------------------------------------
// HTTP Handler
// ---------------------------------------------------------------------------

// NotificationHandler exposes notification operations over HTTP via Echo.
type NotificationHandler struct {
	manager *NotificationManager
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(mgr *NotificationManager) *NotificationHandler {
	return &NotificationHandler{manager: mgr}
}

// RegisterRoutes registers all notification routes on the given Echo group.
func (h *NotificationHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/notifications/send", h.HandleSend)
	g.POST("/notifications/send-template", h.HandleSendTemplate)
	g.GET("/notifications/stats", h.HandleStats)
	g.GET("/notifications/:id", h.HandleGet)
	g.GET("/notifications", h.HandleList)
	g.POST("/notifications/:id/retry", h.HandleRetry)
}

// sendRequest is the JSON body for POST /notifications/send.
type sendRequest struct {
	Type      NotificationType  `json:"type"`
	Recipient string            `json:"recipient"`
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
	Priority  string            `json:"priority"`
	Metadata  map[string]string `json:"metadata"`
}

// HandleSend handles POST /notifications/send.
func (h *NotificationHandler) HandleSend(c echo.Context) error {
	var req sendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	n := &Notification{
		Type:      req.Type,
		Recipient: req.Recipient,
		Subject:   req.Subject,
		Body:      req.Body,
		Priority:  req.Priority,
		Metadata:  req.Metadata,
	}

	err := h.manager.Send(c.Request().Context(), n)
	if err != nil {
		// Still return the notification (with failed status) so the caller can
		// see the ID and error.
		return c.JSON(http.StatusCreated, n)
	}
	return c.JSON(http.StatusCreated, n)
}

// sendTemplateRequest is the JSON body for POST /notifications/send-template.
type sendTemplateRequest struct {
	TemplateID string            `json:"template_id"`
	Recipient  string            `json:"recipient"`
	Data       map[string]string `json:"data"`
}

// HandleSendTemplate handles POST /notifications/send-template.
func (h *NotificationHandler) HandleSendTemplate(c echo.Context) error {
	var req sendTemplateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	n, err := h.manager.SendFromTemplate(c.Request().Context(), req.TemplateID, req.Data, req.Recipient)
	if err != nil && n == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, n)
}

// HandleGet handles GET /notifications/:id.
func (h *NotificationHandler) HandleGet(c echo.Context) error {
	id := c.Param("id")
	n, err := h.manager.GetNotification(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, n)
}

// HandleList handles GET /notifications?recipient=...
func (h *NotificationHandler) HandleList(c echo.Context) error {
	recipient := c.QueryParam("recipient")
	if recipient == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "recipient query parameter is required"})
	}

	list, err := h.manager.ListByRecipient(c.Request().Context(), recipient, 100)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, list)
}

// HandleRetry handles POST /notifications/:id/retry.
func (h *NotificationHandler) HandleRetry(c echo.Context) error {
	id := c.Param("id")
	if err := h.manager.Retry(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	n, _ := h.manager.GetNotification(c.Request().Context(), id)
	return c.JSON(http.StatusOK, n)
}

// HandleStats handles GET /notifications/stats.
func (h *NotificationHandler) HandleStats(c echo.Context) error {
	stats := h.manager.NotificationStats(c.Request().Context())
	return c.JSON(http.StatusOK, stats)
}
