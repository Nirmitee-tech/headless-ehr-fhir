package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

type subscriptionRepoPG struct{ pool *pgxpool.Pool }

// NewSubscriptionRepoPG creates a new PostgreSQL-backed subscription repository.
func NewSubscriptionRepoPG(pool *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepoPG{pool: pool}
}

func (r *subscriptionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const subCols = `id, fhir_id, status, criteria, channel_type, channel_endpoint,
	channel_payload, channel_headers, end_time, error_text, version_id,
	created_at, updated_at`

func scanSub(row pgx.Row) (*Subscription, error) {
	var s Subscription
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.Criteria,
		&s.ChannelType, &s.ChannelEndpoint, &s.ChannelPayload, &s.ChannelHeaders,
		&s.EndTime, &s.ErrorText, &s.VersionID,
		&s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *subscriptionRepoPG) Create(ctx context.Context, sub *Subscription) error {
	sub.ID = uuid.New()
	if sub.FHIRID == "" {
		sub.FHIRID = sub.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO subscription (id, fhir_id, status, criteria, channel_type, channel_endpoint,
			channel_payload, channel_headers, end_time, error_text)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		sub.ID, sub.FHIRID, sub.Status, sub.Criteria,
		sub.ChannelType, sub.ChannelEndpoint, sub.ChannelPayload, sub.ChannelHeaders,
		sub.EndTime, sub.ErrorText)
	return err
}

func (r *subscriptionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	return scanSub(r.conn(ctx).QueryRow(ctx, `SELECT `+subCols+` FROM subscription WHERE id = $1`, id))
}

func (r *subscriptionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Subscription, error) {
	return scanSub(r.conn(ctx).QueryRow(ctx, `SELECT `+subCols+` FROM subscription WHERE fhir_id = $1`, fhirID))
}

func (r *subscriptionRepoPG) Update(ctx context.Context, sub *Subscription) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE subscription SET status=$2, criteria=$3, channel_type=$4, channel_endpoint=$5,
			channel_payload=$6, channel_headers=$7, end_time=$8, error_text=$9,
			version_id=$10, updated_at=NOW()
		WHERE id = $1`,
		sub.ID, sub.Status, sub.Criteria, sub.ChannelType, sub.ChannelEndpoint,
		sub.ChannelPayload, sub.ChannelHeaders, sub.EndTime, sub.ErrorText,
		sub.VersionID)
	return err
}

func (r *subscriptionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM subscription WHERE id = $1`, id)
	return err
}

var subscriptionSearchParams = map[string]fhir.SearchParamConfig{
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"type":     {Type: fhir.SearchParamToken, Column: "channel_type"},
	"criteria": {Type: fhir.SearchParamString, Column: "criteria"},
	"url":      {Type: fhir.SearchParamURI, Column: "channel_endpoint"},
}

func (r *subscriptionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Subscription, int, error) {
	qb := fhir.NewSearchQuery("subscription", subCols)
	qb.ApplyParams(params, subscriptionSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Subscription
	for rows.Next() {
		s, err := scanSub(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *subscriptionRepoPG) ListActive(ctx context.Context) ([]*Subscription, error) {
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+subCols+` FROM subscription WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Subscription
	for rows.Next() {
		s, err := scanSub(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, nil
}

func (r *subscriptionRepoPG) ListExpired(ctx context.Context) ([]*Subscription, error) {
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+subCols+` FROM subscription WHERE status = 'active' AND end_time IS NOT NULL AND end_time < NOW()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Subscription
	for rows.Next() {
		s, err := scanSub(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, nil
}

func (r *subscriptionRepoPG) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorText *string) error {
	_, err := r.conn(ctx).Exec(ctx, `UPDATE subscription SET status=$2, error_text=$3, updated_at=NOW() WHERE id = $1`, id, status, errorText)
	return err
}

// -- Notification methods --

const notifCols = `id, subscription_id, resource_type, resource_id, event_type, status,
	payload, attempt_count, max_attempts, next_attempt_at, last_error, delivered_at, created_at`

func scanNotif(row pgx.Row) (*SubscriptionNotification, error) {
	var n SubscriptionNotification
	err := row.Scan(&n.ID, &n.SubscriptionID, &n.ResourceType, &n.ResourceID,
		&n.EventType, &n.Status, &n.Payload,
		&n.AttemptCount, &n.MaxAttempts, &n.NextAttemptAt,
		&n.LastError, &n.DeliveredAt, &n.CreatedAt)
	return &n, err
}

func (r *subscriptionRepoPG) CreateNotification(ctx context.Context, n *SubscriptionNotification) error {
	n.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO subscription_notification (id, subscription_id, resource_type, resource_id,
			event_type, status, payload, attempt_count, max_attempts, next_attempt_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		n.ID, n.SubscriptionID, n.ResourceType, n.ResourceID,
		n.EventType, n.Status, n.Payload,
		n.AttemptCount, n.MaxAttempts, n.NextAttemptAt)
	return err
}

func (r *subscriptionRepoPG) ListPendingNotifications(ctx context.Context, limit int) ([]*SubscriptionNotification, error) {
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+notifCols+` FROM subscription_notification
		WHERE status = 'pending' AND next_attempt_at <= NOW()
		ORDER BY next_attempt_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SubscriptionNotification
	for rows.Next() {
		n, err := scanNotif(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, nil
}

func (r *subscriptionRepoPG) UpdateNotification(ctx context.Context, n *SubscriptionNotification) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE subscription_notification SET status=$2, attempt_count=$3, next_attempt_at=$4,
			last_error=$5, delivered_at=$6
		WHERE id = $1`,
		n.ID, n.Status, n.AttemptCount, n.NextAttemptAt, n.LastError, n.DeliveredAt)
	return err
}

func (r *subscriptionRepoPG) ListNotificationsBySubscription(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*SubscriptionNotification, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM subscription_notification WHERE subscription_id = $1`, subscriptionID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+notifCols+` FROM subscription_notification WHERE subscription_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, subscriptionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SubscriptionNotification
	for rows.Next() {
		n, err := scanNotif(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

func (r *subscriptionRepoPG) DeleteOldNotifications(ctx context.Context, before time.Time, statuses []string) (int64, error) {
	tag, err := r.conn(ctx).Exec(ctx,
		`DELETE FROM subscription_notification WHERE status = ANY($1) AND created_at < $2`,
		statuses, before)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
