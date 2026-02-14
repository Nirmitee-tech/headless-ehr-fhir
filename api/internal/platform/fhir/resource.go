package fhir

import (
	"time"
)

// Resource is the base FHIR resource representation.
type Resource struct {
	ResourceType string `json:"resourceType"`
	ID           string `json:"id"`
	Meta         *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	VersionID   string    `json:"versionId,omitempty"`
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
	Profile     []string  `json:"profile,omitempty"`
}

type Coding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

type Reference struct {
	Reference string `json:"reference,omitempty"`
	Type      string `json:"type,omitempty"`
	Display   string `json:"display,omitempty"`
}

type Identifier struct {
	Use    string           `json:"use,omitempty"`
	Type   *CodeableConcept `json:"type,omitempty"`
	System string           `json:"system,omitempty"`
	Value  string           `json:"value,omitempty"`
	Period *Period          `json:"period,omitempty"`
}

type HumanName struct {
	Use    string   `json:"use,omitempty"`
	Family string   `json:"family,omitempty"`
	Given  []string `json:"given,omitempty"`
	Prefix []string `json:"prefix,omitempty"`
	Suffix []string `json:"suffix,omitempty"`
}

type Address struct {
	Use        string   `json:"use,omitempty"`
	Type       string   `json:"type,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	District   string   `json:"district,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
	Country    string   `json:"country,omitempty"`
}

type ContactPoint struct {
	System string `json:"system,omitempty"`
	Value  string `json:"value,omitempty"`
	Use    string `json:"use,omitempty"`
	Rank   int    `json:"rank,omitempty"`
}

type Period struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

type Extension struct {
	URL          string `json:"url"`
	ValueString  string `json:"valueString,omitempty"`
	ValueCode    string `json:"valueCode,omitempty"`
	ValueBoolean *bool  `json:"valueBoolean,omitempty"`
	ValueInteger *int   `json:"valueInteger,omitempty"`
}

// OperationOutcome represents a FHIR OperationOutcome for errors.
type OperationOutcome struct {
	ResourceType string               `json:"resourceType"`
	Issue        []OperationOutcomeIssue `json:"issue"`
}

type OperationOutcomeIssue struct {
	Severity    string           `json:"severity"`
	Code        string           `json:"code"`
	Details     *CodeableConcept `json:"details,omitempty"`
	Diagnostics string           `json:"diagnostics,omitempty"`
	Expression  []string         `json:"expression,omitempty"`
}

func NewOperationOutcome(severity, code, diagnostics string) *OperationOutcome {
	return &OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []OperationOutcomeIssue{
			{
				Severity:    severity,
				Code:        code,
				Diagnostics: diagnostics,
			},
		},
	}
}

func ErrorOutcome(diagnostics string) *OperationOutcome {
	return NewOperationOutcome("error", "processing", diagnostics)
}

func NotFoundOutcome(resourceType, id string) *OperationOutcome {
	return NewOperationOutcome("error", "not-found", resourceType+"/"+id+" not found")
}
