package nursing

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

// =========== FlowsheetTemplate Repository ===========

type flowsheetTemplateRepoPG struct{ pool *pgxpool.Pool }

func NewFlowsheetTemplateRepoPG(pool *pgxpool.Pool) FlowsheetTemplateRepository {
	return &flowsheetTemplateRepoPG{pool: pool}
}

func (r *flowsheetTemplateRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const templateCols = `id, name, description, category, is_active, created_by, created_at, updated_at`

func (r *flowsheetTemplateRepoPG) scanTemplate(row pgx.Row) (*FlowsheetTemplate, error) {
	var t FlowsheetTemplate
	err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Category, &t.IsActive, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

func (r *flowsheetTemplateRepoPG) Create(ctx context.Context, t *FlowsheetTemplate) error {
	t.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO flowsheet_template (id, name, description, category, is_active, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		t.ID, t.Name, t.Description, t.Category, t.IsActive, t.CreatedBy)
	return err
}

func (r *flowsheetTemplateRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*FlowsheetTemplate, error) {
	return r.scanTemplate(r.conn(ctx).QueryRow(ctx, `SELECT `+templateCols+` FROM flowsheet_template WHERE id = $1`, id))
}

func (r *flowsheetTemplateRepoPG) Update(ctx context.Context, t *FlowsheetTemplate) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE flowsheet_template SET name=$2, description=$3, category=$4, is_active=$5, updated_at=NOW()
		WHERE id = $1`,
		t.ID, t.Name, t.Description, t.Category, t.IsActive)
	return err
}

func (r *flowsheetTemplateRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM flowsheet_template WHERE id = $1`, id)
	return err
}

func (r *flowsheetTemplateRepoPG) List(ctx context.Context, limit, offset int) ([]*FlowsheetTemplate, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM flowsheet_template`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+templateCols+` FROM flowsheet_template ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FlowsheetTemplate
	for rows.Next() {
		t, err := r.scanTemplate(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *flowsheetTemplateRepoPG) AddRow(ctx context.Context, fr *FlowsheetRow) error {
	fr.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO flowsheet_row (id, template_id, label, data_type, unit, allowed_values, sort_order, is_required)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		fr.ID, fr.TemplateID, fr.Label, fr.DataType, fr.Unit, fr.AllowedValues, fr.SortOrder, fr.IsRequired)
	return err
}

func (r *flowsheetTemplateRepoPG) GetRows(ctx context.Context, templateID uuid.UUID) ([]*FlowsheetRow, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, template_id, label, data_type, unit, allowed_values, sort_order, is_required
		FROM flowsheet_row WHERE template_id = $1 ORDER BY sort_order`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*FlowsheetRow
	for rows.Next() {
		var fr FlowsheetRow
		if err := rows.Scan(&fr.ID, &fr.TemplateID, &fr.Label, &fr.DataType, &fr.Unit, &fr.AllowedValues, &fr.SortOrder, &fr.IsRequired); err != nil {
			return nil, err
		}
		items = append(items, &fr)
	}
	return items, nil
}

// =========== FlowsheetEntry Repository ===========

type flowsheetEntryRepoPG struct{ pool *pgxpool.Pool }

func NewFlowsheetEntryRepoPG(pool *pgxpool.Pool) FlowsheetEntryRepository {
	return &flowsheetEntryRepoPG{pool: pool}
}

func (r *flowsheetEntryRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const entryCols = `id, template_id, row_id, patient_id, encounter_id, value_text, value_numeric, recorded_at, recorded_by_id, note, created_at`

func (r *flowsheetEntryRepoPG) scanEntry(row pgx.Row) (*FlowsheetEntry, error) {
	var e FlowsheetEntry
	err := row.Scan(&e.ID, &e.TemplateID, &e.RowID, &e.PatientID, &e.EncounterID,
		&e.ValueText, &e.ValueNumeric, &e.RecordedAt, &e.RecordedByID, &e.Note, &e.CreatedAt)
	return &e, err
}

func (r *flowsheetEntryRepoPG) Create(ctx context.Context, e *FlowsheetEntry) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO flowsheet_entry (id, template_id, row_id, patient_id, encounter_id,
			value_text, value_numeric, recorded_at, recorded_by_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.ID, e.TemplateID, e.RowID, e.PatientID, e.EncounterID,
		e.ValueText, e.ValueNumeric, e.RecordedAt, e.RecordedByID, e.Note)
	return err
}

func (r *flowsheetEntryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*FlowsheetEntry, error) {
	return r.scanEntry(r.conn(ctx).QueryRow(ctx, `SELECT `+entryCols+` FROM flowsheet_entry WHERE id = $1`, id))
}

func (r *flowsheetEntryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM flowsheet_entry WHERE id = $1`, id)
	return err
}

func (r *flowsheetEntryRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM flowsheet_entry WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+entryCols+` FROM flowsheet_entry WHERE patient_id = $1 ORDER BY recorded_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FlowsheetEntry
	for rows.Next() {
		e, err := r.scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *flowsheetEntryRepoPG) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM flowsheet_entry WHERE encounter_id = $1`, encounterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+entryCols+` FROM flowsheet_entry WHERE encounter_id = $1 ORDER BY recorded_at DESC LIMIT $2 OFFSET $3`, encounterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FlowsheetEntry
	for rows.Next() {
		e, err := r.scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

func (r *flowsheetEntryRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*FlowsheetEntry, int, error) {
	query := `SELECT ` + entryCols + ` FROM flowsheet_entry WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM flowsheet_entry WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient_id"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["encounter_id"]; ok {
		query += fmt.Sprintf(` AND encounter_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND encounter_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["template_id"]; ok {
		query += fmt.Sprintf(` AND template_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND template_id = $%d`, idx)
		args = append(args, p)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY recorded_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FlowsheetEntry
	for rows.Next() {
		e, err := r.scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, nil
}

// =========== NursingAssessment Repository ===========

type nursingAssessmentRepoPG struct{ pool *pgxpool.Pool }

func NewNursingAssessmentRepoPG(pool *pgxpool.Pool) NursingAssessmentRepository {
	return &nursingAssessmentRepoPG{pool: pool}
}

func (r *nursingAssessmentRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const assessCols = `id, patient_id, encounter_id, nurse_id, assessment_type, assessment_data, status, completed_at, note, created_at, updated_at`

func (r *nursingAssessmentRepoPG) scanAssessment(row pgx.Row) (*NursingAssessment, error) {
	var a NursingAssessment
	err := row.Scan(&a.ID, &a.PatientID, &a.EncounterID, &a.NurseID, &a.AssessmentType,
		&a.AssessmentData, &a.Status, &a.CompletedAt, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *nursingAssessmentRepoPG) Create(ctx context.Context, a *NursingAssessment) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO nursing_assessment (id, patient_id, encounter_id, nurse_id, assessment_type,
			assessment_data, status, completed_at, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.PatientID, a.EncounterID, a.NurseID, a.AssessmentType,
		a.AssessmentData, a.Status, a.CompletedAt, a.Note)
	return err
}

func (r *nursingAssessmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*NursingAssessment, error) {
	return r.scanAssessment(r.conn(ctx).QueryRow(ctx, `SELECT `+assessCols+` FROM nursing_assessment WHERE id = $1`, id))
}

func (r *nursingAssessmentRepoPG) Update(ctx context.Context, a *NursingAssessment) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE nursing_assessment SET assessment_data=$2, status=$3, completed_at=$4, note=$5, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.AssessmentData, a.Status, a.CompletedAt, a.Note)
	return err
}

func (r *nursingAssessmentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM nursing_assessment WHERE id = $1`, id)
	return err
}

func (r *nursingAssessmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM nursing_assessment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+assessCols+` FROM nursing_assessment WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NursingAssessment
	for rows.Next() {
		a, err := r.scanAssessment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *nursingAssessmentRepoPG) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM nursing_assessment WHERE encounter_id = $1`, encounterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+assessCols+` FROM nursing_assessment WHERE encounter_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, encounterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NursingAssessment
	for rows.Next() {
		a, err := r.scanAssessment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== FallRisk Repository ===========

type fallRiskRepoPG struct{ pool *pgxpool.Pool }

func NewFallRiskRepoPG(pool *pgxpool.Pool) FallRiskRepository {
	return &fallRiskRepoPG{pool: pool}
}

func (r *fallRiskRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const fallRiskCols = `id, patient_id, encounter_id, assessed_by_id, tool_used, total_score, risk_level,
	history_of_falls, medications, gait_balance, mental_status, interventions, note, assessed_at, created_at`

func (r *fallRiskRepoPG) scanFallRisk(row pgx.Row) (*FallRiskAssessment, error) {
	var a FallRiskAssessment
	err := row.Scan(&a.ID, &a.PatientID, &a.EncounterID, &a.AssessedByID, &a.ToolUsed, &a.TotalScore, &a.RiskLevel,
		&a.HistoryOfFalls, &a.Medications, &a.GaitBalance, &a.MentalStatus, &a.Interventions, &a.Note, &a.AssessedAt, &a.CreatedAt)
	return &a, err
}

func (r *fallRiskRepoPG) Create(ctx context.Context, a *FallRiskAssessment) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO fall_risk_assessment (id, patient_id, encounter_id, assessed_by_id, tool_used, total_score, risk_level,
			history_of_falls, medications, gait_balance, mental_status, interventions, note, assessed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		a.ID, a.PatientID, a.EncounterID, a.AssessedByID, a.ToolUsed, a.TotalScore, a.RiskLevel,
		a.HistoryOfFalls, a.Medications, a.GaitBalance, a.MentalStatus, a.Interventions, a.Note, a.AssessedAt)
	return err
}

func (r *fallRiskRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*FallRiskAssessment, error) {
	return r.scanFallRisk(r.conn(ctx).QueryRow(ctx, `SELECT `+fallRiskCols+` FROM fall_risk_assessment WHERE id = $1`, id))
}

func (r *fallRiskRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*FallRiskAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM fall_risk_assessment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+fallRiskCols+` FROM fall_risk_assessment WHERE patient_id = $1 ORDER BY assessed_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*FallRiskAssessment
	for rows.Next() {
		a, err := r.scanFallRisk(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== SkinAssessment Repository ===========

type skinAssessmentRepoPG struct{ pool *pgxpool.Pool }

func NewSkinAssessmentRepoPG(pool *pgxpool.Pool) SkinAssessmentRepository {
	return &skinAssessmentRepoPG{pool: pool}
}

func (r *skinAssessmentRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const skinCols = `id, patient_id, encounter_id, assessed_by_id, tool_used, total_score, risk_level,
	skin_integrity, moisture_level, mobility, nutrition, wound_present, wound_location, wound_stage,
	interventions, note, assessed_at, created_at`

func (r *skinAssessmentRepoPG) scanSkin(row pgx.Row) (*SkinAssessment, error) {
	var a SkinAssessment
	err := row.Scan(&a.ID, &a.PatientID, &a.EncounterID, &a.AssessedByID, &a.ToolUsed, &a.TotalScore, &a.RiskLevel,
		&a.SkinIntegrity, &a.MoistureLevel, &a.Mobility, &a.Nutrition, &a.WoundPresent, &a.WoundLocation, &a.WoundStage,
		&a.Interventions, &a.Note, &a.AssessedAt, &a.CreatedAt)
	return &a, err
}

func (r *skinAssessmentRepoPG) Create(ctx context.Context, a *SkinAssessment) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO skin_assessment (id, patient_id, encounter_id, assessed_by_id, tool_used, total_score, risk_level,
			skin_integrity, moisture_level, mobility, nutrition, wound_present, wound_location, wound_stage,
			interventions, note, assessed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		a.ID, a.PatientID, a.EncounterID, a.AssessedByID, a.ToolUsed, a.TotalScore, a.RiskLevel,
		a.SkinIntegrity, a.MoistureLevel, a.Mobility, a.Nutrition, a.WoundPresent, a.WoundLocation, a.WoundStage,
		a.Interventions, a.Note, a.AssessedAt)
	return err
}

func (r *skinAssessmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SkinAssessment, error) {
	return r.scanSkin(r.conn(ctx).QueryRow(ctx, `SELECT `+skinCols+` FROM skin_assessment WHERE id = $1`, id))
}

func (r *skinAssessmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SkinAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM skin_assessment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+skinCols+` FROM skin_assessment WHERE patient_id = $1 ORDER BY assessed_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SkinAssessment
	for rows.Next() {
		a, err := r.scanSkin(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== PainAssessment Repository ===========

type painAssessmentRepoPG struct{ pool *pgxpool.Pool }

func NewPainAssessmentRepoPG(pool *pgxpool.Pool) PainAssessmentRepository {
	return &painAssessmentRepoPG{pool: pool}
}

func (r *painAssessmentRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const painCols = `id, patient_id, encounter_id, assessed_by_id, tool_used, pain_score, pain_location,
	pain_character, pain_duration, pain_radiation, aggravating, alleviating, interventions,
	reassess_score, note, assessed_at, created_at`

func (r *painAssessmentRepoPG) scanPain(row pgx.Row) (*PainAssessment, error) {
	var a PainAssessment
	err := row.Scan(&a.ID, &a.PatientID, &a.EncounterID, &a.AssessedByID, &a.ToolUsed, &a.PainScore, &a.PainLocation,
		&a.PainCharacter, &a.PainDuration, &a.PainRadiation, &a.Aggravating, &a.Alleviating, &a.Interventions,
		&a.ReassessScore, &a.Note, &a.AssessedAt, &a.CreatedAt)
	return &a, err
}

func (r *painAssessmentRepoPG) Create(ctx context.Context, a *PainAssessment) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO pain_assessment (id, patient_id, encounter_id, assessed_by_id, tool_used, pain_score, pain_location,
			pain_character, pain_duration, pain_radiation, aggravating, alleviating, interventions,
			reassess_score, note, assessed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		a.ID, a.PatientID, a.EncounterID, a.AssessedByID, a.ToolUsed, a.PainScore, a.PainLocation,
		a.PainCharacter, a.PainDuration, a.PainRadiation, a.Aggravating, a.Alleviating, a.Interventions,
		a.ReassessScore, a.Note, a.AssessedAt)
	return err
}

func (r *painAssessmentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PainAssessment, error) {
	return r.scanPain(r.conn(ctx).QueryRow(ctx, `SELECT `+painCols+` FROM pain_assessment WHERE id = $1`, id))
}

func (r *painAssessmentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PainAssessment, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM pain_assessment WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+painCols+` FROM pain_assessment WHERE patient_id = $1 ORDER BY assessed_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PainAssessment
	for rows.Next() {
		a, err := r.scanPain(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== LinesDrains Repository ===========

type linesDrainsRepoPG struct{ pool *pgxpool.Pool }

func NewLinesDrainsRepoPG(pool *pgxpool.Pool) LinesDrainsRepository {
	return &linesDrainsRepoPG{pool: pool}
}

func (r *linesDrainsRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const linesCols = `id, patient_id, encounter_id, type, description, site, size,
	inserted_at, inserted_by_id, removed_at, removed_by_id, status, device_id, note, created_at, updated_at`

func (r *linesDrainsRepoPG) scanLines(row pgx.Row) (*LinesDrainsAirways, error) {
	var l LinesDrainsAirways
	err := row.Scan(&l.ID, &l.PatientID, &l.EncounterID, &l.Type, &l.Description, &l.Site, &l.Size,
		&l.InsertedAt, &l.InsertedByID, &l.RemovedAt, &l.RemovedByID, &l.Status, &l.DeviceID, &l.Note, &l.CreatedAt, &l.UpdatedAt)
	return &l, err
}

func (r *linesDrainsRepoPG) Create(ctx context.Context, l *LinesDrainsAirways) error {
	l.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO lines_drains_airways (id, patient_id, encounter_id, type, description, site, size,
			inserted_at, inserted_by_id, status, device_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		l.ID, l.PatientID, l.EncounterID, l.Type, l.Description, l.Site, l.Size,
		l.InsertedAt, l.InsertedByID, l.Status, l.DeviceID, l.Note)
	return err
}

func (r *linesDrainsRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*LinesDrainsAirways, error) {
	return r.scanLines(r.conn(ctx).QueryRow(ctx, `SELECT `+linesCols+` FROM lines_drains_airways WHERE id = $1`, id))
}

func (r *linesDrainsRepoPG) Update(ctx context.Context, l *LinesDrainsAirways) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE lines_drains_airways SET status=$2, removed_at=$3, removed_by_id=$4, note=$5, updated_at=NOW()
		WHERE id = $1`,
		l.ID, l.Status, l.RemovedAt, l.RemovedByID, l.Note)
	return err
}

func (r *linesDrainsRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM lines_drains_airways WHERE id = $1`, id)
	return err
}

func (r *linesDrainsRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM lines_drains_airways WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+linesCols+` FROM lines_drains_airways WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*LinesDrainsAirways
	for rows.Next() {
		l, err := r.scanLines(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

func (r *linesDrainsRepoPG) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM lines_drains_airways WHERE encounter_id = $1`, encounterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+linesCols+` FROM lines_drains_airways WHERE encounter_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, encounterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*LinesDrainsAirways
	for rows.Next() {
		l, err := r.scanLines(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

// =========== Restraint Repository ===========

type restraintRepoPG struct{ pool *pgxpool.Pool }

func NewRestraintRepoPG(pool *pgxpool.Pool) RestraintRepository {
	return &restraintRepoPG{pool: pool}
}

func (r *restraintRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const restraintCols = `id, patient_id, encounter_id, restraint_type, reason, body_site,
	applied_at, applied_by_id, removed_at, removed_by_id, order_id,
	last_assessed_at, last_assessed_by_id, skin_condition, circulation, note, created_at, updated_at`

func (r *restraintRepoPG) scanRestraint(row pgx.Row) (*RestraintRecord, error) {
	var rec RestraintRecord
	err := row.Scan(&rec.ID, &rec.PatientID, &rec.EncounterID, &rec.RestraintType, &rec.Reason, &rec.BodySite,
		&rec.AppliedAt, &rec.AppliedByID, &rec.RemovedAt, &rec.RemovedByID, &rec.OrderID,
		&rec.LastAssessedAt, &rec.LastAssessedByID, &rec.SkinCondition, &rec.Circulation, &rec.Note, &rec.CreatedAt, &rec.UpdatedAt)
	return &rec, err
}

func (r *restraintRepoPG) Create(ctx context.Context, rec *RestraintRecord) error {
	rec.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO restraint_record (id, patient_id, encounter_id, restraint_type, reason, body_site,
			applied_at, applied_by_id, order_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rec.ID, rec.PatientID, rec.EncounterID, rec.RestraintType, rec.Reason, rec.BodySite,
		rec.AppliedAt, rec.AppliedByID, rec.OrderID, rec.Note)
	return err
}

func (r *restraintRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*RestraintRecord, error) {
	return r.scanRestraint(r.conn(ctx).QueryRow(ctx, `SELECT `+restraintCols+` FROM restraint_record WHERE id = $1`, id))
}

func (r *restraintRepoPG) Update(ctx context.Context, rec *RestraintRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE restraint_record SET removed_at=$2, removed_by_id=$3,
			last_assessed_at=$4, last_assessed_by_id=$5, skin_condition=$6, circulation=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		rec.ID, rec.RemovedAt, rec.RemovedByID,
		rec.LastAssessedAt, rec.LastAssessedByID, rec.SkinCondition, rec.Circulation, rec.Note)
	return err
}

func (r *restraintRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*RestraintRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM restraint_record WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+restraintCols+` FROM restraint_record WHERE patient_id = $1 ORDER BY applied_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RestraintRecord
	for rows.Next() {
		rec, err := r.scanRestraint(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rec)
	}
	return items, total, nil
}

// =========== IntakeOutput Repository ===========

type intakeOutputRepoPG struct{ pool *pgxpool.Pool }

func NewIntakeOutputRepoPG(pool *pgxpool.Pool) IntakeOutputRepository {
	return &intakeOutputRepoPG{pool: pool}
}

func (r *intakeOutputRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const ioCols = `id, patient_id, encounter_id, category, type, volume, unit, route, recorded_at, recorded_by_id, note, created_at`

func (r *intakeOutputRepoPG) scanIO(row pgx.Row) (*IntakeOutputRecord, error) {
	var rec IntakeOutputRecord
	err := row.Scan(&rec.ID, &rec.PatientID, &rec.EncounterID, &rec.Category, &rec.Type,
		&rec.Volume, &rec.Unit, &rec.Route, &rec.RecordedAt, &rec.RecordedByID, &rec.Note, &rec.CreatedAt)
	return &rec, err
}

func (r *intakeOutputRepoPG) Create(ctx context.Context, rec *IntakeOutputRecord) error {
	rec.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO intake_output_record (id, patient_id, encounter_id, category, type, volume, unit, route,
			recorded_at, recorded_by_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		rec.ID, rec.PatientID, rec.EncounterID, rec.Category, rec.Type, rec.Volume, rec.Unit, rec.Route,
		rec.RecordedAt, rec.RecordedByID, rec.Note)
	return err
}

func (r *intakeOutputRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*IntakeOutputRecord, error) {
	return r.scanIO(r.conn(ctx).QueryRow(ctx, `SELECT `+ioCols+` FROM intake_output_record WHERE id = $1`, id))
}

func (r *intakeOutputRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM intake_output_record WHERE id = $1`, id)
	return err
}

func (r *intakeOutputRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM intake_output_record WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ioCols+` FROM intake_output_record WHERE patient_id = $1 ORDER BY recorded_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*IntakeOutputRecord
	for rows.Next() {
		rec, err := r.scanIO(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rec)
	}
	return items, total, nil
}

func (r *intakeOutputRepoPG) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM intake_output_record WHERE encounter_id = $1`, encounterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+ioCols+` FROM intake_output_record WHERE encounter_id = $1 ORDER BY recorded_at DESC LIMIT $2 OFFSET $3`, encounterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*IntakeOutputRecord
	for rows.Next() {
		rec, err := r.scanIO(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rec)
	}
	return items, total, nil
}
