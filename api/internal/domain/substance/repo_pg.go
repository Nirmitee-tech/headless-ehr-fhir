package substance

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

type substanceRepoPG struct{ pool *pgxpool.Pool }

func NewSubstanceRepoPG(pool *pgxpool.Pool) SubstanceRepository {
	return &substanceRepoPG{pool: pool}
}

func (r *substanceRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const subCols = `id, fhir_id, status, category_code, category_display,
	code_code, code_display, code_system, description, expiry,
	quantity_value, quantity_unit,
	version_id, created_at, updated_at`

func (r *substanceRepoPG) scanRow(row pgx.Row) (*Substance, error) {
	var s Substance
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.CategoryCode, &s.CategoryDisplay,
		&s.CodeCode, &s.CodeDisplay, &s.CodeSystem, &s.Description, &s.Expiry,
		&s.QuantityValue, &s.QuantityUnit,
		&s.VersionID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *substanceRepoPG) Create(ctx context.Context, s *Substance) error {
	id := uuid.New()
	s.ID = id.String()
	if s.FHIRID == "" {
		s.FHIRID = s.ID
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO substance (id, fhir_id, status, category_code, category_display,
			code_code, code_display, code_system, description, expiry,
			quantity_value, quantity_unit)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.FHIRID, s.Status, s.CategoryCode, s.CategoryDisplay,
		s.CodeCode, s.CodeDisplay, s.CodeSystem, s.Description, s.Expiry,
		s.QuantityValue, s.QuantityUnit)
	return err
}

func (r *substanceRepoPG) GetByID(ctx context.Context, id string) (*Substance, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+subCols+` FROM substance WHERE id = $1`, id))
}

func (r *substanceRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Substance, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+subCols+` FROM substance WHERE fhir_id = $1`, fhirID))
}

func (r *substanceRepoPG) Update(ctx context.Context, s *Substance) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE substance SET status=$2, category_code=$3, category_display=$4,
			code_code=$5, code_display=$6, code_system=$7, description=$8, expiry=$9,
			quantity_value=$10, quantity_unit=$11, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.CategoryCode, s.CategoryDisplay,
		s.CodeCode, s.CodeDisplay, s.CodeSystem, s.Description, s.Expiry,
		s.QuantityValue, s.QuantityUnit)
	return err
}

func (r *substanceRepoPG) Delete(ctx context.Context, id string) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM substance WHERE id = $1`, id)
	return err
}

func (r *substanceRepoPG) List(ctx context.Context, limit, offset int) ([]*Substance, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM substance`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+subCols+` FROM substance ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Substance
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *substanceRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Substance, int, error) {
	query := `SELECT ` + subCols + ` FROM substance WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM substance WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["code"]; ok {
		query += fmt.Sprintf(` AND code_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND code_code = $%d`, idx)
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
	var items []*Substance
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}
