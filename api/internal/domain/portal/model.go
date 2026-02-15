package portal

import (
	"encoding/json"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// PortalAccount maps to the portal_account table.
type PortalAccount struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	PatientID           uuid.UUID  `db:"patient_id" json:"patient_id"`
	Username            string     `db:"username" json:"username"`
	Email               string     `db:"email" json:"email"`
	Phone               *string    `db:"phone" json:"phone,omitempty"`
	Status              string     `db:"status" json:"status"`
	EmailVerified       bool       `db:"email_verified" json:"email_verified"`
	LastLoginAt         *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	FailedLoginCount    int        `db:"failed_login_count" json:"failed_login_count"`
	PasswordLastChanged *time.Time `db:"password_last_changed" json:"password_last_changed,omitempty"`
	MFAEnabled          bool       `db:"mfa_enabled" json:"mfa_enabled"`
	PreferredLanguage   *string    `db:"preferred_language" json:"preferred_language,omitempty"`
	Note                *string    `db:"note" json:"note,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// PortalProxyAccess maps to the portal_proxy_access table.
type PortalProxyAccess struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	PortalAccountID uuid.UUID  `db:"portal_account_id" json:"portal_account_id"`
	ProxyPatientID  uuid.UUID  `db:"proxy_patient_id" json:"proxy_patient_id"`
	Relationship    string     `db:"relationship" json:"relationship"`
	AccessLevel     string     `db:"access_level" json:"access_level"`
	Active          bool       `db:"active" json:"active"`
	PeriodStart     *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd       *time.Time `db:"period_end" json:"period_end,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

// PortalMessage maps to the portal_message table.
type PortalMessage struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	PatientID     uuid.UUID  `db:"patient_id" json:"patient_id"`
	PractitionerID *uuid.UUID `db:"practitioner_id" json:"practitioner_id,omitempty"`
	Direction     string     `db:"direction" json:"direction"`
	Subject       *string    `db:"subject" json:"subject,omitempty"`
	Body          string     `db:"body" json:"body"`
	Status        string     `db:"status" json:"status"`
	Priority      *string    `db:"priority" json:"priority,omitempty"`
	Category      *string    `db:"category" json:"category,omitempty"`
	ParentID      *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	ReadAt        *time.Time `db:"read_at" json:"read_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// Questionnaire maps to the questionnaire table (FHIR Questionnaire resource).
type Questionnaire struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FHIRID        string     `db:"fhir_id" json:"fhir_id"`
	Name          string     `db:"name" json:"name"`
	Title         *string    `db:"title" json:"title,omitempty"`
	Status        string     `db:"status" json:"status"`
	Version       *string    `db:"version" json:"version,omitempty"`
	Description   *string    `db:"description" json:"description,omitempty"`
	Purpose       *string    `db:"purpose" json:"purpose,omitempty"`
	SubjectType   *string    `db:"subject_type" json:"subject_type,omitempty"`
	Date          *time.Time `db:"date" json:"date,omitempty"`
	Publisher     *string    `db:"publisher" json:"publisher,omitempty"`
	ApprovalDate  *time.Time `db:"approval_date" json:"approval_date,omitempty"`
	LastReviewDate *time.Time `db:"last_review_date" json:"last_review_date,omitempty"`
	VersionID     int        `db:"version_id" json:"version_id"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (q *Questionnaire) GetVersionID() int { return q.VersionID }

// SetVersionID sets the current version.
func (q *Questionnaire) SetVersionID(v int) { q.VersionID = v }

func (q *Questionnaire) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Questionnaire",
		"id":           q.FHIRID,
		"name":         q.Name,
		"status":       q.Status,
		"meta":         fhir.Meta{LastUpdated: q.UpdatedAt},
	}
	if q.Title != nil {
		result["title"] = *q.Title
	}
	if q.Version != nil {
		result["version"] = *q.Version
	}
	if q.Description != nil {
		result["description"] = *q.Description
	}
	if q.Purpose != nil {
		result["purpose"] = *q.Purpose
	}
	if q.SubjectType != nil {
		result["subjectType"] = []string{*q.SubjectType}
	}
	if q.Date != nil {
		result["date"] = q.Date.Format("2006-01-02")
	}
	if q.Publisher != nil {
		result["publisher"] = *q.Publisher
	}
	if q.ApprovalDate != nil {
		result["approvalDate"] = q.ApprovalDate.Format("2006-01-02")
	}
	if q.LastReviewDate != nil {
		result["lastReviewDate"] = q.LastReviewDate.Format("2006-01-02")
	}
	return result
}

// QuestionnaireItem maps to the questionnaire_item table.
type QuestionnaireItem struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	QuestionnaireID uuid.UUID       `db:"questionnaire_id" json:"questionnaire_id"`
	LinkID          string          `db:"link_id" json:"link_id"`
	Text            string          `db:"text" json:"text"`
	Type            string          `db:"type" json:"type"`
	Required        bool            `db:"required" json:"required"`
	Repeats         bool            `db:"repeats" json:"repeats"`
	ReadOnly        bool            `db:"read_only" json:"read_only"`
	MaxLength       *int            `db:"max_length" json:"max_length,omitempty"`
	AnswerOptions   json.RawMessage `db:"answer_options" json:"answer_options,omitempty"`
	InitialValue    *string         `db:"initial_value" json:"initial_value,omitempty"`
	EnableWhenLinkID *string        `db:"enable_when_link_id" json:"enable_when_link_id,omitempty"`
	EnableWhenOperator *string      `db:"enable_when_operator" json:"enable_when_operator,omitempty"`
	EnableWhenAnswer *string        `db:"enable_when_answer" json:"enable_when_answer,omitempty"`
	SortOrder       int             `db:"sort_order" json:"sort_order"`
}

// QuestionnaireResponse maps to the questionnaire_response table (FHIR QuestionnaireResponse resource).
type QuestionnaireResponse struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	QuestionnaireID uuid.UUID  `db:"questionnaire_id" json:"questionnaire_id"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AuthorID        *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	Status          string     `db:"status" json:"status"`
	Authored        *time.Time `db:"authored" json:"authored,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (qr *QuestionnaireResponse) GetVersionID() int { return qr.VersionID }

// SetVersionID sets the current version.
func (qr *QuestionnaireResponse) SetVersionID(v int) { qr.VersionID = v }

func (qr *QuestionnaireResponse) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "QuestionnaireResponse",
		"id":           qr.FHIRID,
		"status":       qr.Status,
		"questionnaire": qr.QuestionnaireID.String(),
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", qr.PatientID.String())},
		"meta":         fhir.Meta{LastUpdated: qr.UpdatedAt},
	}
	if qr.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", qr.EncounterID.String())}
	}
	if qr.AuthorID != nil {
		result["author"] = fhir.Reference{Reference: fhir.FormatReference("Practitioner", qr.AuthorID.String())}
	}
	if qr.Authored != nil {
		result["authored"] = qr.Authored.Format(time.RFC3339)
	}
	return result
}

// QuestionnaireResponseItem maps to the questionnaire_response_item table.
type QuestionnaireResponseItem struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	ResponseID uuid.UUID  `db:"response_id" json:"response_id"`
	LinkID     string     `db:"link_id" json:"link_id"`
	Text       *string    `db:"text" json:"text,omitempty"`
	AnswerStr  *string    `db:"answer_string" json:"answer_string,omitempty"`
	AnswerInt  *int       `db:"answer_integer" json:"answer_integer,omitempty"`
	AnswerBool *bool      `db:"answer_boolean" json:"answer_boolean,omitempty"`
	AnswerDate *time.Time `db:"answer_date" json:"answer_date,omitempty"`
	AnswerCode *string    `db:"answer_code" json:"answer_code,omitempty"`
}

// PatientCheckin maps to the patient_checkin table.
type PatientCheckin struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	PatientID     uuid.UUID  `db:"patient_id" json:"patient_id"`
	AppointmentID *uuid.UUID `db:"appointment_id" json:"appointment_id,omitempty"`
	Status        string     `db:"status" json:"status"`
	CheckinMethod *string    `db:"checkin_method" json:"checkin_method,omitempty"`
	CheckinTime   *time.Time `db:"checkin_time" json:"checkin_time,omitempty"`
	InsuranceVerified *bool  `db:"insurance_verified" json:"insurance_verified,omitempty"`
	CoPayCollected *bool     `db:"co_pay_collected" json:"co_pay_collected,omitempty"`
	CoPayAmount   *float64   `db:"co_pay_amount" json:"co_pay_amount,omitempty"`
	Note          *string    `db:"note" json:"note,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
