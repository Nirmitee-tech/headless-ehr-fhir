package behavioral

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

// =========== Psychiatric Assessment Repository ===========

type psychAssessmentRepoPG struct{ pool *pgxpool.Pool }

func NewPsychAssessmentRepoPG(pool *pgxpool.Pool) PsychAssessmentRepository {
	return &psychAssessmentRepoPG{pool: pool}
}

func (r *psychAssessmentRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const psychAssessCols = `id, patient_id, encounter_id, assessor_id, assessment_date,
	chief_complaint, history_present_illness, psychiatric_history, substance_use_history,
	family_psych_history, mental_status_exam, appearance, behavior, speech, mood, affect,
	thought_process, thought_content, perceptions, cognition, insight, judgment,
	risk_assessment, suicide_risk_level, homicide_risk_level,
	diagnosis_code, diagnosis_display, diagnosis_system,
	formulation, treatment_plan, disposition, note, created_at, updated_at`

func (r *psychAssessmentRepoPG) scanAssessment(row pgx.Row) (*PsychiatricAssessment, error) {
	var a PsychiatricAssessment
	err := row.Scan(&a.ID, &a.PatientID, &a.EncounterID, &a.AssessorID, &a.AssessmentDate,
		&a.ChiefComplaint, &a.HistoryPresentIllness, &a.PsychiatricHistory, &a.SubstanceUseHistory,
		&a.FamilyPsychHistory, &a.MentalStatusExam, &a.Appearance, &a.Behavior, &a.Speech, &a.Mood, &a.Affect,
		&a.ThoughtProcess, &a.ThoughtContent, &a.Perceptions, &a.Cognition, &a.Insight, &a.Judgment,
		&a.RiskAssessment, &a.SuicideRiskLevel, &a.HomicideRiskLevel,
		&a.DiagnosisCode, &a.DiagnosisDisplay, &a.DiagnosisSystem,
		&a.Formulation, &a.TreatmentPlan, &a.Disposition, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *psychAssessmentRepoPG) Create(ctx context.Context, a *PsychiatricAssessment) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO psychiatric_assessment (id, patient_id, encounter_id, assessor_id, assessment_date,
			chief_complaint, history_present_illness, psychiatric_history, substance_use_history,
			family_psych_history, mental_status_exam, appearance, behavior, speech, mood, affect,
			thought_process, thought_content, perceptions, cognition, insight, judgment,
			risk_assessment, suicide_risk_level, homicide_risk_level,
			diagnosis_code, diagnosis_display, diagnosis_system,
			formulation, treatment_plan, disposition, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32)`,
		a.ID, a.PatientID, a.EncounterID, a.AssessorID, a.AssessmentDate,
		a.ChiefComplaint, a.HistoryPresentIllness, a.PsychiatricHistory, a.SubstanceUseHistory,
		a.FamilyPsychHistory, a.MentalStatusExam, a.Appearance, a.Behavior, a.Speech, a.Mood, a.Affect,
		a.ThoughtProcess, a.ThoughtContent, a.Perceptions, a.Cognition, a.Insight, a.Judgment,
		a.RiskAssessment, a.SuicideRiskLevel, a.HomicideRiskLevel,
		a.DiagnosisCode, a.DiagnosisDisplay, a.DiagnosisSystem,
		a.Formulation, a.TreatmentPlan, a.Disposition, a.Note)
	return err
}

func (r *psychAssessmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PsychiatricAssessment, error) {
	return r.scanAssessment(r.conn(ctx).QueryRow(ctx, `SELECT `+psychAssessCols+` FROM psychiatric_assessment WHERE id = $1`, id))
}

func (r *psychAssessmentRepoPG) Update(ctx context.Context, a *PsychiatricAssessment) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE psychiatric_assessment SET chief_complaint=$2, mental_status_exam=$3,
			risk_assessment=$4, suicide_risk_level=$5, homicide_risk_level=$6,
			diagnosis_code=$7, diagnosis_display=$8, formulation=$9, treatment_plan=$10,
			disposition=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.ChiefComplaint, a.MentalStatusExam,
		a.RiskAssessment, a.SuicideRiskLevel, a.HomicideRiskLevel,
		a.DiagnosisCode, a.DiagnosisDisplay, a.Formulation, a.TreatmentPlan,
		a.Disposition, a.Note)
	return err
}

func (r *psychAssessmentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM psychiatric_assessment WHERE id = $1`, id)
	return err
}

func (r *psychAssessmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PsychiatricAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM psychiatric_assessment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+psychAssessCols+` FROM psychiatric_assessment WHERE patient_id = $1 ORDER BY assessment_date DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PsychiatricAssessment
	for rows.Next() {
		a, err := r.scanAssessment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *psychAssessmentRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PsychiatricAssessment, int, error) {
	query := `SELECT ` + psychAssessCols + ` FROM psychiatric_assessment WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM psychiatric_assessment WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["encounter"]; ok {
		query += fmt.Sprintf(` AND encounter_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND encounter_id = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY assessment_date DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PsychiatricAssessment
	for rows.Next() {
		a, err := r.scanAssessment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== Safety Plan Repository ===========

type safetyPlanRepoPG struct{ pool *pgxpool.Pool }

func NewSafetyPlanRepoPG(pool *pgxpool.Pool) SafetyPlanRepository {
	return &safetyPlanRepoPG{pool: pool}
}

func (r *safetyPlanRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const safetyPlanCols = `id, patient_id, created_by_id, status, plan_date,
	warning_signs, coping_strategies, social_distractions, people_to_contact,
	professionals_to_contact, emergency_contacts, means_restriction, reasons_for_living,
	patient_signature, provider_signature, review_date, note, created_at, updated_at`

func (r *safetyPlanRepoPG) scanSafetyPlan(row pgx.Row) (*SafetyPlan, error) {
	var s SafetyPlan
	err := row.Scan(&s.ID, &s.PatientID, &s.CreatedByID, &s.Status, &s.PlanDate,
		&s.WarningSigns, &s.CopingStrategies, &s.SocialDistractions, &s.PeopleToContact,
		&s.ProfessionalsToContact, &s.EmergencyContacts, &s.MeansRestriction, &s.ReasonsForLiving,
		&s.PatientSignature, &s.ProviderSignature, &s.ReviewDate, &s.Note, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *safetyPlanRepoPG) Create(ctx context.Context, s *SafetyPlan) error {
	s.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO safety_plan (id, patient_id, created_by_id, status, plan_date,
			warning_signs, coping_strategies, social_distractions, people_to_contact,
			professionals_to_contact, emergency_contacts, means_restriction, reasons_for_living,
			patient_signature, provider_signature, review_date, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		s.ID, s.PatientID, s.CreatedByID, s.Status, s.PlanDate,
		s.WarningSigns, s.CopingStrategies, s.SocialDistractions, s.PeopleToContact,
		s.ProfessionalsToContact, s.EmergencyContacts, s.MeansRestriction, s.ReasonsForLiving,
		s.PatientSignature, s.ProviderSignature, s.ReviewDate, s.Note)
	return err
}

func (r *safetyPlanRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SafetyPlan, error) {
	return r.scanSafetyPlan(r.conn(ctx).QueryRow(ctx, `SELECT `+safetyPlanCols+` FROM safety_plan WHERE id = $1`, id))
}

func (r *safetyPlanRepoPG) Update(ctx context.Context, s *SafetyPlan) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE safety_plan SET status=$2, warning_signs=$3, coping_strategies=$4,
			social_distractions=$5, people_to_contact=$6, professionals_to_contact=$7,
			emergency_contacts=$8, means_restriction=$9, reasons_for_living=$10,
			review_date=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.WarningSigns, s.CopingStrategies,
		s.SocialDistractions, s.PeopleToContact, s.ProfessionalsToContact,
		s.EmergencyContacts, s.MeansRestriction, s.ReasonsForLiving,
		s.ReviewDate, s.Note)
	return err
}

func (r *safetyPlanRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM safety_plan WHERE id = $1`, id)
	return err
}

func (r *safetyPlanRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SafetyPlan, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM safety_plan WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+safetyPlanCols+` FROM safety_plan WHERE patient_id = $1 ORDER BY plan_date DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SafetyPlan
	for rows.Next() {
		s, err := r.scanSafetyPlan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *safetyPlanRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SafetyPlan, int, error) {
	query := `SELECT ` + safetyPlanCols + ` FROM safety_plan WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM safety_plan WHERE 1=1`
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

	query += fmt.Sprintf(` ORDER BY plan_date DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SafetyPlan
	for rows.Next() {
		s, err := r.scanSafetyPlan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

// =========== Legal Hold Repository ===========

type legalHoldRepoPG struct{ pool *pgxpool.Pool }

func NewLegalHoldRepoPG(pool *pgxpool.Pool) LegalHoldRepository {
	return &legalHoldRepoPG{pool: pool}
}

func (r *legalHoldRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const legalHoldCols = `id, patient_id, encounter_id, initiated_by_id, status, hold_type,
	authority_statute, start_datetime, end_datetime, duration_hours, reason, criteria_met,
	certifying_physician_id, certification_datetime, court_hearing_date, court_order_number,
	legal_counsel_notified, patient_rights_given, release_reason, release_authorized_by_id,
	note, created_at, updated_at`

func (r *legalHoldRepoPG) scanLegalHold(row pgx.Row) (*LegalHold, error) {
	var h LegalHold
	err := row.Scan(&h.ID, &h.PatientID, &h.EncounterID, &h.InitiatedByID, &h.Status, &h.HoldType,
		&h.AuthorityStatute, &h.StartDatetime, &h.EndDatetime, &h.DurationHours, &h.Reason, &h.CriteriaMet,
		&h.CertifyingPhysicianID, &h.CertificationDatetime, &h.CourtHearingDate, &h.CourtOrderNumber,
		&h.LegalCounselNotified, &h.PatientRightsGiven, &h.ReleaseReason, &h.ReleaseAuthorizedByID,
		&h.Note, &h.CreatedAt, &h.UpdatedAt)
	return &h, err
}

func (r *legalHoldRepoPG) Create(ctx context.Context, h *LegalHold) error {
	h.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO legal_hold (id, patient_id, encounter_id, initiated_by_id, status, hold_type,
			authority_statute, start_datetime, end_datetime, duration_hours, reason, criteria_met,
			certifying_physician_id, certification_datetime, court_hearing_date, court_order_number,
			legal_counsel_notified, patient_rights_given, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		h.ID, h.PatientID, h.EncounterID, h.InitiatedByID, h.Status, h.HoldType,
		h.AuthorityStatute, h.StartDatetime, h.EndDatetime, h.DurationHours, h.Reason, h.CriteriaMet,
		h.CertifyingPhysicianID, h.CertificationDatetime, h.CourtHearingDate, h.CourtOrderNumber,
		h.LegalCounselNotified, h.PatientRightsGiven, h.Note)
	return err
}

func (r *legalHoldRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*LegalHold, error) {
	return r.scanLegalHold(r.conn(ctx).QueryRow(ctx, `SELECT `+legalHoldCols+` FROM legal_hold WHERE id = $1`, id))
}

func (r *legalHoldRepoPG) Update(ctx context.Context, h *LegalHold) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE legal_hold SET status=$2, end_datetime=$3, release_reason=$4,
			release_authorized_by_id=$5, court_hearing_date=$6, court_order_number=$7,
			legal_counsel_notified=$8, patient_rights_given=$9, note=$10, updated_at=NOW()
		WHERE id = $1`,
		h.ID, h.Status, h.EndDatetime, h.ReleaseReason,
		h.ReleaseAuthorizedByID, h.CourtHearingDate, h.CourtOrderNumber,
		h.LegalCounselNotified, h.PatientRightsGiven, h.Note)
	return err
}

func (r *legalHoldRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM legal_hold WHERE id = $1`, id)
	return err
}

func (r *legalHoldRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LegalHold, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM legal_hold WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+legalHoldCols+` FROM legal_hold WHERE patient_id = $1 ORDER BY start_datetime DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*LegalHold
	for rows.Next() {
		h, err := r.scanLegalHold(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, h)
	}
	return items, total, nil
}

func (r *legalHoldRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*LegalHold, int, error) {
	query := `SELECT ` + legalHoldCols + ` FROM legal_hold WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM legal_hold WHERE 1=1`
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

	query += fmt.Sprintf(` ORDER BY start_datetime DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*LegalHold
	for rows.Next() {
		h, err := r.scanLegalHold(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, h)
	}
	return items, total, nil
}

// =========== Seclusion/Restraint Repository ===========

type seclusionRestraintRepoPG struct{ pool *pgxpool.Pool }

func NewSeclusionRestraintRepoPG(pool *pgxpool.Pool) SeclusionRestraintRepository {
	return &seclusionRestraintRepoPG{pool: pool}
}

func (r *seclusionRestraintRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const seclusionCols = `id, patient_id, encounter_id, ordered_by_id, event_type, restraint_type,
	start_datetime, end_datetime, reason, behavior_description, alternatives_attempted,
	monitoring_frequency_min, last_monitoring_check, patient_condition_during, injuries_noted,
	nutrition_offered, toileting_offered, discontinued_by_id, discontinuation_reason,
	debrief_completed, debrief_notes, note, created_at, updated_at`

func (r *seclusionRestraintRepoPG) scanEvent(row pgx.Row) (*SeclusionRestraintEvent, error) {
	var e SeclusionRestraintEvent
	err := row.Scan(&e.ID, &e.PatientID, &e.EncounterID, &e.OrderedByID, &e.EventType, &e.RestraintType,
		&e.StartDatetime, &e.EndDatetime, &e.Reason, &e.BehaviorDescription, &e.AlternativesAttempted,
		&e.MonitoringFrequencyMin, &e.LastMonitoringCheck, &e.PatientConditionDuring, &e.InjuriesNoted,
		&e.NutritionOffered, &e.ToiletingOffered, &e.DiscontinuedByID, &e.DiscontinuationReason,
		&e.DebriefCompleted, &e.DebriefNotes, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *seclusionRestraintRepoPG) Create(ctx context.Context, e *SeclusionRestraintEvent) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO seclusion_restraint_event (id, patient_id, encounter_id, ordered_by_id, event_type,
			restraint_type, start_datetime, end_datetime, reason, behavior_description,
			alternatives_attempted, monitoring_frequency_min, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.PatientID, e.EncounterID, e.OrderedByID, e.EventType,
		e.RestraintType, e.StartDatetime, e.EndDatetime, e.Reason, e.BehaviorDescription,
		e.AlternativesAttempted, e.MonitoringFrequencyMin, e.Note)
	return err
}

func (r *seclusionRestraintRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SeclusionRestraintEvent, error) {
	return r.scanEvent(r.conn(ctx).QueryRow(ctx, `SELECT `+seclusionCols+` FROM seclusion_restraint_event WHERE id = $1`, id))
}

func (r *seclusionRestraintRepoPG) Update(ctx context.Context, e *SeclusionRestraintEvent) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE seclusion_restraint_event SET end_datetime=$2, patient_condition_during=$3,
			injuries_noted=$4, nutrition_offered=$5, toileting_offered=$6,
			discontinued_by_id=$7, discontinuation_reason=$8,
			debrief_completed=$9, debrief_notes=$10, note=$11, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.EndDatetime, e.PatientConditionDuring,
		e.InjuriesNoted, e.NutritionOffered, e.ToiletingOffered,
		e.DiscontinuedByID, e.DiscontinuationReason,
		e.DebriefCompleted, e.DebriefNotes, e.Note)
	return err
}

func (r *seclusionRestraintRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM seclusion_restraint_event WHERE id = $1`, id)
	return err
}

func (r *seclusionRestraintRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SeclusionRestraintEvent, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM seclusion_restraint_event WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+seclusionCols+` FROM seclusion_restraint_event WHERE patient_id = $1 ORDER BY start_datetime DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SeclusionRestraintEvent
	for rows.Next() {
		e, err := r.scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *seclusionRestraintRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SeclusionRestraintEvent, int, error) {
	query := `SELECT ` + seclusionCols + ` FROM seclusion_restraint_event WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM seclusion_restraint_event WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["event_type"]; ok {
		query += fmt.Sprintf(` AND event_type = $%d`, idx)
		countQuery += fmt.Sprintf(` AND event_type = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY start_datetime DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SeclusionRestraintEvent
	for rows.Next() {
		e, err := r.scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

// =========== Group Therapy Repository ===========

type groupTherapyRepoPG struct{ pool *pgxpool.Pool }

func NewGroupTherapyRepoPG(pool *pgxpool.Pool) GroupTherapyRepository {
	return &groupTherapyRepoPG{pool: pool}
}

func (r *groupTherapyRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const groupSessionCols = `id, session_name, session_type, facilitator_id, co_facilitator_id,
	status, scheduled_datetime, actual_start, actual_end, location, max_participants,
	topic, session_goals, session_notes, materials_used, note, created_at, updated_at`

func (r *groupTherapyRepoPG) scanSession(row pgx.Row) (*GroupTherapySession, error) {
	var s GroupTherapySession
	err := row.Scan(&s.ID, &s.SessionName, &s.SessionType, &s.FacilitatorID, &s.CoFacilitatorID,
		&s.Status, &s.ScheduledDatetime, &s.ActualStart, &s.ActualEnd, &s.Location, &s.MaxParticipants,
		&s.Topic, &s.SessionGoals, &s.SessionNotes, &s.MaterialsUsed, &s.Note, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *groupTherapyRepoPG) Create(ctx context.Context, s *GroupTherapySession) error {
	s.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO group_therapy_session (id, session_name, session_type, facilitator_id, co_facilitator_id,
			status, scheduled_datetime, actual_start, actual_end, location, max_participants,
			topic, session_goals, materials_used, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.SessionName, s.SessionType, s.FacilitatorID, s.CoFacilitatorID,
		s.Status, s.ScheduledDatetime, s.ActualStart, s.ActualEnd, s.Location, s.MaxParticipants,
		s.Topic, s.SessionGoals, s.MaterialsUsed, s.Note)
	return err
}

func (r *groupTherapyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*GroupTherapySession, error) {
	return r.scanSession(r.conn(ctx).QueryRow(ctx, `SELECT `+groupSessionCols+` FROM group_therapy_session WHERE id = $1`, id))
}

func (r *groupTherapyRepoPG) Update(ctx context.Context, s *GroupTherapySession) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE group_therapy_session SET status=$2, actual_start=$3, actual_end=$4,
			session_notes=$5, note=$6, updated_at=NOW()
		WHERE id = $1`,
		s.ID, s.Status, s.ActualStart, s.ActualEnd,
		s.SessionNotes, s.Note)
	return err
}

func (r *groupTherapyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM group_therapy_session WHERE id = $1`, id)
	return err
}

func (r *groupTherapyRepoPG) List(ctx context.Context, limit, offset int) ([]*GroupTherapySession, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM group_therapy_session`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+groupSessionCols+` FROM group_therapy_session ORDER BY scheduled_datetime DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*GroupTherapySession
	for rows.Next() {
		s, err := r.scanSession(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *groupTherapyRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*GroupTherapySession, int, error) {
	query := `SELECT ` + groupSessionCols + ` FROM group_therapy_session WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM group_therapy_session WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["facilitator"]; ok {
		query += fmt.Sprintf(` AND facilitator_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND facilitator_id = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY scheduled_datetime DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*GroupTherapySession
	for rows.Next() {
		s, err := r.scanSession(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, nil
}

func (r *groupTherapyRepoPG) AddAttendance(ctx context.Context, a *GroupTherapyAttendance) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO group_therapy_attendance (id, session_id, patient_id, attendance_status,
			participation_level, behavior_notes, mood_before, mood_after, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.SessionID, a.PatientID, a.AttendanceStatus,
		a.ParticipationLevel, a.BehaviorNotes, a.MoodBefore, a.MoodAfter, a.Note)
	return err
}

func (r *groupTherapyRepoPG) GetAttendance(ctx context.Context, sessionID uuid.UUID) ([]*GroupTherapyAttendance, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, session_id, patient_id, attendance_status,
			participation_level, behavior_notes, mood_before, mood_after, note
		FROM group_therapy_attendance WHERE session_id = $1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GroupTherapyAttendance
	for rows.Next() {
		var a GroupTherapyAttendance
		if err := rows.Scan(&a.ID, &a.SessionID, &a.PatientID, &a.AttendanceStatus,
			&a.ParticipationLevel, &a.BehaviorNotes, &a.MoodBefore, &a.MoodAfter, &a.Note); err != nil {
			return nil, err
		}
		items = append(items, &a)
	}
	return items, nil
}
