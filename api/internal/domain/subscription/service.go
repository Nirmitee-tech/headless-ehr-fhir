package subscription

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Service provides business logic for subscription management.
type Service struct {
	repo SubscriptionRepository
	vt   *fhir.VersionTracker
}

// NewService creates a new subscription service.
func NewService(repo SubscriptionRepository) *Service {
	return &Service{repo: repo}
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

var validStatuses = map[string]bool{
	"requested": true, "active": true, "error": true, "off": true,
}

var validChannelTypes = map[string]bool{
	"rest-hook": true,
}

// resolveHost is a variable to allow test injection.
var resolveHost = net.LookupHost

func validateEndpointURL(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("endpoint URL scheme must be http or https, got %q", u.Scheme)
	}

	hostname := u.Hostname()
	lower := strings.ToLower(hostname)
	if lower == "localhost" || lower == "0.0.0.0" || lower == "[::]" || lower == "::" {
		return fmt.Errorf("endpoint hostname %q is not allowed", hostname)
	}

	ips, err := resolveHost(hostname)
	if err != nil {
		return fmt.Errorf("cannot resolve endpoint hostname %q: %w", hostname, err)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("endpoint resolves to private/reserved IP %s", ipStr)
		}
		// Block cloud metadata endpoint
		if ip.Equal(net.ParseIP("169.254.169.254")) {
			return fmt.Errorf("endpoint resolves to cloud metadata IP %s", ipStr)
		}
	}

	env := os.Getenv("ENV")
	if env == "production" && scheme != "https" {
		return fmt.Errorf("endpoint must use HTTPS in production")
	}

	return nil
}

func (s *Service) CreateSubscription(ctx context.Context, sub *Subscription) error {
	if sub.Criteria == "" {
		return fmt.Errorf("criteria is required")
	}
	if !strings.Contains(sub.Criteria, "?") {
		// Simple resource type criteria (e.g., "Patient") - must be non-empty
		if sub.Criteria == "" {
			return fmt.Errorf("criteria must contain a resource type")
		}
	}
	if sub.ChannelEndpoint == "" {
		return fmt.Errorf("channel endpoint is required")
	}
	if err := validateEndpointURL(sub.ChannelEndpoint); err != nil {
		return fmt.Errorf("invalid channel endpoint: %w", err)
	}
	if sub.ChannelType == "" {
		sub.ChannelType = "rest-hook"
	}
	if !validChannelTypes[sub.ChannelType] {
		return fmt.Errorf("invalid channel type: %s (supported: rest-hook)", sub.ChannelType)
	}
	if sub.ChannelPayload == "" {
		sub.ChannelPayload = "application/fhir+json"
	}
	if sub.Status == "" {
		sub.Status = "requested"
	}
	if !validStatuses[sub.Status] {
		return fmt.Errorf("invalid status: %s", sub.Status)
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return err
	}
	sub.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Subscription", sub.FHIRID, sub.ToFHIR())
	}
	return nil
}

func (s *Service) GetSubscription(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetSubscriptionByFHIRID(ctx context.Context, fhirID string) (*Subscription, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateSubscription(ctx context.Context, sub *Subscription) error {
	if sub.Status != "" && !validStatuses[sub.Status] {
		return fmt.Errorf("invalid status: %s", sub.Status)
	}
	if sub.ChannelType != "" && !validChannelTypes[sub.ChannelType] {
		return fmt.Errorf("invalid channel type: %s", sub.ChannelType)
	}
	if sub.ChannelEndpoint != "" {
		if err := validateEndpointURL(sub.ChannelEndpoint); err != nil {
			return fmt.Errorf("invalid channel endpoint: %w", err)
		}
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Subscription", sub.FHIRID, sub.VersionID, sub.ToFHIR())
		if err == nil {
			sub.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, sub)
}

func (s *Service) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		sub, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Subscription", sub.FHIRID, sub.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchSubscriptions(ctx context.Context, params map[string]string, limit, offset int) ([]*Subscription, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}

func (s *Service) ListActive(ctx context.Context) ([]*Subscription, error) {
	return s.repo.ListActive(ctx)
}

func (s *Service) ListExpired(ctx context.Context) ([]*Subscription, error) {
	return s.repo.ListExpired(ctx)
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorText *string) error {
	return s.repo.UpdateStatus(ctx, id, status, errorText)
}

func (s *Service) CreateNotification(ctx context.Context, n *SubscriptionNotification) error {
	return s.repo.CreateNotification(ctx, n)
}

func (s *Service) ListPendingNotifications(ctx context.Context, limit int) ([]*SubscriptionNotification, error) {
	return s.repo.ListPendingNotifications(ctx, limit)
}

func (s *Service) UpdateNotification(ctx context.Context, n *SubscriptionNotification) error {
	return s.repo.UpdateNotification(ctx, n)
}

func (s *Service) ListNotificationsBySubscription(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*SubscriptionNotification, int, error) {
	return s.repo.ListNotificationsBySubscription(ctx, subscriptionID, limit, offset)
}
