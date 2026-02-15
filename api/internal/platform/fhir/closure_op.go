package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
)

// ClosureTable represents a named closure table that tracks hierarchical
// relationships between concepts.
type ClosureTable struct {
	Name    string
	entries map[string]map[string]string // [code]map[ancestor]relationship
	version int
	mu      sync.RWMutex
}

// ClosureManager manages closure tables and computes transitive closures
// using known code system hierarchies.
type ClosureManager struct {
	mu          sync.RWMutex
	tables      map[string]*ClosureTable
	hierarchies map[string]map[string][]string // code system -> code -> parent codes
}

// NewClosureManager creates a new ClosureManager with built-in SNOMED CT
// hierarchies (reused from the subsumption checker data).
func NewClosureManager() *ClosureManager {
	m := &ClosureManager{
		tables:      make(map[string]*ClosureTable),
		hierarchies: make(map[string]map[string][]string),
	}
	m.loadHierarchies()
	return m
}

// loadHierarchies populates the built-in code system hierarchies for closure
// computation. Uses the same SNOMED CT hierarchy as SubsumptionChecker.
func (m *ClosureManager) loadHierarchies() {
	snomed := make(map[string][]string)

	// Clinical finding hierarchy (top-level categories):
	// 404684003 (Clinical finding) -> 64572001 (Disease) -> 40733004 (Infectious disease)
	// 404684003 (Clinical finding) -> 64572001 (Disease) -> 55342001 (Neoplasm)
	snomed["64572001"] = []string{"404684003"}
	snomed["40733004"] = []string{"64572001"}
	snomed["55342001"] = []string{"64572001"}

	// Body structure hierarchy:
	// 123037004 (Body structure) -> 91723000 (Anatomical structure)
	snomed["91723000"] = []string{"123037004"}

	// Diabetes hierarchy (same as subsumes_op.go):
	snomed["44054006"] = []string{"73211009"}
	snomed["46635009"] = []string{"73211009"}
	snomed["11530004"] = []string{"73211009"}
	snomed["313436004"] = []string{"44054006"}
	snomed["422034002"] = []string{"44054006"}
	snomed["420825003"] = []string{"46635009"}

	// Heart disease hierarchy:
	snomed["84114007"] = []string{"56265001"}
	snomed["49436004"] = []string{"56265001"}
	snomed["22298006"] = []string{"56265001"}
	snomed["85232009"] = []string{"84114007"}
	snomed["367363000"] = []string{"84114007"}
	snomed["401303003"] = []string{"22298006"}
	snomed["401314000"] = []string{"22298006"}

	// Respiratory hierarchy:
	snomed["195967001"] = []string{"50043002"}
	snomed["13645005"] = []string{"50043002"}
	snomed["233604007"] = []string{"50043002"}
	snomed["389145006"] = []string{"195967001"}
	snomed["233678006"] = []string{"195967001"}

	// Hypertension hierarchy:
	snomed["59621000"] = []string{"38341003"}
	snomed["70272006"] = []string{"38341003"}
	snomed["48146000"] = []string{"38341003"}

	m.hierarchies[systemSNOMED] = snomed
}

// InitializeClosure creates a new empty closure table with the given name.
func (m *ClosureManager) InitializeClosure(name string) (*ClosureTable, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tables[name]; exists {
		return nil, fmt.Errorf("closure table %q already exists", name)
	}

	table := &ClosureTable{
		Name:    name,
		entries: make(map[string]map[string]string),
		version: 0,
	}
	m.tables[name] = table
	return table, nil
}

// ProcessConcepts adds concepts to a closure table and computes their
// transitive closure relationships. Returns a ClosureConceptMap with all
// discovered relationships.
func (m *ClosureManager) ProcessConcepts(name string, concepts []ClosureConcept) (*ClosureConceptMap, error) {
	m.mu.RLock()
	table, ok := m.tables[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("closure table %q not found", name)
	}

	table.mu.Lock()
	defer table.mu.Unlock()

	table.version++

	// Collect all concept codes by system.
	codesBySystem := make(map[string][]string)
	for _, c := range concepts {
		codesBySystem[c.System] = append(codesBySystem[c.System], c.Code)
		// Ensure entries map exists for each code.
		if table.entries[c.Code] == nil {
			table.entries[c.Code] = make(map[string]string)
		}
	}

	// Build relationships using hierarchy data.
	groups := make(map[string]*ClosureGroup) // keyed by system
	for system, codes := range codesBySystem {
		hierarchy, ok := m.hierarchies[system]
		if !ok {
			continue
		}

		group, exists := groups[system]
		if !exists {
			group = &ClosureGroup{
				Source: system,
				Target: system,
			}
			groups[system] = group
		}

		// Build ancestor map for all codes in this batch.
		codeSet := make(map[string]bool, len(codes))
		for _, c := range codes {
			codeSet[c] = true
		}

		ancestorMap := make(map[string]map[string]bool) // code -> set of ancestors
		for _, code := range codes {
			ancestors := closureGetAllAncestors(hierarchy, code)
			ancestorMap[code] = ancestors
		}

		// Generate relationships: for each code, check if any other code in
		// the batch is an ancestor.
		elementsMap := make(map[string]*ClosureElement)
		for _, code := range codes {
			for _, otherCode := range codes {
				if code == otherCode {
					continue
				}
				// Check if otherCode is a descendant of code (code subsumes otherCode).
				if ancestorMap[otherCode][code] {
					elem, ok := elementsMap[code]
					if !ok {
						elem = &ClosureElement{Code: code}
						elementsMap[code] = elem
					}
					elem.Targets = append(elem.Targets, ClosureTarget{
						Code:        otherCode,
						Equivalence: "subsumes",
					})
					// Record in table entries.
					table.entries[otherCode][code] = "subsumes"
				}
			}
		}

		for _, elem := range elementsMap {
			group.Elements = append(group.Elements, *elem)
		}
	}

	// Build response.
	var groupList []ClosureGroup
	for _, g := range groups {
		groupList = append(groupList, *g)
	}

	return &ClosureConceptMap{
		ResourceType: "ConceptMap",
		Name:         name,
		Version:      strconv.Itoa(table.version),
		Groups:       groupList,
	}, nil
}

// closureGetAllAncestors returns a set of all ancestor codes for the given
// code by walking up the hierarchy transitively.
func closureGetAllAncestors(hierarchy map[string][]string, code string) map[string]bool {
	ancestors := make(map[string]bool)
	visited := make(map[string]bool)
	closureWalkAncestors(hierarchy, code, ancestors, visited)
	return ancestors
}

// closureWalkAncestors recursively walks up the hierarchy collecting all ancestors.
func closureWalkAncestors(hierarchy map[string][]string, code string, ancestors, visited map[string]bool) {
	if visited[code] {
		return
	}
	visited[code] = true

	parents, ok := hierarchy[code]
	if !ok {
		return
	}

	for _, parent := range parents {
		ancestors[parent] = true
		closureWalkAncestors(hierarchy, parent, ancestors, visited)
	}
}

// GetClosure retrieves a closure table by name.
func (m *ClosureManager) GetClosure(name string) (*ClosureTable, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	table, ok := m.tables[name]
	if !ok {
		return nil, fmt.Errorf("closure table %q not found", name)
	}
	return table, nil
}

// DeleteClosure removes a closure table by name.
func (m *ClosureManager) DeleteClosure(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tables[name]; !ok {
		return fmt.Errorf("closure table %q not found", name)
	}
	delete(m.tables, name)
	return nil
}

// =========== Response Types ===========

// ClosureConcept represents a concept to be added to a closure table.
type ClosureConcept struct {
	System  string ` json:"system" `
	Code    string ` json:"code" `
	Display string ` json:"display,omitempty" `
}

// ClosureConceptMap is the FHIR ConceptMap response for the $closure operation.
// Named ClosureConceptMap to avoid conflict with the ConceptMap type in
// translate_op.go.
type ClosureConceptMap struct {
	ResourceType string         ` json:"resourceType" `
	Name         string         ` json:"name" `
	Version      string         ` json:"version" `
	Groups       []ClosureGroup ` json:"group,omitempty" `
}

// ClosureGroup represents a group of closure relationships within the same
// code system.
type ClosureGroup struct {
	Source   string           ` json:"source" `
	Target   string           ` json:"target" `
	Elements []ClosureElement ` json:"element,omitempty" `
}

// ClosureElement represents a source code with its closure targets.
type ClosureElement struct {
	Code    string          ` json:"code" `
	Targets []ClosureTarget ` json:"target,omitempty" `
}

// ClosureTarget represents a target code in a closure relationship.
type ClosureTarget struct {
	Code        string ` json:"code" `
	Equivalence string ` json:"equivalence" `
}

// =========== HTTP Handler ===========

// ClosureHandler provides the CodeSystem/$closure HTTP endpoint.
type ClosureHandler struct {
	manager *ClosureManager
}

// NewClosureHandler creates a new closure HTTP handler.
func NewClosureHandler(manager *ClosureManager) *ClosureHandler {
	return &ClosureHandler{manager: manager}
}

// RegisterRoutes adds the CodeSystem/$closure route to the given FHIR group.
func (h *ClosureHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/CodeSystem/$closure", h.HandleClosure)
}

// closureRequest is the JSON structure for incoming $closure requests.
type closureRequest struct {
	Name    string           ` json:"name" `
	Concept []ClosureConcept ` json:"concept,omitempty" `
}

// HandleClosure handles POST /fhir/CodeSystem/$closure.
// If the request contains only a name, it initializes a new closure table.
// If the request contains name + concepts, it processes the concepts.
func (h *ClosureHandler) HandleClosure(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Failed to read request body"))
	}

	var req closureRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'name' is required"))
	}

	// If no concepts, this is an initialization request.
	if len(req.Concept) == 0 {
		table, err := h.manager.InitializeClosure(req.Name)
		if err != nil {
			return c.JSON(http.StatusBadRequest, operationOutcome("error", "processing", err.Error()))
		}
		return c.JSON(http.StatusOK, &ClosureConceptMap{
			ResourceType: "ConceptMap",
			Name:         table.Name,
			Version:      "0",
		})
	}

	// Process concepts.
	cm, err := h.manager.ProcessConcepts(req.Name, req.Concept)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "processing", err.Error()))
	}

	return c.JSON(http.StatusOK, cm)
}
