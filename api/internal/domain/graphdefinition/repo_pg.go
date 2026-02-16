package graphdefinition

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

type graphDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewGraphDefinitionRepoPG(pool *pgxpool.Pool) GraphDefinitionRepository {
	return &graphDefinitionRepoPG{pool: pool}
}

func (r *graphDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const gdCols = `id, fhir_id, status, url, name, description, publisher, date,
	start_type, profile,
	version_id, created_at, updated_at`

func (r *graphDefinitionRepoPG) scanRow(row pgx.Row) (*GraphDefinition, error) {
	var g GraphDefinition
	err := row.Scan(&g.ID, &g.FHIRID, &g.Status, &g.URL, &g.Name, &g.Description, &g.Publisher, &g.Date,
		&g.StartType, &g.Profile,
		&g.VersionID, &g.CreatedAt, &g.UpdatedAt)
	return &g, err
}

func (r *graphDefinitionRepoPG) Create(ctx context.Context, g *GraphDefinition) error {
	g.ID = uuid.New()
	if g.FHIRID == "" {
		g.FHIRID = g.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO graph_definition (id, fhir_id, status, url, name, description, publisher, date,
			start_type, profile)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		g.ID, g.FHIRID, g.Status, g.URL, g.Name, g.Description, g.Publisher, g.Date,
		g.StartType, g.Profile)
	return err
}

func (r *graphDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*GraphDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+gdCols+` FROM graph_definition WHERE id = $1`, id))
}

func (r *graphDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*GraphDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+gdCols+` FROM graph_definition WHERE fhir_id = $1`, fhirID))
}

func (r *graphDefinitionRepoPG) Update(ctx context.Context, g *GraphDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE graph_definition SET status=$2, url=$3, name=$4, description=$5, publisher=$6, date=$7,
			start_type=$8, profile=$9, updated_at=NOW()
		WHERE id = $1`,
		g.ID, g.Status, g.URL, g.Name, g.Description, g.Publisher, g.Date,
		g.StartType, g.Profile)
	return err
}

func (r *graphDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM graph_definition WHERE id = $1`, id)
	return err
}

func (r *graphDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*GraphDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM graph_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+gdCols+` FROM graph_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*GraphDefinition
	for rows.Next() {
		g, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, g)
	}
	return items, total, nil
}

var graphDefinitionSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"url":    {Type: fhir.SearchParamToken, Column: "url"},
	"start":  {Type: fhir.SearchParamToken, Column: "start_type"},
}

func (r *graphDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*GraphDefinition, int, error) {
	qb := fhir.NewSearchQuery("graph_definition", gdCols)
	qb.ApplyParams(params, graphDefinitionSearchParams)
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
	var items []*GraphDefinition
	for rows.Next() {
		g, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, g)
	}
	return items, total, nil
}
