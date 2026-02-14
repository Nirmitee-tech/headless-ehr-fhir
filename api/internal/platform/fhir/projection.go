package fhir

import (
	"encoding/json"
	"strings"
)

// MandatoryElements are always included regardless of _elements or _summary filters.
var MandatoryElements = map[string]bool{
	"resourceType": true,
	"id":           true,
	"meta":         true,
}

// SummaryElements defines which elements to include for _summary=true per resource type.
// If a resource type is not listed, a default set is used.
var SummaryElements = map[string][]string{
	"Patient": {"identifier", "active", "name", "gender", "birthDate", "address",
		"managingOrganization", "link"},
	"Observation": {"status", "category", "code", "subject", "encounter",
		"effectiveDateTime", "effectivePeriod", "issued", "valueQuantity",
		"valueCodeableConcept", "valueString", "dataAbsentReason", "interpretation"},
	"Condition": {"clinicalStatus", "verificationStatus", "category", "severity",
		"code", "subject", "encounter", "onsetDateTime", "abatementDateTime", "recordedDate"},
	"Encounter": {"identifier", "status", "class", "type", "subject", "participant",
		"period", "reasonCode", "serviceProvider"},
	"MedicationRequest": {"status", "intent", "medicationCodeableConcept",
		"medicationReference", "subject", "encounter", "authoredOn", "requester"},
	"AllergyIntolerance": {"clinicalStatus", "verificationStatus", "type", "category",
		"criticality", "code", "patient", "onsetDateTime", "recordedDate"},
	"Procedure": {"status", "code", "subject", "encounter", "performedDateTime",
		"performedPeriod"},
}

// DefaultSummaryElements is used when a resource type doesn't have specific summary definitions.
var DefaultSummaryElements = []string{
	"status", "code", "subject", "patient", "date", "category",
}

// ApplyElements filters a FHIR resource map to only include the specified elements
// plus mandatory elements (resourceType, id, meta).
func ApplyElements(resource map[string]interface{}, elements string) map[string]interface{} {
	if elements == "" {
		return resource
	}

	requestedFields := strings.Split(elements, ",")
	allowed := make(map[string]bool)
	for k := range MandatoryElements {
		allowed[k] = true
	}
	for _, f := range requestedFields {
		f = strings.TrimSpace(f)
		if f != "" {
			allowed[f] = true
		}
	}

	result := make(map[string]interface{})
	for k, v := range resource {
		if allowed[k] {
			result[k] = v
		}
	}
	return result
}

// ApplySummary applies _summary filtering to a FHIR resource.
// Modes: "true" (summary elements only), "text" (text + id/meta),
// "data" (remove text), "count" (no resources, just count), "false" (no filtering).
func ApplySummary(resource map[string]interface{}, summaryMode string) map[string]interface{} {
	if summaryMode == "" || summaryMode == "false" {
		return resource
	}

	resourceType, _ := resource["resourceType"].(string)

	switch summaryMode {
	case "true":
		summaryFields := SummaryElements[resourceType]
		if summaryFields == nil {
			summaryFields = DefaultSummaryElements
		}
		allowed := make(map[string]bool)
		for k := range MandatoryElements {
			allowed[k] = true
		}
		for _, f := range summaryFields {
			allowed[f] = true
		}
		result := make(map[string]interface{})
		for k, v := range resource {
			if allowed[k] {
				result[k] = v
			}
		}
		// Set the SUBSETTED tag
		addSubsettedTag(result)
		return result

	case "text":
		result := make(map[string]interface{})
		for k := range MandatoryElements {
			if v, ok := resource[k]; ok {
				result[k] = v
			}
		}
		if v, ok := resource["text"]; ok {
			result["text"] = v
		}
		addSubsettedTag(result)
		return result

	case "data":
		result := make(map[string]interface{})
		for k, v := range resource {
			if k != "text" {
				result[k] = v
			}
		}
		return result

	default:
		return resource
	}
}

// ApplyProjection applies both _elements and _summary to a resource.
// _elements takes precedence if both are specified.
func ApplyProjection(resource map[string]interface{}, elements, summary string) map[string]interface{} {
	if elements != "" {
		return ApplyElements(resource, elements)
	}
	if summary != "" {
		return ApplySummary(resource, summary)
	}
	return resource
}

// ApplyProjectionToBundle applies projection to all resources in a bundle.
func ApplyProjectionToBundle(bundle *Bundle, elements, summary string) {
	if elements == "" && summary == "" {
		return
	}
	if summary == "count" {
		bundle.Entry = nil
		return
	}

	for i, entry := range bundle.Entry {
		if len(entry.Resource) == 0 {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(entry.Resource, &m); err != nil {
			continue
		}
		projected := ApplyProjection(m, elements, summary)
		data, err := json.Marshal(projected)
		if err != nil {
			continue
		}
		bundle.Entry[i].Resource = data
	}
}

// addSubsettedTag adds the SUBSETTED meta tag to indicate partial content.
func addSubsettedTag(resource map[string]interface{}) {
	meta, ok := resource["meta"].(map[string]interface{})
	if !ok {
		meta = make(map[string]interface{})
		resource["meta"] = meta
	}

	tags, _ := meta["tag"].([]interface{})
	tags = append(tags, map[string]interface{}{
		"system": "http://terminology.hl7.org/CodeSystem/v3-ObservationValue",
		"code":   "SUBSETTED",
	})
	meta["tag"] = tags
}
