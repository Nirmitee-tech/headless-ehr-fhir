package admin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string       { return &s }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }
func ptrTime(t time.Time) *time.Time { return &t }

// ---------------------------------------------------------------------------
// Organization.ToFHIR
// ---------------------------------------------------------------------------

func TestOrganization_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()

	o := &Organization{
		ID:        uuid.New(),
		FHIRID:    "org-001",
		Name:      "General Hospital",
		TypeCode:  "prov",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := o.ToFHIR()

	if result["resourceType"] != "Organization" {
		t.Errorf("resourceType = %v, want Organization", result["resourceType"])
	}
	if result["id"] != "org-001" {
		t.Errorf("id = %v, want org-001", result["id"])
	}
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}
	if result["name"] != "General Hospital" {
		t.Errorf("name = %v, want General Hospital", result["name"])
	}

	// type
	types, ok := result["type"].([]fhir.CodeableConcept)
	if !ok || len(types) == 0 {
		t.Fatal("type missing or wrong type")
	}
	if len(types[0].Coding) == 0 || types[0].Coding[0].Code != "prov" {
		t.Errorf("type[0].Coding[0].Code = %v, want prov", types[0].Coding[0].Code)
	}

	// meta
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestOrganization_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	parentID := uuid.New()

	o := &Organization{
		ID:           uuid.New(),
		FHIRID:       "org-opt",
		Name:         "Specialty Clinic",
		TypeCode:     "dept",
		Active:       true,
		ParentOrgID:  ptrUUID(parentID),
		NPINumber:    ptrStr("NPI-ORG-001"),
		Phone:        ptrStr("555-1000"),
		Email:        ptrStr("info@clinic.com"),
		AddressLine1: ptrStr("789 Health Ave"),
		AddressLine2: ptrStr("Suite 200"),
		City:         ptrStr("Boston"),
		State:        ptrStr("MA"),
		PostalCode:   ptrStr("02101"),
		Country:      ptrStr("US"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := o.ToFHIR()

	// telecom
	telecoms, ok := result["telecom"].([]fhir.ContactPoint)
	if !ok || len(telecoms) == 0 {
		t.Fatal("telecom missing or wrong type")
	}
	foundPhone := false
	foundEmail := false
	for _, cp := range telecoms {
		if cp.System == "phone" && cp.Value == "555-1000" {
			foundPhone = true
		}
		if cp.System == "email" && cp.Value == "info@clinic.com" {
			foundEmail = true
		}
	}
	if !foundPhone {
		t.Error("telecom missing phone 555-1000")
	}
	if !foundEmail {
		t.Error("telecom missing email info@clinic.com")
	}

	// address
	addrs, ok := result["address"].([]fhir.Address)
	if !ok || len(addrs) == 0 {
		t.Fatal("address missing or wrong type")
	}
	if addrs[0].City != "Boston" {
		t.Errorf("address[0].City = %v, want Boston", addrs[0].City)
	}
	if addrs[0].State != "MA" {
		t.Errorf("address[0].State = %v, want MA", addrs[0].State)
	}
	if addrs[0].PostalCode != "02101" {
		t.Errorf("address[0].PostalCode = %v, want 02101", addrs[0].PostalCode)
	}
	if addrs[0].Country != "US" {
		t.Errorf("address[0].Country = %v, want US", addrs[0].Country)
	}
	if len(addrs[0].Line) != 2 {
		t.Errorf("address[0].Line length = %d, want 2", len(addrs[0].Line))
	}

	// partOf
	partOf, ok := result["partOf"].(fhir.Reference)
	if !ok {
		t.Fatal("partOf missing or wrong type")
	}
	expectedRef := "Organization/" + parentID.String()
	if partOf.Reference != expectedRef {
		t.Errorf("partOf.Reference = %v, want %v", partOf.Reference, expectedRef)
	}
}

func TestOrganization_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	o := &Organization{
		ID:        uuid.New(),
		FHIRID:    "org-nil",
		Name:      "Minimal Org",
		TypeCode:  "prov",
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := o.ToFHIR()

	absentKeys := []string{
		"telecom", "address", "partOf",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Location.ToFHIR
// ---------------------------------------------------------------------------

func TestLocation_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()

	l := &Location{
		ID:        uuid.New(),
		FHIRID:    "loc-001",
		Status:    "active",
		Name:      "Main Building",
		CreatedAt: now,
	}

	result := l.ToFHIR()

	if result["resourceType"] != "Location" {
		t.Errorf("resourceType = %v, want Location", result["resourceType"])
	}
	if result["id"] != "loc-001" {
		t.Errorf("id = %v, want loc-001", result["id"])
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want active", result["status"])
	}
	if result["name"] != "Main Building" {
		t.Errorf("name = %v, want Main Building", result["name"])
	}

	// meta
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}
}

func TestLocation_ToFHIR_OptionalFields(t *testing.T) {
	now := time.Now()
	orgID := uuid.New()
	partOfID := uuid.New()

	l := &Location{
		ID:               uuid.New(),
		FHIRID:           "loc-opt",
		Status:           "active",
		Name:             "Room 101",
		TypeCode:         ptrStr("HOSP"),
		TypeDisplay:      ptrStr("Hospital"),
		PhysicalTypeCode: ptrStr("ro"),
		OrganizationID:   ptrUUID(orgID),
		PartOfLocationID: ptrUUID(partOfID),
		CreatedAt:        now,
	}

	result := l.ToFHIR()

	// type
	types, ok := result["type"].([]fhir.CodeableConcept)
	if !ok || len(types) == 0 {
		t.Fatal("type missing or wrong type")
	}
	if len(types[0].Coding) == 0 || types[0].Coding[0].Code != "HOSP" {
		t.Errorf("type[0].Coding[0].Code = %v, want HOSP", types[0].Coding[0].Code)
	}
	if types[0].Coding[0].Display != "Hospital" {
		t.Errorf("type[0].Coding[0].Display = %v, want Hospital", types[0].Coding[0].Display)
	}

	// physicalType
	pt, ok := result["physicalType"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("physicalType missing or wrong type")
	}
	if len(pt.Coding) == 0 || pt.Coding[0].Code != "ro" {
		t.Errorf("physicalType.Coding[0].Code = %v, want ro", pt.Coding[0].Code)
	}

	// managingOrganization
	mo, ok := result["managingOrganization"].(fhir.Reference)
	if !ok {
		t.Fatal("managingOrganization missing or wrong type")
	}
	expectedOrg := "Organization/" + orgID.String()
	if mo.Reference != expectedOrg {
		t.Errorf("managingOrganization.Reference = %v, want %v", mo.Reference, expectedOrg)
	}

	// partOf
	po, ok := result["partOf"].(fhir.Reference)
	if !ok {
		t.Fatal("partOf missing or wrong type")
	}
	expectedPartOf := "Location/" + partOfID.String()
	if po.Reference != expectedPartOf {
		t.Errorf("partOf.Reference = %v, want %v", po.Reference, expectedPartOf)
	}
}

func TestLocation_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	l := &Location{
		ID:        uuid.New(),
		FHIRID:    "loc-nil",
		Status:    "active",
		Name:      "Minimal Location",
		CreatedAt: now,
	}

	result := l.ToFHIR()

	absentKeys := []string{
		"type", "physicalType", "managingOrganization", "partOf",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Department JSON marshal/unmarshal roundtrip
// ---------------------------------------------------------------------------

func TestDepartment_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	headID := uuid.New()

	d := Department{
		ID:                 uuid.New(),
		OrganizationID:     uuid.New(),
		Name:               "Cardiology",
		Code:               ptrStr("CARD"),
		Description:        ptrStr("Cardiovascular department"),
		HeadPractitionerID: ptrUUID(headID),
		Active:             true,
		CreatedAt:          now,
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Department
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != d.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, d.ID)
	}
	if decoded.Name != "Cardiology" {
		t.Errorf("Name = %v, want Cardiology", decoded.Name)
	}
	if decoded.Code == nil || *decoded.Code != "CARD" {
		t.Errorf("Code = %v, want CARD", decoded.Code)
	}
	if decoded.Active != true {
		t.Errorf("Active = %v, want true", decoded.Active)
	}
	if decoded.HeadPractitionerID == nil || *decoded.HeadPractitionerID != headID {
		t.Errorf("HeadPractitionerID = %v, want %v", decoded.HeadPractitionerID, headID)
	}
}

// ---------------------------------------------------------------------------
// SystemUser JSON marshal/unmarshal roundtrip
// ---------------------------------------------------------------------------

func TestSystemUser_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	practID := uuid.New()
	deptID := uuid.New()

	u := SystemUser{
		ID:                  uuid.New(),
		Username:            "jdoe",
		PractitionerID:      ptrUUID(practID),
		UserType:            "physician",
		Status:              "active",
		DisplayName:         ptrStr("John Doe"),
		Email:               ptrStr("jdoe@hospital.com"),
		Phone:               ptrStr("555-0300"),
		FailedLoginCount:    0,
		MFAEnabled:          true,
		PrimaryDepartmentID: ptrUUID(deptID),
		EmployeeID:          ptrStr("EMP-001"),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded SystemUser
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != u.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, u.ID)
	}
	if decoded.Username != "jdoe" {
		t.Errorf("Username = %v, want jdoe", decoded.Username)
	}
	if decoded.UserType != "physician" {
		t.Errorf("UserType = %v, want physician", decoded.UserType)
	}
	if decoded.Status != "active" {
		t.Errorf("Status = %v, want active", decoded.Status)
	}
	if decoded.MFAEnabled != true {
		t.Errorf("MFAEnabled = %v, want true", decoded.MFAEnabled)
	}
	if decoded.DisplayName == nil || *decoded.DisplayName != "John Doe" {
		t.Errorf("DisplayName = %v, want John Doe", decoded.DisplayName)
	}
	if decoded.PractitionerID == nil || *decoded.PractitionerID != practID {
		t.Errorf("PractitionerID = %v, want %v", decoded.PractitionerID, practID)
	}
}

// ---------------------------------------------------------------------------
// UserRoleAssignment JSON marshal/unmarshal roundtrip
// ---------------------------------------------------------------------------

func TestUserRoleAssignment_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	orgID := uuid.New()
	grantedBy := uuid.New()
	endDate := now.Add(365 * 24 * time.Hour)

	r := UserRoleAssignment{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		RoleName:       "attending_physician",
		OrganizationID: ptrUUID(orgID),
		StartDate:      now,
		EndDate:        ptrTime(endDate),
		Active:         true,
		GrantedByID:    ptrUUID(grantedBy),
		CreatedAt:      now,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded UserRoleAssignment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != r.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, r.ID)
	}
	if decoded.RoleName != "attending_physician" {
		t.Errorf("RoleName = %v, want attending_physician", decoded.RoleName)
	}
	if decoded.Active != true {
		t.Errorf("Active = %v, want true", decoded.Active)
	}
	if decoded.OrganizationID == nil || *decoded.OrganizationID != orgID {
		t.Errorf("OrganizationID = %v, want %v", decoded.OrganizationID, orgID)
	}
	if decoded.GrantedByID == nil || *decoded.GrantedByID != grantedBy {
		t.Errorf("GrantedByID = %v, want %v", decoded.GrantedByID, grantedBy)
	}
	if decoded.EndDate == nil {
		t.Error("EndDate is nil, want non-nil")
	}
}
