package verificationresult

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// VerificationResult maps to the verification_result table (FHIR VerificationResult resource).
type VerificationResult struct {
	ID                       uuid.UUID  `db:"id" json:"id"`
	FHIRID                   string     `db:"fhir_id" json:"fhir_id"`
	Status                   string     `db:"status" json:"status"`
	TargetType               *string    `db:"target_type" json:"target_type,omitempty"`
	TargetReference          *string    `db:"target_reference" json:"target_reference,omitempty"`
	NeedCode                 *string    `db:"need_code" json:"need_code,omitempty"`
	NeedDisplay              *string    `db:"need_display" json:"need_display,omitempty"`
	StatusDate               *time.Time `db:"status_date" json:"status_date,omitempty"`
	ValidationTypeCode       *string    `db:"validation_type_code" json:"validation_type_code,omitempty"`
	ValidationTypeDisplay    *string    `db:"validation_type_display" json:"validation_type_display,omitempty"`
	ValidationProcessCode    *string    `db:"validation_process_code" json:"validation_process_code,omitempty"`
	ValidationProcessDisplay *string    `db:"validation_process_display" json:"validation_process_display,omitempty"`
	FrequencyValue           *int       `db:"frequency_value" json:"frequency_value,omitempty"`
	FrequencyUnit            *string    `db:"frequency_unit" json:"frequency_unit,omitempty"`
	LastPerformed            *time.Time `db:"last_performed" json:"last_performed,omitempty"`
	NextScheduled            *time.Time `db:"next_scheduled" json:"next_scheduled,omitempty"`
	FailureActionCode        *string    `db:"failure_action_code" json:"failure_action_code,omitempty"`
	FailureActionDisplay     *string    `db:"failure_action_display" json:"failure_action_display,omitempty"`
	VersionID                int        `db:"version_id" json:"version_id"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

func (v *VerificationResult) GetVersionID() int  { return v.VersionID }
func (v *VerificationResult) SetVersionID(ver int) { v.VersionID = ver }

func (v *VerificationResult) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "VerificationResult",
		"id":           v.FHIRID,
		"status":       v.Status,
		"meta":         fhir.Meta{LastUpdated: v.UpdatedAt},
	}
	if v.TargetType != nil && v.TargetReference != nil {
		result["target"] = []fhir.Reference{
			{Reference: fhir.FormatReference(*v.TargetType, *v.TargetReference)},
		}
	}
	if v.NeedCode != nil {
		result["need"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *v.NeedCode, Display: strVal(v.NeedDisplay)}},
		}
	}
	if v.StatusDate != nil {
		result["statusDate"] = v.StatusDate.Format("2006-01-02T15:04:05Z")
	}
	if v.ValidationTypeCode != nil {
		result["validationType"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *v.ValidationTypeCode, Display: strVal(v.ValidationTypeDisplay)}},
		}
	}
	if v.ValidationProcessCode != nil {
		result["validationProcess"] = []fhir.CodeableConcept{
			{Coding: []fhir.Coding{{Code: *v.ValidationProcessCode, Display: strVal(v.ValidationProcessDisplay)}}},
		}
	}
	if v.FrequencyValue != nil {
		timing := map[string]interface{}{
			"repeat": map[string]interface{}{},
		}
		repeat := timing["repeat"].(map[string]interface{})
		repeat["frequency"] = *v.FrequencyValue
		if v.FrequencyUnit != nil {
			repeat["periodUnit"] = *v.FrequencyUnit
		}
		result["frequency"] = timing
	}
	if v.LastPerformed != nil {
		result["lastPerformed"] = v.LastPerformed.Format("2006-01-02T15:04:05Z")
	}
	if v.NextScheduled != nil {
		result["nextScheduled"] = v.NextScheduled.Format("2006-01-02")
	}
	if v.FailureActionCode != nil {
		result["failureAction"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *v.FailureActionCode, Display: strVal(v.FailureActionDisplay)}},
		}
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
