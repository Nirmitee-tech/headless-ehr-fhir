package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SubscriptionRepository defines the data access interface for subscriptions.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Subscription, error)
	Update(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Subscription, int, error)
	ListActive(ctx context.Context) ([]*Subscription, error)
	ListExpired(ctx context.Context) ([]*Subscription, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorText *string) error

	// Notification methods
	CreateNotification(ctx context.Context, n *SubscriptionNotification) error
	ListPendingNotifications(ctx context.Context, limit int) ([]*SubscriptionNotification, error)
	UpdateNotification(ctx context.Context, n *SubscriptionNotification) error
	ListNotificationsBySubscription(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*SubscriptionNotification, int, error)
	DeleteOldNotifications(ctx context.Context, before time.Time, statuses []string) (int64, error)
}
