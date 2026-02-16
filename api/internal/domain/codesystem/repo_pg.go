package codesystem

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

type codeSystemRepoPG struct{ pool *pgxpool.Pool }

func NewCodeSystemRepoPG(pool *pgxpool.Pool) CodeSystemRepository {
	return &codeSystemRepoPG{pool: pool}
}

func (r *codeSystemRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const csCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	content, value_set_uri, hierarchy_meaning, compositional, version_needed, count,
	version_id, created_at, updated_at`

func (r *codeSystemRepoPG) scanRow(row pgx.Row) (*CodeSystem, error) {
	var cs CodeSystem
	err := row.Scan(&cs.ID, &cs.FHIRID, &cs.Status, &cs.URL, &cs.Name, &cs.Title,
		&cs.Description, &cs.Publisher, &cs.Date,
		&cs.Content, &cs.ValueSetURI, &cs.HierarchyMeaning,
		&cs.Compositional, &cs.VersionNeeded, &cs.Count,
		&cs.VersionID, &cs.CreatedAt, &cs.UpdatedAt)
	return &cs, err
}

func (r *codeSystemRepoPG) Create(ctx context.Context, cs *CodeSystem) error {
	cs.ID = uuid.New()
	if cs.FHIRID == "" {
		cs.FHIRID = cs.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO code_system (id, fhir_id, status, url, name, title, description, publisher, date,
			content, value_set_uri, hierarchy_meaning, compositional, version_needed, count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		cs.ID, cs.FHIRID, cs.Status, cs.URL, cs.Name, cs.Title,
		cs.Description, cs.Publisher, cs.Date,
		cs.Content, cs.ValueSetURI, cs.HierarchyMeaning,
		cs.Compositional, cs.VersionNeeded, cs.Count)
	return err
}

func (r *codeSystemRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CodeSystem, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+csCols+` FROM code_system WHERE id = $1`, id))
}

func (r *codeSystemRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CodeSystem, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+csCols+` FROM code_system WHERE fhir_id = $1`, fhirID))
}

func (r *codeSystemRepoPG) Update(ctx context.Context, cs *CodeSystem) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE code_system SET status=$2, url=$3, name=$4, title=$5, description=$6,
			publisher=$7, date=$8, content=$9, value_set_uri=$10, hierarchy_meaning=$11,
			compositional=$12, version_needed=$13, count=$14, updated_at=NOW()
		WHERE id = $1`,
		cs.ID, cs.Status, cs.URL, cs.Name, cs.Title, cs.Description,
		cs.Publisher, cs.Date, cs.Content, cs.ValueSetURI, cs.HierarchyMeaning,
		cs.Compositional, cs.VersionNeeded, cs.Count)
	return err
}

func (r *codeSystemRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM code_system WHERE id = $1`, id)
	return err
}

func (r *codeSystemRepoPG) List(ctx context.Context, limit, offset int) ([]*CodeSystem, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM code_system`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+csCols+` FROM code_system ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CodeSystem
	for rows.Next() {
		cs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cs)
	}
	return items, total, nil
}

func (r *codeSystemRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CodeSystem, int, error) {
	query := `SELECT ` + csCols + ` FROM code_system WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM code_system WHERE 1=1`
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
	if p, ok := params["content"]; ok {
		query += fmt.Sprintf(` AND content = $%d`, idx)
		countQuery += fmt.Sprintf(` AND content = $%d`, idx)
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
	var items []*CodeSystem
	for rows.Next() {
		cs, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cs)
	}
	return items, total, nil
}
