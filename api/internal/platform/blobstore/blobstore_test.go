package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func seedBlob(t *testing.T, store BlobStore, patientID, category, fileName, contentType, content string) *BlobMetadata {
	t.Helper()
	meta := BlobMetadata{
		FileName:    fileName,
		ContentType: contentType,
		PatientID:   patientID,
		Category:    category,
		CreatedBy:   "test-user",
		Tags:        map[string]string{"source": "unit-test"},
	}
	result, err := store.Upload(context.Background(), meta, strings.NewReader(content))
	if err != nil {
		t.Fatalf("seedBlob: %v", err)
	}
	return result
}

// ---------------------------------------------------------------------------
// Store tests
// ---------------------------------------------------------------------------

func TestInMemoryBlobStore_Upload(t *testing.T) {
	store := NewInMemoryBlobStore()
	content := "hello world"

	meta := BlobMetadata{
		FileName:    "test.txt",
		ContentType: "text/plain",
		PatientID:   "patient-1",
		Category:    "other",
		CreatedBy:   "user-1",
		Tags:        map[string]string{"env": "test"},
	}

	result, err := store.Upload(context.Background(), meta, strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if result.FileName != "test.txt" {
		t.Errorf("expected FileName=test.txt, got %s", result.FileName)
	}
	if result.ContentType != "text/plain" {
		t.Errorf("expected ContentType=text/plain, got %s", result.ContentType)
	}
	if result.Size != int64(len(content)) {
		t.Errorf("expected Size=%d, got %d", len(content), result.Size)
	}
	if result.Hash == "" {
		t.Fatal("expected non-empty Hash")
	}
	if result.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if result.PatientID != "patient-1" {
		t.Errorf("expected PatientID=patient-1, got %s", result.PatientID)
	}
}

func TestInMemoryBlobStore_Download(t *testing.T) {
	store := NewInMemoryBlobStore()
	content := "binary-content-here"

	uploaded := seedBlob(t, store, "p1", "lab-report", "report.pdf", "application/pdf", content)

	rc, meta, err := store.Download(context.Background(), uploaded.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("error reading content: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected content=%q, got %q", content, string(data))
	}
	if meta.FileName != "report.pdf" {
		t.Errorf("expected FileName=report.pdf, got %s", meta.FileName)
	}
}

func TestInMemoryBlobStore_DownloadNotFound(t *testing.T) {
	store := NewInMemoryBlobStore()

	_, _, err := store.Download(context.Background(), "nonexistent-id")
	if err != ErrBlobNotFound {
		t.Errorf("expected ErrBlobNotFound, got %v", err)
	}
}

func TestInMemoryBlobStore_Delete(t *testing.T) {
	store := NewInMemoryBlobStore()
	uploaded := seedBlob(t, store, "p1", "other", "file.txt", "text/plain", "data")

	err := store.Delete(context.Background(), uploaded.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's gone.
	_, _, err = store.Download(context.Background(), uploaded.ID)
	if err != ErrBlobNotFound {
		t.Errorf("expected ErrBlobNotFound after delete, got %v", err)
	}
}

func TestInMemoryBlobStore_DeleteNotFound(t *testing.T) {
	store := NewInMemoryBlobStore()

	err := store.Delete(context.Background(), "nonexistent-id")
	if err != ErrBlobNotFound {
		t.Errorf("expected ErrBlobNotFound, got %v", err)
	}
}

func TestInMemoryBlobStore_GetMetadata(t *testing.T) {
	store := NewInMemoryBlobStore()
	uploaded := seedBlob(t, store, "p1", "clinical-image", "scan.png", "image/png", "image-data")

	meta, err := store.GetMetadata(context.Background(), uploaded.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.ID != uploaded.ID {
		t.Errorf("expected ID=%s, got %s", uploaded.ID, meta.ID)
	}
	if meta.FileName != "scan.png" {
		t.Errorf("expected FileName=scan.png, got %s", meta.FileName)
	}
	if meta.Category != "clinical-image" {
		t.Errorf("expected Category=clinical-image, got %s", meta.Category)
	}
}

func TestInMemoryBlobStore_ListByPatient(t *testing.T) {
	store := NewInMemoryBlobStore()
	seedBlob(t, store, "patient-A", "lab-report", "a1.pdf", "application/pdf", "a1")
	seedBlob(t, store, "patient-A", "clinical-image", "a2.png", "image/png", "a2")
	seedBlob(t, store, "patient-B", "other", "b1.txt", "text/plain", "b1")

	results, total, err := store.ListByPatient(context.Background(), "patient-A", "", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryBlobStore_ListByPatientAndCategory(t *testing.T) {
	store := NewInMemoryBlobStore()
	seedBlob(t, store, "patient-A", "lab-report", "a1.pdf", "application/pdf", "a1")
	seedBlob(t, store, "patient-A", "clinical-image", "a2.png", "image/png", "a2")
	seedBlob(t, store, "patient-A", "lab-report", "a3.pdf", "application/pdf", "a3")

	results, total, err := store.ListByPatient(context.Background(), "patient-A", "lab-report", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryBlobStore_Search_ByContentType(t *testing.T) {
	store := NewInMemoryBlobStore()
	seedBlob(t, store, "p1", "other", "doc.pdf", "application/pdf", "pdf-content")
	seedBlob(t, store, "p1", "other", "img.png", "image/png", "png-content")

	results, total, err := store.Search(context.Background(), SearchParams{
		ContentType: "application/pdf",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].ContentType != "application/pdf" {
		t.Errorf("expected content type application/pdf, got %s", results[0].ContentType)
	}
}

func TestInMemoryBlobStore_Search_ByDateRange(t *testing.T) {
	store := NewInMemoryBlobStore()

	// Seed some blobs; they will have CreatedAt = now.
	seedBlob(t, store, "p1", "other", "recent.txt", "text/plain", "recent")

	now := time.Now()
	after := now.Add(-1 * time.Hour)
	before := now.Add(1 * time.Hour)

	results, total, err := store.Search(context.Background(), SearchParams{
		CreatedAfter:  &after,
		CreatedBefore: &before,
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Search outside the range.
	pastEnd := now.Add(-2 * time.Hour)
	pastStart := now.Add(-3 * time.Hour)
	results2, total2, err := store.Search(context.Background(), SearchParams{
		CreatedAfter:  &pastStart,
		CreatedBefore: &pastEnd,
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total2 != 0 {
		t.Errorf("expected total=0, got %d", total2)
	}
	if len(results2) != 0 {
		t.Errorf("expected 0 results, got %d", len(results2))
	}
}

func TestInMemoryBlobStore_Search_ByFileName(t *testing.T) {
	store := NewInMemoryBlobStore()
	seedBlob(t, store, "p1", "other", "blood-test-report.pdf", "application/pdf", "data1")
	seedBlob(t, store, "p1", "other", "xray-image.png", "image/png", "data2")

	results, total, err := store.Search(context.Background(), SearchParams{
		FileName: "blood-test",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryBlobStore_Search_ByTags(t *testing.T) {
	store := NewInMemoryBlobStore()

	meta1 := BlobMetadata{
		FileName:    "tagged.txt",
		ContentType: "text/plain",
		Category:    "other",
		CreatedBy:   "user",
		Tags:        map[string]string{"department": "radiology", "priority": "high"},
	}
	store.Upload(context.Background(), meta1, strings.NewReader("tagged-content"))

	meta2 := BlobMetadata{
		FileName:    "other.txt",
		ContentType: "text/plain",
		Category:    "other",
		CreatedBy:   "user",
		Tags:        map[string]string{"department": "cardiology"},
	}
	store.Upload(context.Background(), meta2, strings.NewReader("other-content"))

	results, total, err := store.Search(context.Background(), SearchParams{
		Tags:  map[string]string{"department": "radiology"},
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryBlobStore_Upload_FileTooLarge(t *testing.T) {
	store := NewInMemoryBlobStore()

	// Create a reader that reports a size larger than MaxFileSize.
	largeContent := make([]byte, MaxFileSize+1)

	meta := BlobMetadata{
		FileName:    "huge.bin",
		ContentType: "application/pdf",
		Category:    "other",
		CreatedBy:   "user",
	}

	_, err := store.Upload(context.Background(), meta, bytes.NewReader(largeContent))
	if err != ErrFileTooLarge {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestInMemoryBlobStore_Upload_MissingFileName(t *testing.T) {
	store := NewInMemoryBlobStore()

	meta := BlobMetadata{
		FileName:    "",
		ContentType: "text/plain",
		Category:    "other",
		CreatedBy:   "user",
	}

	_, err := store.Upload(context.Background(), meta, strings.NewReader("data"))
	if err != ErrMissingFileName {
		t.Errorf("expected ErrMissingFileName, got %v", err)
	}
}

func TestInMemoryBlobStore_SHA256Hash(t *testing.T) {
	store := NewInMemoryBlobStore()
	content := "compute-my-hash"

	uploaded := seedBlob(t, store, "p1", "other", "hash.txt", "text/plain", content)

	h := sha256.Sum256([]byte(content))
	expected := fmt.Sprintf("%x", h)

	if uploaded.Hash != expected {
		t.Errorf("expected hash=%s, got %s", expected, uploaded.Hash)
	}
}

func TestInMemoryBlobStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryBlobStore()
	var wg sync.WaitGroup
	const goroutines = 50

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("file-%d.txt", n)
			content := fmt.Sprintf("content-%d", n)
			meta := BlobMetadata{
				FileName:    name,
				ContentType: "text/plain",
				PatientID:   "concurrent-patient",
				Category:    "other",
				CreatedBy:   "user",
			}
			result, err := store.Upload(context.Background(), meta, strings.NewReader(content))
			if err != nil {
				t.Errorf("upload goroutine %d: %v", n, err)
				return
			}

			// Read back.
			rc, _, err := store.Download(context.Background(), result.ID)
			if err != nil {
				t.Errorf("download goroutine %d: %v", n, err)
				return
			}
			rc.Close()

			// Get metadata.
			_, err = store.GetMetadata(context.Background(), result.ID)
			if err != nil {
				t.Errorf("getmetadata goroutine %d: %v", n, err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all uploads visible.
	results, total, err := store.ListByPatient(context.Background(), "concurrent-patient", "", 100, 0)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if total != goroutines {
		t.Errorf("expected total=%d, got %d", goroutines, total)
	}
	if len(results) != goroutines {
		t.Errorf("expected %d results, got %d", goroutines, len(results))
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func newTestHandler() (*BlobHandler, *echo.Echo) {
	store := NewInMemoryBlobStore()
	handler := NewBlobHandler(store)
	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)
	return handler, e
}

func TestBlobHandler_Upload(t *testing.T) {
	_, e := newTestHandler()

	// Build multipart form.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("patient_id", "p-100")
	writer.WriteField("category", "lab-report")
	writer.WriteField("created_by", "dr-smith")

	part, err := writer.CreateFormFile("file", "lab-results.pdf")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("pdf-content-bytes"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/blobs/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var result BlobMetadata
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("error unmarshaling response: %v", err)
	}
	if result.ID == "" {
		t.Error("expected non-empty ID in response")
	}
	if result.FileName != "lab-results.pdf" {
		t.Errorf("expected FileName=lab-results.pdf, got %s", result.FileName)
	}
}

func TestBlobHandler_Download(t *testing.T) {
	store := NewInMemoryBlobStore()
	handler := NewBlobHandler(store)
	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	uploaded := seedBlob(t, store, "p1", "other", "download.txt", "text/plain", "download-me")

	req := httptest.NewRequest(http.MethodGet, "/blobs/"+uploaded.ID, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain" {
		t.Errorf("expected Content-Type=text/plain, got %s", ct)
	}
	if rec.Body.String() != "download-me" {
		t.Errorf("expected body=download-me, got %s", rec.Body.String())
	}
}

func TestBlobHandler_GetMetadata(t *testing.T) {
	store := NewInMemoryBlobStore()
	handler := NewBlobHandler(store)
	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	uploaded := seedBlob(t, store, "p1", "radiology", "xray.png", "image/png", "xray-data")

	req := httptest.NewRequest(http.MethodGet, "/blobs/"+uploaded.ID+"/metadata", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result BlobMetadata
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("error unmarshaling: %v", err)
	}
	if result.ID != uploaded.ID {
		t.Errorf("expected ID=%s, got %s", uploaded.ID, result.ID)
	}
	if result.Category != "radiology" {
		t.Errorf("expected Category=radiology, got %s", result.Category)
	}
}

func TestBlobHandler_Delete(t *testing.T) {
	store := NewInMemoryBlobStore()
	handler := NewBlobHandler(store)
	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	uploaded := seedBlob(t, store, "p1", "other", "delete-me.txt", "text/plain", "bye")

	req := httptest.NewRequest(http.MethodDelete, "/blobs/"+uploaded.ID, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBlobHandler_ListByPatient(t *testing.T) {
	store := NewInMemoryBlobStore()
	handler := NewBlobHandler(store)
	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	seedBlob(t, store, "patient-X", "lab-report", "r1.pdf", "application/pdf", "r1")
	seedBlob(t, store, "patient-X", "clinical-image", "r2.png", "image/png", "r2")
	seedBlob(t, store, "patient-Y", "other", "r3.txt", "text/plain", "r3")

	req := httptest.NewRequest(http.MethodGet, "/blobs/patient/patient-X", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp listResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error unmarshaling: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected Total=2, got %d", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestBlobHandler_Search(t *testing.T) {
	store := NewInMemoryBlobStore()
	handler := NewBlobHandler(store)
	e := echo.New()
	g := e.Group("")
	handler.RegisterRoutes(g)

	seedBlob(t, store, "p1", "lab-report", "search1.pdf", "application/pdf", "s1")
	seedBlob(t, store, "p1", "other", "search2.txt", "text/plain", "s2")
	seedBlob(t, store, "p2", "lab-report", "search3.pdf", "application/pdf", "s3")

	req := httptest.NewRequest(http.MethodGet, "/blobs?patient_id=p1&category=lab-report", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp listResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error unmarshaling: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected Total=1, got %d", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}
