package provenance

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Provenance maps to the provenance table (FHIR Provenance resource).
type Provenance struct {
	ID              uuid.UUID `db:"id" json:"id"`
	FHIRID          string    `db:"fhir_id" json:"fhir_id"`
	TargetType      string    `db:"target_type" json:"target_type"`
	TargetID        string    `db:"target_id" json:"target_id"`
	Recorded        time.Time `db:"recorded" json:"recorded"`
	ActivityCode    *string   `db:"activity_code" json:"activity_code,omitempty"`
	ActivityDisplay *string   `db:"activity_display" json:"activity_display,omitempty"`
	ReasonCode      *string   `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay   *string   `db:"reason_display" json:"reason_display,omitempty"`
	VersionID       int       `db:"version_id" json:"version_id"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (p *Provenance) GetVersionID() int { return p.VersionID }

// SetVersionID sets the current version.
func (p *Provenance) SetVersionID(v int) { p.VersionID = v }

func (p *Provenance) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Provenance",
		"id":           p.FHIRID,
		"target": []fhir.Reference{{
			Reference: fhir.FormatReference(p.TargetType, p.TargetID),
		}},
		"recorded": p.Recorded.Format(time.RFC3339),
		"meta":     fhir.Meta{LastUpdated: p.UpdatedAt},
	}
	if p.ActivityCode != nil {
		result["activity"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *p.ActivityCode, Display: strVal(p.ActivityDisplay)}},
		}
	}
	if p.ReasonCode != nil {
		result["reason"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *p.ReasonCode, Display: strVal(p.ReasonDisplay)}},
		}}
	}
	return result
}

// ProvenanceAgent maps to the provenance_agent table.
type ProvenanceAgent struct {
	ID              uuid.UUID `db:"id" json:"id"`
	ProvenanceID    uuid.UUID `db:"provenance_id" json:"provenance_id"`
	TypeCode        *string   `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay     *string   `db:"type_display" json:"type_display,omitempty"`
	WhoType         string    `db:"who_type" json:"who_type"`
	WhoID           string    `db:"who_id" json:"who_id"`
	OnBehalfOfType  *string   `db:"on_behalf_of_type" json:"on_behalf_of_type,omitempty"`
	OnBehalfOfID    *string   `db:"on_behalf_of_id" json:"on_behalf_of_id,omitempty"`
}

// ProvenanceEntity maps to the provenance_entity table.
type ProvenanceEntity struct {
	ID           uuid.UUID `db:"id" json:"id"`
	ProvenanceID uuid.UUID `db:"provenance_id" json:"provenance_id"`
	Role         string    `db:"role" json:"role"`
	WhatType     string    `db:"what_type" json:"what_type"`
	WhatID       string    `db:"what_id" json:"what_id"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
