package conformance

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

// NamingSystem maps to the naming_system table (FHIR NamingSystem resource).
type NamingSystem struct {
	ID           uuid.UUID `db:"id" json:"id"`
	FHIRID       string    `db:"fhir_id" json:"fhir_id"`
	Name         string    `db:"name" json:"name"`
	Status       string    `db:"status" json:"status"`
	Kind         string    `db:"kind" json:"kind"`
	Date         *string   `db:"date" json:"date,omitempty"`
	Publisher    *string   `db:"publisher" json:"publisher,omitempty"`
	Responsible  *string   `db:"responsible" json:"responsible,omitempty"`
	TypeCode     *string   `db:"type_code" json:"type_code,omitempty"`
	TypeDisplay  *string   `db:"type_display" json:"type_display,omitempty"`
	Description  *string   `db:"description" json:"description,omitempty"`
	UsageNote    *string   `db:"usage_note" json:"usage_note,omitempty"`
	Jurisdiction *string   `db:"jurisdiction" json:"jurisdiction,omitempty"`
	VersionID    int       `db:"version_id" json:"version_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (ns *NamingSystem) GetVersionID() int { return ns.VersionID }

// SetVersionID sets the current version.
func (ns *NamingSystem) SetVersionID(v int) { ns.VersionID = v }

func (ns *NamingSystem) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "NamingSystem",
		"id":           ns.FHIRID,
		"name":         ns.Name,
		"status":       ns.Status,
		"kind":         ns.Kind,
		"meta":         fhir.Meta{LastUpdated: ns.UpdatedAt},
	}
	if ns.Date != nil {
		result["date"] = *ns.Date
	}
	if ns.Publisher != nil {
		result["publisher"] = *ns.Publisher
	}
	if ns.Responsible != nil {
		result["responsible"] = *ns.Responsible
	}
	if ns.TypeCode != nil {
		result["type"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    *ns.TypeCode,
				Display: strVal(ns.TypeDisplay),
			}},
		}
	}
	if ns.Description != nil {
		result["description"] = *ns.Description
	}
	if ns.UsageNote != nil {
		result["usage"] = *ns.UsageNote
	}
	if ns.Jurisdiction != nil {
		result["jurisdiction"] = []fhir.CodeableConcept{{
			Coding: []fhir.Coding{{Code: *ns.Jurisdiction}},
		}}
	}
	return result
}

// NamingSystemUniqueID maps to the naming_system_unique_id table.
type NamingSystemUniqueID struct {
	ID             uuid.UUID `db:"id" json:"id"`
	NamingSystemID uuid.UUID `db:"naming_system_id" json:"naming_system_id"`
	Type           string    `db:"type" json:"type"`
	Value          string    `db:"value" json:"value"`
	Preferred      *bool     `db:"preferred" json:"preferred,omitempty"`
	Comment        *string   `db:"comment" json:"comment,omitempty"`
	PeriodStart    *string   `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd      *string   `db:"period_end" json:"period_end,omitempty"`
}

// OperationDefinition maps to the operation_definition table (FHIR OperationDefinition resource).
type OperationDefinition struct {
	ID            uuid.UUID `db:"id" json:"id"`
	FHIRID        string    `db:"fhir_id" json:"fhir_id"`
	URL           *string   `db:"url" json:"url,omitempty"`
	Name          string    `db:"name" json:"name"`
	Title         *string   `db:"title" json:"title,omitempty"`
	Status        string    `db:"status" json:"status"`
	Kind          string    `db:"kind" json:"kind"`
	Description   *string   `db:"description" json:"description,omitempty"`
	Code          string    `db:"code" json:"code"`
	System        *bool     `db:"system" json:"system,omitempty"`
	Type          *bool     `db:"type" json:"type,omitempty"`
	Instance      *bool     `db:"instance" json:"instance,omitempty"`
	InputProfile  *string   `db:"input_profile" json:"input_profile,omitempty"`
	OutputProfile *string   `db:"output_profile" json:"output_profile,omitempty"`
	Publisher     *string   `db:"publisher" json:"publisher,omitempty"`
	VersionID     int       `db:"version_id" json:"version_id"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (od *OperationDefinition) GetVersionID() int { return od.VersionID }

// SetVersionID sets the current version.
func (od *OperationDefinition) SetVersionID(v int) { od.VersionID = v }

func (od *OperationDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "OperationDefinition",
		"id":           od.FHIRID,
		"name":         od.Name,
		"status":       od.Status,
		"kind":         od.Kind,
		"code":         od.Code,
		"meta":         fhir.Meta{LastUpdated: od.UpdatedAt},
	}
	if od.URL != nil {
		result["url"] = *od.URL
	}
	if od.Title != nil {
		result["title"] = *od.Title
	}
	if od.Description != nil {
		result["description"] = *od.Description
	}
	if od.System != nil {
		result["system"] = *od.System
	}
	if od.Type != nil {
		result["type"] = *od.Type
	}
	if od.Instance != nil {
		result["instance"] = *od.Instance
	}
	if od.InputProfile != nil {
		result["inputProfile"] = *od.InputProfile
	}
	if od.OutputProfile != nil {
		result["outputProfile"] = *od.OutputProfile
	}
	if od.Publisher != nil {
		result["publisher"] = *od.Publisher
	}
	return result
}

// OperationDefinitionParameter maps to the operation_definition_parameter table.
type OperationDefinitionParameter struct {
	ID                    uuid.UUID `db:"id" json:"id"`
	OperationDefinitionID uuid.UUID `db:"operation_definition_id" json:"operation_definition_id"`
	Name                  string    `db:"name" json:"name"`
	Use                   string    `db:"use" json:"use"`
	MinVal                int       `db:"min_val" json:"min_val"`
	MaxVal                *string   `db:"max_val" json:"max_val,omitempty"`
	Documentation         *string   `db:"documentation" json:"documentation,omitempty"`
	Type                  *string   `db:"type" json:"type,omitempty"`
	SearchType            *string   `db:"search_type" json:"search_type,omitempty"`
}

// MessageDefinition maps to the message_definition table (FHIR MessageDefinition resource).
type MessageDefinition struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	FHIRID             string    `db:"fhir_id" json:"fhir_id"`
	URL                *string   `db:"url" json:"url,omitempty"`
	Name               *string   `db:"name" json:"name,omitempty"`
	Title              *string   `db:"title" json:"title,omitempty"`
	Status             string    `db:"status" json:"status"`
	Date               *string   `db:"date" json:"date,omitempty"`
	Publisher          *string   `db:"publisher" json:"publisher,omitempty"`
	Description        *string   `db:"description" json:"description,omitempty"`
	Purpose            *string   `db:"purpose" json:"purpose,omitempty"`
	EventCodingCode    string    `db:"event_coding_code" json:"event_coding_code"`
	EventCodingSystem  *string   `db:"event_coding_system" json:"event_coding_system,omitempty"`
	EventCodingDisplay *string   `db:"event_coding_display" json:"event_coding_display,omitempty"`
	Category           *string   `db:"category" json:"category,omitempty"`
	ResponseRequired   *string   `db:"response_required" json:"response_required,omitempty"`
	VersionID          int       `db:"version_id" json:"version_id"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (md *MessageDefinition) GetVersionID() int { return md.VersionID }

// SetVersionID sets the current version.
func (md *MessageDefinition) SetVersionID(v int) { md.VersionID = v }

func (md *MessageDefinition) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MessageDefinition",
		"id":           md.FHIRID,
		"status":       md.Status,
		"eventCoding": fhir.Coding{
			System:  strVal(md.EventCodingSystem),
			Code:    md.EventCodingCode,
			Display: strVal(md.EventCodingDisplay),
		},
		"meta": fhir.Meta{LastUpdated: md.UpdatedAt},
	}
	if md.URL != nil {
		result["url"] = *md.URL
	}
	if md.Name != nil {
		result["name"] = *md.Name
	}
	if md.Title != nil {
		result["title"] = *md.Title
	}
	if md.Date != nil {
		result["date"] = *md.Date
	}
	if md.Publisher != nil {
		result["publisher"] = *md.Publisher
	}
	if md.Description != nil {
		result["description"] = *md.Description
	}
	if md.Purpose != nil {
		result["purpose"] = *md.Purpose
	}
	if md.Category != nil {
		result["category"] = *md.Category
	}
	if md.ResponseRequired != nil {
		result["responseRequired"] = *md.ResponseRequired
	}
	return result
}

// MessageHeader maps to the message_header table (FHIR MessageHeader resource).
type MessageHeader struct {
	ID                  uuid.UUID `db:"id" json:"id"`
	FHIRID              string    `db:"fhir_id" json:"fhir_id"`
	EventCodingCode     string    `db:"event_coding_code" json:"event_coding_code"`
	EventCodingSystem   *string   `db:"event_coding_system" json:"event_coding_system,omitempty"`
	EventCodingDisplay  *string   `db:"event_coding_display" json:"event_coding_display,omitempty"`
	DestinationName     *string   `db:"destination_name" json:"destination_name,omitempty"`
	DestinationEndpoint *string   `db:"destination_endpoint" json:"destination_endpoint,omitempty"`
	SenderOrgID         *uuid.UUID `db:"sender_org_id" json:"sender_org_id,omitempty"`
	SourceName          *string   `db:"source_name" json:"source_name,omitempty"`
	SourceEndpoint      string    `db:"source_endpoint" json:"source_endpoint"`
	SourceSoftware      *string   `db:"source_software" json:"source_software,omitempty"`
	SourceVersion       *string   `db:"source_version" json:"source_version,omitempty"`
	ReasonCode          *string   `db:"reason_code" json:"reason_code,omitempty"`
	ReasonDisplay       *string   `db:"reason_display" json:"reason_display,omitempty"`
	ResponseIdentifier  *string   `db:"response_identifier" json:"response_identifier,omitempty"`
	ResponseCode        *string   `db:"response_code" json:"response_code,omitempty"`
	FocusReference      *string   `db:"focus_reference" json:"focus_reference,omitempty"`
	DefinitionURL       *string   `db:"definition_url" json:"definition_url,omitempty"`
	VersionID           int       `db:"version_id" json:"version_id"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// GetVersionID returns the current version.
func (mh *MessageHeader) GetVersionID() int { return mh.VersionID }

// SetVersionID sets the current version.
func (mh *MessageHeader) SetVersionID(v int) { mh.VersionID = v }

func (mh *MessageHeader) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "MessageHeader",
		"id":           mh.FHIRID,
		"eventCoding": fhir.Coding{
			System:  strVal(mh.EventCodingSystem),
			Code:    mh.EventCodingCode,
			Display: strVal(mh.EventCodingDisplay),
		},
		"source": map[string]interface{}{
			"endpoint": mh.SourceEndpoint,
		},
		"meta": fhir.Meta{LastUpdated: mh.UpdatedAt},
	}
	if mh.SourceName != nil {
		result["source"].(map[string]interface{})["name"] = *mh.SourceName
	}
	if mh.SourceSoftware != nil {
		result["source"].(map[string]interface{})["software"] = *mh.SourceSoftware
	}
	if mh.SourceVersion != nil {
		result["source"].(map[string]interface{})["version"] = *mh.SourceVersion
	}
	if mh.DestinationName != nil || mh.DestinationEndpoint != nil {
		dest := map[string]string{}
		if mh.DestinationName != nil {
			dest["name"] = *mh.DestinationName
		}
		if mh.DestinationEndpoint != nil {
			dest["endpoint"] = *mh.DestinationEndpoint
		}
		result["destination"] = []map[string]string{dest}
	}
	if mh.SenderOrgID != nil {
		result["sender"] = fhir.Reference{Reference: fhir.FormatReference("Organization", mh.SenderOrgID.String())}
	}
	if mh.ReasonCode != nil {
		result["reason"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{
				Code:    *mh.ReasonCode,
				Display: strVal(mh.ReasonDisplay),
			}},
		}
	}
	if mh.ResponseIdentifier != nil {
		resp := map[string]string{"identifier": *mh.ResponseIdentifier}
		if mh.ResponseCode != nil {
			resp["code"] = *mh.ResponseCode
		}
		result["response"] = resp
	}
	if mh.FocusReference != nil {
		result["focus"] = []fhir.Reference{{Reference: *mh.FocusReference}}
	}
	if mh.DefinitionURL != nil {
		result["definition"] = *mh.DefinitionURL
	}
	return result
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
