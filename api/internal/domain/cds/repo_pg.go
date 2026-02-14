package cds

import (
	"context"

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

// =========== CDS Rule Repository ===========

type cdsRuleRepoPG struct{ pool *pgxpool.Pool }

func NewCDSRuleRepoPG(pool *pgxpool.Pool) CDSRuleRepository { return &cdsRuleRepoPG{pool: pool} }

func (r *cdsRuleRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ruleCols = `id, rule_name, rule_type, description, severity, category,
	trigger_event, condition_expr, action_type, action_detail,
	evidence_source, evidence_url, active, version, created_at, updated_at`

func (r *cdsRuleRepoPG) scanRule(row pgx.Row) (*CDSRule, error) {
	var rule CDSRule
	err := row.Scan(&rule.ID, &rule.RuleName, &rule.RuleType, &rule.Description, &rule.Severity, &rule.Category,
		&rule.TriggerEvent, &rule.ConditionExpr, &rule.ActionType, &rule.ActionDetail,
		&rule.EvidenceSource, &rule.EvidenceURL, &rule.Active, &rule.Version, &rule.CreatedAt, &rule.UpdatedAt)
	return &rule, err
}

func (r *cdsRuleRepoPG) Create(ctx context.Context, rule *CDSRule) error {
	rule.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO cds_rule (id, rule_name, rule_type, description, severity, category,
			trigger_event, condition_expr, action_type, action_detail,
			evidence_source, evidence_url, active, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		rule.ID, rule.RuleName, rule.RuleType, rule.Description, rule.Severity, rule.Category,
		rule.TriggerEvent, rule.ConditionExpr, rule.ActionType, rule.ActionDetail,
		rule.EvidenceSource, rule.EvidenceURL, rule.Active, rule.Version)
	return err
}

func (r *cdsRuleRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CDSRule, error) {
	return r.scanRule(r.conn(ctx).QueryRow(ctx, `SELECT `+ruleCols+` FROM cds_rule WHERE id = $1`, id))
}

func (r *cdsRuleRepoPG) Update(ctx context.Context, rule *CDSRule) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE cds_rule SET rule_name=$2, rule_type=$3, description=$4, severity=$5, category=$6,
			trigger_event=$7, condition_expr=$8, action_type=$9, action_detail=$10,
			evidence_source=$11, evidence_url=$12, active=$13, version=$14, updated_at=NOW()
		WHERE id = $1`,
		rule.ID, rule.RuleName, rule.RuleType, rule.Description, rule.Severity, rule.Category,
		rule.TriggerEvent, rule.ConditionExpr, rule.ActionType, rule.ActionDetail,
		rule.EvidenceSource, rule.EvidenceURL, rule.Active, rule.Version)
	return err
}

func (r *cdsRuleRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM cds_rule WHERE id = $1`, id)
	return err
}

func (r *cdsRuleRepoPG) List(ctx context.Context, limit, offset int) ([]*CDSRule, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cds_rule`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ruleCols+` FROM cds_rule ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CDSRule
	for rows.Next() {
		rule, err := r.scanRule(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rule)
	}
	return items, total, nil
}

// =========== CDS Alert Repository ===========

type cdsAlertRepoPG struct{ pool *pgxpool.Pool }

func NewCDSAlertRepoPG(pool *pgxpool.Pool) CDSAlertRepository { return &cdsAlertRepoPG{pool: pool} }

func (r *cdsAlertRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const alertCols = `id, rule_id, patient_id, encounter_id, practitioner_id, status, severity,
	summary, detail, suggested_action, source, expires_at, fired_at, resolved_at, created_at, updated_at`

func (r *cdsAlertRepoPG) scanAlert(row pgx.Row) (*CDSAlert, error) {
	var a CDSAlert
	err := row.Scan(&a.ID, &a.RuleID, &a.PatientID, &a.EncounterID, &a.PractitionerID, &a.Status, &a.Severity,
		&a.Summary, &a.Detail, &a.SuggestedAction, &a.Source, &a.ExpiresAt, &a.FiredAt, &a.ResolvedAt, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *cdsAlertRepoPG) Create(ctx context.Context, a *CDSAlert) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO cds_alert (id, rule_id, patient_id, encounter_id, practitioner_id, status, severity,
			summary, detail, suggested_action, source, expires_at, fired_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		a.ID, a.RuleID, a.PatientID, a.EncounterID, a.PractitionerID, a.Status, a.Severity,
		a.Summary, a.Detail, a.SuggestedAction, a.Source, a.ExpiresAt, a.FiredAt)
	return err
}

func (r *cdsAlertRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CDSAlert, error) {
	return r.scanAlert(r.conn(ctx).QueryRow(ctx, `SELECT `+alertCols+` FROM cds_alert WHERE id = $1`, id))
}

func (r *cdsAlertRepoPG) Update(ctx context.Context, a *CDSAlert) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE cds_alert SET status=$2, resolved_at=$3, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.Status, a.ResolvedAt)
	return err
}

func (r *cdsAlertRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM cds_alert WHERE id = $1`, id)
	return err
}

func (r *cdsAlertRepoPG) List(ctx context.Context, limit, offset int) ([]*CDSAlert, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cds_alert`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+alertCols+` FROM cds_alert ORDER BY fired_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CDSAlert
	for rows.Next() {
		a, err := r.scanAlert(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *cdsAlertRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CDSAlert, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cds_alert WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+alertCols+` FROM cds_alert WHERE patient_id = $1 ORDER BY fired_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CDSAlert
	for rows.Next() {
		a, err := r.scanAlert(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *cdsAlertRepoPG) AddResponse(ctx context.Context, resp *CDSAlertResponse) error {
	resp.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO cds_alert_response (id, alert_id, practitioner_id, action, reason, comment)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		resp.ID, resp.AlertID, resp.PractitionerID, resp.Action, resp.Reason, resp.Comment)
	return err
}

func (r *cdsAlertRepoPG) GetResponses(ctx context.Context, alertID uuid.UUID) ([]*CDSAlertResponse, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, alert_id, practitioner_id, action, reason, comment, created_at
		FROM cds_alert_response WHERE alert_id = $1 ORDER BY created_at`, alertID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CDSAlertResponse
	for rows.Next() {
		var resp CDSAlertResponse
		if err := rows.Scan(&resp.ID, &resp.AlertID, &resp.PractitionerID, &resp.Action, &resp.Reason, &resp.Comment, &resp.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &resp)
	}
	return items, nil
}

// =========== Drug Interaction Repository ===========

type drugInteractionRepoPG struct{ pool *pgxpool.Pool }

func NewDrugInteractionRepoPG(pool *pgxpool.Pool) DrugInteractionRepository {
	return &drugInteractionRepoPG{pool: pool}
}

func (r *drugInteractionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const diCols = `id, medication_a_id, medication_a_name, medication_b_id, medication_b_name,
	severity, description, clinical_effect, management, evidence_level, source, active, created_at, updated_at`

func (r *drugInteractionRepoPG) scanDI(row pgx.Row) (*DrugInteraction, error) {
	var d DrugInteraction
	err := row.Scan(&d.ID, &d.MedicationAID, &d.MedicationAName, &d.MedicationBID, &d.MedicationBName,
		&d.Severity, &d.Description, &d.ClinicalEffect, &d.Management, &d.EvidenceLevel, &d.Source, &d.Active, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *drugInteractionRepoPG) Create(ctx context.Context, d *DrugInteraction) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO drug_interaction (id, medication_a_id, medication_a_name, medication_b_id, medication_b_name,
			severity, description, clinical_effect, management, evidence_level, source, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		d.ID, d.MedicationAID, d.MedicationAName, d.MedicationBID, d.MedicationBName,
		d.Severity, d.Description, d.ClinicalEffect, d.Management, d.EvidenceLevel, d.Source, d.Active)
	return err
}

func (r *drugInteractionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DrugInteraction, error) {
	return r.scanDI(r.conn(ctx).QueryRow(ctx, `SELECT `+diCols+` FROM drug_interaction WHERE id = $1`, id))
}

func (r *drugInteractionRepoPG) Update(ctx context.Context, d *DrugInteraction) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE drug_interaction SET medication_a_name=$2, medication_b_name=$3, severity=$4,
			description=$5, clinical_effect=$6, management=$7, evidence_level=$8, source=$9, active=$10, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.MedicationAName, d.MedicationBName, d.Severity,
		d.Description, d.ClinicalEffect, d.Management, d.EvidenceLevel, d.Source, d.Active)
	return err
}

func (r *drugInteractionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM drug_interaction WHERE id = $1`, id)
	return err
}

func (r *drugInteractionRepoPG) List(ctx context.Context, limit, offset int) ([]*DrugInteraction, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM drug_interaction`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+diCols+` FROM drug_interaction ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DrugInteraction
	for rows.Next() {
		d, err := r.scanDI(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

// =========== Order Set Repository ===========

type orderSetRepoPG struct{ pool *pgxpool.Pool }

func NewOrderSetRepoPG(pool *pgxpool.Pool) OrderSetRepository { return &orderSetRepoPG{pool: pool} }

func (r *orderSetRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const osCols = `id, name, description, category, status, author_id, version, approval_date, active, created_at, updated_at`

func (r *orderSetRepoPG) scanOS(row pgx.Row) (*OrderSet, error) {
	var o OrderSet
	err := row.Scan(&o.ID, &o.Name, &o.Description, &o.Category, &o.Status, &o.AuthorID, &o.Version, &o.ApprovalDate, &o.Active, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *orderSetRepoPG) Create(ctx context.Context, o *OrderSet) error {
	o.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO order_set (id, name, description, category, status, author_id, version, approval_date, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		o.ID, o.Name, o.Description, o.Category, o.Status, o.AuthorID, o.Version, o.ApprovalDate, o.Active)
	return err
}

func (r *orderSetRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*OrderSet, error) {
	return r.scanOS(r.conn(ctx).QueryRow(ctx, `SELECT `+osCols+` FROM order_set WHERE id = $1`, id))
}

func (r *orderSetRepoPG) Update(ctx context.Context, o *OrderSet) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE order_set SET name=$2, description=$3, category=$4, status=$5,
			version=$6, active=$7, updated_at=NOW()
		WHERE id = $1`,
		o.ID, o.Name, o.Description, o.Category, o.Status, o.Version, o.Active)
	return err
}

func (r *orderSetRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM order_set WHERE id = $1`, id)
	return err
}

func (r *orderSetRepoPG) List(ctx context.Context, limit, offset int) ([]*OrderSet, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM order_set`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+osCols+` FROM order_set ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*OrderSet
	for rows.Next() {
		o, err := r.scanOS(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

func (r *orderSetRepoPG) AddSection(ctx context.Context, s *OrderSetSection) error {
	s.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO order_set_section (id, order_set_id, name, description, sort_order)
		VALUES ($1,$2,$3,$4,$5)`,
		s.ID, s.OrderSetID, s.Name, s.Description, s.SortOrder)
	return err
}

func (r *orderSetRepoPG) GetSections(ctx context.Context, orderSetID uuid.UUID) ([]*OrderSetSection, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, order_set_id, name, description, sort_order
		FROM order_set_section WHERE order_set_id = $1 ORDER BY sort_order`, orderSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*OrderSetSection
	for rows.Next() {
		var s OrderSetSection
		if err := rows.Scan(&s.ID, &s.OrderSetID, &s.Name, &s.Description, &s.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, &s)
	}
	return items, nil
}

func (r *orderSetRepoPG) AddItem(ctx context.Context, item *OrderSetItem) error {
	item.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO order_set_item (id, section_id, item_type, item_name, item_code,
			default_dose, default_frequency, default_duration, instructions, is_required, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		item.ID, item.SectionID, item.ItemType, item.ItemName, item.ItemCode,
		item.DefaultDose, item.DefaultFrequency, item.DefaultDuration, item.Instructions, item.IsRequired, item.SortOrder)
	return err
}

func (r *orderSetRepoPG) GetItems(ctx context.Context, sectionID uuid.UUID) ([]*OrderSetItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, section_id, item_type, item_name, item_code,
			default_dose, default_frequency, default_duration, instructions, is_required, sort_order
		FROM order_set_item WHERE section_id = $1 ORDER BY sort_order`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*OrderSetItem
	for rows.Next() {
		var item OrderSetItem
		if err := rows.Scan(&item.ID, &item.SectionID, &item.ItemType, &item.ItemName, &item.ItemCode,
			&item.DefaultDose, &item.DefaultFrequency, &item.DefaultDuration, &item.Instructions, &item.IsRequired, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// =========== Clinical Pathway Repository ===========

type clinicalPathwayRepoPG struct{ pool *pgxpool.Pool }

func NewClinicalPathwayRepoPG(pool *pgxpool.Pool) ClinicalPathwayRepository {
	return &clinicalPathwayRepoPG{pool: pool}
}

func (r *clinicalPathwayRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const pathCols = `id, name, description, condition, category, version, author_id, active, expected_duration, created_at, updated_at`

func (r *clinicalPathwayRepoPG) scanPathway(row pgx.Row) (*ClinicalPathway, error) {
	var p ClinicalPathway
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Condition, &p.Category, &p.Version, &p.AuthorID, &p.Active, &p.ExpectedDuration, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *clinicalPathwayRepoPG) Create(ctx context.Context, p *ClinicalPathway) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO clinical_pathway (id, name, description, condition, category, version, author_id, active, expected_duration)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.Name, p.Description, p.Condition, p.Category, p.Version, p.AuthorID, p.Active, p.ExpectedDuration)
	return err
}

func (r *clinicalPathwayRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ClinicalPathway, error) {
	return r.scanPathway(r.conn(ctx).QueryRow(ctx, `SELECT `+pathCols+` FROM clinical_pathway WHERE id = $1`, id))
}

func (r *clinicalPathwayRepoPG) Update(ctx context.Context, p *ClinicalPathway) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE clinical_pathway SET name=$2, description=$3, condition=$4, category=$5,
			version=$6, active=$7, expected_duration=$8, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Name, p.Description, p.Condition, p.Category, p.Version, p.Active, p.ExpectedDuration)
	return err
}

func (r *clinicalPathwayRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM clinical_pathway WHERE id = $1`, id)
	return err
}

func (r *clinicalPathwayRepoPG) List(ctx context.Context, limit, offset int) ([]*ClinicalPathway, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM clinical_pathway`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+pathCols+` FROM clinical_pathway ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ClinicalPathway
	for rows.Next() {
		p, err := r.scanPathway(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

func (r *clinicalPathwayRepoPG) AddPhase(ctx context.Context, phase *ClinicalPathwayPhase) error {
	phase.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO clinical_pathway_phase (id, pathway_id, name, description, duration, goals, interventions, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		phase.ID, phase.PathwayID, phase.Name, phase.Description, phase.Duration, phase.Goals, phase.Interventions, phase.SortOrder)
	return err
}

func (r *clinicalPathwayRepoPG) GetPhases(ctx context.Context, pathwayID uuid.UUID) ([]*ClinicalPathwayPhase, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, pathway_id, name, description, duration, goals, interventions, sort_order
		FROM clinical_pathway_phase WHERE pathway_id = $1 ORDER BY sort_order`, pathwayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ClinicalPathwayPhase
	for rows.Next() {
		var phase ClinicalPathwayPhase
		if err := rows.Scan(&phase.ID, &phase.PathwayID, &phase.Name, &phase.Description, &phase.Duration, &phase.Goals, &phase.Interventions, &phase.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, &phase)
	}
	return items, nil
}

// =========== Patient Pathway Enrollment Repository ===========

type patientPathwayEnrollmentRepoPG struct{ pool *pgxpool.Pool }

func NewPatientPathwayEnrollmentRepoPG(pool *pgxpool.Pool) PatientPathwayEnrollmentRepository {
	return &patientPathwayEnrollmentRepoPG{pool: pool}
}

func (r *patientPathwayEnrollmentRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const enrollCols = `id, pathway_id, patient_id, practitioner_id, status, current_phase_id,
	enrolled_at, completed_at, note, created_at, updated_at`

func (r *patientPathwayEnrollmentRepoPG) scanEnrollment(row pgx.Row) (*PatientPathwayEnrollment, error) {
	var e PatientPathwayEnrollment
	err := row.Scan(&e.ID, &e.PathwayID, &e.PatientID, &e.PractitionerID, &e.Status, &e.CurrentPhaseID,
		&e.EnrolledAt, &e.CompletedAt, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	return &e, err
}

func (r *patientPathwayEnrollmentRepoPG) Create(ctx context.Context, e *PatientPathwayEnrollment) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_pathway_enrollment (id, pathway_id, patient_id, practitioner_id, status,
			current_phase_id, enrolled_at, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.PathwayID, e.PatientID, e.PractitionerID, e.Status,
		e.CurrentPhaseID, e.EnrolledAt, e.Note)
	return err
}

func (r *patientPathwayEnrollmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PatientPathwayEnrollment, error) {
	return r.scanEnrollment(r.conn(ctx).QueryRow(ctx, `SELECT `+enrollCols+` FROM patient_pathway_enrollment WHERE id = $1`, id))
}

func (r *patientPathwayEnrollmentRepoPG) Update(ctx context.Context, e *PatientPathwayEnrollment) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE patient_pathway_enrollment SET status=$2, current_phase_id=$3, completed_at=$4, note=$5, updated_at=NOW()
		WHERE id = $1`,
		e.ID, e.Status, e.CurrentPhaseID, e.CompletedAt, e.Note)
	return err
}

func (r *patientPathwayEnrollmentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient_pathway_enrollment WHERE id = $1`, id)
	return err
}

func (r *patientPathwayEnrollmentRepoPG) List(ctx context.Context, limit, offset int) ([]*PatientPathwayEnrollment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient_pathway_enrollment`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+enrollCols+` FROM patient_pathway_enrollment ORDER BY enrolled_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PatientPathwayEnrollment
	for rows.Next() {
		e, err := r.scanEnrollment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *patientPathwayEnrollmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientPathwayEnrollment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient_pathway_enrollment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+enrollCols+` FROM patient_pathway_enrollment WHERE patient_id = $1 ORDER BY enrolled_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PatientPathwayEnrollment
	for rows.Next() {
		e, err := r.scanEnrollment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

// =========== Formulary Repository ===========

type formularyRepoPG struct{ pool *pgxpool.Pool }

func NewFormularyRepoPG(pool *pgxpool.Pool) FormularyRepository { return &formularyRepoPG{pool: pool} }

func (r *formularyRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const formCols = `id, name, description, organization_id, effective_date, expiration_date, version, active, created_at, updated_at`

func (r *formularyRepoPG) scanFormulary(row pgx.Row) (*Formulary, error) {
	var f Formulary
	err := row.Scan(&f.ID, &f.Name, &f.Description, &f.OrganizationID, &f.EffectiveDate, &f.ExpirationDate, &f.Version, &f.Active, &f.CreatedAt, &f.UpdatedAt)
	return &f, err
}

func (r *formularyRepoPG) Create(ctx context.Context, f *Formulary) error {
	f.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO formulary (id, name, description, organization_id, effective_date, expiration_date, version, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		f.ID, f.Name, f.Description, f.OrganizationID, f.EffectiveDate, f.ExpirationDate, f.Version, f.Active)
	return err
}

func (r *formularyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Formulary, error) {
	return r.scanFormulary(r.conn(ctx).QueryRow(ctx, `SELECT `+formCols+` FROM formulary WHERE id = $1`, id))
}

func (r *formularyRepoPG) Update(ctx context.Context, f *Formulary) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE formulary SET name=$2, description=$3, effective_date=$4, expiration_date=$5,
			version=$6, active=$7, updated_at=NOW()
		WHERE id = $1`,
		f.ID, f.Name, f.Description, f.EffectiveDate, f.ExpirationDate, f.Version, f.Active)
	return err
}

func (r *formularyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM formulary WHERE id = $1`, id)
	return err
}

func (r *formularyRepoPG) List(ctx context.Context, limit, offset int) ([]*Formulary, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM formulary`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+formCols+` FROM formulary ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Formulary
	for rows.Next() {
		f, err := r.scanFormulary(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, f)
	}
	return items, total, nil
}

func (r *formularyRepoPG) AddItem(ctx context.Context, item *FormularyItem) error {
	item.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO formulary_item (id, formulary_id, medication_id, medication_name,
			tier_level, requires_prior_auth, step_therapy_req, quantity_limit, preferred_status, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		item.ID, item.FormularyID, item.MedicationID, item.MedicationName,
		item.TierLevel, item.RequiresPriorAuth, item.StepTherapyReq, item.QuantityLimit, item.PreferredStatus, item.Note)
	return err
}

func (r *formularyRepoPG) GetItems(ctx context.Context, formularyID uuid.UUID) ([]*FormularyItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, formulary_id, medication_id, medication_name,
			tier_level, requires_prior_auth, step_therapy_req, quantity_limit, preferred_status, note
		FROM formulary_item WHERE formulary_id = $1`, formularyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*FormularyItem
	for rows.Next() {
		var item FormularyItem
		if err := rows.Scan(&item.ID, &item.FormularyID, &item.MedicationID, &item.MedicationName,
			&item.TierLevel, &item.RequiresPriorAuth, &item.StepTherapyReq, &item.QuantityLimit, &item.PreferredStatus, &item.Note); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// =========== Medication Reconciliation Repository ===========

type medReconcRepoPG struct{ pool *pgxpool.Pool }

func NewMedReconciliationRepoPG(pool *pgxpool.Pool) MedReconciliationRepository {
	return &medReconcRepoPG{pool: pool}
}

func (r *medReconcRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const mrCols = `id, patient_id, encounter_id, practitioner_id, status, reconc_type, completed_at, note, created_at, updated_at`

func (r *medReconcRepoPG) scanMR(row pgx.Row) (*MedicationReconciliation, error) {
	var mr MedicationReconciliation
	err := row.Scan(&mr.ID, &mr.PatientID, &mr.EncounterID, &mr.PractitionerID, &mr.Status, &mr.ReconcType, &mr.CompletedAt, &mr.Note, &mr.CreatedAt, &mr.UpdatedAt)
	return &mr, err
}

func (r *medReconcRepoPG) Create(ctx context.Context, mr *MedicationReconciliation) error {
	mr.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_reconciliation (id, patient_id, encounter_id, practitioner_id, status, reconc_type, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		mr.ID, mr.PatientID, mr.EncounterID, mr.PractitionerID, mr.Status, mr.ReconcType, mr.Note)
	return err
}

func (r *medReconcRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicationReconciliation, error) {
	return r.scanMR(r.conn(ctx).QueryRow(ctx, `SELECT `+mrCols+` FROM medication_reconciliation WHERE id = $1`, id))
}

func (r *medReconcRepoPG) Update(ctx context.Context, mr *MedicationReconciliation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication_reconciliation SET status=$2, completed_at=$3, note=$4, updated_at=NOW()
		WHERE id = $1`,
		mr.ID, mr.Status, mr.CompletedAt, mr.Note)
	return err
}

func (r *medReconcRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_reconciliation WHERE id = $1`, id)
	return err
}

func (r *medReconcRepoPG) List(ctx context.Context, limit, offset int) ([]*MedicationReconciliation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_reconciliation`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mrCols+` FROM medication_reconciliation ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationReconciliation
	for rows.Next() {
		mr, err := r.scanMR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mr)
	}
	return items, total, nil
}

func (r *medReconcRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationReconciliation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_reconciliation WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+mrCols+` FROM medication_reconciliation WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationReconciliation
	for rows.Next() {
		mr, err := r.scanMR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mr)
	}
	return items, total, nil
}

func (r *medReconcRepoPG) AddItem(ctx context.Context, item *MedicationReconciliationItem) error {
	item.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_reconciliation_item (id, reconciliation_id, medication_id, medication_name,
			source_list, dose, frequency, route, action, reason, verified_by_id, verified_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		item.ID, item.ReconciliationID, item.MedicationID, item.MedicationName,
		item.SourceList, item.Dose, item.Frequency, item.Route, item.Action, item.Reason, item.VerifiedByID, item.VerifiedAt)
	return err
}

func (r *medReconcRepoPG) GetItems(ctx context.Context, reconciliationID uuid.UUID) ([]*MedicationReconciliationItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, reconciliation_id, medication_id, medication_name,
			source_list, dose, frequency, route, action, reason, verified_by_id, verified_at
		FROM medication_reconciliation_item WHERE reconciliation_id = $1`, reconciliationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*MedicationReconciliationItem
	for rows.Next() {
		var item MedicationReconciliationItem
		if err := rows.Scan(&item.ID, &item.ReconciliationID, &item.MedicationID, &item.MedicationName,
			&item.SourceList, &item.Dose, &item.Frequency, &item.Route, &item.Action, &item.Reason, &item.VerifiedByID, &item.VerifiedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}
