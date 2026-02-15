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

// DetectedIssue maps to the detected_issue table (FHIR DetectedIssue resource).
type DetectedIssue struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	FHIRID               string     `db:"fhir_id" json:"fhir_id"`
	Status               string     `db:"status" json:"status"`
	CodeCode             *string    `db:"code_code" json:"code_code,omitempty"`
	CodeDisplay          *string    `db:"code_display" json:"code_display,omitempty"`
	CodeSystem           *string    `db:"code_system" json:"code_system,omitempty"`
	Severity             *string    `db:"severity" json:"severity,omitempty"`
	PatientID            *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	IdentifiedDate       *time.Time `db:"identified_date" json:"identified_date,omitempty"`
	AuthorPractitionerID *uuid.UUID `db:"author_practitioner_id" json:"author_practitioner_id,omitempty"`
	Detail               *string    `db:"detail" json:"detail,omitempty"`
	ReferenceURL         *string    `db:"reference_url" json:"reference_url,omitempty"`
	MitigationAction     *string    `db:"mitigation_action" json:"mitigation_action,omitempty"`
	MitigationDate       *time.Time `db:"mitigation_date" json:"mitigation_date,omitempty"`
	MitigationAuthorID   *uuid.UUID `db:"mitigation_author_id" json:"mitigation_author_id,omitempty"`
	VersionID            int        `db:"version_id" json:"version_id"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

func (d *DetectedIssue) GetVersionID() int  { return d.VersionID }
func (d *DetectedIssue) SetVersionID(v int) { d.VersionID = v }

func (d *DetectedIssue) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "DetectedIssue",
		"id":           d.FHIRID,
		"status":       d.Status,
		"meta":         fhir.Meta{LastUpdated: d.UpdatedAt},
	}
	if d.CodeCode != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				System:  strVal(d.CodeSystem),
				Code:    *d.CodeCode,
				Display: strVal(d.CodeDisplay),
			}},
		}
	}
	if d.Severity != nil {
		result["severity"] = *d.Severity
	}
	if d.PatientID != nil {
		result["patient"] = fhir.Reference{Reference: fhir.FormatReference("Patient", d.PatientID.String())}
	}
	if d.IdentifiedDate != nil {
		result["identifiedDateTime"] = d.IdentifiedDate.Format(time.RFC3339)
	}
	if d.AuthorPractitionerID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", d.AuthorPractitionerID.String())}
	}
	if d.Detail != nil {
		result["detail"] = *d.Detail
	}
	if d.ReferenceURL != nil {
		result["reference"] = *d.ReferenceURL
	}
	if d.MitigationAction != nil {
		mitigation := map[string]interface{}{
			"action": fhir.CodeableConcept{
				Coding: []fhir.Coding{{Display: *d.MitigationAction}},
			},
		}
		if d.MitigationDate != nil {
			mitigation["date"] = d.MitigationDate.Format(time.RFC3339)
		}
		if d.MitigationAuthorID != nil {
			mitigation["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", d.MitigationAuthorID.String())}
		}
		result["mitigation"] = []interface{}{mitigation}
	}
	return result
}

// DetectedIssueRepository defines the repository interface.
type DetectedIssueRepository interface {
	Create(ctx context.Context, d *DetectedIssue) error
	GetByID(ctx context.Context, id uuid.UUID) (*DetectedIssue, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DetectedIssue, error)
	Update(ctx context.Context, d *DetectedIssue) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DetectedIssue, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DetectedIssue, int, error)
}

// -- Postgres implementation --
type detectedIssueRepoPG struct{ pool *pgxpool.Pool }

func NewDetectedIssueRepoPG(pool *pgxpool.Pool) DetectedIssueRepository {
	return &detectedIssueRepoPG{pool: pool}
}

func (r *detectedIssueRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const detectedIssueCols = `id, fhir_id, status, code_code, code_display, code_system,
	severity, patient_id, identified_date, author_practitioner_id,
	detail, reference_url, mitigation_action, mitigation_date, mitigation_author_id,
	version_id, created_at, updated_at`

func (r *detectedIssueRepoPG) scanDetectedIssue(row pgx.Row) (*DetectedIssue, error) {
	var d DetectedIssue
	err := row.Scan(&d.ID, &d.FHIRID, &d.Status, &d.CodeCode, &d.CodeDisplay, &d.CodeSystem,
		&d.Severity, &d.PatientID, &d.IdentifiedDate, &d.AuthorPractitionerID,
		&d.Detail, &d.ReferenceURL, &d.MitigationAction, &d.MitigationDate, &d.MitigationAuthorID,
		&d.VersionID, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *detectedIssueRepoPG) Create(ctx context.Context, d *DetectedIssue) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO detected_issue (id, fhir_id, status, code_code, code_display, code_system,
			severity, patient_id, identified_date, author_practitioner_id,
			detail, reference_url, mitigation_action, mitigation_date, mitigation_author_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		d.ID, d.FHIRID, d.Status, d.CodeCode, d.CodeDisplay, d.CodeSystem,
		d.Severity, d.PatientID, d.IdentifiedDate, d.AuthorPractitionerID,
		d.Detail, d.ReferenceURL, d.MitigationAction, d.MitigationDate, d.MitigationAuthorID)
	return err
}

func (r *detectedIssueRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DetectedIssue, error) {
	return r.scanDetectedIssue(r.conn(ctx).QueryRow(ctx, `SELECT `+detectedIssueCols+` FROM detected_issue WHERE id = $1`, id))
}

func (r *detectedIssueRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DetectedIssue, error) {
	return r.scanDetectedIssue(r.conn(ctx).QueryRow(ctx, `SELECT `+detectedIssueCols+` FROM detected_issue WHERE fhir_id = $1`, fhirID))
}

func (r *detectedIssueRepoPG) Update(ctx context.Context, d *DetectedIssue) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE detected_issue SET status=$2, severity=$3, detail=$4,
			mitigation_action=$5, mitigation_date=$6, mitigation_author_id=$7, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Status, d.Severity, d.Detail,
		d.MitigationAction, d.MitigationDate, d.MitigationAuthorID)
	return err
}

func (r *detectedIssueRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM detected_issue WHERE id = $1`, id)
	return err
}

func (r *detectedIssueRepoPG) List(ctx context.Context, limit, offset int) ([]*DetectedIssue, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM detected_issue`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+detectedIssueCols+` FROM detected_issue ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DetectedIssue
	for rows.Next() {
		d, err := r.scanDetectedIssue(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

func (r *detectedIssueRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DetectedIssue, int, error) {
	query := `SELECT ` + detectedIssueCols + ` FROM detected_issue WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM detected_issue WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
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
	var items []*DetectedIssue
	for rows.Next() {
		d, err := r.scanDetectedIssue(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
