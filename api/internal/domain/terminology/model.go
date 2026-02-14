package terminology

// LOINCCode represents a LOINC reference code.
type LOINCCode struct {
	Code       string `db:"code" json:"code"`
	Display    string `db:"display" json:"display"`
	Component  string `db:"component" json:"component,omitempty"`
	Property   string `db:"property" json:"property,omitempty"`
	TimeAspect string `db:"time_aspect" json:"time_aspect,omitempty"`
	SystemURI  string `db:"system_uri" json:"system_uri"`
	Category   string `db:"category" json:"category,omitempty"`
}

// ICD10Code represents an ICD-10-CM diagnosis code.
type ICD10Code struct {
	Code      string `db:"code" json:"code"`
	Display   string `db:"display" json:"display"`
	Category  string `db:"category" json:"category,omitempty"`
	Chapter   string `db:"chapter" json:"chapter,omitempty"`
	SystemURI string `db:"system_uri" json:"system_uri"`
}

// SNOMEDCode represents a SNOMED CT code.
type SNOMEDCode struct {
	Code        string `db:"code" json:"code"`
	Display     string `db:"display" json:"display"`
	SemanticTag string `db:"semantic_tag" json:"semantic_tag,omitempty"`
	Category    string `db:"category" json:"category,omitempty"`
	SystemURI   string `db:"system_uri" json:"system_uri"`
}

// RxNormCode represents a medication RxNorm code.
type RxNormCode struct {
	RxNormCode  string `db:"rxnorm_code" json:"code"`
	Display     string `db:"display" json:"display"`
	GenericName string `db:"generic_name" json:"generic_name,omitempty"`
	DrugClass   string `db:"drug_class" json:"drug_class,omitempty"`
	Route       string `db:"route" json:"route,omitempty"`
	Form        string `db:"form" json:"form,omitempty"`
	SystemURI   string `db:"system_uri" json:"system_uri"`
}

// CPTCode represents a CPT procedure code.
type CPTCode struct {
	Code        string `db:"code" json:"code"`
	Display     string `db:"display" json:"display"`
	Category    string `db:"category" json:"category,omitempty"`
	Subcategory string `db:"subcategory" json:"subcategory,omitempty"`
	SystemURI   string `db:"system_uri" json:"system_uri"`
}

// SearchResult is a generic terminology search result used by the service layer.
type SearchResult struct {
	Code      string `json:"code"`
	Display   string `json:"display"`
	SystemURI string `json:"system"`
}

// LookupRequest represents a FHIR CodeSystem $lookup request.
type LookupRequest struct {
	System string `json:"system"`
	Code   string `json:"code"`
}

// LookupResponse represents a FHIR CodeSystem $lookup response.
type LookupResponse struct {
	ResourceType string            `json:"resourceType"`
	Parameter    []LookupParameter `json:"parameter"`
}

// LookupParameter is a name/value pair in a FHIR Parameters resource.
type LookupParameter struct {
	Name        string `json:"name"`
	ValueString string `json:"valueString,omitempty"`
	ValueCode   string `json:"valueCode,omitempty"`
}

// ValidateCodeRequest represents a FHIR CodeSystem $validate-code request.
type ValidateCodeRequest struct {
	System  string `json:"system"`
	Code    string `json:"code"`
	Display string `json:"display,omitempty"`
}

// ValidateCodeResponse represents a FHIR CodeSystem $validate-code response.
type ValidateCodeResponse struct {
	ResourceType string                  `json:"resourceType"`
	Parameter    []ValidateCodeParameter `json:"parameter"`
}

// ValidateCodeParameter is a name/value pair in a validate-code response.
type ValidateCodeParameter struct {
	Name         string `json:"name"`
	ValueBoolean *bool  `json:"valueBoolean,omitempty"`
	ValueString  string `json:"valueString,omitempty"`
}

// CodeSystemURI constants for well-known terminology systems.
const (
	SystemLOINC  = "http://loinc.org"
	SystemICD10  = "http://hl7.org/fhir/sid/icd-10-cm"
	SystemSNOMED = "http://snomed.info/sct"
	SystemRxNorm = "http://www.nlm.nih.gov/research/umls/rxnorm"
	SystemCPT    = "http://www.ama-assn.org/go/cpt"
)
