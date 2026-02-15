package ccda

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// ParsedDocument represents the extracted data from a C-CDA document.
type ParsedDocument struct {
	Title    string
	Created  time.Time
	Patient  ParsedPatient
	Sections []ParsedSection
}

// ParsedPatient contains the patient demographics extracted from the CDA header.
type ParsedPatient struct {
	Name        string
	DOB         string
	Gender      string
	Identifiers []ParsedID
}

// ParsedID is a parsed identifier.
type ParsedID struct {
	Root      string
	Extension string
}

// ParsedSection holds data extracted from a single CDA section.
type ParsedSection struct {
	Type    string                   // "allergies", "medications", "problems", etc.
	Title   string
	Entries []map[string]interface{} // Extracted data as FHIR-like maps
}

// Parser extracts structured data from C-CDA documents. It is safe for
// concurrent use because it holds no mutable state.
type Parser struct{}

// NewParser creates a new C-CDA parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a C-CDA XML document and extracts structured data.
func (p *Parser) Parse(xmlData []byte) (*ParsedDocument, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("ccda: XML data is empty")
	}

	var doc ClinicalDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return nil, fmt.Errorf("ccda: failed to parse XML: %w", err)
	}

	result := &ParsedDocument{
		Title: doc.Title,
	}

	// Parse effective time
	if doc.EffectiveTime != nil && doc.EffectiveTime.Value != "" {
		if t, err := parseHL7Time(doc.EffectiveTime.Value); err == nil {
			result.Created = t
		}
	}

	// Parse patient
	result.Patient = p.parsePatient(&doc)

	// Parse sections
	if doc.Component != nil && doc.Component.StructuredBody != nil {
		for _, comp := range doc.Component.StructuredBody.Components {
			if comp.Section != nil {
				if ps := p.parseSection(comp.Section); ps != nil {
					result.Sections = append(result.Sections, *ps)
				}
			}
		}
	}

	return result, nil
}

// parsePatient extracts patient demographics from the CDA header.
func (p *Parser) parsePatient(doc *ClinicalDocument) ParsedPatient {
	patient := ParsedPatient{}

	if doc.RecordTarget == nil || doc.RecordTarget.PatientRole == nil {
		return patient
	}

	role := doc.RecordTarget.PatientRole

	// IDs
	for _, id := range role.IDs {
		patient.Identifiers = append(patient.Identifiers, ParsedID{
			Root:      id.Root,
			Extension: id.Extension,
		})
	}

	if role.Patient == nil {
		return patient
	}

	pat := role.Patient

	// Name
	if pat.Name != nil {
		parts := []string{}
		if pat.Name.Given != "" {
			parts = append(parts, pat.Name.Given)
		}
		if pat.Name.Family != "" {
			parts = append(parts, pat.Name.Family)
		}
		patient.Name = strings.Join(parts, " ")
	}

	// Gender
	if pat.AdministrativeGenderCode != nil {
		patient.Gender = pat.AdministrativeGenderCode.DisplayName
		if patient.Gender == "" {
			patient.Gender = pat.AdministrativeGenderCode.Code
		}
	}

	// DOB
	if pat.BirthTime != nil && pat.BirthTime.Value != "" {
		patient.DOB = formatParsedDate(pat.BirthTime.Value)
	}

	return patient
}

// parseSection maps a CDA section to a ParsedSection based on its LOINC code.
func (p *Parser) parseSection(section *Section) *ParsedSection {
	if section.Code == nil {
		return nil
	}

	sectionType := mapLOINCToType(section.Code.Code)
	if sectionType == "" {
		// Unknown section type; still return it with a generic type
		sectionType = "unknown"
	}

	ps := &ParsedSection{
		Type:  sectionType,
		Title: section.Title,
	}

	switch sectionType {
	case "allergies":
		ps.Entries = p.parseAllergyEntries(section)
	case "medications":
		ps.Entries = p.parseMedicationEntries(section)
	case "problems":
		ps.Entries = p.parseProblemEntries(section)
	case "procedures":
		ps.Entries = p.parseProcedureEntries(section)
	case "results":
		ps.Entries = p.parseResultEntries(section)
	case "vital_signs":
		ps.Entries = p.parseVitalSignEntries(section)
	case "immunizations":
		ps.Entries = p.parseImmunizationEntries(section)
	case "social_history":
		ps.Entries = p.parseSocialHistoryEntries(section)
	case "plan_of_care":
		ps.Entries = p.parsePlanOfCareEntries(section)
	case "encounters":
		ps.Entries = p.parseEncounterEntries(section)
	}

	return ps
}

// ---- Entry Parsers ----

func (p *Parser) parseAllergyEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Act == nil {
			continue
		}
		entry := map[string]interface{}{
			"resourceType": "AllergyIntolerance",
		}
		if e.Act.StatusCode != nil {
			entry["clinicalStatus"] = e.Act.StatusCode.Code
		}
		for _, er := range e.Act.EntryRelationships {
			if er.Observation != nil && er.Observation.Value != nil {
				entry["substance"] = er.Observation.Value.DisplayName
				if er.Observation.Value.Code != "" {
					entry["code"] = er.Observation.Value.Code
				}
				if er.Observation.Value.CodeSystem != "" {
					entry["codeSystem"] = er.Observation.Value.CodeSystem
				}
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parseMedicationEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.SubstanceAdministration == nil {
			continue
		}
		sa := e.SubstanceAdministration
		entry := map[string]interface{}{
			"resourceType": "MedicationRequest",
		}
		if sa.StatusCode != nil {
			entry["status"] = sa.StatusCode.Code
		}
		if sa.Consumable != nil && sa.Consumable.ManufacturedProduct != nil &&
			sa.Consumable.ManufacturedProduct.ManufacturedMaterial != nil &&
			sa.Consumable.ManufacturedProduct.ManufacturedMaterial.Code != nil {
			medCode := sa.Consumable.ManufacturedProduct.ManufacturedMaterial.Code
			entry["medication"] = medCode.DisplayName
			if medCode.Code != "" {
				entry["code"] = medCode.Code
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parseProblemEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Act == nil {
			continue
		}
		entry := map[string]interface{}{
			"resourceType": "Condition",
		}
		if e.Act.StatusCode != nil {
			entry["clinicalStatus"] = e.Act.StatusCode.Code
		}
		if e.Act.EffectiveTime != nil && e.Act.EffectiveTime.Low != nil {
			entry["onsetDate"] = formatParsedDate(e.Act.EffectiveTime.Low.Value)
		}
		for _, er := range e.Act.EntryRelationships {
			if er.Observation != nil && er.Observation.Value != nil {
				entry["problem"] = er.Observation.Value.DisplayName
				if er.Observation.Value.Code != "" {
					entry["code"] = er.Observation.Value.Code
				}
				if er.Observation.Value.CodeSystem != "" {
					entry["codeSystem"] = er.Observation.Value.CodeSystem
				}
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parseProcedureEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Procedure == nil {
			continue
		}
		entry := map[string]interface{}{
			"resourceType": "Procedure",
		}
		if e.Procedure.Code != nil {
			entry["procedure"] = e.Procedure.Code.DisplayName
			if e.Procedure.Code.Code != "" {
				entry["code"] = e.Procedure.Code.Code
			}
		}
		if e.Procedure.StatusCode != nil {
			entry["status"] = e.Procedure.StatusCode.Code
		}
		if e.Procedure.EffectiveTime != nil && e.Procedure.EffectiveTime.Low != nil {
			entry["date"] = formatParsedDate(e.Procedure.EffectiveTime.Low.Value)
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parseResultEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Organizer == nil {
			continue
		}
		for _, comp := range e.Organizer.Components {
			if comp.Observation == nil {
				continue
			}
			obs := comp.Observation
			entry := map[string]interface{}{
				"resourceType": "Observation",
			}
			if obs.Code != nil {
				entry["test"] = obs.Code.DisplayName
				if obs.Code.Code != "" {
					entry["code"] = obs.Code.Code
				}
			}
			if obs.Value != nil {
				entry["value"] = obs.Value.Value
				if obs.Value.Unit != "" {
					entry["unit"] = obs.Value.Unit
				}
			}
			if obs.EffectiveTime != nil && obs.EffectiveTime.Low != nil {
				entry["date"] = formatParsedDate(obs.EffectiveTime.Low.Value)
			}
			entries = append(entries, entry)
		}
	}
	return entries
}

func (p *Parser) parseVitalSignEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Organizer == nil {
			continue
		}
		for _, comp := range e.Organizer.Components {
			if comp.Observation == nil {
				continue
			}
			obs := comp.Observation
			entry := map[string]interface{}{
				"resourceType": "Observation",
			}
			if obs.Code != nil {
				entry["vitalSign"] = obs.Code.DisplayName
				if obs.Code.Code != "" {
					entry["code"] = obs.Code.Code
				}
			}
			if obs.Value != nil {
				entry["value"] = obs.Value.Value
				if obs.Value.Unit != "" {
					entry["unit"] = obs.Value.Unit
				}
			}
			if obs.EffectiveTime != nil && obs.EffectiveTime.Low != nil {
				entry["date"] = formatParsedDate(obs.EffectiveTime.Low.Value)
			}
			entries = append(entries, entry)
		}
	}
	return entries
}

func (p *Parser) parseImmunizationEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.SubstanceAdministration == nil {
			continue
		}
		sa := e.SubstanceAdministration
		entry := map[string]interface{}{
			"resourceType": "Immunization",
		}
		if sa.StatusCode != nil {
			entry["status"] = sa.StatusCode.Code
		}
		if sa.EffectiveTime != nil && sa.EffectiveTime.Low != nil {
			entry["date"] = formatParsedDate(sa.EffectiveTime.Low.Value)
		}
		if sa.Consumable != nil && sa.Consumable.ManufacturedProduct != nil &&
			sa.Consumable.ManufacturedProduct.ManufacturedMaterial != nil &&
			sa.Consumable.ManufacturedProduct.ManufacturedMaterial.Code != nil {
			code := sa.Consumable.ManufacturedProduct.ManufacturedMaterial.Code
			entry["vaccine"] = code.DisplayName
			if code.Code != "" {
				entry["code"] = code.Code
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parseSocialHistoryEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Observation == nil {
			continue
		}
		obs := e.Observation
		entry := map[string]interface{}{
			"resourceType": "Observation",
		}
		if obs.Code != nil {
			entry["observation"] = obs.Code.DisplayName
			if obs.Code.Code != "" {
				entry["code"] = obs.Code.Code
			}
		}
		if obs.Value != nil {
			entry["value"] = obs.Value.DisplayName
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parsePlanOfCareEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Act == nil {
			continue
		}
		entry := map[string]interface{}{
			"resourceType": "CarePlan",
		}
		if e.Act.StatusCode != nil {
			entry["status"] = e.Act.StatusCode.Code
		}
		if e.Act.EffectiveTime != nil && e.Act.EffectiveTime.Low != nil {
			entry["date"] = formatParsedDate(e.Act.EffectiveTime.Low.Value)
		}
		entries = append(entries, entry)
	}
	return entries
}

func (p *Parser) parseEncounterEntries(section *Section) []map[string]interface{} {
	var entries []map[string]interface{}
	for _, e := range section.Entries {
		if e.Encounter == nil {
			continue
		}
		enc := e.Encounter
		entry := map[string]interface{}{
			"resourceType": "Encounter",
		}
		if enc.Code != nil {
			entry["type"] = enc.Code.DisplayName
			if enc.Code.Code != "" {
				entry["code"] = enc.Code.Code
			}
		}
		if enc.StatusCode != nil {
			entry["status"] = enc.StatusCode.Code
		}
		if enc.EffectiveTime != nil && enc.EffectiveTime.Low != nil {
			entry["date"] = formatParsedDate(enc.EffectiveTime.Low.Value)
		}
		entries = append(entries, entry)
	}
	return entries
}

// ---- Helpers ----

// mapLOINCToType maps a LOINC section code to a section type name.
func mapLOINCToType(code string) string {
	switch code {
	case LOINCAllergies:
		return "allergies"
	case LOINCMedications:
		return "medications"
	case LOINCProblems:
		return "problems"
	case LOINCProcedures:
		return "procedures"
	case LOINCResults:
		return "results"
	case LOINCVitalSigns:
		return "vital_signs"
	case LOINCImmunizations:
		return "immunizations"
	case LOINCSocialHistory:
		return "social_history"
	case LOINCPlanOfCare:
		return "plan_of_care"
	case LOINCEncounters:
		return "encounters"
	default:
		return ""
	}
}

// parseHL7Time parses an HL7 time string into a time.Time.
func parseHL7Time(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	switch len(s) {
	case 14: // YYYYMMDDHHmmss
		return time.Parse("20060102150405", s)
	case 12: // YYYYMMDDHHmm
		return time.Parse("200601021504", s)
	case 8: // YYYYMMDD
		return time.Parse("20060102", s)
	default:
		if len(s) > 14 {
			return time.Parse("20060102150405", s[:14])
		}
		return time.Time{}, fmt.Errorf("ccda: unrecognized time format: %s", s)
	}
}

// formatParsedDate converts an HL7 date (YYYYMMDD) to a more readable format.
func formatParsedDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}
