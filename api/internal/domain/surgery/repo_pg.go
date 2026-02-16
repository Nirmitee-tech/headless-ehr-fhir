package surgery

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

// =========== OR Room Repository ===========

type orRoomRepoPG struct{ pool *pgxpool.Pool }

func NewORRoomRepoPG(pool *pgxpool.Pool) ORRoomRepository { return &orRoomRepoPG{pool: pool} }

func (r *orRoomRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const orRoomCols = `id, name, location_id, status, room_type, equipment, is_active, decontaminated_at, note, created_at, updated_at`

func (r *orRoomRepoPG) scanORRoom(row pgx.Row) (*ORRoom, error) {
	var o ORRoom
	err := row.Scan(&o.ID, &o.Name, &o.LocationID, &o.Status, &o.RoomType, &o.Equipment,
		&o.IsActive, &o.DecontaminatedAt, &o.Note, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *orRoomRepoPG) Create(ctx context.Context, o *ORRoom) error {
	o.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO or_room (id, name, location_id, status, room_type, equipment, is_active, decontaminated_at, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		o.ID, o.Name, o.LocationID, o.Status, o.RoomType, o.Equipment, o.IsActive, o.DecontaminatedAt, o.Note)
	return err
}

func (r *orRoomRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ORRoom, error) {
	return r.scanORRoom(r.conn(ctx).QueryRow(ctx, `SELECT `+orRoomCols+` FROM or_room WHERE id = $1`, id))
}

func (r *orRoomRepoPG) Update(ctx context.Context, o *ORRoom) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE or_room SET name=$2, status=$3, room_type=$4, equipment=$5, is_active=$6,
			decontaminated_at=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		o.ID, o.Name, o.Status, o.RoomType, o.Equipment, o.IsActive, o.DecontaminatedAt, o.Note)
	return err
}

func (r *orRoomRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM or_room WHERE id = $1`, id)
	return err
}

func (r *orRoomRepoPG) List(ctx context.Context, limit, offset int) ([]*ORRoom, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM or_room`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+orRoomCols+` FROM or_room ORDER BY name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ORRoom
	for rows.Next() {
		o, err := r.scanORRoom(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

var orRoomSearchParams = map[string]fhir.SearchParamConfig{
	"status":    {Type: fhir.SearchParamToken, Column: "status"},
	"is_active": {Type: fhir.SearchParamToken, Column: "is_active"},
}

func (r *orRoomRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ORRoom, int, error) {
	qb := fhir.NewSearchQuery("or_room", orRoomCols)
	qb.ApplyParams(params, orRoomSearchParams)
	qb.OrderBy("name")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ORRoom
	for rows.Next() {
		o, err := r.scanORRoom(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	return items, total, nil
}

// =========== Surgical Case Repository ===========

type surgicalCaseRepoPG struct{ pool *pgxpool.Pool }

func NewSurgicalCaseRepoPG(pool *pgxpool.Pool) SurgicalCaseRepository {
	return &surgicalCaseRepoPG{pool: pool}
}

func (r *surgicalCaseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const caseCols = `id, patient_id, encounter_id, primary_surgeon_id, anesthesiologist_id, or_room_id,
	status, case_class, asa_class, wound_class,
	scheduled_date, scheduled_start, scheduled_end, actual_start, actual_end,
	anesthesia_type, laterality, pre_op_diagnosis, post_op_diagnosis,
	cancel_reason, note, created_at, updated_at`

func (r *surgicalCaseRepoPG) scanCase(row pgx.Row) (*SurgicalCase, error) {
	var sc SurgicalCase
	err := row.Scan(&sc.ID, &sc.PatientID, &sc.EncounterID, &sc.PrimarySurgeonID, &sc.AnesthesiologistID, &sc.ORRoomID,
		&sc.Status, &sc.CaseClass, &sc.ASAClass, &sc.WoundClass,
		&sc.ScheduledDate, &sc.ScheduledStart, &sc.ScheduledEnd, &sc.ActualStart, &sc.ActualEnd,
		&sc.AnesthesiaType, &sc.Laterality, &sc.PreOpDiagnosis, &sc.PostOpDiagnosis,
		&sc.CancelReason, &sc.Note, &sc.CreatedAt, &sc.UpdatedAt)
	return &sc, err
}

func (r *surgicalCaseRepoPG) Create(ctx context.Context, sc *SurgicalCase) error {
	sc.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_case (id, patient_id, encounter_id, primary_surgeon_id, anesthesiologist_id, or_room_id,
			status, case_class, asa_class, wound_class,
			scheduled_date, scheduled_start, scheduled_end,
			anesthesia_type, laterality, pre_op_diagnosis, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		sc.ID, sc.PatientID, sc.EncounterID, sc.PrimarySurgeonID, sc.AnesthesiologistID, sc.ORRoomID,
		sc.Status, sc.CaseClass, sc.ASAClass, sc.WoundClass,
		sc.ScheduledDate, sc.ScheduledStart, sc.ScheduledEnd,
		sc.AnesthesiaType, sc.Laterality, sc.PreOpDiagnosis, sc.Note)
	return err
}

func (r *surgicalCaseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SurgicalCase, error) {
	return r.scanCase(r.conn(ctx).QueryRow(ctx, `SELECT `+caseCols+` FROM surgical_case WHERE id = $1`, id))
}

func (r *surgicalCaseRepoPG) Update(ctx context.Context, sc *SurgicalCase) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE surgical_case SET status=$2, or_room_id=$3, actual_start=$4, actual_end=$5,
			post_op_diagnosis=$6, wound_class=$7, cancel_reason=$8, note=$9, updated_at=NOW()
		WHERE id = $1`,
		sc.ID, sc.Status, sc.ORRoomID, sc.ActualStart, sc.ActualEnd,
		sc.PostOpDiagnosis, sc.WoundClass, sc.CancelReason, sc.Note)
	return err
}

func (r *surgicalCaseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM surgical_case WHERE id = $1`, id)
	return err
}

func (r *surgicalCaseRepoPG) List(ctx context.Context, limit, offset int) ([]*SurgicalCase, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM surgical_case`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+caseCols+` FROM surgical_case ORDER BY scheduled_date DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SurgicalCase
	for rows.Next() {
		sc, err := r.scanCase(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sc)
	}
	return items, total, nil
}

func (r *surgicalCaseRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*SurgicalCase, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM surgical_case WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+caseCols+` FROM surgical_case WHERE patient_id = $1 ORDER BY scheduled_date DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SurgicalCase
	for rows.Next() {
		sc, err := r.scanCase(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sc)
	}
	return items, total, nil
}

var surgicalCaseSearchParams = map[string]fhir.SearchParamConfig{
	"patient_id": {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":     {Type: fhir.SearchParamToken, Column: "status"},
	"surgeon_id": {Type: fhir.SearchParamReference, Column: "primary_surgeon_id"},
	"or_room_id": {Type: fhir.SearchParamReference, Column: "or_room_id"},
}

func (r *surgicalCaseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SurgicalCase, int, error) {
	qb := fhir.NewSearchQuery("surgical_case", caseCols)
	qb.ApplyParams(params, surgicalCaseSearchParams)
	qb.OrderBy("scheduled_date DESC")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SurgicalCase
	for rows.Next() {
		sc, err := r.scanCase(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, sc)
	}
	return items, total, nil
}

// -- Sub-resources: Procedures --

func (r *surgicalCaseRepoPG) AddProcedure(ctx context.Context, p *SurgicalCaseProcedure) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_case_procedure (id, surgical_case_id, procedure_code, procedure_display,
			code_system, cpt_code, is_primary, body_site_code, body_site_display, sequence)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.SurgicalCaseID, p.ProcedureCode, p.ProcedureDisplay,
		p.CodeSystem, p.CPTCode, p.IsPrimary, p.BodySiteCode, p.BodySiteDisplay, p.Sequence)
	return err
}

func (r *surgicalCaseRepoPG) GetProcedures(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCaseProcedure, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, surgical_case_id, procedure_code, procedure_display,
			code_system, cpt_code, is_primary, body_site_code, body_site_display, sequence
		FROM surgical_case_procedure WHERE surgical_case_id = $1 ORDER BY sequence`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SurgicalCaseProcedure
	for rows.Next() {
		var p SurgicalCaseProcedure
		if err := rows.Scan(&p.ID, &p.SurgicalCaseID, &p.ProcedureCode, &p.ProcedureDisplay,
			&p.CodeSystem, &p.CPTCode, &p.IsPrimary, &p.BodySiteCode, &p.BodySiteDisplay, &p.Sequence); err != nil {
			return nil, err
		}
		items = append(items, &p)
	}
	return items, nil
}

func (r *surgicalCaseRepoPG) RemoveProcedure(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM surgical_case_procedure WHERE id = $1`, id)
	return err
}

// -- Sub-resources: Team --

func (r *surgicalCaseRepoPG) AddTeamMember(ctx context.Context, t *SurgicalCaseTeam) error {
	t.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_case_team (id, surgical_case_id, practitioner_id, role, role_display, start_time, end_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		t.ID, t.SurgicalCaseID, t.PractitionerID, t.Role, t.RoleDisplay, t.StartTime, t.EndTime)
	return err
}

func (r *surgicalCaseRepoPG) GetTeamMembers(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCaseTeam, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, surgical_case_id, practitioner_id, role, role_display, start_time, end_time
		FROM surgical_case_team WHERE surgical_case_id = $1`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SurgicalCaseTeam
	for rows.Next() {
		var t SurgicalCaseTeam
		if err := rows.Scan(&t.ID, &t.SurgicalCaseID, &t.PractitionerID, &t.Role, &t.RoleDisplay, &t.StartTime, &t.EndTime); err != nil {
			return nil, err
		}
		items = append(items, &t)
	}
	return items, nil
}

func (r *surgicalCaseRepoPG) RemoveTeamMember(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM surgical_case_team WHERE id = $1`, id)
	return err
}

// -- Sub-resources: Time Events --

func (r *surgicalCaseRepoPG) AddTimeEvent(ctx context.Context, e *SurgicalTimeEvent) error {
	e.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_time_event (id, surgical_case_id, event_type, event_time, recorded_by, note)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		e.ID, e.SurgicalCaseID, e.EventType, e.EventTime, e.RecordedBy, e.Note)
	return err
}

func (r *surgicalCaseRepoPG) GetTimeEvents(ctx context.Context, caseID uuid.UUID) ([]*SurgicalTimeEvent, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, surgical_case_id, event_type, event_time, recorded_by, note
		FROM surgical_time_event WHERE surgical_case_id = $1 ORDER BY event_time`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SurgicalTimeEvent
	for rows.Next() {
		var e SurgicalTimeEvent
		if err := rows.Scan(&e.ID, &e.SurgicalCaseID, &e.EventType, &e.EventTime, &e.RecordedBy, &e.Note); err != nil {
			return nil, err
		}
		items = append(items, &e)
	}
	return items, nil
}

// -- Sub-resources: Counts --

func (r *surgicalCaseRepoPG) AddCount(ctx context.Context, c *SurgicalCount) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_count (id, surgical_case_id, count_type, item_name,
			expected_count, actual_count, is_correct, counted_by, verified_by, count_time, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		c.ID, c.SurgicalCaseID, c.CountType, c.ItemName,
		c.ExpectedCount, c.ActualCount, c.IsCorrect, c.CountedBy, c.VerifiedBy, c.CountTime, c.Note)
	return err
}

func (r *surgicalCaseRepoPG) GetCounts(ctx context.Context, caseID uuid.UUID) ([]*SurgicalCount, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, surgical_case_id, count_type, item_name,
			expected_count, actual_count, is_correct, counted_by, verified_by, count_time, note
		FROM surgical_count WHERE surgical_case_id = $1 ORDER BY count_time`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SurgicalCount
	for rows.Next() {
		var c SurgicalCount
		if err := rows.Scan(&c.ID, &c.SurgicalCaseID, &c.CountType, &c.ItemName,
			&c.ExpectedCount, &c.ActualCount, &c.IsCorrect, &c.CountedBy, &c.VerifiedBy, &c.CountTime, &c.Note); err != nil {
			return nil, err
		}
		items = append(items, &c)
	}
	return items, nil
}

// -- Sub-resources: Supplies --

func (r *surgicalCaseRepoPG) AddSupply(ctx context.Context, su *SurgicalSupplyUsed) error {
	su.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_supply_used (id, surgical_case_id, supply_name, supply_code,
			quantity, unit_of_measure, lot_number, recorded_by, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		su.ID, su.SurgicalCaseID, su.SupplyName, su.SupplyCode,
		su.Quantity, su.UnitOfMeasure, su.LotNumber, su.RecordedBy, su.Note)
	return err
}

func (r *surgicalCaseRepoPG) GetSupplies(ctx context.Context, caseID uuid.UUID) ([]*SurgicalSupplyUsed, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, surgical_case_id, supply_name, supply_code,
			quantity, unit_of_measure, lot_number, recorded_by, note
		FROM surgical_supply_used WHERE surgical_case_id = $1`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*SurgicalSupplyUsed
	for rows.Next() {
		var su SurgicalSupplyUsed
		if err := rows.Scan(&su.ID, &su.SurgicalCaseID, &su.SupplyName, &su.SupplyCode,
			&su.Quantity, &su.UnitOfMeasure, &su.LotNumber, &su.RecordedBy, &su.Note); err != nil {
			return nil, err
		}
		items = append(items, &su)
	}
	return items, nil
}

// =========== Preference Card Repository ===========

type preferenceCardRepoPG struct{ pool *pgxpool.Pool }

func NewPreferenceCardRepoPG(pool *pgxpool.Pool) PreferenceCardRepository {
	return &preferenceCardRepoPG{pool: pool}
}

func (r *preferenceCardRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const prefCardCols = `id, surgeon_id, procedure_code, procedure_display,
	glove_size_l, glove_size_r, gown, skin_prep, position,
	instruments, supplies, sutures, dressings, special_equipment,
	note, is_active, created_at, updated_at`

func (r *preferenceCardRepoPG) scanPrefCard(row pgx.Row) (*SurgicalPreferenceCard, error) {
	var pc SurgicalPreferenceCard
	err := row.Scan(&pc.ID, &pc.SurgeonID, &pc.ProcedureCode, &pc.ProcedureDisplay,
		&pc.GloveSizeL, &pc.GloveSizeR, &pc.Gown, &pc.SkinPrep, &pc.Position,
		&pc.Instruments, &pc.Supplies, &pc.Sutures, &pc.Dressings, &pc.SpecialEquipment,
		&pc.Note, &pc.IsActive, &pc.CreatedAt, &pc.UpdatedAt)
	return &pc, err
}

func (r *preferenceCardRepoPG) Create(ctx context.Context, pc *SurgicalPreferenceCard) error {
	pc.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO surgical_preference_card (id, surgeon_id, procedure_code, procedure_display,
			glove_size_l, glove_size_r, gown, skin_prep, position,
			instruments, supplies, sutures, dressings, special_equipment, note, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		pc.ID, pc.SurgeonID, pc.ProcedureCode, pc.ProcedureDisplay,
		pc.GloveSizeL, pc.GloveSizeR, pc.Gown, pc.SkinPrep, pc.Position,
		pc.Instruments, pc.Supplies, pc.Sutures, pc.Dressings, pc.SpecialEquipment, pc.Note, pc.IsActive)
	return err
}

func (r *preferenceCardRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SurgicalPreferenceCard, error) {
	return r.scanPrefCard(r.conn(ctx).QueryRow(ctx, `SELECT `+prefCardCols+` FROM surgical_preference_card WHERE id = $1`, id))
}

func (r *preferenceCardRepoPG) Update(ctx context.Context, pc *SurgicalPreferenceCard) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE surgical_preference_card SET procedure_code=$2, procedure_display=$3,
			glove_size_l=$4, glove_size_r=$5, gown=$6, skin_prep=$7, position=$8,
			instruments=$9, supplies=$10, sutures=$11, dressings=$12, special_equipment=$13,
			note=$14, is_active=$15, updated_at=NOW()
		WHERE id = $1`,
		pc.ID, pc.ProcedureCode, pc.ProcedureDisplay,
		pc.GloveSizeL, pc.GloveSizeR, pc.Gown, pc.SkinPrep, pc.Position,
		pc.Instruments, pc.Supplies, pc.Sutures, pc.Dressings, pc.SpecialEquipment,
		pc.Note, pc.IsActive)
	return err
}

func (r *preferenceCardRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM surgical_preference_card WHERE id = $1`, id)
	return err
}

func (r *preferenceCardRepoPG) ListBySurgeon(ctx context.Context, surgeonID uuid.UUID, limit, offset int) ([]*SurgicalPreferenceCard, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM surgical_preference_card WHERE surgeon_id = $1`, surgeonID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+prefCardCols+` FROM surgical_preference_card WHERE surgeon_id = $1 ORDER BY procedure_display LIMIT $2 OFFSET $3`, surgeonID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SurgicalPreferenceCard
	for rows.Next() {
		pc, err := r.scanPrefCard(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, pc)
	}
	return items, total, nil
}

var prefCardSearchParams = map[string]fhir.SearchParamConfig{
	"surgeon_id":     {Type: fhir.SearchParamReference, Column: "surgeon_id"},
	"procedure_code": {Type: fhir.SearchParamToken, Column: "procedure_code"},
}

func (r *preferenceCardRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SurgicalPreferenceCard, int, error) {
	qb := fhir.NewSearchQuery("surgical_preference_card", prefCardCols)
	qb.ApplyParams(params, prefCardSearchParams)
	qb.OrderBy("procedure_display")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*SurgicalPreferenceCard
	for rows.Next() {
		pc, err := r.scanPrefCard(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, pc)
	}
	return items, total, nil
}

// =========== Implant Log Repository ===========

type implantLogRepoPG struct{ pool *pgxpool.Pool }

func NewImplantLogRepoPG(pool *pgxpool.Pool) ImplantLogRepository {
	return &implantLogRepoPG{pool: pool}
}

func (r *implantLogRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const implantCols = `id, surgical_case_id, patient_id, device_id, implant_type,
	manufacturer, model_number, serial_number, lot_number, expiration_date,
	body_site_code, body_site_display, implanted_by, implant_date, explant_date,
	note, created_at, updated_at`

func (r *implantLogRepoPG) scanImplant(row pgx.Row) (*ImplantLog, error) {
	var il ImplantLog
	err := row.Scan(&il.ID, &il.SurgicalCaseID, &il.PatientID, &il.DeviceID, &il.ImplantType,
		&il.Manufacturer, &il.ModelNumber, &il.SerialNumber, &il.LotNumber, &il.ExpirationDate,
		&il.BodySiteCode, &il.BodySiteDisplay, &il.ImplantedBy, &il.ImplantDate, &il.ExplantDate,
		&il.Note, &il.CreatedAt, &il.UpdatedAt)
	return &il, err
}

func (r *implantLogRepoPG) Create(ctx context.Context, il *ImplantLog) error {
	il.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO implant_log (id, surgical_case_id, patient_id, device_id, implant_type,
			manufacturer, model_number, serial_number, lot_number, expiration_date,
			body_site_code, body_site_display, implanted_by, implant_date, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		il.ID, il.SurgicalCaseID, il.PatientID, il.DeviceID, il.ImplantType,
		il.Manufacturer, il.ModelNumber, il.SerialNumber, il.LotNumber, il.ExpirationDate,
		il.BodySiteCode, il.BodySiteDisplay, il.ImplantedBy, il.ImplantDate, il.Note)
	return err
}

func (r *implantLogRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ImplantLog, error) {
	return r.scanImplant(r.conn(ctx).QueryRow(ctx, `SELECT `+implantCols+` FROM implant_log WHERE id = $1`, id))
}

func (r *implantLogRepoPG) Update(ctx context.Context, il *ImplantLog) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE implant_log SET implant_type=$2, manufacturer=$3, model_number=$4,
			serial_number=$5, lot_number=$6, body_site_code=$7, body_site_display=$8,
			explant_date=$9, note=$10, updated_at=NOW()
		WHERE id = $1`,
		il.ID, il.ImplantType, il.Manufacturer, il.ModelNumber,
		il.SerialNumber, il.LotNumber, il.BodySiteCode, il.BodySiteDisplay,
		il.ExplantDate, il.Note)
	return err
}

func (r *implantLogRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM implant_log WHERE id = $1`, id)
	return err
}

func (r *implantLogRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM implant_log WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+implantCols+` FROM implant_log WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ImplantLog
	for rows.Next() {
		il, err := r.scanImplant(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, il)
	}
	return items, total, nil
}

func (r *implantLogRepoPG) ListByCase(ctx context.Context, caseID uuid.UUID, limit, offset int) ([]*ImplantLog, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM implant_log WHERE surgical_case_id = $1`, caseID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+implantCols+` FROM implant_log WHERE surgical_case_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, caseID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ImplantLog
	for rows.Next() {
		il, err := r.scanImplant(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, il)
	}
	return items, total, nil
}

var implantLogSearchParams = map[string]fhir.SearchParamConfig{
	"patient_id":       {Type: fhir.SearchParamReference, Column: "patient_id"},
	"surgical_case_id": {Type: fhir.SearchParamReference, Column: "surgical_case_id"},
	"implant_type":     {Type: fhir.SearchParamToken, Column: "implant_type"},
}

func (r *implantLogRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImplantLog, int, error) {
	qb := fhir.NewSearchQuery("implant_log", implantCols)
	qb.ApplyParams(params, implantLogSearchParams)
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
	var items []*ImplantLog
	for rows.Next() {
		il, err := r.scanImplant(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, il)
	}
	return items, total, nil
}
