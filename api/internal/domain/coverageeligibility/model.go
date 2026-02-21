package coverageeligibility

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// CoverageEligibilityRequest maps to the coverage_eligibility_request table (FHIR CoverageEligibilityRequest resource).
type CoverageEligibilityRequest struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FHIRID      string     `db:"fhir_id" json:"fhir_id"`
	Status      string     `db:"status" json:"status"`
	PatientID   uuid.UUID  `db:"patient_id" json:"patient_id"`
	ProviderID  *uuid.UUID `db:"provider_id" json:"provider_id,omitempty"`
	InsurerID   *uuid.UUID `db:"insurer_id" json:"insurer_id,omitempty"`
	Purpose     string     `db:"purpose" json:"purpose"`
	ServicedDate *time.Time `db:"serviced_date" json:"serviced_date,omitempty"`
	Created     *time.Time `db:"created" json:"created,omitempty"`
	VersionID   int        `db:"version_id" json:"version_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func (r *CoverageEligibilityRequest) GetVersionID() int  { return r.VersionID }
func (r *CoverageEligibilityRequest) SetVersionID(v int) { r.VersionID = v }

func (r *CoverageEligibilityRequest) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CoverageEligibilityRequest",
		"id":           r.FHIRID,
		"status":       r.Status,
		"purpose":      []string{r.Purpose},
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", r.PatientID.String())},
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", r.VersionID),
			LastUpdated: r.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/CoverageEligibilityRequest"},
		},
	}
	if r.ProviderID != nil {
		result["provider"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", r.ProviderID.String())}
	}
	if r.InsurerID != nil {
		result["insurer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", r.InsurerID.String())}
	}
	if r.ServicedDate != nil {
		result["servicedDate"] = r.ServicedDate.Format("2006-01-02")
	}
	if r.Created != nil {
		result["created"] = r.Created.Format("2006-01-02T15:04:05Z")
	}
	return result
}

// CoverageEligibilityResponse maps to the coverage_eligibility_response table (FHIR CoverageEligibilityResponse resource).
type CoverageEligibilityResponse struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FHIRID      string     `db:"fhir_id" json:"fhir_id"`
	Status      string     `db:"status" json:"status"`
	PatientID   uuid.UUID  `db:"patient_id" json:"patient_id"`
	RequestID   *uuid.UUID `db:"request_id" json:"request_id,omitempty"`
	InsurerID   *uuid.UUID `db:"insurer_id" json:"insurer_id,omitempty"`
	Outcome     string     `db:"outcome" json:"outcome"`
	Disposition *string    `db:"disposition" json:"disposition,omitempty"`
	Created     *time.Time `db:"created" json:"created,omitempty"`
	VersionID   int        `db:"version_id" json:"version_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func (r *CoverageEligibilityResponse) GetVersionID() int  { return r.VersionID }
func (r *CoverageEligibilityResponse) SetVersionID(v int) { r.VersionID = v }

func (r *CoverageEligibilityResponse) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "CoverageEligibilityResponse",
		"id":           r.FHIRID,
		"status":       r.Status,
		"outcome":      r.Outcome,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", r.PatientID.String())},
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", r.VersionID),
			LastUpdated: r.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/CoverageEligibilityResponse"},
		},
	}
	if r.RequestID != nil {
		result["request"] = fhir.Reference{Reference: fhir.FormatReference("CoverageEligibilityRequest", r.RequestID.String())}
	}
	if r.InsurerID != nil {
		result["insurer"] = fhir.Reference{Reference: fhir.FormatReference("Organization", r.InsurerID.String())}
	}
	if r.Disposition != nil {
		result["disposition"] = *r.Disposition
	}
	if r.Created != nil {
		result["created"] = r.Created.Format("2006-01-02T15:04:05Z")
	}
	return result
}
