package fhir

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// _has Parser Tests
// ---------------------------------------------------------------------------

func TestParseHasParamFromQuery_Valid(t *testing.T) {
	hp, err := ParseHasParamFromQuery("_has:Observation:patient:code", "8867-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hp.TargetType != "Observation" {
		t.Errorf("TargetType = %q, want %q", hp.TargetType, "Observation")
	}
	if hp.TargetParam != "patient" {
		t.Errorf("TargetParam = %q, want %q", hp.TargetParam, "patient")
	}
	if hp.SearchParam != "code" {
		t.Errorf("SearchParam = %q, want %q", hp.SearchParam, "code")
	}
	if hp.Value != "8867-4" {
		t.Errorf("Value = %q, want %q", hp.Value, "8867-4")
	}
}

func TestParseHasParamFromQuery_MissingParts(t *testing.T) {
	_, err := ParseHasParamFromQuery("_has:Observation", "8867-4")
	if err == nil {
		t.Fatal("expected error for missing parts, got nil")
	}
	if !strings.Contains(err.Error(), "must have format") {
		t.Errorf("error = %q, want it to mention format requirement", err.Error())
	}
}

func TestParseHasParamFromQuery_EmptyValue(t *testing.T) {
	_, err := ParseHasParamFromQuery("_has:Observation:patient:code", "")
	if err == nil {
		t.Fatal("expected error for empty value, got nil")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("error = %q, want it to mention empty value", err.Error())
	}
}

func TestParseHasParamFromQuery_ValidCondition(t *testing.T) {
	hp, err := ParseHasParamFromQuery("_has:Condition:patient:clinical-status", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hp.TargetType != "Condition" {
		t.Errorf("TargetType = %q, want %q", hp.TargetType, "Condition")
	}
	if hp.TargetParam != "patient" {
		t.Errorf("TargetParam = %q, want %q", hp.TargetParam, "patient")
	}
	if hp.SearchParam != "clinical-status" {
		t.Errorf("SearchParam = %q, want %q", hp.SearchParam, "clinical-status")
	}
	if hp.Value != "active" {
		t.Errorf("Value = %q, want %q", hp.Value, "active")
	}
}

func TestParseHasParamFromQuery_MedicationRequest(t *testing.T) {
	hp, err := ParseHasParamFromQuery("_has:MedicationRequest:patient:status", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hp.TargetType != "MedicationRequest" {
		t.Errorf("TargetType = %q, want %q", hp.TargetType, "MedicationRequest")
	}
	if hp.SearchParam != "status" {
		t.Errorf("SearchParam = %q, want %q", hp.SearchParam, "status")
	}
}

func TestParseHasParamFromQuery_InvalidFormat(t *testing.T) {
	_, err := ParseHasParamFromQuery("randomstring", "value")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	if !strings.Contains(err.Error(), "must start with '_has:'") {
		t.Errorf("error = %q, want it to mention _has: prefix", err.Error())
	}
}

func TestParseHasParamFromQuery_ExtraParts(t *testing.T) {
	_, err := ParseHasParamFromQuery("_has:Observation:patient:code:extra", "8867-4")
	if err == nil {
		t.Fatal("expected error for extra parts, got nil")
	}
	if !strings.Contains(err.Error(), "too many parts") {
		t.Errorf("error = %q, want it to mention too many parts", err.Error())
	}
}

func TestParseHasParamFromQuery_KeyOnly(t *testing.T) {
	_, err := ParseHasParamFromQuery("_has:", "value")
	if err == nil {
		t.Fatal("expected error for _has: with no parts, got nil")
	}
}

func TestParseHasParamFromQuery_UnknownResourceType(t *testing.T) {
	_, err := ParseHasParamFromQuery("_has:FakeResource:patient:code", "123")
	if err == nil {
		t.Fatal("expected error for unknown resource type, got nil")
	}
	if !strings.Contains(err.Error(), "unknown FHIR resource type") {
		t.Errorf("error = %q, want it to mention unknown resource type", err.Error())
	}
}

func TestParseHasParamFromQuery_TwoParts(t *testing.T) {
	_, err := ParseHasParamFromQuery("_has:Observation:patient", "value")
	if err == nil {
		t.Fatal("expected error for two parts, got nil")
	}
}

// ---------------------------------------------------------------------------
// _has SQL Tests
// ---------------------------------------------------------------------------

func TestHasQueryBuilder_ObservationCode(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "Observation",
		TargetParam: "patient",
		SearchParam: "code",
		Value:       "8867-4",
	}
	sql, args, err := b.BuildSQL("patients", has)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "EXISTS (SELECT 1 FROM observations WHERE observations.patient_id = patients.id AND observations.code = $1)"
	if sql != expected {
		t.Errorf("SQL = %q, want %q", sql, expected)
	}
	if len(args) != 1 || args[0] != "8867-4" {
		t.Errorf("args = %v, want [8867-4]", args)
	}
}

func TestHasQueryBuilder_ConditionStatus(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "Condition",
		TargetParam: "patient",
		SearchParam: "clinical-status",
		Value:       "active",
	}
	sql, args, err := b.BuildSQL("patients", has)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "EXISTS (SELECT 1 FROM conditions WHERE conditions.patient_id = patients.id AND conditions.clinical_status = $1)"
	if sql != expected {
		t.Errorf("SQL = %q, want %q", sql, expected)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("args = %v, want [active]", args)
	}
}

func TestHasQueryBuilder_UnknownResource(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "FakeResource",
		TargetParam: "patient",
		SearchParam: "code",
		Value:       "123",
	}
	_, _, err := b.BuildSQL("patients", has)
	if err == nil {
		t.Fatal("expected error for unknown resource, got nil")
	}
	if !strings.Contains(err.Error(), "unknown resource type") {
		t.Errorf("error = %q, want it to mention unknown resource type", err.Error())
	}
}

func TestHasQueryBuilder_SubjectReference(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "Observation",
		TargetParam: "subject",
		SearchParam: "code",
		Value:       "1234",
	}
	sql, _, err := b.BuildSQL("patients", has)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "subject" should map to "patient_id"
	if !strings.Contains(sql, "observations.patient_id") {
		t.Errorf("SQL = %q, want it to contain 'observations.patient_id'", sql)
	}
}

func TestHasQueryBuilder_BindParams(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "Encounter",
		TargetParam: "patient",
		SearchParam: "status",
		Value:       "finished",
	}
	sql, args, err := b.BuildSQL("patients", has)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "$1") {
		t.Errorf("SQL should contain $1 placeholder, got %q", sql)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 bind param, got %d", len(args))
	}
	if args[0] != "finished" {
		t.Errorf("bind param = %v, want 'finished'", args[0])
	}
}

func TestHasQueryBuilder_NilParam(t *testing.T) {
	b := NewHasQueryBuilder()
	_, _, err := b.BuildSQL("patients", nil)
	if err == nil {
		t.Fatal("expected error for nil has param, got nil")
	}
}

func TestHasQueryBuilder_UnknownReferenceParam(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "Observation",
		TargetParam: "unknownref",
		SearchParam: "code",
		Value:       "123",
	}
	_, _, err := b.BuildSQL("patients", has)
	if err == nil {
		t.Fatal("expected error for unknown reference param, got nil")
	}
	if !strings.Contains(err.Error(), "unknown reference parameter") {
		t.Errorf("error = %q, want it to mention unknown reference parameter", err.Error())
	}
}

func TestHasQueryBuilder_UnknownSearchParam(t *testing.T) {
	b := NewHasQueryBuilder()
	has := &HasParam{
		TargetType:  "Observation",
		TargetParam: "patient",
		SearchParam: "unknownsearch",
		Value:       "123",
	}
	_, _, err := b.BuildSQL("patients", has)
	if err == nil {
		t.Fatal("expected error for unknown search param, got nil")
	}
	if !strings.Contains(err.Error(), "unknown search parameter") {
		t.Errorf("error = %q, want it to mention unknown search parameter", err.Error())
	}
}

// ---------------------------------------------------------------------------
// _filter Parser Tests
// ---------------------------------------------------------------------------

func TestParseFilter_SimpleEq(t *testing.T) {
	filters, err := ParseFilter(`name eq "Smith"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	f := filters[0]
	if f.Field != "name" {
		t.Errorf("Field = %q, want %q", f.Field, "name")
	}
	if f.Op != FilterOpEq {
		t.Errorf("Op = %q, want %q", f.Op, FilterOpEq)
	}
	if f.Value != "Smith" {
		t.Errorf("Value = %q, want %q", f.Value, "Smith")
	}
	if f.Logic != FilterLogicNone {
		t.Errorf("Logic = %q, want empty", f.Logic)
	}
}

func TestParseFilter_DateGe(t *testing.T) {
	filters, err := ParseFilter("birthdate ge 1990-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	f := filters[0]
	if f.Field != "birthdate" {
		t.Errorf("Field = %q, want %q", f.Field, "birthdate")
	}
	if f.Op != FilterOpGe {
		t.Errorf("Op = %q, want %q", f.Op, FilterOpGe)
	}
	if f.Value != "1990-01-01" {
		t.Errorf("Value = %q, want %q", f.Value, "1990-01-01")
	}
}

func TestParseFilter_AndCombined(t *testing.T) {
	filters, err := ParseFilter(`name eq "Smith" and birthdate ge 1990-01-01`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}
	if filters[0].Field != "name" || filters[0].Op != FilterOpEq || filters[0].Value != "Smith" {
		t.Errorf("first filter = %+v, unexpected", filters[0])
	}
	if filters[0].Logic != FilterLogicAnd {
		t.Errorf("first filter Logic = %q, want %q", filters[0].Logic, FilterLogicAnd)
	}
	if filters[1].Field != "birthdate" || filters[1].Op != FilterOpGe || filters[1].Value != "1990-01-01" {
		t.Errorf("second filter = %+v, unexpected", filters[1])
	}
	if filters[1].Logic != FilterLogicNone {
		t.Errorf("second filter Logic = %q, want empty", filters[1].Logic)
	}
}

func TestParseFilter_OrCombined(t *testing.T) {
	filters, err := ParseFilter(`status eq "active" or status eq "inactive"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}
	if filters[0].Logic != FilterLogicOr {
		t.Errorf("first filter Logic = %q, want %q", filters[0].Logic, FilterLogicOr)
	}
	if filters[0].Value != "active" {
		t.Errorf("first filter Value = %q, want %q", filters[0].Value, "active")
	}
	if filters[1].Value != "inactive" {
		t.Errorf("second filter Value = %q, want %q", filters[1].Value, "inactive")
	}
}

func TestParseFilter_Contains(t *testing.T) {
	filters, err := ParseFilter(`name co "mit"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if filters[0].Op != FilterOpCo {
		t.Errorf("Op = %q, want %q", filters[0].Op, FilterOpCo)
	}
	if filters[0].Value != "mit" {
		t.Errorf("Value = %q, want %q", filters[0].Value, "mit")
	}
}

func TestParseFilter_StartsWith(t *testing.T) {
	filters, err := ParseFilter(`name sw "Sm"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if filters[0].Op != FilterOpSw {
		t.Errorf("Op = %q, want %q", filters[0].Op, FilterOpSw)
	}
}

func TestParseFilter_QuotedValue(t *testing.T) {
	filters, err := ParseFilter(`name eq "John Smith"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	// The value should have quotes stripped.
	if filters[0].Value != "John Smith" {
		t.Errorf("Value = %q, want %q", filters[0].Value, "John Smith")
	}
}

func TestParseFilter_UnquotedValue(t *testing.T) {
	filters, err := ParseFilter("birthdate ge 1990-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if filters[0].Value != "1990-01-01" {
		t.Errorf("Value = %q, want %q", filters[0].Value, "1990-01-01")
	}
}

func TestParseFilter_Empty(t *testing.T) {
	_, err := ParseFilter("")
	if err == nil {
		t.Fatal("expected error for empty expression, got nil")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("error = %q, want it to mention empty expression", err.Error())
	}
}

func TestParseFilter_InvalidOp(t *testing.T) {
	_, err := ParseFilter(`name xyz "Smith"`)
	if err == nil {
		t.Fatal("expected error for invalid operator, got nil")
	}
	if !strings.Contains(err.Error(), "unknown filter operator") {
		t.Errorf("error = %q, want it to mention unknown operator", err.Error())
	}
}

func TestParseFilter_EndsWith(t *testing.T) {
	filters, err := ParseFilter(`name ew "son"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if filters[0].Op != FilterOpEw {
		t.Errorf("Op = %q, want %q", filters[0].Op, FilterOpEw)
	}
	if filters[0].Value != "son" {
		t.Errorf("Value = %q, want %q", filters[0].Value, "son")
	}
}

func TestParseFilter_MissingValue(t *testing.T) {
	_, err := ParseFilter("name eq")
	if err == nil {
		t.Fatal("expected error for missing value, got nil")
	}
}

func TestParseFilter_ThreeExpressions(t *testing.T) {
	filters, err := ParseFilter(`name eq "Smith" and gender eq "male" and active eq "true"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(filters))
	}
	if filters[0].Logic != FilterLogicAnd {
		t.Errorf("first Logic = %q, want %q", filters[0].Logic, FilterLogicAnd)
	}
	if filters[1].Logic != FilterLogicAnd {
		t.Errorf("second Logic = %q, want %q", filters[1].Logic, FilterLogicAnd)
	}
	if filters[2].Logic != FilterLogicNone {
		t.Errorf("third Logic = %q, want empty", filters[2].Logic)
	}
}

// ---------------------------------------------------------------------------
// _filter SQL Tests
// ---------------------------------------------------------------------------

func TestFilterToSQL_SimpleEq(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "name", Op: FilterOpEq, Value: "Smith"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "last_name = $1" {
		t.Errorf("SQL = %q, want %q", sql, "last_name = $1")
	}
	if len(args) != 1 || args[0] != "Smith" {
		t.Errorf("args = %v, want [Smith]", args)
	}
}

func TestFilterToSQL_Contains(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "name", Op: FilterOpCo, Value: "mit"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "last_name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "last_name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "%mit%" {
		t.Errorf("args = %v, want [%%mit%%]", args)
	}
}

func TestFilterToSQL_AndCombined(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "name", Op: FilterOpEq, Value: "Smith", Logic: FilterLogicAnd},
		{Field: "gender", Op: FilterOpEq, Value: "male"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "last_name = $1 AND gender = $2"
	if sql != expected {
		t.Errorf("SQL = %q, want %q", sql, expected)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "Smith" || args[1] != "male" {
		t.Errorf("args = %v, want [Smith male]", args)
	}
}

func TestFilterToSQL_DateComparison(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "birthdate", Op: FilterOpGe, Value: "1990-01-01"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "birth_date >= $1" {
		t.Errorf("SQL = %q, want %q", sql, "birth_date >= $1")
	}
	if len(args) != 1 || args[0] != "1990-01-01" {
		t.Errorf("args = %v, want [1990-01-01]", args)
	}
}

func TestFilterToSQL_OrCombined(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "status", Op: FilterOpEq, Value: "active", Logic: FilterLogicOr},
		{Field: "status", Op: FilterOpEq, Value: "inactive"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "status = $1 OR status = $2"
	if sql != expected {
		t.Errorf("SQL = %q, want %q", sql, expected)
	}
	if len(args) != 2 || args[0] != "active" || args[1] != "inactive" {
		t.Errorf("args = %v, want [active inactive]", args)
	}
}

func TestFilterToSQL_StartsWith(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "name", Op: FilterOpSw, Value: "Sm"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "last_name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "last_name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "Sm%" {
		t.Errorf("args = %v, want [Sm%%]", args)
	}
}

func TestFilterToSQL_EndsWith(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "name", Op: FilterOpEw, Value: "son"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "last_name ILIKE $1" {
		t.Errorf("SQL = %q, want %q", sql, "last_name ILIKE $1")
	}
	if len(args) != 1 || args[0] != "%son" {
		t.Errorf("args = %v, want [%%son]", args)
	}
}

func TestFilterToSQL_UnknownField(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "unknownfield", Op: FilterOpEq, Value: "test"},
	}
	_, _, err := FilterToSQL(filters, "Patient")
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "unknown filter field") {
		t.Errorf("error = %q, want it to mention unknown field", err.Error())
	}
}

func TestFilterToSQL_EmptyFilters(t *testing.T) {
	_, _, err := FilterToSQL(nil, "Patient")
	if err == nil {
		t.Fatal("expected error for nil filters, got nil")
	}
}

func TestFilterToSQL_NotEquals(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "status", Op: FilterOpNe, Value: "inactive"},
	}
	sql, args, err := FilterToSQL(filters, "Patient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sql != "status != $1" {
		t.Errorf("SQL = %q, want %q", sql, "status != $1")
	}
	if len(args) != 1 || args[0] != "inactive" {
		t.Errorf("args = %v, want [inactive]", args)
	}
}

func TestFilterToSQL_GtLt(t *testing.T) {
	filters := []*FilterExpression{
		{Field: "date", Op: FilterOpGt, Value: "2023-01-01", Logic: FilterLogicAnd},
		{Field: "date", Op: FilterOpLt, Value: "2023-12-31"},
	}
	sql, args, err := FilterToSQL(filters, "Observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "effective_date > $1 AND effective_date < $2"
	if sql != expected {
		t.Errorf("SQL = %q, want %q", sql, expected)
	}
	if len(args) != 2 || args[0] != "2023-01-01" || args[1] != "2023-12-31" {
		t.Errorf("args = %v, want [2023-01-01 2023-12-31]", args)
	}
}

// ---------------------------------------------------------------------------
// Mapping table tests
// ---------------------------------------------------------------------------

func TestResourceTableMap_Coverage(t *testing.T) {
	expectedResources := []string{
		"Patient", "Observation", "Condition", "Encounter",
		"Procedure", "MedicationRequest", "AllergyIntolerance",
		"DiagnosticReport", "Immunization", "CarePlan",
		"Claim", "ServiceRequest",
	}
	for _, r := range expectedResources {
		if _, ok := resourceTableMap[r]; !ok {
			t.Errorf("resourceTableMap missing resource %q", r)
		}
	}
}

func TestReferenceColumnMap_Coverage(t *testing.T) {
	expectedRefs := []string{"patient", "subject", "encounter", "performer", "requester", "author"}
	for _, r := range expectedRefs {
		if _, ok := referenceColumnMap[r]; !ok {
			t.Errorf("referenceColumnMap missing reference param %q", r)
		}
	}
}

func TestSearchColumnMap_Coverage(t *testing.T) {
	expectedParams := []string{"code", "status", "category", "date", "type", "clinical-status"}
	for _, p := range expectedParams {
		if _, ok := searchColumnMap[p]; !ok {
			t.Errorf("searchColumnMap missing search param %q", p)
		}
	}
}

func TestNewHasQueryBuilder(t *testing.T) {
	b := NewHasQueryBuilder()
	if b == nil {
		t.Fatal("NewHasQueryBuilder returned nil")
	}
	if !b.knownTypes["Patient"] {
		t.Error("knownTypes should contain Patient")
	}
	if !b.knownTypes["Observation"] {
		t.Error("knownTypes should contain Observation")
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`hello`, "hello"},
		{`""`, ""},
		{`"`, `"`},
		{`"a`, `"a`},
		{`a"`, `a"`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripQuotes(tt.input)
			if got != tt.want {
				t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
