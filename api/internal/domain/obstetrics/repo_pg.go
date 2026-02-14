package obstetrics

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

// =========== Pregnancy Repository ===========

type pregnancyRepoPG struct{ pool *pgxpool.Pool }

func NewPregnancyRepoPG(pool *pgxpool.Pool) PregnancyRepository {
	return &pregnancyRepoPG{pool: pool}
}

func (r *pregnancyRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const pregCols = `id, patient_id, status, onset_date, estimated_due_date, last_menstrual_period,
	conception_method, gravida, para, multiple_gestation, number_of_fetuses,
	risk_level, risk_factors, blood_type, rh_factor,
	pre_pregnancy_weight, pre_pregnancy_bmi,
	primary_provider_id, managing_organization_id, note,
	outcome_date, outcome_summary, created_at, updated_at`

func (r *pregnancyRepoPG) scanPregnancy(row pgx.Row) (*Pregnancy, error) {
	var p Pregnancy
	err := row.Scan(&p.ID, &p.PatientID, &p.Status, &p.OnsetDate, &p.EstimatedDueDate, &p.LastMenstrualPeriod,
		&p.ConceptionMethod, &p.Gravida, &p.Para, &p.MultipleGestation, &p.NumberOfFetuses,
		&p.RiskLevel, &p.RiskFactors, &p.BloodType, &p.RhFactor,
		&p.PrePregnancyWeight, &p.PrePregnancyBMI,
		&p.PrimaryProviderID, &p.ManagingOrganizationID, &p.Note,
		&p.OutcomeDate, &p.OutcomeSummary, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *pregnancyRepoPG) Create(ctx context.Context, p *Pregnancy) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO pregnancy (id, patient_id, status, onset_date, estimated_due_date, last_menstrual_period,
			conception_method, gravida, para, multiple_gestation, number_of_fetuses,
			risk_level, risk_factors, blood_type, rh_factor,
			pre_pregnancy_weight, pre_pregnancy_bmi,
			primary_provider_id, managing_organization_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		p.ID, p.PatientID, p.Status, p.OnsetDate, p.EstimatedDueDate, p.LastMenstrualPeriod,
		p.ConceptionMethod, p.Gravida, p.Para, p.MultipleGestation, p.NumberOfFetuses,
		p.RiskLevel, p.RiskFactors, p.BloodType, p.RhFactor,
		p.PrePregnancyWeight, p.PrePregnancyBMI,
		p.PrimaryProviderID, p.ManagingOrganizationID, p.Note)
	return err
}

func (r *pregnancyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Pregnancy, error) {
	return r.scanPregnancy(r.conn(ctx).QueryRow(ctx, `SELECT `+pregCols+` FROM pregnancy WHERE id = $1`, id))
}

func (r *pregnancyRepoPG) Update(ctx context.Context, p *Pregnancy) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE pregnancy SET status=$2, estimated_due_date=$3, risk_level=$4, risk_factors=$5,
			note=$6, outcome_date=$7, outcome_summary=$8, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Status, p.EstimatedDueDate, p.RiskLevel, p.RiskFactors,
		p.Note, p.OutcomeDate, p.OutcomeSummary)
	return err
}

func (r *pregnancyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM pregnancy WHERE id = $1`, id)
	return err
}

func (r *pregnancyRepoPG) List(ctx context.Context, limit, offset int) ([]*Pregnancy, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM pregnancy`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+pregCols+` FROM pregnancy ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Pregnancy
	for rows.Next() {
		p, err := r.scanPregnancy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

func (r *pregnancyRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Pregnancy, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM pregnancy WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+pregCols+` FROM pregnancy WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Pregnancy
	for rows.Next() {
		p, err := r.scanPregnancy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

// =========== Prenatal Visit Repository ===========

type prenatalVisitRepoPG struct{ pool *pgxpool.Pool }

func NewPrenatalVisitRepoPG(pool *pgxpool.Pool) PrenatalVisitRepository {
	return &prenatalVisitRepoPG{pool: pool}
}

func (r *prenatalVisitRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const prenatalCols = `id, pregnancy_id, encounter_id, visit_date,
	gestational_age_weeks, gestational_age_days, weight,
	blood_pressure_systolic, blood_pressure_diastolic,
	fundal_height, fetal_heart_rate, fetal_presentation, fetal_movement,
	urine_protein, urine_glucose, edema,
	cervical_dilation, cervical_effacement, group_b_strep_status,
	provider_id, note, next_visit_date, created_at, updated_at`

func (r *prenatalVisitRepoPG) scanVisit(row pgx.Row) (*PrenatalVisit, error) {
	var v PrenatalVisit
	err := row.Scan(&v.ID, &v.PregnancyID, &v.EncounterID, &v.VisitDate,
		&v.GestationalAgeWeeks, &v.GestationalAgeDays, &v.Weight,
		&v.BloodPressureSystolic, &v.BloodPressureDiastolic,
		&v.FundalHeight, &v.FetalHeartRate, &v.FetalPresentation, &v.FetalMovement,
		&v.UrineProtein, &v.UrineGlucose, &v.Edema,
		&v.CervicalDilation, &v.CervicalEffacement, &v.GroupBStrepStatus,
		&v.ProviderID, &v.Note, &v.NextVisitDate, &v.CreatedAt, &v.UpdatedAt)
	return &v, err
}

func (r *prenatalVisitRepoPG) Create(ctx context.Context, v *PrenatalVisit) error {
	v.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO prenatal_visit (id, pregnancy_id, encounter_id, visit_date,
			gestational_age_weeks, gestational_age_days, weight,
			blood_pressure_systolic, blood_pressure_diastolic,
			fundal_height, fetal_heart_rate, fetal_presentation, fetal_movement,
			urine_protein, urine_glucose, edema,
			cervical_dilation, cervical_effacement, group_b_strep_status,
			provider_id, note, next_visit_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		v.ID, v.PregnancyID, v.EncounterID, v.VisitDate,
		v.GestationalAgeWeeks, v.GestationalAgeDays, v.Weight,
		v.BloodPressureSystolic, v.BloodPressureDiastolic,
		v.FundalHeight, v.FetalHeartRate, v.FetalPresentation, v.FetalMovement,
		v.UrineProtein, v.UrineGlucose, v.Edema,
		v.CervicalDilation, v.CervicalEffacement, v.GroupBStrepStatus,
		v.ProviderID, v.Note, v.NextVisitDate)
	return err
}

func (r *prenatalVisitRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PrenatalVisit, error) {
	return r.scanVisit(r.conn(ctx).QueryRow(ctx, `SELECT `+prenatalCols+` FROM prenatal_visit WHERE id = $1`, id))
}

func (r *prenatalVisitRepoPG) Update(ctx context.Context, v *PrenatalVisit) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE prenatal_visit SET weight=$2, blood_pressure_systolic=$3, blood_pressure_diastolic=$4,
			fundal_height=$5, fetal_heart_rate=$6, fetal_presentation=$7,
			urine_protein=$8, urine_glucose=$9, edema=$10,
			note=$11, next_visit_date=$12, updated_at=NOW()
		WHERE id = $1`,
		v.ID, v.Weight, v.BloodPressureSystolic, v.BloodPressureDiastolic,
		v.FundalHeight, v.FetalHeartRate, v.FetalPresentation,
		v.UrineProtein, v.UrineGlucose, v.Edema,
		v.Note, v.NextVisitDate)
	return err
}

func (r *prenatalVisitRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM prenatal_visit WHERE id = $1`, id)
	return err
}

func (r *prenatalVisitRepoPG) ListByPregnancy(ctx context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*PrenatalVisit, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM prenatal_visit WHERE pregnancy_id = $1`, pregnancyID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+prenatalCols+` FROM prenatal_visit WHERE pregnancy_id = $1 ORDER BY visit_date DESC LIMIT $2 OFFSET $3`, pregnancyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PrenatalVisit
	for rows.Next() {
		v, err := r.scanVisit(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}

// =========== Labor Repository ===========

type laborRepoPG struct{ pool *pgxpool.Pool }

func NewLaborRepoPG(pool *pgxpool.Pool) LaborRepository { return &laborRepoPG{pool: pool} }

func (r *laborRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const laborCols = `id, pregnancy_id, encounter_id, admission_datetime, labor_onset_datetime, labor_onset_type,
	membrane_rupture_datetime, membrane_rupture_type, amniotic_fluid_color, amniotic_fluid_volume,
	induction_method, induction_reason, augmentation_method,
	anesthesia_type, anesthesia_start, status, attending_provider_id, note, created_at, updated_at`

func (r *laborRepoPG) scanLabor(row pgx.Row) (*LaborRecord, error) {
	var l LaborRecord
	err := row.Scan(&l.ID, &l.PregnancyID, &l.EncounterID, &l.AdmissionDatetime, &l.LaborOnsetDatetime, &l.LaborOnsetType,
		&l.MembraneRuptureDatetime, &l.MembraneRuptureType, &l.AmnioticFluidColor, &l.AmnioticFluidVolume,
		&l.InductionMethod, &l.InductionReason, &l.AugmentationMethod,
		&l.AnesthesiaType, &l.AnesthesiaStart, &l.Status, &l.AttendingProviderID, &l.Note, &l.CreatedAt, &l.UpdatedAt)
	return &l, err
}

func (r *laborRepoPG) Create(ctx context.Context, l *LaborRecord) error {
	l.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO labor_record (id, pregnancy_id, encounter_id, admission_datetime, labor_onset_datetime, labor_onset_type,
			membrane_rupture_datetime, membrane_rupture_type, amniotic_fluid_color, amniotic_fluid_volume,
			induction_method, induction_reason, augmentation_method,
			anesthesia_type, anesthesia_start, status, attending_provider_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		l.ID, l.PregnancyID, l.EncounterID, l.AdmissionDatetime, l.LaborOnsetDatetime, l.LaborOnsetType,
		l.MembraneRuptureDatetime, l.MembraneRuptureType, l.AmnioticFluidColor, l.AmnioticFluidVolume,
		l.InductionMethod, l.InductionReason, l.AugmentationMethod,
		l.AnesthesiaType, l.AnesthesiaStart, l.Status, l.AttendingProviderID, l.Note)
	return err
}

func (r *laborRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*LaborRecord, error) {
	return r.scanLabor(r.conn(ctx).QueryRow(ctx, `SELECT `+laborCols+` FROM labor_record WHERE id = $1`, id))
}

func (r *laborRepoPG) Update(ctx context.Context, l *LaborRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE labor_record SET status=$2, anesthesia_type=$3, anesthesia_start=$4, note=$5, updated_at=NOW()
		WHERE id = $1`,
		l.ID, l.Status, l.AnesthesiaType, l.AnesthesiaStart, l.Note)
	return err
}

func (r *laborRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM labor_record WHERE id = $1`, id)
	return err
}

func (r *laborRepoPG) List(ctx context.Context, limit, offset int) ([]*LaborRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM labor_record`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+laborCols+` FROM labor_record ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*LaborRecord
	for rows.Next() {
		l, err := r.scanLabor(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

func (r *laborRepoPG) AddCervicalExam(ctx context.Context, e *LaborCervicalExam) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO labor_cervical_exam (id, labor_record_id, exam_datetime,
			dilation_cm, effacement_pct, station, fetal_position, membrane_status,
			examiner_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.ID, e.LaborRecordID, e.ExamDatetime,
		e.DilationCM, e.EffacementPct, e.Station, e.FetalPosition, e.MembraneStatus,
		e.ExaminerID, e.Note)
	return err
}

func (r *laborRepoPG) GetCervicalExams(ctx context.Context, laborRecordID uuid.UUID) ([]*LaborCervicalExam, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, labor_record_id, exam_datetime,
			dilation_cm, effacement_pct, station, fetal_position, membrane_status,
			examiner_id, note
		FROM labor_cervical_exam WHERE labor_record_id = $1 ORDER BY exam_datetime DESC`, laborRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*LaborCervicalExam
	for rows.Next() {
		var e LaborCervicalExam
		if err := rows.Scan(&e.ID, &e.LaborRecordID, &e.ExamDatetime,
			&e.DilationCM, &e.EffacementPct, &e.Station, &e.FetalPosition, &e.MembraneStatus,
			&e.ExaminerID, &e.Note); err != nil {
			return nil, err
		}
		items = append(items, &e)
	}
	return items, nil
}

func (r *laborRepoPG) AddFetalMonitoring(ctx context.Context, f *FetalMonitoring) error {
	f.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO fetal_monitoring (id, labor_record_id, monitoring_datetime, monitoring_type,
			fetal_heart_rate, baseline_rate, variability, accelerations, decelerations, deceleration_type,
			contraction_frequency, contraction_duration, contraction_intensity, uterine_resting_tone,
			mvus, interpretation, category, recorder_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		f.ID, f.LaborRecordID, f.MonitoringDatetime, f.MonitoringType,
		f.FetalHeartRate, f.BaselineRate, f.Variability, f.Accelerations, f.Decelerations, f.DecelerationType,
		f.ContractionFrequency, f.ContractionDuration, f.ContractionIntensity, f.UterineRestingTone,
		f.MVUs, f.Interpretation, f.Category, f.RecorderID, f.Note)
	return err
}

func (r *laborRepoPG) GetFetalMonitoring(ctx context.Context, laborRecordID uuid.UUID) ([]*FetalMonitoring, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, labor_record_id, monitoring_datetime, monitoring_type,
			fetal_heart_rate, baseline_rate, variability, accelerations, decelerations, deceleration_type,
			contraction_frequency, contraction_duration, contraction_intensity, uterine_resting_tone,
			mvus, interpretation, category, recorder_id, note
		FROM fetal_monitoring WHERE labor_record_id = $1 ORDER BY monitoring_datetime DESC`, laborRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*FetalMonitoring
	for rows.Next() {
		var f FetalMonitoring
		if err := rows.Scan(&f.ID, &f.LaborRecordID, &f.MonitoringDatetime, &f.MonitoringType,
			&f.FetalHeartRate, &f.BaselineRate, &f.Variability, &f.Accelerations, &f.Decelerations, &f.DecelerationType,
			&f.ContractionFrequency, &f.ContractionDuration, &f.ContractionIntensity, &f.UterineRestingTone,
			&f.MVUs, &f.Interpretation, &f.Category, &f.RecorderID, &f.Note); err != nil {
			return nil, err
		}
		items = append(items, &f)
	}
	return items, nil
}

// =========== Delivery Repository ===========

type deliveryRepoPG struct{ pool *pgxpool.Pool }

func NewDeliveryRepoPG(pool *pgxpool.Pool) DeliveryRepository {
	return &deliveryRepoPG{pool: pool}
}

func (r *deliveryRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const deliveryCols = `id, pregnancy_id, labor_record_id, patient_id,
	delivery_datetime, delivery_method, delivery_type,
	delivering_provider_id, assistant_provider_id, delivery_location_id,
	birth_order, placenta_delivery, placenta_datetime, placenta_intact,
	cord_vessels, cord_blood_collected, episiotomy, episiotomy_type,
	laceration_degree, repair_method, blood_loss_ml, complications, note,
	created_at, updated_at`

func (r *deliveryRepoPG) scanDelivery(row pgx.Row) (*DeliveryRecord, error) {
	var d DeliveryRecord
	err := row.Scan(&d.ID, &d.PregnancyID, &d.LaborRecordID, &d.PatientID,
		&d.DeliveryDatetime, &d.DeliveryMethod, &d.DeliveryType,
		&d.DeliveringProviderID, &d.AssistantProviderID, &d.DeliveryLocationID,
		&d.BirthOrder, &d.PlacentaDelivery, &d.PlacentaDatetime, &d.PlacentaIntact,
		&d.CordVessels, &d.CordBloodCollected, &d.Episiotomy, &d.EpisiotomyType,
		&d.LacerationDegree, &d.RepairMethod, &d.BloodLossML, &d.Complications, &d.Note,
		&d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *deliveryRepoPG) Create(ctx context.Context, d *DeliveryRecord) error {
	d.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO delivery_record (id, pregnancy_id, labor_record_id, patient_id,
			delivery_datetime, delivery_method, delivery_type,
			delivering_provider_id, assistant_provider_id, delivery_location_id,
			birth_order, placenta_delivery, placenta_datetime, placenta_intact,
			cord_vessels, cord_blood_collected, episiotomy, episiotomy_type,
			laceration_degree, repair_method, blood_loss_ml, complications, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		d.ID, d.PregnancyID, d.LaborRecordID, d.PatientID,
		d.DeliveryDatetime, d.DeliveryMethod, d.DeliveryType,
		d.DeliveringProviderID, d.AssistantProviderID, d.DeliveryLocationID,
		d.BirthOrder, d.PlacentaDelivery, d.PlacentaDatetime, d.PlacentaIntact,
		d.CordVessels, d.CordBloodCollected, d.Episiotomy, d.EpisiotomyType,
		d.LacerationDegree, d.RepairMethod, d.BloodLossML, d.Complications, d.Note)
	return err
}

func (r *deliveryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DeliveryRecord, error) {
	return r.scanDelivery(r.conn(ctx).QueryRow(ctx, `SELECT `+deliveryCols+` FROM delivery_record WHERE id = $1`, id))
}

func (r *deliveryRepoPG) Update(ctx context.Context, d *DeliveryRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE delivery_record SET delivery_method=$2, delivery_type=$3,
			placenta_delivery=$4, placenta_intact=$5,
			blood_loss_ml=$6, complications=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.DeliveryMethod, d.DeliveryType,
		d.PlacentaDelivery, d.PlacentaIntact,
		d.BloodLossML, d.Complications, d.Note)
	return err
}

func (r *deliveryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM delivery_record WHERE id = $1`, id)
	return err
}

func (r *deliveryRepoPG) ListByPregnancy(ctx context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*DeliveryRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM delivery_record WHERE pregnancy_id = $1`, pregnancyID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+deliveryCols+` FROM delivery_record WHERE pregnancy_id = $1 ORDER BY delivery_datetime DESC LIMIT $2 OFFSET $3`, pregnancyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DeliveryRecord
	for rows.Next() {
		d, err := r.scanDelivery(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

// =========== Newborn Repository ===========

type newbornRepoPG struct{ pool *pgxpool.Pool }

func NewNewbornRepoPG(pool *pgxpool.Pool) NewbornRepository { return &newbornRepoPG{pool: pool} }

func (r *newbornRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const newbornCols = `id, delivery_id, patient_id, birth_datetime, sex,
	birth_weight_grams, birth_length_cm, head_circumference_cm,
	apgar_1min, apgar_5min, apgar_10min, resuscitation_type,
	gestational_age_weeks, gestational_age_days, birth_status,
	nicu_admission, nicu_reason, vitamin_k_given, eye_prophylaxis_given,
	hepatitis_b_given, newborn_screening, feeding_method, note,
	created_at, updated_at`

func (r *newbornRepoPG) scanNewborn(row pgx.Row) (*NewbornRecord, error) {
	var n NewbornRecord
	err := row.Scan(&n.ID, &n.DeliveryID, &n.PatientID, &n.BirthDatetime, &n.Sex,
		&n.BirthWeightGrams, &n.BirthLengthCM, &n.HeadCircumferenceCM,
		&n.Apgar1Min, &n.Apgar5Min, &n.Apgar10Min, &n.ResuscitationType,
		&n.GestationalAgeWeeks, &n.GestationalAgeDays, &n.BirthStatus,
		&n.NICUAdmission, &n.NICUReason, &n.VitaminKGiven, &n.EyeProphylaxisGiven,
		&n.HepatitisBGiven, &n.NewbornScreening, &n.FeedingMethod, &n.Note,
		&n.CreatedAt, &n.UpdatedAt)
	return &n, err
}

func (r *newbornRepoPG) Create(ctx context.Context, n *NewbornRecord) error {
	n.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO newborn_record (id, delivery_id, patient_id, birth_datetime, sex,
			birth_weight_grams, birth_length_cm, head_circumference_cm,
			apgar_1min, apgar_5min, apgar_10min, resuscitation_type,
			gestational_age_weeks, gestational_age_days, birth_status,
			nicu_admission, nicu_reason, vitamin_k_given, eye_prophylaxis_given,
			hepatitis_b_given, newborn_screening, feeding_method, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		n.ID, n.DeliveryID, n.PatientID, n.BirthDatetime, n.Sex,
		n.BirthWeightGrams, n.BirthLengthCM, n.HeadCircumferenceCM,
		n.Apgar1Min, n.Apgar5Min, n.Apgar10Min, n.ResuscitationType,
		n.GestationalAgeWeeks, n.GestationalAgeDays, n.BirthStatus,
		n.NICUAdmission, n.NICUReason, n.VitaminKGiven, n.EyeProphylaxisGiven,
		n.HepatitisBGiven, n.NewbornScreening, n.FeedingMethod, n.Note)
	return err
}

func (r *newbornRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*NewbornRecord, error) {
	return r.scanNewborn(r.conn(ctx).QueryRow(ctx, `SELECT `+newbornCols+` FROM newborn_record WHERE id = $1`, id))
}

func (r *newbornRepoPG) Update(ctx context.Context, n *NewbornRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE newborn_record SET sex=$2, birth_weight_grams=$3, birth_length_cm=$4,
			apgar_1min=$5, apgar_5min=$6, apgar_10min=$7,
			birth_status=$8, nicu_admission=$9, nicu_reason=$10,
			feeding_method=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		n.ID, n.Sex, n.BirthWeightGrams, n.BirthLengthCM,
		n.Apgar1Min, n.Apgar5Min, n.Apgar10Min,
		n.BirthStatus, n.NICUAdmission, n.NICUReason,
		n.FeedingMethod, n.Note)
	return err
}

func (r *newbornRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM newborn_record WHERE id = $1`, id)
	return err
}

func (r *newbornRepoPG) List(ctx context.Context, limit, offset int) ([]*NewbornRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM newborn_record`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+newbornCols+` FROM newborn_record ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NewbornRecord
	for rows.Next() {
		n, err := r.scanNewborn(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

// =========== Postpartum Repository ===========

type postpartumRepoPG struct{ pool *pgxpool.Pool }

func NewPostpartumRepoPG(pool *pgxpool.Pool) PostpartumRepository {
	return &postpartumRepoPG{pool: pool}
}

func (r *postpartumRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const postpartumCols = `id, pregnancy_id, patient_id, encounter_id, visit_date,
	days_postpartum, weeks_postpartum, uterine_involution,
	lochia_type, lochia_amount, perineum_status, incision_status,
	breast_status, breastfeeding_status, contraception_plan,
	mood_screening_score, mood_screening_tool, depression_risk,
	blood_pressure_systolic, blood_pressure_diastolic, weight,
	provider_id, note, created_at, updated_at`

func (r *postpartumRepoPG) scanPostpartum(row pgx.Row) (*PostpartumRecord, error) {
	var p PostpartumRecord
	err := row.Scan(&p.ID, &p.PregnancyID, &p.PatientID, &p.EncounterID, &p.VisitDate,
		&p.DaysPostpartum, &p.WeeksPostpartum, &p.UterineInvolution,
		&p.LochiaType, &p.LochiaAmount, &p.PerineumStatus, &p.IncisionStatus,
		&p.BreastStatus, &p.BreastfeedingStatus, &p.ContraceptionPlan,
		&p.MoodScreeningScore, &p.MoodScreeningTool, &p.DepressionRisk,
		&p.BloodPressureSystolic, &p.BloodPressureDiastolic, &p.Weight,
		&p.ProviderID, &p.Note, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *postpartumRepoPG) Create(ctx context.Context, p *PostpartumRecord) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO postpartum_record (id, pregnancy_id, patient_id, encounter_id, visit_date,
			days_postpartum, weeks_postpartum, uterine_involution,
			lochia_type, lochia_amount, perineum_status, incision_status,
			breast_status, breastfeeding_status, contraception_plan,
			mood_screening_score, mood_screening_tool, depression_risk,
			blood_pressure_systolic, blood_pressure_diastolic, weight,
			provider_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		p.ID, p.PregnancyID, p.PatientID, p.EncounterID, p.VisitDate,
		p.DaysPostpartum, p.WeeksPostpartum, p.UterineInvolution,
		p.LochiaType, p.LochiaAmount, p.PerineumStatus, p.IncisionStatus,
		p.BreastStatus, p.BreastfeedingStatus, p.ContraceptionPlan,
		p.MoodScreeningScore, p.MoodScreeningTool, p.DepressionRisk,
		p.BloodPressureSystolic, p.BloodPressureDiastolic, p.Weight,
		p.ProviderID, p.Note)
	return err
}

func (r *postpartumRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PostpartumRecord, error) {
	return r.scanPostpartum(r.conn(ctx).QueryRow(ctx, `SELECT `+postpartumCols+` FROM postpartum_record WHERE id = $1`, id))
}

func (r *postpartumRepoPG) Update(ctx context.Context, p *PostpartumRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE postpartum_record SET uterine_involution=$2, lochia_type=$3, lochia_amount=$4,
			perineum_status=$5, incision_status=$6, breast_status=$7, breastfeeding_status=$8,
			contraception_plan=$9, mood_screening_score=$10, mood_screening_tool=$11,
			depression_risk=$12, blood_pressure_systolic=$13, blood_pressure_diastolic=$14,
			weight=$15, note=$16, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.UterineInvolution, p.LochiaType, p.LochiaAmount,
		p.PerineumStatus, p.IncisionStatus, p.BreastStatus, p.BreastfeedingStatus,
		p.ContraceptionPlan, p.MoodScreeningScore, p.MoodScreeningTool,
		p.DepressionRisk, p.BloodPressureSystolic, p.BloodPressureDiastolic,
		p.Weight, p.Note)
	return err
}

func (r *postpartumRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM postpartum_record WHERE id = $1`, id)
	return err
}

func (r *postpartumRepoPG) List(ctx context.Context, limit, offset int) ([]*PostpartumRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM postpartum_record`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+postpartumCols+` FROM postpartum_record ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PostpartumRecord
	for rows.Next() {
		p, err := r.scanPostpartum(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}
