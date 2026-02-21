package searchparameter

import (
	"strings"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// SearchParameter maps to the search_parameter table (FHIR SearchParameter resource).
type SearchParameter struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FHIRID      string     `db:"fhir_id" json:"fhir_id"`
	Status      string     `db:"status" json:"status"`
	URL         string     `db:"url" json:"url"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	Code        string     `db:"code" json:"code"`
	Base        string     `db:"base" json:"base"`
	Type        string     `db:"type" json:"type"`
	Expression  *string    `db:"expression" json:"expression,omitempty"`
	XPath       *string    `db:"xpath" json:"xpath,omitempty"`
	Target      *string    `db:"target" json:"target,omitempty"`
	Modifier    *string    `db:"modifier" json:"modifier,omitempty"`
	Comparator  *string    `db:"comparator" json:"comparator,omitempty"`
	Publisher   *string    `db:"publisher" json:"publisher,omitempty"`
	Date        *time.Time `db:"date" json:"date,omitempty"`
	VersionID   int        `db:"version_id" json:"version_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func (s *SearchParameter) GetVersionID() int  { return s.VersionID }
func (s *SearchParameter) SetVersionID(v int) { s.VersionID = v }

func (s *SearchParameter) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "SearchParameter",
		"id":           s.FHIRID,
		"url":          s.URL,
		"name":         s.Name,
		"status":       s.Status,
		"description":  s.Description,
		"code":         s.Code,
		"base":         splitComma(s.Base),
		"type":         s.Type,
		"meta":         fhir.Meta{
			LastUpdated: s.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/SearchParameter"},
		},
	}
	if s.Expression != nil {
		result["expression"] = *s.Expression
	}
	if s.XPath != nil {
		result["xpath"] = *s.XPath
	}
	if s.Target != nil {
		result["target"] = splitComma(*s.Target)
	}
	return result
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
