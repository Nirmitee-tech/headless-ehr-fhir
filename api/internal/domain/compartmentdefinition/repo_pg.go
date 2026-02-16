package compartmentdefinition

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

type compartmentDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewCompartmentDefinitionRepoPG(pool *pgxpool.Pool) CompartmentDefinitionRepository {
	return &compartmentDefinitionRepoPG{pool: pool}
}

func (r *compartmentDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const cdCols = `id, fhir_id, status, url, name, description, publisher, date,
	code, search, resource_type, resource_param,
	version_id, created_at, updated_at`

func (r *compartmentDefinitionRepoPG) scanRow(row pgx.Row) (*CompartmentDefinition, error) {
	var cd CompartmentDefinition
	err := row.Scan(&cd.ID, &cd.FHIRID, &cd.Status, &cd.URL, &cd.Name, &cd.Description, &cd.Publisher, &cd.Date,
		&cd.Code, &cd.Search, &cd.ResourceType, &cd.ResourceParam,
		&cd.VersionID, &cd.CreatedAt, &cd.UpdatedAt)
	return &cd, err
}

func (r *compartmentDefinitionRepoPG) Create(ctx context.Context, cd *CompartmentDefinition) error {
	cd.ID = uuid.New()
	if cd.FHIRID == "" {
		cd.FHIRID = cd.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO compartment_definition (id, fhir_id, status, url, name, description, publisher, date,
			code, search, resource_type, resource_param)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		cd.ID, cd.FHIRID, cd.Status, cd.URL, cd.Name, cd.Description, cd.Publisher, cd.Date,
		cd.Code, cd.Search, cd.ResourceType, cd.ResourceParam)
	return err
}

func (r *compartmentDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CompartmentDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+cdCols+` FROM compartment_definition WHERE id = $1`, id))
}

func (r *compartmentDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CompartmentDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+cdCols+` FROM compartment_definition WHERE fhir_id = $1`, fhirID))
}

func (r *compartmentDefinitionRepoPG) Update(ctx context.Context, cd *CompartmentDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE compartment_definition SET status=$2, url=$3, name=$4, description=$5, publisher=$6, date=$7,
			code=$8, search=$9, resource_type=$10, resource_param=$11, updated_at=NOW()
		WHERE id = $1`,
		cd.ID, cd.Status, cd.URL, cd.Name, cd.Description, cd.Publisher, cd.Date,
		cd.Code, cd.Search, cd.ResourceType, cd.ResourceParam)
	return err
}

func (r *compartmentDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM compartment_definition WHERE id = $1`, id)
	return err
}

func (r *compartmentDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*CompartmentDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM compartment_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cdCols+` FROM compartment_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CompartmentDefinition
	for rows.Next() {
		cd, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cd)
	}
	return items, total, nil
}

func (r *compartmentDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CompartmentDefinition, int, error) {
	query := `SELECT ` + cdCols + ` FROM compartment_definition WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compartment_definition WHERE 1=1`
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
	var items []*CompartmentDefinition
	for rows.Next() {
		cd, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cd)
	}
	return items, total, nil
}
