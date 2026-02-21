package medication

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
)

type queryable interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// =========== Medication Repository ===========

type medicationRepoPG struct{ pool *pgxpool.Pool }

func NewMedicationRepoPG(pool *pgxpool.Pool) MedicationRepository {
	return &medicationRepoPG{pool: pool}
}

func (r *medicationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const medCols = `id, fhir_id, code_system, code_value, code_display, status,
	form_code, form_display, amount_numerator, amount_numerator_unit,
	amount_denominator, amount_denominator_unit, schedule,
	is_brand, is_over_the_counter, manufacturer_id, manufacturer_name,
	lot_number, expiration_date, ndc_code, gtin_code,
	dpco_scheduled, cdsco_approval,
	is_narcotic, is_antibiotic, is_high_alert, requires_reconstitution,
	description, note, created_at, updated_at`

func (r *medicationRepoPG) scanMed(row pgx.Row) (*Medication, error) {
	var m Medication
	err := row.Scan(&m.ID, &m.FHIRID, &m.CodeSystem, &m.CodeValue, &m.CodeDisplay, &m.Status,
		&m.FormCode, &m.FormDisplay, &m.AmountNumerator, &m.AmountNumeratorUnit,
		&m.AmountDenominator, &m.AmountDenominatorUnit, &m.Schedule,
		&m.IsBrand, &m.IsOverTheCounter, &m.ManufacturerID, &m.ManufacturerName,
		&m.LotNumber, &m.ExpirationDate, &m.NDCCode, &m.GTINCode,
		&m.DPCOScheduled, &m.CDSCOApproval,
		&m.IsNarcotic, &m.IsAntibiotic, &m.IsHighAlert, &m.RequiresReconstitution,
		&m.Description, &m.Note, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *medicationRepoPG) Create(ctx context.Context, m *Medication) error {
	m.ID = uuid.New()
	if m.FHIRID == "" {
		m.FHIRID = m.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication (id, fhir_id, code_system, code_value, code_display, status,
			form_code, form_display, amount_numerator, amount_numerator_unit,
			amount_denominator, amount_denominator_unit, schedule,
			is_brand, is_over_the_counter, manufacturer_id, manufacturer_name,
			lot_number, expiration_date, ndc_code, gtin_code,
			is_narcotic, is_antibiotic, is_high_alert, requires_reconstitution,
			description, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)`,
		m.ID, m.FHIRID, m.CodeSystem, m.CodeValue, m.CodeDisplay, m.Status,
		m.FormCode, m.FormDisplay, m.AmountNumerator, m.AmountNumeratorUnit,
		m.AmountDenominator, m.AmountDenominatorUnit, m.Schedule,
		m.IsBrand, m.IsOverTheCounter, m.ManufacturerID, m.ManufacturerName,
		m.LotNumber, m.ExpirationDate, m.NDCCode, m.GTINCode,
		m.IsNarcotic, m.IsAntibiotic, m.IsHighAlert, m.RequiresReconstitution,
		m.Description, m.Note)
	return err
}

func (r *medicationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Medication, error) {
	return r.scanMed(r.conn(ctx).QueryRow(ctx, `SELECT `+medCols+` FROM medication WHERE id = $1`, id))
}

func (r *medicationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Medication, error) {
	return r.scanMed(r.conn(ctx).QueryRow(ctx, `SELECT `+medCols+` FROM medication WHERE fhir_id = $1`, fhirID))
}

func (r *medicationRepoPG) Update(ctx context.Context, m *Medication) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication SET code_system=$2, code_value=$3, code_display=$4, status=$5,
			form_code=$6, form_display=$7, schedule=$8,
			is_brand=$9, is_over_the_counter=$10,
			is_narcotic=$11, is_antibiotic=$12, is_high_alert=$13,
			description=$14, note=$15, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.CodeSystem, m.CodeValue, m.CodeDisplay, m.Status,
		m.FormCode, m.FormDisplay, m.Schedule,
		m.IsBrand, m.IsOverTheCounter,
		m.IsNarcotic, m.IsAntibiotic, m.IsHighAlert,
		m.Description, m.Note)
	return err
}

func (r *medicationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication WHERE id = $1`, id)
	return err
}

var medicationSearchParams = map[string]fhir.SearchParamConfig{
	"code":   {Type: fhir.SearchParamToken, Column: "code_value"},
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"form":   {Type: fhir.SearchParamToken, Column: "form_code"},
}

func (r *medicationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Medication, int, error) {
	qb := fhir.NewSearchQuery("medication", medCols)
	qb.ApplyParams(params, medicationSearchParams)
	qb.OrderBy("created_at DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Medication
	for rows.Next() {
		m, err := r.scanMed(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *medicationRepoPG) AddIngredient(ctx context.Context, ing *MedicationIngredient) error {
	ing.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_ingredient (id, medication_id, item_code, item_display, item_system,
			strength_numerator, strength_numerator_unit, strength_denominator, strength_denominator_unit,
			is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		ing.ID, ing.MedicationID, ing.ItemCode, ing.ItemDisplay, ing.ItemSystem,
		ing.StrengthNumerator, ing.StrengthNumeratorUnit, ing.StrengthDenominator, ing.StrengthDenominatorUnit,
		ing.IsActive)
	return err
}

func (r *medicationRepoPG) GetIngredients(ctx context.Context, medicationID uuid.UUID) ([]*MedicationIngredient, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, medication_id, item_code, item_display, item_system,
			strength_numerator, strength_numerator_unit, strength_denominator, strength_denominator_unit,
			is_active
		FROM medication_ingredient WHERE medication_id = $1`, medicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*MedicationIngredient
	for rows.Next() {
		var ing MedicationIngredient
		if err := rows.Scan(&ing.ID, &ing.MedicationID, &ing.ItemCode, &ing.ItemDisplay, &ing.ItemSystem,
			&ing.StrengthNumerator, &ing.StrengthNumeratorUnit, &ing.StrengthDenominator, &ing.StrengthDenominatorUnit,
			&ing.IsActive); err != nil {
			return nil, err
		}
		items = append(items, &ing)
	}
	return items, nil
}

func (r *medicationRepoPG) RemoveIngredient(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_ingredient WHERE id = $1`, id)
	return err
}

// =========== MedicationRequest Repository ===========

type medRequestRepoPG struct{ pool *pgxpool.Pool }

func NewMedicationRequestRepoPG(pool *pgxpool.Pool) MedicationRequestRepository {
	return &medRequestRepoPG{pool: pool}
}

func (r *medRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const medReqCols = `id, fhir_id, status, status_reason_code, status_reason_display, intent,
	category_code, category_display, priority, medication_id, patient_id, encounter_id,
	requester_id, performer_id, recorder_id,
	reason_code, reason_display, reason_condition_id,
	dosage_text, dosage_timing_code, dosage_timing_display,
	dosage_route_code, dosage_route_display,
	dosage_site_code, dosage_site_display,
	dosage_method_code, dosage_method_display,
	dose_quantity, dose_unit, max_dose_per_period, max_dose_per_period_unit,
	rate_quantity, rate_unit,
	as_needed, as_needed_code, as_needed_display,
	quantity_value, quantity_unit, days_supply, refills_allowed,
	validity_start, validity_end,
	substitution_allowed, substitution_reason,
	authored_on, prior_auth_number, erx_reference, abdm_prescription_id,
	note, created_at, updated_at`

func (r *medRequestRepoPG) scanReq(row pgx.Row) (*MedicationRequest, error) {
	var mr MedicationRequest
	err := row.Scan(&mr.ID, &mr.FHIRID, &mr.Status, &mr.StatusReasonCode, &mr.StatusReasonDisplay, &mr.Intent,
		&mr.CategoryCode, &mr.CategoryDisplay, &mr.Priority, &mr.MedicationID, &mr.PatientID, &mr.EncounterID,
		&mr.RequesterID, &mr.PerformerID, &mr.RecorderID,
		&mr.ReasonCode, &mr.ReasonDisplay, &mr.ReasonConditionID,
		&mr.DosageText, &mr.DosageTimingCode, &mr.DosageTimingDisplay,
		&mr.DosageRouteCode, &mr.DosageRouteDisplay,
		&mr.DosageSiteCode, &mr.DosageSiteDisplay,
		&mr.DosageMethodCode, &mr.DosageMethodDisplay,
		&mr.DoseQuantity, &mr.DoseUnit, &mr.MaxDosePerPeriod, &mr.MaxDosePerPeriodUnit,
		&mr.RateQuantity, &mr.RateUnit,
		&mr.AsNeeded, &mr.AsNeededCode, &mr.AsNeededDisplay,
		&mr.QuantityValue, &mr.QuantityUnit, &mr.DaysSupply, &mr.RefillsAllowed,
		&mr.ValidityStart, &mr.ValidityEnd,
		&mr.SubstitutionAllowed, &mr.SubstitutionReason,
		&mr.AuthoredOn, &mr.PriorAuthNumber, &mr.ERxReference, &mr.ABDMPrescriptionID,
		&mr.Note, &mr.CreatedAt, &mr.UpdatedAt)
	return &mr, err
}

func (r *medRequestRepoPG) Create(ctx context.Context, mr *MedicationRequest) error {
	mr.ID = uuid.New()
	if mr.FHIRID == "" {
		mr.FHIRID = mr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_request (id, fhir_id, status, status_reason_code, status_reason_display, intent,
			category_code, category_display, priority, medication_id, patient_id, encounter_id,
			requester_id, performer_id, recorder_id,
			reason_code, reason_display, reason_condition_id,
			dosage_text, dosage_timing_code, dosage_timing_display,
			dosage_route_code, dosage_route_display,
			dosage_site_code, dosage_site_display,
			dosage_method_code, dosage_method_display,
			dose_quantity, dose_unit, max_dose_per_period, max_dose_per_period_unit,
			rate_quantity, rate_unit,
			as_needed, as_needed_code, as_needed_display,
			quantity_value, quantity_unit, days_supply, refills_allowed,
			validity_start, validity_end,
			substitution_allowed, substitution_reason,
			authored_on, prior_auth_number, erx_reference, abdm_prescription_id, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45,$46,$47,$48,$49)`,
		mr.ID, mr.FHIRID, mr.Status, mr.StatusReasonCode, mr.StatusReasonDisplay, mr.Intent,
		mr.CategoryCode, mr.CategoryDisplay, mr.Priority, mr.MedicationID, mr.PatientID, mr.EncounterID,
		mr.RequesterID, mr.PerformerID, mr.RecorderID,
		mr.ReasonCode, mr.ReasonDisplay, mr.ReasonConditionID,
		mr.DosageText, mr.DosageTimingCode, mr.DosageTimingDisplay,
		mr.DosageRouteCode, mr.DosageRouteDisplay,
		mr.DosageSiteCode, mr.DosageSiteDisplay,
		mr.DosageMethodCode, mr.DosageMethodDisplay,
		mr.DoseQuantity, mr.DoseUnit, mr.MaxDosePerPeriod, mr.MaxDosePerPeriodUnit,
		mr.RateQuantity, mr.RateUnit,
		mr.AsNeeded, mr.AsNeededCode, mr.AsNeededDisplay,
		mr.QuantityValue, mr.QuantityUnit, mr.DaysSupply, mr.RefillsAllowed,
		mr.ValidityStart, mr.ValidityEnd,
		mr.SubstitutionAllowed, mr.SubstitutionReason,
		mr.AuthoredOn, mr.PriorAuthNumber, mr.ERxReference, mr.ABDMPrescriptionID, mr.Note)
	return err
}

func (r *medRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicationRequest, error) {
	return r.scanReq(r.conn(ctx).QueryRow(ctx, `SELECT `+medReqCols+` FROM medication_request WHERE id = $1`, id))
}

func (r *medRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicationRequest, error) {
	return r.scanReq(r.conn(ctx).QueryRow(ctx, `SELECT `+medReqCols+` FROM medication_request WHERE fhir_id = $1`, fhirID))
}

func (r *medRequestRepoPG) Update(ctx context.Context, mr *MedicationRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication_request SET status=$2, status_reason_code=$3, status_reason_display=$4,
			priority=$5, dosage_text=$6, dose_quantity=$7, dose_unit=$8,
			quantity_value=$9, quantity_unit=$10, days_supply=$11, refills_allowed=$12,
			substitution_allowed=$13, note=$14, updated_at=NOW()
		WHERE id = $1`,
		mr.ID, mr.Status, mr.StatusReasonCode, mr.StatusReasonDisplay,
		mr.Priority, mr.DosageText, mr.DoseQuantity, mr.DoseUnit,
		mr.QuantityValue, mr.QuantityUnit, mr.DaysSupply, mr.RefillsAllowed,
		mr.SubstitutionAllowed, mr.Note)
	return err
}

func (r *medRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_request WHERE id = $1`, id)
	return err
}

func (r *medRequestRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_request WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+medReqCols+` FROM medication_request WHERE patient_id = $1 ORDER BY authored_on DESC NULLS LAST LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationRequest
	for rows.Next() {
		mr, err := r.scanReq(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mr)
	}
	return items, total, nil
}

var medRequestSearchParams = map[string]fhir.SearchParamConfig{
	"patient":    {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":     {Type: fhir.SearchParamToken, Column: "status"},
	"intent":     {Type: fhir.SearchParamToken, Column: "intent"},
	"medication": {Type: fhir.SearchParamReference, Column: "medication_id"},
	"authoredon": {Type: fhir.SearchParamDate, Column: "authored_on"},
	"_id":        {Type: fhir.SearchParamToken, Column: "fhir_id"},
}

func (r *medRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationRequest, int, error) {
	qb := fhir.NewSearchQuery("medication_request", medReqCols)
	qb.ApplyParams(params, medRequestSearchParams)
	qb.OrderBy("authored_on DESC NULLS LAST")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationRequest
	for rows.Next() {
		mr, err := r.scanReq(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mr)
	}
	return items, total, nil
}

// =========== MedicationAdministration Repository ===========

type medAdminRepoPG struct{ pool *pgxpool.Pool }

func NewMedicationAdministrationRepoPG(pool *pgxpool.Pool) MedicationAdministrationRepository {
	return &medAdminRepoPG{pool: pool}
}

func (r *medAdminRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const medAdminCols = `id, fhir_id, status, status_reason_code, status_reason_display,
	category_code, category_display, medication_id, patient_id, encounter_id,
	medication_request_id, performer_id, performer_role_code, performer_role_display,
	effective_datetime, effective_start, effective_end,
	reason_code, reason_display, reason_condition_id,
	dosage_text, dosage_route_code, dosage_route_display,
	dosage_site_code, dosage_site_display,
	dosage_method_code, dosage_method_display,
	dose_quantity, dose_unit, rate_quantity, rate_unit,
	note, created_at, updated_at`

func (r *medAdminRepoPG) scanAdmin(row pgx.Row) (*MedicationAdministration, error) {
	var ma MedicationAdministration
	err := row.Scan(&ma.ID, &ma.FHIRID, &ma.Status, &ma.StatusReasonCode, &ma.StatusReasonDisplay,
		&ma.CategoryCode, &ma.CategoryDisplay, &ma.MedicationID, &ma.PatientID, &ma.EncounterID,
		&ma.MedicationRequestID, &ma.PerformerID, &ma.PerformerRoleCode, &ma.PerformerRoleDisplay,
		&ma.EffectiveDatetime, &ma.EffectiveStart, &ma.EffectiveEnd,
		&ma.ReasonCode, &ma.ReasonDisplay, &ma.ReasonConditionID,
		&ma.DosageText, &ma.DosageRouteCode, &ma.DosageRouteDisplay,
		&ma.DosageSiteCode, &ma.DosageSiteDisplay,
		&ma.DosageMethodCode, &ma.DosageMethodDisplay,
		&ma.DoseQuantity, &ma.DoseUnit, &ma.RateQuantity, &ma.RateUnit,
		&ma.Note, &ma.CreatedAt, &ma.UpdatedAt)
	return &ma, err
}

func (r *medAdminRepoPG) Create(ctx context.Context, ma *MedicationAdministration) error {
	ma.ID = uuid.New()
	if ma.FHIRID == "" {
		ma.FHIRID = ma.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_administration (id, fhir_id, status, status_reason_code, status_reason_display,
			category_code, category_display, medication_id, patient_id, encounter_id,
			medication_request_id, performer_id, performer_role_code, performer_role_display,
			effective_datetime, effective_start, effective_end,
			reason_code, reason_display, reason_condition_id,
			dosage_text, dosage_route_code, dosage_route_display,
			dosage_site_code, dosage_site_display,
			dosage_method_code, dosage_method_display,
			dose_quantity, dose_unit, rate_quantity, rate_unit, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32)`,
		ma.ID, ma.FHIRID, ma.Status, ma.StatusReasonCode, ma.StatusReasonDisplay,
		ma.CategoryCode, ma.CategoryDisplay, ma.MedicationID, ma.PatientID, ma.EncounterID,
		ma.MedicationRequestID, ma.PerformerID, ma.PerformerRoleCode, ma.PerformerRoleDisplay,
		ma.EffectiveDatetime, ma.EffectiveStart, ma.EffectiveEnd,
		ma.ReasonCode, ma.ReasonDisplay, ma.ReasonConditionID,
		ma.DosageText, ma.DosageRouteCode, ma.DosageRouteDisplay,
		ma.DosageSiteCode, ma.DosageSiteDisplay,
		ma.DosageMethodCode, ma.DosageMethodDisplay,
		ma.DoseQuantity, ma.DoseUnit, ma.RateQuantity, ma.RateUnit, ma.Note)
	return err
}

func (r *medAdminRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicationAdministration, error) {
	return r.scanAdmin(r.conn(ctx).QueryRow(ctx, `SELECT `+medAdminCols+` FROM medication_administration WHERE id = $1`, id))
}

func (r *medAdminRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicationAdministration, error) {
	return r.scanAdmin(r.conn(ctx).QueryRow(ctx, `SELECT `+medAdminCols+` FROM medication_administration WHERE fhir_id = $1`, fhirID))
}

func (r *medAdminRepoPG) Update(ctx context.Context, ma *MedicationAdministration) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication_administration SET status=$2, status_reason_code=$3, status_reason_display=$4,
			effective_datetime=$5, effective_start=$6, effective_end=$7,
			dose_quantity=$8, dose_unit=$9, rate_quantity=$10, rate_unit=$11,
			note=$12, updated_at=NOW()
		WHERE id = $1`,
		ma.ID, ma.Status, ma.StatusReasonCode, ma.StatusReasonDisplay,
		ma.EffectiveDatetime, ma.EffectiveStart, ma.EffectiveEnd,
		ma.DoseQuantity, ma.DoseUnit, ma.RateQuantity, ma.RateUnit,
		ma.Note)
	return err
}

func (r *medAdminRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_administration WHERE id = $1`, id)
	return err
}

func (r *medAdminRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationAdministration, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_administration WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+medAdminCols+` FROM medication_administration WHERE patient_id = $1 ORDER BY effective_datetime DESC NULLS LAST LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationAdministration
	for rows.Next() {
		ma, err := r.scanAdmin(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ma)
	}
	return items, total, nil
}

var medAdminSearchParams = map[string]fhir.SearchParamConfig{
	"patient":    {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":     {Type: fhir.SearchParamToken, Column: "status"},
	"medication": {Type: fhir.SearchParamReference, Column: "medication_id"},
}

func (r *medAdminRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationAdministration, int, error) {
	qb := fhir.NewSearchQuery("medication_administration", medAdminCols)
	qb.ApplyParams(params, medAdminSearchParams)
	qb.OrderBy("effective_datetime DESC NULLS LAST")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationAdministration
	for rows.Next() {
		ma, err := r.scanAdmin(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ma)
	}
	return items, total, nil
}

// =========== MedicationDispense Repository ===========

type medDispenseRepoPG struct{ pool *pgxpool.Pool }

func NewMedicationDispenseRepoPG(pool *pgxpool.Pool) MedicationDispenseRepository {
	return &medDispenseRepoPG{pool: pool}
}

func (r *medDispenseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const medDispCols = `id, fhir_id, status, status_reason_code, status_reason_display,
	category_code, category_display, medication_id, patient_id, encounter_id,
	medication_request_id, performer_id, location_id,
	quantity_value, quantity_unit, days_supply,
	when_prepared, when_handed_over,
	destination_id, receiver_id,
	was_substituted, substitution_type_code, substitution_reason,
	note, created_at, updated_at`

func (r *medDispenseRepoPG) scanDisp(row pgx.Row) (*MedicationDispense, error) {
	var md MedicationDispense
	err := row.Scan(&md.ID, &md.FHIRID, &md.Status, &md.StatusReasonCode, &md.StatusReasonDisplay,
		&md.CategoryCode, &md.CategoryDisplay, &md.MedicationID, &md.PatientID, &md.EncounterID,
		&md.MedicationRequestID, &md.PerformerID, &md.LocationID,
		&md.QuantityValue, &md.QuantityUnit, &md.DaysSupply,
		&md.WhenPrepared, &md.WhenHandedOver,
		&md.DestinationID, &md.ReceiverID,
		&md.WasSubstituted, &md.SubstitutionTypeCode, &md.SubstitutionReason,
		&md.Note, &md.CreatedAt, &md.UpdatedAt)
	return &md, err
}

func (r *medDispenseRepoPG) Create(ctx context.Context, md *MedicationDispense) error {
	md.ID = uuid.New()
	if md.FHIRID == "" {
		md.FHIRID = md.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_dispense (id, fhir_id, status, status_reason_code, status_reason_display,
			category_code, category_display, medication_id, patient_id, encounter_id,
			medication_request_id, performer_id, location_id,
			quantity_value, quantity_unit, days_supply,
			when_prepared, when_handed_over,
			destination_id, receiver_id,
			was_substituted, substitution_type_code, substitution_reason, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		md.ID, md.FHIRID, md.Status, md.StatusReasonCode, md.StatusReasonDisplay,
		md.CategoryCode, md.CategoryDisplay, md.MedicationID, md.PatientID, md.EncounterID,
		md.MedicationRequestID, md.PerformerID, md.LocationID,
		md.QuantityValue, md.QuantityUnit, md.DaysSupply,
		md.WhenPrepared, md.WhenHandedOver,
		md.DestinationID, md.ReceiverID,
		md.WasSubstituted, md.SubstitutionTypeCode, md.SubstitutionReason, md.Note)
	return err
}

func (r *medDispenseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicationDispense, error) {
	return r.scanDisp(r.conn(ctx).QueryRow(ctx, `SELECT `+medDispCols+` FROM medication_dispense WHERE id = $1`, id))
}

func (r *medDispenseRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicationDispense, error) {
	return r.scanDisp(r.conn(ctx).QueryRow(ctx, `SELECT `+medDispCols+` FROM medication_dispense WHERE fhir_id = $1`, fhirID))
}

func (r *medDispenseRepoPG) Update(ctx context.Context, md *MedicationDispense) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication_dispense SET status=$2, status_reason_code=$3, status_reason_display=$4,
			quantity_value=$5, quantity_unit=$6, days_supply=$7,
			when_prepared=$8, when_handed_over=$9,
			was_substituted=$10, substitution_type_code=$11, substitution_reason=$12,
			note=$13, updated_at=NOW()
		WHERE id = $1`,
		md.ID, md.Status, md.StatusReasonCode, md.StatusReasonDisplay,
		md.QuantityValue, md.QuantityUnit, md.DaysSupply,
		md.WhenPrepared, md.WhenHandedOver,
		md.WasSubstituted, md.SubstitutionTypeCode, md.SubstitutionReason,
		md.Note)
	return err
}

func (r *medDispenseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_dispense WHERE id = $1`, id)
	return err
}

func (r *medDispenseRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationDispense, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_dispense WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+medDispCols+` FROM medication_dispense WHERE patient_id = $1 ORDER BY when_handed_over DESC NULLS LAST LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationDispense
	for rows.Next() {
		md, err := r.scanDisp(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, md)
	}
	return items, total, nil
}

var medDispenseSearchParams = map[string]fhir.SearchParamConfig{
	"patient":    {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":     {Type: fhir.SearchParamToken, Column: "status"},
	"medication": {Type: fhir.SearchParamReference, Column: "medication_id"},
}

func (r *medDispenseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationDispense, int, error) {
	qb := fhir.NewSearchQuery("medication_dispense", medDispCols)
	qb.ApplyParams(params, medDispenseSearchParams)
	qb.OrderBy("when_handed_over DESC NULLS LAST")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationDispense
	for rows.Next() {
		md, err := r.scanDisp(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, md)
	}
	return items, total, nil
}

// =========== MedicationStatement Repository ===========

type medStatementRepoPG struct{ pool *pgxpool.Pool }

func NewMedicationStatementRepoPG(pool *pgxpool.Pool) MedicationStatementRepository {
	return &medStatementRepoPG{pool: pool}
}

func (r *medStatementRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const medStmtCols = `id, fhir_id, status, status_reason_code, status_reason_display,
	category_code, category_display,
	medication_code, medication_display, medication_id,
	patient_id, encounter_id, information_source_id,
	effective_datetime, effective_start, effective_end, date_asserted,
	reason_code, reason_display,
	dosage_text, dosage_route_code, dosage_route_display,
	dose_quantity, dose_unit, dosage_timing_code, dosage_timing_display,
	note, created_at, updated_at`

func (r *medStatementRepoPG) scanStmt(row pgx.Row) (*MedicationStatement, error) {
	var ms MedicationStatement
	err := row.Scan(&ms.ID, &ms.FHIRID, &ms.Status, &ms.StatusReasonCode, &ms.StatusReasonDisplay,
		&ms.CategoryCode, &ms.CategoryDisplay,
		&ms.MedicationCode, &ms.MedicationDisplay, &ms.MedicationID,
		&ms.PatientID, &ms.EncounterID, &ms.InformationSourceID,
		&ms.EffectiveDatetime, &ms.EffectiveStart, &ms.EffectiveEnd, &ms.DateAsserted,
		&ms.ReasonCode, &ms.ReasonDisplay,
		&ms.DosageText, &ms.DosageRouteCode, &ms.DosageRouteDisplay,
		&ms.DoseQuantity, &ms.DoseUnit, &ms.DosageTimingCode, &ms.DosageTimingDisplay,
		&ms.Note, &ms.CreatedAt, &ms.UpdatedAt)
	return &ms, err
}

func (r *medStatementRepoPG) Create(ctx context.Context, ms *MedicationStatement) error {
	ms.ID = uuid.New()
	if ms.FHIRID == "" {
		ms.FHIRID = ms.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO medication_statement (id, fhir_id, status, status_reason_code, status_reason_display,
			category_code, category_display,
			medication_code, medication_display, medication_id,
			patient_id, encounter_id, information_source_id,
			effective_datetime, effective_start, effective_end, date_asserted,
			reason_code, reason_display,
			dosage_text, dosage_route_code, dosage_route_display,
			dose_quantity, dose_unit, dosage_timing_code, dosage_timing_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)`,
		ms.ID, ms.FHIRID, ms.Status, ms.StatusReasonCode, ms.StatusReasonDisplay,
		ms.CategoryCode, ms.CategoryDisplay,
		ms.MedicationCode, ms.MedicationDisplay, ms.MedicationID,
		ms.PatientID, ms.EncounterID, ms.InformationSourceID,
		ms.EffectiveDatetime, ms.EffectiveStart, ms.EffectiveEnd, ms.DateAsserted,
		ms.ReasonCode, ms.ReasonDisplay,
		ms.DosageText, ms.DosageRouteCode, ms.DosageRouteDisplay,
		ms.DoseQuantity, ms.DoseUnit, ms.DosageTimingCode, ms.DosageTimingDisplay, ms.Note)
	return err
}

func (r *medStatementRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MedicationStatement, error) {
	return r.scanStmt(r.conn(ctx).QueryRow(ctx, `SELECT `+medStmtCols+` FROM medication_statement WHERE id = $1`, id))
}

func (r *medStatementRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*MedicationStatement, error) {
	return r.scanStmt(r.conn(ctx).QueryRow(ctx, `SELECT `+medStmtCols+` FROM medication_statement WHERE fhir_id = $1`, fhirID))
}

func (r *medStatementRepoPG) Update(ctx context.Context, ms *MedicationStatement) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE medication_statement SET status=$2, status_reason_code=$3, status_reason_display=$4,
			medication_code=$5, medication_display=$6,
			effective_datetime=$7, effective_start=$8, effective_end=$9,
			dosage_text=$10, dose_quantity=$11, dose_unit=$12,
			note=$13, updated_at=NOW()
		WHERE id = $1`,
		ms.ID, ms.Status, ms.StatusReasonCode, ms.StatusReasonDisplay,
		ms.MedicationCode, ms.MedicationDisplay,
		ms.EffectiveDatetime, ms.EffectiveStart, ms.EffectiveEnd,
		ms.DosageText, ms.DoseQuantity, ms.DoseUnit,
		ms.Note)
	return err
}

func (r *medStatementRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM medication_statement WHERE id = $1`, id)
	return err
}

func (r *medStatementRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationStatement, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM medication_statement WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+medStmtCols+` FROM medication_statement WHERE patient_id = $1 ORDER BY date_asserted DESC NULLS LAST LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationStatement
	for rows.Next() {
		ms, err := r.scanStmt(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ms)
	}
	return items, total, nil
}

var medStatementSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *medStatementRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicationStatement, int, error) {
	qb := fhir.NewSearchQuery("medication_statement", medStmtCols)
	qb.ApplyParams(params, medStatementSearchParams)
	qb.OrderBy("date_asserted DESC NULLS LAST")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MedicationStatement
	for rows.Next() {
		ms, err := r.scanStmt(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, ms)
	}
	return items, total, nil
}
