package diagnostics

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

// =========== ServiceRequest Repository ===========

type serviceRequestRepoPG struct{ pool *pgxpool.Pool }

func NewServiceRequestRepoPG(pool *pgxpool.Pool) ServiceRequestRepository {
	return &serviceRequestRepoPG{pool: pool}
}

func (r *serviceRequestRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const srCols = `id, fhir_id, patient_id, encounter_id, requester_id, performer_id,
	status, intent, priority, category_code, category_display,
	code_system, code_value, code_display,
	order_detail_code, order_detail_display, quantity_value, quantity_unit,
	occurrence_datetime, occurrence_start, occurrence_end, authored_on,
	reason_code, reason_display, reason_condition_id,
	specimen_requirement, body_site_code, body_site_display,
	note, patient_instruction, created_at, updated_at`

func (r *serviceRequestRepoPG) scanSR(row pgx.Row) (*ServiceRequest, error) {
	var sr ServiceRequest
	err := row.Scan(&sr.ID, &sr.FHIRID, &sr.PatientID, &sr.EncounterID, &sr.RequesterID, &sr.PerformerID,
		&sr.Status, &sr.Intent, &sr.Priority, &sr.CategoryCode, &sr.CategoryDisplay,
		&sr.CodeSystem, &sr.CodeValue, &sr.CodeDisplay,
		&sr.OrderDetailCode, &sr.OrderDetailDisplay, &sr.QuantityValue, &sr.QuantityUnit,
		&sr.OccurrenceDatetime, &sr.OccurrenceStart, &sr.OccurrenceEnd, &sr.AuthoredOn,
		&sr.ReasonCode, &sr.ReasonDisplay, &sr.ReasonConditionID,
		&sr.SpecimenRequirement, &sr.BodySiteCode, &sr.BodySiteDisplay,
		&sr.Note, &sr.PatientInstruction, &sr.CreatedAt, &sr.UpdatedAt)
	return &sr, err
}

func (r *serviceRequestRepoPG) Create(ctx context.Context, sr *ServiceRequest) error {
	sr.ID = uuid.New()
	if sr.FHIRID == "" {
		sr.FHIRID = sr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO service_request (id, fhir_id, patient_id, encounter_id, requester_id, performer_id,
			status, intent, priority, category_code, category_display,
			code_system, code_value, code_display,
			order_detail_code, order_detail_display, quantity_value, quantity_unit,
			occurrence_datetime, occurrence_start, occurrence_end, authored_on,
			reason_code, reason_display, reason_condition_id,
			specimen_requirement, body_site_code, body_site_display,
			note, patient_instruction)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)`,
		sr.ID, sr.FHIRID, sr.PatientID, sr.EncounterID, sr.RequesterID, sr.PerformerID,
		sr.Status, sr.Intent, sr.Priority, sr.CategoryCode, sr.CategoryDisplay,
		sr.CodeSystem, sr.CodeValue, sr.CodeDisplay,
		sr.OrderDetailCode, sr.OrderDetailDisplay, sr.QuantityValue, sr.QuantityUnit,
		sr.OccurrenceDatetime, sr.OccurrenceStart, sr.OccurrenceEnd, sr.AuthoredOn,
		sr.ReasonCode, sr.ReasonDisplay, sr.ReasonConditionID,
		sr.SpecimenRequirement, sr.BodySiteCode, sr.BodySiteDisplay,
		sr.Note, sr.PatientInstruction)
	return err
}

func (r *serviceRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ServiceRequest, error) {
	return r.scanSR(r.conn(ctx).QueryRow(ctx, `SELECT `+srCols+` FROM service_request WHERE id = $1`, id))
}

func (r *serviceRequestRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ServiceRequest, error) {
	return r.scanSR(r.conn(ctx).QueryRow(ctx, `SELECT `+srCols+` FROM service_request WHERE fhir_id = $1`, fhirID))
}

func (r *serviceRequestRepoPG) Update(ctx context.Context, sr *ServiceRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE service_request SET status=$2, intent=$3, priority=$4,
			performer_id=$5, occurrence_datetime=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		sr.ID, sr.Status, sr.Intent, sr.Priority,
		sr.PerformerID, sr.OccurrenceDatetime, sr.Note)
	return err
}

func (r *serviceRequestRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM service_request WHERE id = $1`, id)
	return err
}

func (r *serviceRequestRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ServiceRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM service_request WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+srCols+` FROM service_request WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ServiceRequest
	for rows.Next() {
		sr, err := r.scanSR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sr)
	}
	return items, total, nil
}

func (r *serviceRequestRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ServiceRequest, int, error) {
	query := `SELECT ` + srCols + ` FROM service_request WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM service_request WHERE 1=1`
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
	if p, ok := params["category"]; ok {
		query += fmt.Sprintf(` AND category_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND category_code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["code"]; ok {
		query += fmt.Sprintf(` AND code_value = $%d`, idx)
		countQuery += fmt.Sprintf(` AND code_value = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["intent"]; ok {
		query += fmt.Sprintf(` AND intent = $%d`, idx)
		countQuery += fmt.Sprintf(` AND intent = $%d`, idx)
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
	var items []*ServiceRequest
	for rows.Next() {
		sr, err := r.scanSR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sr)
	}
	return items, total, nil
}

// =========== Specimen Repository ===========

type specimenRepoPG struct{ pool *pgxpool.Pool }

func NewSpecimenRepoPG(pool *pgxpool.Pool) SpecimenRepository { return &specimenRepoPG{pool: pool} }

func (r *specimenRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const spCols = `id, fhir_id, patient_id, accession_id, status,
	type_code, type_display, received_time,
	collection_collector, collection_datetime, collection_quantity, collection_unit,
	collection_method, collection_body_site,
	processing_description, processing_procedure, processing_datetime,
	container_description, container_type,
	condition_code, condition_display, note, created_at, updated_at`

func (r *specimenRepoPG) scanSpecimen(row pgx.Row) (*Specimen, error) {
	var sp Specimen
	err := row.Scan(&sp.ID, &sp.FHIRID, &sp.PatientID, &sp.AccessionID, &sp.Status,
		&sp.TypeCode, &sp.TypeDisplay, &sp.ReceivedTime,
		&sp.CollectionCollector, &sp.CollectionDatetime, &sp.CollectionQuantity, &sp.CollectionUnit,
		&sp.CollectionMethod, &sp.CollectionBodySite,
		&sp.ProcessingDesc, &sp.ProcessingProcedure, &sp.ProcessingDatetime,
		&sp.ContainerDesc, &sp.ContainerType,
		&sp.ConditionCode, &sp.ConditionDisplay, &sp.Note, &sp.CreatedAt, &sp.UpdatedAt)
	return &sp, err
}

func (r *specimenRepoPG) Create(ctx context.Context, sp *Specimen) error {
	sp.ID = uuid.New()
	if sp.FHIRID == "" {
		sp.FHIRID = sp.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO specimen (id, fhir_id, patient_id, accession_id, status,
			type_code, type_display, received_time,
			collection_collector, collection_datetime, collection_quantity, collection_unit,
			collection_method, collection_body_site,
			processing_description, processing_procedure, processing_datetime,
			container_description, container_type,
			condition_code, condition_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		sp.ID, sp.FHIRID, sp.PatientID, sp.AccessionID, sp.Status,
		sp.TypeCode, sp.TypeDisplay, sp.ReceivedTime,
		sp.CollectionCollector, sp.CollectionDatetime, sp.CollectionQuantity, sp.CollectionUnit,
		sp.CollectionMethod, sp.CollectionBodySite,
		sp.ProcessingDesc, sp.ProcessingProcedure, sp.ProcessingDatetime,
		sp.ContainerDesc, sp.ContainerType,
		sp.ConditionCode, sp.ConditionDisplay, sp.Note)
	return err
}

func (r *specimenRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Specimen, error) {
	return r.scanSpecimen(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM specimen WHERE id = $1`, id))
}

func (r *specimenRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Specimen, error) {
	return r.scanSpecimen(r.conn(ctx).QueryRow(ctx, `SELECT `+spCols+` FROM specimen WHERE fhir_id = $1`, fhirID))
}

func (r *specimenRepoPG) Update(ctx context.Context, sp *Specimen) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE specimen SET status=$2, type_code=$3, type_display=$4,
			condition_code=$5, condition_display=$6, note=$7, updated_at=NOW()
		WHERE id = $1`,
		sp.ID, sp.Status, sp.TypeCode, sp.TypeDisplay,
		sp.ConditionCode, sp.ConditionDisplay, sp.Note)
	return err
}

func (r *specimenRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM specimen WHERE id = $1`, id)
	return err
}

func (r *specimenRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Specimen, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM specimen WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+spCols+` FROM specimen WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Specimen
	for rows.Next() {
		sp, err := r.scanSpecimen(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sp)
	}
	return items, total, nil
}

func (r *specimenRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Specimen, int, error) {
	query := `SELECT ` + spCols + ` FROM specimen WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM specimen WHERE 1=1`
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
	if p, ok := params["type"]; ok {
		query += fmt.Sprintf(` AND type_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND type_code = $%d`, idx)
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
	var items []*Specimen
	for rows.Next() {
		sp, err := r.scanSpecimen(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sp)
	}
	return items, total, nil
}

// =========== DiagnosticReport Repository ===========

type diagnosticReportRepoPG struct{ pool *pgxpool.Pool }

func NewDiagnosticReportRepoPG(pool *pgxpool.Pool) DiagnosticReportRepository {
	return &diagnosticReportRepoPG{pool: pool}
}

func (r *diagnosticReportRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const drCols = `id, fhir_id, patient_id, encounter_id, performer_id,
	status, category_code, category_display,
	code_system, code_value, code_display,
	effective_datetime, effective_start, effective_end, issued,
	specimen_id, conclusion, conclusion_code, conclusion_display,
	presented_form_url, presented_form_type, note, created_at, updated_at`

func (r *diagnosticReportRepoPG) scanDR(row pgx.Row) (*DiagnosticReport, error) {
	var dr DiagnosticReport
	err := row.Scan(&dr.ID, &dr.FHIRID, &dr.PatientID, &dr.EncounterID, &dr.PerformerID,
		&dr.Status, &dr.CategoryCode, &dr.CategoryDisplay,
		&dr.CodeSystem, &dr.CodeValue, &dr.CodeDisplay,
		&dr.EffectiveDatetime, &dr.EffectiveStart, &dr.EffectiveEnd, &dr.Issued,
		&dr.SpecimenID, &dr.Conclusion, &dr.ConclusionCode, &dr.ConclusionDisplay,
		&dr.PresentedFormURL, &dr.PresentedFormType, &dr.Note, &dr.CreatedAt, &dr.UpdatedAt)
	return &dr, err
}

func (r *diagnosticReportRepoPG) Create(ctx context.Context, dr *DiagnosticReport) error {
	dr.ID = uuid.New()
	if dr.FHIRID == "" {
		dr.FHIRID = dr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO diagnostic_report (id, fhir_id, patient_id, encounter_id, performer_id,
			status, category_code, category_display,
			code_system, code_value, code_display,
			effective_datetime, effective_start, effective_end,
			specimen_id, conclusion, conclusion_code, conclusion_display,
			presented_form_url, presented_form_type, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		dr.ID, dr.FHIRID, dr.PatientID, dr.EncounterID, dr.PerformerID,
		dr.Status, dr.CategoryCode, dr.CategoryDisplay,
		dr.CodeSystem, dr.CodeValue, dr.CodeDisplay,
		dr.EffectiveDatetime, dr.EffectiveStart, dr.EffectiveEnd,
		dr.SpecimenID, dr.Conclusion, dr.ConclusionCode, dr.ConclusionDisplay,
		dr.PresentedFormURL, dr.PresentedFormType, dr.Note)
	return err
}

func (r *diagnosticReportRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DiagnosticReport, error) {
	return r.scanDR(r.conn(ctx).QueryRow(ctx, `SELECT `+drCols+` FROM diagnostic_report WHERE id = $1`, id))
}

func (r *diagnosticReportRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DiagnosticReport, error) {
	return r.scanDR(r.conn(ctx).QueryRow(ctx, `SELECT `+drCols+` FROM diagnostic_report WHERE fhir_id = $1`, fhirID))
}

func (r *diagnosticReportRepoPG) Update(ctx context.Context, dr *DiagnosticReport) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE diagnostic_report SET status=$2, conclusion=$3, conclusion_code=$4,
			conclusion_display=$5, note=$6, updated_at=NOW()
		WHERE id = $1`,
		dr.ID, dr.Status, dr.Conclusion, dr.ConclusionCode,
		dr.ConclusionDisplay, dr.Note)
	return err
}

func (r *diagnosticReportRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM diagnostic_report WHERE id = $1`, id)
	return err
}

func (r *diagnosticReportRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*DiagnosticReport, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM diagnostic_report WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+drCols+` FROM diagnostic_report WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DiagnosticReport
	for rows.Next() {
		dr, err := r.scanDR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, dr)
	}
	return items, total, nil
}

func (r *diagnosticReportRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DiagnosticReport, int, error) {
	query := `SELECT ` + drCols + ` FROM diagnostic_report WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM diagnostic_report WHERE 1=1`
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
	if p, ok := params["category"]; ok {
		query += fmt.Sprintf(` AND category_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND category_code = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if p, ok := params["code"]; ok {
		query += fmt.Sprintf(` AND code_value = $%d`, idx)
		countQuery += fmt.Sprintf(` AND code_value = $%d`, idx)
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
	var items []*DiagnosticReport
	for rows.Next() {
		dr, err := r.scanDR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, dr)
	}
	return items, total, nil
}

func (r *diagnosticReportRepoPG) AddResult(ctx context.Context, reportID uuid.UUID, observationID uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO diagnostic_report_result (diagnostic_report_id, observation_id)
		VALUES ($1, $2)`, reportID, observationID)
	return err
}

func (r *diagnosticReportRepoPG) GetResults(ctx context.Context, reportID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT observation_id FROM diagnostic_report_result WHERE diagnostic_report_id = $1`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *diagnosticReportRepoPG) RemoveResult(ctx context.Context, reportID uuid.UUID, observationID uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `
		DELETE FROM diagnostic_report_result WHERE diagnostic_report_id = $1 AND observation_id = $2`,
		reportID, observationID)
	return err
}

// =========== ImagingStudy Repository ===========

type imagingStudyRepoPG struct{ pool *pgxpool.Pool }

func NewImagingStudyRepoPG(pool *pgxpool.Pool) ImagingStudyRepository {
	return &imagingStudyRepoPG{pool: pool}
}

func (r *imagingStudyRepoPG) conn(ctx context.Context) queryable {
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const isCols = `id, fhir_id, patient_id, encounter_id, referrer_id,
	status, modality_code, modality_display, study_uid,
	number_of_series, number_of_instances, description, started,
	endpoint, reason_code, reason_display, note, created_at, updated_at`

func (r *imagingStudyRepoPG) scanIS(row pgx.Row) (*ImagingStudy, error) {
	var is ImagingStudy
	err := row.Scan(&is.ID, &is.FHIRID, &is.PatientID, &is.EncounterID, &is.ReferrerID,
		&is.Status, &is.ModalityCode, &is.ModalityDisplay, &is.StudyUID,
		&is.NumberOfSeries, &is.NumberOfInstances, &is.Description, &is.Started,
		&is.Endpoint, &is.ReasonCode, &is.ReasonDisplay, &is.Note, &is.CreatedAt, &is.UpdatedAt)
	return &is, err
}

func (r *imagingStudyRepoPG) Create(ctx context.Context, is *ImagingStudy) error {
	is.ID = uuid.New()
	if is.FHIRID == "" {
		is.FHIRID = is.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO imaging_study (id, fhir_id, patient_id, encounter_id, referrer_id,
			status, modality_code, modality_display, study_uid,
			number_of_series, number_of_instances, description, started,
			endpoint, reason_code, reason_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		is.ID, is.FHIRID, is.PatientID, is.EncounterID, is.ReferrerID,
		is.Status, is.ModalityCode, is.ModalityDisplay, is.StudyUID,
		is.NumberOfSeries, is.NumberOfInstances, is.Description, is.Started,
		is.Endpoint, is.ReasonCode, is.ReasonDisplay, is.Note)
	return err
}

func (r *imagingStudyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ImagingStudy, error) {
	return r.scanIS(r.conn(ctx).QueryRow(ctx, `SELECT `+isCols+` FROM imaging_study WHERE id = $1`, id))
}

func (r *imagingStudyRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ImagingStudy, error) {
	return r.scanIS(r.conn(ctx).QueryRow(ctx, `SELECT `+isCols+` FROM imaging_study WHERE fhir_id = $1`, fhirID))
}

func (r *imagingStudyRepoPG) Update(ctx context.Context, is *ImagingStudy) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE imaging_study SET status=$2, number_of_series=$3, number_of_instances=$4,
			description=$5, note=$6, updated_at=NOW()
		WHERE id = $1`,
		is.ID, is.Status, is.NumberOfSeries, is.NumberOfInstances,
		is.Description, is.Note)
	return err
}

func (r *imagingStudyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM imaging_study WHERE id = $1`, id)
	return err
}

func (r *imagingStudyRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImagingStudy, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM imaging_study WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+isCols+` FROM imaging_study WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ImagingStudy
	for rows.Next() {
		is, err := r.scanIS(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, is)
	}
	return items, total, nil
}

func (r *imagingStudyRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImagingStudy, int, error) {
	query := `SELECT ` + isCols + ` FROM imaging_study WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM imaging_study WHERE 1=1`
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
	if p, ok := params["modality"]; ok {
		query += fmt.Sprintf(` AND modality_code = $%d`, idx)
		countQuery += fmt.Sprintf(` AND modality_code = $%d`, idx)
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
	var items []*ImagingStudy
	for rows.Next() {
		is, err := r.scanIS(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, is)
	}
	return items, total, nil
}
