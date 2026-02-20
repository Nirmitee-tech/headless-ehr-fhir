package hipaa

// PHIFieldConfig maps a FHIR resource type to the field paths that contain
// Protected Health Information (PHI) per the HIPAA Safe Harbor de-identification
// standard (45 CFR 164.514(b)(2)). These are the 18 types of identifiers that
// must be removed or encrypted at rest.
type PHIFieldConfig struct {
	// ResourceType is the FHIR resource name (e.g. "Patient").
	ResourceType string
	// Fields lists the field paths within the resource that contain PHI.
	// Paths use dot notation matching the FHIR JSON element names.
	Fields []string
}

// DefaultPHIFields returns the PHI field configuration for standard FHIR
// resources that carry direct patient identifiers. The fields listed are the
// ones most likely to contain the HIPAA Safe Harbor 18 identifier categories:
//
//   - Names (covered by resource-level access control, not listed here)
//   - Geographic data smaller than state (address lines)
//   - Phone numbers, fax numbers
//   - Email addresses
//   - Social Security Numbers (SSN)
//   - Other account / device identifiers stored in FHIR "identifier" elements
//
// This list intentionally focuses on the fields that should be encrypted at
// rest in the database. Display-name fields (e.g. Patient.name) are protected
// by row-level access control and are NOT included here to avoid double
// encryption overhead on high-read paths.
func DefaultPHIFields() []PHIFieldConfig {
	return []PHIFieldConfig{
		{
			ResourceType: "Patient",
			Fields: []string{
				"identifier.ssn",     // SSN stored in identifier where system contains "SSN"
				"address.line",       // Street address lines
				"telecom.phone",      // Phone numbers
				"telecom.email",      // Email addresses
			},
		},
		{
			ResourceType: "Practitioner",
			Fields: []string{
				"address.line",   // Street address lines
				"telecom.phone",  // Phone numbers
				"telecom.email",  // Email addresses
			},
		},
		{
			ResourceType: "RelatedPerson",
			Fields: []string{
				"address.line",   // Street address lines
				"telecom.phone",  // Phone numbers
			},
		},
	}
}

// PHIFieldPaths returns a flat set of "<ResourceType>.<field>" strings for fast
// look-up. Example key: "Patient.telecom.phone".
func PHIFieldPaths() map[string]bool {
	configs := DefaultPHIFields()
	paths := make(map[string]bool, 16)
	for _, c := range configs {
		for _, f := range c.Fields {
			paths[c.ResourceType+"."+f] = true
		}
	}
	return paths
}
