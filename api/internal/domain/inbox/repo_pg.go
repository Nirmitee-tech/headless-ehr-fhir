package inbox

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

// =========== Message Pool Repository ===========

type messagePoolRepoPG struct{ pool *pgxpool.Pool }

func NewMessagePoolRepoPG(pool *pgxpool.Pool) MessagePoolRepository {
	return &messagePoolRepoPG{pool: pool}
}

func (r *messagePoolRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const msgPoolCols = `id, pool_name, pool_type, organization_id, department_id, description,
	is_active, created_at, updated_at`

func (r *messagePoolRepoPG) scanPool(row pgx.Row) (*MessagePool, error) {
	var p MessagePool
	err := row.Scan(&p.ID, &p.PoolName, &p.PoolType, &p.OrganizationID, &p.DepartmentID,
		&p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *messagePoolRepoPG) Create(ctx context.Context, p *MessagePool) error {
	p.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO message_pool (id, pool_name, pool_type, organization_id, department_id,
			description, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.PoolName, p.PoolType, p.OrganizationID, p.DepartmentID,
		p.Description, p.IsActive)
	return err
}

func (r *messagePoolRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*MessagePool, error) {
	return r.scanPool(r.conn(ctx).QueryRow(ctx, `SELECT `+msgPoolCols+` FROM message_pool WHERE id = $1`, id))
}

func (r *messagePoolRepoPG) Update(ctx context.Context, p *MessagePool) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE message_pool SET pool_name=$2, pool_type=$3, organization_id=$4, department_id=$5,
			description=$6, is_active=$7, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.PoolName, p.PoolType, p.OrganizationID, p.DepartmentID,
		p.Description, p.IsActive)
	return err
}

func (r *messagePoolRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM message_pool WHERE id = $1`, id)
	return err
}

func (r *messagePoolRepoPG) List(ctx context.Context, limit, offset int) ([]*MessagePool, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM message_pool`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+msgPoolCols+` FROM message_pool ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*MessagePool
	for rows.Next() {
		p, err := r.scanPool(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, nil
}

// =========== Inbox Message Repository ===========

type inboxMessageRepoPG struct{ pool *pgxpool.Pool }

func NewInboxMessageRepoPG(pool *pgxpool.Pool) InboxMessageRepository {
	return &inboxMessageRepoPG{pool: pool}
}

func (r *inboxMessageRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const inboxMsgCols = `id, message_type, priority, subject, body, patient_id, encounter_id,
	sender_id, recipient_id, pool_id, status, parent_id, thread_id,
	is_urgent, due_date, read_at, completed_at, created_at, updated_at`

func (r *inboxMessageRepoPG) scanMsg(row pgx.Row) (*InboxMessage, error) {
	var m InboxMessage
	err := row.Scan(&m.ID, &m.MessageType, &m.Priority, &m.Subject, &m.Body,
		&m.PatientID, &m.EncounterID, &m.SenderID, &m.RecipientID, &m.PoolID,
		&m.Status, &m.ParentID, &m.ThreadID, &m.IsUrgent, &m.DueDate,
		&m.ReadAt, &m.CompletedAt, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *inboxMessageRepoPG) Create(ctx context.Context, m *InboxMessage) error {
	m.ID = uuid.New()
	if m.ThreadID == nil {
		m.ThreadID = &m.ID
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO inbox_message (id, message_type, priority, subject, body,
			patient_id, encounter_id, sender_id, recipient_id, pool_id,
			status, parent_id, thread_id, is_urgent, due_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		m.ID, m.MessageType, m.Priority, m.Subject, m.Body,
		m.PatientID, m.EncounterID, m.SenderID, m.RecipientID, m.PoolID,
		m.Status, m.ParentID, m.ThreadID, m.IsUrgent, m.DueDate)
	return err
}

func (r *inboxMessageRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*InboxMessage, error) {
	return r.scanMsg(r.conn(ctx).QueryRow(ctx, `SELECT `+inboxMsgCols+` FROM inbox_message WHERE id = $1`, id))
}

func (r *inboxMessageRepoPG) Update(ctx context.Context, m *InboxMessage) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE inbox_message SET status=$2, priority=$3, is_urgent=$4, due_date=$5,
			read_at=$6, completed_at=$7, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.Priority, m.IsUrgent, m.DueDate,
		m.ReadAt, m.CompletedAt)
	return err
}

func (r *inboxMessageRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM inbox_message WHERE id = $1`, id)
	return err
}

func (r *inboxMessageRepoPG) ListByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM inbox_message WHERE recipient_id = $1`, recipientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+inboxMsgCols+` FROM inbox_message WHERE recipient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, recipientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*InboxMessage
	for rows.Next() {
		m, err := r.scanMsg(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *inboxMessageRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*InboxMessage, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM inbox_message WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+inboxMsgCols+` FROM inbox_message WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*InboxMessage
	for rows.Next() {
		m, err := r.scanMsg(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

var inboxMessageSearchParams = map[string]fhir.SearchParamConfig{
	"recipient_id": {Type: fhir.SearchParamReference, Column: "recipient_id"},
	"patient_id":   {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":       {Type: fhir.SearchParamToken, Column: "status"},
	"message_type": {Type: fhir.SearchParamToken, Column: "message_type"},
	"priority":     {Type: fhir.SearchParamToken, Column: "priority"},
}

func (r *inboxMessageRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*InboxMessage, int, error) {
	qb := fhir.NewSearchQuery("inbox_message", inboxMsgCols)
	qb.ApplyParams(params, inboxMessageSearchParams)
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
	var items []*InboxMessage
	for rows.Next() {
		m, err := r.scanMsg(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *inboxMessageRepoPG) AddPoolMember(ctx context.Context, m *MessagePoolMember) error {
	m.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO message_pool_member (id, pool_id, user_id, role, is_active)
		VALUES ($1,$2,$3,$4,$5)`,
		m.ID, m.PoolID, m.UserID, m.Role, m.IsActive)
	return err
}

func (r *inboxMessageRepoPG) GetPoolMembers(ctx context.Context, poolID uuid.UUID) ([]*MessagePoolMember, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, pool_id, user_id, role, is_active, joined_at
		FROM message_pool_member WHERE pool_id = $1 AND is_active = TRUE`, poolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*MessagePoolMember
	for rows.Next() {
		var m MessagePoolMember
		if err := rows.Scan(&m.ID, &m.PoolID, &m.UserID, &m.Role, &m.IsActive, &m.JoinedAt); err != nil {
			return nil, err
		}
		items = append(items, &m)
	}
	return items, nil
}

func (r *inboxMessageRepoPG) RemovePoolMember(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM message_pool_member WHERE id = $1`, id)
	return err
}

// =========== Cosign Request Repository ===========

type cosignRequestRepoPG struct{ pool *pgxpool.Pool }

func NewCosignRequestRepoPG(pool *pgxpool.Pool) CosignRequestRepository {
	return &cosignRequestRepoPG{pool: pool}
}

func (r *cosignRequestRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const cosignCols = `id, document_type, document_id, requester_id, cosigner_id, patient_id,
	encounter_id, status, note, requested_at, responded_at, created_at, updated_at`

func (r *cosignRequestRepoPG) scanCosign(row pgx.Row) (*CosignRequest, error) {
	var cr CosignRequest
	err := row.Scan(&cr.ID, &cr.DocumentType, &cr.DocumentID, &cr.RequesterID, &cr.CosignerID,
		&cr.PatientID, &cr.EncounterID, &cr.Status, &cr.Note, &cr.RequestedAt,
		&cr.RespondedAt, &cr.CreatedAt, &cr.UpdatedAt)
	return &cr, err
}

func (r *cosignRequestRepoPG) Create(ctx context.Context, cr *CosignRequest) error {
	cr.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO cosign_request (id, document_type, document_id, requester_id, cosigner_id,
			patient_id, encounter_id, status, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		cr.ID, cr.DocumentType, cr.DocumentID, cr.RequesterID, cr.CosignerID,
		cr.PatientID, cr.EncounterID, cr.Status, cr.Note)
	return err
}

func (r *cosignRequestRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*CosignRequest, error) {
	return r.scanCosign(r.conn(ctx).QueryRow(ctx, `SELECT `+cosignCols+` FROM cosign_request WHERE id = $1`, id))
}

func (r *cosignRequestRepoPG) Update(ctx context.Context, cr *CosignRequest) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE cosign_request SET status=$2, note=$3, responded_at=$4, updated_at=NOW()
		WHERE id = $1`,
		cr.ID, cr.Status, cr.Note, cr.RespondedAt)
	return err
}

func (r *cosignRequestRepoPG) ListByCosigner(ctx context.Context, cosignerID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cosign_request WHERE cosigner_id = $1`, cosignerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cosignCols+` FROM cosign_request WHERE cosigner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, cosignerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CosignRequest
	for rows.Next() {
		cr, err := r.scanCosign(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cr)
	}
	return items, total, nil
}

func (r *cosignRequestRepoPG) ListByRequester(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]*CosignRequest, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cosign_request WHERE requester_id = $1`, requesterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+cosignCols+` FROM cosign_request WHERE requester_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, requesterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*CosignRequest
	for rows.Next() {
		cr, err := r.scanCosign(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, cr)
	}
	return items, total, nil
}

// =========== Patient List Repository ===========

type patientListRepoPG struct{ pool *pgxpool.Pool }

func NewPatientListRepoPG(pool *pgxpool.Pool) PatientListRepository {
	return &patientListRepoPG{pool: pool}
}

func (r *patientListRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const patListCols = `id, list_name, list_type, owner_id, department_id, description,
	auto_criteria, is_active, created_at, updated_at`

func (r *patientListRepoPG) scanList(row pgx.Row) (*PatientList, error) {
	var l PatientList
	err := row.Scan(&l.ID, &l.ListName, &l.ListType, &l.OwnerID, &l.DepartmentID,
		&l.Description, &l.AutoCriteria, &l.IsActive, &l.CreatedAt, &l.UpdatedAt)
	return &l, err
}

func (r *patientListRepoPG) Create(ctx context.Context, l *PatientList) error {
	l.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_list (id, list_name, list_type, owner_id, department_id,
			description, auto_criteria, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		l.ID, l.ListName, l.ListType, l.OwnerID, l.DepartmentID,
		l.Description, l.AutoCriteria, l.IsActive)
	return err
}

func (r *patientListRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PatientList, error) {
	return r.scanList(r.conn(ctx).QueryRow(ctx, `SELECT `+patListCols+` FROM patient_list WHERE id = $1`, id))
}

func (r *patientListRepoPG) Update(ctx context.Context, l *PatientList) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE patient_list SET list_name=$2, list_type=$3, description=$4,
			auto_criteria=$5, is_active=$6, updated_at=NOW()
		WHERE id = $1`,
		l.ID, l.ListName, l.ListType, l.Description,
		l.AutoCriteria, l.IsActive)
	return err
}

func (r *patientListRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient_list WHERE id = $1`, id)
	return err
}

func (r *patientListRepoPG) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*PatientList, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient_list WHERE owner_id = $1`, ownerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+patListCols+` FROM patient_list WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PatientList
	for rows.Next() {
		l, err := r.scanList(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, l)
	}
	return items, total, nil
}

func (r *patientListRepoPG) AddMember(ctx context.Context, m *PatientListMember) error {
	m.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_list_member (id, list_id, patient_id, priority, flags, one_liner, added_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.ListID, m.PatientID, m.Priority, m.Flags, m.OneLiner, m.AddedBy)
	return err
}

func (r *patientListRepoPG) GetMembers(ctx context.Context, listID uuid.UUID, limit, offset int) ([]*PatientListMember, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient_list_member WHERE list_id = $1 AND removed_at IS NULL`, listID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, list_id, patient_id, priority, flags, one_liner, added_by, added_at, removed_at
		FROM patient_list_member WHERE list_id = $1 AND removed_at IS NULL
		ORDER BY priority DESC, added_at DESC LIMIT $2 OFFSET $3`, listID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PatientListMember
	for rows.Next() {
		var m PatientListMember
		if err := rows.Scan(&m.ID, &m.ListID, &m.PatientID, &m.Priority, &m.Flags,
			&m.OneLiner, &m.AddedBy, &m.AddedAt, &m.RemovedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, &m)
	}
	return items, total, nil
}

func (r *patientListRepoPG) RemoveMember(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `UPDATE patient_list_member SET removed_at=NOW() WHERE id = $1`, id)
	return err
}

func (r *patientListRepoPG) UpdateMember(ctx context.Context, m *PatientListMember) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE patient_list_member SET priority=$2, flags=$3, one_liner=$4
		WHERE id = $1`,
		m.ID, m.Priority, m.Flags, m.OneLiner)
	return err
}

// =========== Handoff Repository ===========

type handoffRepoPG struct{ pool *pgxpool.Pool }

func NewHandoffRepoPG(pool *pgxpool.Pool) HandoffRepository {
	return &handoffRepoPG{pool: pool}
}

func (r *handoffRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const handoffCols = `id, patient_id, encounter_id, from_provider_id, to_provider_id,
	handoff_type, illness_severity, patient_summary, action_list,
	situation_awareness, synthesis, contingency_plan,
	status, acknowledged_at, created_at, updated_at`

func (r *handoffRepoPG) scanHandoff(row pgx.Row) (*HandoffRecord, error) {
	var h HandoffRecord
	err := row.Scan(&h.ID, &h.PatientID, &h.EncounterID, &h.FromProviderID, &h.ToProviderID,
		&h.HandoffType, &h.IllnessSeverity, &h.PatientSummary, &h.ActionList,
		&h.SituationAwareness, &h.Synthesis, &h.ContingencyPlan,
		&h.Status, &h.AcknowledgedAt, &h.CreatedAt, &h.UpdatedAt)
	return &h, err
}

func (r *handoffRepoPG) Create(ctx context.Context, h *HandoffRecord) error {
	h.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO handoff_record (id, patient_id, encounter_id, from_provider_id, to_provider_id,
			handoff_type, illness_severity, patient_summary, action_list,
			situation_awareness, synthesis, contingency_plan, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		h.ID, h.PatientID, h.EncounterID, h.FromProviderID, h.ToProviderID,
		h.HandoffType, h.IllnessSeverity, h.PatientSummary, h.ActionList,
		h.SituationAwareness, h.Synthesis, h.ContingencyPlan, h.Status)
	return err
}

func (r *handoffRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*HandoffRecord, error) {
	return r.scanHandoff(r.conn(ctx).QueryRow(ctx, `SELECT `+handoffCols+` FROM handoff_record WHERE id = $1`, id))
}

func (r *handoffRepoPG) Update(ctx context.Context, h *HandoffRecord) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE handoff_record SET illness_severity=$2, patient_summary=$3, action_list=$4,
			situation_awareness=$5, synthesis=$6, contingency_plan=$7,
			status=$8, acknowledged_at=$9, updated_at=NOW()
		WHERE id = $1`,
		h.ID, h.IllnessSeverity, h.PatientSummary, h.ActionList,
		h.SituationAwareness, h.Synthesis, h.ContingencyPlan,
		h.Status, h.AcknowledgedAt)
	return err
}

func (r *handoffRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM handoff_record WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+handoffCols+` FROM handoff_record WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*HandoffRecord
	for rows.Next() {
		h, err := r.scanHandoff(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, h)
	}
	return items, total, nil
}

func (r *handoffRepoPG) ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]*HandoffRecord, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM handoff_record WHERE from_provider_id = $1 OR to_provider_id = $1`, providerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+handoffCols+` FROM handoff_record WHERE from_provider_id = $1 OR to_provider_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, providerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*HandoffRecord
	for rows.Next() {
		h, err := r.scanHandoff(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, h)
	}
	return items, total, nil
}
