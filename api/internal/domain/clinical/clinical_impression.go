package clinical

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
)

// ClinicalImpression maps to the clinical_impression table (FHIR ClinicalImpression resource).
type ClinicalImpression struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Status           string     `db:"status" json:"status"`
	StatusReason     *string    `db:"status_reason" json:"status_reason,omitempty"`
	CodeCode         *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay      *string    `db:"code_display" json:"code_display,omitempty"`
	Description      *string    `db:"description" json:"description,omitempty"`
	SubjectPatientID uuid.UUID  `db:"subject_patient_id" json:"subject_patient_id"`
	EncounterID      *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	EffectiveDate    *time.Time `db:"effective_date" json:"effective_date,omitempty"`
	Date             *time.Time `db:"date" json:"date,omitempty"`
	AssessorID       *uuid.UUID `db:"assessor_id" json:"assessor_id,omitempty"`
	Summary          *string    `db:"summary" json:"summary,omitempty"`
	PrognosisCode    *string    `db:"prognosis_code" json:"prognosis_code,omitempty"`
	PrognosisDisplay *string    `db:"prognosis_display" json:"prognosis_display,omitempty"`
	Note             *string    `db:"note" json:"note,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (ci *ClinicalImpression) GetVersionID() int  { return ci.VersionID }
func (ci *ClinicalImpression) SetVersionID(v int) { ci.VersionID = v }

func (ci *ClinicalImpression) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ClinicalImpression",
		"id":           ci.FHIRID,
		"status":       ci.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", ci.SubjectPatientID.String())},
		"meta":         fhir.Meta{LastUpdated: ci.UpdatedAt},
	}
	if ci.StatusReason != nil {
		result["statusReason"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Display: *ci.StatusReason}},
		}
	}
	if ci.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    *ci.CodeCode,
				Display: strVal(ci.CodeDisplay),
			}},
		}
	}
	if ci.Description != nil {
		result["description"] = *ci.Description
	}
	if ci.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", ci.EncounterID.String())}
	}
	if ci.EffectiveDate != nil {
		result["effectiveDateTime"] = ci.EffectiveDate.Format(time.RFC3339)
	}
	if ci.Date != nil {
		result["date"] = ci.Date.Format(time.RFC3339)
	}
	if ci.AssessorID != nil {
		result["assessor"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", ci.AssessorID.String())}
	}
	if ci.Summary != nil {
		result["summary"] = *ci.Summary
	}
	if ci.PrognosisCode != nil {
		result["prognosisCodeableConcept"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *ci.PrognosisCode, Display: strVal(ci.PrognosisDisplay)}},
		}}
	}
	if ci.Note != nil {
		result["note"] = []map[string]string{{"text": *ci.Note}}
	}
	return result
}

// ClinicalImpressionRepository defines the repository interface.
type ClinicalImpressionRepository interface {
	Create(ctx context.Context, ci *ClinicalImpression) error
	GetByID(ctx context.Context, id uuid.UUID) (*ClinicalImpression, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ClinicalImpression, error)
	Update(ctx context.Context, ci *ClinicalImpression) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ClinicalImpression, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ClinicalImpression, int, error)
}

// -- Postgres implementation --
type clinicalImpressionRepoPG struct{ pool *pgxpool.Pool }

func NewClinicalImpressionRepoPG(pool *pgxpool.Pool) ClinicalImpressionRepository {
	return &clinicalImpressionRepoPG{pool: pool}
}

func (r *clinicalImpressionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const clinicalImpressionCols = `id, fhir_id, status, status_reason, code_code, code_display,
	description, subject_patient_id, encounter_id, effective_date, date, assessor_id,
	summary, prognosis_code, prognosis_display, note,
	version_id, created_at, updated_at`

func (r *clinicalImpressionRepoPG) scanClinicalImpression(row pgx.Row) (*ClinicalImpression, error) {
	var ci ClinicalImpression
	err := row.Scan(&ci.ID, &ci.FHIRID, &ci.Status, &ci.StatusReason, &ci.CodeCode, &ci.CodeDisplay,
		&ci.Description, &ci.SubjectPatientID, &ci.EncounterID, &ci.EffectiveDate, &ci.Date, &ci.AssessorID,
		&ci.Summary, &ci.PrognosisCode, &ci.PrognosisDisplay, &ci.Note,
		&ci.VersionID, &ci.CreatedAt, &ci.UpdatedAt)
	return &ci, err
}

func (r *clinicalImpressionRepoPG) Create(ctx context.Context, ci *ClinicalImpression) error {
	ci.ID = uuid.New()
	if ci.FHIRID == "" {
		ci.FHIRID = ci.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO clinical_impression (id, fhir_id, status, status_reason, code_code, code_display,
			description, subject_patient_id, encounter_id, effective_date, date, assessor_id,
			summary, prognosis_code, prognosis_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		ci.ID, ci.FHIRID, ci.Status, ci.StatusReason, ci.CodeCode, ci.CodeDisplay,
		ci.Description, ci.SubjectPatientID, ci.EncounterID, ci.EffectiveDate, ci.Date, ci.AssessorID,
		ci.Summary, ci.PrognosisCode, ci.PrognosisDisplay, ci.Note)
	return err
}

func (r *clinicalImpressionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ClinicalImpression, error) {
	return r.scanClinicalImpression(r.conn(ctx).QueryRow(ctx, `SELECT `+clinicalImpressionCols+` FROM clinical_impression WHERE id = $1`, id))
}

func (r *clinicalImpressionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ClinicalImpression, error) {
	return r.scanClinicalImpression(r.conn(ctx).QueryRow(ctx, `SELECT `+clinicalImpressionCols+` FROM clinical_impression WHERE fhir_id = $1`, fhirID))
}

func (r *clinicalImpressionRepoPG) Update(ctx context.Context, ci *ClinicalImpression) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE clinical_impression SET status=$2, status_reason=$3, summary=$4,
			prognosis_code=$5, prognosis_display=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		ci.ID, ci.Status, ci.StatusReason, ci.Summary,
		ci.PrognosisCode, ci.PrognosisDisplay, ci.Note)
	return err
}

func (r *clinicalImpressionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM clinical_impression WHERE id = $1`, id)
	return err
}

func (r *clinicalImpressionRepoPG) List(ctx context.Context, limit, offset int) ([]*ClinicalImpression, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM clinical_impression`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+clinicalImpressionCols+` FROM clinical_impression ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ClinicalImpression
	for rows.Next() {
		ci, err := r.scanClinicalImpression(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ci)
	}
	return items, total, nil
}

func (r *clinicalImpressionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ClinicalImpression, int, error) {
	query := `SELECT ` + clinicalImpressionCols + ` FROM clinical_impression WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM clinical_impression WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND subject_patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND subject_patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
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
	var items []*ClinicalImpression
	for rows.Next() {
		ci, err := r.scanClinicalImpression(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ci)
	}
	return items, total, nil
}
