package subscription

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Subscription maps to the subscription table (FHIR Subscription resource).
type Subscription struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	FHIRID          string          `db:"fhir_id" json:"fhir_id"`
	Status          string          `db:"status" json:"status"`
	Criteria        string          `db:"criteria" json:"criteria"`
	ChannelType     string          `db:"channel_type" json:"channel_type"`
	ChannelEndpoint string          `db:"channel_endpoint" json:"channel_endpoint"`
	ChannelPayload  string          `db:"channel_payload" json:"channel_payload"`
	ChannelHeaders  json.RawMessage `db:"channel_headers" json:"channel_headers,omitempty"`
	EndTime         *time.Time      `db:"end_time" json:"end_time,omitempty"`
	ErrorText       *string         `db:"error_text" json:"error_text,omitempty"`
	VersionID       int             `db:"version_id" json:"version_id"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (s *Subscription) GetVersionID() int { return s.VersionID }

// SetVersionID sets the current version.
func (s *Subscription) SetVersionID(v int) { s.VersionID = v }

// ToFHIR converts the Subscription to a FHIR R4 Subscription resource map.
func (s *Subscription) ToFHIR() map[string]interface{} {
	channel := map[string]interface{}{
		"type":    s.ChannelType,
		"endpoint": s.ChannelEndpoint,
	}
	if s.ChannelPayload != "" {
		channel["payload"] = s.ChannelPayload
	}
	if len(s.ChannelHeaders) > 0 && string(s.ChannelHeaders) != "null" {
		var headers []string
		if err := json.Unmarshal(s.ChannelHeaders, &headers); err == nil && len(headers) > 0 {
			channel["header"] = headers
		}
	}

	result := map[string]interface{}{
		"resourceType": "Subscription",
		"id":           s.FHIRID,
		"status":       s.Status,
		"criteria":     s.Criteria,
		"channel":      channel,
		"meta": map[string]interface{}{
			"versionId":   s.VersionID,
			"lastUpdated": s.UpdatedAt,
		},
	}
	if s.EndTime != nil {
		result["end"] = s.EndTime.Format(time.RFC3339)
	}
	if s.ErrorText != nil {
		result["error"] = *s.ErrorText
	}
	return result
}

// SubscriptionNotification tracks individual webhook delivery attempts.
type SubscriptionNotification struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	SubscriptionID uuid.UUID       `db:"subscription_id" json:"subscription_id"`
	ResourceType   string          `db:"resource_type" json:"resource_type"`
	ResourceID     string          `db:"resource_id" json:"resource_id"`
	EventType      string          `db:"event_type" json:"event_type"`
	Status         string          `db:"status" json:"status"`
	Payload        json.RawMessage `db:"payload" json:"payload,omitempty"`
	AttemptCount   int             `db:"attempt_count" json:"attempt_count"`
	MaxAttempts    int             `db:"max_attempts" json:"max_attempts"`
	NextAttemptAt  time.Time       `db:"next_attempt_at" json:"next_attempt_at"`
	LastError      *string         `db:"last_error" json:"last_error,omitempty"`
	DeliveredAt    *time.Time      `db:"delivered_at" json:"delivered_at,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}
