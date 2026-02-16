package immunization

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

// =========== Immunization Repository ===========

type immunizationRepoPG struct{ pool *pgxpool.Pool }

func NewImmunizationRepoPG(pool *pgxpool.Pool) ImmunizationRepository {
	return &immunizationRepoPG{pool: pool}
}

func (r *immunizationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const immCols = `id, fhir_id, status, patient_id, encounter_id,
	vaccine_code_system, vaccine_code, vaccine_display,
	occurrence_datetime, occurrence_string, primary_source,
	lot_number, expiration_date, site_code, site_display,
	route_code, route_display, dose_quantity, dose_unit,
	performer_id, note, created_at, updated_at`

func (r *immunizationRepoPG) scanImm(row pgx.Row) (*Immunization, error) {
	var im Immunization
	err := row.Scan(&im.ID, &im.FHIRID, &im.Status, &im.PatientID, &im.EncounterID,
		&im.VaccineCodeSystem, &im.VaccineCode, &im.VaccineDisplay,
		&im.OccurrenceDateTime, &im.OccurrenceString, &im.PrimarySource,
		&im.LotNumber, &im.ExpirationDate, &im.SiteCode, &im.SiteDisplay,
		&im.RouteCode, &im.RouteDisplay, &im.DoseQuantity, &im.DoseUnit,
		&im.PerformerID, &im.Note, &im.CreatedAt, &im.UpdatedAt)
	return &im, err
}

func (r *immunizationRepoPG) Create(ctx context.Context, im *Immunization) error {
	im.ID = uuid.New()
	if im.FHIRID == "" {
		im.FHIRID = im.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO immunization (id, fhir_id, status, patient_id, encounter_id,
			vaccine_code_system, vaccine_code, vaccine_display,
			occurrence_datetime, occurrence_string, primary_source,
			lot_number, expiration_date, site_code, site_display,
			route_code, route_display, dose_quantity, dose_unit,
			performer_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		im.ID, im.FHIRID, im.Status, im.PatientID, im.EncounterID,
		im.VaccineCodeSystem, im.VaccineCode, im.VaccineDisplay,
		im.OccurrenceDateTime, im.OccurrenceString, im.PrimarySource,
		im.LotNumber, im.ExpirationDate, im.SiteCode, im.SiteDisplay,
		im.RouteCode, im.RouteDisplay, im.DoseQuantity, im.DoseUnit,
		im.PerformerID, im.Note)
	return err
}

func (r *immunizationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Immunization, error) {
	return r.scanImm(r.conn(ctx).QueryRow(ctx, `SELECT `+immCols+` FROM immunization WHERE id = $1`, id))
}

func (r *immunizationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Immunization, error) {
	return r.scanImm(r.conn(ctx).QueryRow(ctx, `SELECT `+immCols+` FROM immunization WHERE fhir_id = $1`, fhirID))
}

func (r *immunizationRepoPG) Update(ctx context.Context, im *Immunization) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE immunization SET status=$2, vaccine_code=$3, vaccine_display=$4,
			occurrence_datetime=$5, lot_number=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		im.ID, im.Status, im.VaccineCode, im.VaccineDisplay,
		im.OccurrenceDateTime, im.LotNumber, im.Note)
	return err
}

func (r *immunizationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM immunization WHERE id = $1`, id)
	return err
}

func (r *immunizationRepoPG) List(ctx context.Context, limit, offset int) ([]*Immunization, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM immunization`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+immCols+` FROM immunization ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Immunization
	for rows.Next() {
		im, err := r.scanImm(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, im)
	}
	return items, total, nil
}

func (r *immunizationRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Immunization, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM immunization WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+immCols+` FROM immunization WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Immunization
	for rows.Next() {
		im, err := r.scanImm(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, im)
	}
	return items, total, nil
}

var immunizationSearchParams = map[string]fhir.SearchParamConfig{
	"patient":      {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":       {Type: fhir.SearchParamToken, Column: "status"},
	"vaccine-code": {Type: fhir.SearchParamToken, Column: "vaccine_code"},
}

func (r *immunizationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Immunization, int, error) {
	qb := fhir.NewSearchQuery("immunization", immCols)
	qb.ApplyParams(params, immunizationSearchParams)
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
	var items []*Immunization
	for rows.Next() {
		im, err := r.scanImm(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, im)
	}
	return items, total, nil
}

// =========== Recommendation Repository ===========

type recommendationRepoPG struct{ pool *pgxpool.Pool }

func NewRecommendationRepoPG(pool *pgxpool.Pool) RecommendationRepository {
	return &recommendationRepoPG{pool: pool}
}

func (r *recommendationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const recCols = `id, fhir_id, patient_id, date, vaccine_code, vaccine_display,
	forecast_status, forecast_display, date_criterion, series_doses,
	dose_number, description, created_at, updated_at`

func (r *recommendationRepoPG) scanRec(row pgx.Row) (*ImmunizationRecommendation, error) {
	var rec ImmunizationRecommendation
	err := row.Scan(&rec.ID, &rec.FHIRID, &rec.PatientID, &rec.Date,
		&rec.VaccineCode, &rec.VaccineDisplay,
		&rec.ForecastStatus, &rec.ForecastDisplay, &rec.DateCriterion,
		&rec.SeriesDoses, &rec.DoseNumber, &rec.Description,
		&rec.CreatedAt, &rec.UpdatedAt)
	return &rec, err
}

func (r *recommendationRepoPG) Create(ctx context.Context, rec *ImmunizationRecommendation) error {
	rec.ID = uuid.New()
	if rec.FHIRID == "" {
		rec.FHIRID = rec.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO immunization_recommendation (id, fhir_id, patient_id, date,
			vaccine_code, vaccine_display, forecast_status, forecast_display,
			date_criterion, series_doses, dose_number, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		rec.ID, rec.FHIRID, rec.PatientID, rec.Date,
		rec.VaccineCode, rec.VaccineDisplay, rec.ForecastStatus, rec.ForecastDisplay,
		rec.DateCriterion, rec.SeriesDoses, rec.DoseNumber, rec.Description)
	return err
}

func (r *recommendationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ImmunizationRecommendation, error) {
	return r.scanRec(r.conn(ctx).QueryRow(ctx, `SELECT `+recCols+` FROM immunization_recommendation WHERE id = $1`, id))
}

func (r *recommendationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ImmunizationRecommendation, error) {
	return r.scanRec(r.conn(ctx).QueryRow(ctx, `SELECT `+recCols+` FROM immunization_recommendation WHERE fhir_id = $1`, fhirID))
}

func (r *recommendationRepoPG) Update(ctx context.Context, rec *ImmunizationRecommendation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE immunization_recommendation SET forecast_status=$2, forecast_display=$3,
			date_criterion=$4, series_doses=$5, dose_number=$6, description=$7, updated_at=NOW()
		WHERE id = $1`,
		rec.ID, rec.ForecastStatus, rec.ForecastDisplay,
		rec.DateCriterion, rec.SeriesDoses, rec.DoseNumber, rec.Description)
	return err
}

func (r *recommendationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM immunization_recommendation WHERE id = $1`, id)
	return err
}

func (r *recommendationRepoPG) List(ctx context.Context, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM immunization_recommendation`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+recCols+` FROM immunization_recommendation ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ImmunizationRecommendation
	for rows.Next() {
		rec, err := r.scanRec(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rec)
	}
	return items, total, nil
}

func (r *recommendationRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM immunization_recommendation WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+recCols+` FROM immunization_recommendation WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ImmunizationRecommendation
	for rows.Next() {
		rec, err := r.scanRec(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rec)
	}
	return items, total, nil
}

var recommendationSearchParams = map[string]fhir.SearchParamConfig{
	"patient":      {Type: fhir.SearchParamReference, Column: "patient_id"},
	"vaccine-type": {Type: fhir.SearchParamToken, Column: "vaccine_code"},
}

func (r *recommendationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	qb := fhir.NewSearchQuery("immunization_recommendation", recCols)
	qb.ApplyParams(params, recommendationSearchParams)
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
	var items []*ImmunizationRecommendation
	for rows.Next() {
		rec, err := r.scanRec(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rec)
	}
	return items, total, nil
}
