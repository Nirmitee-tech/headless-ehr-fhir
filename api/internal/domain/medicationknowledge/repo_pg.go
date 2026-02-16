package medicationknowledge

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

type medicationKnowledgeRepoPG struct{ pool *pgxpool.Pool }

func NewMedicationKnowledgeRepoPG(pool *pgxpool.Pool) MedicationKnowledgeRepository {
	return &medicationKnowledgeRepoPG{pool: pool}
}

func (r *medicationKnowledgeRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mkCols = `id, fhir_id, status, code_code, code_system, code_display,
	manufacturer_id, dose_form_code, dose_form_display,
	amount_value, amount_unit, synonym, description,
	version_id, created_at, updated_at`

func (r *medicationKnowledgeRepoPG) scanRow(row pgx.Row) (*MedicationKnowledge, error) {
	var m MedicationKnowledge
	err := row.Scan(&m.ID, &m.FHIRID, &m.Status, &m.CodeCode, &m.CodeSystem, &m.CodeDisplay,
		&m.ManufacturerID, &m.DoseFormCode, &m.DoseFormDisplay,
		&m.AmountValue, &m.AmountUnit, &m.Synonym, &m.Description,
		&m.VersionID, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *medicationKnowledgeRepoPG) Create(ctx context.Context, m *MedicationKnowledge) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_knowledge (id, fhir_id, status, code_code, code_system, code_display,
			manufacturer_id, dose_form_code, dose_form_display,
			amount_value, amount_unit, synonym, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.FHIRID, m.Status, m.CodeCode, m.CodeSystem, m.CodeDisplay,
		m.ManufacturerID, m.DoseFormCode, m.DoseFormDisplay,
		m.AmountValue, m.AmountUnit, m.Synonym, m.Description)
	return err
}

func (r *medicationKnowledgeRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicationKnowledge, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mkCols+` FROM medication_knowledge WHERE id = $1`, id))
}

func (r *medicationKnowledgeRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicationKnowledge, error) {
	return r.scanRow(r.conn(ctx).QueryRow(ctx, `SELECT `+mkCols+` FROM medication_knowledge WHERE fhir_id = $1`, fhirID))
}

func (r *medicationKnowledgeRepoPG) Update(ctx context.Context, m *MedicationKnowledge) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication_knowledge SET status=$2, code_code=$3, code_system=$4, code_display=$5,
			manufacturer_id=$6, dose_form_code=$7, dose_form_display=$8,
			amount_value=$9, amount_unit=$10, synonym=$11, description=$12, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.CodeCode, m.CodeSystem, m.CodeDisplay,
		m.ManufacturerID, m.DoseFormCode, m.DoseFormDisplay,
		m.AmountValue, m.AmountUnit, m.Synonym, m.Description)
	return err
}

func (r *medicationKnowledgeRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_knowledge WHERE id = $1`, id)
	return err
}

func (r *medicationKnowledgeRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicationKnowledge, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_knowledge`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mkCols+` FROM medication_knowledge ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationKnowledge
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *medicationKnowledgeRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationKnowledge, int, error) {
	query := `SELECT ` + mkCols + ` FROM medication_knowledge WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM medication_knowledge WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["code"]; ok {
		query += fmt.Sprintf(` AND code_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND code_code = $%d`, idx)
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
	var items []*MedicationKnowledge
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}
