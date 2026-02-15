package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// =========== Mock PatientSearcher ===========

type mockSearcher struct {
	patients []PatientRecord
	err      error
}

func (m *mockSearcher) SearchByDemographics(ctx context.Context, params map[string]string, limit int) ([]PatientRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.patients, nil
}

// =========== Jaro-Winkler Tests ===========

func TestJaroWinkler_ExactMatch(t *testing.T) {
	score := jaroWinklerSimilarity("Smith", "Smith")
	if score != 1.0 {
		t.Errorf("expected 1.0 for exact match, got %f", score)
	}
}

func TestJaroWinkler_Similar(t *testing.T) {
	score := jaroWinklerSimilarity("Martha", "Marhta")
	if score <= 0.9 {
		t.Errorf("expected > 0.9 for Martha/Marhta, got %f", score)
	}
}

func TestJaroWinkler_Different(t *testing.T) {
	score := jaroWinklerSimilarity("Smith", "Jones")
	if score >= 0.5 {
		t.Errorf("expected < 0.5 for Smith/Jones, got %f", score)
	}
}

func TestJaroWinkler_Empty(t *testing.T) {
	score := jaroWinklerSimilarity("", "Smith")
	if score != 0.0 {
		t.Errorf("expected 0.0 for empty string, got %f", score)
	}

	score2 := jaroWinklerSimilarity("Smith", "")
	if score2 != 0.0 {
		t.Errorf("expected 0.0 for empty string, got %f", score2)
	}

	score3 := jaroWinklerSimilarity("", "")
	if score3 != 0.0 {
		t.Errorf("expected 0.0 for both empty, got %f", score3)
	}
}

func TestJaroWinkler_CaseInsensitive(t *testing.T) {
	score := jaroWinklerSimilarity("SMITH", "smith")
	if score != 1.0 {
		t.Errorf("expected 1.0 for case-insensitive match, got %f", score)
	}
}

// =========== Extract Tests ===========

func TestExtractMatchInput_FullPatient(t *testing.T) {
	patient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Smith",
				"given":  []interface{}{"John"},
			},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hospital.org/mrn",
				"value":  "MRN12345",
			},
			map[string]interface{}{
				"system": "http://hl7.org/fhir/sid/us-ssn",
				"value":  "123-45-6789",
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  "555-123-4567",
			},
			map[string]interface{}{
				"system": "email",
				"value":  "john@example.com",
			},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"123 Main St"},
				"city":       "Springfield",
				"postalCode": "62701",
			},
		},
	}

	input := extractMatchInput(patient)

	if input.lastName != "Smith" {
		t.Errorf("expected lastName Smith, got %s", input.lastName)
	}
	if input.firstName != "John" {
		t.Errorf("expected firstName John, got %s", input.firstName)
	}
	if input.birthDate != "1990-01-15" {
		t.Errorf("expected birthDate 1990-01-15, got %s", input.birthDate)
	}
	if input.gender != "male" {
		t.Errorf("expected gender male, got %s", input.gender)
	}
	if input.mrn != "MRN12345" {
		t.Errorf("expected mrn MRN12345, got %s", input.mrn)
	}
	if input.ssnLast4 != "6789" {
		t.Errorf("expected ssnLast4 6789, got %s", input.ssnLast4)
	}
	if input.phone != "555-123-4567" {
		t.Errorf("expected phone 555-123-4567, got %s", input.phone)
	}
	if input.email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", input.email)
	}
	if input.addressLine != "123 Main St" {
		t.Errorf("expected addressLine 123 Main St, got %s", input.addressLine)
	}
	if input.city != "Springfield" {
		t.Errorf("expected city Springfield, got %s", input.city)
	}
	if input.postalCode != "62701" {
		t.Errorf("expected postalCode 62701, got %s", input.postalCode)
	}
}

func TestExtractMatchInput_MinimalPatient(t *testing.T) {
	patient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{
				"family": "Doe",
			},
		},
	}

	input := extractMatchInput(patient)

	if input.lastName != "Doe" {
		t.Errorf("expected lastName Doe, got %s", input.lastName)
	}
	if input.firstName != "" {
		t.Errorf("expected empty firstName, got %s", input.firstName)
	}
	if input.birthDate != "" {
		t.Errorf("expected empty birthDate, got %s", input.birthDate)
	}
}

func TestExtractMatchInput_MRNIdentifier(t *testing.T) {
	patient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Test"},
		},
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://example.org/fhir/mrn",
				"value":  "MRN-999",
			},
		},
	}

	input := extractMatchInput(patient)

	if input.mrn != "MRN-999" {
		t.Errorf("expected mrn MRN-999, got %s", input.mrn)
	}
}

// =========== PatientMatcher Tests ===========

func TestPatientMatcher_ExactMatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID:          "p1",
			FirstName:   "John",
			LastName:    "Smith",
			BirthDate:   "1990-01-15",
			Gender:      "male",
			MRN:         "MRN123",
			Phone:       "555-123-4567",
			Email:       "john@example.com",
			AddressLine: "123 Main St",
			City:        "Springfield",
			PostalCode:  "62701",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p1",
			},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN123"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-123-4567"},
			map[string]interface{}{"system": "email", "value": "john@example.com"},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"123 Main St"},
				"city":       "Springfield",
				"postalCode": "62701",
			},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Score < 0.95 {
		t.Errorf("expected score >= 0.95 for exact match, got %f", results[0].Score)
	}
	if results[0].Grade != "certain" {
		t.Errorf("expected grade certain, got %s", results[0].Grade)
	}
}

func TestPatientMatcher_FuzzyNameMatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID:        "p1",
			FirstName: "John",
			LastName:  "Smith",
			BirthDate: "1990-01-15",
			Gender:    "male",
			MRN:       "MRN123",
			Phone:     "555-1234",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p1",
			},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"Jon"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN123"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-1234"},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for fuzzy name match")
	}
	// Fuzzy match on first name (Jon vs John) means score should be somewhat high
	// but not perfect on the first name component.
	if results[0].Score >= 1.0 {
		t.Errorf("expected score < 1.0 for fuzzy name match, got %f", results[0].Score)
	}
	if results[0].Score < 0.60 {
		t.Errorf("expected score >= 0.60 for fuzzy name match, got %f", results[0].Score)
	}
}

func TestPatientMatcher_MRNMatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID:        "p1",
			FirstName: "Jane",
			LastName:  "Doe",
			BirthDate: "1985-06-20",
			Gender:    "female",
			MRN:       "MRN-999",
			Phone:     "555-9876",
			Email:     "jane@example.com",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p1",
			},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	// Input has matching MRN, DOB, gender, phone, and email but different name.
	// This tests that MRN match significantly contributes to overall score.
	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Different", "given": []interface{}{"Name"}},
		},
		"birthDate": "1985-06-20",
		"gender":    "female",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-999"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-9876"},
			map[string]interface{}{"system": "email", "value": "jane@example.com"},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for MRN match")
	}
	// MRN(0.20) + DOB(0.20) + Gender(0.05) + Phone(0.10) + Email(0.05) = 0.60+
	// Names are different so minimal contribution there.
	if results[0].Score < 0.60 {
		t.Errorf("expected score >= 0.60 from MRN + DOB + gender + phone + email, got %f", results[0].Score)
	}
}

func TestPatientMatcher_NoMatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID:        "p1",
			FirstName: "Alice",
			LastName:  "Wonderland",
			BirthDate: "2000-12-25",
			Gender:    "female",
			MRN:       "MRN-000",
			Phone:     "999-999-9999",
			Email:     "alice@example.com",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p1",
			},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Zzzzz", "given": []interface{}{"Xxxxx"}},
		},
		"birthDate": "1950-01-01",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-DIFFERENT"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "111-111-1111"},
			map[string]interface{}{"system": "email", "value": "zzz@nowhere.com"},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All results with score < 0.60 should be filtered out.
	if len(results) != 0 {
		t.Errorf("expected no results for complete mismatch, got %d with score %f", len(results), results[0].Score)
	}
}

func TestPatientMatcher_MultipleResults(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID:        "p1",
			FirstName: "John",
			LastName:  "Smith",
			BirthDate: "1990-01-15",
			Gender:    "male",
			Phone:     "555-1111",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p1",
			},
		},
		{
			ID:        "p2",
			FirstName: "John",
			LastName:  "Smith",
			BirthDate: "1990-01-15",
			Gender:    "male",
			MRN:       "MRN123",
			Phone:     "555-1111",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p2",
			},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN123"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-1111"},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	// Results should be sorted by score descending.
	if results[0].Score < results[1].Score {
		t.Errorf("expected results sorted by score descending, got %f < %f", results[0].Score, results[1].Score)
	}
	// The one with MRN match should score higher.
	if results[0].Patient.ID != "p2" {
		t.Errorf("expected p2 (with MRN) to be first, got %s", results[0].Patient.ID)
	}
}

func TestPatientMatcher_CountLimit(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		},
		{
			ID: "p2", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p2"},
		},
		{
			ID: "p3", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p3"},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
	}

	results, err := matcher.Match(context.Background(), inputPatient, 1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) > 1 {
		t.Errorf("expected at most 1 result with count=1, got %d", len(results))
	}
}

func TestPatientMatcher_OnlyCertainMatches(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			MRN: "MRN123", Phone: "555-1234", Email: "john@test.com",
			AddressLine: "123 Main", PostalCode: "62701", SSNLast4: "6789",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		},
		{
			ID: "p2", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p2"},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN123"},
			map[string]interface{}{"system": "http://hl7.org/fhir/sid/us-ssn", "value": "xxx-xx-6789"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-1234"},
			map[string]interface{}{"system": "email", "value": "john@test.com"},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{"123 Main"},
				"postalCode": "62701",
			},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Grade != "certain" {
			t.Errorf("expected only certain matches, got grade %s with score %f", r.Grade, r.Score)
		}
	}
}

func TestPatientMatcher_BirthDateMismatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1985-06-20",
		"gender":    "male",
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With matching name (0.15+0.15) + gender (0.05) but different DOB (0.0),
	// score should be lower than with matching DOB.
	for _, r := range results {
		if r.Score >= 0.80 {
			t.Errorf("expected score < 0.80 with DOB mismatch, got %f", r.Score)
		}
	}
}

func TestPatientMatcher_PhoneMatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15",
			Gender: "male", Phone: "555-123-4567", MRN: "MRN-PH1",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputWithPhone := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-PH1"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "phone", "value": "555-123-4567"},
		},
	}

	inputWithoutPhone := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-PH1"},
		},
	}

	resultsWith, err := matcher.Match(context.Background(), inputWithPhone, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resultsWithout, err := matcher.Match(context.Background(), inputWithoutPhone, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resultsWith) == 0 || len(resultsWithout) == 0 {
		t.Fatal("expected results for both queries")
	}

	if resultsWith[0].Score <= resultsWithout[0].Score {
		t.Errorf("expected phone match to increase score: with=%f without=%f",
			resultsWith[0].Score, resultsWithout[0].Score)
	}
}

func TestPatientMatcher_EmailMatch(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15",
			Gender: "male", Email: "john@example.com", MRN: "MRN-EM1",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)

	inputWithEmail := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-EM1"},
		},
		"telecom": []interface{}{
			map[string]interface{}{"system": "email", "value": "john@example.com"},
		},
	}

	inputWithoutEmail := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN-EM1"},
		},
	}

	resultsWith, err := matcher.Match(context.Background(), inputWithEmail, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resultsWithout, err := matcher.Match(context.Background(), inputWithoutEmail, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resultsWith) == 0 || len(resultsWithout) == 0 {
		t.Fatal("expected results for both queries")
	}

	if resultsWith[0].Score <= resultsWithout[0].Score {
		t.Errorf("expected email match to increase score: with=%f without=%f",
			resultsWith[0].Score, resultsWithout[0].Score)
	}
}

func TestPatientMatcher_NilInput(t *testing.T) {
	searcher := &mockSearcher{patients: nil}
	matcher := NewPatientMatcher(searcher)

	_, err := matcher.Match(context.Background(), nil, 5, false)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestPatientMatcher_EmptySearchResults(t *testing.T) {
	searcher := &mockSearcher{patients: []PatientRecord{}}
	matcher := NewPatientMatcher(searcher)

	inputPatient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
	}

	results, err := matcher.Match(context.Background(), inputPatient, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty search, got %d", len(results))
	}
}

// =========== MatchHandler Tests ===========

func buildParametersBody(t *testing.T, patient map[string]interface{}, count *int, onlyCertain *bool) string {
	t.Helper()
	params := map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    []interface{}{},
	}

	paramList := []interface{}{}

	if patient != nil {
		paramList = append(paramList, map[string]interface{}{
			"name":     "resource",
			"resource": patient,
		})
	}

	if count != nil {
		paramList = append(paramList, map[string]interface{}{
			"name":         "count",
			"valueInteger": float64(*count),
		})
	}

	if onlyCertain != nil {
		paramList = append(paramList, map[string]interface{}{
			"name":         "onlyCertainMatches",
			"valueBoolean": *onlyCertain,
		})
	}

	params["parameter"] = paramList

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshal parameters: %v", err)
	}
	return string(data)
}

func TestMatchHandler_ValidRequest(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			MRN: "MRN123",
			FHIRResource: map[string]interface{}{
				"resourceType": "Patient",
				"id":           "p1",
				"name": []interface{}{
					map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
				},
			},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()

	patient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
		"identifier": []interface{}{
			map[string]interface{}{"system": "http://hospital.org/mrn", "value": "MRN123"},
		},
	}

	body := buildParametersBody(t, patient, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$match", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleMatch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected type searchset, got %v", bundle["type"])
	}

	entries, ok := bundle["entry"].([]interface{})
	if !ok || len(entries) == 0 {
		t.Fatal("expected at least one entry in Bundle")
	}

	entry := entries[0].(map[string]interface{})
	search, ok := entry["search"].(map[string]interface{})
	if !ok {
		t.Fatal("expected search element in entry")
	}
	if search["mode"] != "match" {
		t.Errorf("expected search mode match, got %v", search["mode"])
	}

	score, ok := search["score"].(float64)
	if !ok || score <= 0 {
		t.Errorf("expected positive score, got %v", search["score"])
	}

	// Check for match-grade extension.
	extensions, ok := search["extension"].([]interface{})
	if !ok || len(extensions) == 0 {
		t.Fatal("expected extension in search")
	}
	ext := extensions[0].(map[string]interface{})
	if ext["url"] != "http://hl7.org/fhir/StructureDefinition/match-grade" {
		t.Errorf("expected match-grade extension URL, got %v", ext["url"])
	}
}

func TestMatchHandler_MissingPatient(t *testing.T) {
	searcher := &mockSearcher{patients: nil}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()

	// Parameters without a resource parameter.
	body := `{"resourceType":"Parameters","parameter":[{"name":"count","valueInteger":5}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$match", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleMatch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMatchHandler_InvalidJSON(t *testing.T) {
	searcher := &mockSearcher{patients: nil}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$match", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleMatch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMatchHandler_EmptyBody(t *testing.T) {
	searcher := &mockSearcher{patients: nil}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$match", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleMatch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMatchHandler_EmptyResult(t *testing.T) {
	searcher := &mockSearcher{patients: []PatientRecord{}}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()

	patient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Nonexistent"},
		},
	}

	body := buildParametersBody(t, patient, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$match", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleMatch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected Bundle, got %v", bundle["resourceType"])
	}

	total, ok := bundle["total"].(float64)
	if !ok || total != 0 {
		t.Errorf("expected total 0, got %v", bundle["total"])
	}
}

func TestMatchHandler_CustomCount(t *testing.T) {
	candidates := []PatientRecord{
		{
			ID: "p1", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		},
		{
			ID: "p2", FirstName: "John", LastName: "Smith", BirthDate: "1990-01-15", Gender: "male",
			FHIRResource: map[string]interface{}{"resourceType": "Patient", "id": "p2"},
		},
	}

	searcher := &mockSearcher{patients: candidates}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()

	patient := map[string]interface{}{
		"resourceType": "Patient",
		"name": []interface{}{
			map[string]interface{}{"family": "Smith", "given": []interface{}{"John"}},
		},
		"birthDate": "1990-01-15",
		"gender":    "male",
	}

	count := 1
	body := buildParametersBody(t, patient, &count, nil)
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient/$match", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HandleMatch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	entries, ok := bundle["entry"].([]interface{})
	if ok && len(entries) > 1 {
		t.Errorf("expected at most 1 entry with count=1, got %d", len(entries))
	}
}

func TestMatchHandler_RegisterRoutes(t *testing.T) {
	searcher := &mockSearcher{patients: nil}
	matcher := NewPatientMatcher(searcher)
	handler := NewMatchHandler(matcher)

	e := echo.New()
	g := e.Group("/fhir")
	handler.RegisterRoutes(g)

	routes := e.Routes()
	found := false
	for _, r := range routes {
		if r.Method == "POST" && r.Path == "/fhir/Patient/$match" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected POST /fhir/Patient/$match route to be registered")
	}
}
