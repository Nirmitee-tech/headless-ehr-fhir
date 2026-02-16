package auditevent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type AuditEventRepoPG struct {
	pool *pgxpool.Pool
}

func NewAuditEventRepoPG(pool *pgxpool.Pool) *AuditEventRepoPG {
	return &AuditEventRepoPG{pool: pool}
}

func (r *AuditEventRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const auditCols = `id, fhir_id, type_code, type_display, subtype_code, subtype_display,
	action, period_start, period_end, recorded, outcome, outcome_desc,
	agent_type_code, agent_type_display, agent_who_id, agent_who_display,
	agent_alt_id, agent_name, agent_requestor, agent_role_code, agent_role_display,
	agent_network_address, agent_network_type,
	source_site, source_observer_id, source_observer_display, source_type_code,
	entity_what_type, entity_what_id, entity_what_display,
	entity_type_code, entity_role_code, entity_lifecycle, entity_name, entity_description, entity_query,
	purpose_of_use_code, purpose_of_use_display,
	sensitivity_label, user_agent_string, session_id,
	created_at`

func scanAudit(row pgx.Row) (*AuditEvent, error) {
	var a AuditEvent
	err := row.Scan(
		&a.ID, &a.FHIRID, &a.TypeCode, &a.TypeDisplay, &a.SubtypeCode, &a.SubtypeDisplay,
		&a.Action, &a.PeriodStart, &a.PeriodEnd, &a.Recorded, &a.Outcome, &a.OutcomeDesc,
		&a.AgentTypeCode, &a.AgentTypeDisplay, &a.AgentWhoID, &a.AgentWhoDisplay,
		&a.AgentAltID, &a.AgentName, &a.AgentRequestor, &a.AgentRoleCode, &a.AgentRoleDisplay,
		&a.AgentNetworkAddr, &a.AgentNetworkType,
		&a.SourceSite, &a.SourceObserverID, &a.SourceObsDisplay, &a.SourceTypeCode,
		&a.EntityWhatType, &a.EntityWhatID, &a.EntityWhatDisp,
		&a.EntityTypeCode, &a.EntityRoleCode, &a.EntityLifecycle, &a.EntityName, &a.EntityDesc, &a.EntityQuery,
		&a.PurposeCode, &a.PurposeDisplay,
		&a.SensitivityLabel, &a.UserAgentString, &a.SessionID,
		&a.CreatedAt,
	)
	return &a, err
}

func (r *AuditEventRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error) {
	q := fmt.Sprintf("SELECT %s FROM audit_event WHERE id = $1", auditCols)
	return scanAudit(r.conn(ctx).QueryRow(ctx, q, id))
}

func (r *AuditEventRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*AuditEvent, error) {
	q := fmt.Sprintf("SELECT %s FROM audit_event WHERE fhir_id = $1", auditCols)
	return scanAudit(r.conn(ctx).QueryRow(ctx, q, fhirID))
}

func (r *AuditEventRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AuditEvent, int, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if v, ok := params["action"]; ok {
		where = append(where, fmt.Sprintf("action = $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := params["type"]; ok {
		where = append(where, fmt.Sprintf("type_code = $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := params["outcome"]; ok {
		where = append(where, fmt.Sprintf("outcome = $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := params["agent"]; ok {
		where = append(where, fmt.Sprintf("agent_who_display ILIKE $%d", idx))
		args = append(args, "%"+v+"%")
		idx++
	}
	if v, ok := params["entity-type"]; ok {
		where = append(where, fmt.Sprintf("entity_what_type = $%d", idx))
		args = append(args, v)
		idx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM audit_event %s", whereClause)
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf("SELECT %s FROM audit_event %s ORDER BY recorded DESC LIMIT $%d OFFSET $%d",
		auditCols, whereClause, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*AuditEvent
	for rows.Next() {
		a, err := scanAudit(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}
