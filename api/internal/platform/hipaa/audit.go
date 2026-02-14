package hipaa

import (
	"context"
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditEvent represents a FHIR-aligned audit event stored in the audit_event table.
type AuditEvent struct {
	ID               uuid.UUID  `json:"id"`
	FHIRId           string     `json:"fhir_id"`
	TypeCode         string     `json:"type_code"`
	TypeDisplay      string     `json:"type_display"`
	SubtypeCode      string     `json:"subtype_code"`
	SubtypeDisplay   string     `json:"subtype_display"`
	Action           string     `json:"action"` // C/R/U/D/E
	PeriodStart      *time.Time `json:"period_start"`
	PeriodEnd        *time.Time `json:"period_end"`
	Recorded         time.Time  `json:"recorded"`
	Outcome          string     `json:"outcome"` // 0/4/8/12
	OutcomeDesc      string     `json:"outcome_desc"`
	AgentTypeCode    string     `json:"agent_type_code"`
	AgentTypeDisplay string     `json:"agent_type_display"`
	AgentWhoID       *uuid.UUID `json:"agent_who_id"`
	AgentWhoDisplay  string     `json:"agent_who_display"`
	AgentAltID       string     `json:"agent_alt_id"`
	AgentName        string     `json:"agent_name"`
	AgentRequestor   bool       `json:"agent_requestor"`
	AgentRoleCode    string     `json:"agent_role_code"`
	AgentRoleDisplay string     `json:"agent_role_display"`
	AgentNetworkAddr string     `json:"agent_network_address"`
	AgentNetworkType string     `json:"agent_network_type"`
	SourceSite       string     `json:"source_site"`
	SourceObserverID string     `json:"source_observer_id"`
	SourceObsDisplay string     `json:"source_observer_display"`
	SourceTypeCode   string     `json:"source_type_code"`
	EntityWhatType   string     `json:"entity_what_type"`
	EntityWhatID     *uuid.UUID `json:"entity_what_id"`
	EntityWhatDisp   string     `json:"entity_what_display"`
	EntityTypeCode   string     `json:"entity_type_code"`
	EntityRoleCode   string     `json:"entity_role_code"`
	EntityLifecycle  string     `json:"entity_lifecycle"`
	EntityName       string     `json:"entity_name"`
	EntityDesc       string     `json:"entity_description"`
	EntityQuery      string     `json:"entity_query"`
	PurposeCode      string     `json:"purpose_of_use_code"`
	PurposeDisplay   string     `json:"purpose_of_use_display"`
	SensitivityLabel string     `json:"sensitivity_label"`
	UserAgentString  string     `json:"user_agent_string"`
	SessionID        string     `json:"session_id"`
	CreatedAt        time.Time  `json:"created_at"`
}

// PHIAccessLog represents a HIPAA access log entry for the hipaa_access_log table.
type PHIAccessLog struct {
	ID              uuid.UUID  `json:"id"`
	PatientID       uuid.UUID  `json:"patient_id"`
	AccessedByID    uuid.UUID  `json:"accessed_by_id"`
	AccessedByName  string     `json:"accessed_by_name"`
	AccessedByRole  string     `json:"accessed_by_role"`
	ResourceType    string     `json:"resource_type"`
	ResourceID      uuid.UUID  `json:"resource_id"`
	Action          string     `json:"action"`
	ReasonCode      string     `json:"reason_code"`
	ReasonDisplay   string     `json:"reason_display"`
	IsBreakGlass    bool       `json:"is_break_glass"`
	BreakGlassRsn   string     `json:"break_glass_reason"`
	IPAddress       string     `json:"ip_address"`
	UserAgent       string     `json:"user_agent"`
	SessionID       string     `json:"session_id"`
	AccessedAt      time.Time  `json:"accessed_at"`
}

// AuditLogger writes HIPAA-compliant audit events to the database.
type AuditLogger struct {
	pool *pgxpool.Pool
}

// NewAuditLogger creates a new AuditLogger backed by the given connection pool.
func NewAuditLogger(pool *pgxpool.Pool) *AuditLogger {
	return &AuditLogger{pool: pool}
}

// LogEvent writes an AuditEvent to the audit_event table. It uses the tenant-scoped
// connection from context when available, falling back to pool.Acquire.
func (a *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	if event.FHIRId == "" {
		event.FHIRId = uuid.New().String()
	}
	if event.Recorded.IsZero() {
		event.Recorded = time.Now().UTC()
	}

	const query = `
		INSERT INTO audit_event (
			fhir_id, type_code, type_display, subtype_code, subtype_display,
			action, period_start, period_end, recorded, outcome, outcome_desc,
			agent_type_code, agent_type_display, agent_who_id, agent_who_display,
			agent_alt_id, agent_name, agent_requestor, agent_role_code, agent_role_display,
			agent_network_address, agent_network_type,
			source_site, source_observer_id, source_observer_display, source_type_code,
			entity_what_type, entity_what_id, entity_what_display,
			entity_type_code, entity_role_code, entity_lifecycle, entity_name,
			entity_description, entity_query,
			purpose_of_use_code, purpose_of_use_display,
			sensitivity_label, user_agent_string, session_id
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40
		) RETURNING id, created_at`

	args := []any{
		event.FHIRId, event.TypeCode, event.TypeDisplay, event.SubtypeCode, event.SubtypeDisplay,
		event.Action, event.PeriodStart, event.PeriodEnd, event.Recorded, event.Outcome, event.OutcomeDesc,
		event.AgentTypeCode, event.AgentTypeDisplay, event.AgentWhoID, event.AgentWhoDisplay,
		event.AgentAltID, event.AgentName, event.AgentRequestor, event.AgentRoleCode, event.AgentRoleDisplay,
		event.AgentNetworkAddr, event.AgentNetworkType,
		event.SourceSite, event.SourceObserverID, event.SourceObsDisplay, event.SourceTypeCode,
		event.EntityWhatType, event.EntityWhatID, event.EntityWhatDisp,
		event.EntityTypeCode, event.EntityRoleCode, event.EntityLifecycle, event.EntityName,
		event.EntityDesc, event.EntityQuery,
		event.PurposeCode, event.PurposeDisplay,
		event.SensitivityLabel, event.UserAgentString, event.SessionID,
	}

	conn := db.ConnFromContext(ctx)
	if conn != nil {
		return conn.QueryRow(ctx, query, args...).Scan(&event.ID, &event.CreatedAt)
	}

	poolConn, err := a.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("hipaa audit: acquire connection: %w", err)
	}
	defer poolConn.Release()

	return poolConn.QueryRow(ctx, query, args...).Scan(&event.ID, &event.CreatedAt)
}

// LogPHIAccess writes a PHI access log entry to the hipaa_access_log table.
func (a *AuditLogger) LogPHIAccess(ctx context.Context, log *PHIAccessLog) error {
	if log.AccessedAt.IsZero() {
		log.AccessedAt = time.Now().UTC()
	}

	const query = `
		INSERT INTO hipaa_access_log (
			patient_id, accessed_by_id, accessed_by_name, accessed_by_role,
			resource_type, resource_id, action,
			reason_code, reason_display,
			is_break_glass, break_glass_reason,
			ip_address, user_agent, session_id, accessed_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::inet,$13,$14,$15
		) RETURNING id`

	args := []any{
		log.PatientID, log.AccessedByID, log.AccessedByName, log.AccessedByRole,
		log.ResourceType, log.ResourceID, log.Action,
		log.ReasonCode, log.ReasonDisplay,
		log.IsBreakGlass, log.BreakGlassRsn,
		log.IPAddress, log.UserAgent, log.SessionID, log.AccessedAt,
	}

	conn := db.ConnFromContext(ctx)
	if conn != nil {
		return conn.QueryRow(ctx, query, args...).Scan(&log.ID)
	}

	poolConn, err := a.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("hipaa phi access: acquire connection: %w", err)
	}
	defer poolConn.Release()

	return poolConn.QueryRow(ctx, query, args...).Scan(&log.ID)
}

// LogBreakGlass logs an emergency access override (break-the-glass) event.
// It sets IsBreakGlass to true and records both a PHI access log and an audit event.
func (a *AuditLogger) LogBreakGlass(ctx context.Context, log *PHIAccessLog) error {
	log.IsBreakGlass = true
	if log.AccessedAt.IsZero() {
		log.AccessedAt = time.Now().UTC()
	}

	if err := a.LogPHIAccess(ctx, log); err != nil {
		return fmt.Errorf("hipaa break-glass phi access: %w", err)
	}

	event := &AuditEvent{
		FHIRId:           uuid.New().String(),
		TypeCode:         "emergency",
		TypeDisplay:      "Emergency Access (Break-the-Glass)",
		SubtypeCode:      "break-glass",
		SubtypeDisplay:   "Break-the-Glass Override",
		Action:           "R",
		Recorded:         log.AccessedAt,
		Outcome:          "0",
		OutcomeDesc:      fmt.Sprintf("Break-glass access by %s: %s", log.AccessedByName, log.BreakGlassRsn),
		AgentWhoID:       &log.AccessedByID,
		AgentWhoDisplay:  log.AccessedByName,
		AgentRequestor:   true,
		AgentRoleCode:    log.AccessedByRole,
		AgentNetworkAddr: log.IPAddress,
		EntityWhatType:   log.ResourceType,
		EntityWhatID:     &log.ResourceID,
		PurposeCode:      "ETREAT",
		PurposeDisplay:   "Emergency Treatment",
		SensitivityLabel: "R",
		UserAgentString:  log.UserAgent,
		SessionID:        log.SessionID,
	}

	if err := a.LogEvent(ctx, event); err != nil {
		return fmt.Errorf("hipaa break-glass audit event: %w", err)
	}

	return nil
}

// NewReadEvent creates an AuditEvent pre-configured for a read (R) action.
func NewReadEvent(agentWhoID uuid.UUID, agentName, entityType string, entityID uuid.UUID) *AuditEvent {
	return &AuditEvent{
		FHIRId:          uuid.New().String(),
		TypeCode:        "rest",
		TypeDisplay:     "RESTful Operation",
		SubtypeCode:     "read",
		SubtypeDisplay:  "Read",
		Action:          "R",
		Recorded:        time.Now().UTC(),
		Outcome:         "0",
		AgentWhoID:      &agentWhoID,
		AgentWhoDisplay: agentName,
		AgentRequestor:  true,
		EntityWhatType:  entityType,
		EntityWhatID:    &entityID,
		PurposeCode:     "TREAT",
		PurposeDisplay:  "Treatment",
	}
}

// NewWriteEvent creates an AuditEvent pre-configured for a create (C) action.
func NewWriteEvent(agentWhoID uuid.UUID, agentName, entityType string, entityID uuid.UUID) *AuditEvent {
	return &AuditEvent{
		FHIRId:          uuid.New().String(),
		TypeCode:        "rest",
		TypeDisplay:     "RESTful Operation",
		SubtypeCode:     "create",
		SubtypeDisplay:  "Create",
		Action:          "C",
		Recorded:        time.Now().UTC(),
		Outcome:         "0",
		AgentWhoID:      &agentWhoID,
		AgentWhoDisplay: agentName,
		AgentRequestor:  true,
		EntityWhatType:  entityType,
		EntityWhatID:    &entityID,
		PurposeCode:     "TREAT",
		PurposeDisplay:  "Treatment",
	}
}

// NewDeleteEvent creates an AuditEvent pre-configured for a delete (D) action.
func NewDeleteEvent(agentWhoID uuid.UUID, agentName, entityType string, entityID uuid.UUID) *AuditEvent {
	return &AuditEvent{
		FHIRId:          uuid.New().String(),
		TypeCode:        "rest",
		TypeDisplay:     "RESTful Operation",
		SubtypeCode:     "delete",
		SubtypeDisplay:  "Delete",
		Action:          "D",
		Recorded:        time.Now().UTC(),
		Outcome:         "0",
		AgentWhoID:      &agentWhoID,
		AgentWhoDisplay: agentName,
		AgentRequestor:  true,
		EntityWhatType:  entityType,
		EntityWhatID:    &entityID,
		PurposeCode:     "TREAT",
		PurposeDisplay:  "Treatment",
	}
}
