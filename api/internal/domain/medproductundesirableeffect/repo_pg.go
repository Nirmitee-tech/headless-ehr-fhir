package medproductundesirableeffect

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

type mpueRepoPG struct{ pool *pgxpool.Pool }

func NewMedicinalProductUndesirableEffectRepoPG(pool *pgxpool.Pool) MedicinalProductUndesirableEffectRepository {
	return &mpueRepoPG{pool: pool}
}

func (r *mpueRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil { return tx }
	if c := db.ConnFromContext(ctx); c != nil { return c }
	return r.pool
}

const mpueCols = `id, fhir_id, subject_reference, symptom_condition_effect_code, symptom_condition_effect_display,
	classification_code, classification_display, frequency_of_occurrence_code, frequency_of_occurrence_display,
	population_age_low, population_age_high, population_gender_code,
	version_id, created_at, updated_at`

func (r *mpueRepoPG) scanRow(row pgx.Row) (*MedicinalProductUndesirableEffect, error) {
	var m MedicinalProductUndesirableEffect
	err := row.Scan(&m.ID, &m.FHIRID, &m.SubjectReference, &m.SymptomConditionEffectCode, &m.SymptomConditionEffectDisplay,
		&m.ClassificationCode, &m.ClassificationDisplay, &m.FrequencyOfOccurrenceCode, &m.FrequencyOfOccurrenceDisplay,
		&m.PopulationAgeLow, &m.PopulationAgeHigh, &m.PopulationGenderCode,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *mpueRepoPG) Create(ctx context.Context, m *MedicinalProductUndesirableEffect) error {
	m.ID = uuid.New()
	if m.FHIRID == "" { m.FHIRID = m.ID.String() }
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medicinal_product_undesirable_effect (id, fhir_id, subject_reference, symptom_condition_effect_code, symptom_condition_effect_display,
			classification_code, classification_display, frequency_of_occurrence_code, frequency_of_occurrence_display,
			population_age_low, population_age_high, population_gender_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FHIRID, m.SubjectReference, m.SymptomConditionEffectCode, m.SymptomConditionEffectDisplay,
		m.ClassificationCode, m.ClassificationDisplay, m.FrequencyOfOccurrenceCode, m.FrequencyOfOccurrenceDisplay,
		m.PopulationAgeLow, m.PopulationAgeHigh, m.PopulationGenderCode)
	return err
}

func (r *mpueRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductUndesirableEffect, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpueCols+` FROM medicinal_product_undesirable_effect WHERE id = $1`, id)) }
func (r *mpueRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductUndesirableEffect, error) { return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mpueCols+` FROM medicinal_product_undesirable_effect WHERE fhir_id = $1`, fhirID)) }

func (r *mpueRepoPG) Update(ctx context.Context, m *MedicinalProductUndesirableEffect) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medicinal_product_undesirable_effect SET subject_reference=$2, symptom_condition_effect_code=$3, symptom_condition_effect_display=$4,
			classification_code=$5, classification_display=$6, frequency_of_occurrence_code=$7, frequency_of_occurrence_display=$8,
			population_age_low=$9, population_age_high=$10, population_gender_code=$11, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.SubjectReference, m.SymptomConditionEffectCode, m.SymptomConditionEffectDisplay,
		m.ClassificationCode, m.ClassificationDisplay, m.FrequencyOfOccurrenceCode, m.FrequencyOfOccurrenceDisplay,
		m.PopulationAgeLow, m.PopulationAgeHigh, m.PopulationGenderCode)
	return err
}

func (r *mpueRepoPG) Delete(ctx context.Context, id uuid.UUID) error { _, err := r.conn(ctx).Exec(ctx, `DELETE FROM medicinal_product_undesirable_effect WHERE id = $1`, id); return err }

func (r *mpueRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicinalProductUndesirableEffect, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medicinal_product_undesirable_effect`).Scan(&total); err != nil { return nil, 0, err }
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mpueCols+` FROM medicinal_product_undesirable_effect ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductUndesirableEffect
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}

func (r *mpueRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductUndesirableEffect, int, error) {
	query := `SELECT ` + mpueCols + ` FROM medicinal_product_undesirable_effect WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medicinal_product_undesirable_effect WHERE 1=1`
	var args []interface{}; idx := 1
	if p, ok := params["subject"]; ok { query += fmt.Sprintf(` AND subject_reference = $%d`, idx); countQuery += fmt.Sprintf(` AND subject_reference = $%d`, idx); args = append(args, p); idx++ }
	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil, 0, err }
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1); args = append(args, limit, offset)
	rows, err := r.conn(ctx).Query(ctx, query, args...); if err != nil { return nil, 0, err }; defer rows.Close()
	var items []*MedicinalProductUndesirableEffect
	for rows.Next() { m, err := r.scanRow(rows); if err != nil { return nil, 0, err }; items = append(items, m) }
	return items, total, nil
}
