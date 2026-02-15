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

// ===========================================================================
// Test helpers
// ===========================================================================

func newPlanDefinitionEngine() *PlanDefinitionEngine {
	return NewPlanDefinitionEngine(NewFHIRPathEngine())
}

func newPlanDefinitionHandler() *PlanDefinitionHandler {
	return NewPlanDefinitionHandler(NewFHIRPathEngine())
}

// testPatientDiabetes returns a patient with HbA1c > 9.
func testPatientDiabetes() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-dm-1",
		"gender":       "male",
		"birthDate":    "1965-04-10",
		"active":       true,
	}
}

// testPatientSepsis returns a patient with temperature > 38.3.
func testPatientSepsis() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-sepsis-1",
		"gender":       "female",
		"birthDate":    "1980-07-22",
		"active":       true,
	}
}

// testPatientCHF returns a patient for CHF discharge protocol.
func testPatientCHF() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-chf-1",
		"gender":       "male",
		"birthDate":    "1955-11-30",
		"active":       true,
	}
}

// testPatientPreventiveMale50 returns a 50-year-old male for preventive care.
func testPatientPreventiveMale50() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-prev-1",
		"gender":       "male",
		"birthDate":    "1975-01-01",
		"active":       true,
	}
}

// testPatientPreventiveFemale45 returns a 45-year-old female.
func testPatientPreventiveFemale45() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-prev-2",
		"gender":       "female",
		"birthDate":    "1980-01-01",
		"active":       true,
	}
}

func testPatientYoung30() map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           "pt-young-1",
		"gender":       "male",
		"birthDate":    "1995-06-15",
		"active":       true,
	}
}

func setupHandlerTest(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// ===========================================================================
// PlanDefinition CRUD Tests
// ===========================================================================

func TestPlanDefinition_CRUD_Create(t *testing.T) {
	h := newPlanDefinitionHandler()

	pd := PlanDefinition{
		ID:          "pd-test-1",
		URL:         "http://example.org/PlanDefinition/test-1",
		Version:     "1.0",
		Name:        "TestPlan",
		Title:       "Test Plan Definition",
		Status:      "active",
		Type:        "order-set",
		SubjectType: "Patient",
	}
	body, _ := json.Marshal(pd)
	c, rec := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition", string(body))

	err := h.CreatePlanDefinition(c)
	if err != nil {
		t.Fatalf("CreatePlanDefinition error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "PlanDefinition" {
		t.Errorf("expected resourceType PlanDefinition, got %v", result["resourceType"])
	}
	if result["id"] != "pd-test-1" {
		t.Errorf("expected id pd-test-1, got %v", result["id"])
	}
}

func TestPlanDefinition_CRUD_Read(t *testing.T) {
	h := newPlanDefinitionHandler()

	// Create first
	pd := PlanDefinition{
		ID:     "pd-read-1",
		URL:    "http://example.org/PlanDefinition/read-1",
		Name:   "ReadTest",
		Title:  "Read Test",
		Status: "active",
		Type:   "clinical-protocol",
	}
	body, _ := json.Marshal(pd)
	c, _ := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition", string(body))
	h.CreatePlanDefinition(c)

	// Read
	c2, rec2 := setupHandlerTest(http.MethodGet, "/fhir/PlanDefinition/pd-read-1", "")
	c2.SetParamNames("id")
	c2.SetParamValues("pd-read-1")

	err := h.GetPlanDefinition(c2)
	if err != nil {
		t.Fatalf("GetPlanDefinition error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &result)
	if result["id"] != "pd-read-1" {
		t.Errorf("expected id pd-read-1, got %v", result["id"])
	}
}

func TestPlanDefinition_CRUD_List(t *testing.T) {
	h := newPlanDefinitionHandler()

	// Create two plan definitions
	for _, id := range []string{"pd-list-1", "pd-list-2"} {
		pd := PlanDefinition{ID: id, Name: id, Status: "active", Type: "order-set"}
		body, _ := json.Marshal(pd)
		c, _ := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition", string(body))
		h.CreatePlanDefinition(c)
	}

	c, rec := setupHandlerTest(http.MethodGet, "/fhir/PlanDefinition", "")
	err := h.ListPlanDefinitions(c)
	if err != nil {
		t.Fatalf("ListPlanDefinitions error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", result["resourceType"])
	}
	total, _ := result["total"].(float64)
	if total < 2 {
		t.Errorf("expected at least 2 entries, got %v", total)
	}
}

func TestPlanDefinition_CRUD_Update(t *testing.T) {
	h := newPlanDefinitionHandler()

	// Create
	pd := PlanDefinition{ID: "pd-upd-1", Name: "Original", Status: "draft", Type: "order-set"}
	body, _ := json.Marshal(pd)
	c, _ := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition", string(body))
	h.CreatePlanDefinition(c)

	// Update
	pd.Title = "Updated Title"
	pd.Status = "active"
	body, _ = json.Marshal(pd)
	c2, rec2 := setupHandlerTest(http.MethodPut, "/fhir/PlanDefinition/pd-upd-1", string(body))
	c2.SetParamNames("id")
	c2.SetParamValues("pd-upd-1")

	err := h.UpdatePlanDefinition(c2)
	if err != nil {
		t.Fatalf("UpdatePlanDefinition error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}

	// Verify
	c3, rec3 := setupHandlerTest(http.MethodGet, "/fhir/PlanDefinition/pd-upd-1", "")
	c3.SetParamNames("id")
	c3.SetParamValues("pd-upd-1")
	h.GetPlanDefinition(c3)

	var result map[string]interface{}
	json.Unmarshal(rec3.Body.Bytes(), &result)
	if result["title"] != "Updated Title" {
		t.Errorf("expected Updated Title, got %v", result["title"])
	}
	if result["status"] != "active" {
		t.Errorf("expected active, got %v", result["status"])
	}
}

func TestPlanDefinition_CRUD_Delete(t *testing.T) {
	h := newPlanDefinitionHandler()

	// Create
	pd := PlanDefinition{ID: "pd-del-1", Name: "DeleteMe", Status: "draft", Type: "order-set"}
	body, _ := json.Marshal(pd)
	c, _ := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition", string(body))
	h.CreatePlanDefinition(c)

	// Delete
	c2, rec2 := setupHandlerTest(http.MethodDelete, "/fhir/PlanDefinition/pd-del-1", "")
	c2.SetParamNames("id")
	c2.SetParamValues("pd-del-1")

	err := h.DeletePlanDefinition(c2)
	if err != nil {
		t.Fatalf("DeletePlanDefinition error: %v", err)
	}
	if rec2.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec2.Code)
	}

	// Verify not found
	c3, rec3 := setupHandlerTest(http.MethodGet, "/fhir/PlanDefinition/pd-del-1", "")
	c3.SetParamNames("id")
	c3.SetParamValues("pd-del-1")
	h.GetPlanDefinition(c3)
	if rec3.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rec3.Code)
	}
}

// ===========================================================================
// ActivityDefinition CRUD Tests
// ===========================================================================

func TestPlanDefinition_ActivityDefinition_CRUD(t *testing.T) {
	h := newPlanDefinitionHandler()

	// Create
	ad := ActivityDefinition{
		ID:     "ad-test-1",
		URL:    "http://example.org/ActivityDefinition/test-1",
		Name:   "TestActivity",
		Title:  "Test Activity Definition",
		Status: "active",
		Kind:   "MedicationRequest",
		Code: &ActivityCode{
			System:  "http://www.nlm.nih.gov/research/umls/rxnorm",
			Code:    "860975",
			Display: "Metformin 500 MG",
		},
	}
	body, _ := json.Marshal(ad)
	c, rec := setupHandlerTest(http.MethodPost, "/fhir/ActivityDefinition", string(body))
	err := h.CreateActivityDefinition(c)
	if err != nil {
		t.Fatalf("CreateActivityDefinition error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	// Read
	c2, rec2 := setupHandlerTest(http.MethodGet, "/fhir/ActivityDefinition/ad-test-1", "")
	c2.SetParamNames("id")
	c2.SetParamValues("ad-test-1")
	err = h.GetActivityDefinition(c2)
	if err != nil {
		t.Fatalf("GetActivityDefinition error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &result)
	if result["kind"] != "MedicationRequest" {
		t.Errorf("expected kind MedicationRequest, got %v", result["kind"])
	}

	// List
	c3, rec3 := setupHandlerTest(http.MethodGet, "/fhir/ActivityDefinition", "")
	err = h.ListActivityDefinitions(c3)
	if err != nil {
		t.Fatalf("ListActivityDefinitions error: %v", err)
	}
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec3.Code)
	}
}

// ===========================================================================
// $apply — Basic Tests
// ===========================================================================

func TestPlanDefinition_Apply_BasicSingleAction(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-basic-1",
		Name:   "BasicPlan",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:          "action-1",
				Title:       "Order Lab Test",
				Description: "Order a basic lab test",
				Type:        "create",
				DefinitionCanonical: "ActivityDefinition/ad-lab-1",
			},
		},
	}

	ad := &ActivityDefinition{
		ID:     "ad-lab-1",
		Name:   "LabTest",
		Status: "active",
		Kind:   "ServiceRequest",
		Code: &ActivityCode{
			System:  "http://loinc.org",
			Code:    "4548-4",
			Display: "HbA1c",
		},
	}
	engine.RegisterActivityDefinition(ad)

	subject := testPatientDiabetes()
	result, err := engine.Apply(context.Background(), pd, subject, nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	res := result.Resources[0]
	if res["resourceType"] != "ServiceRequest" {
		t.Errorf("expected ServiceRequest, got %v", res["resourceType"])
	}
}

func TestPlanDefinition_Apply_ConditionFalse_ActionSkipped(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-cond-false",
		Name:   "CondFalsePlan",
		Status: "active",
		Type:   "eca-rule",
		Action: []PlanAction{
			{
				ID:    "action-1",
				Title: "Should be skipped",
				Type:  "create",
				Condition: []PlanCondition{
					{
						Kind:       "applicability",
						Expression: "Patient.gender = 'female'",
					},
				},
				DefinitionCanonical: "ActivityDefinition/ad-skip-1",
			},
		},
	}

	ad := &ActivityDefinition{
		ID:     "ad-skip-1",
		Name:   "SkipActivity",
		Status: "active",
		Kind:   "ServiceRequest",
	}
	engine.RegisterActivityDefinition(ad)

	// Patient is male, condition checks for female
	subject := testPatientDiabetes()
	result, err := engine.Apply(context.Background(), pd, subject, nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if len(result.Resources) != 0 {
		t.Errorf("expected 0 resources (skipped), got %d", len(result.Resources))
	}
}

func TestPlanDefinition_Apply_ConditionTrue_ActionIncluded(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-cond-true",
		Name:   "CondTruePlan",
		Status: "active",
		Type:   "eca-rule",
		Action: []PlanAction{
			{
				ID:    "action-1",
				Title: "Should be included",
				Type:  "create",
				Condition: []PlanCondition{
					{
						Kind:       "applicability",
						Expression: "Patient.gender = 'male'",
					},
				},
				DefinitionCanonical: "ActivityDefinition/ad-inc-1",
			},
		},
	}

	ad := &ActivityDefinition{
		ID:     "ad-inc-1",
		Name:   "IncludeActivity",
		Status: "active",
		Kind:   "ServiceRequest",
		Code: &ActivityCode{
			System:  "http://loinc.org",
			Code:    "4548-4",
			Display: "HbA1c",
		},
	}
	engine.RegisterActivityDefinition(ad)

	subject := testPatientDiabetes() // male
	result, err := engine.Apply(context.Background(), pd, subject, nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(result.Resources))
	}
}

func TestPlanDefinition_Apply_GeneratesCarePlan(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-cp-1",
		Name:   "CarePlanGen",
		Title:  "Test CarePlan Generation",
		Status: "active",
		Type:   "clinical-protocol",
		Goal: []PlanGoal{
			{
				Description: "Achieve glycemic control",
				Priority:    "high-priority",
			},
		},
		Action: []PlanAction{
			{
				ID:    "action-1",
				Title: "Monitor HbA1c",
				Type:  "create",
				DefinitionCanonical: "ActivityDefinition/ad-cp-1",
			},
		},
	}

	ad := &ActivityDefinition{
		ID:   "ad-cp-1",
		Kind: "ServiceRequest",
		Code: &ActivityCode{Code: "4548-4", Display: "HbA1c"},
	}
	engine.RegisterActivityDefinition(ad)

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	cp := result.CarePlan
	if cp == nil {
		t.Fatal("expected CarePlan in result")
	}
	if cp["resourceType"] != "CarePlan" {
		t.Errorf("expected CarePlan resourceType, got %v", cp["resourceType"])
	}
	if cp["status"] != "active" {
		t.Errorf("expected status active, got %v", cp["status"])
	}
	if cp["title"] != "Test CarePlan Generation" {
		t.Errorf("expected title from PlanDefinition, got %v", cp["title"])
	}
}

func TestPlanDefinition_Apply_GeneratesRequestGroup(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-rg-1",
		Name:   "RequestGroupGen",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:    "action-1",
				Title: "Lab Order",
				Type:  "create",
				DefinitionCanonical: "ActivityDefinition/ad-rg-1",
			},
			{
				ID:    "action-2",
				Title: "Medication Order",
				Type:  "create",
				DefinitionCanonical: "ActivityDefinition/ad-rg-2",
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-rg-1", Kind: "ServiceRequest",
		Code: &ActivityCode{Code: "lab-1"},
	})
	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-rg-2", Kind: "MedicationRequest",
		Code: &ActivityCode{Code: "med-1"},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	rg := result.RequestGroup
	if rg == nil {
		t.Fatal("expected RequestGroup in result")
	}
	if rg["resourceType"] != "RequestGroup" {
		t.Errorf("expected RequestGroup resourceType, got %v", rg["resourceType"])
	}

	actions, ok := rg["action"].([]interface{})
	if !ok {
		t.Fatal("expected action array in RequestGroup")
	}
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}
}

func TestPlanDefinition_Apply_ResolvesActivityDefinition_MedicationRequest(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-med-1",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:   "action-1",
				Type: "create",
				DefinitionCanonical: "ActivityDefinition/ad-metformin",
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID:     "ad-metformin",
		Name:   "Metformin",
		Status: "active",
		Kind:   "MedicationRequest",
		Code: &ActivityCode{
			System:  "http://www.nlm.nih.gov/research/umls/rxnorm",
			Code:    "860975",
			Display: "Metformin 500 MG Oral Tablet",
		},
		Dosage:  "500mg twice daily",
		Product: "Metformin hydrochloride",
	})

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	med := result.Resources[0]
	if med["resourceType"] != "MedicationRequest" {
		t.Errorf("expected MedicationRequest, got %v", med["resourceType"])
	}
	if med["status"] != "draft" {
		t.Errorf("expected draft status, got %v", med["status"])
	}
	if med["intent"] != "order" {
		t.Errorf("expected intent order, got %v", med["intent"])
	}
}

func TestPlanDefinition_Apply_ResolvesActivityDefinition_ServiceRequest(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-sr-1",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:   "action-1",
				Type: "create",
				DefinitionCanonical: "ActivityDefinition/ad-blood-culture",
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID:     "ad-blood-culture",
		Name:   "BloodCulture",
		Status: "active",
		Kind:   "ServiceRequest",
		Code: &ActivityCode{
			System:  "http://loinc.org",
			Code:    "600-7",
			Display: "Blood culture",
		},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientSepsis(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	sr := result.Resources[0]
	if sr["resourceType"] != "ServiceRequest" {
		t.Errorf("expected ServiceRequest, got %v", sr["resourceType"])
	}
	if sr["status"] != "draft" {
		t.Errorf("expected draft status, got %v", sr["status"])
	}
}

func TestPlanDefinition_Apply_ResolvesActivityDefinition_Task(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-task-1",
		Status: "active",
		Type:   "workflow-definition",
		Action: []PlanAction{
			{
				ID:   "action-1",
				Type: "create",
				DefinitionCanonical: "ActivityDefinition/ad-weight-task",
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID:     "ad-weight-task",
		Name:   "DailyWeight",
		Status: "active",
		Kind:   "Task",
		Code: &ActivityCode{
			Code:    "daily-weight",
			Display: "Daily weight monitoring",
		},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientCHF(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	task := result.Resources[0]
	if task["resourceType"] != "Task" {
		t.Errorf("expected Task, got %v", task["resourceType"])
	}
	if task["status"] != "draft" {
		t.Errorf("expected draft status, got %v", task["status"])
	}
}

func TestPlanDefinition_Apply_DynamicValueExpressions(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-dv-1",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:   "action-1",
				Type: "create",
				DefinitionCanonical: "ActivityDefinition/ad-dv-1",
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID:     "ad-dv-1",
		Status: "active",
		Kind:   "ServiceRequest",
		Code: &ActivityCode{Code: "test-1"},
		DynamicValue: []DynamicValue{
			{
				Path:       "priority",
				Expression: "'urgent'",
			},
			{
				Path:       "note",
				Expression: "Patient.id",
			},
		},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	sr := result.Resources[0]
	if sr["priority"] != "urgent" {
		t.Errorf("expected priority 'urgent', got %v", sr["priority"])
	}
	if sr["note"] != "pt-dm-1" {
		t.Errorf("expected note 'pt-dm-1', got %v", sr["note"])
	}
}

func TestPlanDefinition_Apply_NestedActions(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-nested-1",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:    "group-1",
				Title: "Medication Group",
				Type:  "create",
				GroupingBehavior: "logical-group",
				Action: []PlanAction{
					{
						ID:   "sub-action-1",
						Type: "create",
						DefinitionCanonical: "ActivityDefinition/ad-nested-1",
					},
					{
						ID:   "sub-action-2",
						Type: "create",
						DefinitionCanonical: "ActivityDefinition/ad-nested-2",
					},
				},
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-nested-1", Kind: "MedicationRequest",
		Code: &ActivityCode{Code: "med-1", Display: "Med 1"},
	})
	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-nested-2", Kind: "MedicationRequest",
		Code: &ActivityCode{Code: "med-2", Display: "Med 2"},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Resources) != 2 {
		t.Errorf("expected 2 resources from nested actions, got %d", len(result.Resources))
	}
}

func TestPlanDefinition_Apply_ActionRelationships(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-rel-1",
		Status: "active",
		Type:   "clinical-protocol",
		Action: []PlanAction{
			{
				ID:    "action-1",
				Title: "First Action",
				Type:  "create",
				DefinitionCanonical: "ActivityDefinition/ad-rel-1",
			},
			{
				ID:    "action-2",
				Title: "Second Action (after first)",
				Type:  "create",
				RelatedAction: []RelatedAction{
					{
						ActionID:     "action-1",
						Relationship: "after-end",
					},
				},
				DefinitionCanonical: "ActivityDefinition/ad-rel-2",
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-rel-1", Kind: "ServiceRequest",
		Code: &ActivityCode{Code: "step-1"},
	})
	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-rel-2", Kind: "ServiceRequest",
		Code: &ActivityCode{Code: "step-2"},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(result.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(result.Resources))
	}

	// Check that the RequestGroup captures the relationship
	rg := result.RequestGroup
	if rg == nil {
		t.Fatal("expected RequestGroup")
	}
	actions, ok := rg["action"].([]interface{})
	if !ok || len(actions) < 2 {
		t.Fatal("expected at least 2 actions in RequestGroup")
	}

	// Second action should have relatedAction
	secondAction, ok := actions[1].(map[string]interface{})
	if !ok {
		t.Fatal("expected second action as map")
	}
	relatedActions, ok := secondAction["relatedAction"].([]interface{})
	if !ok || len(relatedActions) == 0 {
		t.Fatal("expected relatedAction on second action")
	}
}

func TestPlanDefinition_Apply_SelectionBehavior(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-sel-1",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:                "group-1",
				Title:             "Select One",
				SelectionBehavior: "exactly-one",
				Action: []PlanAction{
					{
						ID:   "option-a",
						Type: "create",
						DefinitionCanonical: "ActivityDefinition/ad-sel-a",
					},
					{
						ID:   "option-b",
						Type: "create",
						DefinitionCanonical: "ActivityDefinition/ad-sel-b",
					},
				},
			},
		},
	}

	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-sel-a", Kind: "MedicationRequest",
		Code: &ActivityCode{Code: "option-a"},
	})
	engine.RegisterActivityDefinition(&ActivityDefinition{
		ID: "ad-sel-b", Kind: "MedicationRequest",
		Code: &ActivityCode{Code: "option-b"},
	})

	result, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	// The RequestGroup should capture the selection behavior
	rg := result.RequestGroup
	if rg == nil {
		t.Fatal("expected RequestGroup")
	}
	actions, ok := rg["action"].([]interface{})
	if !ok || len(actions) == 0 {
		t.Fatal("expected actions in RequestGroup")
	}

	groupAction, ok := actions[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected group action as map")
	}
	if groupAction["selectionBehavior"] != "exactly-one" {
		t.Errorf("expected selectionBehavior exactly-one, got %v", groupAction["selectionBehavior"])
	}
}

// ===========================================================================
// Built-in Protocol Tests
// ===========================================================================

func TestPlanDefinition_BuiltIn_DiabetesManagement(t *testing.T) {
	engine := newPlanDefinitionEngine()
	engine.RegisterBuiltins()

	pd := engine.GetPlanDefinition("diabetes-management")
	if pd == nil {
		t.Fatal("expected built-in diabetes-management PlanDefinition")
	}
	if pd.Type != "clinical-protocol" {
		t.Errorf("expected clinical-protocol type, got %v", pd.Type)
	}

	// Apply with params that indicate HbA1c > 9
	subject := testPatientDiabetes()
	params := map[string]interface{}{
		"hba1c_value": 9.5,
	}
	result, err := engine.Apply(context.Background(), pd, subject, params)
	if err != nil {
		t.Fatalf("Apply diabetes protocol error: %v", err)
	}

	if result.CarePlan == nil {
		t.Fatal("expected CarePlan from diabetes protocol")
	}

	// Should generate: HbA1c lab, metformin, follow-up, care plan goal
	if len(result.Resources) < 3 {
		t.Errorf("expected at least 3 resources from diabetes protocol, got %d", len(result.Resources))
	}

	// Verify resource types
	resourceTypes := map[string]int{}
	for _, r := range result.Resources {
		rt := r["resourceType"].(string)
		resourceTypes[rt]++
	}

	if resourceTypes["ServiceRequest"] < 1 {
		t.Error("expected at least 1 ServiceRequest (HbA1c lab)")
	}
	if resourceTypes["MedicationRequest"] < 1 {
		t.Error("expected at least 1 MedicationRequest (metformin)")
	}
}

func TestPlanDefinition_BuiltIn_SepsisBundle(t *testing.T) {
	engine := newPlanDefinitionEngine()
	engine.RegisterBuiltins()

	pd := engine.GetPlanDefinition("sepsis-bundle-sep1")
	if pd == nil {
		t.Fatal("expected built-in sepsis-bundle-sep1 PlanDefinition")
	}

	subject := testPatientSepsis()
	params := map[string]interface{}{
		"temperature": 38.5,
	}
	result, err := engine.Apply(context.Background(), pd, subject, params)
	if err != nil {
		t.Fatalf("Apply sepsis bundle error: %v", err)
	}

	// Should generate: blood cultures, antibiotics, IV fluids, lactate lab
	if len(result.Resources) < 3 {
		t.Errorf("expected at least 3 resources from sepsis bundle, got %d", len(result.Resources))
	}

	resourceTypes := map[string]int{}
	for _, r := range result.Resources {
		rt := r["resourceType"].(string)
		resourceTypes[rt]++
	}
	if resourceTypes["ServiceRequest"] < 2 {
		t.Error("expected at least 2 ServiceRequests (blood cultures + lactate)")
	}
	if resourceTypes["MedicationRequest"] < 1 {
		t.Error("expected at least 1 MedicationRequest (antibiotics)")
	}
}

func TestPlanDefinition_BuiltIn_CHFDischarge(t *testing.T) {
	engine := newPlanDefinitionEngine()
	engine.RegisterBuiltins()

	pd := engine.GetPlanDefinition("chf-discharge-protocol")
	if pd == nil {
		t.Fatal("expected built-in chf-discharge-protocol PlanDefinition")
	}

	subject := testPatientCHF()
	params := map[string]interface{}{
		"discharge": true,
	}
	result, err := engine.Apply(context.Background(), pd, subject, params)
	if err != nil {
		t.Fatalf("Apply CHF discharge protocol error: %v", err)
	}

	// Should generate: follow-up within 7 days, ACE inhibitor, daily weight task
	if len(result.Resources) < 3 {
		t.Errorf("expected at least 3 resources from CHF discharge, got %d", len(result.Resources))
	}

	resourceTypes := map[string]int{}
	for _, r := range result.Resources {
		rt := r["resourceType"].(string)
		resourceTypes[rt]++
	}
	if resourceTypes["ServiceRequest"] < 1 {
		t.Error("expected at least 1 ServiceRequest (follow-up)")
	}
	if resourceTypes["MedicationRequest"] < 1 {
		t.Error("expected at least 1 MedicationRequest (ACE inhibitor)")
	}
	if resourceTypes["Task"] < 1 {
		t.Error("expected at least 1 Task (daily weight)")
	}
}

func TestPlanDefinition_BuiltIn_PreventiveCareScreening(t *testing.T) {
	engine := newPlanDefinitionEngine()
	engine.RegisterBuiltins()

	pd := engine.GetPlanDefinition("preventive-care-screening")
	if pd == nil {
		t.Fatal("expected built-in preventive-care-screening PlanDefinition")
	}

	// Test 50-year-old male: should get colonoscopy (>= 45)
	subject := testPatientPreventiveMale50()
	params := map[string]interface{}{
		"age": 50,
	}
	result, err := engine.Apply(context.Background(), pd, subject, params)
	if err != nil {
		t.Fatalf("Apply preventive care error: %v", err)
	}

	if len(result.Resources) < 1 {
		t.Errorf("expected at least 1 resource for 50yo male, got %d", len(result.Resources))
	}

	// Verify colonoscopy is recommended
	foundColonoscopy := false
	for _, r := range result.Resources {
		if code, ok := r["code"].(map[string]interface{}); ok {
			if coding, ok := code["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if c["display"] == "Colonoscopy" {
						foundColonoscopy = true
					}
				}
			}
		}
	}
	if !foundColonoscopy {
		t.Error("expected colonoscopy recommendation for 50-year-old")
	}

	// Test 45-year-old female: should get both mammogram (>= 40) and colonoscopy (>= 45)
	subjectF := testPatientPreventiveFemale45()
	paramsF := map[string]interface{}{
		"age": 45,
	}
	resultF, err := engine.Apply(context.Background(), pd, subjectF, paramsF)
	if err != nil {
		t.Fatalf("Apply preventive care for female error: %v", err)
	}

	if len(resultF.Resources) < 2 {
		t.Errorf("expected at least 2 resources for 45yo female, got %d", len(resultF.Resources))
	}

	foundMammogram := false
	for _, r := range resultF.Resources {
		if code, ok := r["code"].(map[string]interface{}); ok {
			if coding, ok := code["coding"].([]interface{}); ok && len(coding) > 0 {
				if c, ok := coding[0].(map[string]interface{}); ok {
					if c["display"] == "Mammogram" {
						foundMammogram = true
					}
				}
			}
		}
	}
	if !foundMammogram {
		t.Error("expected mammogram recommendation for 45-year-old female")
	}
}

// ===========================================================================
// Error Cases
// ===========================================================================

func TestPlanDefinition_Apply_MissingSubject_Error(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-err-1",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{ID: "a1", Type: "create"},
		},
	}

	_, err := engine.Apply(context.Background(), pd, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil subject")
	}
}

func TestPlanDefinition_Apply_RetiredPlanDefinition_Error(t *testing.T) {
	engine := newPlanDefinitionEngine()

	pd := &PlanDefinition{
		ID:     "pd-retired-1",
		Status: "retired",
		Type:   "order-set",
		Action: []PlanAction{
			{ID: "a1", Type: "create"},
		},
	}

	_, err := engine.Apply(context.Background(), pd, testPatientDiabetes(), nil)
	if err == nil {
		t.Fatal("expected error for retired PlanDefinition")
	}
}

// ===========================================================================
// Handler $apply Test
// ===========================================================================

func TestPlanDefinition_Handler_Apply(t *testing.T) {
	h := newPlanDefinitionHandler()

	// Create a PlanDefinition with an action
	pd := PlanDefinition{
		ID:     "pd-handler-apply",
		Name:   "HandlerApplyTest",
		Status: "active",
		Type:   "order-set",
		Action: []PlanAction{
			{
				ID:    "action-1",
				Title: "Order Test",
				Type:  "create",
				DefinitionCanonical: "ActivityDefinition/ad-handler-1",
			},
		},
	}
	body, _ := json.Marshal(pd)
	c, _ := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition", string(body))
	h.CreatePlanDefinition(c)

	// Create an ActivityDefinition
	ad := ActivityDefinition{
		ID:     "ad-handler-1",
		Status: "active",
		Kind:   "ServiceRequest",
		Code:   &ActivityCode{Code: "test-1", Display: "Test"},
	}
	adBody, _ := json.Marshal(ad)
	c2, _ := setupHandlerTest(http.MethodPost, "/fhir/ActivityDefinition", string(adBody))
	h.CreateActivityDefinition(c2)

	// Apply
	subject := testPatientDiabetes()
	applyBody, _ := json.Marshal(map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []map[string]interface{}{
			{"name": "subject", "resource": subject},
		},
	})
	c3, rec3 := setupHandlerTest(http.MethodPost, "/fhir/PlanDefinition/pd-handler-apply/$apply", string(applyBody))
	c3.SetParamNames("id")
	c3.SetParamValues("pd-handler-apply")

	err := h.ApplyPlanDefinition(c3)
	if err != nil {
		t.Fatalf("ApplyPlanDefinition handler error: %v", err)
	}
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec3.Code, rec3.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rec3.Body.Bytes(), &result)
	if result["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle response, got %v", result["resourceType"])
	}
}

// ===========================================================================
// Concurrent Safety Test
// ===========================================================================

func TestPlanDefinition_Apply_Concurrent(t *testing.T) {
	engine := newPlanDefinitionEngine()
	engine.RegisterBuiltins()

	pd := engine.GetPlanDefinition("diabetes-management")
	if pd == nil {
		t.Fatal("expected built-in diabetes-management")
	}

	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subject := testPatientDiabetes()
			params := map[string]interface{}{
				"hba1c_value": 9.5,
			}
			result, err := engine.Apply(context.Background(), pd, subject, params)
			if err != nil {
				errors <- err
				return
			}
			if result == nil {
				errors <- fmt.Errorf("nil result")
				return
			}
			if result.CarePlan == nil {
				errors <- fmt.Errorf("nil CarePlan")
				return
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent apply error: %v", err)
	}
}
