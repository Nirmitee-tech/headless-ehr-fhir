package conceptmap

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

type conceptMapRepoPG struct{ pool *pgxpool.Pool }

func NewConceptMapRepoPG(pool *pgxpool.Pool) ConceptMapRepository {
	return &conceptMapRepoPG{pool: pool}
}

func (r *conceptMapRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const cmCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	source_uri, target_uri, purpose,
	version_id, created_at, updated_at`

func (r *conceptMapRepoPG) scanRow(row pgx.Row) (*ConceptMap, error) {
	var cm ConceptMap
	err := row.Scan(&cm.ID, &cm.FHIRID, &cm.Status, &cm.URL, &cm.Name, &cm.Title,
		&cm.Description, &cm.Publisher, &cm.Date,
		&cm.SourceURI, &cm.TargetURI, &cm.Purpose,
		&cm.VersionID, &cm.CreatedAt, &cm.UpdatedAt)
	return &cm, err
}

func (r *conceptMapRepoPG) Create(ctx context.Context, cm *ConceptMap) error {
	cm.ID = uuid.New()
	if cm.FHIRID == "" {
		cm.FHIRID = cm.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO concept_map (id, fhir_id, status, url, name, title, description, publisher, date,
			source_uri, target_uri, purpose)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		cm.ID, cm.FHIRID, cm.Status, cm.URL, cm.Name, cm.Title,
		cm.Description, cm.Publisher, cm.Date,
		cm.SourceURI, cm.TargetURI, cm.Purpose)
	return err
}

func (r *conceptMapRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ConceptMap, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+cmCols+` FROM concept_map WHERE id = $1`, id))
}

func (r *conceptMapRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ConceptMap, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+cmCols+` FROM concept_map WHERE fhir_id = $1`, fhirID))
}

func (r *conceptMapRepoPG) Update(ctx context.Context, cm *ConceptMap) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE concept_map SET status=$2, url=$3, name=$4, title=$5, description=$6,
			publisher=$7, date=$8, source_uri=$9, target_uri=$10, purpose=$11, updated_at=NOW()
		WHERE id = $1`,
		cm.ID, cm.Status, cm.URL, cm.Name, cm.Title, cm.Description,
		cm.Publisher, cm.Date, cm.SourceURI, cm.TargetURI, cm.Purpose)
	return err
}

func (r *conceptMapRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM concept_map WHERE id = $1`, id)
	return err
}

func (r *conceptMapRepoPG) List(ctx context.Context, limit, offset int) ([]*ConceptMap, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM concept_map`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cmCols+` FROM concept_map ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ConceptMap
	for rows.Next() {
		cm, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cm)
	}
	return items, total, nil
}

func (r *conceptMapRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ConceptMap, int, error) {
	query := `SELECT ` + cmCols + ` FROM concept_map WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM concept_map WHERE 1=1`
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
	if p, ok := params["source"]; ok {
		query += fmt.Sprintf(` AND source_uri = $%d`, idx)
		countQuery += fmt.Sprintf(` AND source_uri = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["target"]; ok {
		query += fmt.Sprintf(` AND target_uri = $%d`, idx)
		countQuery += fmt.Sprintf(` AND target_uri = $%d`, idx)
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
	var items []*ConceptMap
	for rows.Next() {
		cm, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cm)
	}
	return items, total, nil
}
