package library

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

type libraryRepoPG struct{ pool *pgxpool.Pool }

func NewLibraryRepoPG(pool *pgxpool.Pool) LibraryRepository {
	return &libraryRepoPG{pool: pool}
}

func (r *libraryRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const libCols = `id, fhir_id, status, url, name, title,
	type_code, type_display, description, publisher, date,
	content_type, content_data,
	version_id, created_at, updated_at`

func (r *libraryRepoPG) scanRow(row pgx.Row) (*Library, error) {
	var l Library
	err := row.Scan(&l.ID, &l.FHIRID, &l.Status, &l.URL, &l.Name, &l.Title,
		&l.TypeCode, &l.TypeDisplay, &l.Description, &l.Publisher, &l.Date,
		&l.ContentType, &l.ContentData,
		&l.VersionID, &l.CreatedAt, &l.UpdatedAt)
	return &l, err
}

func (r *libraryRepoPG) Create(ctx context.Context, l *Library) error {
	l.ID = uuid.New()
	if l.FHIRID == "" {
		l.FHIRID = l.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO library (id, fhir_id, status, url, name, title,
			type_code, type_display, description, publisher, date,
			content_type, content_data)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		l.ID, l.FHIRID, l.Status, l.URL, l.Name, l.Title,
		l.TypeCode, l.TypeDisplay, l.Description, l.Publisher, l.Date,
		l.ContentType, l.ContentData)
	return err
}

func (r *libraryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Library, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+libCols+` FROM library WHERE id = $1`, id))
}

func (r *libraryRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Library, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+libCols+` FROM library WHERE fhir_id = $1`, fhirID))
}

func (r *libraryRepoPG) Update(ctx context.Context, l *Library) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE library SET status=$2, url=$3, name=$4, title=$5,
			type_code=$6, type_display=$7, description=$8, publisher=$9, date=$10,
			content_type=$11, content_data=$12, updated_at=NOW()
		WHERE id = $1`,
		l.ID, l.Status, l.URL, l.Name, l.Title,
		l.TypeCode, l.TypeDisplay, l.Description, l.Publisher, l.Date,
		l.ContentType, l.ContentData)
	return err
}

func (r *libraryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM library WHERE id = $1`, id)
	return err
}

func (r *libraryRepoPG) List(ctx context.Context, limit, offset int) ([]*Library, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM library`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+libCols+` FROM library ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Library
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

func (r *libraryRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Library, int, error) {
	query := `SELECT ` + libCols + ` FROM library WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM library WHERE 1=1`
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
	if p, ok := params["title"]; ok {
		query += fmt.Sprintf(` AND title ILIKE '%%' || $%d || '%%'`, idx)
		countQuery += fmt.Sprintf(` AND title ILIKE '%%' || $%d || '%%'`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["url"]; ok {
		query += fmt.Sprintf(` AND url = $%d`, idx)
		countQuery += fmt.Sprintf(` AND url = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type_code = $%d`, idx)
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
	var items []*Library
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}
