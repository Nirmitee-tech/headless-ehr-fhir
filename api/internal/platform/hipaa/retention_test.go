package hipaa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func testLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.Disabled)
}

// --- DefaultRetentionPolicies tests ---

func TestDefaultRetentionPolicies_CoversRequiredTypes(t *testing.T) {
	policies := DefaultRetentionPolicies()
	required := map[string]bool{
		"medical_record":  false,
		"audit_log":       false,
		"billing_record":  false,
		"consent_record":  false,
		"hipaa_access_log": false,
		"temporary_data":  false,
	}

	for _, p := range policies {
		if _, ok := required[p.ResourceType]; ok {
			required[p.ResourceType] = true
		}
	}

	for rt, found := range required {
		if !found {
			t.Errorf("DefaultRetentionPolicies missing required type: %s", rt)
		}
	}
}

func TestDefaultRetentionPolicies_MedicalRecords6Years(t *testing.T) {
	policies := DefaultRetentionPolicies()
	for _, p := range policies {
		if p.ResourceType == "medical_record" {
			if p.RetentionDays < 2190 {
				t.Errorf("medical_record retention should be at least 6 years (2190 days), got %d", p.RetentionDays)
			}
			if p.PurgeAfter != 0 {
				t.Errorf("medical_record should never be purged (PurgeAfter=0), got %d", p.PurgeAfter)
			}
			return
		}
	}
	t.Error("medical_record policy not found")
}

func TestDefaultRetentionPolicies_AuditLog6Years(t *testing.T) {
	policies := DefaultRetentionPolicies()
	for _, p := range policies {
		if p.ResourceType == "audit_log" {
			if p.RetentionDays < 2190 {
				t.Errorf("audit_log retention should be at least 6 years (2190 days), got %d", p.RetentionDays)
			}
			return
		}
	}
	t.Error("audit_log policy not found")
}

func TestDefaultRetentionPolicies_BillingRecords7Years(t *testing.T) {
	policies := DefaultRetentionPolicies()
	for _, p := range policies {
		if p.ResourceType == "billing_record" {
			if p.RetentionDays < 2555 {
				t.Errorf("billing_record retention should be at least 7 years (2555 days), got %d", p.RetentionDays)
			}
			return
		}
	}
	t.Error("billing_record policy not found")
}

func TestDefaultRetentionPolicies_ConsentRecords10Years(t *testing.T) {
	policies := DefaultRetentionPolicies()
	for _, p := range policies {
		if p.ResourceType == "consent_record" {
			if p.RetentionDays < 3650 {
				t.Errorf("consent_record retention should be at least 10 years (3650 days), got %d", p.RetentionDays)
			}
			if p.PurgeAfter != 0 {
				t.Errorf("consent_record should never be purged (PurgeAfter=0), got %d", p.PurgeAfter)
			}
			return
		}
	}
	t.Error("consent_record policy not found")
}

func TestDefaultRetentionPolicies_TempData90Days(t *testing.T) {
	policies := DefaultRetentionPolicies()
	for _, p := range policies {
		if p.ResourceType == "temporary_data" {
			if p.RetentionDays != 90 {
				t.Errorf("temporary_data retention should be 90 days, got %d", p.RetentionDays)
			}
			if p.PurgeAfter != 90 {
				t.Errorf("temporary_data purge should be 90 days, got %d", p.PurgeAfter)
			}
			return
		}
	}
	t.Error("temporary_data policy not found")
}

func TestDefaultRetentionPolicies_AllHaveDescriptions(t *testing.T) {
	policies := DefaultRetentionPolicies()
	for _, p := range policies {
		if p.Description == "" {
			t.Errorf("policy %s has no description", p.ResourceType)
		}
	}
}

// --- RetentionService tests ---

func TestRetentionService_GetPolicy_Known(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	policy := svc.GetPolicy("medical_record")
	if policy == nil {
		t.Fatal("expected policy for medical_record, got nil")
	}
	if policy.ResourceType != "medical_record" {
		t.Errorf("expected resource type medical_record, got %s", policy.ResourceType)
	}
}

func TestRetentionService_GetPolicy_Unknown(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	policy := svc.GetPolicy("nonexistent_type")
	if policy != nil {
		t.Errorf("expected nil for unknown resource type, got %+v", policy)
	}
}

func TestRetentionService_GetAllPolicies(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	policies := svc.GetAllPolicies()
	if len(policies) != 6 {
		t.Errorf("expected 6 policies, got %d", len(policies))
	}
}

func TestRetentionService_CheckRetention_Active(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	// A medical record created 1 year ago should be active
	createdAt := time.Now().UTC().AddDate(-1, 0, 0)
	status := svc.CheckRetention("medical_record", createdAt)

	if status.State != RetentionStateActive {
		t.Errorf("expected state %s, got %s", RetentionStateActive, status.State)
	}
	if status.PolicyName != "medical_record" {
		t.Errorf("expected policy name medical_record, got %s", status.PolicyName)
	}
}

func TestRetentionService_CheckRetention_ArchiveEligible(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	// An audit log created 4 years ago should be archive-eligible (ArchiveAfter=3 years)
	createdAt := time.Now().UTC().AddDate(-4, 0, 0)
	status := svc.CheckRetention("audit_log", createdAt)

	if status.State != RetentionStateArchiveEligible {
		t.Errorf("expected state %s, got %s", RetentionStateArchiveEligible, status.State)
	}
	if status.PolicyName != "audit_log" {
		t.Errorf("expected policy name audit_log, got %s", status.PolicyName)
	}
}

func TestRetentionService_CheckRetention_PurgeEligible(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	// Temporary data created 100 days ago should be purge-eligible (PurgeAfter=90 days)
	createdAt := time.Now().UTC().AddDate(0, 0, -100)
	status := svc.CheckRetention("temporary_data", createdAt)

	if status.State != RetentionStatePurgeEligible {
		t.Errorf("expected state %s, got %s", RetentionStatePurgeEligible, status.State)
	}
	if status.PolicyName != "temporary_data" {
		t.Errorf("expected policy name temporary_data, got %s", status.PolicyName)
	}
}

func TestRetentionService_CheckRetention_UnknownType(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	status := svc.CheckRetention("unknown_type", time.Now().UTC().AddDate(-10, 0, 0))

	if status.State != RetentionStateActive {
		t.Errorf("expected state %s for unknown type, got %s", RetentionStateActive, status.State)
	}
	if status.PolicyName != "unknown" {
		t.Errorf("expected policy name 'unknown', got %s", status.PolicyName)
	}
}

func TestRetentionService_CheckRetention_NeverPurge(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	// Medical records should never reach purge-eligible (PurgeAfter=0)
	createdAt := time.Now().UTC().AddDate(-20, 0, 0) // 20 years old
	status := svc.CheckRetention("medical_record", createdAt)

	if status.State == RetentionStatePurgeEligible {
		t.Error("medical records should never be purge-eligible")
	}
}

func TestRetentionService_CheckRetention_BillingPurge(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	// A billing record created 9 years ago should be purge-eligible (PurgeAfter=8 years)
	createdAt := time.Now().UTC().AddDate(-9, 0, 0)
	status := svc.CheckRetention("billing_record", createdAt)

	if status.State != RetentionStatePurgeEligible {
		t.Errorf("expected state %s, got %s", RetentionStatePurgeEligible, status.State)
	}
}

func TestRetentionService_CheckRetention_BillingActive(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())

	// A billing record created 2 years ago should be active
	createdAt := time.Now().UTC().AddDate(-2, 0, 0)
	status := svc.CheckRetention("billing_record", createdAt)

	if status.State != RetentionStateActive {
		t.Errorf("expected state %s, got %s", RetentionStateActive, status.State)
	}
}

// --- Handler tests ---

func TestRetentionHandler_ListPolicies(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	h := NewRetentionHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/retention-policies", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleListPolicies(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok || int(total) != 6 {
		t.Errorf("expected total 6, got %v", resp["total"])
	}

	policies, ok := resp["policies"].([]interface{})
	if !ok || len(policies) != 6 {
		t.Errorf("expected 6 policies in response, got %v", len(policies))
	}
}

func TestRetentionHandler_GetPolicy_Found(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	h := NewRetentionHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/retention-policies/medical_record", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType")
	c.SetParamValues("medical_record")

	if err := h.HandleGetPolicy(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var policy RetentionPolicy
	if err := json.Unmarshal(rec.Body.Bytes(), &policy); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if policy.ResourceType != "medical_record" {
		t.Errorf("expected resource type medical_record, got %s", policy.ResourceType)
	}
}

func TestRetentionHandler_GetPolicy_NotFound(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	h := NewRetentionHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/retention-policies/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("resourceType")
	c.SetParamValues("nonexistent")

	if err := h.HandleGetPolicy(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestRetentionHandler_RetentionStatus(t *testing.T) {
	svc := NewRetentionService(DefaultRetentionPolicies(), testLogger())
	h := NewRetentionHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/retention-status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.HandleRetentionStatus(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	summaries, ok := resp["summaries"].([]interface{})
	if !ok {
		t.Fatal("expected summaries array in response")
	}
	if len(summaries) != 6 {
		t.Errorf("expected 6 summaries, got %d", len(summaries))
	}
}

func TestRetentionService_CustomPolicies(t *testing.T) {
	custom := []RetentionPolicy{
		{
			ResourceType:  "custom_type",
			RetentionDays: 365,
			ArchiveAfter:  180,
			PurgeAfter:    730,
			Description:   "Custom policy",
		},
	}
	svc := NewRetentionService(custom, testLogger())

	policy := svc.GetPolicy("custom_type")
	if policy == nil {
		t.Fatal("expected custom policy, got nil")
	}
	if policy.RetentionDays != 365 {
		t.Errorf("expected 365 retention days, got %d", policy.RetentionDays)
	}

	all := svc.GetAllPolicies()
	if len(all) != 1 {
		t.Errorf("expected 1 policy, got %d", len(all))
	}
}
