package fhir

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// ParseTransactionBundle tests
// ---------------------------------------------------------------------------

func TestParseTransactionBundle_ValidTransaction(t *testing.T) {
	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:1111",
				"resource": {"resourceType": "Patient", "name": [{"family": "Doe"}]},
				"request": {"method": "POST", "url": "Patient"}
			}
		]
	}`

	b, err := ParseTransactionBundle([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Type != "transaction" {
		t.Errorf("expected type transaction, got %s", b.Type)
	}
	if b.ResourceType != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %s", b.ResourceType)
	}
	if len(b.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(b.Entries))
	}
	if b.Entries[0].FullURL != "urn:uuid:1111" {
		t.Errorf("expected fullUrl urn:uuid:1111, got %s", b.Entries[0].FullURL)
	}
	if b.Entries[0].Request.Method != "POST" {
		t.Errorf("expected method POST, got %s", b.Entries[0].Request.Method)
	}
	if b.Entries[0].Resource["resourceType"] != "Patient" {
		t.Errorf("expected resourceType Patient in resource")
	}
}

func TestParseTransactionBundle_ValidBatch(t *testing.T) {
	body := `{
		"resourceType": "Bundle",
		"type": "batch",
		"entry": [
			{
				"resource": {"resourceType": "Observation"},
				"request": {"method": "POST", "url": "Observation"}
			},
			{
				"request": {"method": "GET", "url": "Patient/123"}
			}
		]
	}`

	b, err := ParseTransactionBundle([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Type != "batch" {
		t.Errorf("expected type batch, got %s", b.Type)
	}
	if len(b.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(b.Entries))
	}
	// Second entry has no resource (GET).
	if b.Entries[1].Resource != nil {
		t.Error("expected nil resource for GET entry")
	}
}

func TestParseTransactionBundle_InvalidJSON(t *testing.T) {
	_, err := ParseTransactionBundle([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in error, got: %v", err)
	}
}

func TestParseTransactionBundle_MissingType(t *testing.T) {
	body := `{"resourceType": "Bundle"}`
	_, err := ParseTransactionBundle([]byte(body))
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !strings.Contains(err.Error(), "bundle type is required") {
		t.Errorf("expected 'bundle type is required' in error, got: %v", err)
	}
}

func TestParseTransactionBundle_WrongResourceType(t *testing.T) {
	body := `{"resourceType": "Patient", "type": "transaction"}`
	_, err := ParseTransactionBundle([]byte(body))
	if err == nil {
		t.Fatal("expected error for wrong resourceType")
	}
	if !strings.Contains(err.Error(), "expected resourceType Bundle") {
		t.Errorf("expected 'expected resourceType Bundle' in error, got: %v", err)
	}
}

func TestParseTransactionBundle_InvalidResourceInEntry(t *testing.T) {
	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:1",
				"resource": "not-a-json-object",
				"request": {"method": "POST", "url": "Patient"}
			}
		]
	}`
	_, err := ParseTransactionBundle([]byte(body))
	if err == nil {
		t.Fatal("expected error for invalid resource")
	}
	if !strings.Contains(err.Error(), "invalid resource in entry 0") {
		t.Errorf("expected 'invalid resource in entry 0' error, got: %v", err)
	}
}

func TestParseTransactionBundle_MultipleEntries(t *testing.T) {
	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:aaa",
				"resource": {"resourceType": "Patient"},
				"request": {"method": "POST", "url": "Patient"}
			},
			{
				"fullUrl": "urn:uuid:bbb",
				"resource": {"resourceType": "Encounter", "subject": {"reference": "urn:uuid:aaa"}},
				"request": {"method": "POST", "url": "Encounter"}
			},
			{
				"fullUrl": "urn:uuid:ccc",
				"request": {"method": "DELETE", "url": "Observation/old-1"}
			}
		]
	}`
	b, err := ParseTransactionBundle([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(b.Entries))
	}
}

func TestParseTransactionBundle_ConditionalHeaders(t *testing.T) {
	body := `{
		"resourceType": "Bundle",
		"type": "batch",
		"entry": [
			{
				"resource": {"resourceType": "Patient"},
				"request": {
					"method": "PUT",
					"url": "Patient/123",
					"ifMatch": "W/\"1\"",
					"ifNoneExist": "identifier=http://example.org|12345"
				}
			}
		]
	}`
	b, err := ParseTransactionBundle([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Entries[0].Request.IfMatch != `W/"1"` {
		t.Errorf("expected ifMatch W/\"1\", got %s", b.Entries[0].Request.IfMatch)
	}
	if b.Entries[0].Request.IfNoneExist != "identifier=http://example.org|12345" {
		t.Errorf("expected ifNoneExist value, got %s", b.Entries[0].Request.IfNoneExist)
	}
}

func TestParseTransactionBundle_EmptyEntries(t *testing.T) {
	body := `{
		"resourceType": "Bundle",
		"type": "batch",
		"entry": []
	}`
	b, err := ParseTransactionBundle([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(b.Entries))
	}
}

// ---------------------------------------------------------------------------
// ValidateTransactionBundle tests
// ---------------------------------------------------------------------------

func TestValidateTransactionBundle_ValidEntries(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:1",
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %+v", len(issues), issues)
	}
}

func TestValidateTransactionBundle_InvalidBundleType(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "searchset",
	}
	issues := ValidateTransactionBundle(bundle)
	hasTypeError := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "bundle type must be") {
			hasTypeError = true
			break
		}
	}
	if !hasTypeError {
		t.Error("expected validation error for invalid bundle type")
	}
}

func TestValidateTransactionBundle_MissingRequest(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	if len(issues) == 0 {
		t.Fatal("expected validation issues for missing request fields")
	}
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "request.method is required") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about missing request.method")
	}
}

func TestValidateTransactionBundle_MissingURL(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{
				Request: BundleEntryRequest{Method: "GET"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "request.url is required") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about missing request.url")
	}
}

func TestValidateTransactionBundle_InvalidMethod(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{
				Request: BundleEntryRequest{Method: "FOOBAR", URL: "Patient/123"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "invalid HTTP method") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about invalid HTTP method")
	}
}

func TestValidateTransactionBundle_TransactionMissingFullUrl(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "fullUrl is required for transaction entries") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about missing fullUrl for transaction entry")
	}
}

func TestValidateTransactionBundle_BatchAllowsMissingFullUrl(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "fullUrl is required") {
			t.Error("batch entries should not require fullUrl")
		}
	}
}

func TestValidateTransactionBundle_DuplicateFullUrl(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL: "urn:uuid:dup",
				Request: BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
			{
				FullURL: "urn:uuid:dup",
				Request: BundleEntryRequest{Method: "POST", URL: "Observation"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "duplicate fullUrl") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about duplicate fullUrl")
	}
}

func TestValidateTransactionBundle_CircularReferences(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:a",
				Resource: map[string]interface{}{"resourceType": "Patient", "link": map[string]interface{}{"reference": "urn:uuid:b"}},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
			{
				FullURL:  "urn:uuid:b",
				Resource: map[string]interface{}{"resourceType": "Patient", "link": map[string]interface{}{"reference": "urn:uuid:a"}},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Diagnostics, "circular reference") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about circular references")
	}
}

func TestValidateTransactionBundle_AllMethodTypes(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"}
	for _, m := range methods {
		bundle := &TransactionBundle{
			ResourceType: "Bundle",
			Type:         "batch",
			Entries: []TransactionEntry{
				{
					Request: BundleEntryRequest{Method: m, URL: "Patient/123"},
				},
			},
		}
		issues := ValidateTransactionBundle(bundle)
		for _, issue := range issues {
			if strings.Contains(issue.Diagnostics, "invalid HTTP method") {
				t.Errorf("method %s should be valid", m)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// SortTransactionEntries tests
// ---------------------------------------------------------------------------

func TestSortTransactionEntries_Order(t *testing.T) {
	entries := []TransactionEntry{
		{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
		{Request: BundleEntryRequest{Method: "POST", URL: "Patient"}},
		{Request: BundleEntryRequest{Method: "PUT", URL: "Patient/2"}},
		{Request: BundleEntryRequest{Method: "DELETE", URL: "Patient/3"}},
		{Request: BundleEntryRequest{Method: "HEAD", URL: "Patient/4"}},
		{Request: BundleEntryRequest{Method: "PATCH", URL: "Patient/5"}},
	}

	sorted := SortTransactionEntries(entries)

	expected := []string{"DELETE", "POST", "PUT", "PATCH", "GET", "HEAD"}
	for i, exp := range expected {
		if sorted[i].Request.Method != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, sorted[i].Request.Method)
		}
	}
}

func TestSortTransactionEntries_StableSort(t *testing.T) {
	entries := []TransactionEntry{
		{FullURL: "a", Request: BundleEntryRequest{Method: "POST", URL: "Patient"}},
		{FullURL: "b", Request: BundleEntryRequest{Method: "POST", URL: "Observation"}},
		{FullURL: "c", Request: BundleEntryRequest{Method: "POST", URL: "Encounter"}},
	}

	sorted := SortTransactionEntries(entries)

	// Same method; order should be preserved.
	if sorted[0].FullURL != "a" || sorted[1].FullURL != "b" || sorted[2].FullURL != "c" {
		t.Error("stable sort not maintained for entries with same method")
	}
}

func TestSortTransactionEntries_EmptySlice(t *testing.T) {
	sorted := SortTransactionEntries(nil)
	if len(sorted) != 0 {
		t.Errorf("expected empty result, got %d entries", len(sorted))
	}
}

func TestSortTransactionEntries_SingleEntry(t *testing.T) {
	entries := []TransactionEntry{
		{Request: BundleEntryRequest{Method: "PUT", URL: "Patient/1"}},
	}
	sorted := SortTransactionEntries(entries)
	if len(sorted) != 1 || sorted[0].Request.Method != "PUT" {
		t.Error("single entry sort failed")
	}
}

// ---------------------------------------------------------------------------
// ParseEntryURL tests
// ---------------------------------------------------------------------------

func TestParseEntryURL_ResourceWithID(t *testing.T) {
	rt, id, isSearch := ParseEntryURL("Patient/123")
	if rt != "Patient" {
		t.Errorf("expected Patient, got %s", rt)
	}
	if id != "123" {
		t.Errorf("expected 123, got %s", id)
	}
	if isSearch {
		t.Error("expected isSearch=false")
	}
}

func TestParseEntryURL_SearchQuery(t *testing.T) {
	rt, id, isSearch := ParseEntryURL("Patient?name=Smith")
	if rt != "Patient" {
		t.Errorf("expected Patient, got %s", rt)
	}
	if id != "" {
		t.Errorf("expected empty id, got %s", id)
	}
	if !isSearch {
		t.Error("expected isSearch=true")
	}
}

func TestParseEntryURL_ResourceTypeOnly(t *testing.T) {
	rt, id, isSearch := ParseEntryURL("Patient")
	if rt != "Patient" {
		t.Errorf("expected Patient, got %s", rt)
	}
	if id != "" {
		t.Errorf("expected empty id, got %s", id)
	}
	if isSearch {
		t.Error("expected isSearch=false")
	}
}

func TestParseEntryURL_VersionedRead(t *testing.T) {
	rt, id, isSearch := ParseEntryURL("Patient/123/_history/2")
	if rt != "Patient" {
		t.Errorf("expected Patient, got %s", rt)
	}
	if id != "123" {
		t.Errorf("expected 123, got %s", id)
	}
	if isSearch {
		t.Error("expected isSearch=false")
	}
}

func TestParseEntryURL_SearchWithMultipleParams(t *testing.T) {
	rt, _, isSearch := ParseEntryURL("Observation?patient=Patient/123&code=8302-2")
	if rt != "Observation" {
		t.Errorf("expected Observation, got %s", rt)
	}
	if !isSearch {
		t.Error("expected isSearch=true")
	}
}

func TestParseEntryURL_EmptyString(t *testing.T) {
	rt, id, isSearch := ParseEntryURL("")
	if rt != "" {
		t.Errorf("expected empty resourceType, got %s", rt)
	}
	if id != "" {
		t.Errorf("expected empty id, got %s", id)
	}
	if isSearch {
		t.Error("expected isSearch=false")
	}
}

// ---------------------------------------------------------------------------
// ResolveInternalReferences tests
// ---------------------------------------------------------------------------

func TestResolveInternalReferences_ReplacesURNUUID(t *testing.T) {
	entries := []TransactionEntry{
		{
			FullURL: "urn:uuid:aaa",
			Resource: map[string]interface{}{
				"resourceType": "Encounter",
				"subject":      map[string]interface{}{"reference": "urn:uuid:bbb"},
			},
			Request: BundleEntryRequest{Method: "POST", URL: "Encounter"},
		},
	}
	idMap := map[string]string{
		"urn:uuid:bbb": "Patient/456",
	}

	ResolveInternalReferences(entries, idMap)

	subject, ok := entries[0].Resource["subject"].(map[string]interface{})
	if !ok {
		t.Fatal("expected subject to be a map")
	}
	if subject["reference"] != "Patient/456" {
		t.Errorf("expected Patient/456, got %v", subject["reference"])
	}
}

func TestResolveInternalReferences_NestedReferences(t *testing.T) {
	entries := []TransactionEntry{
		{
			FullURL: "urn:uuid:enc",
			Resource: map[string]interface{}{
				"resourceType": "Encounter",
				"participant": []interface{}{
					map[string]interface{}{
						"individual": map[string]interface{}{
							"reference": "urn:uuid:prac",
						},
					},
				},
				"subject": map[string]interface{}{
					"reference": "urn:uuid:pat",
				},
			},
			Request: BundleEntryRequest{Method: "POST", URL: "Encounter"},
		},
	}
	idMap := map[string]string{
		"urn:uuid:prac": "Practitioner/789",
		"urn:uuid:pat":  "Patient/123",
	}

	ResolveInternalReferences(entries, idMap)

	// Check nested reference in array.
	participants := entries[0].Resource["participant"].([]interface{})
	part := participants[0].(map[string]interface{})
	individual := part["individual"].(map[string]interface{})
	if individual["reference"] != "Practitioner/789" {
		t.Errorf("expected Practitioner/789, got %v", individual["reference"])
	}

	// Check direct reference.
	subject := entries[0].Resource["subject"].(map[string]interface{})
	if subject["reference"] != "Patient/123" {
		t.Errorf("expected Patient/123, got %v", subject["reference"])
	}
}

func TestResolveInternalReferences_URLResolution(t *testing.T) {
	entries := []TransactionEntry{
		{
			Request: BundleEntryRequest{Method: "PUT", URL: "urn:uuid:pat"},
		},
	}
	idMap := map[string]string{
		"urn:uuid:pat": "Patient/999",
	}

	ResolveInternalReferences(entries, idMap)

	if entries[0].Request.URL != "Patient/999" {
		t.Errorf("expected Patient/999, got %s", entries[0].Request.URL)
	}
}

func TestResolveInternalReferences_NoMatchingRefs(t *testing.T) {
	entries := []TransactionEntry{
		{
			Resource: map[string]interface{}{
				"subject": map[string]interface{}{"reference": "Patient/existing"},
			},
			Request: BundleEntryRequest{Method: "POST", URL: "Encounter"},
		},
	}
	idMap := map[string]string{
		"urn:uuid:other": "Patient/123",
	}

	ResolveInternalReferences(entries, idMap)

	subject := entries[0].Resource["subject"].(map[string]interface{})
	if subject["reference"] != "Patient/existing" {
		t.Errorf("expected unchanged reference, got %v", subject["reference"])
	}
}

func TestResolveInternalReferences_EmptyIDMap(t *testing.T) {
	entries := []TransactionEntry{
		{
			Resource: map[string]interface{}{
				"subject": map[string]interface{}{"reference": "urn:uuid:xyz"},
			},
			Request: BundleEntryRequest{Method: "POST", URL: "Encounter"},
		},
	}

	// Should not panic with empty map.
	ResolveInternalReferences(entries, map[string]string{})

	subject := entries[0].Resource["subject"].(map[string]interface{})
	if subject["reference"] != "urn:uuid:xyz" {
		t.Errorf("expected unchanged reference with empty idMap, got %v", subject["reference"])
	}
}

// ---------------------------------------------------------------------------
// ProcessTransaction tests
// ---------------------------------------------------------------------------

func TestProcessTransaction_AllSuccessful(t *testing.T) {
	callCount := 0
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		callCount++
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: "Patient/" + string(rune('0'+callCount)),
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:a",
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
			{
				FullURL:  "urn:uuid:b",
				Resource: map[string]interface{}{"resourceType": "Observation"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Observation"},
			},
		},
	}

	result, err := processor.ProcessTransaction(bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "transaction-response" {
		t.Errorf("expected transaction-response, got %s", result.Type)
	}
	if len(result.Entry) != 2 {
		t.Fatalf("expected 2 response entries, got %d", len(result.Entry))
	}
	if result.Entry[0].Response.Status != "201 Created" {
		t.Errorf("expected 201 Created, got %s", result.Entry[0].Response.Status)
	}
}

func TestProcessTransaction_FailedEntry_RollsBack(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		if url == "Observation" {
			return nil, errors.New("conflict: resource already exists")
		}
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: "Patient/new1",
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:a",
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
			{
				FullURL:  "urn:uuid:b",
				Resource: map[string]interface{}{"resourceType": "Observation"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Observation"},
			},
		},
	}

	_, err := processor.ProcessTransaction(bundle)
	if err == nil {
		t.Fatal("expected error when entry fails in transaction")
	}
	if !strings.Contains(err.Error(), "transaction failed") {
		t.Errorf("expected 'transaction failed' in error, got: %v", err)
	}
}

func TestProcessTransaction_ResolvesInternalReferences(t *testing.T) {
	var capturedResource map[string]interface{}

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		if method == "POST" && url == "Patient" {
			return &BundleEntryResponse{
				Status:   "201 Created",
				Location: "Patient/actual-id-123",
			}, nil
		}
		// Capture the encounter resource to check if references were resolved.
		capturedResource = resource
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: "Encounter/enc-456",
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:patient-1",
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
			{
				FullURL: "urn:uuid:enc-1",
				Resource: map[string]interface{}{
					"resourceType": "Encounter",
					"subject":      map[string]interface{}{"reference": "urn:uuid:patient-1"},
				},
				Request: BundleEntryRequest{Method: "POST", URL: "Encounter"},
			},
		},
	}

	_, err := processor.ProcessTransaction(bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the urn:uuid reference was resolved before the handler was called.
	if capturedResource != nil {
		subject, ok := capturedResource["subject"].(map[string]interface{})
		if ok {
			if subject["reference"] != "Patient/actual-id-123" {
				t.Errorf("expected resolved reference Patient/actual-id-123, got %v", subject["reference"])
			}
		}
	}
}

func TestProcessTransaction_SortsEntries(t *testing.T) {
	var order []string
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		order = append(order, method)
		return &BundleEntryResponse{Status: "200 OK"}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{FullURL: "urn:uuid:1", Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{FullURL: "urn:uuid:2", Request: BundleEntryRequest{Method: "DELETE", URL: "Patient/2"}},
			{FullURL: "urn:uuid:3", Request: BundleEntryRequest{Method: "POST", URL: "Patient"}, Resource: map[string]interface{}{"resourceType": "Patient"}},
		},
	}

	_, err := processor.ProcessTransaction(bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected order: DELETE, POST, GET.
	if len(order) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(order))
	}
	if order[0] != "DELETE" {
		t.Errorf("expected DELETE first, got %s", order[0])
	}
	if order[1] != "POST" {
		t.Errorf("expected POST second, got %s", order[1])
	}
	if order[2] != "GET" {
		t.Errorf("expected GET third, got %s", order[2])
	}
}

// ---------------------------------------------------------------------------
// ProcessBatch tests
// ---------------------------------------------------------------------------

func TestProcessBatch_MixedSuccessFailure(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		if url == "Patient/bad" {
			return nil, errors.New("not found")
		}
		return &BundleEntryResponse{
			Status:   "200 OK",
			Location: url,
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/bad"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/3"}},
		},
	}

	result := processor.ProcessBatch(bundle)
	if result.Type != "batch-response" {
		t.Errorf("expected batch-response, got %s", result.Type)
	}
	if len(result.Entry) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entry))
	}

	// First entry: success.
	if result.Entry[0].Response.Status != "200 OK" {
		t.Errorf("expected 200 OK for entry 0, got %s", result.Entry[0].Response.Status)
	}

	// Second entry: failure.
	if result.Entry[1].Response.Status != "400 Bad Request" {
		t.Errorf("expected 400 Bad Request for entry 1, got %s", result.Entry[1].Response.Status)
	}
	if result.Entry[1].Response.Outcome == nil {
		t.Error("expected OperationOutcome for failed batch entry")
	}

	// Third entry: success (batch continues after failure).
	if result.Entry[2].Response.Status != "200 OK" {
		t.Errorf("expected 200 OK for entry 2, got %s", result.Entry[2].Response.Status)
	}
}

func TestProcessBatch_ContinuesAfterFailure(t *testing.T) {
	callCount := 0
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		callCount++
		if callCount == 1 {
			return nil, errors.New("first entry fails")
		}
		return &BundleEntryResponse{Status: "200 OK"}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/2"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/3"}},
		},
	}

	result := processor.ProcessBatch(bundle)

	// All 3 entries should be processed.
	if callCount != 3 {
		t.Errorf("expected 3 handler calls, got %d", callCount)
	}
	if len(result.Entry) != 3 {
		t.Fatalf("expected 3 response entries, got %d", len(result.Entry))
	}

	// First failed, rest succeeded.
	if result.Entry[0].Response.Status != "400 Bad Request" {
		t.Errorf("expected 400 for first entry, got %s", result.Entry[0].Response.Status)
	}
	if result.Entry[1].Response.Status != "200 OK" {
		t.Errorf("expected 200 OK for second entry, got %s", result.Entry[1].Response.Status)
	}
	if result.Entry[2].Response.Status != "200 OK" {
		t.Errorf("expected 200 OK for third entry, got %s", result.Entry[2].Response.Status)
	}
}

func TestProcessBatch_AllSuccessful(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{
			Status:       "200 OK",
			Location:     url,
			LastModified: "2024-01-15T10:00:00Z",
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries: []TransactionEntry{
			{Request: BundleEntryRequest{Method: "GET", URL: "Patient/1"}},
			{Request: BundleEntryRequest{Method: "GET", URL: "Observation/2"}},
		},
	}

	result := processor.ProcessBatch(bundle)
	if result.Type != "batch-response" {
		t.Errorf("expected batch-response, got %s", result.Type)
	}
	for i, entry := range result.Entry {
		if entry.Response.Status != "200 OK" {
			t.Errorf("entry %d: expected 200 OK, got %s", i, entry.Response.Status)
		}
	}
}

func TestProcessBatch_EmptyBundle(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{Status: "200 OK"}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "batch",
		Entries:      []TransactionEntry{},
	}

	result := processor.ProcessBatch(bundle)
	if result.Type != "batch-response" {
		t.Errorf("expected batch-response, got %s", result.Type)
	}
	if len(result.Entry) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entry))
	}
}

// ---------------------------------------------------------------------------
// TransactionHandler tests
// ---------------------------------------------------------------------------

func TestTransactionHandler_AcceptsTransactionBundle(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: "Patient/new-1",
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	h := TransactionHandler(processor)

	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:1",
				"resource": {"resourceType": "Patient"},
				"request": {"method": "POST", "url": "Patient"}
			}
		]
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.Type != "transaction-response" {
		t.Errorf("expected transaction-response, got %s", result.Type)
	}
}

func TestTransactionHandler_AcceptsBatchBundle(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return &BundleEntryResponse{Status: "200 OK"}, nil
	}

	processor := NewTransactionProcessor(handler)
	h := TransactionHandler(processor)

	body := `{
		"resourceType": "Bundle",
		"type": "batch",
		"entry": [
			{"request": {"method": "GET", "url": "Patient/1"}}
		]
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result Bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.Type != "batch-response" {
		t.Errorf("expected batch-response, got %s", result.Type)
	}
}

func TestTransactionHandler_RejectsInvalidJSON(t *testing.T) {
	processor := NewTransactionProcessor(func(m, u string, r map[string]interface{}) (*BundleEntryResponse, error) {
		return nil, nil
	})
	h := TransactionHandler(processor)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir", strings.NewReader(`{bad json`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTransactionHandler_RejectsInvalidBundle(t *testing.T) {
	processor := NewTransactionProcessor(func(m, u string, r map[string]interface{}) (*BundleEntryResponse, error) {
		return nil, nil
	})
	h := TransactionHandler(processor)

	// Bundle with invalid method.
	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:1",
				"request": {"method": "INVALID", "url": "Patient"}
			}
		]
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestTransactionHandler_TransactionFailure_Returns400(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		return nil, errors.New("server error")
	}

	processor := NewTransactionProcessor(handler)
	h := TransactionHandler(processor)

	body := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:1",
				"resource": {"resourceType": "Patient"},
				"request": {"method": "POST", "url": "Patient"}
			}
		]
	}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// BundleEntryRequest / BundleEntryResponse serialization tests
// ---------------------------------------------------------------------------

func TestBundleEntryRequest_JSON_Serialization(t *testing.T) {
	req := BundleEntryRequest{
		Method:          "PUT",
		URL:             "Patient/123",
		IfNoneMatch:     "*",
		IfModifiedSince: "2024-01-01T00:00:00Z",
		IfMatch:         `W/"1"`,
		IfNoneExist:     "identifier=http://example.org|12345",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BundleEntryRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("Method: expected %s, got %s", req.Method, decoded.Method)
	}
	if decoded.URL != req.URL {
		t.Errorf("URL: expected %s, got %s", req.URL, decoded.URL)
	}
	if decoded.IfNoneMatch != req.IfNoneMatch {
		t.Errorf("IfNoneMatch: expected %s, got %s", req.IfNoneMatch, decoded.IfNoneMatch)
	}
	if decoded.IfModifiedSince != req.IfModifiedSince {
		t.Errorf("IfModifiedSince: expected %s, got %s", req.IfModifiedSince, decoded.IfModifiedSince)
	}
	if decoded.IfMatch != req.IfMatch {
		t.Errorf("IfMatch: expected %s, got %s", req.IfMatch, decoded.IfMatch)
	}
	if decoded.IfNoneExist != req.IfNoneExist {
		t.Errorf("IfNoneExist: expected %s, got %s", req.IfNoneExist, decoded.IfNoneExist)
	}
}

func TestBundleEntryRequest_JSON_OmitsEmpty(t *testing.T) {
	req := BundleEntryRequest{
		Method: "GET",
		URL:    "Patient/1",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	s := string(data)
	if strings.Contains(s, "ifNoneMatch") {
		t.Error("expected ifNoneMatch to be omitted")
	}
	if strings.Contains(s, "ifModifiedSince") {
		t.Error("expected ifModifiedSince to be omitted")
	}
	if strings.Contains(s, "ifMatch") {
		t.Error("expected ifMatch to be omitted")
	}
	if strings.Contains(s, "ifNoneExist") {
		t.Error("expected ifNoneExist to be omitted")
	}
}

func TestBundleEntryResponse_JSON_Serialization(t *testing.T) {
	resp := BundleEntryResponse{
		Status:       "201 Created",
		Location:     "Patient/123",
		ETag:         `W/"1"`,
		LastModified: "2024-06-15T12:00:00Z",
		Outcome:      SuccessOutcome("created successfully"),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BundleEntryResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Status != resp.Status {
		t.Errorf("Status: expected %s, got %s", resp.Status, decoded.Status)
	}
	if decoded.Location != resp.Location {
		t.Errorf("Location: expected %s, got %s", resp.Location, decoded.Location)
	}
	if decoded.ETag != resp.ETag {
		t.Errorf("ETag: expected %s, got %s", resp.ETag, decoded.ETag)
	}
	if decoded.LastModified != resp.LastModified {
		t.Errorf("LastModified: expected %s, got %s", resp.LastModified, decoded.LastModified)
	}
}

func TestBundleEntryResponse_JSON_OmitsEmpty(t *testing.T) {
	resp := BundleEntryResponse{
		Status: "200 OK",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	s := string(data)
	if strings.Contains(s, "location") {
		t.Error("expected location to be omitted")
	}
	if strings.Contains(s, "etag") {
		t.Error("expected etag to be omitted")
	}
	if strings.Contains(s, "lastModified") {
		t.Error("expected lastModified to be omitted")
	}
}

// ---------------------------------------------------------------------------
// TransactionEntry serialization tests
// ---------------------------------------------------------------------------

func TestTransactionEntry_JSON_RoundTrip(t *testing.T) {
	entry := TransactionEntry{
		FullURL: "urn:uuid:abc",
		Resource: map[string]interface{}{
			"resourceType": "Patient",
			"name": []interface{}{
				map[string]interface{}{"family": "Doe", "given": []interface{}{"John"}},
			},
		},
		Request: BundleEntryRequest{
			Method: "POST",
			URL:    "Patient",
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded TransactionEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.FullURL != entry.FullURL {
		t.Errorf("FullURL: expected %s, got %s", entry.FullURL, decoded.FullURL)
	}
	if decoded.Request.Method != "POST" {
		t.Errorf("Method: expected POST, got %s", decoded.Request.Method)
	}
	if decoded.Resource["resourceType"] != "Patient" {
		t.Error("expected resourceType Patient in decoded resource")
	}
}

// ---------------------------------------------------------------------------
// Internal reference resolution across entries
// ---------------------------------------------------------------------------

func TestProcessTransaction_CrossEntryReferenceResolution(t *testing.T) {
	// Simulates a transaction with Patient, Encounter, and Observation
	// where Encounter references Patient and Observation references both.
	idCounter := 0
	capturedResources := make(map[string]map[string]interface{})

	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		idCounter++
		rt := ""
		if resource != nil {
			rt, _ = resource["resourceType"].(string)
			// Deep copy for verification.
			data, _ := json.Marshal(resource)
			var copy map[string]interface{}
			json.Unmarshal(data, &copy)
			capturedResources[rt] = copy
		}
		location := url
		if method == "POST" {
			location = rt + "/" + string(rune('A'-1+idCounter))
		}
		return &BundleEntryResponse{
			Status:   "201 Created",
			Location: location,
		}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:patient-1",
				Resource: map[string]interface{}{"resourceType": "Patient", "name": "Test"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
			{
				FullURL: "urn:uuid:encounter-1",
				Resource: map[string]interface{}{
					"resourceType": "Encounter",
					"subject":      map[string]interface{}{"reference": "urn:uuid:patient-1"},
				},
				Request: BundleEntryRequest{Method: "POST", URL: "Encounter"},
			},
			{
				FullURL: "urn:uuid:obs-1",
				Resource: map[string]interface{}{
					"resourceType": "Observation",
					"subject":      map[string]interface{}{"reference": "urn:uuid:patient-1"},
					"encounter":    map[string]interface{}{"reference": "urn:uuid:encounter-1"},
				},
				Request: BundleEntryRequest{Method: "POST", URL: "Observation"},
			},
		},
	}

	result, err := processor.ProcessTransaction(bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entry) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entry))
	}

	// Verify the Encounter's subject was resolved.
	if enc, ok := capturedResources["Encounter"]; ok {
		subj := enc["subject"].(map[string]interface{})
		ref := subj["reference"].(string)
		if !strings.HasPrefix(ref, "Patient/") {
			t.Errorf("Encounter subject not resolved, got %s", ref)
		}
	}

	// Verify the Observation references were resolved.
	if obs, ok := capturedResources["Observation"]; ok {
		subj := obs["subject"].(map[string]interface{})
		ref := subj["reference"].(string)
		if !strings.HasPrefix(ref, "Patient/") {
			t.Errorf("Observation subject not resolved, got %s", ref)
		}
		enc := obs["encounter"].(map[string]interface{})
		encRef := enc["reference"].(string)
		if !strings.HasPrefix(encRef, "Encounter/") {
			t.Errorf("Observation encounter not resolved, got %s", encRef)
		}
	}
}

// ---------------------------------------------------------------------------
// NewTransactionProcessor tests
// ---------------------------------------------------------------------------

func TestNewTransactionProcessor(t *testing.T) {
	called := false
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		called = true
		return &BundleEntryResponse{Status: "200 OK"}, nil
	}

	p := NewTransactionProcessor(handler)
	if p == nil {
		t.Fatal("expected non-nil processor")
	}
	if p.ResourceHandler == nil {
		t.Fatal("expected non-nil ResourceHandler")
	}

	// Verify the handler works.
	_, err := p.ResourceHandler("GET", "Patient/1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

// ---------------------------------------------------------------------------
// extractReferences helper tests
// ---------------------------------------------------------------------------

func TestExtractReferences_DeepNesting(t *testing.T) {
	resource := map[string]interface{}{
		"subject": map[string]interface{}{"reference": "Patient/1"},
		"contained": []interface{}{
			map[string]interface{}{
				"author": map[string]interface{}{"reference": "Practitioner/2"},
				"items": []interface{}{
					map[string]interface{}{
						"target": map[string]interface{}{"reference": "Observation/3"},
					},
				},
			},
		},
	}

	refs := extractReferences(resource)
	if len(refs) != 3 {
		t.Fatalf("expected 3 references, got %d: %v", len(refs), refs)
	}

	expected := map[string]bool{
		"Patient/1":      true,
		"Practitioner/2": true,
		"Observation/3":  true,
	}
	for _, ref := range refs {
		if !expected[ref] {
			t.Errorf("unexpected reference: %s", ref)
		}
	}
}

// ---------------------------------------------------------------------------
// TransactionBundle JSON round-trip
// ---------------------------------------------------------------------------

func TestTransactionBundle_JSON_RoundTrip(t *testing.T) {
	bundle := TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL:  "urn:uuid:x",
				Resource: map[string]interface{}{"resourceType": "Patient"},
				Request:  BundleEntryRequest{Method: "POST", URL: "Patient"},
			},
		},
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded TransactionBundle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ResourceType != "Bundle" {
		t.Errorf("expected Bundle, got %s", decoded.ResourceType)
	}
	if decoded.Type != "transaction" {
		t.Errorf("expected transaction, got %s", decoded.Type)
	}
	if len(decoded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(decoded.Entries))
	}
}

// ---------------------------------------------------------------------------
// Edge cases and additional coverage
// ---------------------------------------------------------------------------

func TestProcessTransaction_DeleteEntry(t *testing.T) {
	handler := func(method, url string, resource map[string]interface{}) (*BundleEntryResponse, error) {
		if method == "DELETE" {
			return &BundleEntryResponse{
				Status: "204 No Content",
			}, nil
		}
		return &BundleEntryResponse{Status: "200 OK"}, nil
	}

	processor := NewTransactionProcessor(handler)
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				FullURL: "Patient/to-delete",
				Request: BundleEntryRequest{Method: "DELETE", URL: "Patient/to-delete"},
			},
		},
	}

	result, err := processor.ProcessTransaction(bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Entry[0].Response.Status != "204 No Content" {
		t.Errorf("expected 204 No Content, got %s", result.Entry[0].Response.Status)
	}
}

func TestBundleEntryFromResponse_ParsesLastModified(t *testing.T) {
	resp := &BundleEntryResponse{
		Status:       "200 OK",
		Location:     "Patient/1",
		LastModified: "2024-06-15T12:30:00Z",
	}

	entry := bundleEntryFromResponse(resp)
	if entry.Response.LastModified == nil {
		t.Fatal("expected LastModified to be parsed")
	}
	if entry.Response.LastModified.Year() != 2024 {
		t.Errorf("expected year 2024, got %d", entry.Response.LastModified.Year())
	}
}

func TestBundleEntryFromResponse_InvalidLastModified(t *testing.T) {
	resp := &BundleEntryResponse{
		Status:       "200 OK",
		LastModified: "not-a-date",
	}

	entry := bundleEntryFromResponse(resp)
	if entry.Response.LastModified != nil {
		t.Error("expected nil LastModified for invalid date")
	}
}

func TestTransactionHandler_RejectsWrongResourceType(t *testing.T) {
	processor := NewTransactionProcessor(func(m, u string, r map[string]interface{}) (*BundleEntryResponse, error) {
		return nil, nil
	})
	h := TransactionHandler(processor)

	body := `{"resourceType": "Patient", "type": "transaction"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/fhir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestValidateTransactionBundle_MultipleErrors(t *testing.T) {
	bundle := &TransactionBundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		Entries: []TransactionEntry{
			{
				// Missing fullUrl (required for transaction), missing method, missing url.
				Request: BundleEntryRequest{},
			},
		},
	}
	issues := ValidateTransactionBundle(bundle)
	// Should have at least 3 issues: missing method, missing url, missing fullUrl.
	if len(issues) < 3 {
		t.Errorf("expected at least 3 issues, got %d: %+v", len(issues), issues)
	}
}
