package effectevidencesynthesis

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

type effectEvidenceSynthesisRepoPG struct{ pool *pgxpool.Pool }

func NewEffectEvidenceSynthesisRepoPG(pool *pgxpool.Pool) EffectEvidenceSynthesisRepository {
	return &effectEvidenceSynthesisRepoPG{pool: pool}
}

func (r *effectEvidenceSynthesisRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const eesCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	population_reference, exposure_reference, outcome_reference,
	sample_size_description, result_by_exposure_description,
	version_id, created_at, updated_at`

func (r *effectEvidenceSynthesisRepoPG) scanRow(row pgx.Row) (*EffectEvidenceSynthesis, error) {
	var e EffectEvidenceSynthesis
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.URL, &e.Name, &e.Title, &e.Description, &e.Publisher, &e.Date,
		&e.PopulationReference, &e.ExposureReference, &e.OutcomeReference,
		&e.SampleSizeDescription, &e.ResultByExposureDescription,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *effectEvidenceSynthesisRepoPG) Create(ctx context.Context, e *EffectEvidenceSynthesis) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO effect_evidence_synthesis (id, fhir_id, status, url, name, title, description, publisher, date,
			population_reference, exposure_reference, outcome_reference,
			sample_size_description, result_by_exposure_description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		e.ID, e.FHIRID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.PopulationReference, e.ExposureReference, e.OutcomeReference,
		e.SampleSizeDescription, e.ResultByExposureDescription)
	return err
}

func (r *effectEvidenceSynthesisRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*EffectEvidenceSynthesis, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+eesCols+` FROM effect_evidence_synthesis WHERE id = $1`, id))
}

func (r *effectEvidenceSynthesisRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*EffectEvidenceSynthesis, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+eesCols+` FROM effect_evidence_synthesis WHERE fhir_id = $1`, fhirID))
}

func (r *effectEvidenceSynthesisRepoPG) Update(ctx context.Context, e *EffectEvidenceSynthesis) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE effect_evidence_synthesis SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			population_reference=$9, exposure_reference=$10, outcome_reference=$11,
			sample_size_description=$12, result_by_exposure_description=$13, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.PopulationReference, e.ExposureReference, e.OutcomeReference,
		e.SampleSizeDescription, e.ResultByExposureDescription)
	return err
}

func (r *effectEvidenceSynthesisRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM effect_evidence_synthesis WHERE id = $1`, id)
	return err
}

func (r *effectEvidenceSynthesisRepoPG) List(ctx context.Context, limit, offset int) ([]*EffectEvidenceSynthesis, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM effect_evidence_synthesis`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+eesCols+` FROM effect_evidence_synthesis ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*EffectEvidenceSynthesis
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

var eesSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamToken, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *effectEvidenceSynthesisRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EffectEvidenceSynthesis, int, error) {
	qb := fhir.NewSearchQuery("effect_evidence_synthesis", eesCols)
	qb.ApplyParams(params, eesSearchParams)
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
	var items []*EffectEvidenceSynthesis
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
