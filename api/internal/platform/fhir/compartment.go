package fhir

// CompartmentDefinition maps resource types that belong to a compartment
// to the search parameter that links them.
type CompartmentDefinition struct {
	// Type is the compartment type (e.g., "Patient").
	Type string
	// Resources maps resource type -> search parameter names that link to this compartment.
	Resources map[string][]string
}

// PatientCompartment defines which resources belong to the Patient compartment
// and their linking search parameters per the FHIR R4 spec.
var PatientCompartment = CompartmentDefinition{
	Type: "Patient",
	Resources: map[string][]string{
		"AllergyIntolerance":       {"patient"},
		"Appointment":              {"patient"},
		"CarePlan":                 {"patient"},
		"CareTeam":                 {"patient"},
		"Claim":                    {"patient"},
		"Communication":           {"patient"},
		"Composition":             {"patient"},
		"Condition":               {"patient"},
		"Consent":                 {"patient"},
		"Coverage":                {"patient"},
		"DiagnosticReport":        {"patient"},
		"DocumentReference":       {"patient"},
		"Encounter":               {"patient"},
		"ImagingStudy":            {"patient"},
		"Medication":              {},
		"MedicationAdministration": {"patient"},
		"MedicationDispense":      {"patient"},
		"MedicationRequest":       {"patient"},
		"Observation":             {"patient"},
		"Procedure":               {"patient"},
		"QuestionnaireResponse":   {"patient"},
		"ResearchStudy":           {},
		"Schedule":                {},
		"ServiceRequest":          {"patient"},
		"Slot":                    {},
		"Specimen":                {"patient"},
	},
}

// GetCompartmentParam returns the search parameter that links a resource type
// to the given compartment. Returns empty string if the resource doesn't belong
// to the compartment or has no linking parameter.
func GetCompartmentParam(compartment *CompartmentDefinition, resourceType string) string {
	params, ok := compartment.Resources[resourceType]
	if !ok || len(params) == 0 {
		return ""
	}
	return params[0]
}

// IsInCompartment checks if a resource type is part of the given compartment.
func IsInCompartment(compartment *CompartmentDefinition, resourceType string) bool {
	_, ok := compartment.Resources[resourceType]
	return ok
}
