package coverageeligibility

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

// -- CoverageEligibilityRequest PG Repo --

type coverageEligibilityRequestRepoPG struct{ pool *pgxpool.Pool }

func NewCoverageEligibilityRequestRepoPG(pool *pgxpool.Pool) CoverageEligibilityRequestRepository {
	return &coverageEligibilityRequestRepoPG{pool: pool}
}

func (r *coverageEligibilityRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const reqCols = `id, fhir_id, status, patient_id, provider_id, insurer_id,
	purpose, serviced_date, created,
	version_id, created_at, updated_at`

func (r *coverageEligibilityRequestRepoPG) scanRow(row pgx.Row) (*CoverageEligibilityRequest, error) {
	var e CoverageEligibilityRequest
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.PatientID, &e.ProviderID, &e.InsurerID,
		&e.Purpose, &e.ServicedDate, &e.Created,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *coverageEligibilityRequestRepoPG) Create(ctx context.Context, e *CoverageEligibilityRequest) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO coverage_eligibility_request (id, fhir_id, status, patient_id, provider_id, insurer_id,
			purpose, serviced_date, created)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.FHIRID, e.Status, e.PatientID, e.ProviderID, e.InsurerID,
		e.Purpose, e.ServicedDate, e.Created)
	return err
}

func (r *coverageEligibilityRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CoverageEligibilityRequest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+reqCols+` FROM coverage_eligibility_request WHERE id = $1`, id))
}

func (r *coverageEligibilityRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CoverageEligibilityRequest, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+reqCols+` FROM coverage_eligibility_request WHERE fhir_id = $1`, fhirID))
}

func (r *coverageEligibilityRequestRepoPG) Update(ctx context.Context, e *CoverageEligibilityRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE coverage_eligibility_request SET status=$2, patient_id=$3, provider_id=$4, insurer_id=$5,
			purpose=$6, serviced_date=$7, created=$8, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.PatientID, e.ProviderID, e.InsurerID,
		e.Purpose, e.ServicedDate, e.Created)
	return err
}

func (r *coverageEligibilityRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM coverage_eligibility_request WHERE id = $1`, id)
	return err
}

func (r *coverageEligibilityRequestRepoPG) List(ctx context.Context, limit, offset int) ([]*CoverageEligibilityRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM coverage_eligibility_request`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+reqCols+` FROM coverage_eligibility_request ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CoverageEligibilityRequest
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

var cerSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"purpose": {Type: fhir.SearchParamToken, Column: "purpose"},
}

func (r *coverageEligibilityRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CoverageEligibilityRequest, int, error) {
	qb := fhir.NewSearchQuery("coverage_eligibility_request", reqCols)
	qb.ApplyParams(params, cerSearchParams)
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
	var items []*CoverageEligibilityRequest
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

// -- CoverageEligibilityResponse PG Repo --

type coverageEligibilityResponseRepoPG struct{ pool *pgxpool.Pool }

func NewCoverageEligibilityResponseRepoPG(pool *pgxpool.Pool) CoverageEligibilityResponseRepository {
	return &coverageEligibilityResponseRepoPG{pool: pool}
}

func (r *coverageEligibilityResponseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const respCols = `id, fhir_id, status, patient_id, request_id, insurer_id,
	outcome, disposition, created,
	version_id, created_at, updated_at`

func (r *coverageEligibilityResponseRepoPG) scanRow(row pgx.Row) (*CoverageEligibilityResponse, error) {
	var e CoverageEligibilityResponse
	err := row.Scan(&e.ID, &e.FHIRID, &e.Status, &e.PatientID, &e.RequestID, &e.InsurerID,
		&e.Outcome, &e.Disposition, &e.Created,
		&e.VersionID, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *coverageEligibilityResponseRepoPG) Create(ctx context.Context, e *CoverageEligibilityResponse) error {
	e.ID = uuid.New()
	if e.FHIRID == "" {
		e.FHIRID = e.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO coverage_eligibility_response (id, fhir_id, status, patient_id, request_id, insurer_id,
			outcome, disposition, created)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.FHIRID, e.Status, e.PatientID, e.RequestID, e.InsurerID,
		e.Outcome, e.Disposition, e.Created)
	return err
}

func (r *coverageEligibilityResponseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CoverageEligibilityResponse, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+respCols+` FROM coverage_eligibility_response WHERE id = $1`, id))
}

func (r *coverageEligibilityResponseRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*CoverageEligibilityResponse, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+respCols+` FROM coverage_eligibility_response WHERE fhir_id = $1`, fhirID))
}

func (r *coverageEligibilityResponseRepoPG) Update(ctx context.Context, e *CoverageEligibilityResponse) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE coverage_eligibility_response SET status=$2, patient_id=$3, request_id=$4, insurer_id=$5,
			outcome=$6, disposition=$7, created=$8, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.PatientID, e.RequestID, e.InsurerID,
		e.Outcome, e.Disposition, e.Created)
	return err
}

func (r *coverageEligibilityResponseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM coverage_eligibility_response WHERE id = $1`, id)
	return err
}

func (r *coverageEligibilityResponseRepoPG) List(ctx context.Context, limit, offset int) ([]*CoverageEligibilityResponse, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM coverage_eligibility_response`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+respCols+` FROM coverage_eligibility_response ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CoverageEligibilityResponse
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

var cerspSearchParams = map[string]fhir.SearchParamConfig{
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"outcome": {Type: fhir.SearchParamToken, Column: "outcome"},
	"request": {Type: fhir.SearchParamReference, Column: "request_id"},
}

func (r *coverageEligibilityResponseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CoverageEligibilityResponse, int, error) {
	qb := fhir.NewSearchQuery("coverage_eligibility_response", respCols)
	qb.ApplyParams(params, cerspSearchParams)
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
	var items []*CoverageEligibilityResponse
	for rows.Next() {
		e, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}
