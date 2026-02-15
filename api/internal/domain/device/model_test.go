package device

import (
	"testing"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

func ptrStr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }

func TestDevice_ToFHIR(t *testing.T) {
	now := time.Now()
	patID := uuid.New()
	d := &Device{
		ID: uuid.New(), FHIRID: "dev-001", Status: "active",
		DeviceName: "Pulse Oximeter", DeviceNameType: "manufacturer-name",
		ManufacturerName: ptrStr("Acme Medical"),
		TypeCode: ptrStr("59181002"), TypeDisplay: ptrStr("Oxygen analyzer"),
		TypeSystem: ptrStr("http://snomed.info/sct"),
		PatientID: ptrUUID(patID),
		SerialNumber: ptrStr("SN-12345"),
		ModelNumber:  ptrStr("PO-200"),
		CreatedAt: now, UpdatedAt: now,
	}
	result := d.ToFHIR()

	if result["resourceType"] != "Device" {
		t.Errorf("resourceType = %v, want Device", result["resourceType"])
	}
	if result["id"] != "dev-001" {
		t.Errorf("id = %v, want dev-001", result["id"])
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want active", result["status"])
	}
	// Check deviceName
	names, ok := result["deviceName"].([]map[string]string)
	if !ok {
		t.Fatal("deviceName is not []map[string]string")
	}
	if len(names) != 1 || names[0]["name"] != "Pulse Oximeter" {
		t.Errorf("deviceName = %v", names)
	}
	if names[0]["type"] != "manufacturer-name" {
		t.Errorf("deviceName type = %v, want manufacturer-name", names[0]["type"])
	}
	// Check manufacturer
	if result["manufacturer"] != "Acme Medical" {
		t.Errorf("manufacturer = %v, want Acme Medical", result["manufacturer"])
	}
	// Check type
	tp, ok := result["type"].(fhir.CodeableConcept)
	if !ok {
		t.Fatal("type is not fhir.CodeableConcept")
	}
	if len(tp.Coding) != 1 || tp.Coding[0].Code != "59181002" {
		t.Errorf("type coding = %v", tp.Coding)
	}
	// Check patient
	pat, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	if pat.Reference != "Patient/"+patID.String() {
		t.Errorf("patient.Reference = %v", pat.Reference)
	}
	// Check serialNumber
	if result["serialNumber"] != "SN-12345" {
		t.Errorf("serialNumber = %v", result["serialNumber"])
	}
	// Check modelNumber
	if result["modelNumber"] != "PO-200" {
		t.Errorf("modelNumber = %v", result["modelNumber"])
	}
	// Check meta
	meta, ok := result["meta"].(fhir.Meta)
	if !ok {
		t.Fatal("meta is not fhir.Meta")
	}
	if meta.LastUpdated != now {
		t.Errorf("meta.LastUpdated mismatch")
	}
}

func TestDevice_ToFHIR_AllFields(t *testing.T) {
	now := time.Now()
	mfgDate := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
	expDate := time.Date(2028, 1, 15, 0, 0, 0, 0, time.UTC)
	patID := uuid.New()
	ownerID := uuid.New()
	locID := uuid.New()
	d := &Device{
		ID: uuid.New(), FHIRID: "dev-full", Status: "active",
		StatusReason:       ptrStr("online"),
		DistinctIdentifier: ptrStr("DI-98765"),
		ManufacturerName:   ptrStr("Acme Medical"),
		ManufactureDate:    ptrTime(mfgDate),
		ExpirationDate:     ptrTime(expDate),
		LotNumber:          ptrStr("LOT-2023-A"),
		SerialNumber:       ptrStr("SN-12345"),
		ModelNumber:        ptrStr("PO-200"),
		DeviceName:         "Advanced Pulse Oximeter",
		DeviceNameType:     "manufacturer-name",
		TypeCode:           ptrStr("59181002"),
		TypeDisplay:        ptrStr("Oxygen analyzer"),
		TypeSystem:         ptrStr("http://snomed.info/sct"),
		VersionValue:       ptrStr("2.1.0"),
		PatientID:          ptrUUID(patID),
		OwnerID:            ptrUUID(ownerID),
		LocationID:         ptrUUID(locID),
		ContactPhone:       ptrStr("555-1234"),
		ContactEmail:       ptrStr("support@acme.com"),
		URL:                ptrStr("https://acme.com/po200"),
		Note:               ptrStr("Calibrated monthly"),
		SafetyCode:         ptrStr("C113844"),
		SafetyDisplay:      ptrStr("Latex Not Made with Natural Rubber Latex"),
		UDICarrier:         ptrStr("(01)00844588003288(17)141120(10)7654321D(21)10987654d321"),
		UDIEntryType:       ptrStr("barcode"),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	result := d.ToFHIR()

	// Required fields
	if result["resourceType"] != "Device" {
		t.Errorf("resourceType = %v", result["resourceType"])
	}
	if result["status"] != "active" {
		t.Errorf("status = %v", result["status"])
	}

	// StatusReason
	if _, ok := result["statusReason"]; !ok {
		t.Error("expected statusReason")
	}

	// DistinctIdentifier
	if result["distinctIdentifier"] != "DI-98765" {
		t.Errorf("distinctIdentifier = %v", result["distinctIdentifier"])
	}

	// Manufacturer
	if result["manufacturer"] != "Acme Medical" {
		t.Errorf("manufacturer = %v", result["manufacturer"])
	}

	// ManufactureDate
	if _, ok := result["manufactureDate"]; !ok {
		t.Error("expected manufactureDate")
	}

	// ExpirationDate
	if _, ok := result["expirationDate"]; !ok {
		t.Error("expected expirationDate")
	}

	// LotNumber
	if result["lotNumber"] != "LOT-2023-A" {
		t.Errorf("lotNumber = %v", result["lotNumber"])
	}

	// SerialNumber
	if result["serialNumber"] != "SN-12345" {
		t.Errorf("serialNumber = %v", result["serialNumber"])
	}

	// ModelNumber
	if result["modelNumber"] != "PO-200" {
		t.Errorf("modelNumber = %v", result["modelNumber"])
	}

	// VersionValue
	if _, ok := result["version"]; !ok {
		t.Error("expected version")
	}

	// Patient
	pat, ok := result["patient"].(fhir.Reference)
	if !ok {
		t.Fatal("patient is not fhir.Reference")
	}
	if pat.Reference != "Patient/"+patID.String() {
		t.Errorf("patient.Reference = %v", pat.Reference)
	}

	// Owner
	owner, ok := result["owner"].(fhir.Reference)
	if !ok {
		t.Fatal("owner is not fhir.Reference")
	}
	if owner.Reference != "Organization/"+ownerID.String() {
		t.Errorf("owner.Reference = %v", owner.Reference)
	}

	// Location
	loc, ok := result["location"].(fhir.Reference)
	if !ok {
		t.Fatal("location is not fhir.Reference")
	}
	if loc.Reference != "Location/"+locID.String() {
		t.Errorf("location.Reference = %v", loc.Reference)
	}

	// Contact
	if _, ok := result["contact"]; !ok {
		t.Error("expected contact")
	}

	// URL
	if result["url"] != "https://acme.com/po200" {
		t.Errorf("url = %v", result["url"])
	}

	// Note
	if _, ok := result["note"]; !ok {
		t.Error("expected note")
	}

	// Safety
	if _, ok := result["safety"]; !ok {
		t.Error("expected safety")
	}

	// UDI Carrier
	if _, ok := result["udiCarrier"]; !ok {
		t.Error("expected udiCarrier")
	}
}

func TestDevice_ToFHIR_MinimalFields(t *testing.T) {
	now := time.Now()
	d := &Device{
		ID: uuid.New(), FHIRID: "dev-min", Status: "active",
		DeviceName: "Basic Thermometer", DeviceNameType: "user-friendly-name",
		CreatedAt: now, UpdatedAt: now,
	}
	result := d.ToFHIR()

	// Required fields must be present
	if result["resourceType"] != "Device" {
		t.Errorf("resourceType = %v, want Device", result["resourceType"])
	}
	if result["id"] != "dev-min" {
		t.Errorf("id = %v, want dev-min", result["id"])
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want active", result["status"])
	}

	// deviceName always present (it's required)
	names, ok := result["deviceName"].([]map[string]string)
	if !ok {
		t.Fatal("deviceName should be present")
	}
	if len(names) != 1 || names[0]["name"] != "Basic Thermometer" {
		t.Errorf("deviceName = %v", names)
	}

	// Optional fields must be absent
	for _, key := range []string{
		"statusReason", "distinctIdentifier", "manufacturer",
		"manufactureDate", "expirationDate", "lotNumber",
		"serialNumber", "modelNumber", "type", "version",
		"patient", "owner", "location", "contact", "url",
		"note", "safety", "udiCarrier",
	} {
		if _, ok := result[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}
