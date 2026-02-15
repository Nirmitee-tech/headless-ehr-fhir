// Package fhir provides FHIR R4 Binary resource support.
//
// The Binary resource is used to store opaque content (images, PDFs, CDA
// documents, etc.) within the FHIR ecosystem.  Per the FHIR R4 specification
// (https://hl7.org/fhir/R4/binary.html), reading a Binary with an Accept
// header matching the resource's own contentType returns the raw bytes;
// otherwise the resource is wrapped in the standard FHIR JSON envelope with
// the payload base64-encoded in the "data" element.
package fhir

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Constants and sentinel errors
// ---------------------------------------------------------------------------

// MaxBinarySize is the maximum allowed Binary resource payload in bytes (50 MB).
const MaxBinarySize = 50 * 1024 * 1024

var (
	// ErrBinaryNotFound is returned when a Binary resource cannot be found.
	ErrBinaryNotFound = errors.New("Binary resource not found")

	// ErrBinaryMissingContentType is returned when contentType is empty.
	ErrBinaryMissingContentType = errors.New("Binary resource requires a contentType")

	// ErrBinaryMissingData is returned when data is empty on create.
	ErrBinaryMissingData = errors.New("Binary resource requires data")

	// ErrBinaryTooLarge is returned when the payload exceeds MaxBinarySize.
	ErrBinaryTooLarge = errors.New("Binary resource data exceeds maximum allowed size")
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// BinaryResource represents a FHIR R4 Binary resource in its internal
// (decoded) form.  The Data field holds raw bytes; base64 encoding is
// performed only at the FHIR serialisation boundary.
type BinaryResource struct {
	ID              string     `json:"id"`
	ContentType     string     `json:"contentType"`
	Data            []byte     `json:"-"` // never serialise raw bytes directly
	SecurityContext *Reference `json:"securityContext,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// BinaryFHIRJSON is the wire representation of a FHIR R4 Binary resource.
// The data field is base64-encoded per the specification.
type BinaryFHIRJSON struct {
	ResourceType    string     `json:"resourceType"`
	ID              string     `json:"id,omitempty"`
	ContentType     string     `json:"contentType"`
	Data            string     `json:"data,omitempty"`
	SecurityContext *Reference `json:"securityContext,omitempty"`
}

// ToFHIRJSON converts the internal BinaryResource to its FHIR JSON wire
// representation, base64-encoding the Data payload.
func (b *BinaryResource) ToFHIRJSON() *BinaryFHIRJSON {
	fj := &BinaryFHIRJSON{
		ResourceType: "Binary",
		ID:           b.ID,
		ContentType:  b.ContentType,
	}
	if len(b.Data) > 0 {
		fj.Data = base64.StdEncoding.EncodeToString(b.Data)
	}
	if b.SecurityContext != nil {
		sc := *b.SecurityContext
		fj.SecurityContext = &sc
	}
	return fj
}

// ToBinaryResource converts a FHIR JSON wire representation back to the
// internal BinaryResource, decoding the base64 payload.
func (fj *BinaryFHIRJSON) ToBinaryResource() (*BinaryResource, error) {
	var data []byte
	if fj.Data != "" {
		var err error
		data, err = base64.StdEncoding.DecodeString(fj.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 in Binary.data: %w", err)
		}
	}
	res := &BinaryResource{
		ID:          fj.ID,
		ContentType: fj.ContentType,
		Data:        data,
	}
	if fj.SecurityContext != nil {
		sc := *fj.SecurityContext
		res.SecurityContext = &sc
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// BinaryStore interface
// ---------------------------------------------------------------------------

// BinaryStore defines the contract for Binary resource persistence.
type BinaryStore interface {
	Create(ctx context.Context, res *BinaryResource) (*BinaryResource, error)
	GetByID(ctx context.Context, id string) (*BinaryResource, error)
	Update(ctx context.Context, res *BinaryResource) (*BinaryResource, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*BinaryResource, int, error)
}

// ---------------------------------------------------------------------------
// In-memory implementation
// ---------------------------------------------------------------------------

type storedBinary struct {
	resource BinaryResource
	data     []byte // separate copy for isolation
}

// InMemoryBinaryStore is a thread-safe, in-memory BinaryStore suitable for
// testing and development environments.
type InMemoryBinaryStore struct {
	mu      sync.RWMutex
	entries map[string]*storedBinary
	// order preserves insertion order for deterministic listing.
	order []string
}

// NewInMemoryBinaryStore returns a ready-to-use InMemoryBinaryStore.
func NewInMemoryBinaryStore() *InMemoryBinaryStore {
	return &InMemoryBinaryStore{
		entries: make(map[string]*storedBinary),
	}
}

func (s *InMemoryBinaryStore) validate(res *BinaryResource, requireData bool) error {
	if res.ContentType == "" {
		return ErrBinaryMissingContentType
	}
	if requireData && len(res.Data) == 0 {
		return ErrBinaryMissingData
	}
	if len(res.Data) > MaxBinarySize {
		return ErrBinaryTooLarge
	}
	return nil
}

// copyResource returns a deep copy of a BinaryResource so callers cannot
// mutate store internals.
func copyResource(src *storedBinary) *BinaryResource {
	out := src.resource
	out.Data = make([]byte, len(src.data))
	copy(out.Data, src.data)
	if src.resource.SecurityContext != nil {
		sc := *src.resource.SecurityContext
		out.SecurityContext = &sc
	}
	return &out
}

// Create validates and persists a new BinaryResource, assigning it a UUID and
// timestamps.
func (s *InMemoryBinaryStore) Create(_ context.Context, res *BinaryResource) (*BinaryResource, error) {
	if err := s.validate(res, true); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	entry := &storedBinary{
		resource: BinaryResource{
			ID:          uuid.New().String(),
			ContentType: res.ContentType,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		data: make([]byte, len(res.Data)),
	}
	copy(entry.data, res.Data)
	if res.SecurityContext != nil {
		sc := *res.SecurityContext
		entry.resource.SecurityContext = &sc
	}

	s.mu.Lock()
	s.entries[entry.resource.ID] = entry
	s.order = append(s.order, entry.resource.ID)
	s.mu.Unlock()

	return copyResource(entry), nil
}

// GetByID retrieves a Binary resource by its ID.
func (s *InMemoryBinaryStore) GetByID(_ context.Context, id string) (*BinaryResource, error) {
	s.mu.RLock()
	entry, ok := s.entries[id]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrBinaryNotFound
	}
	return copyResource(entry), nil
}

// Update replaces an existing Binary resource's mutable fields (contentType,
// data, securityContext) and bumps the UpdatedAt timestamp.
func (s *InMemoryBinaryStore) Update(_ context.Context, res *BinaryResource) (*BinaryResource, error) {
	if err := s.validate(res, false); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[res.ID]
	if !ok {
		return nil, ErrBinaryNotFound
	}

	entry.resource.ContentType = res.ContentType
	entry.resource.UpdatedAt = time.Now().UTC()

	if len(res.Data) > 0 {
		entry.data = make([]byte, len(res.Data))
		copy(entry.data, res.Data)
	}

	if res.SecurityContext != nil {
		sc := *res.SecurityContext
		entry.resource.SecurityContext = &sc
	} else {
		entry.resource.SecurityContext = nil
	}

	return copyResource(entry), nil
}

// Delete removes a Binary resource by ID.
func (s *InMemoryBinaryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.entries[id]; !ok {
		return ErrBinaryNotFound
	}
	delete(s.entries, id)

	// Remove from order slice.
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return nil
}

// List returns a page of Binary resources in insertion order.
func (s *InMemoryBinaryStore) List(_ context.Context, limit, offset int) ([]*BinaryResource, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.order)

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	ids := s.order[offset:end]
	results := make([]*BinaryResource, 0, len(ids))
	for _, id := range ids {
		if entry, ok := s.entries[id]; ok {
			results = append(results, copyResource(entry))
		}
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// HTTP handler
// ---------------------------------------------------------------------------

// BinaryHandler provides Echo HTTP handlers for FHIR Binary CRUD operations
// with FHIR R4 content-negotiation semantics.
type BinaryHandler struct {
	store BinaryStore
}

// NewBinaryHandler creates a new BinaryHandler backed by the given store.
func NewBinaryHandler(store BinaryStore) *BinaryHandler {
	return &BinaryHandler{store: store}
}

// RegisterRoutes mounts Binary CRUD routes on the supplied FHIR Echo group.
// The group is expected to be mounted at /fhir.
func (h *BinaryHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/Binary", h.handleList)
	g.POST("/Binary", h.handleCreate)
	g.GET("/Binary/:id", h.handleRead)
	g.PUT("/Binary/:id", h.handleUpdate)
	g.DELETE("/Binary/:id", h.handleDelete)
}

// handleCreate processes POST /fhir/Binary.
func (h *BinaryHandler) handleCreate(c echo.Context) error {
	var fj BinaryFHIRJSON
	if err := json.NewDecoder(c.Request().Body).Decode(&fj); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	if fj.ResourceType != "" && fj.ResourceType != "Binary" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType must be Binary"))
	}

	res, err := fj.ToBinaryResource()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	created, err := h.store.Create(c.Request().Context(), res)
	if err != nil {
		return h.mapStoreError(c, err)
	}

	c.Response().Header().Set("Location", "Binary/"+created.ID)
	return c.JSON(http.StatusCreated, created.ToFHIRJSON())
}

// handleRead processes GET /fhir/Binary/:id with content negotiation.
func (h *BinaryHandler) handleRead(c echo.Context) error {
	id := c.Param("id")

	res, err := h.store.GetByID(c.Request().Context(), id)
	if err != nil {
		return h.mapStoreError(c, err)
	}

	// Content negotiation per FHIR R4 Binary spec:
	// If the Accept header matches the resource's contentType, return raw bytes.
	// Otherwise return the FHIR JSON representation.
	if h.wantsRawContent(c, res.ContentType) {
		return c.Blob(http.StatusOK, res.ContentType, res.Data)
	}

	// If the client explicitly asked for application/fhir+json, honour that
	// media type in the response Content-Type header.
	if h.wantsFHIRContentType(c) {
		c.Response().Header().Set("Content-Type", "application/fhir+json; charset=UTF-8")
		c.Response().WriteHeader(http.StatusOK)
		return json.NewEncoder(c.Response()).Encode(res.ToFHIRJSON())
	}

	return c.JSON(http.StatusOK, res.ToFHIRJSON())
}

// wantsRawContent returns true when the client's Accept header indicates it
// wants the Binary's native content type rather than the FHIR JSON envelope.
func (h *BinaryHandler) wantsRawContent(c echo.Context, resourceContentType string) bool {
	accept := c.Request().Header.Get("Accept")
	if accept == "" {
		return false
	}

	// Parse the Accept header into individual media types.
	for _, part := range strings.Split(accept, ",") {
		mt := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		switch mt {
		case resourceContentType:
			return true
		case "*/*", "application/json", "application/fhir+json":
			return false
		}
	}
	return false
}

// wantsFHIRContentType returns true when the client's Accept header explicitly
// requests the application/fhir+json media type.
func (h *BinaryHandler) wantsFHIRContentType(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	for _, part := range strings.Split(accept, ",") {
		mt := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if mt == "application/fhir+json" {
			return true
		}
	}
	return false
}

// handleUpdate processes PUT /fhir/Binary/:id.
func (h *BinaryHandler) handleUpdate(c echo.Context) error {
	id := c.Param("id")

	var fj BinaryFHIRJSON
	if err := json.NewDecoder(c.Request().Body).Decode(&fj); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	if fj.ResourceType != "" && fj.ResourceType != "Binary" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType must be Binary"))
	}

	// If the body specifies an ID it must match the URL.
	if fj.ID != "" && fj.ID != id {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(
			fmt.Sprintf("resource id in body (%s) does not match URL (%s)", fj.ID, id)))
	}
	fj.ID = id

	res, err := fj.ToBinaryResource()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	updated, err := h.store.Update(c.Request().Context(), res)
	if err != nil {
		return h.mapStoreError(c, err)
	}

	return c.JSON(http.StatusOK, updated.ToFHIRJSON())
}

// handleDelete processes DELETE /fhir/Binary/:id.
func (h *BinaryHandler) handleDelete(c echo.Context) error {
	id := c.Param("id")

	if err := h.store.Delete(c.Request().Context(), id); err != nil {
		return h.mapStoreError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// handleList processes GET /fhir/Binary and returns a FHIR Bundle searchset.
func (h *BinaryHandler) handleList(c echo.Context) error {
	limit := intQueryParam(c, "_count", 20)
	offset := intQueryParam(c, "_offset", 0)

	items, total, err := h.store.List(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	// Sort by ID for deterministic output (store already returns insertion
	// order but we sort by ID for API consistency).
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	entries := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		entries = append(entries, map[string]interface{}{
			"fullUrl":  "Binary/" + item.ID,
			"resource": item.ToFHIRJSON(),
		})
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        total,
		"entry":        entries,
	}

	return c.JSON(http.StatusOK, bundle)
}

// mapStoreError translates store-level errors to appropriate HTTP responses.
func (h *BinaryHandler) mapStoreError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrBinaryNotFound):
		return c.JSON(http.StatusNotFound, NotFoundOutcome("Binary", ""))
	case errors.Is(err, ErrBinaryMissingContentType):
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	case errors.Is(err, ErrBinaryMissingData):
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	case errors.Is(err, ErrBinaryTooLarge):
		return c.JSON(http.StatusRequestEntityTooLarge, ErrorOutcome(err.Error()))
	default:
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}
}

// intQueryParam reads an integer query parameter with a default fallback.
func intQueryParam(c echo.Context, name string, defaultVal int) int {
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
