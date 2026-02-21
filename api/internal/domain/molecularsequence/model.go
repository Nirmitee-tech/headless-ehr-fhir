package molecularsequence

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// MolecularSequence maps to the molecular_sequence table (FHIR MolecularSequence resource).
type MolecularSequence struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	FHIRID             string     `db:"fhir_id" json:"fhir_id"`
	Type               string     `db:"type" json:"type"`
	PatientID          *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	SpecimenID         *uuid.UUID `db:"specimen_id" json:"specimen_id,omitempty"`
	DeviceID           *uuid.UUID `db:"device_id" json:"device_id,omitempty"`
	PerformerID        *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	CoordinateSystem   int        `db:"coordinate_system" json:"coordinate_system"`
	ObservedSeq        *string    `db:"observed_seq" json:"observed_seq,omitempty"`
	ReferenceSeqID     *string    `db:"reference_seq_id" json:"reference_seq_id,omitempty"`
	ReferenceSeqStrand *string    `db:"reference_seq_strand" json:"reference_seq_strand,omitempty"`
	WindowStart        *int       `db:"window_start" json:"window_start,omitempty"`
	WindowEnd          *int       `db:"window_end" json:"window_end,omitempty"`
	VersionID          int        `db:"version_id" json:"version_id"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

func (m *MolecularSequence) GetVersionID() int  { return m.VersionID }
func (m *MolecularSequence) SetVersionID(v int) { m.VersionID = v }

func (m *MolecularSequence) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType":     "MolecularSequence",
		"id":               m.FHIRID,
		"type":             m.Type,
		"coordinateSystem": m.CoordinateSystem,
		"meta":             fhir.Meta{
			VersionID:   fmt.Sprintf("%d", m.VersionID),
			LastUpdated: m.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/MolecularSequence"},
		},
	}
	if m.PatientID != nil {
		result["patient"] = fhir.Reference{Reference: fhir.FormatReference("Patient", m.PatientID.String())}
	}
	if m.SpecimenID != nil {
		result["specimen"] = fhir.Reference{Reference: fhir.FormatReference("Specimen", m.SpecimenID.String())}
	}
	if m.DeviceID != nil {
		result["device"] = fhir.Reference{Reference: fhir.FormatReference("Device", m.DeviceID.String())}
	}
	if m.PerformerID != nil {
		result["performer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", m.PerformerID.String())}
	}
	if m.ObservedSeq != nil {
		result["observedSeq"] = *m.ObservedSeq
	}
	if m.ReferenceSeqID != nil || m.ReferenceSeqStrand != nil {
		refSeq := map[string]interface{}{}
		if m.ReferenceSeqID != nil {
			refSeq["referenceSeqId"] = map[string]interface{}{
				"text": *m.ReferenceSeqID,
			}
		}
		if m.ReferenceSeqStrand != nil {
			refSeq["strand"] = *m.ReferenceSeqStrand
		}
		result["referenceSeq"] = refSeq
	}
	if m.WindowStart != nil || m.WindowEnd != nil {
		variant := map[string]interface{}{}
		if m.WindowStart != nil {
			variant["start"] = *m.WindowStart
		}
		if m.WindowEnd != nil {
			variant["end"] = *m.WindowEnd
		}
		result["variant"] = []map[string]interface{}{variant}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
