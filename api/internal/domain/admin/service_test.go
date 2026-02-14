package admin

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockOrgRepo struct {
	orgs map[uuid.UUID]*Organization
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{orgs: make(map[uuid.UUID]*Organization)}
}

func (m *mockOrgRepo) Create(_ context.Context, org *Organization) error {
	org.ID = uuid.New()
	if org.FHIRID == "" {
		org.FHIRID = org.ID.String()
	}
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepo) GetByID(_ context.Context, id uuid.UUID) (*Organization, error) {
	org, ok := m.orgs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return org, nil
}

func (m *mockOrgRepo) GetByFHIRID(_ context.Context, fhirID string) (*Organization, error) {
	for _, org := range m.orgs {
		if org.FHIRID == fhirID {
			return org, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockOrgRepo) Update(_ context.Context, org *Organization) error {
	if _, ok := m.orgs[org.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.orgs, id)
	return nil
}

func (m *mockOrgRepo) List(_ context.Context, limit, offset int) ([]*Organization, int, error) {
	var result []*Organization
	for _, org := range m.orgs {
		result = append(result, org)
	}
	total := len(result)
	if offset >= len(result) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *mockOrgRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Organization, int, error) {
	return m.List(context.Background(), limit, offset)
}

type mockDeptRepo struct {
	depts map[uuid.UUID]*Department
}

func newMockDeptRepo() *mockDeptRepo {
	return &mockDeptRepo{depts: make(map[uuid.UUID]*Department)}
}

func (m *mockDeptRepo) Create(_ context.Context, dept *Department) error {
	dept.ID = uuid.New()
	m.depts[dept.ID] = dept
	return nil
}

func (m *mockDeptRepo) GetByID(_ context.Context, id uuid.UUID) (*Department, error) {
	d, ok := m.depts[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (m *mockDeptRepo) Update(_ context.Context, dept *Department) error {
	m.depts[dept.ID] = dept
	return nil
}

func (m *mockDeptRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.depts, id)
	return nil
}

func (m *mockDeptRepo) ListByOrganization(_ context.Context, orgID uuid.UUID, limit, offset int) ([]*Department, int, error) {
	var result []*Department
	for _, d := range m.depts {
		if d.OrganizationID == orgID {
			result = append(result, d)
		}
	}
	return result, len(result), nil
}

type mockLocRepo struct {
	locs map[uuid.UUID]*Location
}

func newMockLocRepo() *mockLocRepo {
	return &mockLocRepo{locs: make(map[uuid.UUID]*Location)}
}

func (m *mockLocRepo) Create(_ context.Context, loc *Location) error {
	loc.ID = uuid.New()
	if loc.FHIRID == "" {
		loc.FHIRID = loc.ID.String()
	}
	m.locs[loc.ID] = loc
	return nil
}

func (m *mockLocRepo) GetByID(_ context.Context, id uuid.UUID) (*Location, error) {
	l, ok := m.locs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return l, nil
}

func (m *mockLocRepo) GetByFHIRID(_ context.Context, fhirID string) (*Location, error) {
	for _, l := range m.locs {
		if l.FHIRID == fhirID {
			return l, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockLocRepo) Update(_ context.Context, loc *Location) error {
	m.locs[loc.ID] = loc
	return nil
}

func (m *mockLocRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.locs, id)
	return nil
}

func (m *mockLocRepo) List(_ context.Context, limit, offset int) ([]*Location, int, error) {
	var result []*Location
	for _, l := range m.locs {
		result = append(result, l)
	}
	return result, len(result), nil
}

type mockUserRepo struct {
	users map[uuid.UUID]*SystemUser
	roles map[uuid.UUID]*UserRoleAssignment
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[uuid.UUID]*SystemUser),
		roles: make(map[uuid.UUID]*UserRoleAssignment),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *SystemUser) error {
	user.ID = uuid.New()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*SystemUser, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return u, nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*SystemUser, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockUserRepo) Update(_ context.Context, user *SystemUser) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) List(_ context.Context, limit, offset int) ([]*SystemUser, int, error) {
	var result []*SystemUser
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, len(result), nil
}

func (m *mockUserRepo) AssignRole(_ context.Context, a *UserRoleAssignment) error {
	a.ID = uuid.New()
	m.roles[a.ID] = a
	return nil
}

func (m *mockUserRepo) GetRoles(_ context.Context, userID uuid.UUID) ([]*UserRoleAssignment, error) {
	var result []*UserRoleAssignment
	for _, r := range m.roles {
		if r.UserID == userID && r.Active {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockUserRepo) RemoveRole(_ context.Context, id uuid.UUID) error {
	if r, ok := m.roles[id]; ok {
		r.Active = false
	}
	return nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockOrgRepo(), newMockDeptRepo(), newMockLocRepo(), newMockUserRepo())
}

func TestCreateOrganization(t *testing.T) {
	svc := newTestService()

	org := &Organization{Name: "Test Hospital", TypeCode: "prov"}
	err := svc.CreateOrganization(context.Background(), org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if !org.Active {
		t.Error("expected active to be true")
	}
}

func TestCreateOrganization_NameRequired(t *testing.T) {
	svc := newTestService()

	org := &Organization{TypeCode: "prov"}
	err := svc.CreateOrganization(context.Background(), org)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateOrganization_DefaultTypeCode(t *testing.T) {
	svc := newTestService()

	org := &Organization{Name: "Test"}
	err := svc.CreateOrganization(context.Background(), org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.TypeCode != "prov" {
		t.Errorf("expected default type_code 'prov', got %s", org.TypeCode)
	}
}

func TestGetOrganization(t *testing.T) {
	svc := newTestService()

	org := &Organization{Name: "Test Hospital"}
	svc.CreateOrganization(context.Background(), org)

	fetched, err := svc.GetOrganization(context.Background(), org.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "Test Hospital" {
		t.Errorf("expected name 'Test Hospital', got %s", fetched.Name)
	}
}

func TestDeleteOrganization(t *testing.T) {
	svc := newTestService()

	org := &Organization{Name: "Test"}
	svc.CreateOrganization(context.Background(), org)

	err := svc.DeleteOrganization(context.Background(), org.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetOrganization(context.Background(), org.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCreateDepartment(t *testing.T) {
	svc := newTestService()

	orgID := uuid.New()
	dept := &Department{Name: "Cardiology", OrganizationID: orgID}
	err := svc.CreateDepartment(context.Background(), dept)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dept.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateDepartment_NameRequired(t *testing.T) {
	svc := newTestService()

	dept := &Department{OrganizationID: uuid.New()}
	err := svc.CreateDepartment(context.Background(), dept)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateDepartment_OrgRequired(t *testing.T) {
	svc := newTestService()

	dept := &Department{Name: "Cardiology"}
	err := svc.CreateDepartment(context.Background(), dept)
	if err == nil {
		t.Error("expected error for missing organization_id")
	}
}

func TestCreateLocation(t *testing.T) {
	svc := newTestService()

	loc := &Location{Name: "Main Building"}
	err := svc.CreateLocation(context.Background(), loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.Status != "active" {
		t.Errorf("expected default status 'active', got %s", loc.Status)
	}
}

func TestCreateLocation_NameRequired(t *testing.T) {
	svc := newTestService()

	loc := &Location{}
	err := svc.CreateLocation(context.Background(), loc)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateSystemUser(t *testing.T) {
	svc := newTestService()

	user := &SystemUser{Username: "jdoe", UserType: "provider"}
	err := svc.CreateSystemUser(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Status != "active" {
		t.Errorf("expected default status 'active', got %s", user.Status)
	}
}

func TestCreateSystemUser_UsernameRequired(t *testing.T) {
	svc := newTestService()

	user := &SystemUser{UserType: "provider"}
	err := svc.CreateSystemUser(context.Background(), user)
	if err == nil {
		t.Error("expected error for missing username")
	}
}

func TestCreateSystemUser_TypeRequired(t *testing.T) {
	svc := newTestService()

	user := &SystemUser{Username: "jdoe"}
	err := svc.CreateSystemUser(context.Background(), user)
	if err == nil {
		t.Error("expected error for missing user_type")
	}
}

func TestAssignRole(t *testing.T) {
	svc := newTestService()

	user := &SystemUser{Username: "jdoe", UserType: "provider"}
	svc.CreateSystemUser(context.Background(), user)

	assignment := &UserRoleAssignment{
		UserID:   user.ID,
		RoleName: "physician",
	}
	err := svc.AssignRole(context.Background(), assignment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	roles, err := svc.GetUserRoles(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(roles))
	}
	if roles[0].RoleName != "physician" {
		t.Errorf("expected physician, got %s", roles[0].RoleName)
	}
}

func TestAssignRole_RoleRequired(t *testing.T) {
	svc := newTestService()

	assignment := &UserRoleAssignment{UserID: uuid.New()}
	err := svc.AssignRole(context.Background(), assignment)
	if err == nil {
		t.Error("expected error for missing role_name")
	}
}

func TestOrganizationToFHIR(t *testing.T) {
	phone := "+1-555-1234"
	email := "info@hospital.com"
	addr := "123 Main St"
	city := "Springfield"
	state := "IL"
	postal := "62704"
	country := "US"

	org := &Organization{
		FHIRID:       "org-123",
		Name:         "Springfield General",
		TypeCode:     "prov",
		Active:       true,
		Phone:        &phone,
		Email:        &email,
		AddressLine1: &addr,
		City:         &city,
		State:        &state,
		PostalCode:   &postal,
		Country:      &country,
		UpdatedAt:    time.Now(),
	}

	fhirOrg := org.ToFHIR()

	if fhirOrg["resourceType"] != "Organization" {
		t.Errorf("expected Organization, got %v", fhirOrg["resourceType"])
	}
	if fhirOrg["id"] != "org-123" {
		t.Errorf("expected org-123, got %v", fhirOrg["id"])
	}
	if fhirOrg["active"] != true {
		t.Error("expected active true")
	}
	if fhirOrg["name"] != "Springfield General" {
		t.Errorf("expected Springfield General, got %v", fhirOrg["name"])
	}
	if fhirOrg["telecom"] == nil {
		t.Error("expected telecom to be set")
	}
	if fhirOrg["address"] == nil {
		t.Error("expected address to be set")
	}
}
