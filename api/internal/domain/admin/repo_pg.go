package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/platform/db"
)

// -- Organization Repository --

type orgRepoPG struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepo(pool *pgxpool.Pool) OrganizationRepository {
	return &orgRepoPG{pool: pool}
}

func (r *orgRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *orgRepoPG) Create(ctx context.Context, org *Organization) error {
	org.ID = uuid.New()
	if org.FHIRID == "" {
		org.FHIRID = org.ID.String()
	}

	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO organization (
			id, fhir_id, name, type_code, active, parent_org_id,
			npi_number, tin_number, clia_number,
			rohini_id, abdm_facility_id, nabh_accreditation,
			address_line1, address_line2, city, district, state, postal_code, country,
			phone, email, website
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22
		)`,
		org.ID, org.FHIRID, org.Name, org.TypeCode, org.Active, org.ParentOrgID,
		org.NPINumber, org.TINNumber, org.CLIANumber,
		org.RohiniID, org.ABDMFacilityID, org.NABHAccred,
		org.AddressLine1, org.AddressLine2, org.City, org.District, org.State, org.PostalCode, org.Country,
		org.Phone, org.Email, org.Website,
	)
	return err
}

func (r *orgRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	return r.scanOrg(r.conn(ctx).QueryRow(ctx, `SELECT `+orgColumns+` FROM organization WHERE id = $1`, id))
}

func (r *orgRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Organization, error) {
	return r.scanOrg(r.conn(ctx).QueryRow(ctx, `SELECT `+orgColumns+` FROM organization WHERE fhir_id = $1`, fhirID))
}

func (r *orgRepoPG) Update(ctx context.Context, org *Organization) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE organization SET
			name = $2, type_code = $3, active = $4, parent_org_id = $5,
			npi_number = $6, tin_number = $7, clia_number = $8,
			rohini_id = $9, abdm_facility_id = $10, nabh_accreditation = $11,
			address_line1 = $12, address_line2 = $13, city = $14, district = $15,
			state = $16, postal_code = $17, country = $18,
			phone = $19, email = $20, website = $21, updated_at = NOW()
		WHERE id = $1`,
		org.ID, org.Name, org.TypeCode, org.Active, org.ParentOrgID,
		org.NPINumber, org.TINNumber, org.CLIANumber,
		org.RohiniID, org.ABDMFacilityID, org.NABHAccred,
		org.AddressLine1, org.AddressLine2, org.City, org.District,
		org.State, org.PostalCode, org.Country,
		org.Phone, org.Email, org.Website,
	)
	return err
}

func (r *orgRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM organization WHERE id = $1`, id)
	return err
}

func (r *orgRepoPG) List(ctx context.Context, limit, offset int) ([]*Organization, int, error) {
	var total int
	err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM organization`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx,
		`SELECT `+orgColumns+` FROM organization ORDER BY name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org, err := r.scanOrgRow(rows)
		if err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, org)
	}
	return orgs, total, nil
}

func (r *orgRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Organization, int, error) {
	query := `SELECT ` + orgColumns + ` FROM organization WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM organization WHERE 1=1`
	var args []interface{}
	idx := 1

	if name, ok := params["name"]; ok {
		clause := fmt.Sprintf(` AND name ILIKE $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+name+"%")
		idx++
	}
	if typeCode, ok := params["type"]; ok {
		clause := fmt.Sprintf(` AND type_code = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, typeCode)
		idx++
	}
	if active, ok := params["active"]; ok {
		clause := fmt.Sprintf(` AND active = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, active == "true")
		idx++
	}

	var total int
	err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		org, err := r.scanOrgRow(rows)
		if err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, org)
	}
	return orgs, total, nil
}

const orgColumns = `id, fhir_id, name, type_code, active, parent_org_id,
	npi_number, tin_number, clia_number,
	rohini_id, abdm_facility_id, nabh_accreditation,
	address_line1, address_line2, city, district, state, postal_code, country,
	phone, email, website, created_at, updated_at`

func (r *orgRepoPG) scanOrg(row pgx.Row) (*Organization, error) {
	var o Organization
	err := row.Scan(
		&o.ID, &o.FHIRID, &o.Name, &o.TypeCode, &o.Active, &o.ParentOrgID,
		&o.NPINumber, &o.TINNumber, &o.CLIANumber,
		&o.RohiniID, &o.ABDMFacilityID, &o.NABHAccred,
		&o.AddressLine1, &o.AddressLine2, &o.City, &o.District, &o.State, &o.PostalCode, &o.Country,
		&o.Phone, &o.Email, &o.Website, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orgRepoPG) scanOrgRow(rows pgx.Rows) (*Organization, error) {
	var o Organization
	err := rows.Scan(
		&o.ID, &o.FHIRID, &o.Name, &o.TypeCode, &o.Active, &o.ParentOrgID,
		&o.NPINumber, &o.TINNumber, &o.CLIANumber,
		&o.RohiniID, &o.ABDMFacilityID, &o.NABHAccred,
		&o.AddressLine1, &o.AddressLine2, &o.City, &o.District, &o.State, &o.PostalCode, &o.Country,
		&o.Phone, &o.Email, &o.Website, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// queryable abstracts pgxpool.Pool and pgxpool.Conn for tenant-scoped queries.
type queryable interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// -- Department Repository --

type deptRepoPG struct {
	pool *pgxpool.Pool
}

func NewDepartmentRepo(pool *pgxpool.Pool) DepartmentRepository {
	return &deptRepoPG{pool: pool}
}

func (r *deptRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

func (r *deptRepoPG) Create(ctx context.Context, dept *Department) error {
	dept.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO department (id, organization_id, name, code, description, head_practitioner_id, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		dept.ID, dept.OrganizationID, dept.Name, dept.Code, dept.Description, dept.HeadPractitionerID, dept.Active,
	)
	return err
}

func (r *deptRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Department, error) {
	var d Department
	err := r.conn(ctx).QueryRow(ctx, `
		SELECT id, organization_id, name, code, description, head_practitioner_id, active, created_at
		FROM department WHERE id = $1`, id).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Code, &d.Description, &d.HeadPractitionerID, &d.Active, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deptRepoPG) Update(ctx context.Context, dept *Department) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE department SET
			name = $2, code = $3, description = $4, head_practitioner_id = $5, active = $6
		WHERE id = $1`,
		dept.ID, dept.Name, dept.Code, dept.Description, dept.HeadPractitionerID, dept.Active,
	)
	return err
}

func (r *deptRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM department WHERE id = $1`, id)
	return err
}

func (r *deptRepoPG) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*Department, int, error) {
	var total int
	err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM department WHERE organization_id = $1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, organization_id, name, code, description, head_practitioner_id, active, created_at
		FROM department WHERE organization_id = $1 ORDER BY name LIMIT $2 OFFSET $3`, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var depts []*Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.Name, &d.Code, &d.Description, &d.HeadPractitionerID, &d.Active, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		depts = append(depts, &d)
	}
	return depts, total, nil
}

// -- Location Repository --

type locRepoPG struct {
	pool *pgxpool.Pool
}

func NewLocationRepo(pool *pgxpool.Pool) LocationRepository {
	return &locRepoPG{pool: pool}
}

func (r *locRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const locColumns = `id, fhir_id, status, operational_status, name, description, mode,
	type_code, type_display, physical_type_code,
	organization_id, part_of_location_id,
	address_line1, city, state, postal_code, country,
	latitude, longitude, phone, email, created_at`

func (r *locRepoPG) Create(ctx context.Context, loc *Location) error {
	loc.ID = uuid.New()
	if loc.FHIRID == "" {
		loc.FHIRID = loc.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO location (
			id, fhir_id, status, operational_status, name, description, mode,
			type_code, type_display, physical_type_code,
			organization_id, part_of_location_id,
			address_line1, city, state, postal_code, country,
			latitude, longitude, phone, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		loc.ID, loc.FHIRID, loc.Status, loc.OperationalStatus, loc.Name, loc.Description, loc.Mode,
		loc.TypeCode, loc.TypeDisplay, loc.PhysicalTypeCode,
		loc.OrganizationID, loc.PartOfLocationID,
		loc.AddressLine1, loc.City, loc.State, loc.PostalCode, loc.Country,
		loc.Latitude, loc.Longitude, loc.Phone, loc.Email,
	)
	return err
}

func (r *locRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Location, error) {
	return r.scanLoc(r.conn(ctx).QueryRow(ctx, `SELECT `+locColumns+` FROM location WHERE id = $1`, id))
}

func (r *locRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Location, error) {
	return r.scanLoc(r.conn(ctx).QueryRow(ctx, `SELECT `+locColumns+` FROM location WHERE fhir_id = $1`, fhirID))
}

func (r *locRepoPG) Update(ctx context.Context, loc *Location) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE location SET
			status=$2, operational_status=$3, name=$4, description=$5, mode=$6,
			type_code=$7, type_display=$8, physical_type_code=$9,
			organization_id=$10, part_of_location_id=$11,
			address_line1=$12, city=$13, state=$14, postal_code=$15, country=$16,
			latitude=$17, longitude=$18, phone=$19, email=$20
		WHERE id = $1`,
		loc.ID, loc.Status, loc.OperationalStatus, loc.Name, loc.Description, loc.Mode,
		loc.TypeCode, loc.TypeDisplay, loc.PhysicalTypeCode,
		loc.OrganizationID, loc.PartOfLocationID,
		loc.AddressLine1, loc.City, loc.State, loc.PostalCode, loc.Country,
		loc.Latitude, loc.Longitude, loc.Phone, loc.Email,
	)
	return err
}

func (r *locRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM location WHERE id = $1`, id)
	return err
}

func (r *locRepoPG) List(ctx context.Context, limit, offset int) ([]*Location, int, error) {
	var total int
	err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM location`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, `SELECT `+locColumns+` FROM location ORDER BY name LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var locs []*Location
	for rows.Next() {
		loc, err := r.scanLocRow(rows)
		if err != nil {
			return nil, 0, err
		}
		locs = append(locs, loc)
	}
	return locs, total, nil
}

func (r *locRepoPG) scanLoc(row pgx.Row) (*Location, error) {
	var l Location
	err := row.Scan(
		&l.ID, &l.FHIRID, &l.Status, &l.OperationalStatus, &l.Name, &l.Description, &l.Mode,
		&l.TypeCode, &l.TypeDisplay, &l.PhysicalTypeCode,
		&l.OrganizationID, &l.PartOfLocationID,
		&l.AddressLine1, &l.City, &l.State, &l.PostalCode, &l.Country,
		&l.Latitude, &l.Longitude, &l.Phone, &l.Email, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *locRepoPG) scanLocRow(rows pgx.Rows) (*Location, error) {
	var l Location
	err := rows.Scan(
		&l.ID, &l.FHIRID, &l.Status, &l.OperationalStatus, &l.Name, &l.Description, &l.Mode,
		&l.TypeCode, &l.TypeDisplay, &l.PhysicalTypeCode,
		&l.OrganizationID, &l.PartOfLocationID,
		&l.AddressLine1, &l.City, &l.State, &l.PostalCode, &l.Country,
		&l.Latitude, &l.Longitude, &l.Phone, &l.Email, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// -- SystemUser Repository --

type userRepoPG struct {
	pool *pgxpool.Pool
}

func NewSystemUserRepo(pool *pgxpool.Pool) SystemUserRepository {
	return &userRepoPG{pool: pool}
}

func (r *userRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const userColumns = `id, username, practitioner_id, user_type, status,
	display_name, email, phone, last_login, failed_login_count,
	password_last_changed, mfa_enabled, primary_department_id,
	employee_id, hire_date, termination_date,
	hipaa_training_date, last_compliance_training, note,
	created_at, updated_at`

func (r *userRepoPG) Create(ctx context.Context, user *SystemUser) error {
	user.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO "system_user" (
			id, username, practitioner_id, user_type, status,
			display_name, email, phone, mfa_enabled, primary_department_id,
			employee_id, hire_date, note
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		user.ID, user.Username, user.PractitionerID, user.UserType, user.Status,
		user.DisplayName, user.Email, user.Phone, user.MFAEnabled, user.PrimaryDepartmentID,
		user.EmployeeID, user.HireDate, user.Note,
	)
	return err
}

func (r *userRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*SystemUser, error) {
	return r.scanUser(r.conn(ctx).QueryRow(ctx, `SELECT `+userColumns+` FROM "system_user" WHERE id = $1`, id))
}

func (r *userRepoPG) GetByUsername(ctx context.Context, username string) (*SystemUser, error) {
	return r.scanUser(r.conn(ctx).QueryRow(ctx, `SELECT `+userColumns+` FROM "system_user" WHERE username = $1`, username))
}

func (r *userRepoPG) Update(ctx context.Context, user *SystemUser) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE "system_user" SET
			username=$2, practitioner_id=$3, user_type=$4, status=$5,
			display_name=$6, email=$7, phone=$8, mfa_enabled=$9,
			primary_department_id=$10, employee_id=$11, note=$12, updated_at=NOW()
		WHERE id = $1`,
		user.ID, user.Username, user.PractitionerID, user.UserType, user.Status,
		user.DisplayName, user.Email, user.Phone, user.MFAEnabled,
		user.PrimaryDepartmentID, user.EmployeeID, user.Note,
	)
	return err
}

func (r *userRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM "system_user" WHERE id = $1`, id)
	return err
}

func (r *userRepoPG) List(ctx context.Context, limit, offset int) ([]*SystemUser, int, error) {
	var total int
	err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM "system_user"`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn(ctx).Query(ctx, `SELECT `+userColumns+` FROM "system_user" ORDER BY username LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*SystemUser
	for rows.Next() {
		u, err := r.scanUserRow(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *userRepoPG) AssignRole(ctx context.Context, a *UserRoleAssignment) error {
	a.ID = uuid.New()
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO user_role_assignment (id, user_id, role_name, organization_id, department_id, location_id, start_date, end_date, active, granted_by_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		a.ID, a.UserID, a.RoleName, a.OrganizationID, a.DepartmentID, a.LocationID,
		a.StartDate, a.EndDate, a.Active, a.GrantedByID,
	)
	return err
}

func (r *userRepoPG) GetRoles(ctx context.Context, userID uuid.UUID) ([]*UserRoleAssignment, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, user_id, role_name, organization_id, department_id, location_id, start_date, end_date, active, granted_by_id, created_at
		FROM user_role_assignment WHERE user_id = $1 AND active = true`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*UserRoleAssignment
	for rows.Next() {
		var a UserRoleAssignment
		if err := rows.Scan(&a.ID, &a.UserID, &a.RoleName, &a.OrganizationID, &a.DepartmentID, &a.LocationID,
			&a.StartDate, &a.EndDate, &a.Active, &a.GrantedByID, &a.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, &a)
	}
	return roles, nil
}

func (r *userRepoPG) RemoveRole(ctx context.Context, assignmentID uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `UPDATE user_role_assignment SET active = false, end_date = NOW() WHERE id = $1`, assignmentID)
	return err
}

func (r *userRepoPG) scanUser(row pgx.Row) (*SystemUser, error) {
	var u SystemUser
	err := row.Scan(
		&u.ID, &u.Username, &u.PractitionerID, &u.UserType, &u.Status,
		&u.DisplayName, &u.Email, &u.Phone, &u.LastLogin, &u.FailedLoginCount,
		&u.PasswordLastChanged, &u.MFAEnabled, &u.PrimaryDepartmentID,
		&u.EmployeeID, &u.HireDate, &u.TerminationDate,
		&u.HIPAATrainingDate, &u.LastComplianceTraining, &u.Note,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepoPG) scanUserRow(rows pgx.Rows) (*SystemUser, error) {
	var u SystemUser
	err := rows.Scan(
		&u.ID, &u.Username, &u.PractitionerID, &u.UserType, &u.Status,
		&u.DisplayName, &u.Email, &u.Phone, &u.LastLogin, &u.FailedLoginCount,
		&u.PasswordLastChanged, &u.MFAEnabled, &u.PrimaryDepartmentID,
		&u.EmployeeID, &u.HireDate, &u.TerminationDate,
		&u.HIPAATrainingDate, &u.LastComplianceTraining, &u.Note,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// -- Group Repository --

type groupRepoPG struct {
	pool *pgxpool.Pool
}

// NewGroupRepo creates a new Postgres-backed GroupRepository.
func NewGroupRepo(pool *pgxpool.Pool) GroupRepository {
	return &groupRepoPG{pool: pool}
}

func (r *groupRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const grpColumns = `id, fhir_id, group_type, actual, active, name, code, quantity, managing_entity, version_id, created_at, updated_at`

func (r *groupRepoPG) Create(ctx context.Context, group *Group) error {
	group.ID = uuid.New()
	if group.FHIRID == "" {
		group.FHIRID = group.ID.String()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO fhir_group (
			id, fhir_id, group_type, actual, active, name, code, quantity, managing_entity
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		group.ID, group.FHIRID, string(group.Type), group.Actual, group.Active,
		group.Name, group.Code, group.Quantity, group.ManagingEntity,
	)
	if err != nil {
		return err
	}
	// Insert members
	for i := range group.Members {
		m := &group.Members[i]
		if m.ID == uuid.Nil {
			m.ID = uuid.New()
		}
		_, err := r.conn(ctx).Exec(ctx, `
			INSERT INTO fhir_group_member (id, group_id, entity_type, entity_id, period_start, period_end, inactive)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			m.ID, group.ID, m.EntityType, m.EntityID, m.PeriodStart, m.PeriodEnd, m.Inactive,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *groupRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	g, err := r.scanGroup(r.conn(ctx).QueryRow(ctx, `SELECT `+grpColumns+` FROM fhir_group WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	members, err := r.ListMembers(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.Members = members
	return g, nil
}

func (r *groupRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*Group, error) {
	g, err := r.scanGroup(r.conn(ctx).QueryRow(ctx, `SELECT `+grpColumns+` FROM fhir_group WHERE fhir_id = $1`, fhirID))
	if err != nil {
		return nil, err
	}
	members, err := r.ListMembers(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	g.Members = members
	return g, nil
}

func (r *groupRepoPG) Update(ctx context.Context, group *Group) error {
	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE fhir_group SET
			group_type=$2, actual=$3, active=$4, name=$5, code=$6,
			quantity=$7, managing_entity=$8, version_id=version_id+1, updated_at=NOW()
		WHERE id = $1`,
		group.ID, string(group.Type), group.Actual, group.Active, group.Name,
		group.Code, group.Quantity, group.ManagingEntity,
	)
	return err
}

func (r *groupRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM fhir_group WHERE id = $1`, id)
	return err
}

func (r *groupRepoPG) List(ctx context.Context, filterType string, limit, offset int) ([]*Group, int, error) {
	var total int
	var args []interface{}
	countQuery := `SELECT COUNT(*) FROM fhir_group`
	query := `SELECT ` + grpColumns + ` FROM fhir_group`

	if filterType != "" {
		countQuery += ` WHERE group_type = $1`
		query += ` WHERE group_type = $1`
		args = append(args, filterType)
	}

	err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	idx := len(args) + 1
	query += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		g, err := r.scanGroupRow(rows)
		if err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	return groups, total, nil
}

func (r *groupRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Group, int, error) {
	query := `SELECT ` + grpColumns + ` FROM fhir_group WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM fhir_group WHERE 1=1`
	var args []interface{}
	idx := 1

	if name, ok := params["name"]; ok {
		clause := fmt.Sprintf(` AND name ILIKE $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, "%"+name+"%")
		idx++
	}
	if t, ok := params["type"]; ok {
		clause := fmt.Sprintf(` AND group_type = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, t)
		idx++
	}
	if active, ok := params["active"]; ok {
		clause := fmt.Sprintf(` AND active = $%d`, idx)
		query += clause
		countQuery += clause
		args = append(args, active == "true")
		idx++
	}

	var total int
	err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		g, err := r.scanGroupRow(rows)
		if err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	return groups, total, nil
}

func (r *groupRepoPG) AddMember(ctx context.Context, groupID uuid.UUID, member *GroupMember) error {
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}
	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO fhir_group_member (id, group_id, entity_type, entity_id, period_start, period_end, inactive)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		member.ID, groupID, member.EntityType, member.EntityID, member.PeriodStart, member.PeriodEnd, member.Inactive,
	)
	if err != nil {
		return err
	}
	// Update quantity
	_, err = r.conn(ctx).Exec(ctx, `
		UPDATE fhir_group SET quantity = (SELECT COUNT(*) FROM fhir_group_member WHERE group_id = $1), updated_at = NOW()
		WHERE id = $1`, groupID)
	return err
}

func (r *groupRepoPG) RemoveMember(ctx context.Context, groupID uuid.UUID, memberID uuid.UUID) error {
	tag, err := r.conn(ctx).Exec(ctx, `DELETE FROM fhir_group_member WHERE id = $1 AND group_id = $2`, memberID, groupID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	// Update quantity
	_, err = r.conn(ctx).Exec(ctx, `
		UPDATE fhir_group SET quantity = (SELECT COUNT(*) FROM fhir_group_member WHERE group_id = $1), updated_at = NOW()
		WHERE id = $1`, groupID)
	return err
}

func (r *groupRepoPG) ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMember, error) {
	rows, err := r.conn(ctx).Query(ctx, `
		SELECT id, entity_type, entity_id, period_start, period_end, inactive
		FROM fhir_group_member WHERE group_id = $1 ORDER BY created_at`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []GroupMember
	for rows.Next() {
		var m GroupMember
		if err := rows.Scan(&m.ID, &m.EntityType, &m.EntityID, &m.PeriodStart, &m.PeriodEnd, &m.Inactive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *groupRepoPG) scanGroup(row pgx.Row) (*Group, error) {
	var g Group
	var groupType string
	err := row.Scan(
		&g.ID, &g.FHIRID, &groupType, &g.Actual, &g.Active,
		&g.Name, &g.Code, &g.Quantity, &g.ManagingEntity,
		&g.VersionID, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	g.Type = GroupType(groupType)
	return &g, nil
}

func (r *groupRepoPG) scanGroupRow(rows pgx.Rows) (*Group, error) {
	var g Group
	var groupType string
	err := rows.Scan(
		&g.ID, &g.FHIRID, &groupType, &g.Actual, &g.Active,
		&g.Name, &g.Code, &g.Quantity, &g.ManagingEntity,
		&g.VersionID, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	g.Type = GroupType(groupType)
	return &g, nil
}
