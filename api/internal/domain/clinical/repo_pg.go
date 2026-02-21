package clinical

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

// =========== Condition Repository ===========

type conditionRepoPG struct{ pool *pgxpool.Pool }

func NewConditionRepoPG(pool *pgxpool.Pool) ConditionRepository { return &conditionRepoPG{pool: pool} }

func (r *conditionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const condCols = `id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
	clinical_status, verification_status, category_code, severity_code, severity_display,
	code_system, code_value, code_display, alt_code_system, alt_code_value, alt_code_display,
	body_site_code, body_site_display, onset_datetime, onset_age, onset_string,
	abatement_datetime, abatement_age, abatement_string,
	stage_summary_code, stage_summary_display, stage_type_code,
	evidence_code, evidence_display, recorded_date, note, created_at, updated_at`

func (r *conditionRepoPG) scanCondition(row pgx.Row) (*Condition, error) {
	var c Condition
	err := row.Scan(&c.ID, &c.FHIRID, &c.PatientID, &c.EncounterID, &c.RecorderID, &c.AsserterID,
		&c.ClinicalStatus, &c.VerificationStatus, &c.CategoryCode, &c.SeverityCode, &c.SeverityDisplay,
		&c.CodeSystem, &c.CodeValue, &c.CodeDisplay, &c.AltCodeSystem, &c.AltCodeValue, &c.AltCodeDisplay,
		&c.BodySiteCode, &c.BodySiteDisplay, &c.OnsetDatetime, &c.OnsetAge, &c.OnsetString,
		&c.AbatementDatetime, &c.AbatementAge, &c.AbatementString,
		&c.StageSummaryCode, &c.StageSummaryDisp, &c.StageTypeCode,
		&c.EvidenceCode, &c.EvidenceDisplay, &c.RecordedDate, &c.Note, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *conditionRepoPG) Create(ctx context.Context, c *Condition) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO condition (id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
			clinical_status, verification_status, category_code, severity_code, severity_display,
			code_system, code_value, code_display, body_site_code, body_site_display,
			onset_datetime, onset_age, onset_string, recorded_date, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		c.ID, c.FHIRID, c.PatientID, c.EncounterID, c.RecorderID, c.AsserterID,
		c.ClinicalStatus, c.VerificationStatus, c.CategoryCode, c.SeverityCode, c.SeverityDisplay,
		c.CodeSystem, c.CodeValue, c.CodeDisplay, c.BodySiteCode, c.BodySiteDisplay,
		c.OnsetDatetime, c.OnsetAge, c.OnsetString, c.RecordedDate, c.Note)
	return err
}

func (r *conditionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Condition, error) {
	return r.scanCondition(r.conn(ctx).QueryRow(ctx, `SELECT `+condCols+` FROM condition WHERE id = $1`, id))
}

func (r *conditionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Condition, error) {
	return r.scanCondition(r.conn(ctx).QueryRow(ctx, `SELECT `+condCols+` FROM condition WHERE fhir_id = $1`, fhirID))
}

func (r *conditionRepoPG) Update(ctx context.Context, c *Condition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE condition SET clinical_status=$2, verification_status=$3, category_code=$4,
			severity_code=$5, severity_display=$6, code_system=$7, code_value=$8, code_display=$9,
			onset_datetime=$10, abatement_datetime=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.ClinicalStatus, c.VerificationStatus, c.CategoryCode,
		c.SeverityCode, c.SeverityDisplay, c.CodeSystem, c.CodeValue, c.CodeDisplay,
		c.OnsetDatetime, c.AbatementDatetime, c.Note)
	return err
}

func (r *conditionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM condition WHERE id = $1`, id)
	return err
}

func (r *conditionRepoPG) List(ctx context.Context, limit, offset int) ([]*Condition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM condition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+condCols+` FROM condition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Condition
	for rows.Next() {
		c, err := r.scanCondition(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

func (r *conditionRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Condition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM condition WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+condCols+` FROM condition WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Condition
	for rows.Next() {
		c, err := r.scanCondition(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

var conditionSearchParams = map[string]fhir.SearchParamConfig{
	"patient":         {Type: fhir.SearchParamReference, Column: "patient_id"},
	"clinical-status": {Type: fhir.SearchParamToken, Column: "clinical_status"},
	"category":        {Type: fhir.SearchParamToken, Column: "category_code"},
	"code":            {Type: fhir.SearchParamToken, Column: "code_value", SysColumn: "code_system"},
	"onset-date":      {Type: fhir.SearchParamDate, Column: "onset_datetime"},
	"_id":             {Type: fhir.SearchParamToken, Column: "fhir_id"},
}

func (r *conditionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Condition, int, error) {
	qb := fhir.NewSearchQuery("condition", condCols)
	qb.ApplyParams(params, conditionSearchParams)
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
	var items []*Condition
	for rows.Next() {
		c, err := r.scanCondition(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

// =========== Observation Repository ===========

type observationRepoPG struct{ pool *pgxpool.Pool }

func NewObservationRepoPG(pool *pgxpool.Pool) ObservationRepository {
	return &observationRepoPG{pool: pool}
}

func (r *observationRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const obsCols = `id, fhir_id, status, category_code, category_display,
	code_system, code_value, code_display, patient_id, encounter_id, performer_id,
	effective_datetime, issued,
	value_quantity, value_unit, value_system, value_code, value_string, value_boolean, value_integer,
	value_codeable_code, value_codeable_display,
	reference_range_low, reference_range_high, reference_range_unit, reference_range_text,
	interpretation_code, interpretation_display,
	body_site_code, body_site_display, data_absent_reason, note, created_at, updated_at`

func (r *observationRepoPG) scanObs(row pgx.Row) (*Observation, error) {
	var o Observation
	err := row.Scan(&o.ID, &o.FHIRID, &o.Status, &o.CategoryCode, &o.CategoryDisplay,
		&o.CodeSystem, &o.CodeValue, &o.CodeDisplay, &o.PatientID, &o.EncounterID, &o.PerformerID,
		&o.EffectiveDatetime, &o.Issued,
		&o.ValueQuantity, &o.ValueUnit, &o.ValueSystem, &o.ValueCode, &o.ValueString, &o.ValueBoolean, &o.ValueInteger,
		&o.ValueCodeableCode, &o.ValueCodeableDisplay,
		&o.ReferenceRangeLow, &o.ReferenceRangeHigh, &o.ReferenceRangeUnit, &o.ReferenceRangeText,
		&o.InterpretationCode, &o.InterpretationDisplay,
		&o.BodySiteCode, &o.BodySiteDisplay, &o.DataAbsentReason, &o.Note, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *observationRepoPG) Create(ctx context.Context, o *Observation) error {
	o.ID = uuid.New()
	if o.FHIRID == "" {
		o.FHIRID = o.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO observation (id, fhir_id, status, category_code, category_display,
			code_system, code_value, code_display, patient_id, encounter_id, performer_id,
			effective_datetime, value_quantity, value_unit, value_system, value_code,
			value_string, value_boolean, value_integer,
			value_codeable_code, value_codeable_display,
			interpretation_code, interpretation_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		o.ID, o.FHIRID, o.Status, o.CategoryCode, o.CategoryDisplay,
		o.CodeSystem, o.CodeValue, o.CodeDisplay, o.PatientID, o.EncounterID, o.PerformerID,
		o.EffectiveDatetime, o.ValueQuantity, o.ValueUnit, o.ValueSystem, o.ValueCode,
		o.ValueString, o.ValueBoolean, o.ValueInteger,
		o.ValueCodeableCode, o.ValueCodeableDisplay,
		o.InterpretationCode, o.InterpretationDisplay, o.Note)
	return err
}

func (r *observationRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Observation, error) {
	return r.scanObs(r.conn(ctx).QueryRow(ctx, `SELECT `+obsCols+` FROM observation WHERE id = $1`, id))
}

func (r *observationRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Observation, error) {
	return r.scanObs(r.conn(ctx).QueryRow(ctx, `SELECT `+obsCols+` FROM observation WHERE fhir_id = $1`, fhirID))
}

func (r *observationRepoPG) Update(ctx context.Context, o *Observation) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE observation SET status=$2, value_quantity=$3, value_unit=$4, value_string=$5,
			interpretation_code=$6, interpretation_display=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		o.ID, o.Status, o.ValueQuantity, o.ValueUnit, o.ValueString,
		o.InterpretationCode, o.InterpretationDisplay, o.Note)
	return err
}

func (r *observationRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM observation WHERE id = $1`, id)
	return err
}

func (r *observationRepoPG) List(ctx context.Context, limit, offset int) ([]*Observation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM observation`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+obsCols+` FROM observation ORDER BY effective_datetime DESC NULLS LAST LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Observation
	for rows.Next() {
		o, err := r.scanObs(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

func (r *observationRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Observation, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM observation WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+obsCols+` FROM observation WHERE patient_id = $1 ORDER BY effective_datetime DESC NULLS LAST LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Observation
	for rows.Next() {
		o, err := r.scanObs(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

var observationSearchParams = map[string]fhir.SearchParamConfig{
	"patient":  {Type: fhir.SearchParamReference, Column: "patient_id"},
	"category": {Type: fhir.SearchParamToken, Column: "category_code"},
	"code":     {Type: fhir.SearchParamToken, Column: "code_value", SysColumn: "code_system"},
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"date":     {Type: fhir.SearchParamDate, Column: "effective_datetime"},
	"_id":      {Type: fhir.SearchParamToken, Column: "fhir_id"},
}

func (r *observationRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Observation, int, error) {
	qb := fhir.NewSearchQuery("observation", obsCols)
	qb.ApplyParams(params, observationSearchParams)
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
	var items []*Observation
	for rows.Next() {
		o, err := r.scanObs(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

func (r *observationRepoPG) AddComponent(ctx context.Context, c *ObservationComponent) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO observation_component (id, observation_id, code_system, code_value, code_display,
			value_quantity, value_unit, value_string, value_codeable_code, value_codeable_display,
			interpretation_code, interpretation_display,
			reference_range_low, reference_range_high, reference_range_unit)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		c.ID, c.ObservationID, c.CodeSystem, c.CodeValue, c.CodeDisplay,
		c.ValueQuantity, c.ValueUnit, c.ValueString, c.ValueCodeableCode, c.ValueCodeableDisplay,
		c.InterpretationCode, c.InterpretationDisplay,
		c.ReferenceRangeLow, c.ReferenceRangeHigh, c.ReferenceRangeUnit)
	return err
}

func (r *observationRepoPG) GetComponents(ctx context.Context, observationID uuid.UUID) ([]*ObservationComponent, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, observation_id, code_system, code_value, code_display,
			value_quantity, value_unit, value_string, value_codeable_code, value_codeable_display,
			interpretation_code, interpretation_display,
			reference_range_low, reference_range_high, reference_range_unit
		FROM observation_component WHERE observation_id = $1`, observationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ObservationComponent
	for rows.Next() {
		var c ObservationComponent
		if err := rows.Scan(&c.ID, &c.ObservationID, &c.CodeSystem, &c.CodeValue, &c.CodeDisplay,
			&c.ValueQuantity, &c.ValueUnit, &c.ValueString, &c.ValueCodeableCode, &c.ValueCodeableDisplay,
			&c.InterpretationCode, &c.InterpretationDisplay,
			&c.ReferenceRangeLow, &c.ReferenceRangeHigh, &c.ReferenceRangeUnit); err != nil {
			return nil, err
		}
		items = append(items, &c)
	}
	return items, nil
}

// =========== Allergy Repository ===========

type allergyRepoPG struct{ pool *pgxpool.Pool }

func NewAllergyRepoPG(pool *pgxpool.Pool) AllergyRepository { return &allergyRepoPG{pool: pool} }

func (r *allergyRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const allergyCols = `id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
	clinical_status, verification_status, type, category, criticality,
	code_system, code_value, code_display,
	onset_datetime, onset_age, onset_string, recorded_date, last_occurrence, note, created_at, updated_at`

func (r *allergyRepoPG) scanAllergy(row pgx.Row) (*AllergyIntolerance, error) {
	var a AllergyIntolerance
	err := row.Scan(&a.ID, &a.FHIRID, &a.PatientID, &a.EncounterID, &a.RecorderID, &a.AsserterID,
		&a.ClinicalStatus, &a.VerificationStatus, &a.Type, &a.Category, &a.Criticality,
		&a.CodeSystem, &a.CodeValue, &a.CodeDisplay,
		&a.OnsetDatetime, &a.OnsetAge, &a.OnsetString, &a.RecordedDate, &a.LastOccurrence, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *allergyRepoPG) Create(ctx context.Context, a *AllergyIntolerance) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO allergy_intolerance (id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
			clinical_status, verification_status, type, category, criticality,
			code_system, code_value, code_display,
			onset_datetime, onset_age, onset_string, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		a.ID, a.FHIRID, a.PatientID, a.EncounterID, a.RecorderID, a.AsserterID,
		a.ClinicalStatus, a.VerificationStatus, a.Type, a.Category, a.Criticality,
		a.CodeSystem, a.CodeValue, a.CodeDisplay,
		a.OnsetDatetime, a.OnsetAge, a.OnsetString, a.Note)
	return err
}

func (r *allergyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*AllergyIntolerance, error) {
	return r.scanAllergy(r.conn(ctx).QueryRow(ctx, `SELECT `+allergyCols+` FROM allergy_intolerance WHERE id = $1`, id))
}

func (r *allergyRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*AllergyIntolerance, error) {
	return r.scanAllergy(r.conn(ctx).QueryRow(ctx, `SELECT `+allergyCols+` FROM allergy_intolerance WHERE fhir_id = $1`, fhirID))
}

func (r *allergyRepoPG) Update(ctx context.Context, a *AllergyIntolerance) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE allergy_intolerance SET clinical_status=$2, verification_status=$3, type=$4,
			category=$5, criticality=$6, code_value=$7, code_display=$8, note=$9, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.ClinicalStatus, a.VerificationStatus, a.Type,
		a.Category, a.Criticality, a.CodeValue, a.CodeDisplay, a.Note)
	return err
}

func (r *allergyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM allergy_intolerance WHERE id = $1`, id)
	return err
}

func (r *allergyRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*AllergyIntolerance, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM allergy_intolerance WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+allergyCols+` FROM allergy_intolerance WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*AllergyIntolerance
	for rows.Next() {
		a, err := r.scanAllergy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

var allergySearchParams = map[string]fhir.SearchParamConfig{
	"patient":         {Type: fhir.SearchParamReference, Column: "patient_id"},
	"clinical-status": {Type: fhir.SearchParamToken, Column: "clinical_status"},
	"code":            {Type: fhir.SearchParamToken, Column: "code_value", SysColumn: "code_system"},
	"_id":             {Type: fhir.SearchParamToken, Column: "fhir_id"},
}

func (r *allergyRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*AllergyIntolerance, int, error) {
	qb := fhir.NewSearchQuery("allergy_intolerance", allergyCols)
	qb.ApplyParams(params, allergySearchParams)
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
	var items []*AllergyIntolerance
	for rows.Next() {
		a, err := r.scanAllergy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *allergyRepoPG) AddReaction(ctx context.Context, rx *AllergyReaction) error {
	rx.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO allergy_reaction (id, allergy_id, substance_code, substance_display,
			manifestation_code, manifestation_display, description, severity,
			exposure_route_code, exposure_route_display, onset, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		rx.ID, rx.AllergyID, rx.SubstanceCode, rx.SubstanceDisplay,
		rx.ManifestationCode, rx.ManifestationDisplay, rx.Description, rx.Severity,
		rx.ExposureRouteCode, rx.ExposureRouteDisplay, rx.Onset, rx.Note)
	return err
}

func (r *allergyRepoPG) GetReactions(ctx context.Context, allergyID uuid.UUID) ([]*AllergyReaction, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, allergy_id, substance_code, substance_display,
			manifestation_code, manifestation_display, description, severity,
			exposure_route_code, exposure_route_display, onset, note
		FROM allergy_reaction WHERE allergy_id = $1`, allergyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*AllergyReaction
	for rows.Next() {
		var rx AllergyReaction
		if err := rows.Scan(&rx.ID, &rx.AllergyID, &rx.SubstanceCode, &rx.SubstanceDisplay,
			&rx.ManifestationCode, &rx.ManifestationDisplay, &rx.Description, &rx.Severity,
			&rx.ExposureRouteCode, &rx.ExposureRouteDisplay, &rx.Onset, &rx.Note); err != nil {
			return nil, err
		}
		items = append(items, &rx)
	}
	return items, nil
}

func (r *allergyRepoPG) RemoveReaction(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM allergy_reaction WHERE id = $1`, id)
	return err
}

// =========== Procedure Repository ===========

type procedureRepoPG struct{ pool *pgxpool.Pool }

func NewProcedureRepoPG(pool *pgxpool.Pool) ProcedureRepository {
	return &procedureRepoPG{pool: pool}
}

func (r *procedureRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const procCols = `id, fhir_id, status, status_reason_code, patient_id, encounter_id, recorder_id, asserter_id,
	code_system, code_value, code_display, category_code, category_display,
	performed_datetime, performed_start, performed_end, performed_string,
	body_site_code, body_site_display, outcome_code, outcome_display,
	complication_code, complication_display,
	reason_code, reason_display, reason_condition_id, location_id, anesthesia_type,
	cpt_code, hcpcs_code, note, created_at, updated_at`

func (r *procedureRepoPG) scanProc(row pgx.Row) (*ProcedureRecord, error) {
	var p ProcedureRecord
	err := row.Scan(&p.ID, &p.FHIRID, &p.Status, &p.StatusReasonCode, &p.PatientID, &p.EncounterID, &p.RecorderID, &p.AsserterID,
		&p.CodeSystem, &p.CodeValue, &p.CodeDisplay, &p.CategoryCode, &p.CategoryDisplay,
		&p.PerformedDatetime, &p.PerformedStart, &p.PerformedEnd, &p.PerformedString,
		&p.BodySiteCode, &p.BodySiteDisplay, &p.OutcomeCode, &p.OutcomeDisplay,
		&p.ComplicationCode, &p.ComplicationDisp,
		&p.ReasonCode, &p.ReasonDisplay, &p.ReasonConditionID, &p.LocationID, &p.AnesthesiaType,
		&p.CPTCode, &p.HCPCSCode, &p.Note, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *procedureRepoPG) Create(ctx context.Context, p *ProcedureRecord) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO procedure_record (id, fhir_id, status, status_reason_code, patient_id, encounter_id,
			recorder_id, asserter_id, code_system, code_value, code_display,
			category_code, category_display, performed_datetime, performed_start, performed_end,
			body_site_code, body_site_display, outcome_code, outcome_display,
			reason_code, reason_display, location_id, anesthesia_type, cpt_code, hcpcs_code, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)`,
		p.ID, p.FHIRID, p.Status, p.StatusReasonCode, p.PatientID, p.EncounterID,
		p.RecorderID, p.AsserterID, p.CodeSystem, p.CodeValue, p.CodeDisplay,
		p.CategoryCode, p.CategoryDisplay, p.PerformedDatetime, p.PerformedStart, p.PerformedEnd,
		p.BodySiteCode, p.BodySiteDisplay, p.OutcomeCode, p.OutcomeDisplay,
		p.ReasonCode, p.ReasonDisplay, p.LocationID, p.AnesthesiaType, p.CPTCode, p.HCPCSCode, p.Note)
	return err
}

func (r *procedureRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ProcedureRecord, error) {
	return r.scanProc(r.conn(ctx).QueryRow(ctx, `SELECT `+procCols+` FROM procedure_record WHERE id = $1`, id))
}

func (r *procedureRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ProcedureRecord, error) {
	return r.scanProc(r.conn(ctx).QueryRow(ctx, `SELECT `+procCols+` FROM procedure_record WHERE fhir_id = $1`, fhirID))
}

func (r *procedureRepoPG) Update(ctx context.Context, p *ProcedureRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE procedure_record SET status=$2, outcome_code=$3, outcome_display=$4,
			complication_code=$5, complication_display=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Status, p.OutcomeCode, p.OutcomeDisplay,
		p.ComplicationCode, p.ComplicationDisp, p.Note)
	return err
}

func (r *procedureRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM procedure_record WHERE id = $1`, id)
	return err
}

func (r *procedureRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ProcedureRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM procedure_record WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+procCols+` FROM procedure_record WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ProcedureRecord
	for rows.Next() {
		p, err := r.scanProc(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

var procedureSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
	"code":    {Type: fhir.SearchParamToken, Column: "code_value", SysColumn: "code_system"},
	"date":    {Type: fhir.SearchParamDate, Column: "performed_datetime"},
	"_id":     {Type: fhir.SearchParamToken, Column: "fhir_id"},
}

func (r *procedureRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ProcedureRecord, int, error) {
	qb := fhir.NewSearchQuery("procedure_record", procCols)
	qb.ApplyParams(params, procedureSearchParams)
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
	var items []*ProcedureRecord
	for rows.Next() {
		p, err := r.scanProc(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

func (r *procedureRepoPG) AddPerformer(ctx context.Context, pf *ProcedurePerformer) error {
	pf.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO procedure_performer (id, procedure_id, practitioner_id, role_code, role_display, organization_id)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		pf.ID, pf.ProcedureID, pf.PractitionerID, pf.RoleCode, pf.RoleDisplay, pf.OrganizationID)
	return err
}

func (r *procedureRepoPG) GetPerformers(ctx context.Context, procedureID uuid.UUID) ([]*ProcedurePerformer, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, procedure_id, practitioner_id, role_code, role_display, organization_id
		FROM procedure_performer WHERE procedure_id = $1`, procedureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ProcedurePerformer
	for rows.Next() {
		var pf ProcedurePerformer
		if err := rows.Scan(&pf.ID, &pf.ProcedureID, &pf.PractitionerID, &pf.RoleCode, &pf.RoleDisplay, &pf.OrganizationID); err != nil {
			return nil, err
		}
		items = append(items, &pf)
	}
	return items, nil
}

func (r *procedureRepoPG) RemovePerformer(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM procedure_performer WHERE id = $1`, id)
	return err
}
