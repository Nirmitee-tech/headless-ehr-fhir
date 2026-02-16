package documents

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

// =========== Consent Repository ===========

type consentRepoPG struct{ pool *pgxpool.Pool }

func NewConsentRepoPG(pool *pgxpool.Pool) ConsentRepository { return &consentRepoPG{pool: pool} }

func (r *consentRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const consentCols = `id, fhir_id, status, scope, category_code, category_display,
	patient_id, performer_id, organization_id, policy_authority, policy_uri,
	provision_type, provision_start, provision_end, provision_action,
	hipaa_authorization, abdm_consent, abdm_consent_id,
	signature_type, signature_when, signature_data,
	date_time, note, version_id, created_at, updated_at`

func (r *consentRepoPG) scanConsent(row pgx.Row) (*Consent, error) {
	var c Consent
	err := row.Scan(&c.ID, &c.FHIRID, &c.Status, &c.Scope, &c.CategoryCode, &c.CategoryDisplay,
		&c.PatientID, &c.PerformerID, &c.OrganizationID, &c.PolicyAuthority, &c.PolicyURI,
		&c.ProvisionType, &c.ProvisionStart, &c.ProvisionEnd, &c.ProvisionAction,
		&c.HIPAAAuth, &c.ABDMConsent, &c.ABDMConsentID,
		&c.SignatureType, &c.SignatureWhen, &c.SignatureData,
		&c.DateTime, &c.Note, &c.VersionID, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *consentRepoPG) Create(ctx context.Context, c *Consent) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO consent (id, fhir_id, status, scope, category_code, category_display,
			patient_id, performer_id, organization_id, policy_authority, policy_uri,
			provision_type, provision_start, provision_end, provision_action,
			hipaa_authorization, abdm_consent, abdm_consent_id,
			signature_type, signature_when, signature_data, date_time, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		c.ID, c.FHIRID, c.Status, c.Scope, c.CategoryCode, c.CategoryDisplay,
		c.PatientID, c.PerformerID, c.OrganizationID, c.PolicyAuthority, c.PolicyURI,
		c.ProvisionType, c.ProvisionStart, c.ProvisionEnd, c.ProvisionAction,
		c.HIPAAAuth, c.ABDMConsent, c.ABDMConsentID,
		c.SignatureType, c.SignatureWhen, c.SignatureData, c.DateTime, c.Note)
	return err
}

func (r *consentRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Consent, error) {
	return r.scanConsent(r.conn(ctx).QueryRow(ctx, `SELECT `+consentCols+` FROM consent WHERE id = $1`, id))
}

func (r *consentRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Consent, error) {
	return r.scanConsent(r.conn(ctx).QueryRow(ctx, `SELECT `+consentCols+` FROM consent WHERE fhir_id = $1`, fhirID))
}

func (r *consentRepoPG) Update(ctx context.Context, c *Consent) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE consent SET status=$2, scope=$3, category_code=$4, category_display=$5,
			provision_type=$6, provision_start=$7, provision_end=$8, provision_action=$9,
			hipaa_authorization=$10, abdm_consent=$11, abdm_consent_id=$12,
			signature_type=$13, signature_when=$14, signature_data=$15, note=$16, version_id=$17, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.Status, c.Scope, c.CategoryCode, c.CategoryDisplay,
		c.ProvisionType, c.ProvisionStart, c.ProvisionEnd, c.ProvisionAction,
		c.HIPAAAuth, c.ABDMConsent, c.ABDMConsentID,
		c.SignatureType, c.SignatureWhen, c.SignatureData, c.Note, c.VersionID)
	return err
}

func (r *consentRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM consent WHERE id = $1`, id)
	return err
}

func (r *consentRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Consent, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM consent WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+consentCols+` FROM consent WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Consent
	for rows.Next() {
		c, err := r.scanConsent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

var consentSearchParams = map[string]fhir.SearchParamConfig{
	"patient":  {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"category": {Type: fhir.SearchParamToken, Column: "category_code"},
}

func (r *consentRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Consent, int, error) {
	qb := fhir.NewSearchQuery("consent", consentCols)
	qb.ApplyParams(params, consentSearchParams)
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
	var items []*Consent
	for rows.Next() {
		c, err := r.scanConsent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

// =========== DocumentReference Repository ===========

type docRefRepoPG struct{ pool *pgxpool.Pool }

func NewDocumentReferenceRepoPG(pool *pgxpool.Pool) DocumentReferenceRepository {
	return &docRefRepoPG{pool: pool}
}

func (r *docRefRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const docRefCols = `id, fhir_id, status, doc_status, type_code, type_display,
	category_code, category_display, patient_id, author_id, custodian_id, encounter_id,
	date, description, security_label,
	content_type, content_url, content_size, content_hash, content_title,
	format_code, format_display, version_id, created_at, updated_at`

func (r *docRefRepoPG) scanDocRef(row pgx.Row) (*DocumentReference, error) {
	var d DocumentReference
	err := row.Scan(&d.ID, &d.FHIRID, &d.Status, &d.DocStatus, &d.TypeCode, &d.TypeDisplay,
		&d.CategoryCode, &d.CategoryDisplay, &d.PatientID, &d.AuthorID, &d.CustodianID, &d.EncounterID,
		&d.Date, &d.Description, &d.SecurityLabel,
		&d.ContentType, &d.ContentURL, &d.ContentSize, &d.ContentHash, &d.ContentTitle,
		&d.FormatCode, &d.FormatDisplay, &d.VersionID, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}

func (r *docRefRepoPG) Create(ctx context.Context, d *DocumentReference) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO document_reference (id, fhir_id, status, doc_status, type_code, type_display,
			category_code, category_display, patient_id, author_id, custodian_id, encounter_id,
			date, description, security_label,
			content_type, content_url, content_size, content_hash, content_title,
			format_code, format_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		d.ID, d.FHIRID, d.Status, d.DocStatus, d.TypeCode, d.TypeDisplay,
		d.CategoryCode, d.CategoryDisplay, d.PatientID, d.AuthorID, d.CustodianID, d.EncounterID,
		d.Date, d.Description, d.SecurityLabel,
		d.ContentType, d.ContentURL, d.ContentSize, d.ContentHash, d.ContentTitle,
		d.FormatCode, d.FormatDisplay)
	return err
}

func (r *docRefRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DocumentReference, error) {
	return r.scanDocRef(r.conn(ctx).QueryRow(ctx, `SELECT `+docRefCols+` FROM document_reference WHERE id = $1`, id))
}

func (r *docRefRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*DocumentReference, error) {
	return r.scanDocRef(r.conn(ctx).QueryRow(ctx, `SELECT `+docRefCols+` FROM document_reference WHERE fhir_id = $1`, fhirID))
}

func (r *docRefRepoPG) Update(ctx context.Context, d *DocumentReference) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE document_reference SET status=$2, doc_status=$3, type_code=$4, type_display=$5,
			category_code=$6, category_display=$7, description=$8, security_label=$9,
			content_type=$10, content_url=$11, content_size=$12, content_hash=$13, content_title=$14,
			format_code=$15, format_display=$16, version_id=$17, updated_at=NOW()
		WHERE id = $1`,
		d.ID, d.Status, d.DocStatus, d.TypeCode, d.TypeDisplay,
		d.CategoryCode, d.CategoryDisplay, d.Description, d.SecurityLabel,
		d.ContentType, d.ContentURL, d.ContentSize, d.ContentHash, d.ContentTitle,
		d.FormatCode, d.FormatDisplay, d.VersionID)
	return err
}

func (r *docRefRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM document_reference WHERE id = $1`, id)
	return err
}

func (r *docRefRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*DocumentReference, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM document_reference WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+docRefCols+` FROM document_reference WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DocumentReference
	for rows.Next() {
		d, err := r.scanDocRef(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

var docRefSearchParams = map[string]fhir.SearchParamConfig{
	"patient":  {Type: fhir.SearchParamReference, Column: "patient_id"},
	"status":   {Type: fhir.SearchParamToken, Column: "status"},
	"type":     {Type: fhir.SearchParamToken, Column: "type_code"},
	"category": {Type: fhir.SearchParamToken, Column: "category_code"},
}

func (r *docRefRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DocumentReference, int, error) {
	qb := fhir.NewSearchQuery("document_reference", docRefCols)
	qb.ApplyParams(params, docRefSearchParams)
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
	var items []*DocumentReference
	for rows.Next() {
		d, err := r.scanDocRef(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}

// =========== ClinicalNote Repository ===========

type noteRepoPG struct{ pool *pgxpool.Pool }

func NewClinicalNoteRepoPG(pool *pgxpool.Pool) ClinicalNoteRepository {
	return &noteRepoPG{pool: pool}
}

func (r *noteRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const noteCols = `id, patient_id, encounter_id, author_id, note_type, status, title,
	subjective, objective, assessment, plan, note_text,
	signed_by, signed_at, cosigned_by, cosigned_at,
	amended_by, amended_at, amended_reason, created_at, updated_at`

func (r *noteRepoPG) scanNote(row pgx.Row) (*ClinicalNote, error) {
	var n ClinicalNote
	err := row.Scan(&n.ID, &n.PatientID, &n.EncounterID, &n.AuthorID, &n.NoteType, &n.Status, &n.Title,
		&n.Subjective, &n.Objective, &n.Assessment, &n.Plan, &n.NoteText,
		&n.SignedBy, &n.SignedAt, &n.CosignedBy, &n.CosignedAt,
		&n.AmendedBy, &n.AmendedAt, &n.AmendedReason, &n.CreatedAt, &n.UpdatedAt)
	return &n, err
}

func (r *noteRepoPG) Create(ctx context.Context, n *ClinicalNote) error {
	n.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO clinical_note (id, patient_id, encounter_id, author_id, note_type, status, title,
			subjective, objective, assessment, plan, note_text)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		n.ID, n.PatientID, n.EncounterID, n.AuthorID, n.NoteType, n.Status, n.Title,
		n.Subjective, n.Objective, n.Assessment, n.Plan, n.NoteText)
	return err
}

func (r *noteRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*ClinicalNote, error) {
	return r.scanNote(r.conn(ctx).QueryRow(ctx, `SELECT `+noteCols+` FROM clinical_note WHERE id = $1`, id))
}

func (r *noteRepoPG) Update(ctx context.Context, n *ClinicalNote) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE clinical_note SET status=$2, title=$3, subjective=$4, objective=$5,
			assessment=$6, plan=$7, note_text=$8,
			signed_by=$9, signed_at=$10, cosigned_by=$11, cosigned_at=$12,
			amended_by=$13, amended_at=$14, amended_reason=$15, updated_at=NOW()
		WHERE id = $1`,
		n.ID, n.Status, n.Title, n.Subjective, n.Objective,
		n.Assessment, n.Plan, n.NoteText,
		n.SignedBy, n.SignedAt, n.CosignedBy, n.CosignedAt,
		n.AmendedBy, n.AmendedAt, n.AmendedReason)
	return err
}

func (r *noteRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM clinical_note WHERE id = $1`, id)
	return err
}

func (r *noteRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM clinical_note WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+noteCols+` FROM clinical_note WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ClinicalNote
	for rows.Next() {
		n, err := r.scanNote(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

func (r *noteRepoPG) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*ClinicalNote, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM clinical_note WHERE encounter_id = $1`, encounterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+noteCols+` FROM clinical_note WHERE encounter_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, encounterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*ClinicalNote
	for rows.Next() {
		n, err := r.scanNote(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

// =========== Composition Repository ===========

type compRepoPG struct{ pool *pgxpool.Pool }

func NewCompositionRepoPG(pool *pgxpool.Pool) CompositionRepository {
	return &compRepoPG{pool: pool}
}

func (r *compRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const compCols = `id, fhir_id, status, type_code, type_display, category_code, category_display,
	patient_id, encounter_id, date, author_id, title, confidentiality, custodian_id,
	version_id, created_at, updated_at`

func (r *compRepoPG) scanComp(row pgx.Row) (*Composition, error) {
	var c Composition
	err := row.Scan(&c.ID, &c.FHIRID, &c.Status, &c.TypeCode, &c.TypeDisplay, &c.CategoryCode, &c.CategoryDisplay,
		&c.PatientID, &c.EncounterID, &c.Date, &c.AuthorID, &c.Title, &c.Confidentiality, &c.CustodianID,
		&c.VersionID, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func (r *compRepoPG) Create(ctx context.Context, c *Composition) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO composition (id, fhir_id, status, type_code, type_display,
			category_code, category_display, patient_id, encounter_id, date,
			author_id, title, confidentiality, custodian_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		c.ID, c.FHIRID, c.Status, c.TypeCode, c.TypeDisplay,
		c.CategoryCode, c.CategoryDisplay, c.PatientID, c.EncounterID, c.Date,
		c.AuthorID, c.Title, c.Confidentiality, c.CustodianID)
	return err
}

func (r *compRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Composition, error) {
	return r.scanComp(r.conn(ctx).QueryRow(ctx, `SELECT `+compCols+` FROM composition WHERE id = $1`, id))
}

func (r *compRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Composition, error) {
	return r.scanComp(r.conn(ctx).QueryRow(ctx, `SELECT `+compCols+` FROM composition WHERE fhir_id = $1`, fhirID))
}

func (r *compRepoPG) Update(ctx context.Context, c *Composition) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE composition SET status=$2, type_code=$3, type_display=$4,
			category_code=$5, category_display=$6, title=$7, confidentiality=$8, version_id=$9, updated_at=NOW()
		WHERE id = $1`,
		c.ID, c.Status, c.TypeCode, c.TypeDisplay,
		c.CategoryCode, c.CategoryDisplay, c.Title, c.Confidentiality, c.VersionID)
	return err
}

func (r *compRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM composition WHERE id = $1`, id)
	return err
}

func (r *compRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Composition, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM composition WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+compCols+` FROM composition WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*Composition
	for rows.Next() {
		c, err := r.scanComp(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, nil
}

func (r *compRepoPG) AddSection(ctx context.Context, s *CompositionSection) error {
	s.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO composition_section (id, composition_id, title, code_value, code_display,
			text_status, text_div, mode, ordered_by, entry_reference, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.CompositionID, s.Title, s.CodeValue, s.CodeDisplay,
		s.TextStatus, s.TextDiv, s.Mode, s.OrderedBy, s.EntryReference, s.SortOrder)
	return err
}

func (r *compRepoPG) GetSections(ctx context.Context, compositionID uuid.UUID) ([]*CompositionSection, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, composition_id, title, code_value, code_display,
			text_status, text_div, mode, ordered_by, entry_reference, sort_order
		FROM composition_section WHERE composition_id = $1 ORDER BY sort_order ASC NULLS LAST`, compositionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*CompositionSection
	for rows.Next() {
		var s CompositionSection
		if err := rows.Scan(&s.ID, &s.CompositionID, &s.Title, &s.CodeValue, &s.CodeDisplay,
			&s.TextStatus, &s.TextDiv, &s.Mode, &s.OrderedBy, &s.EntryReference, &s.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, &s)
	}
	return items, nil
}

// =========== DocumentTemplate Repository ===========

type documentTemplateRepoPG struct{ pool *pgxpool.Pool }

func NewDocumentTemplateRepoPG(pool *pgxpool.Pool) DocumentTemplateRepository {
	return &documentTemplateRepoPG{pool: pool}
}

func (r *documentTemplateRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const templateCols = `id, name, description, status, type_code, type_display, created_at, updated_at, created_by`

func (r *documentTemplateRepoPG) scanTemplate(row pgx.Row) (*DocumentTemplate, error) {
	var t DocumentTemplate
	err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Status, &t.TypeCode, &t.TypeDisplay,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy)
	return &t, err
}

func (r *documentTemplateRepoPG) Create(ctx context.Context, t *DocumentTemplate) error {
	t.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO document_template (id, name, description, status, type_code, type_display, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		t.ID, t.Name, t.Description, t.Status, t.TypeCode, t.TypeDisplay, t.CreatedBy)
	return err
}

func (r *documentTemplateRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*DocumentTemplate, error) {
	return r.scanTemplate(r.conn(ctx).QueryRow(ctx, `SELECT `+templateCols+` FROM document_template WHERE id = $1`, id))
}

func (r *documentTemplateRepoPG) Update(ctx context.Context, t *DocumentTemplate) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE document_template SET name=$2, description=$3, status=$4, type_code=$5, type_display=$6, updated_at=NOW()
		WHERE id = $1`,
		t.ID, t.Name, t.Description, t.Status, t.TypeCode, t.TypeDisplay)
	return err
}

func (r *documentTemplateRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM document_template WHERE id = $1`, id)
	return err
}

func (r *documentTemplateRepoPG) List(ctx context.Context, limit, offset int) ([]*DocumentTemplate, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM document_template`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+templateCols+` FROM document_template ORDER BY name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*DocumentTemplate
	for rows.Next() {
		t, err := r.scanTemplate(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, nil
}

func (r *documentTemplateRepoPG) AddSection(ctx context.Context, s *TemplateSection) error {
	s.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO template_section (id, template_id, title, sort_order, content_template, required)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		s.ID, s.TemplateID, s.Title, s.SortOrder, s.ContentTemplate, s.Required)
	return err
}

func (r *documentTemplateRepoPG) GetSections(ctx context.Context, templateID uuid.UUID) ([]*TemplateSection, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, template_id, title, sort_order, content_template, required
		FROM template_section WHERE template_id = $1 ORDER BY sort_order ASC`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*TemplateSection
	for rows.Next() {
		var s TemplateSection
		if err := rows.Scan(&s.ID, &s.TemplateID, &s.Title, &s.SortOrder, &s.ContentTemplate, &s.Required); err != nil {
			return nil, err
		}
		items = append(items, &s)
	}
	return items, nil
}
