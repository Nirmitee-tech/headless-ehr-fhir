package researchdefinition

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

type researchDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewResearchDefinitionRepoPG(pool *pgxpool.Pool) ResearchDefinitionRepository {
	return &researchDefinitionRepoPG{pool: pool}
}

func (r *researchDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const rdCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	population_reference, exposure_reference, outcome_reference,
	version_id, created_at, updated_at`

func (r *researchDefinitionRepoPG) scanRow(row pgx.Row) (*ResearchDefinition, error) {
	var e ResearchDefinition
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.URL, &e.Name, &e.Title, &e.Description, &e.Publisher, &e.Date,
		&e.PopulationReference, &e.ExposureReference, &e.OutcomeReference,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *researchDefinitionRepoPG) Create(ctx context.Context, e *ResearchDefinition) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_definition (id, fhir_id, status, url, name, title, description, publisher, date,
			population_reference, exposure_reference, outcome_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		e.ID, e.FHIRID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.PopulationReference, e.ExposureReference, e.OutcomeReference)
	return err
}

func (r *researchDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ResearchDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+rdCols+` FROM research_definition WHERE id = $1`, id))
}

func (r *researchDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ResearchDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+rdCols+` FROM research_definition WHERE fhir_id = $1`, fhirID))
}

func (r *researchDefinitionRepoPG) Update(ctx context.Context, e *ResearchDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE research_definition SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			population_reference=$9, exposure_reference=$10, outcome_reference=$11, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.PopulationReference, e.ExposureReference, e.OutcomeReference)
	return err
}

func (r *researchDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM research_definition WHERE id = $1`, id)
	return err
}

func (r *researchDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*ResearchDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+rdCols+` FROM research_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchDefinition
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *researchDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchDefinition, int, error) {
	query := `SELECT ` + rdCols + ` FROM research_definition WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM research_definition WHERE 1=1`
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
	var items []*ResearchDefinition
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
