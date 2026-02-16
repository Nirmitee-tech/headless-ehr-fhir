package medproductcontraindication

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

type mpcRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductContraindicationRepoPG(pool *pgxpool.Pool) MedicinalProductContraindicationRepository {
	return &mpcRepoPG{pool: pool}
}

func (r *mpcRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mpcCols = `id, fhir_id, subject_reference, disease_code, disease_display,
	disease_status_code, disease_status_display, comorbidity_code, comorbidity_display,
	therapeutic_indication_reference, population_age_low, population_age_high, population_gender_code,
	version_id, created_at, updated_at`

func (r *mpcRepoPG) scanRow(row pgx.Row) (*MedicinalProductContraindication, error) {
	var m MedicinalProductContraindication
	err := row.Scan(&m.ID, &m.FHIRID, &m.SubjectReference, &m.DiseaseCode, &m.DiseaseDisplay,
		&m.DiseaseStatusCode, &m.DiseaseStatusDisplay, &m.ComorbidityCode, &m.ComorbidityDisplay,
		&m.TherapeuticIndicationRef, &m.PopulationAgeLow, &m.PopulationAgeHigh, &m.PopulationGenderCode,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpcRepoPG) Create(ctx context.Context, m *MedicinalProductContraindication) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_contraindication (id, fhir_id, subject_reference, disease_code, disease_display,
			disease_status_code, disease_status_display, comorbidity_code, comorbidity_display,
			therapeutic_indication_reference, population_age_low, population_age_high, population_gender_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.FHIRID, m.SubjectReference, m.DiseaseCode, m.DiseaseDisplay,
		m.DiseaseStatusCode, m.DiseaseStatusDisplay, m.ComorbidityCode, m.ComorbidityDisplay,
		m.TherapeuticIndicationRef, m.PopulationAgeLow, m.PopulationAgeHigh, m.PopulationGenderCode)
	return err
}

func (r *mpcRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductContraindication, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpcCols+` FROM medicinal_product_contraindication WHERE id = $1`, id))
}
func (r *mpcRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductContraindication, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpcCols+` FROM medicinal_product_contraindication WHERE fhir_id = $1`, fhirID))
}

func (r *mpcRepoPG) Update(ctx context.Context, m *MedicinalProductContraindication) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_contraindication SET subject_reference=$2, disease_code=$3, disease_display=$4,
			disease_status_code=$5, disease_status_display=$6, comorbidity_code=$7, comorbidity_display=$8,
			therapeutic_indication_reference=$9, population_age_low=$10, population_age_high=$11, population_gender_code=$12, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SubjectReference, m.DiseaseCode, m.DiseaseDisplay,
		m.DiseaseStatusCode, m.DiseaseStatusDisplay, m.ComorbidityCode, m.ComorbidityDisplay,
		m.TherapeuticIndicationRef, m.PopulationAgeLow, m.PopulationAgeHigh, m.PopulationGenderCode)
	return err
}

func (r *mpcRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_contraindication WHERE id = $1`, id)
	return err
}

func (r *mpcRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductContraindication, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_contraindication`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpcCols+` FROM medicinal_product_contraindication ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var items []*MedicinalProductContraindication
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

var mpcSearchParams = map[string]fhir.SearchParamConfig{
	"subject": {Type: fhir.SearchParamReference, Column: "subject_reference"},
	"disease": {Type: fhir.SearchParamToken, Column: "disease_code"},
}

func (r *mpcRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductContraindication, int, error) {
	qb := fhir.NewSearchQuery("medicinal_product_contraindication", mpcCols)
	qb.ApplyParams(params, mpcSearchParams)
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
	var items []*MedicinalProductContraindication
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
