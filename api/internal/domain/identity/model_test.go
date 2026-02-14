package identity

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string       { return &s }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

// ---------------------------------------------------------------------------
// Patient.ToFHIR
// ---------------------------------------------------------------------------

func TestPatient_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	p := &Patient{
		ID:        uuid.New(),
		FHIRID:    "pat-001",
		Active:    true,
		MRN:       "MRN-12345",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := p.ToFHIR()

	if result["resourceType"] != "Patient" {
		t.Errorf("resourceType = %v, want Patient", result["resourceType"])
	}
	if result["id"] != "pat-001" {
		t.Errorf("id = %v, want pat-001", result["id"])
	}
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}

	// meta
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}

	// name
	names, ok := result["name"].([]fhir.HumanName)
	if !ok || len(names) == 0 {
		t.Fatal("name missing or wrong type")
	}
	if names[0].Family != "Doe" {
		t.Errorf("name[0].Family = %v, want Doe", names[0].Family)
	}
	if len(names[0].Given) < 1 || names[0].Given[0] != "John" {
		t.Errorf("name[0].Given = %v, want [John]", names[0].Given)
	}

	// identifier with MRN
	ids, ok := result["identifier"].([]fhir.Identifier)
	if !ok || len(ids) == 0 {
		t.Fatal("identifier missing or wrong type")
	}
	if ids[0].Value != "MRN-12345" {
		t.Errorf("identifier[0].Value = %v, want MRN-12345", ids[0].Value)
	}
}

func TestPatient_ToFHIR_AllFields(t *testing.T) {
	now := time.Now()
	bd := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	dd := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	orgID := uuid.New()
	pcpID := uuid.New()

	p := &Patient{
		ID:                    uuid.New(),
		FHIRID:                "pat-all",
		Active:                true,
		MRN:                   "MRN-999",
		Prefix:                ptrStr("Dr"),
		FirstName:             "Jane",
		MiddleName:            ptrStr("Marie"),
		LastName:              "Smith",
		Suffix:                ptrStr("Jr"),
		BirthDate:             ptrTime(bd),
		Gender:                ptrStr("female"),
		DeceasedBoolean:       true,
		DeceasedDatetime:      ptrTime(dd),
		PhoneMobile:           ptrStr("555-0100"),
		Email:                 ptrStr("jane@example.com"),
		AddressLine1:          ptrStr("123 Main St"),
		City:                  ptrStr("Springfield"),
		State:                 ptrStr("IL"),
		PreferredLanguage:     ptrStr("en"),
		ManagingOrgID:         ptrUUID(orgID),
		PrimaryCareProviderID: ptrUUID(pcpID),
		AbhaID:                ptrStr("ABHA-123"),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	result := p.ToFHIR()

	// gender
	if result["gender"] != "female" {
		t.Errorf("gender = %v, want female", result["gender"])
	}

	// birthDate
	if result["birthDate"] != "1990-05-15" {
		t.Errorf("birthDate = %v, want 1990-05-15", result["birthDate"])
	}

	// deceasedBoolean
	if result["deceasedBoolean"] != true {
		t.Errorf("deceasedBoolean = %v, want true", result["deceasedBoolean"])
	}

	// deceasedDateTime
	if _, ok := result["deceasedDateTime"]; !ok {
		t.Error("expected deceasedDateTime to be present")
	}

	// telecom
	telecoms, ok := result["telecom"].([]fhir.ContactPoint)
	if !ok || len(telecoms) == 0 {
		t.Fatal("telecom missing or wrong type")
	}
	foundPhone := false
	foundEmail := false
	for _, cp := range telecoms {
		if cp.System == "phone" && cp.Value == "555-0100" {
			foundPhone = true
		}
		if cp.System == "email" && cp.Value == "jane@example.com" {
			foundEmail = true
		}
	}
	if !foundPhone {
		t.Error("telecom missing phone 555-0100")
	}
	if !foundEmail {
		t.Error("telecom missing email jane@example.com")
	}

	// address
	addrs, ok := result["address"].([]fhir.Address)
	if !ok || len(addrs) == 0 {
		t.Fatal("address missing or wrong type")
	}
	if addrs[0].City != "Springfield" {
		t.Errorf("address[0].City = %v, want Springfield", addrs[0].City)
	}
	if addrs[0].State != "IL" {
		t.Errorf("address[0].State = %v, want IL", addrs[0].State)
	}

	// communication
	if _, ok := result["communication"]; !ok {
		t.Error("expected communication to be present")
	}

	// managingOrganization
	if _, ok := result["managingOrganization"]; !ok {
		t.Error("expected managingOrganization to be present")
	}

	// generalPractitioner
	gps, ok := result["generalPractitioner"].([]fhir.Reference)
	if !ok || len(gps) == 0 {
		t.Fatal("generalPractitioner missing or wrong type")
	}
	expected := "Practitioner/" + pcpID.String()
	if gps[0].Reference != expected {
		t.Errorf("generalPractitioner[0].Reference = %v, want %v", gps[0].Reference, expected)
	}

	// name with prefix, middle, suffix
	names := result["name"].([]fhir.HumanName)
	if len(names[0].Given) != 2 || names[0].Given[1] != "Marie" {
		t.Errorf("name[0].Given = %v, want [Jane Marie]", names[0].Given)
	}
	if len(names[0].Prefix) != 1 || names[0].Prefix[0] != "Dr" {
		t.Errorf("name[0].Prefix = %v, want [Dr]", names[0].Prefix)
	}
	if len(names[0].Suffix) != 1 || names[0].Suffix[0] != "Jr" {
		t.Errorf("name[0].Suffix = %v, want [Jr]", names[0].Suffix)
	}

	// ABHA identifier
	ids := result["identifier"].([]fhir.Identifier)
	foundAbha := false
	for _, id := range ids {
		if id.System == "https://healthid.ndhm.gov.in" && id.Value == "ABHA-123" {
			foundAbha = true
		}
	}
	if !foundAbha {
		t.Error("identifier missing ABHA-123")
	}
}

func TestPatient_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	p := &Patient{
		ID:        uuid.New(),
		FHIRID:    "pat-nil",
		Active:    true,
		MRN:       "MRN-000",
		FirstName: "Min",
		LastName:  "Fields",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := p.ToFHIR()

	absentKeys := []string{
		"gender", "birthDate", "telecom", "address",
		"communication", "generalPractitioner", "managingOrganization",
		"deceasedBoolean", "deceasedDateTime",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Practitioner.ToFHIR
// ---------------------------------------------------------------------------

func TestPractitioner_ToFHIR_RequiredFields(t *testing.T) {
	now := time.Now()
	p := &Practitioner{
		ID:        uuid.New(),
		FHIRID:    "prac-001",
		Active:    true,
		FirstName: "Alice",
		LastName:  "Wong",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := p.ToFHIR()

	if result["resourceType"] != "Practitioner" {
		t.Errorf("resourceType = %v, want Practitioner", result["resourceType"])
	}
	if result["id"] != "prac-001" {
		t.Errorf("id = %v, want prac-001", result["id"])
	}
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}

	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated = %v, want %v", meta.LastUpdated, now)
	}

	names, ok := result["name"].([]fhir.HumanName)
	if !ok || len(names) == 0 {
		t.Fatal("name missing or wrong type")
	}
	if names[0].Family != "Wong" {
		t.Errorf("name[0].Family = %v, want Wong", names[0].Family)
	}
	if len(names[0].Given) < 1 || names[0].Given[0] != "Alice" {
		t.Errorf("name[0].Given = %v, want [Alice]", names[0].Given)
	}
}

func TestPractitioner_ToFHIR_AllFields(t *testing.T) {
	now := time.Now()
	bd := time.Date(1980, 3, 20, 0, 0, 0, 0, time.UTC)

	p := &Practitioner{
		ID:           uuid.New(),
		FHIRID:       "prac-all",
		Active:       true,
		Prefix:       ptrStr("Dr"),
		FirstName:    "Bob",
		MiddleName:   ptrStr("Edward"),
		LastName:     "Chen",
		Suffix:       ptrStr("MD"),
		Gender:       ptrStr("male"),
		BirthDate:    ptrTime(bd),
		NPINumber:    ptrStr("NPI-999"),
		HPRID:        ptrStr("HPR-111"),
		Phone:        ptrStr("555-0200"),
		Email:        ptrStr("bob@hospital.com"),
		AddressLine1: ptrStr("456 Hospital Dr"),
		City:         ptrStr("Chicago"),
		State:        ptrStr("IL"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := p.ToFHIR()

	// gender
	if result["gender"] != "male" {
		t.Errorf("gender = %v, want male", result["gender"])
	}

	// birthDate
	if result["birthDate"] != "1980-03-20" {
		t.Errorf("birthDate = %v, want 1980-03-20", result["birthDate"])
	}

	// identifier with NPI and HPRID
	ids, ok := result["identifier"].([]fhir.Identifier)
	if !ok || len(ids) == 0 {
		t.Fatal("identifier missing or wrong type")
	}
	foundNPI := false
	foundHPR := false
	for _, id := range ids {
		if id.System == "http://hl7.org/fhir/sid/us-npi" && id.Value == "NPI-999" {
			foundNPI = true
		}
		if id.System == "https://hpr.ndhm.gov.in" && id.Value == "HPR-111" {
			foundHPR = true
		}
	}
	if !foundNPI {
		t.Error("identifier missing NPI-999")
	}
	if !foundHPR {
		t.Error("identifier missing HPR-111")
	}

	// telecom
	telecoms, ok := result["telecom"].([]fhir.ContactPoint)
	if !ok || len(telecoms) == 0 {
		t.Fatal("telecom missing or wrong type")
	}
	foundPhone := false
	foundEmail := false
	for _, cp := range telecoms {
		if cp.System == "phone" && cp.Value == "555-0200" {
			foundPhone = true
		}
		if cp.System == "email" && cp.Value == "bob@hospital.com" {
			foundEmail = true
		}
	}
	if !foundPhone {
		t.Error("telecom missing phone 555-0200")
	}
	if !foundEmail {
		t.Error("telecom missing email bob@hospital.com")
	}

	// address
	addrs, ok := result["address"].([]fhir.Address)
	if !ok || len(addrs) == 0 {
		t.Fatal("address missing or wrong type")
	}
	if addrs[0].City != "Chicago" {
		t.Errorf("address[0].City = %v, want Chicago", addrs[0].City)
	}
	if addrs[0].State != "IL" {
		t.Errorf("address[0].State = %v, want IL", addrs[0].State)
	}
	if addrs[0].Use != "work" {
		t.Errorf("address[0].Use = %v, want work", addrs[0].Use)
	}

	// name with prefix, middle, suffix
	names := result["name"].([]fhir.HumanName)
	if len(names[0].Given) != 2 || names[0].Given[1] != "Edward" {
		t.Errorf("name[0].Given = %v, want [Bob Edward]", names[0].Given)
	}
	if len(names[0].Prefix) != 1 || names[0].Prefix[0] != "Dr" {
		t.Errorf("name[0].Prefix = %v, want [Dr]", names[0].Prefix)
	}
	if len(names[0].Suffix) != 1 || names[0].Suffix[0] != "MD" {
		t.Errorf("name[0].Suffix = %v, want [MD]", names[0].Suffix)
	}
}

func TestPractitioner_ToFHIR_OptionalFieldsNil(t *testing.T) {
	now := time.Now()
	p := &Practitioner{
		ID:        uuid.New(),
		FHIRID:    "prac-nil",
		Active:    true,
		FirstName: "Min",
		LastName:  "Fields",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := p.ToFHIR()

	absentKeys := []string{
		"gender", "birthDate", "identifier", "telecom", "address",
	}
	for _, key := range absentKeys {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
