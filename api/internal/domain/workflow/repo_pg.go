package workflow

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

// =========== ActivityDefinition Repository ===========

type activityDefinitionRepoPG struct{ pool *pgxpool.Pool }

func NewActivityDefinitionRepoPG(pool *pgxpool.Pool) ActivityDefinitionRepository {
	return &activityDefinitionRepoPG{pool: pool}
}

func (r *activityDefinitionRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const adCols = `id, fhir_id, url, status, name, title, description, purpose,
	kind, code_code, code_display, code_system, intent, priority,
	do_not_perform, timing_description, location_id,
	quantity_value, quantity_unit, dosage_text, publisher,
	effective_start, effective_end, approval_date, last_review_date,
	created_at, updated_at`

func (r *activityDefinitionRepoPG) scanAD(row pgx.Row) (*ActivityDefinition, error) {
	var a ActivityDefinition
	err := row.Scan(&a.ID, &a.FHIRID, &a.URL, &a.Status, &a.Name, &a.Title,
		&a.Description, &a.Purpose, &a.Kind, &a.CodeCode, &a.CodeDisplay,
		&a.CodeSystem, &a.Intent, &a.Priority, &a.DoNotPerform,
		&a.TimingDescription, &a.LocationID, &a.QuantityValue, &a.QuantityUnit,
		&a.DosageText, &a.Publisher, &a.EffectiveStart, &a.EffectiveEnd,
		&a.ApprovalDate, &a.LastReviewDate, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *activityDefinitionRepoPG) Create(ctx context.Context, a *ActivityDefinition) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO activity_definition (id, fhir_id, url, status, name, title,
			description, purpose, kind, code_code, code_display, code_system,
			intent, priority, do_not_perform, timing_description, location_id,
			quantity_value, quantity_unit, dosage_text, publisher,
			effective_start, effective_end, approval_date, last_review_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		a.ID, a.FHIRID, a.URL, a.Status, a.Name, a.Title,
		a.Description, a.Purpose, a.Kind, a.CodeCode, a.CodeDisplay, a.CodeSystem,
		a.Intent, a.Priority, a.DoNotPerform, a.TimingDescription, a.LocationID,
		a.QuantityValue, a.QuantityUnit, a.DosageText, a.Publisher,
		a.EffectiveStart, a.EffectiveEnd, a.ApprovalDate, a.LastReviewDate)
	return err
}

func (r *activityDefinitionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ActivityDefinition, error) {
	return r.scanAD(r.conn(ctx).QueryRow(ctx, `SELECT `+adCols+` FROM activity_definition WHERE id = $1`, id))
}

func (r *activityDefinitionRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*ActivityDefinition, error) {
	return r.scanAD(r.conn(ctx).QueryRow(ctx, `SELECT `+adCols+` FROM activity_definition WHERE fhir_id = $1`, fhirID))
}

func (r *activityDefinitionRepoPG) Update(ctx context.Context, a *ActivityDefinition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE activity_definition SET status=$2, name=$3, title=$4,
			description=$5, purpose=$6, kind=$7, intent=$8, priority=$9,
			do_not_perform=$10, publisher=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.Status, a.Name, a.Title,
		a.Description, a.Purpose, a.Kind, a.Intent, a.Priority,
		a.DoNotPerform, a.Publisher, a.DosageText)
	return err
}

func (r *activityDefinitionRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM activity_definition WHERE id = $1`, id)
	return err
}

func (r *activityDefinitionRepoPG) List(ctx context.Context, limit, offset int) ([]*ActivityDefinition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM activity_definition`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+adCols+` FROM activity_definition ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ActivityDefinition
	for rows.Next() {
		a, err := r.scanAD(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

var adSearchParams = map[string]fhir.SearchParamConfig{
	"status": {Type: fhir.SearchParamToken, Column: "status"},
	"name":   {Type: fhir.SearchParamString, Column: "name"},
}

func (r *activityDefinitionRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ActivityDefinition, int, error) {
	qb := fhir.NewSearchQuery("activity_definition", adCols)
	qb.ApplyParams(params, adSearchParams)
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
	var items []*ActivityDefinition
	for rows.Next() {
		a, err := r.scanAD(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== RequestGroup Repository ===========

type requestGroupRepoPG struct{ pool *pgxpool.Pool }

func NewRequestGroupRepoPG(pool *pgxpool.Pool) RequestGroupRepository {
	return &requestGroupRepoPG{pool: pool}
}

func (r *requestGroupRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const rgCols = `id, fhir_id, status, intent, priority, code_code, code_display,
	subject_patient_id, encounter_id, authored_on, author_id,
	reason_code, reason_display, note,
	created_at, updated_at`

func (r *requestGroupRepoPG) scanRG(row pgx.Row) (*RequestGroup, error) {
	var rg RequestGroup
	err := row.Scan(&rg.ID, &rg.FHIRID, &rg.Status, &rg.Intent, &rg.Priority,
		&rg.CodeCode, &rg.CodeDisplay, &rg.SubjectPatientID, &rg.EncounterID,
		&rg.AuthoredOn, &rg.AuthorID, &rg.ReasonCode, &rg.ReasonDisplay,
		&rg.Note, &rg.CreatedAt, &rg.UpdatedAt)
	return &rg, err
}

func (r *requestGroupRepoPG) Create(ctx context.Context, rg *RequestGroup) error {
	rg.ID = uuid.New()
	if rg.FHIRID == "" {
		rg.FHIRID = rg.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO request_group (id, fhir_id, status, intent, priority,
			code_code, code_display, subject_patient_id, encounter_id,
			authored_on, author_id, reason_code, reason_display, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		rg.ID, rg.FHIRID, rg.Status, rg.Intent, rg.Priority,
		rg.CodeCode, rg.CodeDisplay, rg.SubjectPatientID, rg.EncounterID,
		rg.AuthoredOn, rg.AuthorID, rg.ReasonCode, rg.ReasonDisplay, rg.Note)
	return err
}

func (r *requestGroupRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*RequestGroup, error) {
	return r.scanRG(r.conn(ctx).QueryRow(ctx, `SELECT `+rgCols+` FROM request_group WHERE id = $1`, id))
}

func (r *requestGroupRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*RequestGroup, error) {
	return r.scanRG(r.conn(ctx).QueryRow(ctx, `SELECT `+rgCols+` FROM request_group WHERE fhir_id = $1`, fhirID))
}

func (r *requestGroupRepoPG) Update(ctx context.Context, rg *RequestGroup) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE request_group SET status=$2, intent=$3, priority=$4,
			code_code=$5, code_display=$6, reason_code=$7, reason_display=$8,
			note=$9, updated_at=NOW()
		WHERE id = $1`,
		rg.ID, rg.Status, rg.Intent, rg.Priority,
		rg.CodeCode, rg.CodeDisplay, rg.ReasonCode, rg.ReasonDisplay, rg.Note)
	return err
}

func (r *requestGroupRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM request_group WHERE id = $1`, id)
	return err
}

func (r *requestGroupRepoPG) List(ctx context.Context, limit, offset int) ([]*RequestGroup, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM request_group`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+rgCols+` FROM request_group ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*RequestGroup
	for rows.Next() {
		rg, err := r.scanRG(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rg)
	}
	return items, total, nil
}

var rgSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *requestGroupRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RequestGroup, int, error) {
	qb := fhir.NewSearchQuery("request_group", rgCols)
	qb.ApplyParams(params, rgSearchParams)
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
	var items []*RequestGroup
	for rows.Next() {
		rg, err := r.scanRG(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rg)
	}
	return items, total, nil
}

func (r *requestGroupRepoPG) AddAction(ctx context.Context, a *RequestGroupAction) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO request_group_action (id, request_group_id, prefix, title,
			description, priority, resource_reference,
			selection_behavior, required_behavior, precheck_behavior, cardinality_behavior)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		a.ID, a.RequestGroupID, a.Prefix, a.Title,
		a.Description, a.Priority, a.ResourceReference,
		a.SelectionBehavior, a.RequiredBehavior, a.PrecheckBehavior, a.CardinalityBehavior)
	return err
}

func (r *requestGroupRepoPG) GetActions(ctx context.Context, requestGroupID uuid.UUID) ([]*RequestGroupAction, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, request_group_id, prefix, title, description, priority,
			resource_reference, selection_behavior, required_behavior,
			precheck_behavior, cardinality_behavior
		FROM request_group_action WHERE request_group_id = $1 ORDER BY id`, requestGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*RequestGroupAction
	for rows.Next() {
		var a RequestGroupAction
		if err := rows.Scan(&a.ID, &a.RequestGroupID, &a.Prefix, &a.Title,
			&a.Description, &a.Priority, &a.ResourceReference,
			&a.SelectionBehavior, &a.RequiredBehavior,
			&a.PrecheckBehavior, &a.CardinalityBehavior); err != nil {
			return nil, err
		}
		items = append(items, &a)
	}
	return items, nil
}

// =========== GuidanceResponse Repository ===========

type guidanceResponseRepoPG struct{ pool *pgxpool.Pool }

func NewGuidanceResponseRepoPG(pool *pgxpool.Pool) GuidanceResponseRepository {
	return &guidanceResponseRepoPG{pool: pool}
}

func (r *guidanceResponseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const grCols = `id, fhir_id, request_identifier, module_uri, status,
	subject_patient_id, encounter_id, occurrence_date, performer_id,
	reason_code, reason_display, note, result_reference,
	created_at, updated_at`

func (r *guidanceResponseRepoPG) scanGR(row pgx.Row) (*GuidanceResponse, error) {
	var gr GuidanceResponse
	err := row.Scan(&gr.ID, &gr.FHIRID, &gr.RequestIdentifier, &gr.ModuleURI,
		&gr.Status, &gr.SubjectPatientID, &gr.EncounterID, &gr.OccurrenceDate,
		&gr.PerformerID, &gr.ReasonCode, &gr.ReasonDisplay, &gr.Note,
		&gr.ResultReference, &gr.CreatedAt, &gr.UpdatedAt)
	return &gr, err
}

func (r *guidanceResponseRepoPG) Create(ctx context.Context, gr *GuidanceResponse) error {
	gr.ID = uuid.New()
	if gr.FHIRID == "" {
		gr.FHIRID = gr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO guidance_response (id, fhir_id, request_identifier, module_uri,
			status, subject_patient_id, encounter_id, occurrence_date, performer_id,
			reason_code, reason_display, note, result_reference)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		gr.ID, gr.FHIRID, gr.RequestIdentifier, gr.ModuleURI,
		gr.Status, gr.SubjectPatientID, gr.EncounterID, gr.OccurrenceDate,
		gr.PerformerID, gr.ReasonCode, gr.ReasonDisplay, gr.Note, gr.ResultReference)
	return err
}

func (r *guidanceResponseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*GuidanceResponse, error) {
	return r.scanGR(r.conn(ctx).QueryRow(ctx, `SELECT `+grCols+` FROM guidance_response WHERE id = $1`, id))
}

func (r *guidanceResponseRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*GuidanceResponse, error) {
	return r.scanGR(r.conn(ctx).QueryRow(ctx, `SELECT `+grCols+` FROM guidance_response WHERE fhir_id = $1`, fhirID))
}

func (r *guidanceResponseRepoPG) Update(ctx context.Context, gr *GuidanceResponse) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE guidance_response SET status=$2, module_uri=$3,
			reason_code=$4, reason_display=$5, note=$6, result_reference=$7,
			updated_at=NOW()
		WHERE id = $1`,
		gr.ID, gr.Status, gr.ModuleURI,
		gr.ReasonCode, gr.ReasonDisplay, gr.Note, gr.ResultReference)
	return err
}

func (r *guidanceResponseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM guidance_response WHERE id = $1`, id)
	return err
}

func (r *guidanceResponseRepoPG) List(ctx context.Context, limit, offset int) ([]*GuidanceResponse, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM guidance_response`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+grCols+` FROM guidance_response ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*GuidanceResponse
	for rows.Next() {
		gr, err := r.scanGR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, gr)
	}
	return items, total, nil
}

var grSearchParams = map[string]fhir.SearchParamConfig{
	"patient": {Type: fhir.SearchParamReference, Column: "subject_patient_id"},
	"status":  {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *guidanceResponseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*GuidanceResponse, int, error) {
	qb := fhir.NewSearchQuery("guidance_response", grCols)
	qb.ApplyParams(params, grSearchParams)
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
	var items []*GuidanceResponse
	for rows.Next() {
		gr, err := r.scanGR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, gr)
	}
	return items, total, nil
}
