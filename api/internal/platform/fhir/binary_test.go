package fhir

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// BinaryStore in-memory implementation tests
// ---------------------------------------------------------------------------

func TestBinaryStore_Create(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ContentType:     "application/pdf",
		Data:            []byte("hello world"),
		SecurityContext: &Reference{Reference: "Patient/123"},
	}

	created, err := store.Create(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.ContentType != "application/pdf" {
		t.Errorf("expected contentType application/pdf, got %s", created.ContentType)
	}
	if !bytes.Equal(created.Data, []byte("hello world")) {
		t.Errorf("expected data 'hello world', got %s", string(created.Data))
	}
	if created.SecurityContext == nil || created.SecurityContext.Reference != "Patient/123" {
		t.Error("expected securityContext Patient/123")
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestBinaryStore_Create_EmptyContentType(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		Data: []byte("data"),
	}

	_, err := store.Create(ctx, res)
	if err == nil {
		t.Fatal("expected error for empty content type")
	}
	if err != ErrBinaryMissingContentType {
		t.Errorf("expected ErrBinaryMissingContentType, got %v", err)
	}
}

func TestBinaryStore_Create_EmptyData(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ContentType: "text/plain",
	}

	_, err := store.Create(ctx, res)
	if err == nil {
		t.Fatal("expected error for empty data")
	}
	if err != ErrBinaryMissingData {
		t.Errorf("expected ErrBinaryMissingData, got %v", err)
	}
}

func TestBinaryStore_Create_ExceedsSizeLimit(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	// MaxBinarySize is 50 MB; create data exceeding that.
	bigData := make([]byte, MaxBinarySize+1)
	res := &BinaryResource{
		ContentType: "application/octet-stream",
		Data:        bigData,
	}

	_, err := store.Create(ctx, res)
	if err == nil {
		t.Fatal("expected error for oversized data")
	}
	if err != ErrBinaryTooLarge {
		t.Errorf("expected ErrBinaryTooLarge, got %v", err)
	}
}

func TestBinaryStore_GetByID(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ContentType: "image/png",
		Data:        []byte{0x89, 0x50, 0x4E, 0x47},
	}

	created, _ := store.Create(ctx, res)

	fetched, err := store.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, fetched.ID)
	}
	if !bytes.Equal(fetched.Data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		t.Error("data mismatch")
	}
}

func TestBinaryStore_GetByID_NotFound(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	_, err := store.GetByID(ctx, "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
	if err != ErrBinaryNotFound {
		t.Errorf("expected ErrBinaryNotFound, got %v", err)
	}
}

func TestBinaryStore_Update(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ContentType: "text/plain",
		Data:        []byte("original"),
	}
	created, _ := store.Create(ctx, res)

	created.Data = []byte("updated")
	created.ContentType = "text/html"
	created.SecurityContext = &Reference{Reference: "Practitioner/456"}

	updated, err := store.Update(ctx, created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ContentType != "text/html" {
		t.Errorf("expected text/html, got %s", updated.ContentType)
	}
	if !bytes.Equal(updated.Data, []byte("updated")) {
		t.Error("data not updated")
	}
	if updated.SecurityContext == nil || updated.SecurityContext.Reference != "Practitioner/456" {
		t.Error("securityContext not updated")
	}
}

func TestBinaryStore_Update_NotFound(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ID:          "nonexistent",
		ContentType: "text/plain",
		Data:        []byte("data"),
	}

	_, err := store.Update(ctx, res)
	if err == nil {
		t.Fatal("expected error for nonexistent resource")
	}
	if err != ErrBinaryNotFound {
		t.Errorf("expected ErrBinaryNotFound, got %v", err)
	}
}

func TestBinaryStore_Update_EmptyContentType(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ContentType: "text/plain",
		Data:        []byte("data"),
	}
	created, _ := store.Create(ctx, res)

	created.ContentType = ""
	_, err := store.Update(ctx, created)
	if err != ErrBinaryMissingContentType {
		t.Errorf("expected ErrBinaryMissingContentType, got %v", err)
	}
}

func TestBinaryStore_Delete(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	res := &BinaryResource{
		ContentType: "text/plain",
		Data:        []byte("to delete"),
	}
	created, _ := store.Create(ctx, res)

	err := store.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = store.GetByID(ctx, created.ID)
	if err != ErrBinaryNotFound {
		t.Errorf("expected ErrBinaryNotFound after deletion, got %v", err)
	}
}

func TestBinaryStore_Delete_NotFound(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent resource")
	}
	if err != ErrBinaryNotFound {
		t.Errorf("expected ErrBinaryNotFound, got %v", err)
	}
}

func TestBinaryStore_List_Empty(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	items, total, err := store.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestBinaryStore_List_Pagination(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	for i := 0; i < 15; i++ {
		store.Create(ctx, &BinaryResource{
			ContentType: "text/plain",
			Data:        []byte(fmt.Sprintf("item %d", i)),
		})
	}

	// First page
	items, total, err := store.List(ctx, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 15 {
		t.Errorf("expected total 15, got %d", total)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}

	// Beyond range
	items4, total4, _ := store.List(ctx, 5, 20)
	if total4 != 15 {
		t.Errorf("expected total 15, got %d", total4)
	}
	if len(items4) != 0 {
		t.Errorf("expected 0 items, got %d", len(items4))
	}
}

func TestBinaryStore_List_DefaultLimit(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		store.Create(ctx, &BinaryResource{
			ContentType: "text/plain",
			Data:        []byte(fmt.Sprintf("item %d", i)),
		})
	}

	items, _, err := store.List(ctx, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 20 {
		t.Errorf("expected default limit of 20 items, got %d", len(items))
	}
}

func TestBinaryStore_IsolatesMutations(t *testing.T) {
	store := NewInMemoryBinaryStore()
	ctx := context.Background()

	original := &BinaryResource{
		ContentType: "text/plain",
		Data:        []byte("immutable"),
	}
	created, _ := store.Create(ctx, original)

	// Mutate the returned value
	created.Data[0] = 'X'

	// Fetch again - should be unmodified
	fetched, _ := store.GetByID(ctx, created.ID)
	if fetched.Data[0] == 'X' {
		t.Error("store did not isolate data; mutation leaked through")
	}
}

// ---------------------------------------------------------------------------
// FHIR JSON mapping tests
// ---------------------------------------------------------------------------

func TestBinaryResource_ToFHIRJSON(t *testing.T) {
	res := &BinaryResource{
		ID:              "abc-123",
		ContentType:     "application/pdf",
		Data:            []byte("PDF content here"),
		SecurityContext: &Reference{Reference: "Patient/42"},
	}

	fhirJSON := res.ToFHIRJSON()
	if fhirJSON.ResourceType != "Binary" {
		t.Errorf("expected resourceType Binary, got %s", fhirJSON.ResourceType)
	}
	if fhirJSON.ID != "abc-123" {
		t.Errorf("expected ID abc-123, got %s", fhirJSON.ID)
	}
	if fhirJSON.ContentType != "application/pdf" {
		t.Errorf("expected contentType application/pdf, got %s", fhirJSON.ContentType)
	}

	expectedB64 := base64.StdEncoding.EncodeToString([]byte("PDF content here"))
	if fhirJSON.Data != expectedB64 {
		t.Errorf("expected base64 data %s, got %s", expectedB64, fhirJSON.Data)
	}
	if fhirJSON.SecurityContext == nil || fhirJSON.SecurityContext.Reference != "Patient/42" {
		t.Error("expected securityContext Patient/42")
	}
}

func TestBinaryResource_ToFHIRJSON_NilSecurityContext(t *testing.T) {
	res := &BinaryResource{
		ID:          "def-456",
		ContentType: "text/plain",
		Data:        []byte("text"),
	}

	fhirJSON := res.ToFHIRJSON()
	if fhirJSON.SecurityContext != nil {
		t.Error("expected nil securityContext")
	}
}

func TestBinaryFHIRJSON_ToBinaryResource(t *testing.T) {
	raw := []byte("raw binary data")
	b64 := base64.StdEncoding.EncodeToString(raw)

	fhirJSON := &BinaryFHIRJSON{
		ResourceType:    "Binary",
		ID:              "id-1",
		ContentType:     "image/png",
		Data:            b64,
		SecurityContext: &Reference{Reference: "Organization/1"},
	}

	res, err := fhirJSON.ToBinaryResource()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != "id-1" {
		t.Errorf("expected ID id-1, got %s", res.ID)
	}
	if res.ContentType != "image/png" {
		t.Errorf("expected contentType image/png, got %s", res.ContentType)
	}
	if !bytes.Equal(res.Data, raw) {
		t.Errorf("data mismatch: expected %v, got %v", raw, res.Data)
	}
	if res.SecurityContext == nil || res.SecurityContext.Reference != "Organization/1" {
		t.Error("expected securityContext Organization/1")
	}
}

func TestBinaryFHIRJSON_ToBinaryResource_InvalidBase64(t *testing.T) {
	fhirJSON := &BinaryFHIRJSON{
		ResourceType: "Binary",
		ContentType:  "text/plain",
		Data:         "!!!not-base64!!!",
	}

	_, err := fhirJSON.ToBinaryResource()
	if err == nil {
		t.Fatal("expected error for invalid base64 data")
	}
}

func TestBinaryFHIRJSON_Roundtrip(t *testing.T) {
	original := &BinaryResource{
		ID:              "roundtrip-1",
		ContentType:     "application/octet-stream",
		Data:            []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		SecurityContext: &Reference{Reference: "Patient/999"},
	}

	fhirJSON := original.ToFHIRJSON()

	// Serialize to actual JSON
	jsonBytes, err := json.Marshal(fhirJSON)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Parse back
	var parsed BinaryFHIRJSON
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	restored, err := parsed.ToBinaryResource()
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", restored.ID, original.ID)
	}
	if restored.ContentType != original.ContentType {
		t.Errorf("ContentType mismatch: %s vs %s", restored.ContentType, original.ContentType)
	}
	if !bytes.Equal(restored.Data, original.Data) {
		t.Error("Data mismatch after roundtrip")
	}
}

func TestBinaryFHIRJSON_EmptyData(t *testing.T) {
	fhirJSON := &BinaryFHIRJSON{
		ResourceType: "Binary",
		ContentType:  "text/plain",
		Data:         "",
	}

	res, err := fhirJSON.ToBinaryResource()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Data) != 0 {
		t.Errorf("expected empty data, got %d bytes", len(res.Data))
	}
}

// ---------------------------------------------------------------------------
// HTTP Handler tests
// ---------------------------------------------------------------------------

func setupBinaryTestServer() (*echo.Echo, *BinaryHandler) {
	e := echo.New()
	store := NewInMemoryBinaryStore()
	handler := NewBinaryHandler(store)
	g := e.Group("/fhir")
	handler.RegisterRoutes(g)
	return e, handler
}

func TestHandler_CreateBinary(t *testing.T) {
	e, _ := setupBinaryTestServer()

	body := `{
		"resourceType": "Binary",
		"contentType": "application/pdf",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("pdf content")) + `",
		"securityContext": {"reference": "Patient/1"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var result BinaryFHIRJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result.ResourceType != "Binary" {
		t.Errorf("expected resourceType Binary, got %s", result.ResourceType)
	}
	if result.ID == "" {
		t.Error("expected non-empty ID")
	}
	if result.ContentType != "application/pdf" {
		t.Errorf("expected contentType application/pdf, got %s", result.ContentType)
	}

	// Check Location header
	loc := rec.Header().Get("Location")
	if loc == "" {
		t.Error("expected Location header")
	}
}

func TestHandler_CreateBinary_InvalidJSON(t *testing.T) {
	e, _ := setupBinaryTestServer()

	req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_CreateBinary_MissingContentType(t *testing.T) {
	e, _ := setupBinaryTestServer()

	body := `{
		"resourceType": "Binary",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("data")) + `"
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_CreateBinary_InvalidBase64(t *testing.T) {
	e, _ := setupBinaryTestServer()

	body := `{
		"resourceType": "Binary",
		"contentType": "text/plain",
		"data": "!!!not-valid-base64!!!"
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_GetBinary_FHIRJson(t *testing.T) {
	e, _ := setupBinaryTestServer()

	// Create a resource first
	createBody := `{
		"resourceType": "Binary",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("hello")) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Get with Accept: application/fhir+json (FHIR JSON)
	getReq := httptest.NewRequest(http.MethodGet, "/fhir/Binary/"+created.ID, nil)
	getReq.Header.Set("Accept", "application/fhir+json")
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	ct := getRec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/fhir+json") {
		t.Errorf("expected Content-Type application/fhir+json, got %s", ct)
	}

	var result BinaryFHIRJSON
	if err := json.Unmarshal(getRec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result.ResourceType != "Binary" {
		t.Errorf("expected resourceType Binary, got %s", result.ResourceType)
	}
	if result.ContentType != "text/plain" {
		t.Errorf("expected contentType text/plain, got %s", result.ContentType)
	}

	decoded, _ := base64.StdEncoding.DecodeString(result.Data)
	if string(decoded) != "hello" {
		t.Errorf("expected data 'hello', got '%s'", string(decoded))
	}
}

func TestHandler_GetBinary_RawContent(t *testing.T) {
	e, _ := setupBinaryTestServer()

	rawContent := []byte("This is raw text content")

	// Create a resource
	createBody := `{
		"resourceType": "Binary",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString(rawContent) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Get with Accept matching the binary's own content type -> raw bytes
	getReq := httptest.NewRequest(http.MethodGet, "/fhir/Binary/"+created.ID, nil)
	getReq.Header.Set("Accept", "text/plain")
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	ct := getRec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got %s", ct)
	}

	if !bytes.Equal(getRec.Body.Bytes(), rawContent) {
		t.Errorf("expected raw content, got %s", getRec.Body.String())
	}
}

func TestHandler_GetBinary_DefaultToFHIRJson(t *testing.T) {
	e, _ := setupBinaryTestServer()

	createBody := `{
		"resourceType": "Binary",
		"contentType": "image/png",
		"data": "` + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50}) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Get without Accept header -> default to FHIR JSON
	getReq := httptest.NewRequest(http.MethodGet, "/fhir/Binary/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}

	var result BinaryFHIRJSON
	if err := json.Unmarshal(getRec.Body.Bytes(), &result); err != nil {
		t.Fatalf("expected FHIR JSON response, got parse error: %v", err)
	}
	if result.ResourceType != "Binary" {
		t.Error("expected resourceType Binary")
	}
}

func TestHandler_GetBinary_NotFound(t *testing.T) {
	e, _ := setupBinaryTestServer()

	req := httptest.NewRequest(http.MethodGet, "/fhir/Binary/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_UpdateBinary(t *testing.T) {
	e, _ := setupBinaryTestServer()

	// Create
	createBody := `{
		"resourceType": "Binary",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("original")) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Update
	updateBody := `{
		"resourceType": "Binary",
		"id": "` + created.ID + `",
		"contentType": "text/html",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("<h1>updated</h1>")) + `",
		"securityContext": {"reference": "Patient/42"}
	}`
	updateReq := httptest.NewRequest(http.MethodPut, "/fhir/Binary/"+created.ID, strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/fhir+json")
	updateRec := httptest.NewRecorder()
	e.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", updateRec.Code, updateRec.Body.String())
	}

	var result BinaryFHIRJSON
	json.Unmarshal(updateRec.Body.Bytes(), &result)
	if result.ContentType != "text/html" {
		t.Errorf("expected text/html, got %s", result.ContentType)
	}

	decoded, _ := base64.StdEncoding.DecodeString(result.Data)
	if string(decoded) != "<h1>updated</h1>" {
		t.Errorf("expected updated data, got %s", string(decoded))
	}
}

func TestHandler_UpdateBinary_IDMismatch(t *testing.T) {
	e, _ := setupBinaryTestServer()

	// Create
	createBody := `{
		"resourceType": "Binary",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("data")) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Update with mismatched ID
	updateBody := `{
		"resourceType": "Binary",
		"id": "different-id",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("data")) + `"
	}`
	updateReq := httptest.NewRequest(http.MethodPut, "/fhir/Binary/"+created.ID, strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/fhir+json")
	updateRec := httptest.NewRecorder()
	e.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for ID mismatch, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
}

func TestHandler_UpdateBinary_NotFound(t *testing.T) {
	e, _ := setupBinaryTestServer()

	updateBody := `{
		"resourceType": "Binary",
		"id": "nonexistent",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("data")) + `"
	}`
	updateReq := httptest.NewRequest(http.MethodPut, "/fhir/Binary/nonexistent", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/fhir+json")
	updateRec := httptest.NewRecorder()
	e.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", updateRec.Code)
	}
}

func TestHandler_DeleteBinary(t *testing.T) {
	e, _ := setupBinaryTestServer()

	// Create
	createBody := `{
		"resourceType": "Binary",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("to delete")) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Delete
	delReq := httptest.NewRequest(http.MethodDelete, "/fhir/Binary/"+created.ID, nil)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", delRec.Code)
	}

	// Verify deleted
	getReq := httptest.NewRequest(http.MethodGet, "/fhir/Binary/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", getRec.Code)
	}
}

func TestHandler_DeleteBinary_NotFound(t *testing.T) {
	e, _ := setupBinaryTestServer()

	req := httptest.NewRequest(http.MethodDelete, "/fhir/Binary/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_ListBinaries(t *testing.T) {
	e, _ := setupBinaryTestServer()

	// Create a few resources
	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{
			"resourceType": "Binary",
			"contentType": "text/plain",
			"data": "%s"
		}`, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("item %d", i))))

		req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/fhir+json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}

	// List
	req := httptest.NewRequest(http.MethodGet, "/fhir/Binary", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &bundle); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("expected resourceType Bundle, got %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("expected type searchset, got %v", bundle["type"])
	}
	total, ok := bundle["total"].(float64)
	if !ok || int(total) != 3 {
		t.Errorf("expected total 3, got %v", bundle["total"])
	}
	entries, ok := bundle["entry"].([]interface{})
	if !ok || len(entries) != 3 {
		t.Errorf("expected 3 entries, got %v", len(entries))
	}
}

func TestHandler_ListBinaries_Pagination(t *testing.T) {
	e, _ := setupBinaryTestServer()

	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{
			"resourceType": "Binary",
			"contentType": "text/plain",
			"data": "%s"
		}`, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("item %d", i))))

		req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/fhir+json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/fhir/Binary?_count=2&_offset=0", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var bundle map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &bundle)

	total := int(bundle["total"].(float64))
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	entries := bundle["entry"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandler_CreateBinary_SizeLimit(t *testing.T) {
	e, _ := setupBinaryTestServer()

	// Create data that exceeds the max binary size
	bigData := make([]byte, MaxBinarySize+1)
	b64Data := base64.StdEncoding.EncodeToString(bigData)

	body := `{
		"resourceType": "Binary",
		"contentType": "application/octet-stream",
		"data": "` + b64Data + `"
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ContentNegotiation_WildcardAccept(t *testing.T) {
	e, _ := setupBinaryTestServer()

	createBody := `{
		"resourceType": "Binary",
		"contentType": "application/pdf",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("pdf")) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Accept: */* should return FHIR JSON
	getReq := httptest.NewRequest(http.MethodGet, "/fhir/Binary/"+created.ID, nil)
	getReq.Header.Set("Accept", "*/*")
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}

	var result BinaryFHIRJSON
	if err := json.Unmarshal(getRec.Body.Bytes(), &result); err != nil {
		t.Fatalf("expected FHIR JSON for */* accept, got parse error: %v", err)
	}
}

func TestHandler_ContentNegotiation_JsonAccept(t *testing.T) {
	e, _ := setupBinaryTestServer()

	createBody := `{
		"resourceType": "Binary",
		"contentType": "image/png",
		"data": "` + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50}) + `"
	}`
	createReq := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/fhir+json")
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var created BinaryFHIRJSON
	json.Unmarshal(createRec.Body.Bytes(), &created)

	// Accept: application/json should also return FHIR JSON
	getReq := httptest.NewRequest(http.MethodGet, "/fhir/Binary/"+created.ID, nil)
	getReq.Header.Set("Accept", "application/json")
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Code)
	}

	var result BinaryFHIRJSON
	if err := json.Unmarshal(getRec.Body.Bytes(), &result); err != nil {
		t.Fatalf("expected FHIR JSON for application/json accept: %v", err)
	}
}

func TestHandler_CreateBinary_WrongResourceType(t *testing.T) {
	e, _ := setupBinaryTestServer()

	body := `{
		"resourceType": "Patient",
		"contentType": "text/plain",
		"data": "` + base64.StdEncoding.EncodeToString([]byte("data")) + `"
	}`

	req := httptest.NewRequest(http.MethodPost, "/fhir/Binary", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong resourceType, got %d", rec.Code)
	}
}
