package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Test helpers — prefixed with "vd" to avoid collisions with fhirpath_test.go
// ---------------------------------------------------------------------------

func newTestViewEngine() *ViewDefinitionEngine {
	fp := NewFHIRPathEngine()
	return NewViewDefinitionEngine(fp)
}

func newTestViewHandler() *ViewDefinitionHandler {
	engine := newTestViewEngine()
	h := NewViewDefinitionHandler(engine)
	// Register built-in views
	for _, v := range BuiltInViewDefinitions() {
		vc := v
		h.views[vc.ID] = &vc
	}
	return h
}

func vdPatient() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-100",
		"active":       true,
		"birthDate":    "1985-07-23",
		"gender":       "female",
		"name": []interface{}{
			map[string]interface{}{
				"use":    "official",
				"family": "Doe",
				"given":  []interface{}{"Jane", "Marie"},
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-1234",
			},
			map[string]interface{}{
				"system": "email",
				"value":  "jane@example.com",
			},
		},
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hospital.org/mrn",
				"value":  "MRN-12345",
			},
		},
	}
}

func vdConditionActive() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Condition",
		"id":           "cond-1",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://snomed.info/sct",
					"code":    "73211009",
					"display": "Diabetes mellitus",
				},
			},
		},
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code": "active",
				},
			},
		},
		"verificationStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code": "confirmed",
				},
			},
		},
		"onsetDateTime": "2020-01-15",
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code": "encounter-diagnosis",
					},
				},
			},
		},
	}
}

func vdConditionInactive() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Condition",
		"id":           "cond-2",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://snomed.info/sct",
					"code":    "386661006",
					"display": "Fever",
				},
			},
		},
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code": "resolved",
				},
			},
		},
		"verificationStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"code": "confirmed",
				},
			},
		},
	}
}

func vdLabObservation() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-lab-1",
		"status":       "final",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://terminology.hl7.org/CodeSystem/observation-category",
						"code":   "laboratory",
					},
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "2339-0",
					"display": "Glucose",
				},
			},
		},
		"valueQuantity": map[string]interface{}{
			"value": float64(95),
			"unit":  "mg/dL",
		},
		"effectiveDateTime": "2024-03-15T08:00:00Z",
	}
}

func vdVitalSignsObservation() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-bp-100",
		"status":       "final",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"category": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system": "http://terminology.hl7.org/CodeSystem/observation-category",
						"code":   "vital-signs",
					},
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    "85354-9",
					"display": "Blood pressure panel",
				},
			},
		},
		"effectiveDateTime": "2024-06-01T14:30:00Z",
		"component": []interface{}{
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system":  "http://loinc.org",
							"code":    "8480-6",
							"display": "Systolic blood pressure",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": float64(130),
					"unit":  "mmHg",
				},
			},
			map[string]interface{}{
				"code": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system":  "http://loinc.org",
							"code":    "8462-4",
							"display": "Diastolic blood pressure",
						},
					},
				},
				"valueQuantity": map[string]interface{}{
					"value": float64(85),
					"unit":  "mmHg",
				},
			},
		},
	}
}

func vdMedRequest() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "medrx-1",
		"status":       "active",
		"intent":       "order",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
					"code":    "860975",
					"display": "Metformin 500 MG",
				},
			},
		},
		"authoredOn": "2024-01-10",
		"dosageInstruction": []interface{}{
			map[string]interface{}{
				"text": "Take 1 tablet by mouth twice daily",
			},
		},
	}
}

func vdMedRequestInactive() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           "medrx-2",
		"status":       "completed",
		"intent":       "order",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
					"code":    "197361",
					"display": "Amoxicillin 500 MG",
				},
			},
		},
		"authoredOn": "2023-06-01",
	}
}

func vdEncounter() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           "enc-1",
		"status":       "finished",
		"subject": map[string]interface{}{
			"reference": "Patient/pt-100",
		},
		"class": map[string]interface{}{
			"code": "AMB",
		},
		"type": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "99213",
						"display": "Office visit",
					},
				},
			},
		},
		"period": map[string]interface{}{
			"start": "2024-03-01T09:00:00Z",
			"end":   "2024-03-01T09:30:00Z",
		},
		"reasonCode": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"code":    "J06.9",
						"display": "Upper respiratory infection",
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// ViewDefinition CRUD Tests
// ---------------------------------------------------------------------------

func TestViewDefinition_CRUD_Create(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	body := `{
		"id": "test-view-1",
		"name": "TestView",
		"resource": "Patient",
		"status": "active",
		"select": [
			{"path": "id", "name": "id", "type": "string"}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/ViewDefinition", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var result ViewDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Name != "TestView" {
		t.Errorf("expected name TestView, got %q", result.Name)
	}
}

func TestViewDefinition_CRUD_List(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/ViewDefinition", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result []ViewDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Should have built-in views
	if len(result) < 6 {
		t.Errorf("expected at least 6 built-in views, got %d", len(result))
	}
}

func TestViewDefinition_CRUD_Get(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/ViewDefinition/patient_demographics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result ViewDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.ID != "patient_demographics" {
		t.Errorf("expected ID patient_demographics, got %q", result.ID)
	}
}

func TestViewDefinition_CRUD_Update(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	// Create first
	body := `{
		"id": "update-test",
		"name": "OriginalName",
		"resource": "Patient",
		"status": "active",
		"select": [{"path": "id", "name": "id", "type": "string"}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/ViewDefinition", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rec.Code)
	}

	// Update
	updateBody := `{
		"id": "update-test",
		"name": "UpdatedName",
		"resource": "Patient",
		"status": "active",
		"select": [{"path": "id", "name": "id", "type": "string"}, {"path": "gender", "name": "gender", "type": "string"}]
	}`
	req = httptest.NewRequest(http.MethodPut, "/fhir/ViewDefinition/update-test", strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result ViewDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Name != "UpdatedName" {
		t.Errorf("expected updated name, got %q", result.Name)
	}
	if len(result.Select) != 2 {
		t.Errorf("expected 2 columns, got %d", len(result.Select))
	}
}

func TestViewDefinition_CRUD_Delete(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	// Create
	body := `{
		"id": "delete-test",
		"name": "DeleteMe",
		"resource": "Patient",
		"status": "active",
		"select": [{"path": "id", "name": "id", "type": "string"}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/ViewDefinition", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rec.Code)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/fhir/ViewDefinition/delete-test", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}

	// Verify gone
	req = httptest.NewRequest(http.MethodGet, "/fhir/ViewDefinition/delete-test", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Execute Tests
// ---------------------------------------------------------------------------

func TestViewDefinition_Execute_SimplePath(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "simple-test",
		Name:     "SimpleTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
	if result.Rows[0][0] != "pt-100" {
		t.Errorf("expected id pt-100, got %v", result.Rows[0][0])
	}
}

func TestViewDefinition_Execute_NestedPath(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "nested-test",
		Name:     "NestedTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "name.where(use='official').family", Name: "family_name", Type: "string"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
	if result.Rows[0][0] != "Doe" {
		t.Errorf("expected family_name Doe, got %v", result.Rows[0][0])
	}
}

func TestViewDefinition_Execute_WhereClause(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "where-test",
		Name:     "WhereTest",
		Resource: "Condition",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
		},
		Where: []ViewWhere{
			{Path: "clinicalStatus.coding.code = 'active'"},
		},
	}

	resources := []map[string]interface{}{
		vdConditionActive(),
		vdConditionInactive(),
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row (only active), got %d", len(result.Rows))
	}
	if result.Rows[0][0] != "cond-1" {
		t.Errorf("expected cond-1, got %v", result.Rows[0][0])
	}
}

func TestViewDefinition_Execute_MultipleWhereClauses(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "multi-where",
		Name:     "MultiWhere",
		Resource: "Condition",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
		},
		Where: []ViewWhere{
			{Path: "clinicalStatus.coding.code = 'active'"},
			{Path: "verificationStatus.coding.code = 'confirmed'"},
		},
	}

	resources := []map[string]interface{}{
		vdConditionActive(),
		vdConditionInactive(),
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Only cond-1 is active AND confirmed
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
}

func TestViewDefinition_Execute_TypeCoercion(t *testing.T) {
	engine := newTestViewEngine()

	view := &ViewDefinition{
		ID:       "type-test",
		Name:     "TypeTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "active", Name: "active", Type: "boolean"},
			{Path: "birthDate", Name: "birth_date", Type: "date"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	row := result.Rows[0]
	if row[0] != "pt-100" {
		t.Errorf("string coercion: expected pt-100, got %v", row[0])
	}
	if row[1] != true {
		t.Errorf("boolean coercion: expected true, got %v (%T)", row[1], row[1])
	}
	if row[2] != "1985-07-23" {
		t.Errorf("date coercion: expected 1985-07-23, got %v", row[2])
	}
}

func TestViewDefinition_Execute_CollectionColumns(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "collection-test",
		Name:     "CollectionTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "name.where(use='official').given", Name: "given_names", Type: "string", Collection: true},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	val := result.Rows[0][0]
	arr, ok := val.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{} for collection column, got %T: %v", val, val)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 given names, got %d", len(arr))
	}
}

func TestViewDefinition_Execute_NullMissingValues(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "null-test",
		Name:     "NullTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "maritalStatus.coding.code", Name: "marital_status", Type: "string"},
			{Path: "deceasedDateTime", Name: "deceased_date", Type: "dateTime"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	row := result.Rows[0]
	if row[0] != "pt-100" {
		t.Errorf("expected id pt-100, got %v", row[0])
	}
	if row[1] != nil {
		t.Errorf("expected nil for missing marital_status, got %v", row[1])
	}
	if row[2] != nil {
		t.Errorf("expected nil for missing deceased_date, got %v", row[2])
	}
}

// ---------------------------------------------------------------------------
// Built-in View Tests
// ---------------------------------------------------------------------------

func TestViewDefinition_Execute_PatientDemographics(t *testing.T) {
	engine := newTestViewEngine()
	views := BuiltInViewDefinitions()
	var view *ViewDefinition
	for i := range views {
		if views[i].ID == "patient_demographics" {
			view = &views[i]
			break
		}
	}
	if view == nil {
		t.Fatal("patient_demographics view not found")
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	colNames := make(map[string]int)
	for i, c := range result.Columns {
		colNames[c.Name] = i
	}

	expectedCols := []string{"id", "family_name", "given_name", "birth_date", "gender", "mrn", "phone", "email", "active"}
	for _, name := range expectedCols {
		if _, ok := colNames[name]; !ok {
			t.Errorf("expected column %q not found in result", name)
		}
	}

	row := result.Rows[0]
	if idx, ok := colNames["id"]; ok && row[idx] != "pt-100" {
		t.Errorf("expected id pt-100, got %v", row[idx])
	}
	if idx, ok := colNames["family_name"]; ok && row[idx] != "Doe" {
		t.Errorf("expected family_name Doe, got %v", row[idx])
	}
	if idx, ok := colNames["gender"]; ok && row[idx] != "female" {
		t.Errorf("expected gender female, got %v", row[idx])
	}
	if idx, ok := colNames["mrn"]; ok && row[idx] != "MRN-12345" {
		t.Errorf("expected mrn MRN-12345, got %v", row[idx])
	}
	if idx, ok := colNames["active"]; ok && row[idx] != true {
		t.Errorf("expected active true, got %v", row[idx])
	}
}

func TestViewDefinition_Execute_ActiveConditions(t *testing.T) {
	engine := newTestViewEngine()
	views := BuiltInViewDefinitions()
	var view *ViewDefinition
	for i := range views {
		if views[i].ID == "active_conditions" {
			view = &views[i]
			break
		}
	}
	if view == nil {
		t.Fatal("active_conditions view not found")
	}

	resources := []map[string]interface{}{
		vdConditionActive(),
		vdConditionInactive(),
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	colNames := make(map[string]int)
	for i, c := range result.Columns {
		colNames[c.Name] = i
	}

	if idx, ok := colNames["id"]; ok && result.Rows[0][idx] != "cond-1" {
		t.Errorf("expected id cond-1, got %v", result.Rows[0][idx])
	}
	if idx, ok := colNames["code_display"]; ok && result.Rows[0][idx] != "Diabetes mellitus" {
		t.Errorf("expected code_display Diabetes mellitus, got %v", result.Rows[0][idx])
	}
}

func TestViewDefinition_Execute_LabResults(t *testing.T) {
	engine := newTestViewEngine()
	views := BuiltInViewDefinitions()
	var view *ViewDefinition
	for i := range views {
		if views[i].ID == "lab_results" {
			view = &views[i]
			break
		}
	}
	if view == nil {
		t.Fatal("lab_results view not found")
	}

	resources := []map[string]interface{}{
		vdLabObservation(),
		vdVitalSignsObservation(), // should be filtered out
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row (lab only), got %d", len(result.Rows))
	}

	colNames := make(map[string]int)
	for i, c := range result.Columns {
		colNames[c.Name] = i
	}

	if idx, ok := colNames["code_display"]; ok && result.Rows[0][idx] != "Glucose" {
		t.Errorf("expected code_display Glucose, got %v", result.Rows[0][idx])
	}
}

func TestViewDefinition_Execute_ActiveMedications(t *testing.T) {
	engine := newTestViewEngine()
	views := BuiltInViewDefinitions()
	var view *ViewDefinition
	for i := range views {
		if views[i].ID == "active_medications" {
			view = &views[i]
			break
		}
	}
	if view == nil {
		t.Fatal("active_medications view not found")
	}

	resources := []map[string]interface{}{
		vdMedRequest(),
		vdMedRequestInactive(),
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row (active only), got %d", len(result.Rows))
	}

	colNames := make(map[string]int)
	for i, c := range result.Columns {
		colNames[c.Name] = i
	}

	if idx, ok := colNames["medication_display"]; ok && result.Rows[0][idx] != "Metformin 500 MG" {
		t.Errorf("expected medication_display Metformin 500 MG, got %v", result.Rows[0][idx])
	}
	if idx, ok := colNames["status"]; ok && result.Rows[0][idx] != "active" {
		t.Errorf("expected status active, got %v", result.Rows[0][idx])
	}
}

func TestViewDefinition_Execute_VitalSigns_BP(t *testing.T) {
	engine := newTestViewEngine()
	views := BuiltInViewDefinitions()
	var view *ViewDefinition
	for i := range views {
		if views[i].ID == "vital_signs" {
			view = &views[i]
			break
		}
	}
	if view == nil {
		t.Fatal("vital_signs view not found")
	}

	resources := []map[string]interface{}{
		vdVitalSignsObservation(),
		vdLabObservation(), // should be filtered out
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row (vital signs only), got %d", len(result.Rows))
	}

	colNames := make(map[string]int)
	for i, c := range result.Columns {
		colNames[c.Name] = i
	}

	if idx, ok := colNames["systolic"]; ok {
		v := result.Rows[0][idx]
		if v == nil {
			t.Error("expected systolic value, got nil")
		} else if fmt.Sprintf("%v", v) != "130" {
			t.Errorf("expected systolic 130, got %v", v)
		}
	} else {
		t.Error("systolic column not found")
	}

	if idx, ok := colNames["diastolic"]; ok {
		v := result.Rows[0][idx]
		if v == nil {
			t.Error("expected diastolic value, got nil")
		} else if fmt.Sprintf("%v", v) != "85" {
			t.Errorf("expected diastolic 85, got %v", v)
		}
	} else {
		t.Error("diastolic column not found")
	}
}

func TestViewDefinition_Execute_EncountersSummary(t *testing.T) {
	engine := newTestViewEngine()
	views := BuiltInViewDefinitions()
	var view *ViewDefinition
	for i := range views {
		if views[i].ID == "encounters_summary" {
			view = &views[i]
			break
		}
	}
	if view == nil {
		t.Fatal("encounters_summary view not found")
	}

	resources := []map[string]interface{}{vdEncounter()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	colNames := make(map[string]int)
	for i, c := range result.Columns {
		colNames[c.Name] = i
	}

	if idx, ok := colNames["status"]; ok && result.Rows[0][idx] != "finished" {
		t.Errorf("expected status finished, got %v", result.Rows[0][idx])
	}
	if idx, ok := colNames["class_code"]; ok && result.Rows[0][idx] != "AMB" {
		t.Errorf("expected class_code AMB, got %v", result.Rows[0][idx])
	}
	if idx, ok := colNames["type_display"]; ok && result.Rows[0][idx] != "Office visit" {
		t.Errorf("expected type_display Office visit, got %v", result.Rows[0][idx])
	}
}

// ---------------------------------------------------------------------------
// Output Format Tests
// ---------------------------------------------------------------------------

func TestViewDefinition_ToCSV(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "csv-test",
		Name:     "CSVTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "gender", Name: "gender", Type: "string"},
			{Path: "birthDate", Name: "birth_date", Type: "date"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	csv := engine.ToCSV(result)
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}
	if lines[0] != "id,gender,birth_date" {
		t.Errorf("unexpected header: %q", lines[0])
	}
	if !strings.Contains(lines[1], "pt-100") {
		t.Errorf("expected row to contain pt-100: %q", lines[1])
	}
}

func TestViewDefinition_ToJSON(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "json-test",
		Name:     "JSONTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "gender", Name: "gender", Type: "string"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	jsonResult := engine.ToJSON(result)
	if len(jsonResult) != 1 {
		t.Fatalf("expected 1 object, got %d", len(jsonResult))
	}
	if jsonResult[0]["id"] != "pt-100" {
		t.Errorf("expected id pt-100, got %v", jsonResult[0]["id"])
	}
	if jsonResult[0]["gender"] != "female" {
		t.Errorf("expected gender female, got %v", jsonResult[0]["gender"])
	}
}

func TestViewDefinition_ToNDJSON(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "ndjson-test",
		Name:     "NDJSONTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "gender", Name: "gender", Type: "string"},
		},
	}

	resources := []map[string]interface{}{
		vdPatient(),
		{
			"resourceType": "Patient",
			"id":           "pt-200",
			"gender":       "male",
		},
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	ndjson := engine.ToNDJSON(result)
	lines := strings.Split(strings.TrimSpace(ndjson), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var obj1 map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &obj1); err != nil {
		t.Fatalf("unmarshal line 1: %v", err)
	}
	if obj1["id"] != "pt-100" {
		t.Errorf("expected id pt-100, got %v", obj1["id"])
	}
}

func TestViewDefinition_GenerateSQL(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "sql-test",
		Name:     "patient_view",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "gender", Name: "gender", Type: "string"},
			{Path: "birthDate", Name: "birth_date", Type: "date"},
		},
	}

	sql := engine.GenerateSQL(view)
	if !strings.Contains(sql, "CREATE") {
		t.Error("expected CREATE in SQL")
	}
	if !strings.Contains(sql, "VIEW") {
		t.Error("expected VIEW in SQL")
	}
	if !strings.Contains(sql, "patient_view") {
		t.Error("expected view name in SQL")
	}
}

// ---------------------------------------------------------------------------
// Handler Tests
// ---------------------------------------------------------------------------

func TestViewDefinition_Handler_Execute_JSON(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	body, _ := json.Marshal([]map[string]interface{}{vdPatient()})
	req := httptest.NewRequest(http.MethodPost, "/fhir/ViewDefinition/patient_demographics/$execute?_format=json", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0]["id"] != "pt-100" {
		t.Errorf("expected id pt-100, got %v", result[0]["id"])
	}
}

func TestViewDefinition_Handler_Execute_CSV(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	body, _ := json.Marshal([]map[string]interface{}{vdPatient()})
	req := httptest.NewRequest(http.MethodPost, "/fhir/ViewDefinition/patient_demographics/$execute?_format=csv", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	csvStr := rec.Body.String()
	if !strings.Contains(csvStr, "id,") || !strings.Contains(csvStr, "family_name") {
		t.Errorf("CSV missing expected headers: %s", csvStr)
	}
	if !strings.Contains(csvStr, "pt-100") {
		t.Errorf("CSV missing expected data: %s", csvStr)
	}
}

func TestViewDefinition_Handler_SQL(t *testing.T) {
	h := newTestViewHandler()
	e := echo.New()
	h.RegisterRoutes(e.Group("/fhir"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/ViewDefinition/patient_demographics/$sql", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	sql := rec.Body.String()
	if !strings.Contains(sql, "CREATE") {
		t.Error("expected CREATE in SQL output")
	}
}

func TestViewDefinition_Execute_EmptyResources(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "empty-test",
		Name:     "EmptyTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "gender", Name: "gender", Type: "string"},
		},
	}

	result, err := engine.Execute(context.Background(), view, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(result.Columns))
	}
	if len(result.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(result.Rows))
	}
}

func TestViewDefinition_Execute_WithConstants(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "const-test",
		Name:     "ConstTest",
		Resource: "Observation",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "status", Name: "status", Type: "string"},
		},
		Constants: []ViewConstant{
			{Name: "target_status", Value: "final"},
		},
		Where: []ViewWhere{
			{Path: "status = 'final'"},
		},
	}

	resources := []map[string]interface{}{
		vdLabObservation(),
	}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
}

func TestViewDefinition_Execute_Concurrent(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "concurrent-test",
		Name:     "ConcurrentTest",
		Resource: "Patient",
		Select: []ViewColumn{
			{Path: "id", Name: "id", Type: "string"},
			{Path: "gender", Name: "gender", Type: "string"},
		},
	}

	resources := []map[string]interface{}{vdPatient()}

	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			result, err := engine.Execute(context.Background(), view, resources)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: %w", idx, err)
				return
			}
			if len(result.Rows) != 1 {
				errors <- fmt.Errorf("goroutine %d: expected 1 row, got %d", idx, len(result.Rows))
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// ---------------------------------------------------------------------------
// Integer / Decimal type coercion tests
// ---------------------------------------------------------------------------

func TestViewDefinition_Execute_IntegerCoercion(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "int-test",
		Name:     "IntTest",
		Resource: "Observation",
		Select: []ViewColumn{
			{Path: "valueQuantity.value", Name: "value", Type: "integer"},
		},
	}

	resources := []map[string]interface{}{vdLabObservation()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	val := result.Rows[0][0]
	switch v := val.(type) {
	case int64:
		if v != 95 {
			t.Errorf("expected 95, got %d", v)
		}
	case int:
		if v != 95 {
			t.Errorf("expected 95, got %d", v)
		}
	case float64:
		if v != 95 {
			t.Errorf("expected 95, got %f", v)
		}
	default:
		t.Errorf("unexpected type for integer coercion: %T = %v", val, val)
	}
}

func TestViewDefinition_Execute_DecimalCoercion(t *testing.T) {
	engine := newTestViewEngine()
	view := &ViewDefinition{
		ID:       "dec-test",
		Name:     "DecTest",
		Resource: "Observation",
		Select: []ViewColumn{
			{Path: "valueQuantity.value", Name: "value", Type: "decimal"},
		},
	}

	resources := []map[string]interface{}{vdLabObservation()}
	result, err := engine.Execute(context.Background(), view, resources)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	val := result.Rows[0][0]
	switch v := val.(type) {
	case float64:
		if v != 95.0 {
			t.Errorf("expected 95.0, got %f", v)
		}
	default:
		t.Errorf("unexpected type for decimal coercion: %T = %v", val, val)
	}
}
