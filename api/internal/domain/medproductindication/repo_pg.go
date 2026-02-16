package medproductindication

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

type mpiRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductIndicationRepoPG(pool *pgxpool.Pool) MedicinalProductIndicationRepository {
	return &mpiRepoPG{pool: pool}
}

func (r *mpiRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mpiCols = `id, fhir_id, subject_reference, disease_symptom_procedure_code, disease_symptom_procedure_display,
	disease_status_code, disease_status_display, comorbidity_code, comorbidity_display,
	intended_effect_code, intended_effect_display, duration_value, duration_unit,
	undesirable_effect_reference, population_age_low, population_age_high, population_gender_code,
	version_id, created_at, updated_at`

func (r *mpiRepoPG) scanRow(row pgx.Row) (*MedicinalProductIndication, error) {
	var m MedicinalProductIndication
	err := row.Scan(&m.ID, &m.FHIRID, &m.SubjectReference, &m.DiseaseSymptomProcedureCode, &m.DiseaseSymptomProcedureDisplay,
		&m.DiseaseStatusCode, &m.DiseaseStatusDisplay, &m.ComorbidityCode, &m.ComorbidityDisplay,
		&m.IntendedEffectCode, &m.IntendedEffectDisplay, &m.DurationValue, &m.DurationUnit,
		&m.UndesirableEffectReference, &m.PopulationAgeLow, &m.PopulationAgeHigh, &m.PopulationGenderCode,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpiRepoPG) Create(ctx context.Context, m *MedicinalProductIndication) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_indication (id, fhir_id, subject_reference, disease_symptom_procedure_code, disease_symptom_procedure_display,
			disease_status_code, disease_status_display, comorbidity_code, comorbidity_display,
			intended_effect_code, intended_effect_display, duration_value, duration_unit,
			undesirable_effect_reference, population_age_low, population_age_high, population_gender_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		m.ID, m.FHIRID, m.SubjectReference, m.DiseaseSymptomProcedureCode, m.DiseaseSymptomProcedureDisplay,
		m.DiseaseStatusCode, m.DiseaseStatusDisplay, m.ComorbidityCode, m.ComorbidityDisplay,
		m.IntendedEffectCode, m.IntendedEffectDisplay, m.DurationValue, m.DurationUnit,
		m.UndesirableEffectReference, m.PopulationAgeLow, m.PopulationAgeHigh, m.PopulationGenderCode)
	return err
}

func (r *mpiRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductIndication, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpiCols+` FROM medicinal_product_indication WHERE id = $1`, id)) }
func (r *mpiRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductIndication, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpiCols+` FROM medicinal_product_indication WHERE fhir_id = $1`, fhirID)) }

func (r *mpiRepoPG) Update(ctx context.Context, m *MedicinalProductIndication) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_indication SET subject_reference=$2, disease_symptom_procedure_code=$3, disease_symptom_procedure_display=$4,
			disease_status_code=$5, disease_status_display=$6, comorbidity_code=$7, comorbidity_display=$8,
			intended_effect_code=$9, intended_effect_display=$10, duration_value=$11, duration_unit=$12,
			undesirable_effect_reference=$13, population_age_low=$14, population_age_high=$15, population_gender_code=$16, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SubjectReference, m.DiseaseSymptomProcedureCode, m.DiseaseSymptomProcedureDisplay,
		m.DiseaseStatusCode, m.DiseaseStatusDisplay, m.ComorbidityCode, m.ComorbidityDisplay,
		m.IntendedEffectCode, m.IntendedEffectDisplay, m.DurationValue, m.DurationUnit,
		m.UndesirableEffectReference, m.PopulationAgeLow, m.PopulationAgeHigh, m.PopulationGenderCode)
	return err
}

func (r *mpiRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_indication WHERE id = $1`, id); return err }

func (r *mpiRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductIndication, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_indication`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpiCols+` FROM medicinal_product_indication ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductIndication
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

var mpiSearchParams = map[string]fhir.SearchParamConfig{
	"subject": {Type: fhir.SearchParamReference, Column: "subject_reference"},
	"disease": {Type: fhir.SearchParamToken, Column: "disease_symptom_procedure_code"},
}

func (r *mpiRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductIndication, int, error) {
	qb := fhir.NewSearchQuery("medicinal_product_indication", mpiCols)
	qb.ApplyParams(params, mpiSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil { return nil, 0, err }

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductIndication
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
