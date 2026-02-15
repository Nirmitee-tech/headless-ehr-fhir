package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func newGroupTestService() *GroupService {
	return NewGroupService(NewInMemoryGroupRepo())
}

func newGroupTestHandler() (*GroupHandler, *echo.Echo) {
	svc := newGroupTestService()
	h := NewGroupHandler(svc)
	e := echo.New()
	return h, e
}

func createTestGroup(t *testing.T, svc *GroupService, name string, groupType GroupType) *Group {
	t.Helper()
	g := &Group{Name: name, Type: groupType, Actual: true}
	if err := svc.CreateGroup(context.Background(), g); err != nil {
		t.Fatalf("failed to create test group: %v", err)
	}
	return g
}

// --------------------------------------------------------------------------
// Model Tests
// --------------------------------------------------------------------------

func TestIsValidGroupType(t *testing.T) {
	valid := []string{"person", "animal", "practitioner", "device", "medication", "substance"}
	for _, v := range valid {
		if !IsValidGroupType(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	invalid := []string{"", "unknown", "patient", "Person", "PERSON"}
	for _, v := range invalid {
		if IsValidGroupType(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestGroup_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	g := &Group{
		ID:        uuid.New(),
		FHIRID:    "grp-001",
		Type:      GroupTypePerson,
		Actual:    true,
		Name:      "Test Group",
		Quantity:  5,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := g.ToFHIR()

	if result["resourceType"] != "Group" {
		t.Errorf("resourceType = %v, want Group", result["resourceType"])
	}
	if result["id"] != "grp-001" {
		t.Errorf("id = %v, want grp-001", result["id"])
	}
	if result["type"] != "person" {
		t.Errorf("type = %v, want person", result["type"])
	}
	if result["actual"] != true {
		t.Errorf("actual = %v, want true", result["actual"])
	}
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}
	if result["name"] != "Test Group" {
		t.Errorf("name = %v, want Test Group", result["name"])
	}
	if result["quantity"] != 5 {
		t.Errorf("quantity = %v, want 5", result["quantity"])
	}
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestGroup_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	code := "team-code"
	entity := "Organization/org-123"
	start := now.Add(-24 * time.Hour)

	g := &Group{
		ID:             uuid.New(),
		FHIRID:         "grp-opt",
		Type:           GroupTypePractitioner,
		Actual:         false,
		Code:           &code,
		Name:           "Care Team",
		Quantity:       1,
		ManagingEntity: &entity,
		Members: []GroupMember{
			{
				ID:          uuid.New(),
				EntityID:    "pract-001",
				EntityType:  "Practitioner",
				PeriodStart: &start,
				Inactive:    false,
			},
		},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := g.ToFHIR()

	// code
	codeCC, ok := result["code"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("code missing or wrong type")
	}
	if len(codeCC.Coding) == 0 || codeCC.Coding[0].Code != "team-code" {
		t.Errorf("code.coding[0].code = %v, want team-code", codeCC.Coding[0].Code)
	}

	// managingEntity
	me, ok := result["managingEntity"].(fhir.Reference)
	if !ok {
		t.Fatal("managingEntity missing or wrong type")
	}
	if me.Reference != "Organization/org-123" {
		t.Errorf("managingEntity.reference = %v, want Organization/org-123", me.Reference)
	}

	// member
	members, ok := result["member"].([]map[string]interface{})
	if !ok || len(members) == 0 {
		t.Fatal("member missing or wrong type")
	}
	memberEntity, ok := members[0]["entity"].(fhir.Reference)
	if !ok {
		t.Fatal("member[0].entity missing or wrong type")
	}
	if memberEntity.Reference != "Practitioner/pract-001" {
		t.Errorf("member[0].entity.reference = %v, want Practitioner/pract-001", memberEntity.Reference)
	}
	if members[0]["inactive"] != false {
		t.Errorf("member[0].inactive = %v, want false", members[0]["inactive"])
	}
	_, hasPeriod := members[0]["period"]
	if !hasPeriod {
		t.Error("expected member[0].period to be present")
	}
}

func TestGroup_ToFHIR_NoOptionalFields(t *testing.T) {
	now := time.Now()
	g := &Group{
		ID:        uuid.New(),
		FHIRID:    "grp-min",
		Type:      GroupTypeDevice,
		Actual:    true,
		Name:      "Minimal Group",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := g.ToFHIR()

	absentKeys := []string{"code", "managingEntity", "member"}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

func TestGroup_ToFHIR_MemberWithoutPeriod(t *testing.T) {
	now := time.Now()
	g := &Group{
		ID:     uuid.New(),
		FHIRID: "grp-nperiod",
		Type:   GroupTypePerson,
		Actual: true,
		Name:   "No Period Members",
		Members: []GroupMember{
			{
				ID:         uuid.New(),
				EntityID:   "patient-001",
				EntityType: "Patient",
				Inactive:   true,
			},
		},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := g.ToFHIR()
	members := result["member"].([]map[string]interface{})
	if _, hasPeriod := members[0]["period"]; hasPeriod {
		t.Error("expected member[0].period to be absent when no dates set")
	}
	if members[0]["inactive"] != true {
		t.Errorf("member[0].inactive = %v, want true", members[0]["inactive"])
	}
}

func TestGroupFromFHIR_FullResource(t *testing.T) {
	data := map[string]interface{}{
		"resourceType": "Group",
		"type":         "person",
		"actual":       true,
		"active":       true,
		"name":         "Test Group",
		"quantity":     float64(3),
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{"code": "test-code"},
			},
		},
		"managingEntity": map[string]interface{}{
			"reference": "Organization/org-1",
		},
		"member": []interface{}{
			map[string]interface{}{
				"entity": map[string]interface{}{
					"reference": "Patient/pat-1",
				},
				"inactive": false,
				"period": map[string]interface{}{
					"start": "2025-01-01T00:00:00Z",
				},
			},
		},
	}

	g, err := GroupFromFHIR(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Type != GroupTypePerson {
		t.Errorf("type = %v, want person", g.Type)
	}
	if g.Actual != true {
		t.Errorf("actual = %v, want true", g.Actual)
	}
	if g.Active != true {
		t.Errorf("active = %v, want true", g.Active)
	}
	if g.Name != "Test Group" {
		t.Errorf("name = %v, want Test Group", g.Name)
	}
	if g.Quantity != 3 {
		t.Errorf("quantity = %v, want 3", g.Quantity)
	}
	if g.Code == nil || *g.Code != "test-code" {
		t.Errorf("code = %v, want test-code", g.Code)
	}
	if g.ManagingEntity == nil || *g.ManagingEntity != "Organization/org-1" {
		t.Errorf("managingEntity = %v, want Organization/org-1", g.ManagingEntity)
	}
	if len(g.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(g.Members))
	}
	if g.Members[0].EntityType != "Patient" {
		t.Errorf("member entity_type = %v, want Patient", g.Members[0].EntityType)
	}
	if g.Members[0].EntityID != "pat-1" {
		t.Errorf("member entity_id = %v, want pat-1", g.Members[0].EntityID)
	}
	if g.Members[0].Inactive != false {
		t.Errorf("member inactive = %v, want false", g.Members[0].Inactive)
	}
	if g.Members[0].PeriodStart == nil {
		t.Error("expected member period_start to be set")
	}
}

func TestGroupFromFHIR_InvalidType(t *testing.T) {
	data := map[string]interface{}{
		"type": "invalid",
		"name": "Bad Group",
	}
	_, err := GroupFromFHIR(data)
	if err == nil {
		t.Error("expected error for invalid group type")
	}
}

func TestGroupFromFHIR_MinimalResource(t *testing.T) {
	data := map[string]interface{}{
		"type":   "device",
		"name":   "Devices",
		"actual": false,
	}
	g, err := GroupFromFHIR(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Type != GroupTypeDevice {
		t.Errorf("type = %v, want device", g.Type)
	}
	if g.Actual != false {
		t.Errorf("actual = %v, want false", g.Actual)
	}
	if g.Code != nil {
		t.Errorf("code should be nil, got %v", g.Code)
	}
	if g.ManagingEntity != nil {
		t.Errorf("managingEntity should be nil, got %v", g.ManagingEntity)
	}
	if len(g.Members) != 0 {
		t.Errorf("expected 0 members, got %d", len(g.Members))
	}
}

func TestGroup_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	code := "grp-code"
	entity := "Organization/org-1"
	start := now.Add(-time.Hour).Truncate(time.Second)

	g := Group{
		ID:             uuid.New(),
		FHIRID:         "grp-rt",
		Type:           GroupTypePerson,
		Actual:         true,
		Code:           &code,
		Name:           "Roundtrip Group",
		Quantity:       1,
		ManagingEntity: &entity,
		Members: []GroupMember{
			{
				ID:          uuid.New(),
				EntityID:    "pat-1",
				EntityType:  "Patient",
				PeriodStart: &start,
				Inactive:    false,
			},
		},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Group
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != g.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, g.ID)
	}
	if decoded.Name != g.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, g.Name)
	}
	if decoded.Type != g.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, g.Type)
	}
	if decoded.Actual != g.Actual {
		t.Errorf("Actual = %v, want %v", decoded.Actual, g.Actual)
	}
	if decoded.Code == nil || *decoded.Code != code {
		t.Errorf("Code = %v, want %v", decoded.Code, code)
	}
	if len(decoded.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(decoded.Members))
	}
	if decoded.Members[0].EntityID != "pat-1" {
		t.Errorf("member entity_id = %v, want pat-1", decoded.Members[0].EntityID)
	}
}

// --------------------------------------------------------------------------
// Repository Tests
// --------------------------------------------------------------------------

func TestInMemoryGroupRepo_CreateAndGet(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "Test", Type: GroupTypePerson, Actual: true}
	err := repo.Create(context.Background(), g)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if g.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if g.FHIRID == "" {
		t.Error("expected FHIRID to be set")
	}
	if g.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	fetched, err := repo.GetByID(context.Background(), g.ID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if fetched.Name != "Test" {
		t.Errorf("Name = %v, want Test", fetched.Name)
	}
}

func TestInMemoryGroupRepo_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	_, err := repo.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestInMemoryGroupRepo_Update(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "Original", Type: GroupTypePerson, Actual: true}
	repo.Create(context.Background(), g)

	g.Name = "Updated"
	err := repo.Update(context.Background(), g)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	fetched, _ := repo.GetByID(context.Background(), g.ID)
	if fetched.Name != "Updated" {
		t.Errorf("Name = %v, want Updated", fetched.Name)
	}
}

func TestInMemoryGroupRepo_Update_NotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{ID: uuid.New(), Name: "Ghost", Type: GroupTypePerson}
	err := repo.Update(context.Background(), g)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestInMemoryGroupRepo_Delete(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "ToDelete", Type: GroupTypePerson, Actual: true}
	repo.Create(context.Background(), g)

	err := repo.Delete(context.Background(), g.ID)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), g.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestInMemoryGroupRepo_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestInMemoryGroupRepo_List(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	repo.Create(context.Background(), &Group{Name: "G1", Type: GroupTypePerson, Actual: true})
	repo.Create(context.Background(), &Group{Name: "G2", Type: GroupTypeDevice, Actual: true})
	repo.Create(context.Background(), &Group{Name: "G3", Type: GroupTypePerson, Actual: true})

	// List all
	groups, total, err := repo.List(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(groups) != 3 {
		t.Errorf("len = %d, want 3", len(groups))
	}
}

func TestInMemoryGroupRepo_ListWithTypeFilter(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	repo.Create(context.Background(), &Group{Name: "G1", Type: GroupTypePerson, Actual: true})
	repo.Create(context.Background(), &Group{Name: "G2", Type: GroupTypeDevice, Actual: true})
	repo.Create(context.Background(), &Group{Name: "G3", Type: GroupTypePerson, Actual: true})

	// Filter by type
	groups, total, err := repo.List(context.Background(), "person", 10, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(groups) != 2 {
		t.Errorf("len = %d, want 2", len(groups))
	}
}

func TestInMemoryGroupRepo_ListPagination(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &Group{Name: "G", Type: GroupTypePerson, Actual: true})
	}

	groups, total, err := repo.List(context.Background(), "", 2, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(groups) != 2 {
		t.Errorf("len = %d, want 2", len(groups))
	}

	// Offset past end
	groups2, total2, err := repo.List(context.Background(), "", 10, 10)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total2 != 5 {
		t.Errorf("total = %d, want 5", total2)
	}
	if len(groups2) != 0 {
		t.Errorf("len = %d, want 0", len(groups2))
	}
}

func TestInMemoryGroupRepo_AddMember(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "G1", Type: GroupTypePerson, Actual: true}
	repo.Create(context.Background(), g)

	member := &GroupMember{EntityID: "pat-1", EntityType: "Patient"}
	err := repo.AddMember(context.Background(), g.ID, member)
	if err != nil {
		t.Fatalf("AddMember error: %v", err)
	}
	if member.ID == uuid.Nil {
		t.Error("expected member ID to be set")
	}

	fetched, _ := repo.GetByID(context.Background(), g.ID)
	if fetched.Quantity != 1 {
		t.Errorf("quantity = %d, want 1", fetched.Quantity)
	}
	if len(fetched.Members) != 1 {
		t.Errorf("len = %d, want 1", len(fetched.Members))
	}
}

func TestInMemoryGroupRepo_AddMember_NotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	member := &GroupMember{EntityID: "pat-1", EntityType: "Patient"}
	err := repo.AddMember(context.Background(), uuid.New(), member)
	if err == nil {
		t.Error("expected error for group not found")
	}
}

func TestInMemoryGroupRepo_RemoveMember(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "G1", Type: GroupTypePerson, Actual: true}
	repo.Create(context.Background(), g)

	m1 := &GroupMember{EntityID: "pat-1", EntityType: "Patient"}
	m2 := &GroupMember{EntityID: "pat-2", EntityType: "Patient"}
	repo.AddMember(context.Background(), g.ID, m1)
	repo.AddMember(context.Background(), g.ID, m2)

	err := repo.RemoveMember(context.Background(), g.ID, m1.ID)
	if err != nil {
		t.Fatalf("RemoveMember error: %v", err)
	}

	members, _ := repo.ListMembers(context.Background(), g.ID)
	if len(members) != 1 {
		t.Errorf("len = %d, want 1", len(members))
	}
	if members[0].EntityID != "pat-2" {
		t.Errorf("remaining member entity_id = %v, want pat-2", members[0].EntityID)
	}
}

func TestInMemoryGroupRepo_RemoveMember_NotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "G1", Type: GroupTypePerson, Actual: true}
	repo.Create(context.Background(), g)

	err := repo.RemoveMember(context.Background(), g.ID, uuid.New())
	if err == nil {
		t.Error("expected error for member not found")
	}
}

func TestInMemoryGroupRepo_RemoveMember_GroupNotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	err := repo.RemoveMember(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error for group not found")
	}
}

func TestInMemoryGroupRepo_ListMembers(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	g := &Group{Name: "G1", Type: GroupTypePerson, Actual: true}
	repo.Create(context.Background(), g)

	repo.AddMember(context.Background(), g.ID, &GroupMember{EntityID: "pat-1", EntityType: "Patient"})
	repo.AddMember(context.Background(), g.ID, &GroupMember{EntityID: "pat-2", EntityType: "Patient"})

	members, err := repo.ListMembers(context.Background(), g.ID)
	if err != nil {
		t.Fatalf("ListMembers error: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("len = %d, want 2", len(members))
	}
}

func TestInMemoryGroupRepo_ListMembers_GroupNotFound(t *testing.T) {
	repo := NewInMemoryGroupRepo()
	_, err := repo.ListMembers(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for group not found")
	}
}

// --------------------------------------------------------------------------
// Service Tests
// --------------------------------------------------------------------------

func TestGroupService_CreateGroup(t *testing.T) {
	svc := newGroupTestService()
	g := &Group{Name: "Test Group", Type: GroupTypePerson, Actual: true}
	err := svc.CreateGroup(context.Background(), g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if !g.Active {
		t.Error("expected active to be true")
	}
}

func TestGroupService_CreateGroup_NameRequired(t *testing.T) {
	svc := newGroupTestService()
	g := &Group{Type: GroupTypePerson}
	err := svc.CreateGroup(context.Background(), g)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestGroupService_CreateGroup_InvalidType(t *testing.T) {
	svc := newGroupTestService()
	g := &Group{Name: "Bad Type", Type: "invalid"}
	err := svc.CreateGroup(context.Background(), g)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestGroupService_GetGroup(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "Test Group", GroupTypePerson)

	fetched, err := svc.GetGroup(context.Background(), g.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "Test Group" {
		t.Errorf("Name = %v, want Test Group", fetched.Name)
	}
}

func TestGroupService_GetGroup_NotFound(t *testing.T) {
	svc := newGroupTestService()
	_, err := svc.GetGroup(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGroupService_UpdateGroup(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "Original", GroupTypePerson)

	g.Name = "Updated"
	err := svc.UpdateGroup(context.Background(), g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := svc.GetGroup(context.Background(), g.ID)
	if fetched.Name != "Updated" {
		t.Errorf("Name = %v, want Updated", fetched.Name)
	}
}

func TestGroupService_UpdateGroup_NameRequired(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "Test", GroupTypePerson)

	g.Name = ""
	err := svc.UpdateGroup(context.Background(), g)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestGroupService_UpdateGroup_InvalidType(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "Test", GroupTypePerson)

	g.Type = "invalid"
	err := svc.UpdateGroup(context.Background(), g)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestGroupService_DeleteGroup(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "ToDelete", GroupTypePerson)

	err := svc.DeleteGroup(context.Background(), g.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetGroup(context.Background(), g.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestGroupService_ListGroups(t *testing.T) {
	svc := newGroupTestService()
	createTestGroup(t, svc, "G1", GroupTypePerson)
	createTestGroup(t, svc, "G2", GroupTypeDevice)

	groups, total, err := svc.ListGroups(context.Background(), "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(groups) != 2 {
		t.Errorf("len = %d, want 2", len(groups))
	}
}

func TestGroupService_ListGroups_FilterByType(t *testing.T) {
	svc := newGroupTestService()
	createTestGroup(t, svc, "G1", GroupTypePerson)
	createTestGroup(t, svc, "G2", GroupTypeDevice)
	createTestGroup(t, svc, "G3", GroupTypePerson)

	groups, total, err := svc.ListGroups(context.Background(), "person", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(groups) != 2 {
		t.Errorf("len = %d, want 2", len(groups))
	}
}

func TestGroupService_ListGroups_InvalidTypeFilter(t *testing.T) {
	svc := newGroupTestService()
	_, _, err := svc.ListGroups(context.Background(), "invalid", 10, 0)
	if err == nil {
		t.Error("expected error for invalid type filter")
	}
}

func TestGroupService_AddMember(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "G1", GroupTypePerson)

	member := &GroupMember{EntityID: "pat-1", EntityType: "Patient"}
	err := svc.AddMember(context.Background(), g.ID, member)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	members, _ := svc.ListMembers(context.Background(), g.ID)
	if len(members) != 1 {
		t.Errorf("len = %d, want 1", len(members))
	}
}

func TestGroupService_AddMember_EntityIDRequired(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "G1", GroupTypePerson)

	member := &GroupMember{EntityType: "Patient"}
	err := svc.AddMember(context.Background(), g.ID, member)
	if err == nil {
		t.Error("expected error for missing entity_id")
	}
}

func TestGroupService_AddMember_EntityTypeRequired(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "G1", GroupTypePerson)

	member := &GroupMember{EntityID: "pat-1"}
	err := svc.AddMember(context.Background(), g.ID, member)
	if err == nil {
		t.Error("expected error for missing entity_type")
	}
}

func TestGroupService_RemoveMember(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "G1", GroupTypePerson)

	m := &GroupMember{EntityID: "pat-1", EntityType: "Patient"}
	svc.AddMember(context.Background(), g.ID, m)

	err := svc.RemoveMember(context.Background(), g.ID, m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	members, _ := svc.ListMembers(context.Background(), g.ID)
	if len(members) != 0 {
		t.Errorf("len = %d, want 0", len(members))
	}
}

func TestGroupService_ListMembers(t *testing.T) {
	svc := newGroupTestService()
	g := createTestGroup(t, svc, "G1", GroupTypePerson)

	svc.AddMember(context.Background(), g.ID, &GroupMember{EntityID: "pat-1", EntityType: "Patient"})
	svc.AddMember(context.Background(), g.ID, &GroupMember{EntityID: "pat-2", EntityType: "Patient"})

	members, err := svc.ListMembers(context.Background(), g.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("len = %d, want 2", len(members))
	}
}

// --------------------------------------------------------------------------
// Handler Tests - Operational API
// --------------------------------------------------------------------------

func TestGroupHandler_CreateGroup(t *testing.T) {
	h, e := newGroupTestHandler()

	body := `{"name":"Test Group","type":"person","actual":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var g Group
	json.Unmarshal(rec.Body.Bytes(), &g)
	if g.Name != "Test Group" {
		t.Errorf("name = %v, want Test Group", g.Name)
	}
	if g.Type != GroupTypePerson {
		t.Errorf("type = %v, want person", g.Type)
	}
}

func TestGroupHandler_CreateGroup_BadRequest(t *testing.T) {
	h, e := newGroupTestHandler()

	body := `{"type":"person"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGroup(c)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestGroupHandler_GetGroup(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "Test", GroupTypePerson)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.GetGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGroupHandler_GetGroup_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetGroup(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGroupHandler_GetGroup_InvalidID(t *testing.T) {
	h, e := newGroupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetGroup(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
}

func TestGroupHandler_UpdateGroup(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "Original", GroupTypePerson)

	body := `{"name":"Updated","type":"person","actual":true}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.UpdateGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var updated Group
	json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated.Name != "Updated" {
		t.Errorf("name = %v, want Updated", updated.Name)
	}
}

func TestGroupHandler_UpdateGroup_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	body := `{"name":"Updated","type":"person","actual":true}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.UpdateGroup(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGroupHandler_DeleteGroup(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "ToDelete", GroupTypePerson)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.DeleteGroup(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestGroupHandler_DeleteGroup_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.DeleteGroup(c)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestGroupHandler_ListGroups(t *testing.T) {
	h, e := newGroupTestHandler()
	createTestGroup(t, h.svc, "G1", GroupTypePerson)
	createTestGroup(t, h.svc, "G2", GroupTypeDevice)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListGroups(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGroupHandler_ListGroups_FilterByType(t *testing.T) {
	h, e := newGroupTestHandler()
	createTestGroup(t, h.svc, "G1", GroupTypePerson)
	createTestGroup(t, h.svc, "G2", GroupTypeDevice)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?type=person", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListGroups(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp pagination.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}

func TestGroupHandler_AddMember(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "G1", GroupTypePerson)

	body := `{"entity_id":"pat-1","entity_type":"Patient"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.AddMember(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestGroupHandler_AddMember_MissingEntityID(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "G1", GroupTypePerson)

	body := `{"entity_type":"Patient"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.AddMember(c)
	if err == nil {
		t.Error("expected error for missing entity_id")
	}
}

func TestGroupHandler_RemoveMember(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "G1", GroupTypePerson)

	m := &GroupMember{EntityID: "pat-1", EntityType: "Patient"}
	h.svc.AddMember(context.Background(), g.ID, m)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "member_id")
	c.SetParamValues(g.ID.String(), m.ID.String())

	err := h.RemoveMember(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestGroupHandler_RemoveMember_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "G1", GroupTypePerson)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "member_id")
	c.SetParamValues(g.ID.String(), uuid.New().String())

	err := h.RemoveMember(c)
	if err == nil {
		t.Error("expected error for member not found")
	}
}

func TestGroupHandler_ListMembers(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "G1", GroupTypePerson)

	h.svc.AddMember(context.Background(), g.ID, &GroupMember{EntityID: "pat-1", EntityType: "Patient"})
	h.svc.AddMember(context.Background(), g.ID, &GroupMember{EntityID: "pat-2", EntityType: "Patient"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.ListMembers(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var members []GroupMember
	json.Unmarshal(rec.Body.Bytes(), &members)
	if len(members) != 2 {
		t.Errorf("len = %d, want 2", len(members))
	}
}

func TestGroupHandler_ListMembers_GroupNotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.ListMembers(c)
	if err == nil {
		t.Error("expected error for group not found")
	}
}

// --------------------------------------------------------------------------
// Handler Tests - FHIR API
// --------------------------------------------------------------------------

func TestGroupHandler_SearchGroupsFHIR(t *testing.T) {
	h, e := newGroupTestHandler()
	createTestGroup(t, h.svc, "G1", GroupTypePerson)

	req := httptest.NewRequest(http.MethodGet, "/fhir/Group", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SearchGroupsFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("resourceType = %v, want Bundle", bundle["resourceType"])
	}
}

func TestGroupHandler_GetGroupFHIR(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "G1", GroupTypePerson)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.GetGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resource map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resource)
	if resource["resourceType"] != "Group" {
		t.Errorf("resourceType = %v, want Group", resource["resourceType"])
	}
}

func TestGroupHandler_GetGroupFHIR_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.GetGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGroupHandler_CreateGroupFHIR(t *testing.T) {
	h, e := newGroupTestHandler()

	body := `{"resourceType":"Group","type":"practitioner","actual":true,"name":"FHIR Group","quantity":0}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Group", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Error("expected Location header")
	}

	var resource map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resource)
	if resource["resourceType"] != "Group" {
		t.Errorf("resourceType = %v, want Group", resource["resourceType"])
	}
	if resource["type"] != "practitioner" {
		t.Errorf("type = %v, want practitioner", resource["type"])
	}
}

func TestGroupHandler_CreateGroupFHIR_InvalidType(t *testing.T) {
	h, e := newGroupTestHandler()

	body := `{"resourceType":"Group","type":"invalid","name":"Bad"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Group", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGroupHandler_UpdateGroupFHIR(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "Original", GroupTypePerson)

	body := `{"resourceType":"Group","type":"person","actual":false,"name":"Updated FHIR"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.UpdateGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resource map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resource)
	if resource["name"] != "Updated FHIR" {
		t.Errorf("name = %v, want Updated FHIR", resource["name"])
	}
}

func TestGroupHandler_UpdateGroupFHIR_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	body := `{"resourceType":"Group","type":"person","name":"Ghost"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.UpdateGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGroupHandler_DeleteGroupFHIR(t *testing.T) {
	h, e := newGroupTestHandler()
	g := createTestGroup(t, h.svc, "ToDelete", GroupTypePerson)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID.String())

	err := h.DeleteGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestGroupHandler_DeleteGroupFHIR_NotFound(t *testing.T) {
	h, e := newGroupTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.DeleteGroupFHIR(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGroupHandler_RegisterGroupRoutes(t *testing.T) {
	h, e := newGroupTestHandler()
	api := e.Group("/api/v1")
	fhirG := e.Group("/fhir")

	h.RegisterGroupRoutes(api, fhirG)

	routes := e.Routes()
	if len(routes) == 0 {
		t.Error("expected routes to be registered")
	}

	routePaths := make(map[string]bool)
	for _, r := range routes {
		routePaths[r.Method+":"+r.Path] = true
	}

	expected := []string{
		"GET:/api/v1/groups",
		"POST:/api/v1/groups",
		"GET:/api/v1/groups/:id",
		"PUT:/api/v1/groups/:id",
		"DELETE:/api/v1/groups/:id",
		"POST:/api/v1/groups/:id/members",
		"DELETE:/api/v1/groups/:id/members/:member_id",
		"GET:/api/v1/groups/:id/members",
		"GET:/fhir/Group",
		"POST:/fhir/Group",
		"GET:/fhir/Group/:id",
		"PUT:/fhir/Group/:id",
		"DELETE:/fhir/Group/:id",
	}
	for _, path := range expected {
		if !routePaths[path] {
			t.Errorf("missing expected route: %s", path)
		}
	}
}

// --------------------------------------------------------------------------
// FHIR Roundtrip Test
// --------------------------------------------------------------------------

func TestGroup_FHIRRoundtrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	code := "care-team"
	entity := "Organization/org-1"
	start := now.Add(-24 * time.Hour)

	original := &Group{
		ID:             uuid.New(),
		FHIRID:         "grp-rt",
		Type:           GroupTypePractitioner,
		Actual:         true,
		Code:           &code,
		Name:           "Care Team Alpha",
		Quantity:       1,
		ManagingEntity: &entity,
		Members: []GroupMember{
			{
				ID:          uuid.New(),
				EntityID:    "pract-1",
				EntityType:  "Practitioner",
				PeriodStart: &start,
				Inactive:    false,
			},
		},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	fhirMap := original.ToFHIR()

	// Serialize to JSON and back to map (simulating FHIR transmission)
	jsonBytes, err := json.Marshal(fhirMap)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify key FHIR fields survived the roundtrip
	if decoded["resourceType"] != "Group" {
		t.Errorf("resourceType = %v, want Group", decoded["resourceType"])
	}
	if decoded["type"] != "practitioner" {
		t.Errorf("type = %v, want practitioner", decoded["type"])
	}
	if decoded["actual"] != true {
		t.Errorf("actual = %v, want true", decoded["actual"])
	}
	if decoded["name"] != "Care Team Alpha" {
		t.Errorf("name = %v, want Care Team Alpha", decoded["name"])
	}

	// Parse back to domain from the FHIR map
	parsed, err := GroupFromFHIR(decoded)
	if err != nil {
		t.Fatalf("GroupFromFHIR error: %v", err)
	}
	if parsed.Type != original.Type {
		t.Errorf("parsed type = %v, want %v", parsed.Type, original.Type)
	}
	if parsed.Actual != original.Actual {
		t.Errorf("parsed actual = %v, want %v", parsed.Actual, original.Actual)
	}
	if parsed.Name != original.Name {
		t.Errorf("parsed name = %v, want %v", parsed.Name, original.Name)
	}
	if parsed.Code == nil || *parsed.Code != *original.Code {
		t.Errorf("parsed code = %v, want %v", parsed.Code, original.Code)
	}
	if len(parsed.Members) != 1 {
		t.Fatalf("parsed members = %d, want 1", len(parsed.Members))
	}
	if parsed.Members[0].EntityType != "Practitioner" {
		t.Errorf("parsed member entity_type = %v, want Practitioner", parsed.Members[0].EntityType)
	}
	if parsed.Members[0].EntityID != "pract-1" {
		t.Errorf("parsed member entity_id = %v, want pract-1", parsed.Members[0].EntityID)
	}
}
