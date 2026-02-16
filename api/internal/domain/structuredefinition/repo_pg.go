package structuredefinition

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

type structureDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewStructureDefinitionRepoPG(pool *pgxpool.Pool) StructureDefinitionRepository {
	return &structureDefinitionRepoPG{pool: pool}
}

func (r *structureDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const sdCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	kind, abstract, type, base_definition, derivation, context_type,
	version_id, created_at, updated_at`

func (r *structureDefinitionRepoPG) scanRow(row pgx.Row) (*StructureDefinition, error) {
	var s StructureDefinition
	err := row.Scan(&s.ID, &s.FHIRID, &s.Status, &s.URL, &s.Name, &s.Title, &s.Description, &s.Publisher, &s.Date,
		&s.Kind, &s.Abstract, &s.Type, &s.BaseDefinition, &s.Derivation, &s.ContextType,
		&s.VersionID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *structureDefinitionRepoPG) Create(ctx context.Context, s *StructureDefinition) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO structure_definition (id, fhir_id, status, url, name, title, description, publisher, date,
			kind, abstract, type, base_definition, derivation, context_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.FHIRID, s.Status, s.URL, s.Name, s.Title, s.Description, s.Publisher, s.Date,
		s.Kind, s.Abstract, s.Type, s.BaseDefinition, s.Derivation, s.ContextType)
	return err
}

func (r *structureDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*StructureDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+sdCols+` FROM structure_definition WHERE id = $1`, id))
}

func (r *structureDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*StructureDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+sdCols+` FROM structure_definition WHERE fhir_id = $1`, fhirID))
}

func (r *structureDefinitionRepoPG) Update(ctx context.Context, s *StructureDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE structure_definition SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			kind=$9, abstract=$10, type=$11, base_definition=$12, derivation=$13, context_type=$14, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.URL, s.Name, s.Title, s.Description, s.Publisher, s.Date,
		s.Kind, s.Abstract, s.Type, s.BaseDefinition, s.Derivation, s.ContextType)
	return err
}

func (r *structureDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM structure_definition WHERE id = $1`, id)
	return err
}

func (r *structureDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*StructureDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM structure_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+sdCols+` FROM structure_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*StructureDefinition
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

var structureDefinitionSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamURI, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
	"type":   {Type: fhir.SearchParamToken, Column: "type"},
	"kind":   {Type: fhir.SearchParamToken, Column: "kind"},
	"base":   {Type: fhir.SearchParamURI, Column: "base_definition"},
}

func (r *structureDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*StructureDefinition, int, error) {
	qb := fhir.NewSearchQuery("structure_definition", sdCols)
	qb.ApplyParams(params, structureDefinitionSearchParams)
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
	var items []*StructureDefinition
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}
