package endpoint

import (
	"context"
	"fmt"

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

type endpointRepoPG struct{ pool *pgxpool.Pool }

func NewEndpointRepoPG(pool *pgxpool.Pool) EndpointRepository {
	return &endpointRepoPG{pool: pool}
}

func (r *endpointRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const epCols = `id, fhir_id, status, connection_type_code, connection_type_display,
	name, managing_org_id, contact_phone, contact_email,
	period_start, period_end, payload_type_code, payload_type_display,
	payload_mime_type, address, header,
	version_id, created_at, updated_at`

func (r *endpointRepoPG) scanRow(row pgx.Row) (*Endpoint, error) {
	var e Endpoint
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.ConnectionTypeCode, &e.ConnectionTypeDisplay,
		&e.Name, &e.ManagingOrgID, &e.ContactPhone, &e.ContactEmail,
		&e.PeriodStart, &e.PeriodEnd, &e.PayloadTypeCode, &e.PayloadTypeDisplay,
		&e.PayloadMimeType, &e.Address, &e.Header,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *endpointRepoPG) Create(ctx context.Context, e *Endpoint) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO endpoint (id, fhir_id, status, connection_type_code, connection_type_display,
			name, managing_org_id, contact_phone, contact_email,
			period_start, period_end, payload_type_code, payload_type_display,
			payload_mime_type, address, header)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		e.ID, e.FHIRID, e.Status, e.ConnectionTypeCode, e.ConnectionTypeDisplay,
		e.Name, e.ManagingOrgID, e.ContactPhone, e.ContactEmail,
		e.PeriodStart, e.PeriodEnd, e.PayloadTypeCode, e.PayloadTypeDisplay,
		e.PayloadMimeType, e.Address, e.Header)
	return err
}

func (r *endpointRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Endpoint, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+epCols+` FROM endpoint WHERE id = $1`, id))
}

func (r *endpointRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Endpoint, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+epCols+` FROM endpoint WHERE fhir_id = $1`, fhirID))
}

func (r *endpointRepoPG) Update(ctx context.Context, e *Endpoint) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE endpoint SET status=$2, connection_type_code=$3, connection_type_display=$4,
			name=$5, managing_org_id=$6, contact_phone=$7, contact_email=$8,
			period_start=$9, period_end=$10, payload_type_code=$11, payload_type_display=$12,
			payload_mime_type=$13, address=$14, header=$15, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.ConnectionTypeCode, e.ConnectionTypeDisplay,
		e.Name, e.ManagingOrgID, e.ContactPhone, e.ContactEmail,
		e.PeriodStart, e.PeriodEnd, e.PayloadTypeCode, e.PayloadTypeDisplay,
		e.PayloadMimeType, e.Address, e.Header)
	return err
}

func (r *endpointRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM endpoint WHERE id = $1`, id)
	return err
}

func (r *endpointRepoPG) List(ctx context.Context, limit, offset int) ([]*Endpoint, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM endpoint`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+epCols+` FROM endpoint ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Endpoint
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *endpointRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Endpoint, int, error) {
	query := `SELECT ` + epCols + ` FROM endpoint WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM endpoint WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["name"]; ok {
		query += fmt.Sprintf(` AND name ILIKE '%%' || $%d || '%%'`, idx)
		countQuery += fmt.Sprintf(` AND name ILIKE '%%' || $%d || '%%'`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["organization"]; ok {
		query += fmt.Sprintf(` AND managing_org_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND managing_org_id = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Endpoint
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
