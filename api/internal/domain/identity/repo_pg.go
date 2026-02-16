package identity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/internal/platform/hipaa"
)

// -- Patient Repository --

type patientRepoPG struct {
	pool      *pgxpool.Pool
	encryptor hipaa.FieldEncryptor
}

func NewPatientRepo(pool *pgxpool.Pool) PatientRepository {
	return &patientRepoPG{pool: pool}
}

// NewPatientRepoWithEncryption creates a patient repository with PHI field-level encryption.
// The encryptor is used to encrypt fields before storage and decrypt after retrieval.
// Pass nil to disable encryption (equivalent to NewPatientRepo).
func NewPatientRepoWithEncryption(pool *pgxpool.Pool, enc hipaa.FieldEncryptor) PatientRepository {
	return &patientRepoPG{pool: pool, encryptor: enc}
}

func (r *patientRepoPG) conn(ctx context.Context) querier {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const patientCols = `id, fhir_id, active, mrn, prefix, first_name, middle_name, last_name, suffix, maiden_name,
	birth_date, gender, deceased_boolean, deceased_datetime, marital_status,
	multiple_birth, multiple_birth_int, photo_url,
	ssn_hash, abha_id, abha_address, aadhaar_hash,
	phone_home, phone_mobile, phone_work, email,
	address_use, address_line1, address_line2, city, district, state, postal_code, country,
	preferred_language, interpreter_needed,
	primary_care_provider_id, managing_org_id,
	version_id, created_at, updated_at`

func (r *patientRepoPG) Create(ctx context.Context, p *Patient) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}

	// Encrypt PHI fields before storage, then restore originals for the caller.
	if err := r.encryptPatientPHI(p); err != nil {
		return fmt.Errorf("patient create: %w", err)
	}
	defer r.decryptPatientPHI(p) //nolint:errcheck // best-effort restore

	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient (
			id, fhir_id, active, mrn, prefix, first_name, middle_name, last_name, suffix, maiden_name,
			birth_date, gender, deceased_boolean, deceased_datetime, marital_status,
			multiple_birth, multiple_birth_int, photo_url,
			ssn_hash, abha_id, abha_address, aadhaar_hash,
			phone_home, phone_mobile, phone_work, email,
			address_use, address_line1, address_line2, city, district, state, postal_code, country,
			preferred_language, interpreter_needed,
			primary_care_provider_id, managing_org_id
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15,$16,$17,$18,
			$19,$20,$21,$22,
			$23,$24,$25,$26,
			$27,$28,$29,$30,$31,$32,$33,$34,
			$35,$36,$37,$38
		)`,
		p.ID, p.FHIRID, p.Active, p.MRN, p.Prefix, p.FirstName, p.MiddleName, p.LastName, p.Suffix, p.MaidenName,
		p.BirthDate, p.Gender, p.DeceasedBoolean, p.DeceasedDatetime, p.MaritalStatus,
		p.MultipleBirth, p.MultipleBirthInt, p.PhotoURL,
		p.SSNHash, p.AbhaID, p.AbhaAddress, p.AadhaarHash,
		p.PhoneHome, p.PhoneMobile, p.PhoneWork, p.Email,
		p.AddressUse, p.AddressLine1, p.AddressLine2, p.City, p.District, p.State, p.PostalCode, p.Country,
		p.PreferredLanguage, p.InterpreterNeeded,
		p.PrimaryCareProviderID, p.ManagingOrgID,
	)
	return err
}

func (r *patientRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Patient, error) {
	p, err := scanPatient(r.conn(ctx).QueryRow(ctx, `SELECT `+patientCols+` FROM patient WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	if err := r.decryptPatientPHI(p); err != nil {
		return nil, fmt.Errorf("patient get by id: %w", err)
	}
	return p, nil
}

func (r *patientRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Patient, error) {
	p, err := scanPatient(r.conn(ctx).QueryRow(ctx, `SELECT `+patientCols+` FROM patient WHERE fhir_id = $1`, fhirID))
	if err != nil {
		return nil, err
	}
	if err := r.decryptPatientPHI(p); err != nil {
		return nil, fmt.Errorf("patient get by fhir id: %w", err)
	}
	return p, nil
}

func (r *patientRepoPG) GetByMRN(ctx context.Context, mrn string) (*Patient, error) {
	p, err := scanPatient(r.conn(ctx).QueryRow(ctx, `SELECT `+patientCols+` FROM patient WHERE mrn = $1`, mrn))
	if err != nil {
		return nil, err
	}
	if err := r.decryptPatientPHI(p); err != nil {
		return nil, fmt.Errorf("patient get by mrn: %w", err)
	}
	return p, nil
}

func (r *patientRepoPG) Update(ctx context.Context, p *Patient) error {
	// Encrypt PHI fields before storage, then restore originals for the caller.
	if err := r.encryptPatientPHI(p); err != nil {
		return fmt.Errorf("patient update: %w", err)
	}
	defer r.decryptPatientPHI(p) //nolint:errcheck // best-effort restore

	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE patient SET
			active=$2, mrn=$3, prefix=$4, first_name=$5, middle_name=$6, last_name=$7, suffix=$8, maiden_name=$9,
			birth_date=$10, gender=$11, deceased_boolean=$12, deceased_datetime=$13, marital_status=$14,
			multiple_birth=$15, multiple_birth_int=$16, photo_url=$17,
			abha_id=$18, abha_address=$19,
			phone_home=$20, phone_mobile=$21, phone_work=$22, email=$23,
			address_use=$24, address_line1=$25, address_line2=$26, city=$27, district=$28, state=$29, postal_code=$30, country=$31,
			preferred_language=$32, interpreter_needed=$33,
			primary_care_provider_id=$34, managing_org_id=$35, version_id=$36, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Active, p.MRN, p.Prefix, p.FirstName, p.MiddleName, p.LastName, p.Suffix, p.MaidenName,
		p.BirthDate, p.Gender, p.DeceasedBoolean, p.DeceasedDatetime, p.MaritalStatus,
		p.MultipleBirth, p.MultipleBirthInt, p.PhotoURL,
		p.AbhaID, p.AbhaAddress,
		p.PhoneHome, p.PhoneMobile, p.PhoneWork, p.Email,
		p.AddressUse, p.AddressLine1, p.AddressLine2, p.City, p.District, p.State, p.PostalCode, p.Country,
		p.PreferredLanguage, p.InterpreterNeeded,
		p.PrimaryCareProviderID, p.ManagingOrgID, p.VersionID,
	)
	return err
}

func (r *patientRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient WHERE id = $1`, id)
	return err
}

func (r *patientRepoPG) List(ctx context.Context, limit, offset int) ([]*Patient, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM patient`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+patientCols+` FROM patient ORDER BY last_name, first_name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var patients []*Patient
	for rows.Next() {
		p, err := scanPatientRows(rows)
		if err != nil {
			return nil, 0, err
		}
		if err := r.decryptPatientPHI(p); err != nil {
			return nil, 0, fmt.Errorf("patient list: %w", err)
		}
		patients = append(patients, p)
	}
	return patients, total, nil
}

var patientSearchParams = map[string]fhir.SearchParamConfig{
	"family":     {Type: fhir.SearchParamString, Column: "last_name"},
	"given":      {Type: fhir.SearchParamString, Column: "first_name"},
	"birthdate":  {Type: fhir.SearchParamDate, Column: "birth_date"},
	"gender":     {Type: fhir.SearchParamToken, Column: "gender"},
	"identifier": {Type: fhir.SearchParamToken, Column: "mrn"},
}

func (r *patientRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Patient, int, error) {
	qb := fhir.NewSearchQuery("patient", patientCols)
	if name, ok := params["name"]; ok {
		qb.Add(fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d)", qb.Idx(), qb.Idx()), "%"+name+"%")
	}
	qb.ApplyParams(params, patientSearchParams)
	qb.OrderBy("last_name, first_name")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var patients []*Patient
	for rows.Next() {
		p, err := scanPatientRows(rows)
		if err != nil {
			return nil, 0, err
		}
		if err := r.decryptPatientPHI(p); err != nil {
			return nil, 0, fmt.Errorf("patient search: %w", err)
		}
		patients = append(patients, p)
	}
	return patients, total, nil
}

// Contacts
func (r *patientRepoPG) AddContact(ctx context.Context, c *PatientContact) error {
	c.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_contact (id, patient_id, relationship, prefix, first_name, last_name,
			phone, email, address_line1, city, state, postal_code, country, gender, is_primary_contact, period_start, period_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		c.ID, c.PatientID, c.Relationship, c.Prefix, c.FirstName, c.LastName,
		c.Phone, c.Email, c.AddressLine1, c.City, c.State, c.PostalCode, c.Country,
		c.Gender, c.IsPrimaryContact, c.PeriodStart, c.PeriodEnd,
	)
	return err
}

func (r *patientRepoPG) GetContacts(ctx context.Context, patientID uuid.UUID) ([]*PatientContact, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, patient_id, relationship, prefix, first_name, last_name,
			phone, email, address_line1, city, state, postal_code, country, gender, is_primary_contact, period_start, period_end
		FROM patient_contact WHERE patient_id = $1`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*PatientContact
	for rows.Next() {
		var c PatientContact
		if err := rows.Scan(&c.ID, &c.PatientID, &c.Relationship, &c.Prefix, &c.FirstName, &c.LastName,
			&c.Phone, &c.Email, &c.AddressLine1, &c.City, &c.State, &c.PostalCode, &c.Country,
			&c.Gender, &c.IsPrimaryContact, &c.PeriodStart, &c.PeriodEnd); err != nil {
			return nil, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, nil
}

func (r *patientRepoPG) RemoveContact(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient_contact WHERE id = $1`, id)
	return err
}

// Identifiers
func (r *patientRepoPG) AddIdentifier(ctx context.Context, ident *PatientIdentifier) error {
	ident.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_identifier (id, patient_id, system_uri, value, type_code, type_display, assigner, period_start, period_end)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		ident.ID, ident.PatientID, ident.SystemURI, ident.Value,
		ident.TypeCode, ident.TypeDisplay, ident.Assigner, ident.PeriodStart, ident.PeriodEnd,
	)
	return err
}

func (r *patientRepoPG) GetIdentifiers(ctx context.Context, patientID uuid.UUID) ([]*PatientIdentifier, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, patient_id, system_uri, value, type_code, type_display, assigner, period_start, period_end
		FROM patient_identifier WHERE patient_id = $1`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var idents []*PatientIdentifier
	for rows.Next() {
		var i PatientIdentifier
		if err := rows.Scan(&i.ID, &i.PatientID, &i.SystemURI, &i.Value,
			&i.TypeCode, &i.TypeDisplay, &i.Assigner, &i.PeriodStart, &i.PeriodEnd); err != nil {
			return nil, err
		}
		idents = append(idents, &i)
	}
	return idents, nil
}

func (r *patientRepoPG) RemoveIdentifier(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient_identifier WHERE id = $1`, id)
	return err
}

// -- PHI Encryption Helpers --

func (r *patientRepoPG) encryptField(value *string) (*string, error) {
	if r.encryptor == nil || value == nil || *value == "" {
		return value, nil
	}
	encrypted, err := r.encryptor.Encrypt(*value)
	if err != nil {
		return nil, fmt.Errorf("encrypting PHI field: %w", err)
	}
	return &encrypted, nil
}

func (r *patientRepoPG) decryptField(value *string) (*string, error) {
	if r.encryptor == nil || value == nil || *value == "" {
		return value, nil
	}
	decrypted, err := r.encryptor.Decrypt(*value)
	if err != nil {
		return nil, fmt.Errorf("decrypting PHI field: %w", err)
	}
	return &decrypted, nil
}

// encryptPatientPHI encrypts all PHI fields on a Patient in place before database storage.
func (r *patientRepoPG) encryptPatientPHI(p *Patient) error {
	var err error
	if p.SSNHash, err = r.encryptField(p.SSNHash); err != nil {
		return err
	}
	if p.AadhaarHash, err = r.encryptField(p.AadhaarHash); err != nil {
		return err
	}
	if p.PhoneHome, err = r.encryptField(p.PhoneHome); err != nil {
		return err
	}
	if p.PhoneMobile, err = r.encryptField(p.PhoneMobile); err != nil {
		return err
	}
	if p.PhoneWork, err = r.encryptField(p.PhoneWork); err != nil {
		return err
	}
	if p.Email, err = r.encryptField(p.Email); err != nil {
		return err
	}
	if p.AddressLine1, err = r.encryptField(p.AddressLine1); err != nil {
		return err
	}
	if p.AddressLine2, err = r.encryptField(p.AddressLine2); err != nil {
		return err
	}
	if p.City, err = r.encryptField(p.City); err != nil {
		return err
	}
	if p.District, err = r.encryptField(p.District); err != nil {
		return err
	}
	if p.State, err = r.encryptField(p.State); err != nil {
		return err
	}
	if p.PostalCode, err = r.encryptField(p.PostalCode); err != nil {
		return err
	}
	return nil
}

// decryptPatientPHI decrypts all PHI fields on a Patient in place after database retrieval.
func (r *patientRepoPG) decryptPatientPHI(p *Patient) error {
	var err error
	if p.SSNHash, err = r.decryptField(p.SSNHash); err != nil {
		return err
	}
	if p.AadhaarHash, err = r.decryptField(p.AadhaarHash); err != nil {
		return err
	}
	if p.PhoneHome, err = r.decryptField(p.PhoneHome); err != nil {
		return err
	}
	if p.PhoneMobile, err = r.decryptField(p.PhoneMobile); err != nil {
		return err
	}
	if p.PhoneWork, err = r.decryptField(p.PhoneWork); err != nil {
		return err
	}
	if p.Email, err = r.decryptField(p.Email); err != nil {
		return err
	}
	if p.AddressLine1, err = r.decryptField(p.AddressLine1); err != nil {
		return err
	}
	if p.AddressLine2, err = r.decryptField(p.AddressLine2); err != nil {
		return err
	}
	if p.City, err = r.decryptField(p.City); err != nil {
		return err
	}
	if p.District, err = r.decryptField(p.District); err != nil {
		return err
	}
	if p.State, err = r.decryptField(p.State); err != nil {
		return err
	}
	if p.PostalCode, err = r.decryptField(p.PostalCode); err != nil {
		return err
	}
	return nil
}

func scanPatient(row pgx.Row) (*Patient, error) {
	var p Patient
	err := row.Scan(
		&p.ID, &p.FHIRID, &p.Active, &p.MRN, &p.Prefix, &p.FirstName, &p.MiddleName, &p.LastName, &p.Suffix, &p.MaidenName,
		&p.BirthDate, &p.Gender, &p.DeceasedBoolean, &p.DeceasedDatetime, &p.MaritalStatus,
		&p.MultipleBirth, &p.MultipleBirthInt, &p.PhotoURL,
		&p.SSNHash, &p.AbhaID, &p.AbhaAddress, &p.AadhaarHash,
		&p.PhoneHome, &p.PhoneMobile, &p.PhoneWork, &p.Email,
		&p.AddressUse, &p.AddressLine1, &p.AddressLine2, &p.City, &p.District, &p.State, &p.PostalCode, &p.Country,
		&p.PreferredLanguage, &p.InterpreterNeeded,
		&p.PrimaryCareProviderID, &p.ManagingOrgID,
		&p.VersionID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPatientRows(rows pgx.Rows) (*Patient, error) {
	var p Patient
	err := rows.Scan(
		&p.ID, &p.FHIRID, &p.Active, &p.MRN, &p.Prefix, &p.FirstName, &p.MiddleName, &p.LastName, &p.Suffix, &p.MaidenName,
		&p.BirthDate, &p.Gender, &p.DeceasedBoolean, &p.DeceasedDatetime, &p.MaritalStatus,
		&p.MultipleBirth, &p.MultipleBirthInt, &p.PhotoURL,
		&p.SSNHash, &p.AbhaID, &p.AbhaAddress, &p.AadhaarHash,
		&p.PhoneHome, &p.PhoneMobile, &p.PhoneWork, &p.Email,
		&p.AddressUse, &p.AddressLine1, &p.AddressLine2, &p.City, &p.District, &p.State, &p.PostalCode, &p.Country,
		&p.PreferredLanguage, &p.InterpreterNeeded,
		&p.PrimaryCareProviderID, &p.ManagingOrgID,
		&p.VersionID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// -- PatientLink Repository --

type patientLinkRepoPG struct {
	pool *pgxpool.Pool
}

func NewPatientLinkRepo(pool *pgxpool.Pool) PatientLinkRepository {
	return &patientLinkRepoPG{pool: pool}
}

func (r *patientLinkRepoPG) conn(ctx context.Context) querier {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *patientLinkRepoPG) Create(ctx context.Context, link *PatientLink) error {
	link.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO patient_link (id, patient_id, linked_patient_id, link_type, confidence, match_method, match_details, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		link.ID, link.PatientID, link.LinkedPatientID, link.LinkType,
		link.Confidence, link.MatchMethod, link.MatchDetails, link.CreatedBy,
	)
	return err
}

func (r *patientLinkRepoPG) GetByPatientID(ctx context.Context, patientID uuid.UUID) ([]*PatientLink, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, patient_id, linked_patient_id, link_type, confidence, match_method, match_details, created_at, created_by
		FROM patient_link WHERE patient_id = $1 ORDER BY created_at DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*PatientLink
	for rows.Next() {
		var l PatientLink
		if err := rows.Scan(&l.ID, &l.PatientID, &l.LinkedPatientID, &l.LinkType,
			&l.Confidence, &l.MatchMethod, &l.MatchDetails, &l.CreatedAt, &l.CreatedBy); err != nil {
			return nil, err
		}
		links = append(links, &l)
	}
	return links, nil
}

func (r *patientLinkRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM patient_link WHERE id = $1`, id)
	return err
}

// -- Practitioner Repository --

type practRepoPG struct {
	pool      *pgxpool.Pool
	encryptor hipaa.FieldEncryptor
}

func NewPractitionerRepo(pool *pgxpool.Pool) PractitionerRepository {
	return &practRepoPG{pool: pool}
}

// NewPractitionerRepoWithEncryption creates a practitioner repository with PII field-level encryption.
// The encryptor is used to encrypt fields before storage and decrypt after retrieval.
// Pass nil to disable encryption (equivalent to NewPractitionerRepo).
func NewPractitionerRepoWithEncryption(pool *pgxpool.Pool, enc hipaa.FieldEncryptor) PractitionerRepository {
	return &practRepoPG{pool: pool, encryptor: enc}
}

func (r *practRepoPG) conn(ctx context.Context) querier {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const practCols = `id, fhir_id, active, prefix, first_name, middle_name, last_name, suffix,
	gender, birth_date, photo_url,
	npi_number, dea_number, state_license_num, state_license_state,
	medical_council_reg, abha_id, hpr_id,
	phone, email, address_line1, city, state, postal_code, country,
	qualification_summary, version_id, created_at, updated_at`

func (r *practRepoPG) Create(ctx context.Context, p *Practitioner) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}

	// Encrypt PII fields before storage, then restore originals for the caller.
	if err := r.encryptPractitionerPII(p); err != nil {
		return fmt.Errorf("practitioner create: %w", err)
	}
	defer r.decryptPractitionerPII(p) //nolint:errcheck // best-effort restore

	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO practitioner (
			id, fhir_id, active, prefix, first_name, middle_name, last_name, suffix,
			gender, birth_date, photo_url,
			npi_number, dea_number, state_license_num, state_license_state,
			medical_council_reg, abha_id, hpr_id,
			phone, email, address_line1, city, state, postal_code, country,
			qualification_summary
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)`,
		p.ID, p.FHIRID, p.Active, p.Prefix, p.FirstName, p.MiddleName, p.LastName, p.Suffix,
		p.Gender, p.BirthDate, p.PhotoURL,
		p.NPINumber, p.DEANumber, p.StateLicenseNum, p.StateLicenseState,
		p.MedicalCouncilReg, p.AbhaID, p.HPRID,
		p.Phone, p.Email, p.AddressLine1, p.City, p.State, p.PostalCode, p.Country,
		p.QualificationSummary,
	)
	return err
}

func (r *practRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Practitioner, error) {
	p, err := scanPractitioner(r.conn(ctx).QueryRow(ctx, `SELECT `+practCols+` FROM practitioner WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	if err := r.decryptPractitionerPII(p); err != nil {
		return nil, fmt.Errorf("practitioner get by id: %w", err)
	}
	return p, nil
}

func (r *practRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Practitioner, error) {
	p, err := scanPractitioner(r.conn(ctx).QueryRow(ctx, `SELECT `+practCols+` FROM practitioner WHERE fhir_id = $1`, fhirID))
	if err != nil {
		return nil, err
	}
	if err := r.decryptPractitionerPII(p); err != nil {
		return nil, fmt.Errorf("practitioner get by fhir id: %w", err)
	}
	return p, nil
}

func (r *practRepoPG) GetByNPI(ctx context.Context, npi string) (*Practitioner, error) {
	p, err := scanPractitioner(r.conn(ctx).QueryRow(ctx, `SELECT `+practCols+` FROM practitioner WHERE npi_number = $1`, npi))
	if err != nil {
		return nil, err
	}
	if err := r.decryptPractitionerPII(p); err != nil {
		return nil, fmt.Errorf("practitioner get by npi: %w", err)
	}
	return p, nil
}

func (r *practRepoPG) Update(ctx context.Context, p *Practitioner) error {
	// Encrypt PII fields before storage, then restore originals for the caller.
	if err := r.encryptPractitionerPII(p); err != nil {
		return fmt.Errorf("practitioner update: %w", err)
	}
	defer r.decryptPractitionerPII(p) //nolint:errcheck // best-effort restore

	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE practitioner SET
			active=$2, prefix=$3, first_name=$4, middle_name=$5, last_name=$6, suffix=$7,
			gender=$8, birth_date=$9, photo_url=$10,
			npi_number=$11, dea_number=$12, state_license_num=$13, state_license_state=$14,
			medical_council_reg=$15, abha_id=$16, hpr_id=$17,
			phone=$18, email=$19, address_line1=$20, city=$21, state=$22, postal_code=$23, country=$24,
			qualification_summary=$25, version_id=$26, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Active, p.Prefix, p.FirstName, p.MiddleName, p.LastName, p.Suffix,
		p.Gender, p.BirthDate, p.PhotoURL,
		p.NPINumber, p.DEANumber, p.StateLicenseNum, p.StateLicenseState,
		p.MedicalCouncilReg, p.AbhaID, p.HPRID,
		p.Phone, p.Email, p.AddressLine1, p.City, p.State, p.PostalCode, p.Country,
		p.QualificationSummary, p.VersionID,
	)
	return err
}

func (r *practRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM practitioner WHERE id = $1`, id)
	return err
}

func (r *practRepoPG) List(ctx context.Context, limit, offset int) ([]*Practitioner, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM practitioner`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+practCols+` FROM practitioner ORDER BY last_name, first_name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var practs []*Practitioner
	for rows.Next() {
		p, err := scanPractitionerRows(rows)
		if err != nil {
			return nil, 0, err
		}
		if err := r.decryptPractitionerPII(p); err != nil {
			return nil, 0, fmt.Errorf("practitioner list: %w", err)
		}
		practs = append(practs, p)
	}
	return practs, total, nil
}

var practitionerSearchParams = map[string]fhir.SearchParamConfig{
	"family":     {Type: fhir.SearchParamString, Column: "last_name"},
	"identifier": {Type: fhir.SearchParamToken, Column: "npi_number"},
}

func (r *practRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Practitioner, int, error) {
	qb := fhir.NewSearchQuery("practitioner", practCols)
	if name, ok := params["name"]; ok {
		qb.Add(fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d)", qb.Idx(), qb.Idx()), "%"+name+"%")
	}
	qb.ApplyParams(params, practitionerSearchParams)
	qb.OrderBy("last_name, first_name")

	var total int
	if err := r.conn(ctx).QueryRow(ctx, qb.CountSQL(), qb.CountArgs()...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, qb.DataSQL(limit, offset), qb.DataArgs(limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var practs []*Practitioner
	for rows.Next() {
		p, err := scanPractitionerRows(rows)
		if err != nil {
			return nil, 0, err
		}
		if err := r.decryptPractitionerPII(p); err != nil {
			return nil, 0, fmt.Errorf("practitioner search: %w", err)
		}
		practs = append(practs, p)
	}
	return practs, total, nil
}

func (r *practRepoPG) AddRole(ctx context.Context, role *PractitionerRole) error {
	role.ID = uuid.New()
	if role.FHIRID == "" {
		role.FHIRID = role.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO practitioner_role (id, fhir_id, practitioner_id, organization_id, department_id,
			role_code, role_display, period_start, period_end, active, telehealth_capable, accepting_patients)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		role.ID, role.FHIRID, role.PractitionerID, role.OrganizationID, role.DepartmentID,
		role.RoleCode, role.RoleDisplay, role.PeriodStart, role.PeriodEnd, role.Active,
		role.TelehealthCapable, role.AcceptingPatients,
	)
	return err
}

func (r *practRepoPG) GetRoles(ctx context.Context, practitionerID uuid.UUID) ([]*PractitionerRole, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, fhir_id, practitioner_id, organization_id, department_id,
			role_code, role_display, period_start, period_end, active, telehealth_capable, accepting_patients, created_at
		FROM practitioner_role WHERE practitioner_id = $1`, practitionerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*PractitionerRole
	for rows.Next() {
		var r PractitionerRole
		if err := rows.Scan(&r.ID, &r.FHIRID, &r.PractitionerID, &r.OrganizationID, &r.DepartmentID,
			&r.RoleCode, &r.RoleDisplay, &r.PeriodStart, &r.PeriodEnd, &r.Active,
			&r.TelehealthCapable, &r.AcceptingPatients, &r.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, &r)
	}
	return roles, nil
}

func (r *practRepoPG) RemoveRole(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM practitioner_role WHERE id = $1`, id)
	return err
}

func scanPractitioner(row pgx.Row) (*Practitioner, error) {
	var p Practitioner
	err := row.Scan(
		&p.ID, &p.FHIRID, &p.Active, &p.Prefix, &p.FirstName, &p.MiddleName, &p.LastName, &p.Suffix,
		&p.Gender, &p.BirthDate, &p.PhotoURL,
		&p.NPINumber, &p.DEANumber, &p.StateLicenseNum, &p.StateLicenseState,
		&p.MedicalCouncilReg, &p.AbhaID, &p.HPRID,
		&p.Phone, &p.Email, &p.AddressLine1, &p.City, &p.State, &p.PostalCode, &p.Country,
		&p.QualificationSummary, &p.VersionID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPractitionerRows(rows pgx.Rows) (*Practitioner, error) {
	var p Practitioner
	err := rows.Scan(
		&p.ID, &p.FHIRID, &p.Active, &p.Prefix, &p.FirstName, &p.MiddleName, &p.LastName, &p.Suffix,
		&p.Gender, &p.BirthDate, &p.PhotoURL,
		&p.NPINumber, &p.DEANumber, &p.StateLicenseNum, &p.StateLicenseState,
		&p.MedicalCouncilReg, &p.AbhaID, &p.HPRID,
		&p.Phone, &p.Email, &p.AddressLine1, &p.City, &p.State, &p.PostalCode, &p.Country,
		&p.QualificationSummary, &p.VersionID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// -- Practitioner PII Encryption Helpers --

func (r *practRepoPG) encryptPractitionerField(value *string) (*string, error) {
	if r.encryptor == nil || value == nil || *value == "" {
		return value, nil
	}
	encrypted, err := r.encryptor.Encrypt(*value)
	if err != nil {
		return nil, fmt.Errorf("encrypting PII field: %w", err)
	}
	return &encrypted, nil
}

func (r *practRepoPG) decryptPractitionerField(value *string) (*string, error) {
	if r.encryptor == nil || value == nil || *value == "" {
		return value, nil
	}
	decrypted, err := r.encryptor.Decrypt(*value)
	if err != nil {
		return nil, fmt.Errorf("decrypting PII field: %w", err)
	}
	return &decrypted, nil
}

// encryptPractitionerPII encrypts all PII fields on a Practitioner in place before database storage.
func (r *practRepoPG) encryptPractitionerPII(p *Practitioner) error {
	var err error
	if p.Phone, err = r.encryptPractitionerField(p.Phone); err != nil {
		return err
	}
	if p.Email, err = r.encryptPractitionerField(p.Email); err != nil {
		return err
	}
	if p.AddressLine1, err = r.encryptPractitionerField(p.AddressLine1); err != nil {
		return err
	}
	if p.City, err = r.encryptPractitionerField(p.City); err != nil {
		return err
	}
	if p.State, err = r.encryptPractitionerField(p.State); err != nil {
		return err
	}
	if p.PostalCode, err = r.encryptPractitionerField(p.PostalCode); err != nil {
		return err
	}
	if p.Country, err = r.encryptPractitionerField(p.Country); err != nil {
		return err
	}
	if p.NPINumber, err = r.encryptPractitionerField(p.NPINumber); err != nil {
		return err
	}
	if p.DEANumber, err = r.encryptPractitionerField(p.DEANumber); err != nil {
		return err
	}
	if p.StateLicenseNum, err = r.encryptPractitionerField(p.StateLicenseNum); err != nil {
		return err
	}
	if p.MedicalCouncilReg, err = r.encryptPractitionerField(p.MedicalCouncilReg); err != nil {
		return err
	}
	if p.AbhaID, err = r.encryptPractitionerField(p.AbhaID); err != nil {
		return err
	}
	return nil
}

// decryptPractitionerPII decrypts all PII fields on a Practitioner in place after database retrieval.
func (r *practRepoPG) decryptPractitionerPII(p *Practitioner) error {
	var err error
	if p.Phone, err = r.decryptPractitionerField(p.Phone); err != nil {
		return err
	}
	if p.Email, err = r.decryptPractitionerField(p.Email); err != nil {
		return err
	}
	if p.AddressLine1, err = r.decryptPractitionerField(p.AddressLine1); err != nil {
		return err
	}
	if p.City, err = r.decryptPractitionerField(p.City); err != nil {
		return err
	}
	if p.State, err = r.decryptPractitionerField(p.State); err != nil {
		return err
	}
	if p.PostalCode, err = r.decryptPractitionerField(p.PostalCode); err != nil {
		return err
	}
	if p.Country, err = r.decryptPractitionerField(p.Country); err != nil {
		return err
	}
	if p.NPINumber, err = r.decryptPractitionerField(p.NPINumber); err != nil {
		return err
	}
	if p.DEANumber, err = r.decryptPractitionerField(p.DEANumber); err != nil {
		return err
	}
	if p.StateLicenseNum, err = r.decryptPractitionerField(p.StateLicenseNum); err != nil {
		return err
	}
	if p.MedicalCouncilReg, err = r.decryptPractitionerField(p.MedicalCouncilReg); err != nil {
		return err
	}
	if p.AbhaID, err = r.decryptPractitionerField(p.AbhaID); err != nil {
		return err
	}
	return nil
}

// -- PractitionerRole Repository --

type practRoleRepoPG struct {
	pool *pgxpool.Pool
}

func NewPractitionerRoleRepoPG(pool *pgxpool.Pool) PractitionerRoleRepository {
	return &practRoleRepoPG{pool: pool}
}

func (r *practRoleRepoPG) conn(ctx context.Context) querier {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const practRoleCols = `id, fhir_id, practitioner_id, organization_id, department_id,
	role_code, role_display, period_start, period_end, active,
	telehealth_capable, accepting_patients, version_id, created_at, updated_at`

func scanPractitionerRole(row pgx.Row) (*PractitionerRole, error) {
	var r PractitionerRole
	err := row.Scan(
		&r.ID, &r.FHIRID, &r.PractitionerID, &r.OrganizationID, &r.DepartmentID,
		&r.RoleCode, &r.RoleDisplay, &r.PeriodStart, &r.PeriodEnd, &r.Active,
		&r.TelehealthCapable, &r.AcceptingPatients, &r.VersionID, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func scanPractitionerRoleRows(rows pgx.Rows) (*PractitionerRole, error) {
	var r PractitionerRole
	err := rows.Scan(
		&r.ID, &r.FHIRID, &r.PractitionerID, &r.OrganizationID, &r.DepartmentID,
		&r.RoleCode, &r.RoleDisplay, &r.PeriodStart, &r.PeriodEnd, &r.Active,
		&r.TelehealthCapable, &r.AcceptingPatients, &r.VersionID, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *practRoleRepoPG) Create(ctx context.Context, role *PractitionerRole) error {
	role.ID = uuid.New()
	if role.FHIRID == "" {
		role.FHIRID = role.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO practitioner_role (
			id, fhir_id, practitioner_id, organization_id, department_id,
			role_code, role_display, period_start, period_end, active,
			telehealth_capable, accepting_patients
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		role.ID, role.FHIRID, role.PractitionerID, role.OrganizationID, role.DepartmentID,
		role.RoleCode, role.RoleDisplay, role.PeriodStart, role.PeriodEnd, role.Active,
		role.TelehealthCapable, role.AcceptingPatients,
	)
	return err
}

func (r *practRoleRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*PractitionerRole, error) {
	return scanPractitionerRole(r.conn(ctx).QueryRow(ctx, `SELECT `+practRoleCols+` FROM practitioner_role WHERE id = $1`, id))
}

func (r *practRoleRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*PractitionerRole, error) {
	return scanPractitionerRole(r.conn(ctx).QueryRow(ctx, `SELECT `+practRoleCols+` FROM practitioner_role WHERE fhir_id = $1`, fhirID))
}

func (r *practRoleRepoPG) Update(ctx context.Context, role *PractitionerRole) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE practitioner_role SET
			role_code=$2, role_display=$3, active=$4, period_start=$5, period_end=$6,
			accepting_patients=$7, telehealth_capable=$8, version_id=$9, updated_at=NOW()
		WHERE id = $1`,
		role.ID, role.RoleCode, role.RoleDisplay, role.Active, role.PeriodStart, role.PeriodEnd,
		role.AcceptingPatients, role.TelehealthCapable, role.VersionID,
	)
	return err
}

func (r *practRoleRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM practitioner_role WHERE id = $1`, id)
	return err
}

func (r *practRoleRepoPG) List(ctx context.Context, limit, offset int) ([]*PractitionerRole, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM practitioner_role`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+practRoleCols+` FROM practitioner_role ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []*PractitionerRole
	for rows.Next() {
		role, err := scanPractitionerRoleRows(rows)
		if err != nil {
			return nil, 0, err
		}
		roles = append(roles, role)
	}
	return roles, total, nil
}

var practitionerRoleSearchParams = map[string]fhir.SearchParamConfig{
	"practitioner": {Type: fhir.SearchParamReference, Column: "practitioner_id"},
	"organization": {Type: fhir.SearchParamReference, Column: "organization_id"},
	"role":         {Type: fhir.SearchParamToken, Column: "role_code"},
	"active":       {Type: fhir.SearchParamToken, Column: "active"},
}

func (r *practRoleRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PractitionerRole, int, error) {
	qb := fhir.NewSearchQuery("practitioner_role", practRoleCols)
	qb.ApplyParams(params, practitionerRoleSearchParams)
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

	var roles []*PractitionerRole
	for rows.Next() {
		role, err := scanPractitionerRoleRows(rows)
		if err != nil {
			return nil, 0, err
		}
		roles = append(roles, role)
	}
	return roles, total, nil
}

type querier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}
