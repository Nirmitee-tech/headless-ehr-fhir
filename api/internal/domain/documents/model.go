package documents

import (
	"fmt"
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// Consent maps to the consent table (FHIR Consent resource).
type Consent struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	Scope           *string    `db:"scope" json:"scope,omitempty"`
	CategoryCode    *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay *string    `db:"category_display" json:"category_display,omitempty"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	PerformerID     *uuid.UUID `db:"performer_id" json:"performer_id,omitempty"`
	OrganizationID  *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	PolicyAuthority *string    `db:"policy_authority" json:"policy_authority,omitempty"`
	PolicyURI       *string    `db:"policy_uri" json:"policy_uri,omitempty"`
	ProvisionType   *string    `db:"provision_type" json:"provision_type,omitempty"`
	ProvisionStart  *time.Time `db:"provision_start" json:"provision_start,omitempty"`
	ProvisionEnd    *time.Time `db:"provision_end" json:"provision_end,omitempty"`
	ProvisionAction *string    `db:"provision_action" json:"provision_action,omitempty"`
	HIPAAAuth       *bool      `db:"hipaa_authorization" json:"hipaa_authorization,omitempty"`
	ABDMConsent     *bool      `db:"abdm_consent" json:"abdm_consent,omitempty"`
	ABDMConsentID   *string    `db:"abdm_consent_id" json:"abdm_consent_id,omitempty"`
	SignatureType   *string    `db:"signature_type" json:"signature_type,omitempty"`
	SignatureWhen   *time.Time `db:"signature_when" json:"signature_when,omitempty"`
	SignatureData   *string    `db:"signature_data" json:"signature_data,omitempty"`
	DateTime        *time.Time `db:"date_time" json:"date_time,omitempty"`
	Note            *string    `db:"note" json:"note,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (c *Consent) GetVersionID() int  { return c.VersionID }
func (c *Consent) SetVersionID(v int) { c.VersionID = v }

func (c *Consent) ToFHIR() map[string]interface{} {
	versionID := c.VersionID
	if versionID == 0 {
		versionID = 1
	}
	result := map[string]interface{}{
		"resourceType": "Consent",
		"id":           c.FHIRID,
		"status":       c.Status,
		"patient":      fhir.Reference{Reference: fhir.FormatReference("Patient", c.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", versionID),
			LastUpdated: c.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Consent"},
		},
	}
	if c.Scope != nil {
		result["scope"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *c.Scope}},
		}
	}
	if c.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *c.CategoryCode, Display: strVal(c.CategoryDisplay)}},
		}}
	}
	if c.PerformerID != nil {
		result["performer"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", c.PerformerID.String())}}
	}
	if c.OrganizationID != nil {
		result["organization"] = []fhir.Reference{{Reference: fhir.FormatReference("Organization", c.OrganizationID.String())}}
	}
	if c.PolicyURI != nil {
		result["policy"] = []map[string]string{{"uri": *c.PolicyURI}}
	}
	if c.ProvisionType != nil {
		provision := map[string]interface{}{
			"type": *c.ProvisionType,
		}
		if c.ProvisionStart != nil || c.ProvisionEnd != nil {
			provision["period"] = fhir.Period{Start: c.ProvisionStart, End: c.ProvisionEnd}
		}
		if c.ProvisionAction != nil {
			provision["action"] = []fhir.CodeableConcept{{
				Coding: []fhir.Coding{{Code: *c.ProvisionAction}},
			}}
		}
		result["provision"] = provision
	}
	if c.DateTime != nil {
		result["dateTime"] = c.DateTime.Format(time.RFC3339)
	}
	return result
}

// DocumentReference maps to the document_reference table (FHIR DocumentReference resource).
type DocumentReference struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	DocStatus       *string    `db:"doc_status" json:"doc_status,omitempty"`
	TypeCode        *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay     *string    `db:"type_display" json:"type_display,omitempty"`
	CategoryCode    *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay *string    `db:"category_display" json:"category_display,omitempty"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	AuthorID        *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	CustodianID     *uuid.UUID `db:"custodian_id" json:"custodian_id,omitempty"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	Date            *time.Time `db:"date" json:"date,omitempty"`
	Description     *string    `db:"description" json:"description,omitempty"`
	SecurityLabel   *string    `db:"security_label" json:"security_label,omitempty"`
	ContentType     *string    `db:"content_type" json:"content_type,omitempty"`
	ContentURL      *string    `db:"content_url" json:"content_url,omitempty"`
	ContentSize     *int       `db:"content_size" json:"content_size,omitempty"`
	ContentHash     *string    `db:"content_hash" json:"content_hash,omitempty"`
	ContentTitle    *string    `db:"content_title" json:"content_title,omitempty"`
	FormatCode      *string    `db:"format_code" json:"format_code,omitempty"`
	FormatDisplay   *string    `db:"format_display" json:"format_display,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (d *DocumentReference) GetVersionID() int  { return d.VersionID }
func (d *DocumentReference) SetVersionID(v int) { d.VersionID = v }

func (d *DocumentReference) ToFHIR() map[string]interface{} {
	versionID := d.VersionID
	if versionID == 0 {
		versionID = 1
	}
	result := map[string]interface{}{
		"resourceType": "DocumentReference",
		"id":           d.FHIRID,
		"status":       d.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", d.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", versionID),
			LastUpdated: d.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-documentreference"},
		},
	}
	if d.DocStatus != nil {
		result["docStatus"] = *d.DocStatus
	}
	if d.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *d.TypeCode, Display: strVal(d.TypeDisplay)}},
		}
	}
	if d.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *d.CategoryCode, Display: strVal(d.CategoryDisplay)}},
		}}
	}
	if d.AuthorID != nil {
		result["author"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", d.AuthorID.String())}}
	}
	if d.CustodianID != nil {
		result["custodian"] = fhir.Reference{Reference: fhir.FormatReference("Organization", d.CustodianID.String())}
	}
	if d.EncounterID != nil {
		result["context"] = map[string]interface{}{
			"encounter": []fhir.Reference{{Reference: fhir.FormatReference("Encounter", d.EncounterID.String())}},
		}
	}
	if d.Date != nil {
		result["date"] = d.Date.Format(time.RFC3339)
	}
	if d.Description != nil {
		result["description"] = *d.Description
	}
	if d.SecurityLabel != nil {
		result["securityLabel"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *d.SecurityLabel}},
		}}
	}
	content := map[string]interface{}{}
	hasContent := false
	if d.ContentType != nil || d.ContentURL != nil {
		attachment := map[string]interface{}{}
		if d.ContentType != nil {
			attachment["contentType"] = *d.ContentType
		}
		if d.ContentURL != nil {
			attachment["url"] = *d.ContentURL
		}
		if d.ContentSize != nil {
			attachment["size"] = *d.ContentSize
		}
		if d.ContentHash != nil {
			attachment["hash"] = *d.ContentHash
		}
		if d.ContentTitle != nil {
			attachment["title"] = *d.ContentTitle
		}
		content["attachment"] = attachment
		hasContent = true
	}
	if d.FormatCode != nil {
		content["format"] = fhir.Coding{Code: *d.FormatCode, Display: strVal(d.FormatDisplay)}
		hasContent = true
	}
	if hasContent {
		result["content"] = []interface{}{content}
	}
	return result
}

// ClinicalNote maps to the clinical_note table (structured clinical notes).
type ClinicalNote struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	PatientID     uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID   *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	AuthorID      uuid.UUID  `db:"author_id" json:"author_id"`
	NoteType      string     `db:"note_type" json:"note_type"`
	Status        string     `db:"status" json:"status"`
	Title         *string    `db:"title" json:"title,omitempty"`
	Subjective    *string    `db:"subjective" json:"subjective,omitempty"`
	Objective     *string    `db:"objective" json:"objective,omitempty"`
	Assessment    *string    `db:"assessment" json:"assessment,omitempty"`
	Plan          *string    `db:"plan" json:"plan,omitempty"`
	NoteText      *string    `db:"note_text" json:"note_text,omitempty"`
	SignedBy      *uuid.UUID `db:"signed_by" json:"signed_by,omitempty"`
	SignedAt      *time.Time `db:"signed_at" json:"signed_at,omitempty"`
	CosignedBy    *uuid.UUID `db:"cosigned_by" json:"cosigned_by,omitempty"`
	CosignedAt    *time.Time `db:"cosigned_at" json:"cosigned_at,omitempty"`
	AmendedBy     *uuid.UUID `db:"amended_by" json:"amended_by,omitempty"`
	AmendedAt     *time.Time `db:"amended_at" json:"amended_at,omitempty"`
	AmendedReason *string    `db:"amended_reason" json:"amended_reason,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// Composition maps to the composition table (FHIR Composition resource).
type Composition struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	FHIRID          string     `db:"fhir_id" json:"fhir_id"`
	Status          string     `db:"status" json:"status"`
	TypeCode        *string    `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay     *string    `db:"type_display" json:"type_display,omitempty"`
	CategoryCode    *string    `db:"category_code" json:"category_code,omitempty"`
	CategoryDisplay *string    `db:"category_display" json:"category_display,omitempty"`
	PatientID       uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID     *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	Date            *time.Time `db:"date" json:"date,omitempty"`
	AuthorID        *uuid.UUID `db:"author_id" json:"author_id,omitempty"`
	Title           *string    `db:"title" json:"title,omitempty"`
	Confidentiality *string    `db:"confidentiality" json:"confidentiality,omitempty"`
	CustodianID     *uuid.UUID `db:"custodian_id" json:"custodian_id,omitempty"`
	VersionID       int        `db:"version_id" json:"version_id"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (comp *Composition) GetVersionID() int  { return comp.VersionID }
func (comp *Composition) SetVersionID(v int) { comp.VersionID = v }

func (comp *Composition) ToFHIR() map[string]interface{} {
	versionID := comp.VersionID
	if versionID == 0 {
		versionID = 1
	}
	result := map[string]interface{}{
		"resourceType": "Composition",
		"id":           comp.FHIRID,
		"status":       comp.Status,
		"subject":      fhir.Reference{Reference: fhir.FormatReference("Patient", comp.PatientID.String())},
		"meta": fhir.Meta{
			VersionID:   fmt.Sprintf("%d", versionID),
			LastUpdated: comp.UpdatedAt,
			Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Composition"},
		},
	}
	if comp.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *comp.TypeCode, Display: strVal(comp.TypeDisplay)}},
		}
	}
	if comp.CategoryCode != nil {
		result["category"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *comp.CategoryCode, Display: strVal(comp.CategoryDisplay)}},
		}}
	}
	if comp.EncounterID != nil {
		result["encounter"] = fhir.Reference{Reference: fhir.FormatReference("Encounter", comp.EncounterID.String())}
	}
	if comp.Date != nil {
		result["date"] = comp.Date.Format(time.RFC3339)
	}
	if comp.AuthorID != nil {
		result["author"] = []fhir.Reference{{Reference: fhir.FormatReference("Practitioner", comp.AuthorID.String())}}
	}
	if comp.Title != nil {
		result["title"] = *comp.Title
	}
	if comp.Confidentiality != nil {
		result["confidentiality"] = *comp.Confidentiality
	}
	if comp.CustodianID != nil {
		result["custodian"] = fhir.Reference{Reference: fhir.FormatReference("Organization", comp.CustodianID.String())}
	}
	return result
}

// CompositionSection maps to the composition_section table.
type CompositionSection struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	CompositionID   uuid.UUID  `db:"composition_id" json:"composition_id"`
	Title           *string    `db:"title" json:"title,omitempty"`
	CodeValue       *string    `db:"code_value" json:"code_value,omitempty"`
	CodeDisplay     *string    `db:"code_display" json:"code_display,omitempty"`
	TextStatus      *string    `db:"text_status" json:"text_status,omitempty"`
	TextDiv         *string    `db:"text_div" json:"text_div,omitempty"`
	Mode            *string    `db:"mode" json:"mode,omitempty"`
	OrderedBy       *string    `db:"ordered_by" json:"ordered_by,omitempty"`
	EntryReference  *string    `db:"entry_reference" json:"entry_reference,omitempty"`
	SortOrder       *int       `db:"sort_order" json:"sort_order,omitempty"`
}

// DocumentTemplate maps to the document_template table.
type DocumentTemplate struct {
	ID          uuid.UUID         `db:"id" json:"id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description,omitempty"`
	Status      string            `db:"status" json:"status"`
	TypeCode    *string           `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay *string           `db:"type_display" json:"type_display,omitempty"`
	Sections    []TemplateSection `json:"sections,omitempty"` // populated from separate table
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	CreatedBy   string            `db:"created_by" json:"created_by"`
}

// TemplateSection maps to the template_section table.
type TemplateSection struct {
	ID              uuid.UUID `db:"id" json:"id"`
	TemplateID      uuid.UUID `db:"template_id" json:"template_id"`
	Title           string    `db:"title" json:"title"`
	SortOrder       int       `db:"sort_order" json:"sort_order"`
	ContentTemplate string    `db:"content_template" json:"content_template"`
	Required        bool      `db:"required" json:"required"`
}

// RenderedDocument represents a template rendered with variable substitution.
type RenderedDocument struct {
	TemplateID   uuid.UUID         `json:"template_id"`
	TemplateName string            `json:"template_name"`
	Sections     []RenderedSection `json:"sections"`
	RenderedAt   time.Time         `json:"rendered_at"`
}

// RenderedSection represents a single section after template rendering.
type RenderedSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
