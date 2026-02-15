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

// AdverseEvent maps to the adverse_event table (FHIR AdverseEvent resource).
type AdverseEvent struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	FHIRID           string     `db:"fhir_id" json:"fhir_id"`
	Actuality        string     `db:"actuality" json:"actuality"`
	CategoryCode     *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay  *string    `db:"category_display" json:"category_display,omitempty"`
	EventCode        *string    `db:"event_code" json:"event_code,omitempty"`
	EventDisplay     *string    `db:"event_display" json:"event_display,omitempty"`
	EventSystem      *string    `db:"event_system" json:"event_system,omitempty"`
	SubjectPatientID uuid.UUID  `db:"subject_patient_id" json:"subject_patient_id"`
	EncounterID      *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	Date             *time.Time `db:"date" json:"date,omitempty"`
	Detected         *time.Time `db:"detected" json:"detected,omitempty"`
	RecordedDate     *time.Time `db:"recorded_date" json:"recorded_date,omitempty"`
	RecorderID       *uuid.UUID `db:"recorder_id" json:"recorder_id,omitempty"`
	SeriousnessCode  *string    `db:"seriousness_code" json:"seriousness_code,omitempty"`
	SeverityCode     *string    `db:"severity_code" json:"severity_code,omitempty"`
	OutcomeCode      *string    `db:"outcome_code" json:"outcome_code,omitempty"`
	LocationID       *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	Description      *string    `db:"description" json:"description,omitempty"`
	VersionID        int        `db:"version_id" json:"version_id"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

func (a *AdverseEvent) GetVersionID() int  { return a.VersionID }
func (a *AdverseEvent) SetVersionID(v int) { a.VersionID = v }

func (a *AdverseEvent) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "AdverseEvent",
		"id":           a.FHIRID,
		"actuality":    a.Actuality,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", a.SubjectPatientID.String())},
		"meta":         fhir.Meta{LastUpdated: a.UpdatedAt},
	}
	if a.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *a.CategoryCode, Display: strVal(a.CategoryDisplay)}},
		}}
	}
	if a.EventCode != nil {
		result["event"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(a.EventSystem),
				Code:    *a.EventCode,
				Display: strVal(a.EventDisplay),
			}},
		}
	}
	if a.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", a.EncounterID.String())}
	}
	if a.Date != nil {
		result["date"] = a.Date.Format(time.RFC3339)
	}
	if a.Detected != nil {
		result["detected"] = a.Detected.Format(time.RFC3339)
	}
	if a.RecordedDate != nil {
		result["recordedDate"] = a.RecordedDate.Format(time.RFC3339)
	}
	if a.RecorderID != nil {
		result["recorder"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", a.RecorderID.String())}
	}
	if a.SeriousnessCode != nil {
		result["seriousness"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *a.SeriousnessCode}},
		}
	}
	if a.SeverityCode != nil {
		result["severity"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *a.SeverityCode}},
		}
	}
	if a.OutcomeCode != nil {
		result["outcome"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *a.OutcomeCode}},
		}
	}
	if a.LocationID != nil {
		result["location"] = fhir.Reference{Reference: fhir.FormatReference("Location", a.LocationID.String())}
	}
	if a.Description != nil {
		result["description"] = *a.Description
	}
	return result
}

// AdverseEventRepository defines the repository interface.
type AdverseEventRepository interface {
	Create(ctx context.Context, a *AdverseEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*AdverseEvent, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*AdverseEvent, error)
	Update(ctx context.Context, a *AdverseEvent) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*AdverseEvent, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AdverseEvent, int, error)
}

// -- Postgres implementation --
type adverseEventRepoPG struct{ pool *pgxpool.Pool }

func NewAdverseEventRepoPG(pool *pgxpool.Pool) AdverseEventRepository {
	return &adverseEventRepoPG{pool: pool}
}

func (r *adverseEventRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const adverseEventCols = `id, fhir_id, actuality, category_code, category_display,
	event_code, event_display, event_system, subject_patient_id, encounter_id,
	date, detected, recorded_date, recorder_id,
	seriousness_code, severity_code, outcome_code, location_id, description,
	version_id, created_at, updated_at`

func (r *adverseEventRepoPG) scanAdverseEvent(row pgx.Row) (*AdverseEvent, error) {
	var a AdverseEvent
	err := row.Scan(&a.ID, &a.FHIRID, &a.Actuality, &a.CategoryCode, &a.CategoryDisplay,
		&a.EventCode, &a.EventDisplay, &a.EventSystem, &a.SubjectPatientID, &a.EncounterID,
		&a.Date, &a.Detected, &a.RecordedDate, &a.RecorderID,
		&a.SeriousnessCode, &a.SeverityCode, &a.OutcomeCode, &a.LocationID, &a.Description,
		&a.VersionID, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *adverseEventRepoPG) Create(ctx context.Context, a *AdverseEvent) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO adverse_event (id, fhir_id, actuality, category_code, category_display,
			event_code, event_display, event_system, subject_patient_id, encounter_id,
			date, detected, recorded_date, recorder_id,
			seriousness_code, severity_code, outcome_code, location_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		a.ID, a.FHIRID, a.Actuality, a.CategoryCode, a.CategoryDisplay,
		a.EventCode, a.EventDisplay, a.EventSystem, a.SubjectPatientID, a.EncounterID,
		a.Date, a.Detected, a.RecordedDate, a.RecorderID,
		a.SeriousnessCode, a.SeverityCode, a.OutcomeCode, a.LocationID, a.Description)
	return err
}

func (r *adverseEventRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*AdverseEvent, error) {
	return r.scanAdverseEvent(r.conn(ctx).QueryRow(ctx, `SELECT `+adverseEventCols+` FROM adverse_event WHERE id = $1`, id))
}

func (r *adverseEventRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*AdverseEvent, error) {
	return r.scanAdverseEvent(r.conn(ctx).QueryRow(ctx, `SELECT `+adverseEventCols+` FROM adverse_event WHERE fhir_id = $1`, fhirID))
}

func (r *adverseEventRepoPG) Update(ctx context.Context, a *AdverseEvent) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE adverse_event SET actuality=$2, seriousness_code=$3, severity_code=$4,
			outcome_code=$5, description=$6, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.Actuality, a.SeriousnessCode, a.SeverityCode,
		a.OutcomeCode, a.Description)
	return err
}

func (r *adverseEventRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM adverse_event WHERE id = $1`, id)
	return err
}

func (r *adverseEventRepoPG) List(ctx context.Context, limit, offset int) ([]*AdverseEvent, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM adverse_event`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+adverseEventCols+` FROM adverse_event ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*AdverseEvent
	for rows.Next() {
		a, err := r.scanAdverseEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *adverseEventRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AdverseEvent, int, error) {
	query := `SELECT ` + adverseEventCols + ` FROM adverse_event WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM adverse_event WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND subject_patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND subject_patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["actuality"]; ok {
		query += fmt.Sprintf(` AND actuality = $%d`, idx)
		countQuery += fmt.Sprintf(` AND actuality = $%d`, idx)
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
	var items []*AdverseEvent
	for rows.Next() {
		a, err := r.scanAdverseEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}
