package structuremap

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

type structureMapRepoPG struct{ pool *pgxpool.Pool }

func NewStructureMapRepoPG(pool *pgxpool.Pool) StructureMapRepository {
	return &structureMapRepoPG{pool: pool}
}

func (r *structureMapRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const smCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	structure_url, structure_mode, import_uri,
	version_id, created_at, updated_at`

func (r *structureMapRepoPG) scanRow(row pgx.Row) (*StructureMap, error) {
	var sm StructureMap
	err := row.Scan(&sm.ID, &sm.FHIRID, &sm.Status, &sm.URL, &sm.Name, &sm.Title, &sm.Description, &sm.Publisher, &sm.Date,
		&sm.StructureURL, &sm.StructureMode, &sm.ImportURI,
		&sm.VersionID, &sm.CreatedAt, &sm.UpdatedAt)
	return &sm, err
}

func (r *structureMapRepoPG) Create(ctx context.Context, sm *StructureMap) error {
	sm.ID = uuid.New()
	if sm.FHIRID == "" {
		sm.FHIRID = sm.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO structure_map (id, fhir_id, status, url, name, title, description, publisher, date,
			structure_url, structure_mode, import_uri)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		sm.ID, sm.FHIRID, sm.Status, sm.URL, sm.Name, sm.Title, sm.Description, sm.Publisher, sm.Date,
		sm.StructureURL, sm.StructureMode, sm.ImportURI)
	return err
}

func (r *structureMapRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*StructureMap, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+smCols+` FROM structure_map WHERE id = $1`, id))
}

func (r *structureMapRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*StructureMap, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+smCols+` FROM structure_map WHERE fhir_id = $1`, fhirID))
}

func (r *structureMapRepoPG) Update(ctx context.Context, sm *StructureMap) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE structure_map SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			structure_url=$9, structure_mode=$10, import_uri=$11, updated_at=NOW()
		WHERE id = $1`,
		sm.ID, sm.Status, sm.URL, sm.Name, sm.Title, sm.Description, sm.Publisher, sm.Date,
		sm.StructureURL, sm.StructureMode, sm.ImportURI)
	return err
}

func (r *structureMapRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM structure_map WHERE id = $1`, id)
	return err
}

func (r *structureMapRepoPG) List(ctx context.Context, limit, offset int) ([]*StructureMap, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM structure_map`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+smCols+` FROM structure_map ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*StructureMap
	for rows.Next() {
		sm, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sm)
	}
	return items, total, nil
}

func (r *structureMapRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*StructureMap, int, error) {
	query := `SELECT ` + smCols + ` FROM structure_map WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM structure_map WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["url"]; ok {
		query += fmt.Sprintf(` AND url = $%d`, idx)
		countQuery += fmt.Sprintf(` AND url = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["name"]; ok {
		query += fmt.Sprintf(` AND name ILIKE '%%' || $%d || '%%'`, idx)
		countQuery += fmt.Sprintf(` AND name ILIKE '%%' || $%d || '%%'`, idx)
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
	var items []*StructureMap
	for rows.Next() {
		sm, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sm)
	}
	return items, total, nil
}
