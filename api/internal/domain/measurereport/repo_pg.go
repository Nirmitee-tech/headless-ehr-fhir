package measurereport

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

type measureReportRepoPG struct{ pool *pgxpool.Pool }

func NewMeasureReportRepoPG(pool *pgxpool.Pool) MeasureReportRepository {
	return &measureReportRepoPG{pool: pool}
}

func (r *measureReportRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mrCols = `id, fhir_id, status, type, measure_url,
	subject_patient_id, date, reporter_org_id,
	period_start, period_end, improvement_notation,
	group_code, group_population_code, group_population_count,
	group_measure_score, version_id, created_at, updated_at`

func (r *measureReportRepoPG) scanRow(row pgx.Row) (*MeasureReport, error) {
	var mr MeasureReport
	err := row.Scan(&mr.ID, &mr.FHIRID, &mr.Status, &mr.Type, &mr.MeasureURL,
		&mr.SubjectPatientID, &mr.Date, &mr.ReporterOrgID,
		&mr.PeriodStart, &mr.PeriodEnd, &mr.ImprovementNotation,
		&mr.GroupCode, &mr.GroupPopulationCode, &mr.GroupPopulationCount,
		&mr.GroupMeasureScore, &mr.VersionID, &mr.CreatedAt, &mr.UpdatedAt)
	return &mr, err
}

func (r *measureReportRepoPG) Create(ctx context.Context, mr *MeasureReport) error {
	mr.ID = uuid.New()
	if mr.FHIRID == "" {
		mr.FHIRID = mr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO measure_report (id, fhir_id, status, type, measure_url,
			subject_patient_id, date, reporter_org_id,
			period_start, period_end, improvement_notation,
			group_code, group_population_code, group_population_count,
			group_measure_score)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		mr.ID, mr.FHIRID, mr.Status, mr.Type, mr.MeasureURL,
		mr.SubjectPatientID, mr.Date, mr.ReporterOrgID,
		mr.PeriodStart, mr.PeriodEnd, mr.ImprovementNotation,
		mr.GroupCode, mr.GroupPopulationCode, mr.GroupPopulationCount,
		mr.GroupMeasureScore)
	return err
}

func (r *measureReportRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MeasureReport, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mrCols+` FROM measure_report WHERE id = $1`, id))
}

func (r *measureReportRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MeasureReport, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mrCols+` FROM measure_report WHERE fhir_id = $1`, fhirID))
}

func (r *measureReportRepoPG) Update(ctx context.Context, mr *MeasureReport) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE measure_report SET status=$2, type=$3, measure_url=$4,
			period_start=$5, period_end=$6, improvement_notation=$7,
			group_code=$8, group_population_code=$9, group_population_count=$10,
			group_measure_score=$11, updated_at=NOW()
		WHERE id = $1`,
		mr.ID, mr.Status, mr.Type, mr.MeasureURL,
		mr.PeriodStart, mr.PeriodEnd, mr.ImprovementNotation,
		mr.GroupCode, mr.GroupPopulationCode, mr.GroupPopulationCount,
		mr.GroupMeasureScore)
	return err
}

func (r *measureReportRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM measure_report WHERE id = $1`, id)
	return err
}

func (r *measureReportRepoPG) List(ctx context.Context, limit, offset int) ([]*MeasureReport, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM measure_report`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mrCols+` FROM measure_report ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MeasureReport
	for rows.Next() {
		mr, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mr)
	}
	return items, total, nil
}

var measureReportSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *measureReportRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MeasureReport, int, error) {
	qb := fhir.NewSearchQuery("measure_report", mrCols)
	qb.ApplyParams(params, measureReportSearchParams)
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
	var items []*MeasureReport
	for rows.Next() {
		mr, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mr)
	}
	return items, total, nil
}
