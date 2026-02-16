package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// MetaStore defines persistence operations for resource-level meta management.
type MetaStore interface {
	GetMeta(ctx context.Context, resourceType, resourceID string) (*Meta, error)
	AddMeta(ctx context.Context, resourceType, resourceID string, meta *Meta) (*Meta, error)
	DeleteMeta(ctx context.Context, resourceType, resourceID string, meta *Meta) (*Meta, error)
}

// InMemoryMetaStore is a thread-safe in-memory implementation of MetaStore.
type InMemoryMetaStore struct {
	mu   sync.RWMutex
	data map[string]*Meta // key: "resourceType/resourceID"
}

// NewInMemoryMetaStore creates a new InMemoryMetaStore.
func NewInMemoryMetaStore() *InMemoryMetaStore {
	return &InMemoryMetaStore{
		data: make(map[string]*Meta),
	}
}

func metaKey(resourceType, resourceID string) string {
	return resourceType + "/" + resourceID
}

// GetMeta returns the meta for a resource, or an empty Meta if none exists.
func (s *InMemoryMetaStore) GetMeta(_ context.Context, resourceType, resourceID string) (*Meta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := metaKey(resourceType, resourceID)
	if m, ok := s.data[key]; ok {
		return cloneMeta(m), nil
	}
	return &Meta{}, nil
}

// AddMeta merges the provided meta into the existing meta for a resource.
// Profiles, security labels, and tags are added if not already present.
func (s *InMemoryMetaStore) AddMeta(_ context.Context, resourceType, resourceID string, meta *Meta) (*Meta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := metaKey(resourceType, resourceID)
	existing, ok := s.data[key]
	if !ok {
		existing = &Meta{}
		s.data[key] = existing
	}

	existing.Profile = mergeStrings(existing.Profile, meta.Profile)
	existing.Security = mergeCodings(existing.Security, meta.Security)
	existing.Tag = mergeCodings(existing.Tag, meta.Tag)

	return cloneMeta(existing), nil
}

// DeleteMeta removes the specified profiles, security labels, and tags from a resource's meta.
func (s *InMemoryMetaStore) DeleteMeta(_ context.Context, resourceType, resourceID string, meta *Meta) (*Meta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := metaKey(resourceType, resourceID)
	existing, ok := s.data[key]
	if !ok {
		return &Meta{}, nil
	}

	existing.Profile = removeStrings(existing.Profile, meta.Profile)
	existing.Security = removeCodings(existing.Security, meta.Security)
	existing.Tag = removeCodings(existing.Tag, meta.Tag)

	return cloneMeta(existing), nil
}

// mergeStrings adds values from src to dst if not already present.
func mergeStrings(dst, src []string) []string {
	set := make(map[string]bool, len(dst))
	for _, v := range dst {
		set[v] = true
	}
	for _, v := range src {
		if !set[v] {
			dst = append(dst, v)
			set[v] = true
		}
	}
	return dst
}

// mergeCodings adds codings from src to dst if no coding with the same system+code exists.
func mergeCodings(dst, src []Coding) []Coding {
	type key struct{ system, code string }
	set := make(map[key]bool, len(dst))
	for _, c := range dst {
		set[key{c.System, c.Code}] = true
	}
	for _, c := range src {
		k := key{c.System, c.Code}
		if !set[k] {
			dst = append(dst, c)
			set[k] = true
		}
	}
	return dst
}

// removeStrings removes values in toRemove from src.
func removeStrings(src, toRemove []string) []string {
	set := make(map[string]bool, len(toRemove))
	for _, v := range toRemove {
		set[v] = true
	}
	result := make([]string, 0, len(src))
	for _, v := range src {
		if !set[v] {
			result = append(result, v)
		}
	}
	return result
}

// removeCodings removes codings from src that match system+code in toRemove.
func removeCodings(src, toRemove []Coding) []Coding {
	type key struct{ system, code string }
	set := make(map[key]bool, len(toRemove))
	for _, c := range toRemove {
		set[key{c.System, c.Code}] = true
	}
	result := make([]Coding, 0, len(src))
	for _, c := range src {
		if !set[key{c.System, c.Code}] {
			result = append(result, c)
		}
	}
	return result
}

// cloneMeta returns a deep copy of a Meta value.
func cloneMeta(m *Meta) *Meta {
	c := &Meta{
		VersionID:   m.VersionID,
		LastUpdated: m.LastUpdated,
	}
	if len(m.Profile) > 0 {
		c.Profile = make([]string, len(m.Profile))
		copy(c.Profile, m.Profile)
	}
	if len(m.Security) > 0 {
		c.Security = make([]Coding, len(m.Security))
		copy(c.Security, m.Security)
	}
	if len(m.Tag) > 0 {
		c.Tag = make([]Coding, len(m.Tag))
		copy(c.Tag, m.Tag)
	}
	return c
}

// MetaHandler handles FHIR $meta, $meta-add, and $meta-delete operations.
type MetaHandler struct {
	store MetaStore
}

// NewMetaHandler creates a new MetaHandler.
func NewMetaHandler(store MetaStore) *MetaHandler {
	return &MetaHandler{store: store}
}

// RegisterRoutes registers the $meta, $meta-add, and $meta-delete endpoints.
func (h *MetaHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/:resourceType/:id/$meta", h.GetMeta)
	fhirGroup.POST("/:resourceType/:id/$meta", h.GetMeta)
	fhirGroup.POST("/:resourceType/:id/$meta-add", h.AddMeta)
	fhirGroup.POST("/:resourceType/:id/$meta-delete", h.DeleteMeta)
}

// GetMeta handles GET/POST /fhir/:resourceType/:id/$meta
func (h *MetaHandler) GetMeta(c echo.Context) error {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("id")

	if resourceType == "" || resourceID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType and id are required"))
	}

	meta, err := h.store.GetMeta(c.Request().Context(), resourceType, resourceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, metaToParameters(meta))
}

// AddMeta handles POST /fhir/:resourceType/:id/$meta-add
func (h *MetaHandler) AddMeta(c echo.Context) error {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("id")

	if resourceType == "" || resourceID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType and id are required"))
	}

	inputMeta, err := parseMetaFromParameters(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid Parameters resource: %v", err)))
	}

	result, err := h.store.AddMeta(c.Request().Context(), resourceType, resourceID, inputMeta)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, metaToParameters(result))
}

// DeleteMeta handles POST /fhir/:resourceType/:id/$meta-delete
func (h *MetaHandler) DeleteMeta(c echo.Context) error {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("id")

	if resourceType == "" || resourceID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("resourceType and id are required"))
	}

	inputMeta, err := parseMetaFromParameters(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid Parameters resource: %v", err)))
	}

	result, err := h.store.DeleteMeta(c.Request().Context(), resourceType, resourceID, inputMeta)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, metaToParameters(result))
}

// metaToParameters converts a Meta to a FHIR Parameters resource containing the meta.
func metaToParameters(meta *Meta) map[string]interface{} {
	metaMap := map[string]interface{}{}

	if meta.VersionID != "" {
		metaMap["versionId"] = meta.VersionID
	}
	if !meta.LastUpdated.IsZero() {
		metaMap["lastUpdated"] = meta.LastUpdated
	}
	if len(meta.Profile) > 0 {
		metaMap["profile"] = meta.Profile
	}
	if len(meta.Security) > 0 {
		secList := make([]map[string]interface{}, 0, len(meta.Security))
		for _, s := range meta.Security {
			entry := map[string]interface{}{}
			if s.System != "" {
				entry["system"] = s.System
			}
			if s.Code != "" {
				entry["code"] = s.Code
			}
			if s.Display != "" {
				entry["display"] = s.Display
			}
			secList = append(secList, entry)
		}
		metaMap["security"] = secList
	}
	if len(meta.Tag) > 0 {
		tagList := make([]map[string]interface{}, 0, len(meta.Tag))
		for _, t := range meta.Tag {
			entry := map[string]interface{}{}
			if t.System != "" {
				entry["system"] = t.System
			}
			if t.Code != "" {
				entry["code"] = t.Code
			}
			if t.Display != "" {
				entry["display"] = t.Display
			}
			tagList = append(tagList, entry)
		}
		metaMap["tag"] = tagList
	}

	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []map[string]interface{}{
			{
				"name":      "return",
				"valueMeta": metaMap,
			},
		},
	}
}

// parseMetaFromParameters extracts a Meta from a FHIR Parameters resource in the request body.
func parseMetaFromParameters(c echo.Context) (*Meta, error) {
	var body map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %w", err)
	}

	params, ok := body["parameter"]
	if !ok {
		return nil, fmt.Errorf("missing 'parameter' field")
	}

	paramList, ok := params.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'parameter' must be an array")
	}

	meta := &Meta{}

	for _, p := range paramList {
		param, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := param["name"].(string)
		if name != "meta" && name != "return" {
			continue
		}

		valueMeta, ok := param["valueMeta"].(map[string]interface{})
		if !ok {
			continue
		}

		if profiles, ok := valueMeta["profile"].([]interface{}); ok {
			for _, pr := range profiles {
				if s, ok := pr.(string); ok {
					meta.Profile = append(meta.Profile, s)
				}
			}
		}

		if securityList, ok := valueMeta["security"].([]interface{}); ok {
			for _, s := range securityList {
				if sm, ok := s.(map[string]interface{}); ok {
					coding := Coding{}
					if v, ok := sm["system"].(string); ok {
						coding.System = v
					}
					if v, ok := sm["code"].(string); ok {
						coding.Code = v
					}
					if v, ok := sm["display"].(string); ok {
						coding.Display = v
					}
					meta.Security = append(meta.Security, coding)
				}
			}
		}

		if tagList, ok := valueMeta["tag"].([]interface{}); ok {
			for _, t := range tagList {
				if tm, ok := t.(map[string]interface{}); ok {
					coding := Coding{}
					if v, ok := tm["system"].(string); ok {
						coding.System = v
					}
					if v, ok := tm["code"].(string); ok {
						coding.Code = v
					}
					if v, ok := tm["display"].(string); ok {
						coding.Display = v
					}
					meta.Tag = append(meta.Tag, coding)
				}
			}
		}
	}

	return meta, nil
}
