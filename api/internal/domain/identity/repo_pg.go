package identity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
)

// -- Patient Repository --

type patientRepoPG struct {
	pool *pgxpool.Pool
}

func NewPatientRepo(pool *pgxpool.Pool) PatientRepository {
	return &patientRepoPG{pool: pool}
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
	created_at, updated_at`

func (r *patientRepoPG) Create(ctx context.Context, p *Patient) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
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
	return scanPatient(r.conn(ctx).QueryRow(ctx, `SELECT `+patientCols+` FROM patient WHERE id = $1`, id))
}

func (r *patientRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Patient, error) {
	return scanPatient(r.conn(ctx).QueryRow(ctx, `SELECT `+patientCols+` FROM patient WHERE fhir_id = $1`, fhirID))
}

func (r *patientRepoPG) GetByMRN(ctx context.Context, mrn string) (*Patient, error) {
	return scanPatient(r.conn(ctx).QueryRow(ctx, `SELECT `+patientCols+` FROM patient WHERE mrn = $1`, mrn))
}

func (r *patientRepoPG) Update(ctx context.Context, p *Patient) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE patient SET
			active=$2, mrn=$3, prefix=$4, first_name=$5, middle_name=$6, last_name=$7, suffix=$8, maiden_name=$9,
			birth_date=$10, gender=$11, deceased_boolean=$12, deceased_datetime=$13, marital_status=$14,
			multiple_birth=$15, multiple_birth_int=$16, photo_url=$17,
			abha_id=$18, abha_address=$19,
			phone_home=$20, phone_mobile=$21, phone_work=$22, email=$23,
			address_use=$24, address_line1=$25, address_line2=$26, city=$27, district=$28, state=$29, postal_code=$30, country=$31,
			preferred_language=$32, interpreter_needed=$33,
			primary_care_provider_id=$34, managing_org_id=$35, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Active, p.MRN, p.Prefix, p.FirstName, p.MiddleName, p.LastName, p.Suffix, p.MaidenName,
		p.BirthDate, p.Gender, p.DeceasedBoolean, p.DeceasedDatetime, p.MaritalStatus,
		p.MultipleBirth, p.MultipleBirthInt, p.PhotoURL,
		p.AbhaID, p.AbhaAddress,
		p.PhoneHome, p.PhoneMobile, p.PhoneWork, p.Email,
		p.AddressUse, p.AddressLine1, p.AddressLine2, p.City, p.District, p.State, p.PostalCode, p.Country,
		p.PreferredLanguage, p.InterpreterNeeded,
		p.PrimaryCareProviderID, p.ManagingOrgID,
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
		patients = append(patients, p)
	}
	return patients, total, nil
}

func (r *patientRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Patient, int, error) {
	query := `SELECT ` + patientCols + ` FROM patient WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM patient WHERE 1=1`
	var args []interface{}
	idx := 1

	if name, ok := params["name"]; ok {
		clause := fmt.Sprintf(` AND (first_name ILIKE $%d OR last_name ILIKE $%d)`, idx, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+name+"%")
		idx++
	}
	if family, ok := params["family"]; ok {
		clause := fmt.Sprintf(` AND last_name ILIKE $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+family+"%")
		idx++
	}
	if given, ok := params["given"]; ok {
		clause := fmt.Sprintf(` AND first_name ILIKE $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+given+"%")
		idx++
	}
	if birthdate, ok := params["birthdate"]; ok {
		clause := fmt.Sprintf(` AND birth_date = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, birthdate)
		idx++
	}
	if gender, ok := params["gender"]; ok {
		clause := fmt.Sprintf(` AND gender = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, gender)
		idx++
	}
	if identifier, ok := params["identifier"]; ok {
		clause := fmt.Sprintf(` AND mrn = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, identifier)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY last_name, first_name LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
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
		&p.CreatedAt, &p.UpdatedAt,
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
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// -- Practitioner Repository --

type practRepoPG struct {
	pool *pgxpool.Pool
}

func NewPractitionerRepo(pool *pgxpool.Pool) PractitionerRepository {
	return &practRepoPG{pool: pool}
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
	qualification_summary, created_at, updated_at`

func (r *practRepoPG) Create(ctx context.Context, p *Practitioner) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
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
	return scanPractitioner(r.conn(ctx).QueryRow(ctx, `SELECT `+practCols+` FROM practitioner WHERE id = $1`, id))
}

func (r *practRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Practitioner, error) {
	return scanPractitioner(r.conn(ctx).QueryRow(ctx, `SELECT `+practCols+` FROM practitioner WHERE fhir_id = $1`, fhirID))
}

func (r *practRepoPG) GetByNPI(ctx context.Context, npi string) (*Practitioner, error) {
	return scanPractitioner(r.conn(ctx).QueryRow(ctx, `SELECT `+practCols+` FROM practitioner WHERE npi_number = $1`, npi))
}

func (r *practRepoPG) Update(ctx context.Context, p *Practitioner) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE practitioner SET
			active=$2, prefix=$3, first_name=$4, middle_name=$5, last_name=$6, suffix=$7,
			gender=$8, birth_date=$9, photo_url=$10,
			npi_number=$11, dea_number=$12, state_license_num=$13, state_license_state=$14,
			medical_council_reg=$15, abha_id=$16, hpr_id=$17,
			phone=$18, email=$19, address_line1=$20, city=$21, state=$22, postal_code=$23, country=$24,
			qualification_summary=$25, updated_at=NOW()
		WHERE id = $1`,
		p.ID, p.Active, p.Prefix, p.FirstName, p.MiddleName, p.LastName, p.Suffix,
		p.Gender, p.BirthDate, p.PhotoURL,
		p.NPINumber, p.DEANumber, p.StateLicenseNum, p.StateLicenseState,
		p.MedicalCouncilReg, p.AbhaID, p.HPRID,
		p.Phone, p.Email, p.AddressLine1, p.City, p.State, p.PostalCode, p.Country,
		p.QualificationSummary,
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
		practs = append(practs, p)
	}
	return practs, total, nil
}

func (r *practRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Practitioner, int, error) {
	query := `SELECT ` + practCols + ` FROM practitioner WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM practitioner WHERE 1=1`
	var args []interface{}
	idx := 1

	if name, ok := params["name"]; ok {
		clause := fmt.Sprintf(` AND (first_name ILIKE $%d OR last_name ILIKE $%d)`, idx, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+name+"%")
		idx++
	}
	if family, ok := params["family"]; ok {
		clause := fmt.Sprintf(` AND last_name ILIKE $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+family+"%")
		idx++
	}
	if identifier, ok := params["identifier"]; ok {
		clause := fmt.Sprintf(` AND npi_number = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, identifier)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY last_name, first_name LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
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
		&p.QualificationSummary, &p.CreatedAt, &p.UpdatedAt,
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
		&p.QualificationSummary, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type querier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}
