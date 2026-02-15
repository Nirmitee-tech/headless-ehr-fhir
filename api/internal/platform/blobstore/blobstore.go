// Package blobstore provides document/blob storage for the EHR platform.
// It defines the BlobStore interface, an in-memory implementation suitable for
// testing and development, and Echo HTTP handlers for multipart upload,
// download, metadata retrieval, deletion, and search.
package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

var (
	ErrBlobNotFound       = errors.New("blob not found")
	ErrFileTooLarge       = errors.New("file exceeds maximum allowed size")
	ErrInvalidContentType = errors.New("content type is not allowed")
	ErrMissingFileName    = errors.New("file name is required")
)

// ---------------------------------------------------------------------------
// Validation constants
// ---------------------------------------------------------------------------

// MaxFileSize is the maximum allowed blob size in bytes (100 MB).
const MaxFileSize = 100 * 1024 * 1024

// AllowedCategories lists valid blob category values.
var AllowedCategories = map[string]bool{
	"clinical-image": true,
	"lab-report":     true,
	"consent-form":   true,
	"radiology":      true,
	"pathology":      true,
	"other":          true,
}

// AllowedContentTypes lists common medical file MIME types.
var AllowedContentTypes = map[string]bool{
	"image/png":              true,
	"image/jpeg":             true,
	"image/dicom":            true,
	"application/pdf":        true,
	"application/dicom":      true,
	"text/plain":             true,
	"text/html":              true,
	"application/hl7-v2":     true,
	"application/fhir+json":  true,
	"application/fhir+xml":   true,
}

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// BlobMetadata describes a stored blob.
type BlobMetadata struct {
	ID          string            `json:"id"`
	FileName    string            `json:"file_name"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	PatientID   string            `json:"patient_id,omitempty"`
	EncounterID string            `json:"encounter_id,omitempty"`
	Category    string            `json:"category"`
	Hash        string            `json:"hash"`
	CreatedAt   time.Time         `json:"created_at"`
	CreatedBy   string            `json:"created_by"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// SearchParams specifies search/filter criteria for blobs.
type SearchParams struct {
	PatientID     string
	Category      string
	ContentType   string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	FileName      string // partial match
	Tags          map[string]string
	Limit         int
	Offset        int
}

// ---------------------------------------------------------------------------
// BlobStore interface
// ---------------------------------------------------------------------------

// BlobStore defines the contract for blob storage backends.
type BlobStore interface {
	Upload(ctx context.Context, meta BlobMetadata, content io.Reader) (*BlobMetadata, error)
	Download(ctx context.Context, id string) (io.ReadCloser, *BlobMetadata, error)
	Delete(ctx context.Context, id string) error
	GetMetadata(ctx context.Context, id string) (*BlobMetadata, error)
	ListByPatient(ctx context.Context, patientID string, category string, limit, offset int) ([]*BlobMetadata, int, error)
	Search(ctx context.Context, params SearchParams) ([]*BlobMetadata, int, error)
}

// ---------------------------------------------------------------------------
// In-memory implementation
// ---------------------------------------------------------------------------

type storedBlob struct {
	metadata BlobMetadata
	content  []byte
}

// InMemoryBlobStore is a thread-safe, in-memory BlobStore for testing/dev.
type InMemoryBlobStore struct {
	mu    sync.RWMutex
	blobs map[string]*storedBlob
}

// NewInMemoryBlobStore returns a ready-to-use InMemoryBlobStore.
func NewInMemoryBlobStore() *InMemoryBlobStore {
	return &InMemoryBlobStore{
		blobs: make(map[string]*storedBlob),
	}
}

// Upload validates inputs, reads the content, computes a SHA-256 hash, and
// stores the blob in memory.
func (s *InMemoryBlobStore) Upload(_ context.Context, meta BlobMetadata, content io.Reader) (*BlobMetadata, error) {
	if meta.FileName == "" {
		return nil, ErrMissingFileName
	}

	// Read content into memory so we can measure size and compute hash.
	data, err := io.ReadAll(io.LimitReader(content, MaxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}
	if int64(len(data)) > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Compute SHA-256.
	h := sha256.Sum256(data)

	meta.ID = uuid.New().String()
	meta.Size = int64(len(data))
	meta.Hash = fmt.Sprintf("%x", h)
	meta.CreatedAt = time.Now().UTC()

	if meta.Tags == nil {
		meta.Tags = make(map[string]string)
	}

	s.mu.Lock()
	s.blobs[meta.ID] = &storedBlob{
		metadata: meta,
		content:  data,
	}
	s.mu.Unlock()

	out := meta // copy
	return &out, nil
}

// Download returns an io.ReadCloser over the blob content and its metadata.
func (s *InMemoryBlobStore) Download(_ context.Context, id string) (io.ReadCloser, *BlobMetadata, error) {
	s.mu.RLock()
	blob, ok := s.blobs[id]
	s.mu.RUnlock()

	if !ok {
		return nil, nil, ErrBlobNotFound
	}

	meta := blob.metadata // copy
	return io.NopCloser(bytes.NewReader(blob.content)), &meta, nil
}

// Delete removes a blob by ID.
func (s *InMemoryBlobStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.blobs[id]; !ok {
		return ErrBlobNotFound
	}
	delete(s.blobs, id)
	return nil
}

// GetMetadata returns blob metadata without content.
func (s *InMemoryBlobStore) GetMetadata(_ context.Context, id string) (*BlobMetadata, error) {
	s.mu.RLock()
	blob, ok := s.blobs[id]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrBlobNotFound
	}

	meta := blob.metadata // copy
	return &meta, nil
}

// ListByPatient returns blobs for a given patient, optionally filtered by
// category. It returns the matching page and the total count.
func (s *InMemoryBlobStore) ListByPatient(_ context.Context, patientID, category string, limit, offset int) ([]*BlobMetadata, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []*BlobMetadata
	for _, b := range s.blobs {
		if b.metadata.PatientID != patientID {
			continue
		}
		if category != "" && b.metadata.Category != category {
			continue
		}
		m := b.metadata // copy
		matched = append(matched, &m)
	}

	total := len(matched)
	if limit <= 0 {
		limit = 20
	}
	if offset > len(matched) {
		offset = len(matched)
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

// Search returns blobs matching the given search parameters.
func (s *InMemoryBlobStore) Search(_ context.Context, params SearchParams) ([]*BlobMetadata, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []*BlobMetadata
	for _, b := range s.blobs {
		if !matchesSearch(&b.metadata, params) {
			continue
		}
		m := b.metadata // copy
		matched = append(matched, &m)
	}

	total := len(matched)
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset > len(matched) {
		offset = len(matched)
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func matchesSearch(m *BlobMetadata, p SearchParams) bool {
	if p.PatientID != "" && m.PatientID != p.PatientID {
		return false
	}
	if p.Category != "" && m.Category != p.Category {
		return false
	}
	if p.ContentType != "" && m.ContentType != p.ContentType {
		return false
	}
	if p.CreatedAfter != nil && m.CreatedAt.Before(*p.CreatedAfter) {
		return false
	}
	if p.CreatedBefore != nil && m.CreatedAt.After(*p.CreatedBefore) {
		return false
	}
	if p.FileName != "" && !strings.Contains(strings.ToLower(m.FileName), strings.ToLower(p.FileName)) {
		return false
	}
	if len(p.Tags) > 0 {
		for k, v := range p.Tags {
			if mv, ok := m.Tags[k]; !ok || mv != v {
				return false
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// HTTP handler
// ---------------------------------------------------------------------------

// listResponse is the JSON envelope returned by list/search endpoints.
type listResponse struct {
	Items []*BlobMetadata `json:"items"`
	Total int             `json:"total"`
}

// BlobHandler provides Echo HTTP handlers for blob operations.
type BlobHandler struct {
	store BlobStore
}

// NewBlobHandler creates a new BlobHandler.
func NewBlobHandler(store BlobStore) *BlobHandler {
	return &BlobHandler{store: store}
}

// RegisterRoutes mounts blob routes on the supplied Echo group.
func (h *BlobHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/blobs/upload", h.handleUpload)
	g.GET("/blobs/patient/:patientId", h.handleListByPatient)
	g.GET("/blobs/:id/metadata", h.handleGetMetadata)
	g.GET("/blobs/:id", h.handleDownload)
	g.DELETE("/blobs/:id", h.handleDelete)
	g.GET("/blobs", h.handleSearch)
}

func (h *BlobHandler) handleUpload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open uploaded file"})
	}
	defer src.Close()

	// Detect content type from file header or form field.
	contentType := file.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = "application/octet-stream"
	}

	meta := BlobMetadata{
		FileName:    file.Filename,
		ContentType: contentType,
		PatientID:   c.FormValue("patient_id"),
		EncounterID: c.FormValue("encounter_id"),
		Category:    c.FormValue("category"),
		CreatedBy:   c.FormValue("created_by"),
	}

	result, err := h.store.Upload(c.Request().Context(), meta, src)
	if err != nil {
		switch {
		case errors.Is(err, ErrFileTooLarge):
			return c.JSON(http.StatusRequestEntityTooLarge, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrMissingFileName):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrInvalidContentType):
			return c.JSON(http.StatusUnsupportedMediaType, map[string]string{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *BlobHandler) handleDownload(c echo.Context) error {
	id := c.Param("id")

	rc, meta, err := h.store.Download(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer rc.Close()

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, meta.FileName))
	return c.Stream(http.StatusOK, meta.ContentType, rc)
}

func (h *BlobHandler) handleGetMetadata(c echo.Context) error {
	id := c.Param("id")

	meta, err := h.store.GetMetadata(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, meta)
}

func (h *BlobHandler) handleDelete(c echo.Context) error {
	id := c.Param("id")

	err := h.store.Delete(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *BlobHandler) handleListByPatient(c echo.Context) error {
	patientID := c.Param("patientId")
	category := c.QueryParam("category")
	limit := intParam(c, "limit", 20)
	offset := intParam(c, "offset", 0)

	items, total, err := h.store.ListByPatient(c.Request().Context(), patientID, category, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if items == nil {
		items = []*BlobMetadata{}
	}

	return c.JSON(http.StatusOK, listResponse{Items: items, Total: total})
}

func (h *BlobHandler) handleSearch(c echo.Context) error {
	params := SearchParams{
		PatientID:   c.QueryParam("patient_id"),
		Category:    c.QueryParam("category"),
		ContentType: c.QueryParam("content_type"),
		FileName:    c.QueryParam("file_name"),
		Limit:       intParam(c, "limit", 20),
		Offset:      intParam(c, "offset", 0),
	}

	items, total, err := h.store.Search(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if items == nil {
		items = []*BlobMetadata{}
	}

	return c.JSON(http.StatusOK, listResponse{Items: items, Total: total})
}

func intParam(c echo.Context, name string, defaultVal int) int {
	v := c.QueryParam(name)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}
