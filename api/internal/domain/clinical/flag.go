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

// Flag maps to the flag table (FHIR Flag resource).
type Flag struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	FHIRID               string     `db:"fhir_id" json:"fhir_id"`
	Status               string     `db:"status" json:"status"`
	CategoryCode         *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay      *string    `db:"category_display" json:"category_display,omitempty"`
	CodeCode             string     `db:"code_code" json:"code_code"`
	CodeDisplay          *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem           *string    `db:"code_system" json:"code_system,omitempty"`
	SubjectPatientID     *uuid.UUID `db:"subject_patient_id" json:"subject_patient_id,omitempty"`
	PeriodStart          *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd            *time.Time `db:"period_end" json:"period_end,omitempty"`
	EncounterID          *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AuthorPractitionerID *uuid.UUID `db:"author_practitioner_id" json:"author_practitioner_id,omitempty"`
	VersionID            int        `db:"version_id" json:"version_id"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

func (f *Flag) GetVersionID() int  { return f.VersionID }
func (f *Flag) SetVersionID(v int) { f.VersionID = v }

func (f *Flag) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Flag",
		"id":           f.FHIRID,
		"status":       f.Status,
		"code": fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(f.CodeSystem),
				Code:    f.CodeCode,
				Display: strVal(f.CodeDisplay),
			}},
		},
		"meta": fhir.Meta{LastUpdated: f.UpdatedAt},
	}
	if f.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *f.CategoryCode, Display: strVal(f.CategoryDisplay)}},
		}}
	}
	if f.SubjectPatientID != nil {
		result["subject"] = fhir.Reference{Reference: fhir.FormatReference("Patient", f.SubjectPatientID.String())}
	}
	if f.PeriodStart != nil || f.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: f.PeriodStart, End: f.PeriodEnd}
	}
	if f.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", f.EncounterID.String())}
	}
	if f.AuthorPractitionerID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", f.AuthorPractitionerID.String())}
	}
	return result
}

// FlagRepository defines the repository interface.
type FlagRepository interface {
	Create(ctx context.Context, f *Flag) error
	GetByID(ctx context.Context, id uuid.UUID) (*Flag, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Flag, error)
	Update(ctx context.Context, f *Flag) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Flag, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Flag, int, error)
}

// -- Postgres implementation --
type flagRepoPG struct{ pool *pgxpool.Pool }

func NewFlagRepoPG(pool *pgxpool.Pool) FlagRepository { return &flagRepoPG{pool: pool} }

func (r *flagRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const flagCols = `id, fhir_id, status, category_code, category_display,
	code_code, code_display, code_system, subject_patient_id,
	period_start, period_end, encounter_id, author_practitioner_id,
	version_id, created_at, updated_at`

func (r *flagRepoPG) scanFlag(row pgx.Row) (*Flag, error) {
	var f Flag
	err := row.Scan(&f.ID, &f.FHIRID, &f.Status, &f.CategoryCode, &f.CategoryDisplay,
		&f.CodeCode, &f.CodeDisplay, &f.CodeSystem, &f.SubjectPatientID,
		&f.PeriodStart, &f.PeriodEnd, &f.EncounterID, &f.AuthorPractitionerID,
		&f.VersionID, &f.CreatedAt, &f.UpdatedAt)
	return &f, err
}

func (r *flagRepoPG) Create(ctx context.Context, f *Flag) error {
	f.ID = uuid.New()
	if f.FHIRID == "" {
		f.FHIRID = f.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO flag (id, fhir_id, status, category_code, category_display,
			code_code, code_display, code_system, subject_patient_id,
			period_start, period_end, encounter_id, author_practitioner_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		f.ID, f.FHIRID, f.Status, f.CategoryCode, f.CategoryDisplay,
		f.CodeCode, f.CodeDisplay, f.CodeSystem, f.SubjectPatientID,
		f.PeriodStart, f.PeriodEnd, f.EncounterID, f.AuthorPractitionerID)
	return err
}

func (r *flagRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Flag, error) {
	return r.scanFlag(r.conn(ctx).QueryRow(ctx, `SELECT `+flagCols+` FROM flag WHERE id = $1`, id))
}

func (r *flagRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Flag, error) {
	return r.scanFlag(r.conn(ctx).QueryRow(ctx, `SELECT `+flagCols+` FROM flag WHERE fhir_id = $1`, fhirID))
}

func (r *flagRepoPG) Update(ctx context.Context, f *Flag) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE flag SET status=$2, code_code=$3, code_display=$4, period_start=$5, period_end=$6, updated_at=NOW()
		WHERE id = $1`,
		f.ID, f.Status, f.CodeCode, f.CodeDisplay, f.PeriodStart, f.PeriodEnd)
	return err
}

func (r *flagRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM flag WHERE id = $1`, id)
	return err
}

func (r *flagRepoPG) List(ctx context.Context, limit, offset int) ([]*Flag, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM flag`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+flagCols+` FROM flag ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Flag
	for rows.Next() {
		f, err := r.scanFlag(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, nil
}

func (r *flagRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Flag, int, error) {
	query := `SELECT ` + flagCols + ` FROM flag WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM flag WHERE 1=1`
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
	var items []*Flag
	for rows.Next() {
		f, err := r.scanFlag(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, nil
}
