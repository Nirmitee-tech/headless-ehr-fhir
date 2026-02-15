package fhir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SubscriptionInfo holds the data the notification engine needs from an active subscription.
type SubscriptionInfo struct {
	ID              uuid.UUID
	FHIRID          string
	Criteria        string
	ChannelEndpoint string
	ChannelPayload  string
	ChannelHeaders  []string
}

// NotificationRepo is the subset of the subscription repository the engine needs.
type NotificationRepo interface {
	ListActive(ctx context.Context) ([]SubscriptionInfo, error)
	CreateNotification(ctx context.Context, n *NotificationRecord) error
	ListPendingNotifications(ctx context.Context, limit int) ([]*NotificationRecord, error)
	UpdateNotification(ctx context.Context, n *NotificationRecord) error
	UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status string, errorText *string) error
	ListExpiredSubscriptions(ctx context.Context) ([]ExpiredSubscription, error)
	DeleteOldNotifications(ctx context.Context, before time.Time, statuses []string) (int64, error)
}

// ExpiredSubscription holds minimal info for expiry processing.
type ExpiredSubscription struct {
	ID     uuid.UUID
	FHIRID string
}

// NotificationRecord mirrors the subscription_notification table for the engine.
type NotificationRecord struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	ResourceType   string
	ResourceID     string
	EventType      string
	Status         string
	Payload        json.RawMessage
	AttemptCount   int
	MaxAttempts    int
	NextAttemptAt  time.Time
	LastError      *string
	DeliveredAt    *time.Time

	// Populated by join for delivery
	ChannelEndpoint string
	ChannelPayload  string
	ChannelHeaders  []string
	SubFHIRID       string
}

// NotificationEngine listens for resource events, evaluates them against
// cached subscription criteria, and manages webhook delivery.
type NotificationEngine struct {
	repo   NotificationRepo
	logger zerolog.Logger
	client *http.Client

	mu    sync.RWMutex
	cache []cachedSubscription

	// CacheRefreshInterval controls how often the subscription cache is refreshed.
	CacheRefreshInterval time.Duration
	// DeliveryInterval controls how often pending notifications are polled.
	DeliveryInterval time.Duration
	// DeliveryBatchSize is the max number of pending notifications fetched per tick.
	DeliveryBatchSize int
	// CleanupInterval controls how often old notifications are purged.
	CleanupInterval time.Duration
}

type cachedSubscription struct {
	info         SubscriptionInfo
	resourceType string
	params       map[string]string
}

// NewNotificationEngine creates a new engine. Pass nil logger for a no-op logger.
func NewNotificationEngine(repo NotificationRepo, logger zerolog.Logger) *NotificationEngine {
	return &NotificationEngine{
		repo:                 repo,
		logger:               logger,
		client:               &http.Client{Timeout: 10 * time.Second},
		CacheRefreshInterval: 30 * time.Second,
		DeliveryInterval:     5 * time.Second,
		DeliveryBatchSize:    50,
		CleanupInterval:      1 * time.Hour,
	}
}

// OnResourceEvent implements ResourceEventListener. It evaluates the event
// against cached subscriptions and creates notification rows for matches.
func (ne *NotificationEngine) OnResourceEvent(ctx context.Context, event ResourceEvent) {
	ne.mu.RLock()
	cached := ne.cache
	ne.mu.RUnlock()

	for _, cs := range cached {
		if !matchesCriteria(cs, event) {
			continue
		}
		n := &NotificationRecord{
			SubscriptionID: cs.info.ID,
			ResourceType:   event.ResourceType,
			ResourceID:     event.ResourceID,
			EventType:      event.Action,
			Status:         "pending",
			Payload:        event.Resource,
			MaxAttempts:    5,
			NextAttemptAt:  time.Now(),
		}
		if err := ne.repo.CreateNotification(ctx, n); err != nil {
			ne.logger.Error().Err(err).
				Str("subscription", cs.info.FHIRID).
				Str("resource", event.ResourceType+"/"+event.ResourceID).
				Msg("failed to create notification")
		}
	}
}

// Start runs the background cache refresh, delivery, and expiry loops.
// It blocks until ctx is cancelled.
func (ne *NotificationEngine) Start(ctx context.Context) {
	ne.refreshCache(ctx)

	cacheTicker := time.NewTicker(ne.CacheRefreshInterval)
	deliveryTicker := time.NewTicker(ne.DeliveryInterval)
	expiryTicker := time.NewTicker(5 * time.Minute)
	cleanupTicker := time.NewTicker(ne.CleanupInterval)
	defer cacheTicker.Stop()
	defer deliveryTicker.Stop()
	defer expiryTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cacheTicker.C:
			ne.refreshCache(ctx)
		case <-deliveryTicker.C:
			ne.deliverPending(ctx)
		case <-expiryTicker.C:
			ne.expireSubscriptions(ctx)
		case <-cleanupTicker.C:
			ne.cleanupOldNotifications(ctx)
		}
	}
}

// RefreshCache forces an immediate cache refresh. Useful after subscription CRUD.
func (ne *NotificationEngine) RefreshCache(ctx context.Context) {
	ne.refreshCache(ctx)
}

func (ne *NotificationEngine) refreshCache(ctx context.Context) {
	subs, err := ne.repo.ListActive(ctx)
	if err != nil {
		ne.logger.Error().Err(err).Msg("failed to refresh subscription cache")
		return
	}
	cached := make([]cachedSubscription, 0, len(subs))
	for _, s := range subs {
		rt, params := ParseCriteria(s.Criteria)
		cached = append(cached, cachedSubscription{
			info:         s,
			resourceType: rt,
			params:       params,
		})
	}
	ne.mu.Lock()
	ne.cache = cached
	ne.mu.Unlock()
}

func (ne *NotificationEngine) deliverPending(ctx context.Context) {
	notifications, err := ne.repo.ListPendingNotifications(ctx, ne.DeliveryBatchSize)
	if err != nil {
		ne.logger.Error().Err(err).Msg("failed to list pending notifications")
		return
	}
	for _, n := range notifications {
		ne.deliverOne(ctx, n)
	}
}

func (ne *NotificationEngine) deliverOne(ctx context.Context, n *NotificationRecord) {
	bundle := buildNotificationBundle(n)
	body, err := json.Marshal(bundle)
	if err != nil {
		ne.markFailed(ctx, n, "marshal bundle: "+err.Error())
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.ChannelEndpoint, bytes.NewReader(body))
	if err != nil {
		ne.markFailed(ctx, n, "build request: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", n.ChannelPayload)
	for _, h := range n.ChannelHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := ne.client.Do(req)
	if err != nil {
		ne.markFailed(ctx, n, "http post: "+err.Error())
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		now := time.Now()
		n.Status = "delivered"
		n.DeliveredAt = &now
		n.AttemptCount++
		if err := ne.repo.UpdateNotification(ctx, n); err != nil {
			ne.logger.Error().Err(err).Str("notification", n.ID.String()).Msg("failed to mark delivered")
		}
		return
	}

	ne.markFailed(ctx, n, fmt.Sprintf("http status %d", resp.StatusCode))
}

func (ne *NotificationEngine) markFailed(ctx context.Context, n *NotificationRecord, errMsg string) {
	n.AttemptCount++
	n.LastError = &errMsg

	if n.AttemptCount >= n.MaxAttempts {
		n.Status = "abandoned"
		if err := ne.repo.UpdateNotification(ctx, n); err != nil {
			ne.logger.Error().Err(err).Msg("failed to abandon notification")
		}
		errText := fmt.Sprintf("max delivery attempts reached: %s", errMsg)
		_ = ne.repo.UpdateSubscriptionStatus(ctx, n.SubscriptionID, "error", &errText)
		return
	}

	n.NextAttemptAt = time.Now().Add(retryBackoff(n.AttemptCount))
	if err := ne.repo.UpdateNotification(ctx, n); err != nil {
		ne.logger.Error().Err(err).Msg("failed to update notification retry")
	}
}

func (ne *NotificationEngine) expireSubscriptions(ctx context.Context) {
	expired, err := ne.repo.ListExpiredSubscriptions(ctx)
	if err != nil {
		ne.logger.Error().Err(err).Msg("failed to list expired subscriptions")
		return
	}
	for _, s := range expired {
		_ = ne.repo.UpdateSubscriptionStatus(ctx, s.ID, "off", nil)
		ne.logger.Info().Str("subscription", s.FHIRID).Msg("subscription expired")
	}
	if len(expired) > 0 {
		ne.refreshCache(ctx)
	}
}

func (ne *NotificationEngine) cleanupOldNotifications(ctx context.Context) {
	now := time.Now()
	deliveredCutoff := now.AddDate(0, 0, -30)
	deliveredCount, err := ne.repo.DeleteOldNotifications(ctx, deliveredCutoff, []string{"delivered"})
	if err != nil {
		ne.logger.Error().Err(err).Msg("failed to cleanup delivered notifications")
	} else if deliveredCount > 0 {
		ne.logger.Info().Int64("count", deliveredCount).Msg("cleaned up old delivered notifications")
	}

	abandonedCutoff := now.AddDate(0, 0, -90)
	abandonedCount, err := ne.repo.DeleteOldNotifications(ctx, abandonedCutoff, []string{"abandoned"})
	if err != nil {
		ne.logger.Error().Err(err).Msg("failed to cleanup abandoned notifications")
	} else if abandonedCount > 0 {
		ne.logger.Info().Int64("count", abandonedCount).Msg("cleaned up old abandoned notifications")
	}
}

// ParseCriteria splits a FHIR subscription criteria string into resource type and parameters.
// Examples:
//
//	"Observation?code=1234&status=final" -> ("Observation", {"code":"1234","status":"final"})
//	"Patient" -> ("Patient", {})
func ParseCriteria(criteria string) (string, map[string]string) {
	parts := strings.SplitN(criteria, "?", 2)
	resourceType := strings.TrimSpace(parts[0])
	params := make(map[string]string)
	if len(parts) == 2 && parts[1] != "" {
		for _, param := range strings.Split(parts[1], "&") {
			kv := strings.SplitN(param, "=", 2)
			if len(kv) == 2 {
				params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	return resourceType, params
}

// matchesCriteria checks if a resource event matches a cached subscription.
func matchesCriteria(cs cachedSubscription, event ResourceEvent) bool {
	if cs.resourceType != event.ResourceType {
		return false
	}
	if len(cs.params) == 0 {
		return true
	}

	// Parse the resource JSON for field matching
	var resource map[string]interface{}
	if err := json.Unmarshal(event.Resource, &resource); err != nil {
		return false
	}

	for key, expected := range cs.params {
		actual := extractField(resource, key)
		if actual != expected {
			return false
		}
	}
	return true
}

// extractField retrieves a value from a FHIR resource map for criteria matching.
// Supports dotted paths (e.g., "subject.reference") and common FHIR patterns.
func extractField(resource map[string]interface{}, key string) string {
	// Handle dotted paths like "subject.reference"
	parts := strings.Split(key, ".")
	var current interface{} = resource
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return ""
		}
	}

	switch v := current.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return ""
	}
}

func buildNotificationBundle(n *NotificationRecord) map[string]interface{} {
	now := time.Now().UTC().Format(time.RFC3339)
	entry := map[string]interface{}{
		"fullUrl":  fmt.Sprintf("%s/%s", n.ResourceType, n.ResourceID),
		"resource": json.RawMessage(n.Payload),
		"request": map[string]string{
			"method": actionToMethod(n.EventType),
			"url":    fmt.Sprintf("%s/%s", n.ResourceType, n.ResourceID),
		},
	}
	total := 1
	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "history",
		"timestamp":    now,
		"total":        &total,
		"entry":        []map[string]interface{}{entry},
	}
}

func actionToMethod(action string) string {
	switch action {
	case "create":
		return "POST"
	case "delete":
		return "DELETE"
	default:
		return "PUT"
	}
}

// RetryBackoff returns the delay for a given attempt number (1-indexed).
// Schedule: 30s, 1m, 5m, 15m, 1h
func retryBackoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 30 * time.Second
	case 2:
		return 1 * time.Minute
	case 3:
		return 5 * time.Minute
	case 4:
		return 15 * time.Minute
	default:
		return 1 * time.Hour
	}
}

// PerformHandshake sends a handshake POST to the subscription endpoint.
// Returns nil if the endpoint responds with 2xx.
func (ne *NotificationEngine) PerformHandshake(ctx context.Context, endpoint string, headers []string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader([]byte("{}")))
	if err != nil {
		return fmt.Errorf("build handshake request: %w", err)
	}
	req.Header.Set("Content-Type", "application/fhir+json")
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	resp, err := ne.client.Do(req)
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("handshake returned status %d", resp.StatusCode)
	}
	return nil
}
