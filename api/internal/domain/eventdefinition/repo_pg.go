package eventdefinition

import (
	"context"

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

type eventDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewEventDefinitionRepoPG(pool *pgxpool.Pool) EventDefinitionRepository {
	return &eventDefinitionRepoPG{pool: pool}
}

func (r *eventDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const edCols = `id, fhir_id, status, url, name, title, description, publisher, date, purpose,
	trigger_type, trigger_name, trigger_condition,
	version_id, created_at, updated_at`

func (r *eventDefinitionRepoPG) scanRow(row pgx.Row) (*EventDefinition, error) {
	var e EventDefinition
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.URL, &e.Name, &e.Title, &e.Description, &e.Publisher, &e.Date, &e.Purpose,
		&e.TriggerType, &e.TriggerName, &e.TriggerCondition,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *eventDefinitionRepoPG) Create(ctx context.Context, e *EventDefinition) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO event_definition (id, fhir_id, status, url, name, title, description, publisher, date, purpose,
			trigger_type, trigger_name, trigger_condition)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.FHIRID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date, e.Purpose,
		e.TriggerType, e.TriggerName, e.TriggerCondition)
	return err
}

func (r *eventDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*EventDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+edCols+` FROM event_definition WHERE id = $1`, id))
}

func (r *eventDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*EventDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+edCols+` FROM event_definition WHERE fhir_id = $1`, fhirID))
}

func (r *eventDefinitionRepoPG) Update(ctx context.Context, e *EventDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE event_definition SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8, purpose=$9,
			trigger_type=$10, trigger_name=$11, trigger_condition=$12, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date, e.Purpose,
		e.TriggerType, e.TriggerName, e.TriggerCondition)
	return err
}

func (r *eventDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM event_definition WHERE id = $1`, id)
	return err
}

func (r *eventDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*EventDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM event_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+edCols+` FROM event_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EventDefinition
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

var eventDefinitionSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"url":    {Type: fhir.SearchParamToken, Column: "url"},
}

func (r *eventDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EventDefinition, int, error) {
	qb := fhir.NewSearchQuery("event_definition", edCols)
	qb.ApplyParams(params, eventDefinitionSearchParams)
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
	var items []*EventDefinition
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
