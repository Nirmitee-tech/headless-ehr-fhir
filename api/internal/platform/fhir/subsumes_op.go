package fhir

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// SubsumptionResult represents the outcome of a $subsumes operation.
type SubsumptionResult string

const (
	// Subsumes means code A is an ancestor of code B.
	Subsumes SubsumptionResult = "subsumes"
	// SubsumedBy means code A is a descendant of code B.
	SubsumedBy SubsumptionResult = "subsumed-by"
	// Equivalent means code A and code B are the same.
	Equivalent SubsumptionResult = "equivalent"
	// NotSubsumed means there is no hierarchical relationship between A and B.
	NotSubsumed SubsumptionResult = "not-subsumed"
)

// SubsumptionChecker tests hierarchical relationships between codes.
type SubsumptionChecker struct {
	// hierarchies maps system URI → code → parent codes.
	// A code can have multiple parents (e.g., in SNOMED CT).
	hierarchies map[string]map[string][]string

	// icd10Hierarchy stores ICD-10 chapter-level parent→children relationships
	// using prefix-based subsumption.
	icd10Hierarchy []icd10Entry
}

// icd10Entry represents a parent code in ICD-10 with its child prefixes.
type icd10Entry struct {
	Code     string
	Children []string
}

// NewSubsumptionChecker creates a SubsumptionChecker with built-in SNOMED CT
// and ICD-10 hierarchies for common clinical concepts.
func NewSubsumptionChecker() *SubsumptionChecker {
	c := &SubsumptionChecker{
		hierarchies: make(map[string]map[string][]string),
	}
	c.loadSNOMEDHierarchies()
	c.loadICD10Hierarchies()
	return c
}

// loadSNOMEDHierarchies populates the SNOMED CT hierarchy data.
// Each entry maps a child code to its parent code(s).
func (c *SubsumptionChecker) loadSNOMEDHierarchies() {
	snomed := make(map[string][]string)

	// Diabetes hierarchy:
	// 73211009 (Diabetes mellitus)
	//   ├── 44054006 (Type 2 diabetes mellitus)
	//   │   ├── 313436004 (Type 2 diabetes with renal complications)
	//   │   └── 422034002 (Diabetic retinopathy associated with type 2 diabetes)
	//   ├── 46635009 (Type 1 diabetes mellitus)
	//   │   └── 420825003 (Diabetic ketoacidosis in type 1 diabetes)
	//   └── 11530004 (Gestational diabetes)
	snomed["44054006"] = []string{"73211009"}
	snomed["46635009"] = []string{"73211009"}
	snomed["11530004"] = []string{"73211009"}
	snomed["313436004"] = []string{"44054006"}
	snomed["422034002"] = []string{"44054006"}
	snomed["420825003"] = []string{"46635009"}

	// Hypertension hierarchy:
	// 38341003 (Hypertensive disorder)
	//   ├── 59621000 (Essential hypertension)
	//   ├── 70272006 (Malignant hypertension)
	//   └── 48146000 (Diastolic hypertension)
	snomed["59621000"] = []string{"38341003"}
	snomed["70272006"] = []string{"38341003"}
	snomed["48146000"] = []string{"38341003"}

	// Heart disease hierarchy:
	// 56265001 (Heart disease)
	//   ├── 84114007 (Heart failure)
	//   │   ├── 85232009 (Left heart failure)
	//   │   └── 367363000 (Right heart failure)
	//   ├── 49436004 (Atrial fibrillation)
	//   └── 22298006 (Myocardial infarction)
	//       ├── 401303003 (Acute ST elevation myocardial infarction)
	//       └── 401314000 (Acute non-ST elevation myocardial infarction)
	snomed["84114007"] = []string{"56265001"}
	snomed["49436004"] = []string{"56265001"}
	snomed["22298006"] = []string{"56265001"}
	snomed["85232009"] = []string{"84114007"}
	snomed["367363000"] = []string{"84114007"}
	snomed["401303003"] = []string{"22298006"}
	snomed["401314000"] = []string{"22298006"}

	// Respiratory hierarchy:
	// 50043002 (Respiratory disorder)
	//   ├── 195967001 (Asthma)
	//   │   ├── 389145006 (Allergic asthma)
	//   │   └── 233678006 (Childhood asthma)
	//   ├── 13645005 (COPD)
	//   └── 233604007 (Pneumonia)
	snomed["195967001"] = []string{"50043002"}
	snomed["13645005"] = []string{"50043002"}
	snomed["233604007"] = []string{"50043002"}
	snomed["389145006"] = []string{"195967001"}
	snomed["233678006"] = []string{"195967001"}

	c.hierarchies[systemSNOMED] = snomed
}

// loadICD10Hierarchies populates the ICD-10 chapter-level hierarchy.
// ICD-10 uses prefix-based subsumption: E11 subsumes E11.9, E11.65, etc.
func (c *SubsumptionChecker) loadICD10Hierarchies() {
	icd10 := make(map[string][]string)

	// ICD-10 chapter-level hierarchy:
	// E00-E89 (Endocrine diseases)
	//   ├── E08-E13 (Diabetes mellitus)
	//   │   ├── E10 (Type 1 diabetes)
	//   │   ├── E11 (Type 2 diabetes)
	//   │   └── E13 (Other specified diabetes)
	// I00-I99 (Circulatory diseases)
	//   ├── I10-I16 (Hypertensive diseases)
	//   │   └── I10 (Essential hypertension)
	//   ├── I20-I25 (Ischemic heart diseases)
	//   │   └── I21 (Acute myocardial infarction)
	//   └── I48 (Atrial fibrillation and flutter)

	// Diabetes chapter
	icd10["E10"] = []string{"E08-E13"}
	icd10["E11"] = []string{"E08-E13"}
	icd10["E13"] = []string{"E08-E13"}
	icd10["E08-E13"] = []string{"E00-E89"}

	// Hypertensive diseases
	icd10["I10"] = []string{"I10-I16"}
	icd10["I10-I16"] = []string{"I00-I99"}

	// Ischemic heart diseases
	icd10["I21"] = []string{"I20-I25"}
	icd10["I20-I25"] = []string{"I00-I99"}

	// Atrial fibrillation
	icd10["I48"] = []string{"I00-I99"}

	c.hierarchies[systemICD10] = icd10
}

// CheckSubsumption determines the hierarchical relationship between codeA and
// codeB within the given code system.
func (c *SubsumptionChecker) CheckSubsumption(system, codeA, codeB string) (SubsumptionResult, error) {
	// Step 1: Equivalent if codes are the same.
	if codeA == codeB {
		return Equivalent, nil
	}

	switch system {
	case systemSNOMED:
		return c.checkSNOMED(codeA, codeB), nil

	case systemICD10:
		return c.checkICD10(codeA, codeB), nil

	default:
		return "", fmt.Errorf("unsupported code system: %s", system)
	}
}

// checkSNOMED checks subsumption using the SNOMED CT hierarchy by walking
// parent chains.
func (c *SubsumptionChecker) checkSNOMED(codeA, codeB string) SubsumptionResult {
	hierarchy := c.hierarchies[systemSNOMED]

	// Check if codeA is an ancestor of codeB (codeA subsumes codeB).
	if c.isAncestor(hierarchy, codeA, codeB) {
		return Subsumes
	}

	// Check if codeB is an ancestor of codeA (codeA subsumed-by codeB).
	if c.isAncestor(hierarchy, codeB, codeA) {
		return SubsumedBy
	}

	return NotSubsumed
}

// isAncestor walks up the parent chain of descendant to see if ancestor
// appears anywhere in the chain.
func (c *SubsumptionChecker) isAncestor(hierarchy map[string][]string, ancestor, descendant string) bool {
	// Use a visited set to avoid infinite loops in case of cycles.
	visited := make(map[string]bool)
	return c.walkParents(hierarchy, ancestor, descendant, visited)
}

// walkParents recursively walks up the parent chain looking for the ancestor code.
func (c *SubsumptionChecker) walkParents(hierarchy map[string][]string, ancestor, current string, visited map[string]bool) bool {
	if visited[current] {
		return false
	}
	visited[current] = true

	parents, ok := hierarchy[current]
	if !ok {
		return false
	}

	for _, parent := range parents {
		if parent == ancestor {
			return true
		}
		if c.walkParents(hierarchy, ancestor, parent, visited) {
			return true
		}
	}

	return false
}

// checkICD10 checks subsumption using prefix-based matching for ICD-10 codes
// and the explicit ICD-10 chapter hierarchy.
func (c *SubsumptionChecker) checkICD10(codeA, codeB string) SubsumptionResult {
	// Prefix-based subsumption: E11 subsumes E11.9, E11.65 etc.
	if c.icd10PrefixSubsumes(codeA, codeB) {
		return Subsumes
	}
	if c.icd10PrefixSubsumes(codeB, codeA) {
		return SubsumedBy
	}

	// Check explicit hierarchy.
	hierarchy := c.hierarchies[systemICD10]
	if c.isAncestor(hierarchy, codeA, codeB) {
		return Subsumes
	}
	if c.isAncestor(hierarchy, codeB, codeA) {
		return SubsumedBy
	}

	return NotSubsumed
}

// icd10PrefixSubsumes returns true if parent is a proper prefix of child,
// where the child either continues with a dot or more characters after the
// parent prefix. For example, "E11" is a prefix of "E11.9" and "E11.65".
func (c *SubsumptionChecker) icd10PrefixSubsumes(parent, child string) bool {
	if len(parent) >= len(child) {
		return false
	}
	if !strings.HasPrefix(child, parent) {
		return false
	}
	// The next character after the prefix must be a dot or the prefix must
	// end at a natural boundary.
	rest := child[len(parent):]
	return strings.HasPrefix(rest, ".") || len(rest) > 0
}

// SubsumesHandler provides the CodeSystem/$subsumes HTTP endpoints.
type SubsumesHandler struct {
	checker *SubsumptionChecker
}

// NewSubsumesHandler creates a new SubsumesHandler.
func NewSubsumesHandler(checker *SubsumptionChecker) *SubsumesHandler {
	return &SubsumesHandler{checker: checker}
}

// RegisterRoutes adds CodeSystem/$subsumes routes to the given FHIR group.
func (h *SubsumesHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/CodeSystem/$subsumes", h.HandleSubsumes)
	g.POST("/CodeSystem/$subsumes", h.HandleSubsumesPost)
}

// HandleSubsumes handles GET /fhir/CodeSystem/$subsumes with query parameters.
func (h *SubsumesHandler) HandleSubsumes(c echo.Context) error {
	system := c.QueryParam("system")
	codeA := c.QueryParam("codeA")
	codeB := c.QueryParam("codeB")

	if system == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'system' is required"))
	}
	if codeA == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'codeA' is required"))
	}
	if codeB == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'codeB' is required"))
	}

	return h.doSubsumes(c, system, codeA, codeB)
}

// HandleSubsumesPost handles POST /fhir/CodeSystem/$subsumes with a Parameters
// resource body.
func (h *SubsumesHandler) HandleSubsumesPost(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Failed to read request body"))
	}

	var params struct {
		ResourceType string `json:"resourceType"`
		Parameter    []struct {
			Name        string `json:"name"`
			ValueCode   string `json:"valueCode,omitempty"`
			ValueUri    string `json:"valueUri,omitempty"`
			ValueString string `json:"valueString,omitempty"`
		} `json:"parameter"`
	}

	if err := json.Unmarshal(body, &params); err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
	}

	var system, codeA, codeB string
	for _, p := range params.Parameter {
		switch p.Name {
		case "system":
			system = p.ValueUri
		case "codeA":
			codeA = p.ValueCode
		case "codeB":
			codeB = p.ValueCode
		}
	}

	if system == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'system' is required"))
	}
	if codeA == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'codeA' is required"))
	}
	if codeB == "" {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "required", "Parameter 'codeB' is required"))
	}

	return h.doSubsumes(c, system, codeA, codeB)
}

// doSubsumes performs the subsumption check and returns a FHIR Parameters
// response.
func (h *SubsumesHandler) doSubsumes(c echo.Context, system, codeA, codeB string) error {
	result, err := h.checker.CheckSubsumption(system, codeA, codeB)
	if err != nil {
		return c.JSON(http.StatusBadRequest, operationOutcome("error", "not-supported", err.Error()))
	}

	return c.JSON(http.StatusOK, buildSubsumesResponse(result))
}

// buildSubsumesResponse builds a FHIR Parameters resource with the subsumption
// outcome.
func buildSubsumesResponse(result SubsumptionResult) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter": []map[string]interface{}{
			{
				"name":      "outcome",
				"valueCode": string(result),
			},
		},
	}
}
