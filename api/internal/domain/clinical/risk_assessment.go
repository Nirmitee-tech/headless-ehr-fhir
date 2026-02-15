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

// RiskAssessment maps to the risk_assessment table (FHIR RiskAssessment resource).
type RiskAssessment struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	FHIRID                 string     `db:"fhir_id" json:"fhir_id"`
	Status                 string     `db:"status" json:"status"`
	MethodCode             *string    `db:"method_code" json:"method_code,omitempty"`
	MethodDisplay          *string    `db:"method_display" json:"method_display,omitempty"`
	CodeCode               *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay            *string    `db:"code_display" json:"code_display,omitempty"`
	SubjectPatientID       uuid.UUID  `db:"subject_patient_id" json:"subject_patient_id"`
	EncounterID            *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	OccurrenceDate         *time.Time `db:"occurrence_date" json:"occurrence_date,omitempty"`
	PerformerID            *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	PredictionOutcome      *string    `db:"prediction_outcome" json:"prediction_outcome,omitempty"`
	PredictionProbability  *float64   `db:"prediction_probability" json:"prediction_probability,omitempty"`
	PredictionQualitative  *string    `db:"prediction_qualitative" json:"prediction_qualitative,omitempty"`
	Mitigation             *string    `db:"mitigation" json:"mitigation,omitempty"`
	Note                   *string    `db:"note" json:"note,omitempty"`
	VersionID              int        `db:"version_id" json:"version_id"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

func (ra *RiskAssessment) GetVersionID() int  { return ra.VersionID }
func (ra *RiskAssessment) SetVersionID(v int) { ra.VersionID = v }

func (ra *RiskAssessment) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "RiskAssessment",
		"id":           ra.FHIRID,
		"status":       ra.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", ra.SubjectPatientID.String())},
		"meta":         fhir.Meta{LastUpdated: ra.UpdatedAt},
	}
	if ra.MethodCode != nil {
		result["method"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ra.MethodCode, Display: strVal(ra.MethodDisplay)}},
		}
	}
	if ra.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *ra.CodeCode, Display: strVal(ra.CodeDisplay)}},
		}
	}
	if ra.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", ra.EncounterID.String())}
	}
	if ra.OccurrenceDate != nil {
		result["occurrenceDateTime"] = ra.OccurrenceDate.Format(time.RFC3339)
	}
	if ra.PerformerID != nil {
		result["performer"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", ra.PerformerID.String())}
	}
	if ra.PredictionOutcome != nil {
		prediction := map[string]interface{}{
			"outcome": fhir.CodeableConcept{
				Coding: []fhir.Coding{{Display: *ra.PredictionOutcome}},
			},
		}
		if ra.PredictionProbability != nil {
			prediction["probabilityDecimal"] = *ra.PredictionProbability
		}
		if ra.PredictionQualitative != nil {
			prediction["qualitativeRisk"] = fhir.CodeableConcept{
				Coding: []fhir.Coding{{Display: *ra.PredictionQualitative}},
			}
		}
		result["prediction"] = []interface{}{prediction}
	}
	if ra.Mitigation != nil {
		result["mitigation"] = *ra.Mitigation
	}
	if ra.Note != nil {
		result["note"] = []map[string]string{{"text": *ra.Note}}
	}
	return result
}

// RiskAssessmentRepository defines the repository interface.
type RiskAssessmentRepository interface {
	Create(ctx context.Context, ra *RiskAssessment) error
	GetByID(ctx context.Context, id uuid.UUID) (*RiskAssessment, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*RiskAssessment, error)
	Update(ctx context.Context, ra *RiskAssessment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*RiskAssessment, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RiskAssessment, int, error)
}

// -- Postgres implementation --
type riskAssessmentRepoPG struct{ pool *pgxpool.Pool }

func NewRiskAssessmentRepoPG(pool *pgxpool.Pool) RiskAssessmentRepository {
	return &riskAssessmentRepoPG{pool: pool}
}

func (r *riskAssessmentRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const riskAssessmentCols = `id, fhir_id, status, method_code, method_display,
	code_code, code_display, subject_patient_id, encounter_id, occurrence_date, performer_id,
	prediction_outcome, prediction_probability, prediction_qualitative,
	mitigation, note,
	version_id, created_at, updated_at`

func (r *riskAssessmentRepoPG) scanRiskAssessment(row pgx.Row) (*RiskAssessment, error) {
	var ra RiskAssessment
	err := row.Scan(&ra.ID, &ra.FHIRID, &ra.Status, &ra.MethodCode, &ra.MethodDisplay,
		&ra.CodeCode, &ra.CodeDisplay, &ra.SubjectPatientID, &ra.EncounterID, &ra.OccurrenceDate, &ra.PerformerID,
		&ra.PredictionOutcome, &ra.PredictionProbability, &ra.PredictionQualitative,
		&ra.Mitigation, &ra.Note,
		&ra.VersionID, &ra.CreatedAt, &ra.UpdatedAt)
	return &ra, err
}

func (r *riskAssessmentRepoPG) Create(ctx context.Context, ra *RiskAssessment) error {
	ra.ID = uuid.New()
	if ra.FHIRID == "" {
		ra.FHIRID = ra.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO risk_assessment (id, fhir_id, status, method_code, method_display,
			code_code, code_display, subject_patient_id, encounter_id, occurrence_date, performer_id,
			prediction_outcome, prediction_probability, prediction_qualitative,
			mitigation, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		ra.ID, ra.FHIRID, ra.Status, ra.MethodCode, ra.MethodDisplay,
		ra.CodeCode, ra.CodeDisplay, ra.SubjectPatientID, ra.EncounterID, ra.OccurrenceDate, ra.PerformerID,
		ra.PredictionOutcome, ra.PredictionProbability, ra.PredictionQualitative,
		ra.Mitigation, ra.Note)
	return err
}

func (r *riskAssessmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*RiskAssessment, error) {
	return r.scanRiskAssessment(r.conn(ctx).QueryRow(ctx, `SELECT `+riskAssessmentCols+` FROM risk_assessment WHERE id = $1`, id))
}

func (r *riskAssessmentRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*RiskAssessment, error) {
	return r.scanRiskAssessment(r.conn(ctx).QueryRow(ctx, `SELECT `+riskAssessmentCols+` FROM risk_assessment WHERE fhir_id = $1`, fhirID))
}

func (r *riskAssessmentRepoPG) Update(ctx context.Context, ra *RiskAssessment) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE risk_assessment SET status=$2, prediction_outcome=$3, prediction_probability=$4,
			prediction_qualitative=$5, mitigation=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		ra.ID, ra.Status, ra.PredictionOutcome, ra.PredictionProbability,
		ra.PredictionQualitative, ra.Mitigation, ra.Note)
	return err
}

func (r *riskAssessmentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM risk_assessment WHERE id = $1`, id)
	return err
}

func (r *riskAssessmentRepoPG) List(ctx context.Context, limit, offset int) ([]*RiskAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM risk_assessment`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+riskAssessmentCols+` FROM risk_assessment ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RiskAssessment
	for rows.Next() {
		ra, err := r.scanRiskAssessment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ra)
	}
	return items, total, nil
}

func (r *riskAssessmentRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RiskAssessment, int, error) {
	query := `SELECT ` + riskAssessmentCols + ` FROM risk_assessment WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM risk_assessment WHERE 1=1`
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
	var items []*RiskAssessment
	for rows.Next() {
		ra, err := r.scanRiskAssessment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ra)
	}
	return items, total, nil
}
