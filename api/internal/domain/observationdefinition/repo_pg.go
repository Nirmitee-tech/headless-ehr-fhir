package observationdefinition

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

type observationDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewObservationDefinitionRepoPG(pool *pgxpool.Pool) ObservationDefinitionRepository {
	return &observationDefinitionRepoPG{pool: pool}
}

func (r *observationDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const odCols = `id, fhir_id, status, category_code, category_display,
	code_code, code_system, code_display,
	permitted_data_type, multiple_results_allowed,
	method_code, method_display, preferred_report_name,
	unit_code, unit_display,
	normal_value_low, normal_value_high,
	version_id, created_at, updated_at`

func (r *observationDefinitionRepoPG) scanRow(row pgx.Row) (*ObservationDefinition, error) {
	var od ObservationDefinition
	err := row.Scan(&od.ID, &od.FHIRID, &od.Status, &od.CategoryCode, &od.CategoryDisplay,
		&od.CodeCode, &od.CodeSystem, &od.CodeDisplay,
		&od.PermittedDataType, &od.MultipleResultsAllowed,
		&od.MethodCode, &od.MethodDisplay, &od.PreferredReportName,
		&od.UnitCode, &od.UnitDisplay,
		&od.NormalValueLow, &od.NormalValueHigh,
		&od.VersionID, &od.CreatedAt, &od.UpdatedAt)
	return &od, err
}

func (r *observationDefinitionRepoPG) Create(ctx context.Context, od *ObservationDefinition) error {
	od.ID = uuid.New()
	if od.FHIRID == "" {
		od.FHIRID = od.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO observation_definition (id, fhir_id, status, category_code, category_display,
			code_code, code_system, code_display,
			permitted_data_type, multiple_results_allowed,
			method_code, method_display, preferred_report_name,
			unit_code, unit_display,
			normal_value_low, normal_value_high)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		od.ID, od.FHIRID, od.Status, od.CategoryCode, od.CategoryDisplay,
		od.CodeCode, od.CodeSystem, od.CodeDisplay,
		od.PermittedDataType, od.MultipleResultsAllowed,
		od.MethodCode, od.MethodDisplay, od.PreferredReportName,
		od.UnitCode, od.UnitDisplay,
		od.NormalValueLow, od.NormalValueHigh)
	return err
}

func (r *observationDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ObservationDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+odCols+` FROM observation_definition WHERE id = $1`, id))
}

func (r *observationDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ObservationDefinition, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+odCols+` FROM observation_definition WHERE fhir_id = $1`, fhirID))
}

func (r *observationDefinitionRepoPG) Update(ctx context.Context, od *ObservationDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE observation_definition SET status=$2, category_code=$3, category_display=$4,
			code_code=$5, code_system=$6, code_display=$7,
			permitted_data_type=$8, multiple_results_allowed=$9,
			method_code=$10, method_display=$11, preferred_report_name=$12,
			unit_code=$13, unit_display=$14,
			normal_value_low=$15, normal_value_high=$16, updated_at=NOW()
		WHERE id = $1`,
		od.ID, od.Status, od.CategoryCode, od.CategoryDisplay,
		od.CodeCode, od.CodeSystem, od.CodeDisplay,
		od.PermittedDataType, od.MultipleResultsAllowed,
		od.MethodCode, od.MethodDisplay, od.PreferredReportName,
		od.UnitCode, od.UnitDisplay,
		od.NormalValueLow, od.NormalValueHigh)
	return err
}

func (r *observationDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM observation_definition WHERE id = $1`, id)
	return err
}

func (r *observationDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*ObservationDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM observation_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+odCols+` FROM observation_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ObservationDefinition
	for rows.Next() {
		od, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, od)
	}
	return items, total, nil
}

var odSearchParams = map[string]fhir.SearchParamConfig{
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"code":     {Type: fhir.SearchParamToken, Column: "code_code"},
	"category": {Type: fhir.SearchParamToken, Column: "category_code"},
}

func (r *observationDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ObservationDefinition, int, error) {
	qb := fhir.NewSearchQuery("observation_definition", odCols)
	qb.ApplyParams(params, odSearchParams)
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
	var items []*ObservationDefinition
	for rows.Next() {
		od, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, od)
	}
	return items, total, nil
}
