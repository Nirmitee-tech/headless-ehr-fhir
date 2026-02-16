package encounter

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
)

type repoPG struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) Repository {
	return &repoPG{pool: pool}
}

type querier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (r *repoPG) conn(ctx context.Context) querier {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const encCols = `id, fhir_id, status, class_code, class_display, type_code, type_display,
	service_type_code, service_type_display, priority_code,
	patient_id, primary_practitioner_id, service_provider_id, department_id,
	period_start, period_end, length_minutes,
	location_id, bed_id,
	admit_source_code, admit_source_display,
	discharge_disposition_code, discharge_disposition_display,
	re_admission, is_telehealth, telehealth_platform, reason_text,
	drg_code, drg_type, created_at, updated_at`

func (r *repoPG) Create(ctx context.Context, enc *Encounter) error {
	enc.ID = uuid.New()
	if enc.FHIRID == "" {
		enc.FHIRID = enc.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO encounter (
			id, fhir_id, status, class_code, class_display, type_code, type_display,
			service_type_code, service_type_display, priority_code,
			patient_id, primary_practitioner_id, service_provider_id, department_id,
			period_start, period_end, length_minutes,
			location_id, bed_id,
			admit_source_code, admit_source_display,
			discharge_disposition_code, discharge_disposition_display,
			re_admission, is_telehealth, telehealth_platform, reason_text,
			drg_code, drg_type
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15,$16,$17,$18,$19,
			$20,$21,$22,$23,$24,$25,$26,$27,$28,$29
		)`,
		enc.ID, enc.FHIRID, enc.Status, enc.ClassCode, enc.ClassDisplay, enc.TypeCode, enc.TypeDisplay,
		enc.ServiceTypeCode, enc.ServiceTypeDisplay, enc.PriorityCode,
		enc.PatientID, enc.PrimaryPractitionerID, enc.ServiceProviderID, enc.DepartmentID,
		enc.PeriodStart, enc.PeriodEnd, enc.LengthMinutes,
		enc.LocationID, enc.BedID,
		enc.AdmitSourceCode, enc.AdmitSourceDisplay,
		enc.DischargeDispositionCode, enc.DischargeDispositionDisp,
		enc.ReAdmission, enc.IsTelehealth, enc.TelehealthPlatform, enc.ReasonText,
		enc.DRGCode, enc.DRGType,
	)
	return err
}

func (r *repoPG) GetByID(ctx context.Context, id uuid.UUID) (*Encounter, error) {
	return scanEnc(r.conn(ctx).QueryRow(ctx, `SELECT `+encCols+` FROM encounter WHERE id = $1`, id))
}

func (r *repoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Encounter, error) {
	return scanEnc(r.conn(ctx).QueryRow(ctx, `SELECT `+encCols+` FROM encounter WHERE fhir_id = $1`, fhirID))
}

func (r *repoPG) Update(ctx context.Context, enc *Encounter) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE encounter SET
			status=$2, class_code=$3, class_display=$4, type_code=$5, type_display=$6,
			service_type_code=$7, service_type_display=$8, priority_code=$9,
			primary_practitioner_id=$10, service_provider_id=$11, department_id=$12,
			period_start=$13, period_end=$14, length_minutes=$15,
			location_id=$16, bed_id=$17,
			admit_source_code=$18, admit_source_display=$19,
			discharge_disposition_code=$20, discharge_disposition_display=$21,
			re_admission=$22, is_telehealth=$23, telehealth_platform=$24, reason_text=$25,
			drg_code=$26, drg_type=$27, updated_at=NOW()
		WHERE id = $1`,
		enc.ID, enc.Status, enc.ClassCode, enc.ClassDisplay, enc.TypeCode, enc.TypeDisplay,
		enc.ServiceTypeCode, enc.ServiceTypeDisplay, enc.PriorityCode,
		enc.PrimaryPractitionerID, enc.ServiceProviderID, enc.DepartmentID,
		enc.PeriodStart, enc.PeriodEnd, enc.LengthMinutes,
		enc.LocationID, enc.BedID,
		enc.AdmitSourceCode, enc.AdmitSourceDisplay,
		enc.DischargeDispositionCode, enc.DischargeDispositionDisp,
		enc.ReAdmission, enc.IsTelehealth, enc.TelehealthPlatform, enc.ReasonText,
		enc.DRGCode, enc.DRGType,
	)
	return err
}

func (r *repoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM encounter WHERE id = $1`, id)
	return err
}

func (r *repoPG) List(ctx context.Context, limit, offset int) ([]*Encounter, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM encounter`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+encCols+` FROM encounter ORDER BY period_start DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return collectEncs(rows, total)
}

func (r *repoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Encounter, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM encounter WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx,
		`SELECT `+encCols+` FROM encounter WHERE patient_id = $1 ORDER BY period_start DESC LIMIT $2 OFFSET $3`,
		patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return collectEncs(rows, total)
}

var encounterSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"class":   {Type: fhir.SearchParamToken, Column: "class_code"},
	"date":    {Type: fhir.SearchParamDate, Column: "period_start"},
}

func (r *repoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Encounter, int, error) {
	qb := fhir.NewSearchQuery("encounter", encCols)
	qb.ApplyParams(params, encounterSearchParams)
	qb.OrderBy("period_start DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return collectEncs(rows, total)
}

// Participants
func (r *repoPG) AddParticipant(ctx context.Context, p *EncounterParticipant) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO encounter_participant (id, encounter_id, practitioner_id, type_code, type_display, period_start, period_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.EncounterID, p.PractitionerID, p.TypeCode, p.TypeDisplay, p.PeriodStart, p.PeriodEnd,
	)
	return err
}

func (r *repoPG) GetParticipants(ctx context.Context, encounterID uuid.UUID) ([]*EncounterParticipant, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, encounter_id, practitioner_id, type_code, type_display, period_start, period_end
		FROM encounter_participant WHERE encounter_id = $1`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []*EncounterParticipant
	for rows.Next() {
		var p EncounterParticipant
		if err := rows.Scan(&p.ID, &p.EncounterID, &p.PractitionerID, &p.TypeCode, &p.TypeDisplay, &p.PeriodStart, &p.PeriodEnd); err != nil {
			return nil, err
		}
		parts = append(parts, &p)
	}
	return parts, nil
}

func (r *repoPG) RemoveParticipant(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM encounter_participant WHERE id = $1`, id)
	return err
}

// Diagnoses
func (r *repoPG) AddDiagnosis(ctx context.Context, d *EncounterDiagnosis) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO encounter_diagnosis (id, encounter_id, condition_id, use_code, rank)
		VALUES ($1,$2,$3,$4,$5)`,
		d.ID, d.EncounterID, d.ConditionID, d.UseCode, d.Rank,
	)
	return err
}

func (r *repoPG) GetDiagnoses(ctx context.Context, encounterID uuid.UUID) ([]*EncounterDiagnosis, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, encounter_id, condition_id, use_code, rank, created_at
		FROM encounter_diagnosis WHERE encounter_id = $1 ORDER BY rank`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diags []*EncounterDiagnosis
	for rows.Next() {
		var d EncounterDiagnosis
		if err := rows.Scan(&d.ID, &d.EncounterID, &d.ConditionID, &d.UseCode, &d.Rank, &d.CreatedAt); err != nil {
			return nil, err
		}
		diags = append(diags, &d)
	}
	return diags, nil
}

func (r *repoPG) RemoveDiagnosis(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM encounter_diagnosis WHERE id = $1`, id)
	return err
}

// Status History
func (r *repoPG) AddStatusHistory(ctx context.Context, sh *EncounterStatusHistory) error {
	sh.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO encounter_status_history (id, encounter_id, status, period_start, period_end)
		VALUES ($1,$2,$3,$4,$5)`,
		sh.ID, sh.EncounterID, sh.Status, sh.PeriodStart, sh.PeriodEnd,
	)
	return err
}

func (r *repoPG) GetStatusHistory(ctx context.Context, encounterID uuid.UUID) ([]*EncounterStatusHistory, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, encounter_id, status, period_start, period_end
		FROM encounter_status_history WHERE encounter_id = $1 ORDER BY period_start`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*EncounterStatusHistory
	for rows.Next() {
		var sh EncounterStatusHistory
		if err := rows.Scan(&sh.ID, &sh.EncounterID, &sh.Status, &sh.PeriodStart, &sh.PeriodEnd); err != nil {
			return nil, err
		}
		history = append(history, &sh)
	}
	return history, nil
}

func scanEnc(row pgx.Row) (*Encounter, error) {
	var e Encounter
	err := row.Scan(
		&e.ID, &e.FHIRID, &e.Status, &e.ClassCode, &e.ClassDisplay, &e.TypeCode, &e.TypeDisplay,
		&e.ServiceTypeCode, &e.ServiceTypeDisplay, &e.PriorityCode,
		&e.PatientID, &e.PrimaryPractitionerID, &e.ServiceProviderID, &e.DepartmentID,
		&e.PeriodStart, &e.PeriodEnd, &e.LengthMinutes,
		&e.LocationID, &e.BedID,
		&e.AdmitSourceCode, &e.AdmitSourceDisplay,
		&e.DischargeDispositionCode, &e.DischargeDispositionDisp,
		&e.ReAdmission, &e.IsTelehealth, &e.TelehealthPlatform, &e.ReasonText,
		&e.DRGCode, &e.DRGType, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func collectEncs(rows pgx.Rows, total int) ([]*Encounter, int, error) {
	var encs []*Encounter
	for rows.Next() {
		var e Encounter
		err := rows.Scan(
			&e.ID, &e.FHIRID, &e.Status, &e.ClassCode, &e.ClassDisplay, &e.TypeCode, &e.TypeDisplay,
			&e.ServiceTypeCode, &e.ServiceTypeDisplay, &e.PriorityCode,
			&e.PatientID, &e.PrimaryPractitionerID, &e.ServiceProviderID, &e.DepartmentID,
			&e.PeriodStart, &e.PeriodEnd, &e.LengthMinutes,
			&e.LocationID, &e.BedID,
			&e.AdmitSourceCode, &e.AdmitSourceDisplay,
			&e.DischargeDispositionCode, &e.DischargeDispositionDisp,
			&e.ReAdmission, &e.IsTelehealth, &e.TelehealthPlatform, &e.ReasonText,
			&e.DRGCode, &e.DRGType, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		encs = append(encs, &e)
	}
	return encs, total, nil
}
