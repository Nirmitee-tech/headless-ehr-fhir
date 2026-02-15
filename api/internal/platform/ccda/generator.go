package ccda

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PatientData holds all clinical data needed to generate a CCD.
type PatientData struct {
	Patient       map[string]interface{} // FHIR Patient resource
	Allergies     []map[string]interface{} // FHIR AllergyIntolerance resources
	Medications   []map[string]interface{} // FHIR MedicationRequest resources
	Conditions    []map[string]interface{} // FHIR Condition resources
	Procedures    []map[string]interface{} // FHIR Procedure resources
	Results       []map[string]interface{} // FHIR Observation (lab) resources
	VitalSigns    []map[string]interface{} // FHIR Observation (vital) resources
	Immunizations []map[string]interface{} // FHIR Immunization resources
	Encounters    []map[string]interface{} // FHIR Encounter resources
	SocialHistory []map[string]interface{} // FHIR Observation (social) resources
	CarePlans     []map[string]interface{} // FHIR CarePlan resources
}

// Generator creates C-CDA 2.1 CCD documents from FHIR data. It is safe
// for concurrent use because it holds only immutable configuration.
type Generator struct {
	orgName string // Custodian organization name
	orgOID  string // Custodian OID
}

// NewGenerator creates a new C-CDA generator.
func NewGenerator(orgName, orgOID string) *Generator {
	return &Generator{
		orgName: orgName,
		orgOID:  orgOID,
	}
}

// GenerateCCD produces a complete CCD XML document from the provided patient data.
func (g *Generator) GenerateCCD(data *PatientData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("ccda: patient data is nil")
	}
	if data.Patient == nil {
		return nil, fmt.Errorf("ccda: patient resource is required")
	}

	doc := g.buildDocument(data)

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ccda: failed to marshal XML: %w", err)
	}

	// Prepend XML declaration
	header := []byte(xml.Header)
	result := make([]byte, len(header)+len(output))
	copy(result, header)
	copy(result[len(header):], output)
	return result, nil
}

// buildDocument constructs the full ClinicalDocument from patient data.
func (g *Generator) buildDocument(data *PatientData) *ClinicalDocument {
	now := time.Now().UTC()
	docID := uuid.New().String()

	doc := &ClinicalDocument{
		XSI:  XSINamespace,
		SDTC: SDTCNamespace,
		RealmCode: &Code{Code: "US"},
		TypeID: &TypeID{
			Root:      "2.16.840.1.113883.1.3",
			Extension: "POCD_HD000040",
		},
		TemplateIDs: []TemplateID{
			{Root: OIDUSRealmHeader},
			{Root: OIDCCDDocument},
		},
		ID: &InstanceID{Root: docID},
		Code: &Code{
			Code:           "34133-9",
			CodeSystem:     OIDLOINC,
			CodeSystemName: "LOINC",
			DisplayName:    "Summarization of Episode Note",
		},
		Title:         "Continuity of Care Document",
		EffectiveTime: &TimeValue{Value: formatHL7Time(now)},
		ConfidentialityCode: &Code{
			Code:       "N",
			CodeSystem: "2.16.840.1.113883.5.25",
		},
		LanguageCode: &Code{Code: "en-US"},
	}

	// Header components
	doc.RecordTarget = g.buildRecordTarget(data.Patient)
	doc.Author = g.buildAuthor(now)
	doc.Custodian = g.buildCustodian()
	doc.DocumentationOf = g.buildDocumentationOf(now)

	// Body sections
	sections := g.buildSections(data)
	if len(sections) > 0 {
		components := make([]SectionComponent, len(sections))
		for i, s := range sections {
			sec := s
			components[i] = SectionComponent{Section: &sec}
		}
		doc.Component = &Component{
			StructuredBody: &StructuredBody{
				Components: components,
			},
		}
	}

	return doc
}

// buildRecordTarget constructs the patient header from a FHIR Patient resource.
func (g *Generator) buildRecordTarget(patient map[string]interface{}) *RecordTarget {
	role := &PatientRole{}

	// Patient ID
	if id, ok := getString(patient, "id"); ok {
		role.IDs = []InstanceID{{Root: g.orgOID, Extension: id}}
	}

	// Address
	if addrs, ok := getArray(patient, "address"); ok && len(addrs) > 0 {
		if addr, ok := addrs[0].(map[string]interface{}); ok {
			role.Addr = buildAddress(addr)
		}
	}

	// Telecom
	if telecoms, ok := getArray(patient, "telecom"); ok && len(telecoms) > 0 {
		if t, ok := telecoms[0].(map[string]interface{}); ok {
			val, _ := getString(t, "value")
			use, _ := getString(t, "use")
			role.Telecom = &Telecom{Value: val, Use: use}
		}
	}

	// Demographics
	pat := &Patient{}

	// Name
	if names, ok := getArray(patient, "name"); ok && len(names) > 0 {
		if n, ok := names[0].(map[string]interface{}); ok {
			pat.Name = buildName(n)
		}
	}

	// Gender
	if gender, ok := getString(patient, "gender"); ok {
		genderCode := mapGenderCode(gender)
		pat.AdministrativeGenderCode = &Code{
			Code:           genderCode,
			CodeSystem:     OIDAdminGender,
			DisplayName:    gender,
		}
	}

	// Birth date
	if dob, ok := getString(patient, "birthDate"); ok {
		pat.BirthTime = &TimeValue{Value: strings.ReplaceAll(dob, "-", "")}
	}

	role.Patient = pat
	return &RecordTarget{PatientRole: role}
}

// buildAuthor creates the document author section.
func (g *Generator) buildAuthor(now time.Time) *Author {
	return &Author{
		Time: &TimeValue{Value: formatHL7Time(now)},
		AssignedAuthor: &AssignedAuthor{
			ID: &InstanceID{Root: g.orgOID},
			AssignedAuthoringDevice: &AuthoringDevice{
				SoftwareName: "EHR System",
			},
			RepresentedOrganization: &Organization{
				IDs:   []InstanceID{{Root: g.orgOID}},
				Names: []string{g.orgName},
			},
		},
	}
}

// buildCustodian creates the custodian section.
func (g *Generator) buildCustodian() *Custodian {
	return &Custodian{
		AssignedCustodian: &AssignedCustodian{
			RepresentedCustodianOrganization: &CustodianOrganization{
				IDs:   []InstanceID{{Root: g.orgOID}},
				Names: []string{g.orgName},
			},
		},
	}
}

// buildDocumentationOf creates the documentation-of section.
func (g *Generator) buildDocumentationOf(now time.Time) *DocumentationOf {
	return &DocumentationOf{
		ServiceEvent: &ServiceEvent{
			ClassCode: "PCPR",
			EffectiveTime: &TimeRange{
				Low:  &TimeLow{Value: formatHL7Time(now)},
				High: &TimeHigh{Value: formatHL7Time(now)},
			},
		},
	}
}

// buildSections creates all applicable CDA sections from the patient data.
// Sections with no data are omitted.
func (g *Generator) buildSections(data *PatientData) []Section {
	var sections []Section

	if len(data.Allergies) > 0 {
		sections = append(sections, buildAllergiesSection(data.Allergies))
	}
	if len(data.Medications) > 0 {
		sections = append(sections, buildMedicationsSection(data.Medications))
	}
	if len(data.Conditions) > 0 {
		sections = append(sections, buildProblemsSection(data.Conditions))
	}
	if len(data.Procedures) > 0 {
		sections = append(sections, buildProceduresSection(data.Procedures))
	}
	if len(data.Results) > 0 {
		sections = append(sections, buildResultsSection(data.Results))
	}
	if len(data.VitalSigns) > 0 {
		sections = append(sections, buildVitalSignsSection(data.VitalSigns))
	}
	if len(data.Immunizations) > 0 {
		sections = append(sections, buildImmunizationsSection(data.Immunizations))
	}
	if len(data.SocialHistory) > 0 {
		sections = append(sections, buildSocialHistorySection(data.SocialHistory))
	}
	if len(data.CarePlans) > 0 {
		sections = append(sections, buildPlanOfCareSection(data.CarePlans))
	}
	if len(data.Encounters) > 0 {
		sections = append(sections, buildEncountersSection(data.Encounters))
	}

	return sections
}

// ---- Section Builders ----

func buildAllergiesSection(allergies []map[string]interface{}) Section {
	section := newSection(OIDAllergiesSection, LOINCAllergies, "LOINC", "Allergies and Adverse Reactions")

	headers := []string{"Substance", "Reaction", "Status"}
	var rows []NarrativeTr
	var entries []Entry

	for _, a := range allergies {
		substance := "Unknown"
		reaction := "Unknown"
		status := "active"

		if code, ok := getNestedMap(a, "code"); ok {
			if disp, ok := getString(code, "text"); ok {
				substance = disp
			} else if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						substance = d
					}
				}
			}
		}

		if reactions, ok := getArray(a, "reaction"); ok && len(reactions) > 0 {
			if r, ok := reactions[0].(map[string]interface{}); ok {
				if manifestations, ok := getArray(r, "manifestation"); ok && len(manifestations) > 0 {
					if m, ok := manifestations[0].(map[string]interface{}); ok {
						if codings, ok := getArray(m, "coding"); ok && len(codings) > 0 {
							if c, ok := codings[0].(map[string]interface{}); ok {
								if d, ok := getString(c, "display"); ok {
									reaction = d
								}
							}
						}
					}
				}
			}
		}

		if cs, ok := getString(a, "clinicalStatus"); ok {
			status = cs
		} else if csObj, ok := getNestedMap(a, "clinicalStatus"); ok {
			if codings, ok := getArray(csObj, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if code, ok := getString(c, "code"); ok {
						status = code
					}
				}
			}
		}

		rows = append(rows, NarrativeTr{Tds: []string{substance, reaction, status}})

		allergyCode, allergySystem := extractCode(a, "code")
		entry := Entry{
			TypeCode: "DRIV",
			Act: &Act{
				ClassCode: "ACT",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDAllergyEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				Code: &Code{
					Code:       "CONC",
					CodeSystem: "2.16.840.1.113883.5.6",
				},
				StatusCode: &Code{Code: status},
				EntryRelationships: []EntryRelationship{
					{
						TypeCode: "SUBJ",
						Observation: &ObservationEntry{
							ClassCode: "OBS",
							MoodCode:  "EVN",
							Code: &Code{
								Code:       allergyCode,
								CodeSystem: allergySystem,
								DisplayName: substance,
							},
							Value: &Value{
								Type:        "CD",
								Code:        allergyCode,
								CodeSystem:  allergySystem,
								DisplayName: substance,
							},
						},
					},
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildMedicationsSection(meds []map[string]interface{}) Section {
	section := newSection(OIDMedicationsSection, LOINCMedications, "LOINC", "Medications")

	headers := []string{"Medication", "Dosage", "Status"}
	var rows []NarrativeTr
	var entries []Entry

	for _, m := range meds {
		medName := "Unknown"
		dosage := ""
		status := "active"

		if mc, ok := getNestedMap(m, "medicationCodeableConcept"); ok {
			if disp, ok := getString(mc, "text"); ok {
				medName = disp
			} else if codings, ok := getArray(mc, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						medName = d
					}
				}
			}
		}

		if doses, ok := getArray(m, "dosageInstruction"); ok && len(doses) > 0 {
			if d, ok := doses[0].(map[string]interface{}); ok {
				if txt, ok := getString(d, "text"); ok {
					dosage = txt
				}
			}
		}

		if s, ok := getString(m, "status"); ok {
			status = s
		}

		rows = append(rows, NarrativeTr{Tds: []string{medName, dosage, status}})

		medCode, medSystem := extractCodeFromCodeableConcept(m, "medicationCodeableConcept")
		entry := Entry{
			TypeCode: "DRIV",
			SubstanceAdministration: &SubstanceAdministration{
				ClassCode: "SBADM",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDMedicationEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				StatusCode: &Code{Code: status},
				Consumable: &Consumable{
					ManufacturedProduct: &ManufacturedProduct{
						ManufacturedMaterial: &ManufacturedMaterial{
							Code: &Code{
								Code:           medCode,
								CodeSystem:     medSystem,
								DisplayName:    medName,
							},
						},
					},
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildProblemsSection(conditions []map[string]interface{}) Section {
	section := newSection(OIDProblemsSection, LOINCProblems, "LOINC", "Problems")

	headers := []string{"Problem", "Status", "Date"}
	var rows []NarrativeTr
	var entries []Entry

	for _, cond := range conditions {
		problem := "Unknown"
		status := "active"
		dateStr := ""

		if code, ok := getNestedMap(cond, "code"); ok {
			if disp, ok := getString(code, "text"); ok {
				problem = disp
			} else if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						problem = d
					}
				}
			}
		}

		if cs, ok := getString(cond, "clinicalStatus"); ok {
			status = cs
		} else if csObj, ok := getNestedMap(cond, "clinicalStatus"); ok {
			if codings, ok := getArray(csObj, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if code, ok := getString(c, "code"); ok {
						status = code
					}
				}
			}
		}

		if onset, ok := getString(cond, "onsetDateTime"); ok {
			dateStr = onset
		}

		rows = append(rows, NarrativeTr{Tds: []string{problem, status, dateStr}})

		condCode, condSystem := extractCode(cond, "code")
		entry := Entry{
			TypeCode: "DRIV",
			Act: &Act{
				ClassCode: "ACT",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDProblemEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				Code: &Code{
					Code:       "CONC",
					CodeSystem: "2.16.840.1.113883.5.6",
				},
				StatusCode: &Code{Code: status},
				EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
				EntryRelationships: []EntryRelationship{
					{
						TypeCode: "SUBJ",
						Observation: &ObservationEntry{
							ClassCode: "OBS",
							MoodCode:  "EVN",
							Code: &Code{
								Code:           condCode,
								CodeSystem:     condSystem,
								DisplayName:    problem,
							},
							Value: &Value{
								Type:        "CD",
								Code:        condCode,
								CodeSystem:  condSystem,
								DisplayName: problem,
							},
						},
					},
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildProceduresSection(procedures []map[string]interface{}) Section {
	section := newSection(OIDProceduresSection, LOINCProcedures, "LOINC", "Procedures")

	headers := []string{"Procedure", "Date", "Status"}
	var rows []NarrativeTr
	var entries []Entry

	for _, proc := range procedures {
		procName := "Unknown"
		dateStr := ""
		status := "completed"

		if code, ok := getNestedMap(proc, "code"); ok {
			if disp, ok := getString(code, "text"); ok {
				procName = disp
			} else if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						procName = d
					}
				}
			}
		}

		if dt, ok := getString(proc, "performedDateTime"); ok {
			dateStr = dt
		} else if period, ok := getNestedMap(proc, "performedPeriod"); ok {
			if start, ok := getString(period, "start"); ok {
				dateStr = start
			}
		}

		if s, ok := getString(proc, "status"); ok {
			status = s
		}

		rows = append(rows, NarrativeTr{Tds: []string{procName, dateStr, status}})

		procCode, procSystem := extractCode(proc, "code")
		entry := Entry{
			TypeCode: "DRIV",
			Procedure: &ProcedureEntry{
				ClassCode: "PROC",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDProcedureEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				Code: &Code{
					Code:        procCode,
					CodeSystem:  procSystem,
					DisplayName: procName,
				},
				StatusCode: &Code{Code: status},
				EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildResultsSection(results []map[string]interface{}) Section {
	section := newSection(OIDResultsSection, LOINCResults, "LOINC", "Results")

	headers := []string{"Test", "Value", "Date"}
	var rows []NarrativeTr
	var entries []Entry

	for _, obs := range results {
		testName := "Unknown"
		valueStr := ""
		dateStr := ""

		if code, ok := getNestedMap(obs, "code"); ok {
			if disp, ok := getString(code, "text"); ok {
				testName = disp
			} else if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						testName = d
					}
				}
			}
		}

		valueStr = extractFHIRValue(obs)

		if dt, ok := getString(obs, "effectiveDateTime"); ok {
			dateStr = dt
		}

		rows = append(rows, NarrativeTr{Tds: []string{testName, valueStr, dateStr}})

		obsCode, obsSystem := extractCode(obs, "code")
		obsVal, obsUnit := extractQuantityValue(obs)

		entry := Entry{
			TypeCode: "DRIV",
			Organizer: &Organizer{
				ClassCode: "CLUSTER",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDResultEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				Code: &Code{
					Code:           obsCode,
					CodeSystem:     obsSystem,
					DisplayName:    testName,
				},
				StatusCode: &Code{Code: "completed"},
				Components: []OrganizerComponent{
					{
						Observation: &ObservationEntry{
							ClassCode: "OBS",
							MoodCode:  "EVN",
							Code: &Code{
								Code:           obsCode,
								CodeSystem:     obsSystem,
								DisplayName:    testName,
							},
							StatusCode: &Code{Code: "completed"},
							EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
							Value: &Value{
								Type:  "PQ",
								Value: obsVal,
								Unit:  obsUnit,
							},
						},
					},
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildVitalSignsSection(vitals []map[string]interface{}) Section {
	section := newSection(OIDVitalSignsSection, LOINCVitalSigns, "LOINC", "Vital Signs")

	headers := []string{"Vital Sign", "Value", "Date"}
	var rows []NarrativeTr
	var entries []Entry

	for _, obs := range vitals {
		vitalName := "Unknown"
		valueStr := ""
		dateStr := ""

		if code, ok := getNestedMap(obs, "code"); ok {
			if disp, ok := getString(code, "text"); ok {
				vitalName = disp
			} else if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						vitalName = d
					}
				}
			}
		}

		valueStr = extractFHIRValue(obs)

		if dt, ok := getString(obs, "effectiveDateTime"); ok {
			dateStr = dt
		}

		rows = append(rows, NarrativeTr{Tds: []string{vitalName, valueStr, dateStr}})

		obsCode, obsSystem := extractCode(obs, "code")
		obsVal, obsUnit := extractQuantityValue(obs)

		entry := Entry{
			TypeCode: "DRIV",
			Organizer: &Organizer{
				ClassCode: "CLUSTER",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDVitalSignEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				StatusCode: &Code{Code: "completed"},
				Components: []OrganizerComponent{
					{
						Observation: &ObservationEntry{
							ClassCode: "OBS",
							MoodCode:  "EVN",
							Code: &Code{
								Code:           obsCode,
								CodeSystem:     obsSystem,
								DisplayName:    vitalName,
							},
							StatusCode: &Code{Code: "completed"},
							EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
							Value: &Value{
								Type:  "PQ",
								Value: obsVal,
								Unit:  obsUnit,
							},
						},
					},
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildImmunizationsSection(immunizations []map[string]interface{}) Section {
	section := newSection(OIDImmunizationsSection, LOINCImmunizations, "LOINC", "Immunizations")

	headers := []string{"Vaccine", "Date", "Status"}
	var rows []NarrativeTr
	var entries []Entry

	for _, imm := range immunizations {
		vaccine := "Unknown"
		dateStr := ""
		status := "completed"

		if vc, ok := getNestedMap(imm, "vaccineCode"); ok {
			if disp, ok := getString(vc, "text"); ok {
				vaccine = disp
			} else if codings, ok := getArray(vc, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						vaccine = d
					}
				}
			}
		}

		if dt, ok := getString(imm, "occurrenceDateTime"); ok {
			dateStr = dt
		}

		if s, ok := getString(imm, "status"); ok {
			status = s
		}

		rows = append(rows, NarrativeTr{Tds: []string{vaccine, dateStr, status}})

		vaccCode, vaccSystem := extractCodeFromCodeableConcept(imm, "vaccineCode")
		entry := Entry{
			TypeCode: "DRIV",
			SubstanceAdministration: &SubstanceAdministration{
				ClassCode: "SBADM",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDImmunizationEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				StatusCode: &Code{Code: status},
				EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
				Consumable: &Consumable{
					ManufacturedProduct: &ManufacturedProduct{
						ManufacturedMaterial: &ManufacturedMaterial{
							Code: &Code{
								Code:        vaccCode,
								CodeSystem:  vaccSystem,
								DisplayName: vaccine,
							},
						},
					},
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildSocialHistorySection(observations []map[string]interface{}) Section {
	section := newSection(OIDSocialHistorySection, LOINCSocialHistory, "LOINC", "Social History")

	headers := []string{"Observation", "Value", "Date"}
	var rows []NarrativeTr
	var entries []Entry

	for _, obs := range observations {
		obsName := "Unknown"
		valueStr := ""
		dateStr := ""

		if code, ok := getNestedMap(obs, "code"); ok {
			if disp, ok := getString(code, "text"); ok {
				obsName = disp
			} else if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if d, ok := getString(c, "display"); ok {
						obsName = d
					}
				}
			}
		}

		valueStr = extractFHIRValue(obs)

		if dt, ok := getString(obs, "effectiveDateTime"); ok {
			dateStr = dt
		}

		rows = append(rows, NarrativeTr{Tds: []string{obsName, valueStr, dateStr}})

		obsCode, obsSystem := extractCode(obs, "code")
		entry := Entry{
			TypeCode: "DRIV",
			Observation: &ObservationEntry{
				ClassCode: "OBS",
				MoodCode:  "EVN",
				Code: &Code{
					Code:        obsCode,
					CodeSystem:  obsSystem,
					DisplayName: obsName,
				},
				StatusCode: &Code{Code: "completed"},
				EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
				Value: &Value{
					Type:        "ST",
					DisplayName: valueStr,
				},
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildPlanOfCareSection(carePlans []map[string]interface{}) Section {
	section := newSection(OIDPlanOfCareSection, LOINCPlanOfCare, "LOINC", "Plan of Care")

	headers := []string{"Plan", "Status", "Date"}
	var rows []NarrativeTr
	var entries []Entry

	for _, cp := range carePlans {
		planTitle := "Unknown"
		status := "active"
		dateStr := ""

		if t, ok := getString(cp, "title"); ok {
			planTitle = t
		} else if desc, ok := getString(cp, "description"); ok {
			planTitle = desc
		}

		if s, ok := getString(cp, "status"); ok {
			status = s
		}

		if period, ok := getNestedMap(cp, "period"); ok {
			if start, ok := getString(period, "start"); ok {
				dateStr = start
			}
		}

		rows = append(rows, NarrativeTr{Tds: []string{planTitle, status, dateStr}})

		entry := Entry{
			TypeCode: "DRIV",
			Act: &Act{
				ClassCode: "ACT",
				MoodCode:  "INT",
				TemplateIDs: []TemplateID{{Root: OIDPlanOfCareSection}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				Code: &Code{
					NullFlavor: "NA",
				},
				StatusCode: &Code{Code: status},
				EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

func buildEncountersSection(encounters []map[string]interface{}) Section {
	section := newSection(OIDEncountersSection, LOINCEncounters, "LOINC", "Encounters")

	headers := []string{"Encounter", "Date", "Status"}
	var rows []NarrativeTr
	var entries []Entry

	for _, enc := range encounters {
		encType := "Unknown"
		dateStr := ""
		status := "finished"

		if types, ok := getArray(enc, "type"); ok && len(types) > 0 {
			if t, ok := types[0].(map[string]interface{}); ok {
				if disp, ok := getString(t, "text"); ok {
					encType = disp
				} else if codings, ok := getArray(t, "coding"); ok && len(codings) > 0 {
					if c, ok := codings[0].(map[string]interface{}); ok {
						if d, ok := getString(c, "display"); ok {
							encType = d
						}
					}
				}
			}
		}

		if period, ok := getNestedMap(enc, "period"); ok {
			if start, ok := getString(period, "start"); ok {
				dateStr = start
			}
		}

		if s, ok := getString(enc, "status"); ok {
			status = s
		}

		rows = append(rows, NarrativeTr{Tds: []string{encType, dateStr, status}})

		encCode, encSystem := extractCodeFromArray(enc, "type")
		entry := Entry{
			TypeCode: "DRIV",
			Encounter: &EncounterEntry{
				ClassCode: "ENC",
				MoodCode:  "EVN",
				TemplateIDs: []TemplateID{{Root: OIDEncounterEntry}},
				IDs:       []InstanceID{{Root: uuid.New().String()}},
				Code: &Code{
					Code:        encCode,
					CodeSystem:  encSystem,
					DisplayName: encType,
				},
				StatusCode: &Code{Code: status},
				EffectiveTime: buildTimeRangeFromFHIR(dateStr, ""),
			},
		}
		entries = append(entries, entry)
	}

	section.Text = buildNarrativeTable(headers, rows)
	section.Entries = entries
	return section
}

// ---- Helpers ----

// newSection creates a Section with standard template ID, code, and title.
func newSection(templateID, loincCode, codeSystemName, title string) Section {
	return Section{
		TemplateIDs: []TemplateID{{Root: templateID}},
		Code: &Code{
			Code:           loincCode,
			CodeSystem:     OIDLOINC,
			CodeSystemName: codeSystemName,
			DisplayName:    title,
		},
		Title: title,
	}
}

// buildNarrativeTable constructs a narrative table from headers and rows.
func buildNarrativeTable(headers []string, rows []NarrativeTr) *Narrative {
	return &Narrative{
		Table: &NarrativeTable{
			Thead: &NarrativeThead{
				Tr: &NarrativeTr{Ths: headers},
			},
			Tbody: &NarrativeTbody{
				Trs: rows,
			},
		},
	}
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// getArray safely extracts a slice from a map.
func getArray(m map[string]interface{}, key string) ([]interface{}, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}
	arr, ok := v.([]interface{})
	return arr, ok
}

// getNestedMap safely extracts a nested map from a map.
func getNestedMap(m map[string]interface{}, key string) (map[string]interface{}, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}
	nested, ok := v.(map[string]interface{})
	return nested, ok
}

// extractCode extracts code and system from a FHIR CodeableConcept at the given key.
func extractCode(resource map[string]interface{}, key string) (string, string) {
	return extractCodeFromCodeableConcept(resource, key)
}

// extractCodeFromCodeableConcept extracts code and system from a FHIR CodeableConcept.
func extractCodeFromCodeableConcept(resource map[string]interface{}, key string) (string, string) {
	cc, ok := getNestedMap(resource, key)
	if !ok {
		return "", ""
	}
	codings, ok := getArray(cc, "coding")
	if !ok || len(codings) == 0 {
		return "", ""
	}
	c, ok := codings[0].(map[string]interface{})
	if !ok {
		return "", ""
	}
	code, _ := getString(c, "code")
	system, _ := getString(c, "system")
	return code, mapFHIRSystemToOID(system)
}

// extractCodeFromArray extracts code and system from the first element of a CodeableConcept array.
func extractCodeFromArray(resource map[string]interface{}, key string) (string, string) {
	arr, ok := getArray(resource, key)
	if !ok || len(arr) == 0 {
		return "", ""
	}
	cc, ok := arr[0].(map[string]interface{})
	if !ok {
		return "", ""
	}
	codings, ok := getArray(cc, "coding")
	if !ok || len(codings) == 0 {
		return "", ""
	}
	c, ok := codings[0].(map[string]interface{})
	if !ok {
		return "", ""
	}
	code, _ := getString(c, "code")
	system, _ := getString(c, "system")
	return code, mapFHIRSystemToOID(system)
}

// extractFHIRValue extracts a display-friendly value from a FHIR Observation.
func extractFHIRValue(obs map[string]interface{}) string {
	if vq, ok := getNestedMap(obs, "valueQuantity"); ok {
		val := ""
		unit := ""
		if v, ok := vq["value"]; ok {
			val = fmt.Sprintf("%v", v)
		}
		if u, ok := getString(vq, "unit"); ok {
			unit = u
		}
		if unit != "" {
			return val + " " + unit
		}
		return val
	}
	if vc, ok := getNestedMap(obs, "valueCodeableConcept"); ok {
		if disp, ok := getString(vc, "text"); ok {
			return disp
		}
		if codings, ok := getArray(vc, "coding"); ok && len(codings) > 0 {
			if c, ok := codings[0].(map[string]interface{}); ok {
				if d, ok := getString(c, "display"); ok {
					return d
				}
			}
		}
	}
	if vs, ok := getString(obs, "valueString"); ok {
		return vs
	}
	return ""
}

// extractQuantityValue extracts the numeric value and unit from a FHIR Observation.
func extractQuantityValue(obs map[string]interface{}) (string, string) {
	if vq, ok := getNestedMap(obs, "valueQuantity"); ok {
		val := ""
		unit := ""
		if v, ok := vq["value"]; ok {
			val = fmt.Sprintf("%v", v)
		}
		if u, ok := getString(vq, "unit"); ok {
			unit = u
		}
		return val, unit
	}
	return "", ""
}

// mapFHIRSystemToOID converts a FHIR system URL to the corresponding OID.
func mapFHIRSystemToOID(system string) string {
	switch system {
	case "http://snomed.info/sct":
		return OIDSNOMED
	case "http://www.nlm.nih.gov/research/umls/rxnorm":
		return OIDRxNorm
	case "http://loinc.org":
		return OIDLOINC
	case "http://hl7.org/fhir/sid/icd-10-cm":
		return OIDICD10
	case "http://hl7.org/fhir/sid/cvx":
		return OIDCVX
	default:
		return system
	}
}

// mapGenderCode maps a FHIR gender string to CDA administrative gender code.
func mapGenderCode(gender string) string {
	switch strings.ToLower(gender) {
	case "male":
		return "M"
	case "female":
		return "F"
	case "other":
		return "UN"
	default:
		return "UN"
	}
}

// formatHL7Time formats a time.Time as an HL7 timestamp (YYYYMMDDHHmmss).
func formatHL7Time(t time.Time) string {
	return t.Format("20060102150405")
}

// buildTimeRangeFromFHIR creates a TimeRange from FHIR date strings.
func buildTimeRangeFromFHIR(low, high string) *TimeRange {
	tr := &TimeRange{}
	if low != "" {
		tr.Low = &TimeLow{Value: strings.ReplaceAll(low[:min(len(low), 10)], "-", "")}
	}
	if high != "" {
		tr.High = &TimeHigh{Value: strings.ReplaceAll(high[:min(len(high), 10)], "-", "")}
	}
	if tr.Low == nil && tr.High == nil {
		return nil
	}
	return tr
}

// buildAddress creates a CDA Address from a FHIR address map.
func buildAddress(addr map[string]interface{}) *Address {
	a := &Address{}
	if lines, ok := getArray(addr, "line"); ok && len(lines) > 0 {
		if line, ok := lines[0].(string); ok {
			a.StreetAddress = line
		}
	}
	if city, ok := getString(addr, "city"); ok {
		a.City = city
	}
	if state, ok := getString(addr, "state"); ok {
		a.State = state
	}
	if zip, ok := getString(addr, "postalCode"); ok {
		a.PostalCode = zip
	}
	if country, ok := getString(addr, "country"); ok {
		a.Country = country
	}
	if use, ok := getString(addr, "use"); ok {
		a.Use = use
	}
	return a
}

// buildName creates a CDA Name from a FHIR name map.
func buildName(n map[string]interface{}) *Name {
	name := &Name{}
	if givens, ok := getArray(n, "given"); ok && len(givens) > 0 {
		if given, ok := givens[0].(string); ok {
			name.Given = given
		}
	}
	if family, ok := getString(n, "family"); ok {
		name.Family = family
	}
	return name
}
