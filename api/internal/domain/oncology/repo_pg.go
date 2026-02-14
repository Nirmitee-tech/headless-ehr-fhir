package oncology

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

// =========== Cancer Diagnosis Repository ===========

type cancerDiagnosisRepoPG struct{ pool *pgxpool.Pool }

func NewCancerDiagnosisRepoPG(pool *pgxpool.Pool) CancerDiagnosisRepository {
	return &cancerDiagnosisRepoPG{pool: pool}
}

func (r *cancerDiagnosisRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const dxCols = `id, patient_id, condition_id, diagnosis_date,
	cancer_type, cancer_site, histology_code, histology_display,
	morphology_code, morphology_display, staging_system, stage_group,
	t_stage, n_stage, m_stage, grade, laterality, current_status,
	diagnosing_provider_id, managing_provider_id,
	icd10_code, icd10_display, icdo3_topography, icdo3_morphology,
	note, created_at, updated_at`

func (r *cancerDiagnosisRepoPG) scanDx(row pgx.Row) (*CancerDiagnosis, error) {
	var d CancerDiagnosis
	err := row.Scan(&d.ID, &d.PatientID, &d.ConditionID, &d.DiagnosisDate,
		&d.CancerType, &d.CancerSite, &d.HistologyCode, &d.HistologyDisplay,
		&d.MorphologyCode, &d.MorphologyDisplay, &d.StagingSystem, &d.StageGroup,
		&d.TStage, &d.NStage, &d.MStage, &d.Grade, &d.Laterality, &d.CurrentStatus,
		&d.DiagnosingProviderID, &d.ManagingProviderID,
		&d.ICD10Code, &d.ICD10Display, &d.ICDO3Topography, &d.ICDO3Morphology,
		&d.Note, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *cancerDiagnosisRepoPG) Create(ctx context.Context, d *CancerDiagnosis) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO cancer_diagnosis (id, patient_id, condition_id, diagnosis_date,
			cancer_type, cancer_site, histology_code, histology_display,
			morphology_code, morphology_display, staging_system, stage_group,
			t_stage, n_stage, m_stage, grade, laterality, current_status,
			diagnosing_provider_id, managing_provider_id,
			icd10_code, icd10_display, icdo3_topography, icdo3_morphology, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		d.ID, d.PatientID, d.ConditionID, d.DiagnosisDate,
		d.CancerType, d.CancerSite, d.HistologyCode, d.HistologyDisplay,
		d.MorphologyCode, d.MorphologyDisplay, d.StagingSystem, d.StageGroup,
		d.TStage, d.NStage, d.MStage, d.Grade, d.Laterality, d.CurrentStatus,
		d.DiagnosingProviderID, d.ManagingProviderID,
		d.ICD10Code, d.ICD10Display, d.ICDO3Topography, d.ICDO3Morphology, d.Note)
	return err
}

func (r *cancerDiagnosisRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CancerDiagnosis, error) {
	return r.scanDx(r.conn(ctx).QueryRow(ctx, `SELECT `+dxCols+` FROM cancer_diagnosis WHERE id = $1`, id))
}

func (r *cancerDiagnosisRepoPG) Update(ctx context.Context, d *CancerDiagnosis) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE cancer_diagnosis SET current_status=$2, stage_group=$3, t_stage=$4, n_stage=$5, m_stage=$6,
			grade=$7, managing_provider_id=$8, note=$9, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.CurrentStatus, d.StageGroup, d.TStage, d.NStage, d.MStage,
		d.Grade, d.ManagingProviderID, d.Note)
	return err
}

func (r *cancerDiagnosisRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM cancer_diagnosis WHERE id = $1`, id)
	return err
}

func (r *cancerDiagnosisRepoPG) List(ctx context.Context, limit, offset int) ([]*CancerDiagnosis, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cancer_diagnosis`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+dxCols+` FROM cancer_diagnosis ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CancerDiagnosis
	for rows.Next() {
		d, err := r.scanDx(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

func (r *cancerDiagnosisRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CancerDiagnosis, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cancer_diagnosis WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+dxCols+` FROM cancer_diagnosis WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CancerDiagnosis
	for rows.Next() {
		d, err := r.scanDx(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

// =========== Treatment Protocol Repository ===========

type treatmentProtocolRepoPG struct{ pool *pgxpool.Pool }

func NewTreatmentProtocolRepoPG(pool *pgxpool.Pool) TreatmentProtocolRepository {
	return &treatmentProtocolRepoPG{pool: pool}
}

func (r *treatmentProtocolRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const protoCols = `id, cancer_diagnosis_id, protocol_name, protocol_code, protocol_type,
	intent, number_of_cycles, cycle_length_days, start_date, end_date, status,
	prescribing_provider_id, clinical_trial_id, note, created_at, updated_at`

func (r *treatmentProtocolRepoPG) scanProto(row pgx.Row) (*TreatmentProtocol, error) {
	var p TreatmentProtocol
	err := row.Scan(&p.ID, &p.CancerDiagnosisID, &p.ProtocolName, &p.ProtocolCode, &p.ProtocolType,
		&p.Intent, &p.NumberOfCycles, &p.CycleLengthDays, &p.StartDate, &p.EndDate, &p.Status,
		&p.PrescribingProviderID, &p.ClinicalTrialID, &p.Note, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *treatmentProtocolRepoPG) Create(ctx context.Context, p *TreatmentProtocol) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO treatment_protocol (id, cancer_diagnosis_id, protocol_name, protocol_code, protocol_type,
			intent, number_of_cycles, cycle_length_days, start_date, end_date, status,
			prescribing_provider_id, clinical_trial_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.CancerDiagnosisID, p.ProtocolName, p.ProtocolCode, p.ProtocolType,
		p.Intent, p.NumberOfCycles, p.CycleLengthDays, p.StartDate, p.EndDate, p.Status,
		p.PrescribingProviderID, p.ClinicalTrialID, p.Note)
	return err
}

func (r *treatmentProtocolRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TreatmentProtocol, error) {
	return r.scanProto(r.conn(ctx).QueryRow(ctx, `SELECT `+protoCols+` FROM treatment_protocol WHERE id = $1`, id))
}

func (r *treatmentProtocolRepoPG) Update(ctx context.Context, p *TreatmentProtocol) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE treatment_protocol SET protocol_name=$2, status=$3, number_of_cycles=$4,
			start_date=$5, end_date=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.ProtocolName, p.Status, p.NumberOfCycles,
		p.StartDate, p.EndDate, p.Note)
	return err
}

func (r *treatmentProtocolRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM treatment_protocol WHERE id = $1`, id)
	return err
}

func (r *treatmentProtocolRepoPG) List(ctx context.Context, limit, offset int) ([]*TreatmentProtocol, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM treatment_protocol`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+protoCols+` FROM treatment_protocol ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TreatmentProtocol
	for rows.Next() {
		p, err := r.scanProto(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

func (r *treatmentProtocolRepoPG) AddDrug(ctx context.Context, d *TreatmentProtocolDrug) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO treatment_protocol_drug (id, protocol_id, drug_name, drug_code, drug_code_system,
			route, dose_value, dose_unit, dose_calculation_method, frequency,
			administration_day, infusion_duration_min, premedication, sequence_order, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		d.ID, d.ProtocolID, d.DrugName, d.DrugCode, d.DrugCodeSystem,
		d.Route, d.DoseValue, d.DoseUnit, d.DoseCalculationMethod, d.Frequency,
		d.AdministrationDay, d.InfusionDurationMin, d.Premedication, d.SequenceOrder, d.Note)
	return err
}

func (r *treatmentProtocolRepoPG) GetDrugs(ctx context.Context, protocolID uuid.UUID) ([]*TreatmentProtocolDrug, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, protocol_id, drug_name, drug_code, drug_code_system,
			route, dose_value, dose_unit, dose_calculation_method, frequency,
			administration_day, infusion_duration_min, premedication, sequence_order, note
		FROM treatment_protocol_drug WHERE protocol_id = $1 ORDER BY sequence_order NULLS LAST`, protocolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*TreatmentProtocolDrug
	for rows.Next() {
		var d TreatmentProtocolDrug
		if err := rows.Scan(&d.ID, &d.ProtocolID, &d.DrugName, &d.DrugCode, &d.DrugCodeSystem,
			&d.Route, &d.DoseValue, &d.DoseUnit, &d.DoseCalculationMethod, &d.Frequency,
			&d.AdministrationDay, &d.InfusionDurationMin, &d.Premedication, &d.SequenceOrder, &d.Note); err != nil {
			return nil, err
		}
		items = append(items, &d)
	}
	return items, nil
}

// =========== Chemo Cycle Repository ===========

type chemoCycleRepoPG struct{ pool *pgxpool.Pool }

func NewChemoCycleRepoPG(pool *pgxpool.Pool) ChemoCycleRepository {
	return &chemoCycleRepoPG{pool: pool}
}

func (r *chemoCycleRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const chemoCycleCols = `id, protocol_id, cycle_number, planned_start_date, actual_start_date, actual_end_date,
	status, dose_reduction_pct, dose_reduction_reason, delay_days, delay_reason,
	bsa_m2, weight_kg, height_cm, creatinine_clearance, provider_id, note, created_at, updated_at`

func (r *chemoCycleRepoPG) scanCycle(row pgx.Row) (*ChemoCycle, error) {
	var c ChemoCycle
	err := row.Scan(&c.ID, &c.ProtocolID, &c.CycleNumber, &c.PlannedStartDate, &c.ActualStartDate, &c.ActualEndDate,
		&c.Status, &c.DoseReductionPct, &c.DoseReductionReason, &c.DelayDays, &c.DelayReason,
		&c.BSAM2, &c.WeightKG, &c.HeightCM, &c.CreatinineClearance, &c.ProviderID, &c.Note, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *chemoCycleRepoPG) Create(ctx context.Context, c *ChemoCycle) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO chemotherapy_cycle (id, protocol_id, cycle_number, planned_start_date, actual_start_date, actual_end_date,
			status, dose_reduction_pct, dose_reduction_reason, delay_days, delay_reason,
			bsa_m2, weight_kg, height_cm, creatinine_clearance, provider_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		c.ID, c.ProtocolID, c.CycleNumber, c.PlannedStartDate, c.ActualStartDate, c.ActualEndDate,
		c.Status, c.DoseReductionPct, c.DoseReductionReason, c.DelayDays, c.DelayReason,
		c.BSAM2, c.WeightKG, c.HeightCM, c.CreatinineClearance, c.ProviderID, c.Note)
	return err
}

func (r *chemoCycleRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ChemoCycle, error) {
	return r.scanCycle(r.conn(ctx).QueryRow(ctx, `SELECT `+chemoCycleCols+` FROM chemotherapy_cycle WHERE id = $1`, id))
}

func (r *chemoCycleRepoPG) Update(ctx context.Context, c *ChemoCycle) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE chemotherapy_cycle SET status=$2, actual_start_date=$3, actual_end_date=$4,
			dose_reduction_pct=$5, dose_reduction_reason=$6,
			delay_days=$7, delay_reason=$8, note=$9, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.Status, c.ActualStartDate, c.ActualEndDate,
		c.DoseReductionPct, c.DoseReductionReason,
		c.DelayDays, c.DelayReason, c.Note)
	return err
}

func (r *chemoCycleRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM chemotherapy_cycle WHERE id = $1`, id)
	return err
}

func (r *chemoCycleRepoPG) List(ctx context.Context, limit, offset int) ([]*ChemoCycle, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM chemotherapy_cycle`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+chemoCycleCols+` FROM chemotherapy_cycle ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ChemoCycle
	for rows.Next() {
		c, err := r.scanCycle(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

func (r *chemoCycleRepoPG) AddAdministration(ctx context.Context, a *ChemoAdministration) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO chemotherapy_administration (id, cycle_id, protocol_drug_id, drug_name,
			administration_datetime, dose_given, dose_unit, route,
			infusion_duration_min, infusion_rate, site, sequence_number,
			reaction_type, reaction_severity, reaction_action,
			administering_nurse_id, supervising_provider_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		a.ID, a.CycleID, a.ProtocolDrugID, a.DrugName,
		a.AdministrationDatetime, a.DoseGiven, a.DoseUnit, a.Route,
		a.InfusionDurationMin, a.InfusionRate, a.Site, a.SequenceNumber,
		a.ReactionType, a.ReactionSeverity, a.ReactionAction,
		a.AdministeringNurseID, a.SupervisingProviderID, a.Note)
	return err
}

func (r *chemoCycleRepoPG) GetAdministrations(ctx context.Context, cycleID uuid.UUID) ([]*ChemoAdministration, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, cycle_id, protocol_drug_id, drug_name,
			administration_datetime, dose_given, dose_unit, route,
			infusion_duration_min, infusion_rate, site, sequence_number,
			reaction_type, reaction_severity, reaction_action,
			administering_nurse_id, supervising_provider_id, note
		FROM chemotherapy_administration WHERE cycle_id = $1 ORDER BY administration_datetime`, cycleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ChemoAdministration
	for rows.Next() {
		var a ChemoAdministration
		if err := rows.Scan(&a.ID, &a.CycleID, &a.ProtocolDrugID, &a.DrugName,
			&a.AdministrationDatetime, &a.DoseGiven, &a.DoseUnit, &a.Route,
			&a.InfusionDurationMin, &a.InfusionRate, &a.Site, &a.SequenceNumber,
			&a.ReactionType, &a.ReactionSeverity, &a.ReactionAction,
			&a.AdministeringNurseID, &a.SupervisingProviderID, &a.Note); err != nil {
			return nil, err
		}
		items = append(items, &a)
	}
	return items, nil
}

// =========== Radiation Therapy Repository ===========

type radiationRepoPG struct{ pool *pgxpool.Pool }

func NewRadiationTherapyRepoPG(pool *pgxpool.Pool) RadiationTherapyRepository {
	return &radiationRepoPG{pool: pool}
}

func (r *radiationRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const radiationCols = `id, cancer_diagnosis_id, therapy_type, modality, technique,
	target_site, laterality, total_dose_cgy, dose_per_fraction_cgy,
	planned_fractions, completed_fractions, start_date, end_date, status,
	prescribing_provider_id, treating_facility_id,
	energy_type, energy_value, treatment_volume_cc, note, created_at, updated_at`

func (r *radiationRepoPG) scanRadiation(row pgx.Row) (*RadiationTherapy, error) {
	var rt RadiationTherapy
	err := row.Scan(&rt.ID, &rt.CancerDiagnosisID, &rt.TherapyType, &rt.Modality, &rt.Technique,
		&rt.TargetSite, &rt.Laterality, &rt.TotalDoseCGY, &rt.DosePerFractionCGY,
		&rt.PlannedFractions, &rt.CompletedFractions, &rt.StartDate, &rt.EndDate, &rt.Status,
		&rt.PrescribingProviderID, &rt.TreatingFacilityID,
		&rt.EnergyType, &rt.EnergyValue, &rt.TreatmentVolumeCC, &rt.Note, &rt.CreatedAt, &rt.UpdatedAt)
	return &rt, err
}

func (r *radiationRepoPG) Create(ctx context.Context, rt *RadiationTherapy) error {
	rt.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO radiation_therapy (id, cancer_diagnosis_id, therapy_type, modality, technique,
			target_site, laterality, total_dose_cgy, dose_per_fraction_cgy,
			planned_fractions, completed_fractions, start_date, end_date, status,
			prescribing_provider_id, treating_facility_id,
			energy_type, energy_value, treatment_volume_cc, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		rt.ID, rt.CancerDiagnosisID, rt.TherapyType, rt.Modality, rt.Technique,
		rt.TargetSite, rt.Laterality, rt.TotalDoseCGY, rt.DosePerFractionCGY,
		rt.PlannedFractions, rt.CompletedFractions, rt.StartDate, rt.EndDate, rt.Status,
		rt.PrescribingProviderID, rt.TreatingFacilityID,
		rt.EnergyType, rt.EnergyValue, rt.TreatmentVolumeCC, rt.Note)
	return err
}

func (r *radiationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*RadiationTherapy, error) {
	return r.scanRadiation(r.conn(ctx).QueryRow(ctx, `SELECT `+radiationCols+` FROM radiation_therapy WHERE id = $1`, id))
}

func (r *radiationRepoPG) Update(ctx context.Context, rt *RadiationTherapy) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE radiation_therapy SET status=$2, completed_fractions=$3, end_date=$4, note=$5, updated_at=NOW()
		WHERE id = $1`,
		rt.ID, rt.Status, rt.CompletedFractions, rt.EndDate, rt.Note)
	return err
}

func (r *radiationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM radiation_therapy WHERE id = $1`, id)
	return err
}

func (r *radiationRepoPG) List(ctx context.Context, limit, offset int) ([]*RadiationTherapy, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM radiation_therapy`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+radiationCols+` FROM radiation_therapy ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RadiationTherapy
	for rows.Next() {
		rt, err := r.scanRadiation(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rt)
	}
	return items, total, nil
}

func (r *radiationRepoPG) AddSession(ctx context.Context, s *RadiationSession) error {
	s.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO radiation_therapy_session (id, radiation_therapy_id, session_number, session_date,
			dose_delivered_cgy, field_name, setup_verified, imaging_type,
			skin_reaction_grade, fatigue_grade, other_toxicity, toxicity_grade,
			machine_id, therapist_id, physicist_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		s.ID, s.RadiationTherapyID, s.SessionNumber, s.SessionDate,
		s.DoseDeliveredCGY, s.FieldName, s.SetupVerified, s.ImagingType,
		s.SkinReactionGrade, s.FatigueGrade, s.OtherToxicity, s.ToxicityGrade,
		s.MachineID, s.TherapistID, s.PhysicistID, s.Note)
	return err
}

func (r *radiationRepoPG) GetSessions(ctx context.Context, radiationID uuid.UUID) ([]*RadiationSession, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, radiation_therapy_id, session_number, session_date,
			dose_delivered_cgy, field_name, setup_verified, imaging_type,
			skin_reaction_grade, fatigue_grade, other_toxicity, toxicity_grade,
			machine_id, therapist_id, physicist_id, note
		FROM radiation_therapy_session WHERE radiation_therapy_id = $1 ORDER BY session_number`, radiationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*RadiationSession
	for rows.Next() {
		var s RadiationSession
		if err := rows.Scan(&s.ID, &s.RadiationTherapyID, &s.SessionNumber, &s.SessionDate,
			&s.DoseDeliveredCGY, &s.FieldName, &s.SetupVerified, &s.ImagingType,
			&s.SkinReactionGrade, &s.FatigueGrade, &s.OtherToxicity, &s.ToxicityGrade,
			&s.MachineID, &s.TherapistID, &s.PhysicistID, &s.Note); err != nil {
			return nil, err
		}
		items = append(items, &s)
	}
	return items, nil
}

// =========== Tumor Marker Repository ===========

type tumorMarkerRepoPG struct{ pool *pgxpool.Pool }

func NewTumorMarkerRepoPG(pool *pgxpool.Pool) TumorMarkerRepository {
	return &tumorMarkerRepoPG{pool: pool}
}

func (r *tumorMarkerRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const markerCols = `id, cancer_diagnosis_id, patient_id, marker_name, marker_code, marker_code_system,
	value_quantity, value_unit, value_string, value_interpretation,
	reference_range_low, reference_range_high, reference_range_text,
	specimen_type, collection_datetime, result_datetime,
	performing_lab, ordering_provider_id, note, created_at, updated_at`

func (r *tumorMarkerRepoPG) scanMarker(row pgx.Row) (*TumorMarker, error) {
	var m TumorMarker
	err := row.Scan(&m.ID, &m.CancerDiagnosisID, &m.PatientID, &m.MarkerName, &m.MarkerCode, &m.MarkerCodeSystem,
		&m.ValueQuantity, &m.ValueUnit, &m.ValueString, &m.ValueInterpretation,
		&m.ReferenceRangeLow, &m.ReferenceRangeHigh, &m.ReferenceRangeText,
		&m.SpecimenType, &m.CollectionDatetime, &m.ResultDatetime,
		&m.PerformingLab, &m.OrderingProviderID, &m.Note, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *tumorMarkerRepoPG) Create(ctx context.Context, m *TumorMarker) error {
	m.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO tumor_marker (id, cancer_diagnosis_id, patient_id, marker_name, marker_code, marker_code_system,
			value_quantity, value_unit, value_string, value_interpretation,
			reference_range_low, reference_range_high, reference_range_text,
			specimen_type, collection_datetime, result_datetime,
			performing_lab, ordering_provider_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		m.ID, m.CancerDiagnosisID, m.PatientID, m.MarkerName, m.MarkerCode, m.MarkerCodeSystem,
		m.ValueQuantity, m.ValueUnit, m.ValueString, m.ValueInterpretation,
		m.ReferenceRangeLow, m.ReferenceRangeHigh, m.ReferenceRangeText,
		m.SpecimenType, m.CollectionDatetime, m.ResultDatetime,
		m.PerformingLab, m.OrderingProviderID, m.Note)
	return err
}

func (r *tumorMarkerRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TumorMarker, error) {
	return r.scanMarker(r.conn(ctx).QueryRow(ctx, `SELECT `+markerCols+` FROM tumor_marker WHERE id = $1`, id))
}

func (r *tumorMarkerRepoPG) Update(ctx context.Context, m *TumorMarker) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE tumor_marker SET value_quantity=$2, value_unit=$3, value_string=$4,
			value_interpretation=$5, result_datetime=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.ValueQuantity, m.ValueUnit, m.ValueString,
		m.ValueInterpretation, m.ResultDatetime, m.Note)
	return err
}

func (r *tumorMarkerRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM tumor_marker WHERE id = $1`, id)
	return err
}

func (r *tumorMarkerRepoPG) List(ctx context.Context, limit, offset int) ([]*TumorMarker, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM tumor_marker`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+markerCols+` FROM tumor_marker ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TumorMarker
	for rows.Next() {
		m, err := r.scanMarker(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

// =========== Tumor Board Repository ===========

type tumorBoardRepoPG struct{ pool *pgxpool.Pool }

func NewTumorBoardRepoPG(pool *pgxpool.Pool) TumorBoardRepository {
	return &tumorBoardRepoPG{pool: pool}
}

func (r *tumorBoardRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const boardCols = `id, cancer_diagnosis_id, patient_id, review_date, review_type,
	presenting_provider_id, attendees, clinical_summary, pathology_summary,
	imaging_summary, discussion, recommendations, treatment_decision,
	clinical_trial_discussed, clinical_trial_id, next_review_date, note,
	created_at, updated_at`

func (r *tumorBoardRepoPG) scanBoard(row pgx.Row) (*TumorBoardReview, error) {
	var b TumorBoardReview
	err := row.Scan(&b.ID, &b.CancerDiagnosisID, &b.PatientID, &b.ReviewDate, &b.ReviewType,
		&b.PresentingProviderID, &b.Attendees, &b.ClinicalSummary, &b.PathologySummary,
		&b.ImagingSummary, &b.Discussion, &b.Recommendations, &b.TreatmentDecision,
		&b.ClinicalTrialDiscussed, &b.ClinicalTrialID, &b.NextReviewDate, &b.Note,
		&b.CreatedAt, &b.UpdatedAt)
	return &b, err
}

func (r *tumorBoardRepoPG) Create(ctx context.Context, b *TumorBoardReview) error {
	b.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO tumor_board_review (id, cancer_diagnosis_id, patient_id, review_date, review_type,
			presenting_provider_id, attendees, clinical_summary, pathology_summary,
			imaging_summary, discussion, recommendations, treatment_decision,
			clinical_trial_discussed, clinical_trial_id, next_review_date, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		b.ID, b.CancerDiagnosisID, b.PatientID, b.ReviewDate, b.ReviewType,
		b.PresentingProviderID, b.Attendees, b.ClinicalSummary, b.PathologySummary,
		b.ImagingSummary, b.Discussion, b.Recommendations, b.TreatmentDecision,
		b.ClinicalTrialDiscussed, b.ClinicalTrialID, b.NextReviewDate, b.Note)
	return err
}

func (r *tumorBoardRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*TumorBoardReview, error) {
	return r.scanBoard(r.conn(ctx).QueryRow(ctx, `SELECT `+boardCols+` FROM tumor_board_review WHERE id = $1`, id))
}

func (r *tumorBoardRepoPG) Update(ctx context.Context, b *TumorBoardReview) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE tumor_board_review SET recommendations=$2, treatment_decision=$3,
			clinical_trial_discussed=$4, clinical_trial_id=$5,
			next_review_date=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		b.ID, b.Recommendations, b.TreatmentDecision,
		b.ClinicalTrialDiscussed, b.ClinicalTrialID,
		b.NextReviewDate, b.Note)
	return err
}

func (r *tumorBoardRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM tumor_board_review WHERE id = $1`, id)
	return err
}

func (r *tumorBoardRepoPG) List(ctx context.Context, limit, offset int) ([]*TumorBoardReview, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM tumor_board_review`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+boardCols+` FROM tumor_board_review ORDER BY review_date DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*TumorBoardReview
	for rows.Next() {
		b, err := r.scanBoard(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, nil
}
