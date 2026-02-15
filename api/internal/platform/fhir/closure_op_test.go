package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== ClosureManager Tests ===========

func TestClosureManager_Initialize(t *testing.T) {
	mgr := NewClosureManager()
	table, err := mgr.InitializeClosure("test-closure")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if table == nil {
		t.Fatal("expected non-nil table")
	}
	if table.Name != "test-closure" {
		t.Errorf("expected name test-closure, got %s", table.Name)
	}
	if table.version != 0 {
		t.Errorf("expected version 0, got %d", table.version)
	}
}

func TestClosureManager_InitializeDuplicate(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("dup-name")
	if err != nil {
		t.Fatalf("unexpected error on first init: %v", err)
	}
	_, err = mgr.InitializeClosure("dup-name")
	if err == nil {
		t.Fatal("expected error for duplicate closure name")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got %s", err.Error())
	}
}

func TestClosureManager_ProcessConcepts(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("concept-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	concepts := []ClosureConcept{
		{System: systemSNOMED, Code: "73211009", Display: "Diabetes mellitus"},
		{System: systemSNOMED, Code: "44054006", Display: "Type 2 diabetes mellitus"},
	}

	cm, err := mgr.ProcessConcepts("concept-test", concepts)
	if err != nil {
		t.Fatalf("process error: %v", err)
	}
	if cm == nil {
		t.Fatal("expected non-nil ConceptMap")
	}
	if cm.ResourceType != "ConceptMap" {
		t.Errorf("expected resourceType ConceptMap, got %s", cm.ResourceType)
	}
	if cm.Name != "concept-test" {
		t.Errorf("expected name concept-test, got %s", cm.Name)
	}
}

func TestClosureManager_TransitiveClosure(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("transitive-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	// Diabetes mellitus → Type 2 diabetes → Type 2 with renal complications
	// Transitive: Diabetes mellitus should subsume Type 2 with renal complications
	concepts := []ClosureConcept{
		{System: systemSNOMED, Code: "73211009", Display: "Diabetes mellitus"},
		{System: systemSNOMED, Code: "44054006", Display: "Type 2 diabetes mellitus"},
		{System: systemSNOMED, Code: "313436004", Display: "Type 2 diabetes with renal complications"},
	}

	cm, err := mgr.ProcessConcepts("transitive-test", concepts)
	if err != nil {
		t.Fatalf("process error: %v", err)
	}

	// Check that transitive relationship is captured
	foundTransitive := false
	for _, group := range cm.Groups {
		for _, elem := range group.Elements {
			if elem.Code == "73211009" {
				for _, target := range elem.Targets {
					if target.Code == "313436004" && target.Equivalence == "subsumes" {
						foundTransitive = true
					}
				}
			}
		}
	}
	if !foundTransitive {
		t.Error("expected transitive relationship: 73211009 subsumes 313436004")
	}
}

func TestClosureManager_GetClosure(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("get-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}
	table, err := mgr.GetClosure("get-test")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if table.Name != "get-test" {
		t.Errorf("expected name get-test, got %s", table.Name)
	}
}

func TestClosureManager_GetNotFound(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.GetClosure("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent closure")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %s", err.Error())
	}
}

func TestClosureManager_DeleteClosure(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("delete-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}
	err = mgr.DeleteClosure("delete-test")
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	_, err = mgr.GetClosure("delete-test")
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestClosureManager_VersionIncrement(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("version-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	concepts := []ClosureConcept{
		{System: systemSNOMED, Code: "73211009"},
	}

	_, err = mgr.ProcessConcepts("version-test", concepts)
	if err != nil {
		t.Fatalf("process error: %v", err)
	}

	table, _ := mgr.GetClosure("version-test")
	if table.version != 1 {
		t.Errorf("expected version 1 after first process, got %d", table.version)
	}

	_, err = mgr.ProcessConcepts("version-test", concepts)
	if err != nil {
		t.Fatalf("second process error: %v", err)
	}

	table, _ = mgr.GetClosure("version-test")
	if table.version != 2 {
		t.Errorf("expected version 2 after second process, got %d", table.version)
	}
}

func TestClosureManager_SNOMEDHierarchy(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("snomed-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	// Use known SNOMED hierarchy: Heart disease → Heart failure → Left heart failure
	concepts := []ClosureConcept{
		{System: systemSNOMED, Code: "56265001", Display: "Heart disease"},
		{System: systemSNOMED, Code: "84114007", Display: "Heart failure"},
		{System: systemSNOMED, Code: "85232009", Display: "Left heart failure"},
	}

	cm, err := mgr.ProcessConcepts("snomed-test", concepts)
	if err != nil {
		t.Fatalf("process error: %v", err)
	}

	// Verify direct: Heart disease subsumes Heart failure
	foundDirect := false
	foundTransitive := false
	for _, group := range cm.Groups {
		for _, elem := range group.Elements {
			if elem.Code == "56265001" {
				for _, target := range elem.Targets {
					if target.Code == "84114007" && target.Equivalence == "subsumes" {
						foundDirect = true
					}
					if target.Code == "85232009" && target.Equivalence == "subsumes" {
						foundTransitive = true
					}
				}
			}
		}
	}
	if !foundDirect {
		t.Error("expected 56265001 (Heart disease) subsumes 84114007 (Heart failure)")
	}
	if !foundTransitive {
		t.Error("expected 56265001 (Heart disease) subsumes 85232009 (Left heart failure)")
	}
}

func TestClosureManager_MultipleConcepts(t *testing.T) {
	mgr := NewClosureManager()
	_, err := mgr.InitializeClosure("multi-test")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	concepts := []ClosureConcept{
		{System: systemSNOMED, Code: "50043002", Display: "Respiratory disorder"},
		{System: systemSNOMED, Code: "195967001", Display: "Asthma"},
		{System: systemSNOMED, Code: "389145006", Display: "Allergic asthma"},
		{System: systemSNOMED, Code: "13645005", Display: "COPD"},
		{System: systemSNOMED, Code: "233604007", Display: "Pneumonia"},
	}

	cm, err := mgr.ProcessConcepts("multi-test", concepts)
	if err != nil {
		t.Fatalf("process error: %v", err)
	}

	if len(cm.Groups) == 0 {
		t.Fatal("expected at least one group")
	}

	// Count total relationships found
	totalRelationships := 0
	for _, group := range cm.Groups {
		for _, elem := range group.Elements {
			totalRelationships += len(elem.Targets)
		}
	}

	// Respiratory disorder subsumes: Asthma, Allergic asthma, COPD, Pneumonia
	// Asthma subsumes: Allergic asthma
	// Expected at least 5 relationships
	if totalRelationships < 5 {
		t.Errorf("expected at least 5 relationships, got %d", totalRelationships)
	}
}

func TestClosureManager_ConcurrentAccess(t *testing.T) {
	mgr := NewClosureManager()

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	// Concurrently initialize multiple tables
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "concurrent-" + strings.Repeat("x", idx)
			_, err := mgr.InitializeClosure(name)
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent error: %v", err)
	}

	// Concurrently process concepts
	_, _ = mgr.InitializeClosure("race-test")
	var wg2 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			concepts := []ClosureConcept{
				{System: systemSNOMED, Code: "73211009"},
				{System: systemSNOMED, Code: "44054006"},
			}
			_, _ = mgr.ProcessConcepts("race-test", concepts)
		}()
	}
	wg2.Wait()
}

// =========== ClosureHandler Tests ===========

func TestClosureHandler_Initialize(t *testing.T) {
	mgr := NewClosureManager()
	handler := NewClosureHandler(mgr)
	e := echo.New()

	body := `{"name": "handler-init-test"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$closure",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleClosure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["resourceType"] != "ConceptMap" {
		t.Errorf("expected resourceType ConceptMap, got %v", result["resourceType"])
	}
}

func TestClosureHandler_ProcessConcepts(t *testing.T) {
	mgr := NewClosureManager()
	_, _ = mgr.InitializeClosure("handler-process-test")
	handler := NewClosureHandler(mgr)
	e := echo.New()

	body := `{
		"name": "handler-process-test",
		"concept": [
			{"system": "http://snomed.info/sct", "code": "73211009", "display": "Diabetes mellitus"},
			{"system": "http://snomed.info/sct", "code": "44054006", "display": "Type 2 diabetes"}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$closure",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleClosure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["resourceType"] != "ConceptMap" {
		t.Errorf("expected resourceType ConceptMap, got %v", result["resourceType"])
	}
}

func TestClosureHandler_MissingName(t *testing.T) {
	mgr := NewClosureManager()
	handler := NewClosureHandler(mgr)
	e := echo.New()

	body := `{"concept": [{"system": "http://snomed.info/sct", "code": "73211009"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/CodeSystem/$closure",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleClosure(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
