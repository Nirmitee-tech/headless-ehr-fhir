package fhir

import (
	"encoding/json"
	"fmt"
	"time"
)

// Bundle represents a FHIR Bundle resource.
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	ID           string        `json:"id,omitempty"`
	Type         string        `json:"type"`
	Total        *int          `json:"total,omitempty"`
	Link         []BundleLink  `json:"link,omitempty"`
	Entry        []BundleEntry `json:"entry,omitempty"`
	Timestamp    *time.Time    `json:"timestamp,omitempty"`
}

type BundleLink struct {
	Relation string `json:"relation"`
	URL      string `json:"url"`
}

type BundleEntry struct {
	FullURL  string          `json:"fullUrl,omitempty"`
	Resource json.RawMessage `json:"resource,omitempty"`
	Search   *BundleSearch   `json:"search,omitempty"`
	Request  *BundleRequest  `json:"request,omitempty"`
	Response *BundleResponse `json:"response,omitempty"`
}

type BundleSearch struct {
	Mode  string   `json:"mode,omitempty"`
	Score *float64 `json:"score,omitempty"`
}

type BundleRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type BundleResponse struct {
	Status       string      `json:"status"`
	Location     string      `json:"location,omitempty"`
	LastModified *time.Time  `json:"lastModified,omitempty"`
	Outcome      interface{} `json:"outcome,omitempty"`
}

// SearchBundleParams holds pagination and link information for a search bundle.
type SearchBundleParams struct {
	BaseURL    string
	QueryStr   string
	Count      int
	Offset     int
	Total      int
}

// NewSearchBundle creates a searchset Bundle from a list of resources.
// It populates fullUrl for each entry and sets self/next/previous links.
func NewSearchBundle(resources []interface{}, total int, baseURL string) *Bundle {
	now := time.Now().UTC()
	entries := make([]BundleEntry, len(resources))
	for i, r := range resources {
		raw, _ := json.Marshal(r)
		fullURL := extractFullURL(r, baseURL)
		entries[i] = BundleEntry{
			FullURL:  fullURL,
			Resource: raw,
			Search: &BundleSearch{
				Mode: "match",
			},
		}
	}

	return &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &total,
		Timestamp:    &now,
		Link: []BundleLink{
			{Relation: "self", URL: baseURL},
		},
		Entry: entries,
	}
}

// NewSearchBundleWithLinks creates a searchset Bundle with proper pagination links.
func NewSearchBundleWithLinks(resources []interface{}, params SearchBundleParams) *Bundle {
	now := time.Now().UTC()
	entries := make([]BundleEntry, len(resources))
	for i, r := range resources {
		raw, _ := json.Marshal(r)
		fullURL := extractFullURL(r, params.BaseURL)
		entries[i] = BundleEntry{
			FullURL:  fullURL,
			Resource: raw,
			Search: &BundleSearch{
				Mode: "match",
			},
		}
	}

	links := buildPaginationLinks(params)

	return &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &params.Total,
		Timestamp:    &now,
		Link:         links,
		Entry:        entries,
	}
}

// NewTransactionResponse creates a transaction-response Bundle from entry outcomes.
func NewTransactionResponse(entries []BundleEntry) *Bundle {
	now := time.Now().UTC()
	return &Bundle{
		ResourceType: "Bundle",
		Type:         "transaction-response",
		Timestamp:    &now,
		Entry:        entries,
	}
}

// NewBatchResponse creates a batch-response Bundle from entry outcomes.
func NewBatchResponse(entries []BundleEntry) *Bundle {
	now := time.Now().UTC()
	return &Bundle{
		ResourceType: "Bundle",
		Type:         "batch-response",
		Timestamp:    &now,
		Entry:        entries,
	}
}

// extractFullURL attempts to build a fullUrl from a resource's resourceType and id.
func extractFullURL(r interface{}, baseURL string) string {
	m, ok := toMap(r)
	if !ok {
		return ""
	}
	rt, _ := m["resourceType"].(string)
	id, _ := m["id"].(string)
	if rt != "" && id != "" {
		return fmt.Sprintf("%s/%s", rt, id)
	}
	return ""
}

// toMap converts an interface{} to map[string]interface{} if possible.
func toMap(v interface{}) (map[string]interface{}, bool) {
	switch val := v.(type) {
	case map[string]interface{}:
		return val, true
	case map[string]string:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = v
		}
		return m, true
	default:
		// Try via JSON round-trip for struct types.
		data, err := json.Marshal(v)
		if err != nil {
			return nil, false
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, false
		}
		return m, true
	}
}

// buildPaginationLinks creates self, next, and previous links for searchset bundles.
func buildPaginationLinks(params SearchBundleParams) []BundleLink {
	links := []BundleLink{
		{
			Relation: "self",
			URL:      fmt.Sprintf("%s?%s_count=%d&_offset=%d", params.BaseURL, conditionalAmpersand(params.QueryStr), params.Count, params.Offset),
		},
	}

	// Next link: only if there are more results
	nextOffset := params.Offset + params.Count
	if nextOffset < params.Total {
		links = append(links, BundleLink{
			Relation: "next",
			URL:      fmt.Sprintf("%s?%s_count=%d&_offset=%d", params.BaseURL, conditionalAmpersand(params.QueryStr), params.Count, nextOffset),
		})
	}

	// Previous link: only if not at the first page
	if params.Offset > 0 {
		prevOffset := params.Offset - params.Count
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, BundleLink{
			Relation: "previous",
			URL:      fmt.Sprintf("%s?%s_count=%d&_offset=%d", params.BaseURL, conditionalAmpersand(params.QueryStr), params.Count, prevOffset),
		})
	}

	return links
}

// conditionalAmpersand returns the query string with a trailing & if non-empty.
func conditionalAmpersand(qs string) string {
	if qs == "" {
		return ""
	}
	return qs + "&"
}

// CapabilityStatement represents the FHIR CapabilityStatement (metadata).
type CapabilityStatement struct {
	ResourceType  string                  `json:"resourceType"`
	Status        string                  `json:"status"`
	Date          string                  `json:"date"`
	Kind          string                  `json:"kind"`
	FHIRVersion   string                  `json:"fhirVersion"`
	Format        []string                `json:"format"`
	Implementation *CSImplementation      `json:"implementation,omitempty"`
	Rest          []CSRest                `json:"rest"`
}

type CSImplementation struct {
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
}

type CSRest struct {
	Mode     string       `json:"mode"`
	Resource []CSResource `json:"resource"`
	Security *CSSecurity  `json:"security,omitempty"`
}

type CSResource struct {
	Type          string            `json:"type"`
	Interaction   []CSInteraction   `json:"interaction"`
	SearchParam   []CSSearchParam   `json:"searchParam,omitempty"`
	Versioning    string            `json:"versioning,omitempty"`
	ReadHistory   bool              `json:"readHistory,omitempty"`
}

type CSInteraction struct {
	Code string `json:"code"`
}

type CSSearchParam struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Definition string `json:"definition,omitempty"`
}

type CSSecurity struct {
	CORS    bool             `json:"cors"`
	Service []CodeableConcept `json:"service,omitempty"`
}

// NewCapabilityStatement creates the server's capability statement.
func NewCapabilityStatement(baseURL string, resources []CSResource) *CapabilityStatement {
	return &CapabilityStatement{
		ResourceType: "CapabilityStatement",
		Status:       "active",
		Date:         time.Now().UTC().Format("2006-01-02"),
		Kind:         "instance",
		FHIRVersion:  "4.0.1",
		Format:       []string{"json"},
		Implementation: &CSImplementation{
			Description: "Headless EHR FHIR R4 Server",
			URL:         baseURL,
		},
		Rest: []CSRest{
			{
				Mode:     "server",
				Resource: resources,
				Security: &CSSecurity{
					CORS: true,
					Service: []CodeableConcept{
						{
							Coding: []Coding{
								{
									System:  "http://terminology.hl7.org/CodeSystem/restful-security-service",
									Code:    "SMART-on-FHIR",
									Display: "SMART on FHIR",
								},
							},
							Text: "OAuth2 using SMART on FHIR profile",
						},
					},
				},
			},
		},
	}
}

// ResourceCapability creates a CSResource with standard CRUD interactions.
func ResourceCapability(resourceType string, searchParams []CSSearchParam) CSResource {
	return CSResource{
		Type: resourceType,
		Interaction: []CSInteraction{
			{Code: "read"},
			{Code: "vread"},
			{Code: "search-type"},
			{Code: "create"},
			{Code: "update"},
			{Code: "delete"},
		},
		SearchParam: searchParams,
		Versioning:  "versioned",
	}
}

// FormatReference creates a FHIR reference string.
func FormatReference(resourceType, id string) string {
	return fmt.Sprintf("%s/%s", resourceType, id)
}
