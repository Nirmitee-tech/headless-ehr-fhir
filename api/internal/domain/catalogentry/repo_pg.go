package catalogentry

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

type catalogEntryRepoPG struct{ pool *pgxpool.Pool }

func NewCatalogEntryRepoPG(pool *pgxpool.Pool) CatalogEntryRepository {
	return &catalogEntryRepoPG{pool: pool}
}

func (r *catalogEntryRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ceCols = `id, fhir_id, type, orderable, referenced_item_type, referenced_item_reference,
	status, effective_period_start, effective_period_end,
	additional_identifier, classification_code, classification_display,
	validity_period_start, validity_period_end, last_updated_ts,
	version_id, created_at, updated_at`

func (r *catalogEntryRepoPG) scanRow(row pgx.Row) (*CatalogEntry, error) {
	var ce CatalogEntry
	err := row.Scan(&ce.ID, &ce.FHIRID, &ce.Type, &ce.Orderable, &ce.ReferencedItemType, &ce.ReferencedItemReference,
		&ce.Status, &ce.EffectivePeriodStart, &ce.EffectivePeriodEnd,
		&ce.AdditionalIdentifier, &ce.ClassificationCode, &ce.ClassificationDisplay,
		&ce.ValidityPeriodStart, &ce.ValidityPeriodEnd, &ce.LastUpdatedTS,
		&ce.VersionID, &ce.CreatedAt, &ce.UpdatedAt)
	return &ce, err
}

func (r *catalogEntryRepoPG) Create(ctx context.Context, ce *CatalogEntry) error {
	ce.ID = uuid.New()
	if ce.FHIRID == "" {
		ce.FHIRID = ce.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO catalog_entry (id, fhir_id, type, orderable, referenced_item_type, referenced_item_reference,
			status, effective_period_start, effective_period_end,
			additional_identifier, classification_code, classification_display,
			validity_period_start, validity_period_end, last_updated_ts)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		ce.ID, ce.FHIRID, ce.Type, ce.Orderable, ce.ReferencedItemType, ce.ReferencedItemReference,
		ce.Status, ce.EffectivePeriodStart, ce.EffectivePeriodEnd,
		ce.AdditionalIdentifier, ce.ClassificationCode, ce.ClassificationDisplay,
		ce.ValidityPeriodStart, ce.ValidityPeriodEnd, ce.LastUpdatedTS)
	return err
}

func (r *catalogEntryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CatalogEntry, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ceCols+` FROM catalog_entry WHERE id = $1`, id))
}

func (r *catalogEntryRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CatalogEntry, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+ceCols+` FROM catalog_entry WHERE fhir_id = $1`, fhirID))
}

func (r *catalogEntryRepoPG) Update(ctx context.Context, ce *CatalogEntry) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE catalog_entry SET type=$2, orderable=$3, referenced_item_type=$4, referenced_item_reference=$5,
			status=$6, effective_period_start=$7, effective_period_end=$8,
			additional_identifier=$9, classification_code=$10, classification_display=$11,
			validity_period_start=$12, validity_period_end=$13, last_updated_ts=$14, updated_at=NOW()
		WHERE id = $1`,
		ce.ID, ce.Type, ce.Orderable, ce.ReferencedItemType, ce.ReferencedItemReference,
		ce.Status, ce.EffectivePeriodStart, ce.EffectivePeriodEnd,
		ce.AdditionalIdentifier, ce.ClassificationCode, ce.ClassificationDisplay,
		ce.ValidityPeriodStart, ce.ValidityPeriodEnd, ce.LastUpdatedTS)
	return err
}

func (r *catalogEntryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM catalog_entry WHERE id = $1`, id)
	return err
}

func (r *catalogEntryRepoPG) List(ctx context.Context, limit, offset int) ([]*CatalogEntry, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM catalog_entry`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ceCols+` FROM catalog_entry ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CatalogEntry
	for rows.Next() {
		ce, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ce)
	}
	return items, total, nil
}

func (r *catalogEntryRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CatalogEntry, int, error) {
	query := `SELECT ` + ceCols + ` FROM catalog_entry WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM catalog_entry WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["orderable"]; ok {
		query += fmt.Sprintf(` AND orderable = $%d`, idx)
		countQuery += fmt.Sprintf(` AND orderable = $%d`, idx)
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
	var items []*CatalogEntry
	for rows.Next() {
		ce, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ce)
	}
	return items, total, nil
}
