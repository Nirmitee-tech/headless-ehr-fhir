package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ResourceResolver fetches FHIR resources by reference.
type ResourceResolver interface {
	ResolveReference(ctx context.Context, reference string) (map[string]interface{}, error)
}

// DocumentGenerator creates FHIR Document Bundles from Compositions.
type DocumentGenerator struct {
	resolver ResourceResolver
}

// NewDocumentGenerator creates a new document generator.
func NewDocumentGenerator(resolver ResourceResolver) *DocumentGenerator {
	return &DocumentGenerator{resolver: resolver}
}

// GenerateDocument generates a FHIR Document Bundle from a Composition resource.
// The Composition is validated, then all referenced resources are resolved and
// assembled into a Bundle of type "document" per the FHIR specification.
func (g *DocumentGenerator) GenerateDocument(ctx context.Context, composition map[string]interface{}, persist bool) (map[string]interface{}, error) {
	if composition == nil {
		return nil, fmt.Errorf("composition is nil")
	}

	// Validate the Composition resource.
	rt, _ := composition["resourceType"].(string)
	if rt != "Composition" {
		return nil, fmt.Errorf("resource is not a Composition (resourceType=%q)", rt)
	}

	requiredFields := []string{"status", "type", "date", "author", "title"}
	for _, field := range requiredFields {
		if _, ok := composition[field]; !ok {
			return nil, fmt.Errorf("Composition missing required field %q", field)
		}
	}

	// Build the Document Bundle.
	now := time.Now().UTC().Format(time.RFC3339)
	bundleID := uuid.New().String()

	// First entry is always the Composition.
	compositionRaw, err := json.Marshal(composition)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Composition: %w", err)
	}

	compID, _ := composition["id"].(string)
	compFullURL := "Composition/" + compID

	entries := []interface{}{
		map[string]interface{}{
			"fullUrl":  compFullURL,
			"resource": json.RawMessage(compositionRaw),
		},
	}

	// Collect and deduplicate references.
	refs := collectReferences(composition)

	// Track which references we have already added to avoid duplicates.
	seen := make(map[string]bool)
	seen[compFullURL] = true

	for _, ref := range refs {
		if seen[ref] {
			continue
		}
		seen[ref] = true

		resource, err := g.resolver.ResolveReference(ctx, ref)
		if err != nil || resource == nil {
			// Skip unresolvable references.
			continue
		}

		raw, err := json.Marshal(resource)
		if err != nil {
			continue
		}

		entries = append(entries, map[string]interface{}{
			"fullUrl":  ref,
			"resource": json.RawMessage(raw),
		})
	}

	total := len(entries)

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "document",
		"identifier": map[string]interface{}{
			"system": "urn:ietf:rfc:3986",
			"value":  "urn:uuid:" + bundleID,
		},
		"timestamp": now,
		"total":     total,
		"entry":     entries,
	}

	return bundle, nil
}

// collectReferences walks a Composition and extracts all unique reference strings.
// It handles subject, author, custodian, encounter, attester[].party,
// and section[].entry[] / section[].author[] references, including nested sections.
func collectReferences(composition map[string]interface{}) []string {
	seen := make(map[string]bool)
	var refs []string

	addRef := func(ref string) {
		if ref != "" && !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
	}

	// Extract a single reference from a reference object.
	extractRef := func(obj interface{}) string {
		m, ok := obj.(map[string]interface{})
		if !ok {
			return ""
		}
		ref, _ := m["reference"].(string)
		return ref
	}

	// Extract references from an array of reference objects.
	extractArrayRefs := func(arr interface{}) {
		slice, ok := arr.([]interface{})
		if !ok {
			return
		}
		for _, item := range slice {
			if ref := extractRef(item); ref != "" {
				addRef(ref)
			}
		}
	}

	// subject
	if subj, ok := composition["subject"]; ok {
		if ref := extractRef(subj); ref != "" {
			addRef(ref)
		}
	}

	// author (array)
	if authors, ok := composition["author"]; ok {
		extractArrayRefs(authors)
	}

	// custodian
	if cust, ok := composition["custodian"]; ok {
		if ref := extractRef(cust); ref != "" {
			addRef(ref)
		}
	}

	// encounter
	if enc, ok := composition["encounter"]; ok {
		if ref := extractRef(enc); ref != "" {
			addRef(ref)
		}
	}

	// attester[].party
	if attesters, ok := composition["attester"]; ok {
		if arr, ok := attesters.([]interface{}); ok {
			for _, att := range arr {
				if attMap, ok := att.(map[string]interface{}); ok {
					if party, ok := attMap["party"]; ok {
						if ref := extractRef(party); ref != "" {
							addRef(ref)
						}
					}
				}
			}
		}
	}

	// sections (recursive)
	var walkSections func(sections []interface{})
	walkSections = func(sections []interface{}) {
		for _, sec := range sections {
			secMap, ok := sec.(map[string]interface{})
			if !ok {
				continue
			}

			// section.entry[] references
			if entries, ok := secMap["entry"]; ok {
				extractArrayRefs(entries)
			}

			// section.author[] references
			if authors, ok := secMap["author"]; ok {
				extractArrayRefs(authors)
			}

			// Recurse into sub-sections.
			if subSections, ok := secMap["section"]; ok {
				if arr, ok := subSections.([]interface{}); ok {
					walkSections(arr)
				}
			}
		}
	}

	if sections, ok := composition["section"]; ok {
		if arr, ok := sections.([]interface{}); ok {
			walkSections(arr)
		}
	}

	return refs
}

// DocumentHandler provides HTTP endpoints for the $document operation.
type DocumentHandler struct {
	generator *DocumentGenerator
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(generator *DocumentGenerator) *DocumentHandler {
	return &DocumentHandler{generator: generator}
}

// RegisterRoutes registers the $document routes on the given FHIR group.
func (h *DocumentHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/Composition/:id/$document", h.GenerateDocument)
	g.POST("/Composition/$document", h.GenerateDocumentFromBody)
}

// GenerateDocument handles GET /fhir/Composition/:id/$document.
// It resolves the Composition by ID and generates a Document Bundle.
func (h *DocumentHandler) GenerateDocument(c echo.Context) error {
	compositionID := c.Param("id")
	if compositionID == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("composition id is required"))
	}

	ctx := c.Request().Context()

	ref := "Composition/" + compositionID
	composition, err := h.generator.resolver.ResolveReference(ctx, ref)
	if err != nil || composition == nil {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("Composition", compositionID))
	}

	persist := c.QueryParam("persist") == "true"

	bundle, err := h.generator.GenerateDocument(ctx, composition, persist)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, bundle)
}

// GenerateDocumentFromBody handles POST /fhir/Composition/$document.
// It reads the Composition from the request body and generates a Document Bundle.
func (h *DocumentHandler) GenerateDocumentFromBody(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to read request body"))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("request body is empty"))
	}

	var composition map[string]interface{}
	if err := json.Unmarshal(body, &composition); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	ctx := c.Request().Context()
	persist := c.QueryParam("persist") == "true"

	bundle, err := h.generator.GenerateDocument(ctx, composition, persist)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, bundle)
}
