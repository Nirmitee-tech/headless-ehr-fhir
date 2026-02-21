package testreport

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// TestReport maps to the test_report table (FHIR TestReport resource).
type TestReport struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	Status              string     `db:"status" json:"status"`
	Name                *string    `db:"name" json:"name,omitempty"`
	TestScriptReference *string    `db:"test_script_reference" json:"test_script_reference,omitempty"`
	Result              string     `db:"result" json:"result"`
	Score               *float64   `db:"score" json:"score,omitempty"`
	Tester              *string    `db:"tester" json:"tester,omitempty"`
	Issued              *time.Time `db:"issued" json:"issued,omitempty"`
	ParticipantType     *string    `db:"participant_type" json:"participant_type,omitempty"`
	ParticipantURI      *string    `db:"participant_uri" json:"participant_uri,omitempty"`
	VersionID           int        `db:"version_id" json:"version_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

func (e *TestReport) GetVersionID() int  { return e.VersionID }
func (e *TestReport) SetVersionID(v int) { e.VersionID = v }

func (e *TestReport) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "TestReport",
		"id":           e.FHIRID,
		"status":       e.Status,
		"result":       e.Result,
		"meta":         fhir.Meta{
			VersionID:   fmt.Sprintf("%d", e.VersionID),
			LastUpdated: e.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/TestReport"},
		},
	}
	if e.Name != nil {
		result["name"] = *e.Name
	}
	if e.TestScriptReference != nil {
		result["testScript"] = fhir.Reference{Reference: *e.TestScriptReference}
	}
	if e.Score != nil {
		result["score"] = *e.Score
	}
	if e.Tester != nil {
		result["tester"] = *e.Tester
	}
	if e.Issued != nil {
		result["issued"] = e.Issued.Format(time.RFC3339)
	}
	if e.ParticipantType != nil || e.ParticipantURI != nil {
		participant := map[string]interface{}{}
		if e.ParticipantType != nil {
			participant["type"] = *e.ParticipantType
		}
		if e.ParticipantURI != nil {
			participant["uri"] = *e.ParticipantURI
		}
		result["participant"] = []interface{}{participant}
	}
	return result
}
