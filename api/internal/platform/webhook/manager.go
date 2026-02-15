// Package webhook provides production-grade webhook management for the EHR platform.
// It supports endpoint registration, event-driven delivery with HMAC-SHA256 signing,
// retry logic, delivery logging, and an Echo HTTP handler for API exposure.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Domain structs
// ---------------------------------------------------------------------------

// WebhookEndpoint represents a registered webhook destination.
type WebhookEndpoint struct {
	ID        string            `json:"id"`
	URL       string            `json:"url"`
	Secret    string            `json:"secret,omitempty"`
	Events    []string          `json:"events"`
	TenantID  string            `json:"tenant_id"`
	ClientID  string            `json:"client_id"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// DeliveryAttempt records a single delivery attempt for a webhook event.
type DeliveryAttempt struct {
	ID           string        `json:"id"`
	WebhookID    string        `json:"webhook_id"`
	EventType    string        `json:"event_type"`
	EventID      string        `json:"event_id"`
	Payload      []byte        `json:"payload"`
	Signature    string        `json:"signature"`
	StatusCode   int           `json:"status_code"`
	ResponseBody string        `json:"response_body"`
	Duration     time.Duration `json:"duration_ns"`
	Attempt      int           `json:"attempt"`
	Status       string        `json:"status"` // "success", "failed", "pending"
	Error        string        `json:"error,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// WebhookEvent represents an event to be delivered to webhook endpoints.
type WebhookEvent struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	TenantID     string          `json:"tenant_id"`
	Payload      json.RawMessage `json:"payload"`
	Timestamp    time.Time       `json:"timestamp"`
}

// DeliveryResult summarises the outcome of delivering an event to one endpoint.
type DeliveryResult struct {
	EndpointID string `json:"endpoint_id"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Store interface
// ---------------------------------------------------------------------------

// WebhookStore defines the persistence interface for webhook endpoints and delivery attempts.
type WebhookStore interface {
	CreateEndpoint(ctx context.Context, endpoint *WebhookEndpoint) error
	GetEndpoint(ctx context.Context, id string) (*WebhookEndpoint, error)
	ListEndpoints(ctx context.Context, tenantID string, limit, offset int) ([]*WebhookEndpoint, int, error)
	UpdateEndpoint(ctx context.Context, endpoint *WebhookEndpoint) error
	DeleteEndpoint(ctx context.Context, id string) error
	RecordDelivery(ctx context.Context, attempt *DeliveryAttempt) error
	ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*DeliveryAttempt, int, error)
	GetDelivery(ctx context.Context, id string) (*DeliveryAttempt, error)
}

// ---------------------------------------------------------------------------
// InMemoryWebhookStore
// ---------------------------------------------------------------------------

// InMemoryWebhookStore is a thread-safe, in-memory implementation of WebhookStore.
type InMemoryWebhookStore struct {
	mu         sync.RWMutex
	endpoints  map[string]*WebhookEndpoint
	deliveries map[string]*DeliveryAttempt
	// ordered keys for deterministic pagination
	endpointOrder  []string
	deliveryOrder  []string
}

// NewInMemoryWebhookStore creates a new empty in-memory store.
func NewInMemoryWebhookStore() *InMemoryWebhookStore {
	return &InMemoryWebhookStore{
		endpoints:  make(map[string]*WebhookEndpoint),
		deliveries: make(map[string]*DeliveryAttempt),
	}
}

func (s *InMemoryWebhookStore) CreateEndpoint(_ context.Context, ep *WebhookEndpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[ep.ID] = ep
	s.endpointOrder = append(s.endpointOrder, ep.ID)
	return nil
}

func (s *InMemoryWebhookStore) GetEndpoint(_ context.Context, id string) (*WebhookEndpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ep, ok := s.endpoints[id]
	if !ok {
		return nil, fmt.Errorf("endpoint %s not found", id)
	}
	return ep, nil
}

func (s *InMemoryWebhookStore) ListEndpoints(_ context.Context, tenantID string, limit, offset int) ([]*WebhookEndpoint, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*WebhookEndpoint
	for _, id := range s.endpointOrder {
		ep := s.endpoints[id]
		if ep == nil {
			continue
		}
		if tenantID == "" || ep.TenantID == tenantID {
			filtered = append(filtered, ep)
		}
	}
	total := len(filtered)
	if offset >= total {
		return []*WebhookEndpoint{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (s *InMemoryWebhookStore) UpdateEndpoint(_ context.Context, ep *WebhookEndpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.endpoints[ep.ID]; !ok {
		return fmt.Errorf("endpoint %s not found", ep.ID)
	}
	s.endpoints[ep.ID] = ep
	return nil
}

func (s *InMemoryWebhookStore) DeleteEndpoint(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.endpoints[id]; !ok {
		return fmt.Errorf("endpoint %s not found", id)
	}
	delete(s.endpoints, id)
	// Remove from ordered list
	for i, eid := range s.endpointOrder {
		if eid == id {
			s.endpointOrder = append(s.endpointOrder[:i], s.endpointOrder[i+1:]...)
			break
		}
	}
	return nil
}

func (s *InMemoryWebhookStore) RecordDelivery(_ context.Context, attempt *DeliveryAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries[attempt.ID] = attempt
	s.deliveryOrder = append(s.deliveryOrder, attempt.ID)
	return nil
}

func (s *InMemoryWebhookStore) ListDeliveries(_ context.Context, webhookID string, limit, offset int) ([]*DeliveryAttempt, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*DeliveryAttempt
	for _, id := range s.deliveryOrder {
		d := s.deliveries[id]
		if d == nil {
			continue
		}
		if d.WebhookID == webhookID {
			filtered = append(filtered, d)
		}
	}
	total := len(filtered)
	if offset >= total {
		return []*DeliveryAttempt{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (s *InMemoryWebhookStore) GetDelivery(_ context.Context, id string) (*DeliveryAttempt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deliveries[id]
	if !ok {
		return nil, fmt.Errorf("delivery %s not found", id)
	}
	return d, nil
}

// ---------------------------------------------------------------------------
// Signature helpers
// ---------------------------------------------------------------------------

// SignPayload computes an HMAC-SHA256 signature of the payload using the given secret,
// returning the hex-encoded result.
func SignPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature returns true when the hex-encoded signature matches the HMAC-SHA256
// of payload under the given secret.
func VerifySignature(payload []byte, secret, signature string) bool {
	expected := SignPayload(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ---------------------------------------------------------------------------
// WebhookManager
// ---------------------------------------------------------------------------

// ManagerOption configures a WebhookManager.
type ManagerOption func(*WebhookManager)

// WithHTTPClient overrides the default HTTP client used for deliveries.
func WithHTTPClient(c *http.Client) ManagerOption {
	return func(m *WebhookManager) { m.httpClient = c }
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) ManagerOption {
	return func(m *WebhookManager) { m.maxRetries = n }
}

// WebhookManager orchestrates endpoint registration, event delivery, and retries.
type WebhookManager struct {
	store       WebhookStore
	httpClient  *http.Client
	maxRetries  int
	retryDelays []time.Duration
}

// NewWebhookManager creates a WebhookManager with sensible defaults.
func NewWebhookManager(store WebhookStore, opts ...ManagerOption) *WebhookManager {
	m := &WebhookManager{
		store: store,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries:  3,
		retryDelays: []time.Duration{1 * time.Second, 30 * time.Second, 5 * time.Minute},
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// generateSecret produces a cryptographically random 32-byte hex string.
func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// validateWebhookURL checks that the URL is non-empty and uses http or https.
func validateWebhookURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("url scheme must be http or https, got %q", u.Scheme)
	}
	return nil
}

// RegisterEndpoint validates and persists a new webhook endpoint. If secret is
// empty, a cryptographically random one is generated.
func (m *WebhookManager) RegisterEndpoint(ctx context.Context, rawURL, secret, tenantID, clientID string, events []string) (*WebhookEndpoint, error) {
	if err := validateWebhookURL(rawURL); err != nil {
		return nil, err
	}
	if secret == "" {
		s, err := generateSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secret: %w", err)
		}
		secret = s
	}

	ep := &WebhookEndpoint{
		ID:        uuid.New().String(),
		URL:       rawURL,
		Secret:    secret,
		Events:    events,
		TenantID:  tenantID,
		ClientID:  clientID,
		Status:    "active",
		CreatedAt: time.Now(),
		Metadata:  map[string]string{},
	}
	if err := m.store.CreateEndpoint(ctx, ep); err != nil {
		return nil, err
	}
	return ep, nil
}

// PauseEndpoint sets the endpoint status to "paused".
func (m *WebhookManager) PauseEndpoint(ctx context.Context, id string) error {
	ep, err := m.store.GetEndpoint(ctx, id)
	if err != nil {
		return err
	}
	ep.Status = "paused"
	return m.store.UpdateEndpoint(ctx, ep)
}

// ResumeEndpoint sets the endpoint status to "active".
func (m *WebhookManager) ResumeEndpoint(ctx context.Context, id string) error {
	ep, err := m.store.GetEndpoint(ctx, id)
	if err != nil {
		return err
	}
	ep.Status = "active"
	return m.store.UpdateEndpoint(ctx, ep)
}

// eventMatches returns true if the event type matches a subscription pattern.
// Patterns can be exact ("Patient.create") or wildcard ("*.delete").
func eventMatches(pattern, eventType string) bool {
	if pattern == eventType {
		return true
	}
	// Wildcard matching: *.action
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".action"
		return strings.HasSuffix(eventType, suffix)
	}
	// Wildcard matching: Resource.*
	if strings.HasSuffix(pattern, ".*") {
		prefix := pattern[:len(pattern)-1] // "Resource."
		return strings.HasPrefix(eventType, prefix)
	}
	return false
}

// endpointMatchesEvent returns true if the endpoint subscribes to the event type.
func endpointMatchesEvent(ep *WebhookEndpoint, eventType string) bool {
	for _, pat := range ep.Events {
		if eventMatches(pat, eventType) {
			return true
		}
	}
	return false
}

// Deliver sends the event to all matching, active endpoints for the tenant.
func (m *WebhookManager) Deliver(ctx context.Context, event WebhookEvent) []DeliveryResult {
	// Fetch all endpoints for the tenant.
	endpoints, _, err := m.store.ListEndpoints(ctx, event.TenantID, 1000, 0)
	if err != nil {
		return nil
	}

	var results []DeliveryResult
	for _, ep := range endpoints {
		if ep.Status != "active" {
			continue
		}
		if !endpointMatchesEvent(ep, event.Type) {
			continue
		}
		attempt := m.DeliverToEndpoint(ctx, ep, event)
		results = append(results, DeliveryResult{
			EndpointID: ep.ID,
			Success:    attempt.Status == "success",
			StatusCode: attempt.StatusCode,
			Error:      attempt.Error,
		})
	}
	return results
}

// DeliverToEndpoint signs the payload and POSTs it to the endpoint, recording the result.
func (m *WebhookManager) DeliverToEndpoint(ctx context.Context, ep *WebhookEndpoint, event WebhookEvent) *DeliveryAttempt {
	payload, _ := json.Marshal(event)
	sig := SignPayload(payload, ep.Secret)
	now := time.Now()

	attempt := &DeliveryAttempt{
		ID:        uuid.New().String(),
		WebhookID: ep.ID,
		EventType: event.Type,
		EventID:   event.ID,
		Payload:   payload,
		Signature: sig,
		Attempt:   1,
		Status:    "pending",
		CreatedAt: now,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(payload))
	if err != nil {
		attempt.Status = "failed"
		attempt.Error = err.Error()
		m.store.RecordDelivery(ctx, attempt)
		return attempt
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "sha256="+sig)
	req.Header.Set("X-Webhook-ID", ep.ID)
	req.Header.Set("X-Webhook-Timestamp", now.UTC().Format(time.RFC3339))

	start := time.Now()
	resp, err := m.httpClient.Do(req)
	attempt.Duration = time.Since(start)

	if err != nil {
		attempt.Status = "failed"
		attempt.Error = err.Error()
		attempt.StatusCode = 0
		m.store.RecordDelivery(ctx, attempt)
		return attempt
	}
	defer resp.Body.Close()

	attempt.StatusCode = resp.StatusCode

	// Read at most 1KB of response body.
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	attempt.ResponseBody = string(bodyBytes)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		attempt.Status = "success"
	} else {
		attempt.Status = "failed"
		attempt.Error = fmt.Sprintf("non-2xx response: %d", resp.StatusCode)
	}

	m.store.RecordDelivery(ctx, attempt)
	return attempt
}

// RetryDelivery re-delivers a previously failed attempt, incrementing the attempt counter.
func (m *WebhookManager) RetryDelivery(ctx context.Context, deliveryID string) (*DeliveryAttempt, error) {
	original, err := m.store.GetDelivery(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("delivery not found: %w", err)
	}

	ep, err := m.store.GetEndpoint(ctx, original.WebhookID)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	// Reconstruct the event from the original delivery payload.
	var event WebhookEvent
	if err := json.Unmarshal(original.Payload, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal original payload: %w", err)
	}

	attempt := m.DeliverToEndpoint(ctx, ep, event)
	attempt.Attempt = original.Attempt + 1

	// Update stored delivery with correct attempt number.
	m.store.RecordDelivery(ctx, attempt)

	return attempt, nil
}

// TestEndpoint sends a synthetic test event to verify endpoint connectivity.
func (m *WebhookManager) TestEndpoint(ctx context.Context, endpointID string) (*DeliveryAttempt, error) {
	ep, err := m.store.GetEndpoint(ctx, endpointID)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	testEvent := WebhookEvent{
		ID:           uuid.New().String(),
		Type:         "webhook.test",
		ResourceType: "Webhook",
		ResourceID:   ep.ID,
		TenantID:     ep.TenantID,
		Payload:      json.RawMessage(`{"test":true}`),
		Timestamp:    time.Now(),
	}

	attempt := m.DeliverToEndpoint(ctx, ep, testEvent)
	return attempt, nil
}

// GetDeliveryLogs returns paginated delivery attempts for a webhook endpoint.
func (m *WebhookManager) GetDeliveryLogs(ctx context.Context, webhookID string, limit, offset int) ([]*DeliveryAttempt, int, error) {
	return m.store.ListDeliveries(ctx, webhookID, limit, offset)
}

// ---------------------------------------------------------------------------
// WebhookHandler — Echo HTTP handler
// ---------------------------------------------------------------------------

// WebhookHandler exposes webhook management via Echo HTTP routes.
type WebhookHandler struct {
	manager *WebhookManager
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(manager *WebhookManager) *WebhookHandler {
	return &WebhookHandler{manager: manager}
}

// RegisterRoutes binds all webhook management routes to the given Echo group.
func (h *WebhookHandler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.RegisterEndpoint)
	g.GET("", h.ListEndpoints)
	g.GET("/:id", h.GetEndpoint)
	g.PUT("/:id", h.UpdateEndpoint)
	g.DELETE("/:id", h.DeleteEndpoint)
	g.POST("/:id/test", h.TestEndpointHandler)
	g.GET("/:id/deliveries", h.GetDeliveryLogs)
	g.POST("/:id/pause", h.PauseEndpointHandler)
	g.POST("/:id/resume", h.ResumeEndpointHandler)
	g.POST("/deliveries/:id/retry", h.RetryDeliveryHandler)
}

// registerRequest is the JSON body for endpoint registration.
type registerRequest struct {
	URL      string   `json:"url"`
	Secret   string   `json:"secret"`
	TenantID string   `json:"tenant_id"`
	ClientID string   `json:"client_id"`
	Events   []string `json:"events"`
}

// RegisterEndpoint handles POST /webhooks.
func (h *WebhookHandler) RegisterEndpoint(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ep, err := h.manager.RegisterEndpoint(c.Request().Context(), req.URL, req.Secret, req.TenantID, req.ClientID, req.Events)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ep)
}

// ListEndpoints handles GET /webhooks.
func (h *WebhookHandler) ListEndpoints(c echo.Context) error {
	tenantID := c.QueryParam("tenant_id")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	eps, total, err := h.manager.store.ListEndpoints(c.Request().Context(), tenantID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    eps,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"has_more": offset+limit < total,
	})
}

// GetEndpoint handles GET /webhooks/:id.
func (h *WebhookHandler) GetEndpoint(c echo.Context) error {
	id := c.Param("id")
	ep, err := h.manager.store.GetEndpoint(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "endpoint not found")
	}
	return c.JSON(http.StatusOK, ep)
}

// updateRequest is the JSON body for endpoint updates.
type updateRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Status string   `json:"status"`
}

// UpdateEndpoint handles PUT /webhooks/:id.
func (h *WebhookHandler) UpdateEndpoint(c echo.Context) error {
	id := c.Param("id")
	ep, err := h.manager.store.GetEndpoint(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "endpoint not found")
	}
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.URL != "" {
		if err := validateWebhookURL(req.URL); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		ep.URL = req.URL
	}
	if len(req.Events) > 0 {
		ep.Events = req.Events
	}
	if req.Status != "" {
		ep.Status = req.Status
	}
	if err := h.manager.store.UpdateEndpoint(c.Request().Context(), ep); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, ep)
}

// DeleteEndpoint handles DELETE /webhooks/:id.
func (h *WebhookHandler) DeleteEndpoint(c echo.Context) error {
	id := c.Param("id")
	if err := h.manager.store.DeleteEndpoint(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "endpoint not found")
	}
	return c.NoContent(http.StatusNoContent)
}

// TestEndpointHandler handles POST /webhooks/:id/test.
func (h *WebhookHandler) TestEndpointHandler(c echo.Context) error {
	id := c.Param("id")
	attempt, err := h.manager.TestEndpoint(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, attempt)
}

// GetDeliveryLogs handles GET /webhooks/:id/deliveries.
func (h *WebhookHandler) GetDeliveryLogs(c echo.Context) error {
	webhookID := c.Param("id")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	logs, total, err := h.manager.GetDeliveryLogs(c.Request().Context(), webhookID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    logs,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"has_more": offset+limit < total,
	})
}

// RetryDeliveryHandler handles POST /webhooks/deliveries/:id/retry.
func (h *WebhookHandler) RetryDeliveryHandler(c echo.Context) error {
	id := c.Param("id")
	attempt, err := h.manager.RetryDelivery(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, attempt)
}

// PauseEndpointHandler handles POST /webhooks/:id/pause.
func (h *WebhookHandler) PauseEndpointHandler(c echo.Context) error {
	id := c.Param("id")
	if err := h.manager.PauseEndpoint(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "paused"})
}

// ResumeEndpointHandler handles POST /webhooks/:id/resume.
func (h *WebhookHandler) ResumeEndpointHandler(c echo.Context) error {
	id := c.Param("id")
	if err := h.manager.ResumeEndpoint(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "active"})
}
