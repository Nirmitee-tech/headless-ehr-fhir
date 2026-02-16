package fhir

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		sortValue string
		id        string
	}{
		{"simple values", "2024-01-15", "patient-123"},
		{"empty sort value", "", "abc"},
		{"empty id", "2024-01-15", ""},
		{"both empty", "", ""},
		{"special characters", "Smith, John", "urn:uuid:550e8400-e29b-41d4-a716-446655440000"},
		{"unicode values", "Muller", "id-42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := EncodeCursor(tt.sortValue, tt.id)
			if token == "" {
				t.Fatal("expected non-empty token")
			}

			cursor, err := DecodeCursor(token)
			if err != nil {
				t.Fatalf("DecodeCursor returned error: %v", err)
			}

			if cursor.Value != tt.sortValue {
				t.Errorf("sort value = %q, want %q", cursor.Value, tt.sortValue)
			}
			if cursor.ID != tt.id {
				t.Errorf("id = %q, want %q", cursor.ID, tt.id)
			}
		})
	}
}

func TestEncodeCursor_Format(t *testing.T) {
	token := EncodeCursor("2024-01-15", "patient-123")

	// Verify the token is valid base64url
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token is not valid base64url: %v", err)
	}

	// Verify the decoded data is valid JSON with expected fields
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("decoded token is not valid JSON: %v", err)
	}

	if parsed["v"] != "2024-01-15" {
		t.Errorf("expected v=2024-01-15, got %v", parsed["v"])
	}
	if parsed["id"] != "patient-123" {
		t.Errorf("expected id=patient-123, got %v", parsed["id"])
	}
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	_, err := DecodeCursor("!!!not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
	if !strings.Contains(err.Error(), "invalid cursor token") {
		t.Errorf("expected 'invalid cursor token' in error, got: %v", err)
	}
}

func TestDecodeCursor_InvalidJSON(t *testing.T) {
	// Valid base64 but not valid JSON
	token := base64.RawURLEncoding.EncodeToString([]byte("not json"))
	_, err := DecodeCursor(token)
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
	if !strings.Contains(err.Error(), "invalid cursor payload") {
		t.Errorf("expected 'invalid cursor payload' in error, got: %v", err)
	}
}

func TestDecodeCursor_EmptyToken(t *testing.T) {
	// Empty string decodes to empty bytes, which is not valid JSON
	_, err := DecodeCursor("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNewSearchBundleWithCursor_WithNextPage(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{"resourceType": "Patient", "id": "p1"},
		map[string]interface{}{"resourceType": "Patient", "id": "p2"},
	}

	nextToken := EncodeCursor("2024-06-01", "p2")

	params := CursorBundleParams{
		BaseURL:    "/fhir/Patient",
		QueryStr:   "name=Smith",
		Count:      10,
		Total:      42,
		HasMore:    true,
		NextCursor: nextToken,
	}

	bundle := NewSearchBundleWithCursor(resources, params)

	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %s", bundle.ResourceType)
	}
	if bundle.Type != "searchset" {
		t.Errorf("expected type searchset, got %s", bundle.Type)
	}
	if *bundle.Total != 42 {
		t.Errorf("expected total 42, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(bundle.Entry))
	}
	if bundle.Timestamp == nil {
		t.Error("expected timestamp to be set")
	}

	// Should have self and next links
	if len(bundle.Link) != 2 {
		t.Fatalf("expected 2 links (self, next), got %d", len(bundle.Link))
	}

	selfLink := bundle.Link[0]
	if selfLink.Relation != "self" {
		t.Errorf("expected first link relation 'self', got %q", selfLink.Relation)
	}
	if selfLink.URL != "/fhir/Patient?name=Smith&_count=10" {
		t.Errorf("unexpected self URL: %s", selfLink.URL)
	}

	nextLink := bundle.Link[1]
	if nextLink.Relation != "next" {
		t.Errorf("expected second link relation 'next', got %q", nextLink.Relation)
	}
	expectedNextURL := "/fhir/Patient?name=Smith&_count=10&_pageToken=" + nextToken
	if nextLink.URL != expectedNextURL {
		t.Errorf("unexpected next URL:\n  got:  %s\n  want: %s", nextLink.URL, expectedNextURL)
	}
}

func TestNewSearchBundleWithCursor_NoMoreResults(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{"resourceType": "Patient", "id": "p1"},
	}

	params := CursorBundleParams{
		BaseURL:    "/fhir/Patient",
		QueryStr:   "name=Smith",
		Count:      10,
		Total:      1,
		HasMore:    false,
		NextCursor: "",
	}

	bundle := NewSearchBundleWithCursor(resources, params)

	// Should only have self link, no next
	if len(bundle.Link) != 1 {
		t.Fatalf("expected 1 link (self only), got %d", len(bundle.Link))
	}
	if bundle.Link[0].Relation != "self" {
		t.Errorf("expected self link, got %q", bundle.Link[0].Relation)
	}
}

func TestNewSearchBundleWithCursor_HasMoreButEmptyCursor(t *testing.T) {
	params := CursorBundleParams{
		BaseURL:    "/fhir/Patient",
		QueryStr:   "",
		Count:      10,
		Total:      20,
		HasMore:    true,
		NextCursor: "", // HasMore is true but cursor is empty
	}

	bundle := NewSearchBundleWithCursor(nil, params)

	// Should only have self link since NextCursor is empty
	if len(bundle.Link) != 1 {
		t.Fatalf("expected 1 link (self only), got %d", len(bundle.Link))
	}
}

func TestNewSearchBundleWithCursor_EmptyQuery(t *testing.T) {
	params := CursorBundleParams{
		BaseURL:    "/fhir/Patient",
		QueryStr:   "",
		Count:      10,
		Total:      5,
		HasMore:    false,
		NextCursor: "",
	}

	bundle := NewSearchBundleWithCursor(nil, params)

	if len(bundle.Link) != 1 {
		t.Fatalf("expected 1 link (self), got %d", len(bundle.Link))
	}
	if bundle.Link[0].URL != "/fhir/Patient?_count=10" {
		t.Errorf("unexpected self URL: %s", bundle.Link[0].URL)
	}
}

func TestNewSearchBundleWithCursor_EntriesFullURL(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{"resourceType": "Observation", "id": "obs-1"},
		map[string]interface{}{"resourceType": "Observation", "id": "obs-2"},
	}

	params := CursorBundleParams{
		BaseURL: "/fhir/Observation",
		Count:   10,
		Total:   2,
	}

	bundle := NewSearchBundleWithCursor(resources, params)

	if len(bundle.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(bundle.Entry))
	}
	if bundle.Entry[0].FullURL != "Observation/obs-1" {
		t.Errorf("expected fullUrl 'Observation/obs-1', got %q", bundle.Entry[0].FullURL)
	}
	if bundle.Entry[1].FullURL != "Observation/obs-2" {
		t.Errorf("expected fullUrl 'Observation/obs-2', got %q", bundle.Entry[1].FullURL)
	}
}

func TestNewSearchBundleWithCursor_SearchMode(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{"resourceType": "Patient", "id": "p1"},
	}

	params := CursorBundleParams{
		BaseURL: "/fhir/Patient",
		Count:   10,
		Total:   1,
	}

	bundle := NewSearchBundleWithCursor(resources, params)

	if len(bundle.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(bundle.Entry))
	}
	if bundle.Entry[0].Search == nil {
		t.Fatal("expected search to be set")
	}
	if bundle.Entry[0].Search.Mode != "match" {
		t.Errorf("expected search mode 'match', got %q", bundle.Entry[0].Search.Mode)
	}
}

func TestParsePageToken_Present(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_pageToken=abc123&_count=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := ParsePageToken(c)
	if token != "abc123" {
		t.Errorf("expected 'abc123', got %q", token)
	}
}

func TestParsePageToken_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := ParsePageToken(c)
	if token != "" {
		t.Errorf("expected empty string, got %q", token)
	}
}

func TestParsePageToken_EmptyValue(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_pageToken=&_count=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := ParsePageToken(c)
	if token != "" {
		t.Errorf("expected empty string, got %q", token)
	}
}
