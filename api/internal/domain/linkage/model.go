package linkage

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Linkage maps to the linkage table (FHIR Linkage resource).
type Linkage struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Active             bool       `db:"active" json:"active"`
	AuthorID           *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	SourceType         string     `db:"source_type" json:"source_type"`
	SourceReference    string     `db:"source_reference" json:"source_reference"`
	AlternateType      *string    `db:"alternate_type" json:"alternate_type,omitempty"`
	AlternateReference *string    `db:"alternate_reference" json:"alternate_reference,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (l *Linkage) GetVersionID() int  { return l.VersionID }
func (l *Linkage) SetVersionID(v int) { l.VersionID = v }

func (l *Linkage) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Linkage",
		"id":           l.FHIRID,
		"active":       l.Active,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", l.VersionID),
			LastUpdated: l.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Linkage"},
		},
	}
	if l.AuthorID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", l.AuthorID.String())}
	}
	items := []map[string]interface{}{
		{
			"type":     "source",
			"resource": fhir.Reference{Reference: fhir.FormatReference(l.SourceType, l.SourceReference)},
		},
	}
	if l.AlternateType != nil && l.AlternateReference != nil {
		items = append(items, map[string]interface{}{
			"type":     "alternate",
			"resource": fhir.Reference{Reference: fhir.FormatReference(*l.AlternateType, *l.AlternateReference)},
		})
	}
	result["item"] = items
	return result
}
