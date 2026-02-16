package auditevent

import (
	"time"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type AuditEvent struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	FHIRID              string     `db:"fhir_id" json:"fhir_id"`
	TypeCode            string     `db:"type_code" json:"type_code"`
	TypeDisplay         string     `db:"type_display" json:"type_display"`
	SubtypeCode         string     `db:"subtype_code" json:"subtype_code"`
	SubtypeDisplay      string     `db:"subtype_display" json:"subtype_display"`
	Action              string     `db:"action" json:"action"`
	PeriodStart         *time.Time `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd           *time.Time `db:"period_end" json:"period_end,omitempty"`
	Recorded            time.Time  `db:"recorded" json:"recorded"`
	Outcome             string     `db:"outcome" json:"outcome"`
	OutcomeDesc         string     `db:"outcome_desc" json:"outcome_desc"`
	AgentTypeCode       string     `db:"agent_type_code" json:"agent_type_code"`
	AgentTypeDisplay    string     `db:"agent_type_display" json:"agent_type_display"`
	AgentWhoID          *uuid.UUID `db:"agent_who_id" json:"agent_who_id,omitempty"`
	AgentWhoDisplay     string     `db:"agent_who_display" json:"agent_who_display"`
	AgentAltID          string     `db:"agent_alt_id" json:"agent_alt_id"`
	AgentName           string     `db:"agent_name" json:"agent_name"`
	AgentRequestor      bool       `db:"agent_requestor" json:"agent_requestor"`
	AgentRoleCode       string     `db:"agent_role_code" json:"agent_role_code"`
	AgentRoleDisplay    string     `db:"agent_role_display" json:"agent_role_display"`
	AgentNetworkAddr    string     `db:"agent_network_address" json:"agent_network_address"`
	AgentNetworkType    string     `db:"agent_network_type" json:"agent_network_type"`
	SourceSite          string     `db:"source_site" json:"source_site"`
	SourceObserverID    string     `db:"source_observer_id" json:"source_observer_id"`
	SourceObsDisplay    string     `db:"source_observer_display" json:"source_observer_display"`
	SourceTypeCode      string     `db:"source_type_code" json:"source_type_code"`
	EntityWhatType      string     `db:"entity_what_type" json:"entity_what_type"`
	EntityWhatID        *uuid.UUID `db:"entity_what_id" json:"entity_what_id,omitempty"`
	EntityWhatDisp      string     `db:"entity_what_display" json:"entity_what_display"`
	EntityTypeCode      string     `db:"entity_type_code" json:"entity_type_code"`
	EntityRoleCode      string     `db:"entity_role_code" json:"entity_role_code"`
	EntityLifecycle     string     `db:"entity_lifecycle" json:"entity_lifecycle"`
	EntityName          string     `db:"entity_name" json:"entity_name"`
	EntityDesc          string     `db:"entity_description" json:"entity_description"`
	EntityQuery         string     `db:"entity_query" json:"entity_query"`
	PurposeCode         string     `db:"purpose_of_use_code" json:"purpose_of_use_code"`
	PurposeDisplay      string     `db:"purpose_of_use_display" json:"purpose_of_use_display"`
	SensitivityLabel    string     `db:"sensitivity_label" json:"sensitivity_label"`
	UserAgentString     string     `db:"user_agent_string" json:"user_agent_string"`
	SessionID           string     `db:"session_id" json:"session_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
}

func (a *AuditEvent) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "AuditEvent",
		"id":           a.FHIRID,
		"type": fhir.Coding{
			Code:    a.TypeCode,
			Display: a.TypeDisplay,
		},
		"action":   a.Action,
		"recorded": a.Recorded.Format("2006-01-02T15:04:05Z"),
		"outcome":  a.Outcome,
	}
	if a.SubtypeCode != "" {
		result["subtype"] = []fhir.Coding{{Code: a.SubtypeCode, Display: a.SubtypeDisplay}}
	}
	if a.PeriodStart != nil || a.PeriodEnd != nil {
		result["period"] = fhir.Period{Start: a.PeriodStart, End: a.PeriodEnd}
	}
	if a.OutcomeDesc != "" {
		result["outcomeDesc"] = a.OutcomeDesc
	}
	agent := map[string]interface{}{
		"requestor": a.AgentRequestor,
	}
	if a.AgentTypeCode != "" {
		agent["type"] = fhir.CodeableConcept{Coding: []fhir.Coding{{Code: a.AgentTypeCode, Display: a.AgentTypeDisplay}}}
	}
	if a.AgentWhoID != nil {
		agent["who"] = fhir.Reference{Reference: a.AgentWhoDisplay, Display: a.AgentWhoDisplay}
	}
	if a.AgentName != "" {
		agent["name"] = a.AgentName
	}
	if a.AgentRoleCode != "" {
		agent["role"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: a.AgentRoleCode, Display: a.AgentRoleDisplay}}}}
	}
	if a.AgentNetworkAddr != "" {
		agent["network"] = map[string]interface{}{
			"address": a.AgentNetworkAddr,
			"type":    a.AgentNetworkType,
		}
	}
	result["agent"] = []interface{}{agent}

	source := map[string]interface{}{}
	if a.SourceSite != "" {
		source["site"] = a.SourceSite
	}
	if a.SourceObserverID != "" {
		source["observer"] = fhir.Reference{Reference: a.SourceObserverID, Display: a.SourceObsDisplay}
	}
	if a.SourceTypeCode != "" {
		source["type"] = []fhir.Coding{{Code: a.SourceTypeCode}}
	}
	if len(source) > 0 {
		result["source"] = source
	}

	if a.EntityWhatType != "" || a.EntityWhatID != nil {
		entity := map[string]interface{}{}
		if a.EntityWhatID != nil {
			ref := a.EntityWhatType + "/" + a.EntityWhatID.String()
			entity["what"] = fhir.Reference{Reference: ref, Display: a.EntityWhatDisp}
		}
		if a.EntityTypeCode != "" {
			entity["type"] = fhir.Coding{Code: a.EntityTypeCode}
		}
		if a.EntityRoleCode != "" {
			entity["role"] = fhir.Coding{Code: a.EntityRoleCode}
		}
		if a.EntityLifecycle != "" {
			entity["lifecycle"] = fhir.Coding{Code: a.EntityLifecycle}
		}
		if a.EntityName != "" {
			entity["name"] = a.EntityName
		}
		if a.EntityDesc != "" {
			entity["description"] = a.EntityDesc
		}
		result["entity"] = []interface{}{entity}
	}

	if a.PurposeCode != "" {
		result["purposeOfEvent"] = []fhir.CodeableConcept{{Coding: []fhir.Coding{{Code: a.PurposeCode, Display: a.PurposeDisplay}}}}
	}

	return result
}
