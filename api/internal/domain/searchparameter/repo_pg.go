package searchparameter

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

type searchParameterRepoPG struct{ pool *pgxpool.Pool }

func NewSearchParameterRepoPG(pool *pgxpool.Pool) SearchParameterRepository {
	return &searchParameterRepoPG{pool: pool}
}

func (r *searchParameterRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const spCols = `id, fhir_id, status, url, name, description, code,
	base, type, expression, xpath, target, modifier, comparator,
	publisher, date,
	version_id, created_at, updated_at`

func (r *searchParameterRepoPG) scanRow(row pgx.Row) (*SearchParameter, error) {
	var s SearchParameter
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.URL, &s.Name, &s.Description, &s.Code,
		&s.Base, &s.Type, &s.Expression, &s.XPath, &s.Target, &s.Modifier, &s.Comparator,
		&s.Publisher, &s.Date,
		&s.VersionID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *searchParameterRepoPG) Create(ctx context.Context, s *SearchParameter) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO search_parameter (id, fhir_id, status, url, name, description, code,
			base, type, expression, xpath, target, modifier, comparator,
			publisher, date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		s.ID, s.FHIRID, s.Status, s.URL, s.Name, s.Description, s.Code,
		s.Base, s.Type, s.Expression, s.XPath, s.Target, s.Modifier, s.Comparator,
		s.Publisher, s.Date)
	return err
}

func (r *searchParameterRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SearchParameter, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM search_parameter WHERE id = $1`, id))
}

func (r *searchParameterRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*SearchParameter, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM search_parameter WHERE fhir_id = $1`, fhirID))
}

func (r *searchParameterRepoPG) Update(ctx context.Context, s *SearchParameter) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE search_parameter SET status=$2, url=$3, name=$4, description=$5, code=$6,
			base=$7, type=$8, expression=$9, xpath=$10, target=$11, modifier=$12, comparator=$13,
			publisher=$14, date=$15, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.URL, s.Name, s.Description, s.Code,
		s.Base, s.Type, s.Expression, s.XPath, s.Target, s.Modifier, s.Comparator,
		s.Publisher, s.Date)
	return err
}

func (r *searchParameterRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM search_parameter WHERE id = $1`, id)
	return err
}

func (r *searchParameterRepoPG) List(ctx context.Context, limit, offset int) ([]*SearchParameter, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM search_parameter`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+spCols+` FROM search_parameter ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SearchParameter
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *searchParameterRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SearchParameter, int, error) {
	query := `SELECT ` + spCols + ` FROM search_parameter WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM search_parameter WHERE 1=1`
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
	if p, ok := params["code"]; ok {
		query += fmt.Sprintf(` AND code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["base"]; ok {
		query += fmt.Sprintf(` AND base = $%d`, idx)
		countQuery += fmt.Sprintf(` AND base = $%d`, idx)
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
	var items []*SearchParameter
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}
