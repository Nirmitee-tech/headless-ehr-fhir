package riskevidencesynthesis

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

type riskEvidenceSynthesisRepoPG struct{ pool *pgxpool.Pool }

func NewRiskEvidenceSynthesisRepoPG(pool *pgxpool.Pool) RiskEvidenceSynthesisRepository {
	return &riskEvidenceSynthesisRepoPG{pool: pool}
}

func (r *riskEvidenceSynthesisRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const resCols = `id, fhir_id, status, url, name, title, description, publisher, date,
	population_reference, outcome_reference,
	sample_size_description, risk_estimate_description, risk_estimate_value,
	version_id, created_at, updated_at`

func (r *riskEvidenceSynthesisRepoPG) scanRow(row pgx.Row) (*RiskEvidenceSynthesis, error) {
	var e RiskEvidenceSynthesis
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.URL, &e.Name, &e.Title, &e.Description, &e.Publisher, &e.Date,
		&e.PopulationReference, &e.OutcomeReference,
		&e.SampleSizeDescription, &e.RiskEstimateDescription, &e.RiskEstimateValue,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *riskEvidenceSynthesisRepoPG) Create(ctx context.Context, e *RiskEvidenceSynthesis) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO risk_evidence_synthesis (id, fhir_id, status, url, name, title, description, publisher, date,
			population_reference, outcome_reference,
			sample_size_description, risk_estimate_description, risk_estimate_value)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		e.ID, e.FHIRID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.PopulationReference, e.OutcomeReference,
		e.SampleSizeDescription, e.RiskEstimateDescription, e.RiskEstimateValue)
	return err
}

func (r *riskEvidenceSynthesisRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*RiskEvidenceSynthesis, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+resCols+` FROM risk_evidence_synthesis WHERE id = $1`, id))
}

func (r *riskEvidenceSynthesisRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*RiskEvidenceSynthesis, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+resCols+` FROM risk_evidence_synthesis WHERE fhir_id = $1`, fhirID))
}

func (r *riskEvidenceSynthesisRepoPG) Update(ctx context.Context, e *RiskEvidenceSynthesis) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE risk_evidence_synthesis SET status=$2, url=$3, name=$4, title=$5, description=$6, publisher=$7, date=$8,
			population_reference=$9, outcome_reference=$10,
			sample_size_description=$11, risk_estimate_description=$12, risk_estimate_value=$13, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.URL, e.Name, e.Title, e.Description, e.Publisher, e.Date,
		e.PopulationReference, e.OutcomeReference,
		e.SampleSizeDescription, e.RiskEstimateDescription, e.RiskEstimateValue)
	return err
}

func (r *riskEvidenceSynthesisRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM risk_evidence_synthesis WHERE id = $1`, id)
	return err
}

func (r *riskEvidenceSynthesisRepoPG) List(ctx context.Context, limit, offset int) ([]*RiskEvidenceSynthesis, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM risk_evidence_synthesis`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+resCols+` FROM risk_evidence_synthesis ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RiskEvidenceSynthesis
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

var riskEvidenceSynthesisSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"url":    {Type: fhir.SearchParamURI, Column: "url"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *riskEvidenceSynthesisRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RiskEvidenceSynthesis, int, error) {
	qb := fhir.NewSearchQuery("risk_evidence_synthesis", resCols)
	qb.ApplyParams(params, riskEvidenceSynthesisSearchParams)
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
	var items []*RiskEvidenceSynthesis
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
