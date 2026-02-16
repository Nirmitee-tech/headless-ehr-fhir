package portal

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

// =========== Portal Account Repository ===========

type portalAccountRepoPG struct{ pool *pgxpool.Pool }

func NewPortalAccountRepoPG(pool *pgxpool.Pool) PortalAccountRepository {
	return &portalAccountRepoPG{pool: pool}
}

func (r *portalAccountRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const acctCols = `id, patient_id, username, email, phone, status, email_verified,
	last_login_at, failed_login_count, password_last_changed, mfa_enabled,
	preferred_language, note, created_at, updated_at`

func (r *portalAccountRepoPG) scanAccount(row pgx.Row) (*PortalAccount, error) {
	var a PortalAccount
	err := row.Scan(&a.ID, &a.PatientID, &a.Username, &a.Email, &a.Phone, &a.Status, &a.EmailVerified,
		&a.LastLoginAt, &a.FailedLoginCount, &a.PasswordLastChanged, &a.MFAEnabled,
		&a.PreferredLanguage, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	return &a, err
}

func (r *portalAccountRepoPG) Create(ctx context.Context, a *PortalAccount) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO portal_account (id, patient_id, username, email, phone, status,
			email_verified, mfa_enabled, preferred_language, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		a.ID, a.PatientID, a.Username, a.Email, a.Phone, a.Status,
		a.EmailVerified, a.MFAEnabled, a.PreferredLanguage, a.Note)
	return err
}

func (r *portalAccountRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PortalAccount, error) {
	return r.scanAccount(r.conn(ctx).QueryRow(ctx, `SELECT `+acctCols+` FROM portal_account WHERE id = $1`, id))
}

func (r *portalAccountRepoPG) Update(ctx context.Context, a *PortalAccount) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE portal_account SET email=$2, phone=$3, status=$4, email_verified=$5,
			mfa_enabled=$6, preferred_language=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		a.ID, a.Email, a.Phone, a.Status, a.EmailVerified,
		a.MFAEnabled, a.PreferredLanguage, a.Note)
	return err
}

func (r *portalAccountRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM portal_account WHERE id = $1`, id)
	return err
}

func (r *portalAccountRepoPG) List(ctx context.Context, limit, offset int) ([]*PortalAccount, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM portal_account`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+acctCols+` FROM portal_account ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PortalAccount
	for rows.Next() {
		a, err := r.scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

func (r *portalAccountRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalAccount, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM portal_account WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+acctCols+` FROM portal_account WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PortalAccount
	for rows.Next() {
		a, err := r.scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// =========== Portal Message Repository ===========

type portalMessageRepoPG struct{ pool *pgxpool.Pool }

func NewPortalMessageRepoPG(pool *pgxpool.Pool) PortalMessageRepository {
	return &portalMessageRepoPG{pool: pool}
}

func (r *portalMessageRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const msgCols = `id, patient_id, practitioner_id, direction, subject, body, status,
	priority, category, parent_id, read_at, created_at, updated_at`

func (r *portalMessageRepoPG) scanMessage(row pgx.Row) (*PortalMessage, error) {
	var m PortalMessage
	err := row.Scan(&m.ID, &m.PatientID, &m.PractitionerID, &m.Direction, &m.Subject, &m.Body, &m.Status,
		&m.Priority, &m.Category, &m.ParentID, &m.ReadAt, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func (r *portalMessageRepoPG) Create(ctx context.Context, m *PortalMessage) error {
	m.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO portal_message (id, patient_id, practitioner_id, direction, subject, body, status,
			priority, category, parent_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		m.ID, m.PatientID, m.PractitionerID, m.Direction, m.Subject, m.Body, m.Status,
		m.Priority, m.Category, m.ParentID)
	return err
}

func (r *portalMessageRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PortalMessage, error) {
	return r.scanMessage(r.conn(ctx).QueryRow(ctx, `SELECT `+msgCols+` FROM portal_message WHERE id = $1`, id))
}

func (r *portalMessageRepoPG) Update(ctx context.Context, m *PortalMessage) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE portal_message SET status=$2, read_at=$3, updated_at=NOW()
		WHERE id = $1`,
		m.ID, m.Status, m.ReadAt)
	return err
}

func (r *portalMessageRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM portal_message WHERE id = $1`, id)
	return err
}

func (r *portalMessageRepoPG) List(ctx context.Context, limit, offset int) ([]*PortalMessage, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM portal_message`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+msgCols+` FROM portal_message ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PortalMessage
	for rows.Next() {
		m, err := r.scanMessage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

func (r *portalMessageRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalMessage, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM portal_message WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+msgCols+` FROM portal_message WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PortalMessage
	for rows.Next() {
		m, err := r.scanMessage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, nil
}

// =========== Questionnaire Repository ===========

type questionnaireRepoPG struct{ pool *pgxpool.Pool }

func NewQuestionnaireRepoPG(pool *pgxpool.Pool) QuestionnaireRepository {
	return &questionnaireRepoPG{pool: pool}
}

func (r *questionnaireRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const questCols = `id, fhir_id, name, title, status, version, description, purpose,
	subject_type, date, publisher, approval_date, last_review_date, created_at, updated_at`

func (r *questionnaireRepoPG) scanQuestionnaire(row pgx.Row) (*Questionnaire, error) {
	var q Questionnaire
	err := row.Scan(&q.ID, &q.FHIRID, &q.Name, &q.Title, &q.Status, &q.Version, &q.Description, &q.Purpose,
		&q.SubjectType, &q.Date, &q.Publisher, &q.ApprovalDate, &q.LastReviewDate, &q.CreatedAt, &q.UpdatedAt)
	return &q, err
}

func (r *questionnaireRepoPG) Create(ctx context.Context, q *Questionnaire) error {
	q.ID = uuid.New()
	if q.FHIRID == "" {
		q.FHIRID = q.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO questionnaire (id, fhir_id, name, title, status, version, description, purpose,
			subject_type, date, publisher, approval_date, last_review_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		q.ID, q.FHIRID, q.Name, q.Title, q.Status, q.Version, q.Description, q.Purpose,
		q.SubjectType, q.Date, q.Publisher, q.ApprovalDate, q.LastReviewDate)
	return err
}

func (r *questionnaireRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Questionnaire, error) {
	return r.scanQuestionnaire(r.conn(ctx).QueryRow(ctx, `SELECT `+questCols+` FROM questionnaire WHERE id = $1`, id))
}

func (r *questionnaireRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Questionnaire, error) {
	return r.scanQuestionnaire(r.conn(ctx).QueryRow(ctx, `SELECT `+questCols+` FROM questionnaire WHERE fhir_id = $1`, fhirID))
}

func (r *questionnaireRepoPG) Update(ctx context.Context, q *Questionnaire) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE questionnaire SET name=$2, title=$3, status=$4, version=$5, description=$6,
			purpose=$7, publisher=$8, updated_at=NOW()
		WHERE id = $1`,
		q.ID, q.Name, q.Title, q.Status, q.Version, q.Description,
		q.Purpose, q.Publisher)
	return err
}

func (r *questionnaireRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM questionnaire WHERE id = $1`, id)
	return err
}

func (r *questionnaireRepoPG) List(ctx context.Context, limit, offset int) ([]*Questionnaire, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM questionnaire`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+questCols+` FROM questionnaire ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Questionnaire
	for rows.Next() {
		q, err := r.scanQuestionnaire(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, q)
	}
	return items, total, nil
}

var questSearchParams = map[string]fhir.SearchParamConfig{
	"name":      {Type: fhir.SearchParamString, Column: "name"},
	"status":    {Type: fhir.SearchParamToken, Column: "status"},
	"publisher": {Type: fhir.SearchParamString, Column: "publisher"},
}

func (r *questionnaireRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Questionnaire, int, error) {
	qb := fhir.NewSearchQuery("questionnaire", questCols)
	qb.ApplyParams(params, questSearchParams)
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
	var items []*Questionnaire
	for rows.Next() {
		q, err := r.scanQuestionnaire(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, q)
	}
	return items, total, nil
}

func (r *questionnaireRepoPG) AddItem(ctx context.Context, item *QuestionnaireItem) error {
	item.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO questionnaire_item (id, questionnaire_id, link_id, text, type, required, repeats,
			read_only, max_length, answer_options, initial_value,
			enable_when_link_id, enable_when_operator, enable_when_answer, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		item.ID, item.QuestionnaireID, item.LinkID, item.Text, item.Type, item.Required, item.Repeats,
		item.ReadOnly, item.MaxLength, item.AnswerOptions, item.InitialValue,
		item.EnableWhenLinkID, item.EnableWhenOperator, item.EnableWhenAnswer, item.SortOrder)
	return err
}

func (r *questionnaireRepoPG) GetItems(ctx context.Context, questionnaireID uuid.UUID) ([]*QuestionnaireItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, questionnaire_id, link_id, text, type, required, repeats,
			read_only, max_length, answer_options, initial_value,
			enable_when_link_id, enable_when_operator, enable_when_answer, sort_order
		FROM questionnaire_item WHERE questionnaire_id = $1 ORDER BY sort_order`, questionnaireID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*QuestionnaireItem
	for rows.Next() {
		var item QuestionnaireItem
		if err := rows.Scan(&item.ID, &item.QuestionnaireID, &item.LinkID, &item.Text, &item.Type,
			&item.Required, &item.Repeats, &item.ReadOnly, &item.MaxLength, &item.AnswerOptions,
			&item.InitialValue, &item.EnableWhenLinkID, &item.EnableWhenOperator, &item.EnableWhenAnswer,
			&item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// =========== Questionnaire Response Repository ===========

type questionnaireResponseRepoPG struct{ pool *pgxpool.Pool }

func NewQuestionnaireResponseRepoPG(pool *pgxpool.Pool) QuestionnaireResponseRepository {
	return &questionnaireResponseRepoPG{pool: pool}
}

func (r *questionnaireResponseRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const qrCols = `id, fhir_id, questionnaire_id, patient_id, encounter_id, author_id,
	status, authored, created_at, updated_at`

func (r *questionnaireResponseRepoPG) scanQR(row pgx.Row) (*QuestionnaireResponse, error) {
	var qr QuestionnaireResponse
	err := row.Scan(&qr.ID, &qr.FHIRID, &qr.QuestionnaireID, &qr.PatientID, &qr.EncounterID, &qr.AuthorID,
		&qr.Status, &qr.Authored, &qr.CreatedAt, &qr.UpdatedAt)
	return &qr, err
}

func (r *questionnaireResponseRepoPG) Create(ctx context.Context, qr *QuestionnaireResponse) error {
	qr.ID = uuid.New()
	if qr.FHIRID == "" {
		qr.FHIRID = qr.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO questionnaire_response (id, fhir_id, questionnaire_id, patient_id, encounter_id,
			author_id, status, authored)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		qr.ID, qr.FHIRID, qr.QuestionnaireID, qr.PatientID, qr.EncounterID,
		qr.AuthorID, qr.Status, qr.Authored)
	return err
}

func (r *questionnaireResponseRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*QuestionnaireResponse, error) {
	return r.scanQR(r.conn(ctx).QueryRow(ctx, `SELECT `+qrCols+` FROM questionnaire_response WHERE id = $1`, id))
}

func (r *questionnaireResponseRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*QuestionnaireResponse, error) {
	return r.scanQR(r.conn(ctx).QueryRow(ctx, `SELECT `+qrCols+` FROM questionnaire_response WHERE fhir_id = $1`, fhirID))
}

func (r *questionnaireResponseRepoPG) Update(ctx context.Context, qr *QuestionnaireResponse) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE questionnaire_response SET status=$2, authored=$3, updated_at=NOW()
		WHERE id = $1`,
		qr.ID, qr.Status, qr.Authored)
	return err
}

func (r *questionnaireResponseRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM questionnaire_response WHERE id = $1`, id)
	return err
}

func (r *questionnaireResponseRepoPG) List(ctx context.Context, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM questionnaire_response`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+qrCols+` FROM questionnaire_response ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*QuestionnaireResponse
	for rows.Next() {
		qr, err := r.scanQR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, qr)
	}
	return items, total, nil
}

func (r *questionnaireResponseRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM questionnaire_response WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+qrCols+` FROM questionnaire_response WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*QuestionnaireResponse
	for rows.Next() {
		qr, err := r.scanQR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, qr)
	}
	return items, total, nil
}

var qrSearchParams = map[string]fhir.SearchParamConfig{
	"patient":       {Type: fhir.SearchParamReference, Column: "patient_id"},
	"questionnaire": {Type: fhir.SearchParamReference, Column: "questionnaire_id"},
	"status":        {Type: fhir.SearchParamToken, Column: "status"},
}

func (r *questionnaireResponseRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	qb := fhir.NewSearchQuery("questionnaire_response", qrCols)
	qb.ApplyParams(params, qrSearchParams)
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
	var items []*QuestionnaireResponse
	for rows.Next() {
		qr, err := r.scanQR(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, qr)
	}
	return items, total, nil
}

func (r *questionnaireResponseRepoPG) AddResponseItem(ctx context.Context, item *QuestionnaireResponseItem) error {
	item.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO questionnaire_response_item (id, response_id, link_id, text,
			answer_string, answer_integer, answer_boolean, answer_date, answer_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		item.ID, item.ResponseID, item.LinkID, item.Text,
		item.AnswerStr, item.AnswerInt, item.AnswerBool, item.AnswerDate, item.AnswerCode)
	return err
}

func (r *questionnaireResponseRepoPG) GetResponseItems(ctx context.Context, responseID uuid.UUID) ([]*QuestionnaireResponseItem, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, response_id, link_id, text,
			answer_string, answer_integer, answer_boolean, answer_date, answer_code
		FROM questionnaire_response_item WHERE response_id = $1`, responseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*QuestionnaireResponseItem
	for rows.Next() {
		var item QuestionnaireResponseItem
		if err := rows.Scan(&item.ID, &item.ResponseID, &item.LinkID, &item.Text,
			&item.AnswerStr, &item.AnswerInt, &item.AnswerBool, &item.AnswerDate, &item.AnswerCode); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// =========== Patient Checkin Repository ===========

type patientCheckinRepoPG struct{ pool *pgxpool.Pool }

func NewPatientCheckinRepoPG(pool *pgxpool.Pool) PatientCheckinRepository {
	return &patientCheckinRepoPG{pool: pool}
}

func (r *patientCheckinRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const checkinCols = `id, patient_id, appointment_id, status, checkin_method, checkin_time,
	insurance_verified, co_pay_collected, co_pay_amount, note, created_at, updated_at`

func (r *patientCheckinRepoPG) scanCheckin(row pgx.Row) (*PatientCheckin, error) {
	var c PatientCheckin
	err := row.Scan(&c.ID, &c.PatientID, &c.AppointmentID, &c.Status, &c.CheckinMethod, &c.CheckinTime,
		&c.InsuranceVerified, &c.CoPayCollected, &c.CoPayAmount, &c.Note, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *patientCheckinRepoPG) Create(ctx context.Context, c *PatientCheckin) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_checkin (id, patient_id, appointment_id, status, checkin_method, checkin_time,
			insurance_verified, co_pay_collected, co_pay_amount, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.PatientID, c.AppointmentID, c.Status, c.CheckinMethod, c.CheckinTime,
		c.InsuranceVerified, c.CoPayCollected, c.CoPayAmount, c.Note)
	return err
}

func (r *patientCheckinRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PatientCheckin, error) {
	return r.scanCheckin(r.conn(ctx).QueryRow(ctx, `SELECT `+checkinCols+` FROM patient_checkin WHERE id = $1`, id))
}

func (r *patientCheckinRepoPG) Update(ctx context.Context, c *PatientCheckin) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE patient_checkin SET status=$2, checkin_method=$3, checkin_time=$4,
			insurance_verified=$5, co_pay_collected=$6, co_pay_amount=$7, note=$8, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.Status, c.CheckinMethod, c.CheckinTime,
		c.InsuranceVerified, c.CoPayCollected, c.CoPayAmount, c.Note)
	return err
}

func (r *patientCheckinRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient_checkin WHERE id = $1`, id)
	return err
}

func (r *patientCheckinRepoPG) List(ctx context.Context, limit, offset int) ([]*PatientCheckin, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient_checkin`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+checkinCols+` FROM patient_checkin ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PatientCheckin
	for rows.Next() {
		c, err := r.scanCheckin(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

func (r *patientCheckinRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientCheckin, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient_checkin WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+checkinCols+` FROM patient_checkin WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*PatientCheckin
	for rows.Next() {
		c, err := r.scanCheckin(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}
