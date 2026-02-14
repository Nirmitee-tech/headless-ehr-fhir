package research

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

// =========== Research Study Repository ===========

type studyRepoPG struct{ pool *pgxpool.Pool }

func NewStudyRepoPG(pool *pgxpool.Pool) ResearchStudyRepository {
	return &studyRepoPG{pool: pool}
}

func (r *studyRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const studyCols = `id, fhir_id, title, protocol_number, status, phase, category, focus,
	description, sponsor_name, sponsor_contact, principal_investigator_id,
	site_name, site_contact, irb_number, irb_approval_date, irb_expiration_date,
	start_date, end_date, enrollment_target, enrollment_actual,
	primary_endpoint, secondary_endpoints, inclusion_criteria, exclusion_criteria,
	note, created_at, updated_at`

func (r *studyRepoPG) scanStudy(row pgx.Row) (*ResearchStudy, error) {
	var s ResearchStudy
	err := row.Scan(&s.ID, &s.FHIRID, &s.Title, &s.ProtocolNumber, &s.Status, &s.Phase, &s.Category, &s.Focus,
		&s.Description, &s.SponsorName, &s.SponsorContact, &s.PrincipalInvestigatorID,
		&s.SiteName, &s.SiteContact, &s.IRBNumber, &s.IRBApprovalDate, &s.IRBExpirationDate,
		&s.StartDate, &s.EndDate, &s.EnrollmentTarget, &s.EnrollmentActual,
		&s.PrimaryEndpoint, &s.SecondaryEndpoints, &s.InclusionCriteria, &s.ExclusionCriteria,
		&s.Note, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *studyRepoPG) Create(ctx context.Context, s *ResearchStudy) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = s.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_study (id, fhir_id, title, protocol_number, status, phase, category, focus,
			description, sponsor_name, sponsor_contact, principal_investigator_id,
			site_name, site_contact, irb_number, irb_approval_date, irb_expiration_date,
			start_date, end_date, enrollment_target,
			primary_endpoint, secondary_endpoints, inclusion_criteria, exclusion_criteria, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		s.ID, s.FHIRID, s.Title, s.ProtocolNumber, s.Status, s.Phase, s.Category, s.Focus,
		s.Description, s.SponsorName, s.SponsorContact, s.PrincipalInvestigatorID,
		s.SiteName, s.SiteContact, s.IRBNumber, s.IRBApprovalDate, s.IRBExpirationDate,
		s.StartDate, s.EndDate, s.EnrollmentTarget,
		s.PrimaryEndpoint, s.SecondaryEndpoints, s.InclusionCriteria, s.ExclusionCriteria, s.Note)
	return err
}

func (r *studyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ResearchStudy, error) {
	return r.scanStudy(r.conn(ctx).QueryRow(ctx, `SELECT `+studyCols+` FROM research_study WHERE id = $1`, id))
}

func (r *studyRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ResearchStudy, error) {
	return r.scanStudy(r.conn(ctx).QueryRow(ctx, `SELECT `+studyCols+` FROM research_study WHERE fhir_id = $1`, fhirID))
}

func (r *studyRepoPG) Update(ctx context.Context, s *ResearchStudy) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE research_study SET status=$2, title=$3, description=$4,
			start_date=$5, end_date=$6, enrollment_target=$7,
			primary_endpoint=$8, secondary_endpoints=$9, note=$10, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.Title, s.Description,
		s.StartDate, s.EndDate, s.EnrollmentTarget,
		s.PrimaryEndpoint, s.SecondaryEndpoints, s.Note)
	return err
}

func (r *studyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM research_study WHERE id = $1`, id)
	return err
}

func (r *studyRepoPG) List(ctx context.Context, limit, offset int) ([]*ResearchStudy, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_study`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+studyCols+` FROM research_study ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchStudy
	for rows.Next() {
		s, err := r.scanStudy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *studyRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchStudy, int, error) {
	query := `SELECT ` + studyCols + ` FROM research_study WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM research_study WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["title"]; ok {
		query += fmt.Sprintf(` AND title ILIKE '%%' || $%d || '%%'`, idx)
		countQuery += fmt.Sprintf(` AND title ILIKE '%%' || $%d || '%%'`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["protocol"]; ok {
		query += fmt.Sprintf(` AND protocol_number = $%d`, idx)
		countQuery += fmt.Sprintf(` AND protocol_number = $%d`, idx)
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
	var items []*ResearchStudy
	for rows.Next() {
		s, err := r.scanStudy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *studyRepoPG) AddArm(ctx context.Context, a *ResearchArm) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_arm (id, study_id, name, arm_type, description, target_enrollment)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		a.ID, a.StudyID, a.Name, a.ArmType, a.Description, a.TargetEnrollment)
	return err
}

func (r *studyRepoPG) GetArms(ctx context.Context, studyID uuid.UUID) ([]*ResearchArm, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, study_id, name, arm_type, description, target_enrollment, actual_enrollment
		FROM research_arm WHERE study_id = $1`, studyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ResearchArm
	for rows.Next() {
		var a ResearchArm
		if err := rows.Scan(&a.ID, &a.StudyID, &a.Name, &a.ArmType, &a.Description, &a.TargetEnrollment, &a.ActualEnrollment); err != nil {
			return nil, err
		}
		items = append(items, &a)
	}
	return items, nil
}

// =========== Enrollment Repository ===========

type enrollmentRepoPG struct{ pool *pgxpool.Pool }

func NewEnrollmentRepoPG(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentRepoPG{pool: pool}
}

func (r *enrollmentRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const enrollCols = `id, study_id, arm_id, patient_id, consent_id, status,
	enrolled_date, screening_date, randomization_date, completion_date,
	withdrawal_date, withdrawal_reason, randomization_number, subject_number,
	enrolled_by_id, note, created_at, updated_at`

func (r *enrollmentRepoPG) scanEnrollment(row pgx.Row) (*ResearchEnrollment, error) {
	var e ResearchEnrollment
	err := row.Scan(&e.ID, &e.StudyID, &e.ArmID, &e.PatientID, &e.ConsentID, &e.Status,
		&e.EnrolledDate, &e.ScreeningDate, &e.RandomizationDate, &e.CompletionDate,
		&e.WithdrawalDate, &e.WithdrawalReason, &e.RandomizationNumber, &e.SubjectNumber,
		&e.EnrolledByID, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *enrollmentRepoPG) Create(ctx context.Context, e *ResearchEnrollment) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_enrollment (id, study_id, arm_id, patient_id, consent_id, status,
			enrolled_date, screening_date, randomization_date,
			randomization_number, subject_number, enrolled_by_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.StudyID, e.ArmID, e.PatientID, e.ConsentID, e.Status,
		e.EnrolledDate, e.ScreeningDate, e.RandomizationDate,
		e.RandomizationNumber, e.SubjectNumber, e.EnrolledByID, e.Note)
	return err
}

func (r *enrollmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ResearchEnrollment, error) {
	return r.scanEnrollment(r.conn(ctx).QueryRow(ctx, `SELECT `+enrollCols+` FROM research_enrollment WHERE id = $1`, id))
}

func (r *enrollmentRepoPG) Update(ctx context.Context, e *ResearchEnrollment) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE research_enrollment SET status=$2, arm_id=$3,
			enrolled_date=$4, completion_date=$5, withdrawal_date=$6,
			withdrawal_reason=$7, subject_number=$8, note=$9, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.ArmID,
		e.EnrolledDate, e.CompletionDate, e.WithdrawalDate,
		e.WithdrawalReason, e.SubjectNumber, e.Note)
	return err
}

func (r *enrollmentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM research_enrollment WHERE id = $1`, id)
	return err
}

func (r *enrollmentRepoPG) ListByStudy(ctx context.Context, studyID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_enrollment WHERE study_id = $1`, studyID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+enrollCols+` FROM research_enrollment WHERE study_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, studyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchEnrollment
	for rows.Next() {
		e, err := r.scanEnrollment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *enrollmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_enrollment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+enrollCols+` FROM research_enrollment WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchEnrollment
	for rows.Next() {
		e, err := r.scanEnrollment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

// =========== Adverse Event Repository ===========

type adverseEventRepoPG struct{ pool *pgxpool.Pool }

func NewAdverseEventRepoPG(pool *pgxpool.Pool) AdverseEventRepository {
	return &adverseEventRepoPG{pool: pool}
}

func (r *adverseEventRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const aeCols = `id, enrollment_id, event_date, reported_date, reported_by_id,
	description, severity, seriousness, causality, expectedness, outcome, action_taken,
	resolution_date, reported_to_irb, irb_report_date,
	reported_to_sponsor, sponsor_report_date, note, created_at, updated_at`

func (r *adverseEventRepoPG) scanAE(row pgx.Row) (*ResearchAdverseEvent, error) {
	var ae ResearchAdverseEvent
	err := row.Scan(&ae.ID, &ae.EnrollmentID, &ae.EventDate, &ae.ReportedDate, &ae.ReportedByID,
		&ae.Description, &ae.Severity, &ae.Seriousness, &ae.Causality, &ae.Expectedness, &ae.Outcome, &ae.ActionTaken,
		&ae.ResolutionDate, &ae.ReportedToIRB, &ae.IRBReportDate,
		&ae.ReportedToSponsor, &ae.SponsorReportDate, &ae.Note, &ae.CreatedAt, &ae.UpdatedAt)
	return &ae, err
}

func (r *adverseEventRepoPG) Create(ctx context.Context, ae *ResearchAdverseEvent) error {
	ae.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_adverse_event (id, enrollment_id, event_date, reported_date, reported_by_id,
			description, severity, seriousness, causality, expectedness, outcome, action_taken, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		ae.ID, ae.EnrollmentID, ae.EventDate, ae.ReportedDate, ae.ReportedByID,
		ae.Description, ae.Severity, ae.Seriousness, ae.Causality, ae.Expectedness, ae.Outcome, ae.ActionTaken, ae.Note)
	return err
}

func (r *adverseEventRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ResearchAdverseEvent, error) {
	return r.scanAE(r.conn(ctx).QueryRow(ctx, `SELECT `+aeCols+` FROM research_adverse_event WHERE id = $1`, id))
}

func (r *adverseEventRepoPG) Update(ctx context.Context, ae *ResearchAdverseEvent) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE research_adverse_event SET severity=$2, seriousness=$3, causality=$4,
			outcome=$5, action_taken=$6, resolution_date=$7,
			reported_to_irb=$8, irb_report_date=$9,
			reported_to_sponsor=$10, sponsor_report_date=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		ae.ID, ae.Severity, ae.Seriousness, ae.Causality,
		ae.Outcome, ae.ActionTaken, ae.ResolutionDate,
		ae.ReportedToIRB, ae.IRBReportDate,
		ae.ReportedToSponsor, ae.SponsorReportDate, ae.Note)
	return err
}

func (r *adverseEventRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM research_adverse_event WHERE id = $1`, id)
	return err
}

func (r *adverseEventRepoPG) ListByEnrollment(ctx context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchAdverseEvent, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_adverse_event WHERE enrollment_id = $1`, enrollmentID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+aeCols+` FROM research_adverse_event WHERE enrollment_id = $1 ORDER BY event_date DESC LIMIT $2 OFFSET $3`, enrollmentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchAdverseEvent
	for rows.Next() {
		ae, err := r.scanAE(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ae)
	}
	return items, total, nil
}

// =========== Protocol Deviation Repository ===========

type deviationRepoPG struct{ pool *pgxpool.Pool }

func NewDeviationRepoPG(pool *pgxpool.Pool) DeviationRepository {
	return &deviationRepoPG{pool: pool}
}

func (r *deviationRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const devCols = `id, enrollment_id, deviation_date, reported_date, reported_by_id,
	category, description, severity, corrective_action, preventive_action,
	impact_on_subject, impact_on_study,
	reported_to_irb, irb_report_date, reported_to_sponsor, sponsor_report_date,
	note, created_at, updated_at`

func (r *deviationRepoPG) scanDev(row pgx.Row) (*ResearchProtocolDeviation, error) {
	var d ResearchProtocolDeviation
	err := row.Scan(&d.ID, &d.EnrollmentID, &d.DeviationDate, &d.ReportedDate, &d.ReportedByID,
		&d.Category, &d.Description, &d.Severity, &d.CorrectiveAction, &d.PreventiveAction,
		&d.ImpactOnSubject, &d.ImpactOnStudy,
		&d.ReportedToIRB, &d.IRBReportDate, &d.ReportedToSponsor, &d.SponsorReportDate,
		&d.Note, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *deviationRepoPG) Create(ctx context.Context, d *ResearchProtocolDeviation) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO research_protocol_deviation (id, enrollment_id, deviation_date, reported_date,
			reported_by_id, category, description, severity,
			corrective_action, preventive_action, impact_on_subject, impact_on_study, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		d.ID, d.EnrollmentID, d.DeviationDate, d.ReportedDate,
		d.ReportedByID, d.Category, d.Description, d.Severity,
		d.CorrectiveAction, d.PreventiveAction, d.ImpactOnSubject, d.ImpactOnStudy, d.Note)
	return err
}

func (r *deviationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ResearchProtocolDeviation, error) {
	return r.scanDev(r.conn(ctx).QueryRow(ctx, `SELECT `+devCols+` FROM research_protocol_deviation WHERE id = $1`, id))
}

func (r *deviationRepoPG) Update(ctx context.Context, d *ResearchProtocolDeviation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE research_protocol_deviation SET severity=$2, corrective_action=$3,
			preventive_action=$4, impact_on_subject=$5, impact_on_study=$6,
			reported_to_irb=$7, irb_report_date=$8,
			reported_to_sponsor=$9, sponsor_report_date=$10, note=$11, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Severity, d.CorrectiveAction,
		d.PreventiveAction, d.ImpactOnSubject, d.ImpactOnStudy,
		d.ReportedToIRB, d.IRBReportDate,
		d.ReportedToSponsor, d.SponsorReportDate, d.Note)
	return err
}

func (r *deviationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM research_protocol_deviation WHERE id = $1`, id)
	return err
}

func (r *deviationRepoPG) ListByEnrollment(ctx context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchProtocolDeviation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM research_protocol_deviation WHERE enrollment_id = $1`, enrollmentID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+devCols+` FROM research_protocol_deviation WHERE enrollment_id = $1 ORDER BY deviation_date DESC LIMIT $2 OFFSET $3`, enrollmentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ResearchProtocolDeviation
	for rows.Next() {
		d, err := r.scanDev(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
