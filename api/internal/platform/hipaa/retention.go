package hipaa

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// RetentionPolicy defines how long data of a specific type should be retained.
type RetentionPolicy struct {
	ResourceType  string `json:"resource_type"`
	RetentionDays int    `json:"retention_days"`
	ArchiveAfter  int    `json:"archive_after_days,omitempty"` // days before archival
	PurgeAfter    int    `json:"purge_after_days,omitempty"`   // days before purge (0 = never)
	Description   string `json:"description"`
}

// RetentionStatus represents the lifecycle state of a resource.
type RetentionStatus struct {
	State     string    `json:"state"`      // "active", "archive_eligible", "purge_eligible"
	ExpiresAt time.Time `json:"expires_at"` // when current state expires
	PolicyName string   `json:"policy_name"`
}

// Retention state constants.
const (
	RetentionStateActive          = "active"
	RetentionStateArchiveEligible = "archive_eligible"
	RetentionStatePurgeEligible   = "purge_eligible"
)

// DefaultRetentionPolicies returns HIPAA-compliant retention policies.
//
// HIPAA requires covered entities to retain certain records for a minimum of
// 6 years. State laws may extend these requirements.
func DefaultRetentionPolicies() []RetentionPolicy {
	return []RetentionPolicy{
		{
			ResourceType:  "medical_record",
			RetentionDays: 2190, // 6 years
			ArchiveAfter:  1825, // 5 years
			PurgeAfter:    0,    // never purge medical records
			Description:   "Medical records: 6 years from last date of service (HIPAA minimum; state law may require longer)",
		},
		{
			ResourceType:  "audit_log",
			RetentionDays: 2190, // 6 years
			ArchiveAfter:  1095, // 3 years
			PurgeAfter:    2555, // 7 years
			Description:   "Audit logs: HIPAA requires minimum 6-year retention for policies and procedures, including audit trails",
		},
		{
			ResourceType:  "billing_record",
			RetentionDays: 2555, // 7 years
			ArchiveAfter:  1825, // 5 years
			PurgeAfter:    2920, // 8 years
			Description:   "Billing records: 7 years per IRS and CMS requirements",
		},
		{
			ResourceType:  "consent_record",
			RetentionDays: 3650, // 10 years
			ArchiveAfter:  2555, // 7 years
			PurgeAfter:    0,    // never purge consent records
			Description:   "Consent records: 10 years or indefinite; critical for demonstrating authorization",
		},
		{
			ResourceType:  "hipaa_access_log",
			RetentionDays: 2190, // 6 years
			ArchiveAfter:  1095, // 3 years
			PurgeAfter:    2555, // 7 years
			Description:   "HIPAA access logs: 6 years per HIPAA Administrative Simplification regulation",
		},
		{
			ResourceType:  "temporary_data",
			RetentionDays: 90,
			ArchiveAfter:  0,  // no archival for temp data
			PurgeAfter:    90, // purge after 90 days
			Description:   "Temporary/staging data: 90 days maximum retention",
		},
	}
}

// RetentionService manages data lifecycle based on configured retention policies.
type RetentionService struct {
	mu       sync.RWMutex
	policies map[string]RetentionPolicy
	logger   zerolog.Logger
}

// NewRetentionService creates a new RetentionService with the given policies.
func NewRetentionService(policies []RetentionPolicy, logger zerolog.Logger) *RetentionService {
	policyMap := make(map[string]RetentionPolicy, len(policies))
	for _, p := range policies {
		policyMap[p.ResourceType] = p
	}
	return &RetentionService{
		policies: policyMap,
		logger:   logger.With().Str("component", "retention-service").Logger(),
	}
}

// GetPolicy returns the retention policy for a resource type, or nil if not found.
func (s *RetentionService) GetPolicy(resourceType string) *RetentionPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[resourceType]
	if !ok {
		return nil
	}
	return &p
}

// GetAllPolicies returns all configured retention policies.
func (s *RetentionService) GetAllPolicies() []RetentionPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]RetentionPolicy, 0, len(s.policies))
	for _, p := range s.policies {
		result = append(result, p)
	}
	return result
}

// CheckRetention checks if a resource has exceeded its retention period.
// Returns a RetentionStatus indicating whether the resource is active,
// eligible for archival, or eligible for purging.
func (s *RetentionService) CheckRetention(resourceType string, createdAt time.Time) RetentionStatus {
	s.mu.RLock()
	policy, ok := s.policies[resourceType]
	s.mu.RUnlock()

	if !ok {
		// Unknown resource type: treat as active with no expiration
		return RetentionStatus{
			State:      RetentionStateActive,
			ExpiresAt:  time.Time{},
			PolicyName: "unknown",
		}
	}

	now := time.Now().UTC()
	age := now.Sub(createdAt)
	ageDays := int(age.Hours() / 24)

	// Check purge eligibility first (most expired state)
	if policy.PurgeAfter > 0 && ageDays >= policy.PurgeAfter {
		return RetentionStatus{
			State:      RetentionStatePurgeEligible,
			ExpiresAt:  createdAt.AddDate(0, 0, policy.PurgeAfter),
			PolicyName: policy.ResourceType,
		}
	}

	// Check archive eligibility
	if policy.ArchiveAfter > 0 && ageDays >= policy.ArchiveAfter {
		expiresAt := createdAt.AddDate(0, 0, policy.RetentionDays)
		if policy.PurgeAfter > 0 {
			expiresAt = createdAt.AddDate(0, 0, policy.PurgeAfter)
		}
		return RetentionStatus{
			State:      RetentionStateArchiveEligible,
			ExpiresAt:  expiresAt,
			PolicyName: policy.ResourceType,
		}
	}

	// Resource is still active
	expiresAt := createdAt.AddDate(0, 0, policy.RetentionDays)
	if policy.ArchiveAfter > 0 {
		expiresAt = createdAt.AddDate(0, 0, policy.ArchiveAfter)
	}
	return RetentionStatus{
		State:      RetentionStateActive,
		ExpiresAt:  expiresAt,
		PolicyName: policy.ResourceType,
	}
}
