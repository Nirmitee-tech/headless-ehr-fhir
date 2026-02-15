package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/ehr/ehr/internal/platform/fhir"
)

// NotifyRepoAdapter adapts SubscriptionRepository to fhir.NotificationRepo,
// bridging the domain and platform layers.
type NotifyRepoAdapter struct {
	repo SubscriptionRepository
}

// NewNotifyRepoAdapter creates a new adapter.
func NewNotifyRepoAdapter(repo SubscriptionRepository) *NotifyRepoAdapter {
	return &NotifyRepoAdapter{repo: repo}
}

func (a *NotifyRepoAdapter) ListActive(ctx context.Context) ([]fhir.SubscriptionInfo, error) {
	subs, err := a.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fhir.SubscriptionInfo, len(subs))
	for i, s := range subs {
		var headers []string
		if len(s.ChannelHeaders) > 0 && string(s.ChannelHeaders) != "null" {
			_ = json.Unmarshal(s.ChannelHeaders, &headers)
		}
		out[i] = fhir.SubscriptionInfo{
			ID:              s.ID,
			FHIRID:          s.FHIRID,
			Criteria:        s.Criteria,
			ChannelEndpoint: s.ChannelEndpoint,
			ChannelPayload:  s.ChannelPayload,
			ChannelHeaders:  headers,
		}
	}
	return out, nil
}

func (a *NotifyRepoAdapter) CreateNotification(ctx context.Context, n *fhir.NotificationRecord) error {
	dn := &SubscriptionNotification{
		SubscriptionID: n.SubscriptionID,
		ResourceType:   n.ResourceType,
		ResourceID:     n.ResourceID,
		EventType:      n.EventType,
		Status:         n.Status,
		Payload:        n.Payload,
		AttemptCount:   n.AttemptCount,
		MaxAttempts:    n.MaxAttempts,
		NextAttemptAt:  n.NextAttemptAt,
	}
	if err := a.repo.CreateNotification(ctx, dn); err != nil {
		return err
	}
	n.ID = dn.ID
	return nil
}

func (a *NotifyRepoAdapter) ListPendingNotifications(ctx context.Context, limit int) ([]*fhir.NotificationRecord, error) {
	notifs, err := a.repo.ListPendingNotifications(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*fhir.NotificationRecord, len(notifs))
	for i, n := range notifs {
		out[i] = a.domainToEngine(n)
	}
	// We need channel info for delivery. Look up subscription for each notification.
	for _, rec := range out {
		sub, err := a.repo.GetByID(ctx, rec.SubscriptionID)
		if err != nil {
			continue
		}
		rec.ChannelEndpoint = sub.ChannelEndpoint
		rec.ChannelPayload = sub.ChannelPayload
		rec.SubFHIRID = sub.FHIRID
		var headers []string
		if len(sub.ChannelHeaders) > 0 && string(sub.ChannelHeaders) != "null" {
			_ = json.Unmarshal(sub.ChannelHeaders, &headers)
		}
		rec.ChannelHeaders = headers
	}
	return out, nil
}

func (a *NotifyRepoAdapter) UpdateNotification(ctx context.Context, n *fhir.NotificationRecord) error {
	dn := &SubscriptionNotification{
		ID:            n.ID,
		Status:        n.Status,
		AttemptCount:  n.AttemptCount,
		NextAttemptAt: n.NextAttemptAt,
		LastError:     n.LastError,
		DeliveredAt:   n.DeliveredAt,
	}
	return a.repo.UpdateNotification(ctx, dn)
}

func (a *NotifyRepoAdapter) UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status string, errorText *string) error {
	return a.repo.UpdateStatus(ctx, id, status, errorText)
}

func (a *NotifyRepoAdapter) ListExpiredSubscriptions(ctx context.Context) ([]fhir.ExpiredSubscription, error) {
	subs, err := a.repo.ListExpired(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]fhir.ExpiredSubscription, len(subs))
	for i, s := range subs {
		out[i] = fhir.ExpiredSubscription{
			ID:     s.ID,
			FHIRID: s.FHIRID,
		}
	}
	return out, nil
}

func (a *NotifyRepoAdapter) DeleteOldNotifications(ctx context.Context, before time.Time, statuses []string) (int64, error) {
	return a.repo.DeleteOldNotifications(ctx, before, statuses)
}

func (a *NotifyRepoAdapter) domainToEngine(n *SubscriptionNotification) *fhir.NotificationRecord {
	return &fhir.NotificationRecord{
		ID:             n.ID,
		SubscriptionID: n.SubscriptionID,
		ResourceType:   n.ResourceType,
		ResourceID:     n.ResourceID,
		EventType:      n.EventType,
		Status:         n.Status,
		Payload:        n.Payload,
		AttemptCount:   n.AttemptCount,
		MaxAttempts:    n.MaxAttempts,
		NextAttemptAt:  n.NextAttemptAt,
		LastError:      n.LastError,
		DeliveredAt:    n.DeliveredAt,
	}
}

// ListPendingNotificationsRepo is a helper that returns SubscriptionNotification from repo.
// Exported only for wiring layer use.
func (a *NotifyRepoAdapter) ListPendingNotificationsRepo(ctx context.Context, limit int) ([]*SubscriptionNotification, error) {
	return a.repo.ListPendingNotifications(ctx, limit)
}

// placeholder to satisfy type constraints if needed
var _ fhir.NotificationRepo = (*NotifyRepoAdapter)(nil)

// timeNow is a helper for testing.
var timeNow = func() time.Time { return time.Now() }
