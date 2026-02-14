package fhir

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewSearchBundle(t *testing.T) {
	resources := []interface{}{
		map[string]string{"id": "1", "resourceType": "Patient"},
		map[string]string{"id": "2", "resourceType": "Patient"},
	}

	bundle := NewSearchBundle(resources, 10, "/fhir/Patient")

	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %s", bundle.ResourceType)
	}
	if bundle.Type != "searchset" {
		t.Errorf("expected type searchset, got %s", bundle.Type)
	}
	if *bundle.Total != 10 {
		t.Errorf("expected total 10, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(bundle.Entry))
	}
	if bundle.Entry[0].Search == nil || bundle.Entry[0].Search.Mode != "match" {
		t.Error("expected search mode 'match'")
	}
	if bundle.Timestamp == nil {
		t.Error("expected timestamp to be set")
	}
	// Self link should be present
	if len(bundle.Link) < 1 {
		t.Fatal("expected at least 1 link (self)")
	}
	if bundle.Link[0].Relation != "self" {
		t.Errorf("expected first link relation 'self', got %q", bundle.Link[0].Relation)
	}
}

func TestNewSearchBundle_FullURL(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{"resourceType": "Patient", "id": "abc-123"},
	}

	bundle := NewSearchBundle(resources, 1, "/fhir/Patient")

	if len(bundle.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(bundle.Entry))
	}
	if bundle.Entry[0].FullURL != "Patient/abc-123" {
		t.Errorf("expected fullUrl 'Patient/abc-123', got '%s'", bundle.Entry[0].FullURL)
	}
}

func TestNewSearchBundle_Empty(t *testing.T) {
	bundle := NewSearchBundle(nil, 0, "/fhir/Patient")

	if *bundle.Total != 0 {
		t.Errorf("expected total 0, got %d", *bundle.Total)
	}
	if len(bundle.Entry) != 0 {
		t.Errorf("expected 0 entries, got %d", len(bundle.Entry))
	}
}

func TestNewSearchBundle_ResourceSerialization(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Patient",
			"id":           "test-1",
			"active":       true,
		},
	}

	bundle := NewSearchBundle(resources, 1, "/fhir/Patient")

	if len(bundle.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(bundle.Entry))
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bundle.Entry[0].Resource, &parsed); err != nil {
		t.Fatalf("failed to parse resource JSON: %v", err)
	}
	if parsed["resourceType"] != "Patient" {
		t.Errorf("expected resourceType Patient, got %v", parsed["resourceType"])
	}
	if parsed["id"] != "test-1" {
		t.Errorf("expected id test-1, got %v", parsed["id"])
	}
}

func TestNewSearchBundleWithLinks_FirstPage(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{"id": "1", "resourceType": "Patient"},
		map[string]interface{}{"id": "2", "resourceType": "Patient"},
	}

	params := SearchBundleParams{
		BaseURL:  "/fhir/Patient",
		QueryStr: "name=Smith",
		Count:    10,
		Offset:   0,
		Total:    42,
	}

	bundle := NewSearchBundleWithLinks(resources, params)

	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %s", bundle.ResourceType)
	}
	if bundle.Type != "searchset" {
		t.Errorf("expected type searchset, got %s", bundle.Type)
	}
	if *bundle.Total != 42 {
		t.Errorf("expected total 42, got %d", *bundle.Total)
	}

	// Should have self and next links (offset=0, total=42, count=10)
	if len(bundle.Link) < 2 {
		t.Fatalf("expected at least 2 links (self, next), got %d", len(bundle.Link))
	}

	selfLink := bundle.Link[0]
	if selfLink.Relation != "self" {
		t.Errorf("expected first link to be 'self', got '%s'", selfLink.Relation)
	}
	if selfLink.URL != "/fhir/Patient?name=Smith&_count=10&_offset=0" {
		t.Errorf("unexpected self URL: %s", selfLink.URL)
	}

	nextLink := bundle.Link[1]
	if nextLink.Relation != "next" {
		t.Errorf("expected second link to be 'next', got '%s'", nextLink.Relation)
	}
	if nextLink.URL != "/fhir/Patient?name=Smith&_count=10&_offset=10" {
		t.Errorf("unexpected next URL: %s", nextLink.URL)
	}
}

func TestNewSearchBundleWithLinks_MiddlePage(t *testing.T) {
	params := SearchBundleParams{
		BaseURL:  "/fhir/Patient",
		QueryStr: "name=Smith",
		Count:    10,
		Offset:   20,
		Total:    42,
	}

	bundle := NewSearchBundleWithLinks(nil, params)

	// Should have self, next, and previous links
	if len(bundle.Link) != 3 {
		t.Fatalf("expected 3 links (self, next, previous), got %d", len(bundle.Link))
	}

	relations := map[string]string{}
	for _, l := range bundle.Link {
		relations[l.Relation] = l.URL
	}

	if _, ok := relations["self"]; !ok {
		t.Error("missing self link")
	}
	if _, ok := relations["next"]; !ok {
		t.Error("missing next link")
	}
	if _, ok := relations["previous"]; !ok {
		t.Error("missing previous link")
	}
}

func TestNewSearchBundleWithLinks_LastPage(t *testing.T) {
	params := SearchBundleParams{
		BaseURL:  "/fhir/Patient",
		QueryStr: "",
		Count:    10,
		Offset:   40,
		Total:    42,
	}

	bundle := NewSearchBundleWithLinks(nil, params)

	// Should have self and previous links, but NOT next
	relations := map[string]bool{}
	for _, l := range bundle.Link {
		relations[l.Relation] = true
	}

	if !relations["self"] {
		t.Error("missing self link")
	}
	if relations["next"] {
		t.Error("should not have next link on last page")
	}
	if !relations["previous"] {
		t.Error("missing previous link")
	}
}

func TestNewSearchBundleWithLinks_EmptyQuery(t *testing.T) {
	params := SearchBundleParams{
		BaseURL:  "/fhir/Patient",
		QueryStr: "",
		Count:    10,
		Offset:   0,
		Total:    5,
	}

	bundle := NewSearchBundleWithLinks(nil, params)

	// Only self, no next or previous
	if len(bundle.Link) != 1 {
		t.Fatalf("expected 1 link (self only), got %d", len(bundle.Link))
	}
	if bundle.Link[0].URL != "/fhir/Patient?_count=10&_offset=0" {
		t.Errorf("unexpected self URL: %s", bundle.Link[0].URL)
	}
}

func TestNewTransactionResponse(t *testing.T) {
	entries := []BundleEntry{
		{
			Response: &BundleResponse{
				Status:   "201 Created",
				Location: "Patient/123",
			},
		},
		{
			Response: &BundleResponse{
				Status:   "200 OK",
				Location: "Observation/456",
			},
		},
	}

	bundle := NewTransactionResponse(entries)

	if bundle.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %s", bundle.ResourceType)
	}
	if bundle.Type != "transaction-response" {
		t.Errorf("expected type transaction-response, got %s", bundle.Type)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(bundle.Entry))
	}
	if bundle.Timestamp == nil {
		t.Error("expected timestamp to be set")
	}
	if bundle.Entry[0].Response.Status != "201 Created" {
		t.Errorf("expected status '201 Created', got '%s'", bundle.Entry[0].Response.Status)
	}
}

func TestNewBatchResponse(t *testing.T) {
	entries := []BundleEntry{
		{
			Response: &BundleResponse{
				Status: "201 Created",
			},
		},
		{
			Response: &BundleResponse{
				Status: "400 Bad Request",
			},
		},
	}

	bundle := NewBatchResponse(entries)

	if bundle.Type != "batch-response" {
		t.Errorf("expected type batch-response, got %s", bundle.Type)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(bundle.Entry))
	}
}

func TestExtractFullURL(t *testing.T) {
	tests := []struct {
		name     string
		resource interface{}
		baseURL  string
		want     string
	}{
		{
			name:     "map with resourceType and id",
			resource: map[string]interface{}{"resourceType": "Patient", "id": "123"},
			baseURL:  "/fhir/Patient",
			want:     "Patient/123",
		},
		{
			name:     "map missing id",
			resource: map[string]interface{}{"resourceType": "Patient"},
			baseURL:  "/fhir/Patient",
			want:     "",
		},
		{
			name:     "map[string]string type",
			resource: map[string]string{"resourceType": "Observation", "id": "obs-1"},
			baseURL:  "/fhir/Observation",
			want:     "Observation/obs-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFullURL(tt.resource, tt.baseURL)
			if got != tt.want {
				t.Errorf("extractFullURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPaginationLinks(t *testing.T) {
	tests := []struct {
		name          string
		params        SearchBundleParams
		expectSelf    bool
		expectNext    bool
		expectPrev    bool
		expectedCount int
	}{
		{
			name: "first page with more results",
			params: SearchBundleParams{
				BaseURL: "/fhir/Patient", QueryStr: "name=Smith",
				Count: 10, Offset: 0, Total: 50,
			},
			expectSelf: true, expectNext: true, expectPrev: false,
			expectedCount: 2,
		},
		{
			name: "middle page",
			params: SearchBundleParams{
				BaseURL: "/fhir/Patient", QueryStr: "name=Smith",
				Count: 10, Offset: 20, Total: 50,
			},
			expectSelf: true, expectNext: true, expectPrev: true,
			expectedCount: 3,
		},
		{
			name: "last page",
			params: SearchBundleParams{
				BaseURL: "/fhir/Patient", QueryStr: "name=Smith",
				Count: 10, Offset: 40, Total: 50,
			},
			expectSelf: true, expectNext: false, expectPrev: true,
			expectedCount: 2,
		},
		{
			name: "single page",
			params: SearchBundleParams{
				BaseURL: "/fhir/Patient", QueryStr: "",
				Count: 10, Offset: 0, Total: 5,
			},
			expectSelf: true, expectNext: false, expectPrev: false,
			expectedCount: 1,
		},
		{
			name: "no results",
			params: SearchBundleParams{
				BaseURL: "/fhir/Patient",
				Count: 10, Offset: 0, Total: 0,
			},
			expectSelf: true, expectNext: false, expectPrev: false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links := buildPaginationLinks(tt.params)
			if len(links) != tt.expectedCount {
				t.Errorf("expected %d links, got %d", tt.expectedCount, len(links))
			}
			hasRelation := func(rel string) bool {
				for _, l := range links {
					if l.Relation == rel {
						return true
					}
				}
				return false
			}
			if tt.expectSelf && !hasRelation("self") {
				t.Error("expected self link")
			}
			if tt.expectNext && !hasRelation("next") {
				t.Error("expected next link")
			}
			if tt.expectPrev && !hasRelation("previous") {
				t.Error("expected previous link")
			}
		})
	}
}

func TestConditionalAmpersand(t *testing.T) {
	if got := conditionalAmpersand(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if got := conditionalAmpersand("foo=bar"); got != "foo=bar&" {
		t.Errorf("expected 'foo=bar&', got %q", got)
	}
}

func TestParseEntryRequest(t *testing.T) {
	tests := []struct {
		name       string
		entry      BundleEntry
		wantMethod string
		wantType   string
		wantID     string
	}{
		{
			name: "POST Patient",
			entry: BundleEntry{
				Request: &BundleRequest{Method: "POST", URL: "Patient"},
			},
			wantMethod: "POST", wantType: "Patient", wantID: "",
		},
		{
			name: "PUT Patient/123",
			entry: BundleEntry{
				Request: &BundleRequest{Method: "PUT", URL: "Patient/123"},
			},
			wantMethod: "PUT", wantType: "Patient", wantID: "123",
		},
		{
			name: "DELETE Patient/123",
			entry: BundleEntry{
				Request: &BundleRequest{Method: "DELETE", URL: "Patient/123"},
			},
			wantMethod: "DELETE", wantType: "Patient", wantID: "123",
		},
		{
			name: "GET with query params",
			entry: BundleEntry{
				Request: &BundleRequest{Method: "GET", URL: "Patient?name=Smith"},
			},
			wantMethod: "GET", wantType: "Patient", wantID: "",
		},
		{
			name: "leading slash",
			entry: BundleEntry{
				Request: &BundleRequest{Method: "PUT", URL: "/Observation/obs-1"},
			},
			wantMethod: "PUT", wantType: "Observation", wantID: "obs-1",
		},
		{
			name:       "nil request",
			entry:      BundleEntry{},
			wantMethod: "", wantType: "", wantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, resType, resID := parseEntryRequest(tt.entry)
			if method != tt.wantMethod {
				t.Errorf("method = %q, want %q", method, tt.wantMethod)
			}
			if resType != tt.wantType {
				t.Errorf("resourceType = %q, want %q", resType, tt.wantType)
			}
			if resID != tt.wantID {
				t.Errorf("resourceID = %q, want %q", resID, tt.wantID)
			}
		})
	}
}

func TestBundleJSON_RoundTrip(t *testing.T) {
	resources := []interface{}{
		map[string]interface{}{
			"resourceType": "Patient",
			"id":           "p1",
			"active":       true,
		},
	}

	bundle := NewSearchBundle(resources, 1, "/fhir/Patient")

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("failed to marshal bundle: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal bundle: %v", err)
	}

	if parsed["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle in JSON")
	}
	if parsed["type"] != "searchset" {
		t.Errorf("expected type searchset in JSON")
	}

	total, ok := parsed["total"].(float64)
	if !ok || int(total) != 1 {
		t.Errorf("expected total 1, got %v", parsed["total"])
	}

	entries, ok := parsed["entry"].([]interface{})
	if !ok || len(entries) != 1 {
		t.Fatal("expected 1 entry in JSON")
	}

	entry := entries[0].(map[string]interface{})
	resource := entry["resource"].(map[string]interface{})
	if resource["resourceType"] != "Patient" {
		t.Errorf("expected Patient resource in entry")
	}
}

func TestDefaultBundleProcessor(t *testing.T) {
	proc := &DefaultBundleProcessor{}

	tests := []struct {
		method     string
		resType    string
		resID      string
		wantStatus string
		wantErr    bool
	}{
		{"POST", "Patient", "123", "201 Created", false},
		{"PUT", "Patient", "123", "200 OK", false},
		{"DELETE", "Patient", "123", "204 No Content", false},
		{"GET", "Patient", "123", "200 OK", false},
		{"PATCH", "Patient", "123", "", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.resType), func(t *testing.T) {
			entry, err := proc.ProcessEntry(nil, tt.method, tt.resType, tt.resID, nil)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.Response.Status != tt.wantStatus {
				t.Errorf("expected status %q, got %q", tt.wantStatus, entry.Response.Status)
			}
		})
	}
}

func TestNewCapabilityStatement(t *testing.T) {
	resources := []CSResource{
		ResourceCapability("Patient", []CSSearchParam{
			{Name: "name", Type: "string"},
		}),
	}

	cs := NewCapabilityStatement("http://localhost:8000/fhir", resources)

	if cs.ResourceType != "CapabilityStatement" {
		t.Errorf("expected CapabilityStatement, got %s", cs.ResourceType)
	}
	if cs.FHIRVersion != "4.0.1" {
		t.Errorf("expected FHIR version 4.0.1, got %s", cs.FHIRVersion)
	}
	if cs.Kind != "instance" {
		t.Errorf("expected kind instance, got %s", cs.Kind)
	}
	if len(cs.Rest) != 1 {
		t.Fatalf("expected 1 rest entry, got %d", len(cs.Rest))
	}
	if cs.Rest[0].Mode != "server" {
		t.Errorf("expected mode server, got %s", cs.Rest[0].Mode)
	}
	if len(cs.Rest[0].Resource) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(cs.Rest[0].Resource))
	}
	if cs.Rest[0].Resource[0].Type != "Patient" {
		t.Errorf("expected Patient resource, got %s", cs.Rest[0].Resource[0].Type)
	}
}

func TestResourceCapability(t *testing.T) {
	rc := ResourceCapability("Encounter", []CSSearchParam{
		{Name: "patient", Type: "reference"},
		{Name: "status", Type: "token"},
	})

	if rc.Type != "Encounter" {
		t.Errorf("expected Encounter, got %s", rc.Type)
	}
	if len(rc.Interaction) != 6 {
		t.Errorf("expected 6 interactions (CRUD+search+vread), got %d", len(rc.Interaction))
	}
	if len(rc.SearchParam) != 2 {
		t.Errorf("expected 2 search params, got %d", len(rc.SearchParam))
	}
	if rc.Versioning != "versioned" {
		t.Errorf("expected versioned, got %s", rc.Versioning)
	}
}
