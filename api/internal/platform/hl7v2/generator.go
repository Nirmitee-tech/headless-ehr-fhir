package hl7v2

import (
	"fmt"
	"strings"
	"time"
)

// GenerateADT generates an ADT (Admit/Discharge/Transfer) HL7v2 message.
// event is the ADT event code: "A01" (admit), "A02" (transfer), "A03" (discharge),
// "A04" (register), "A08" (update).
// patient is a FHIR Patient resource as a map.
// encounter is a FHIR Encounter resource as a map.
func GenerateADT(event string, patient map[string]interface{}, encounter map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}

	var segments []string

	segments = append(segments, buildMSH("ADT", event))
	segments = append(segments, buildEVN(event))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildPV1(encounter))

	return []byte(strings.Join(segments, "\r")), nil
}

// GenerateORM generates an ORM (Order) HL7v2 message.
// serviceRequest is a FHIR ServiceRequest resource as a map.
// patient is a FHIR Patient resource as a map.
func GenerateORM(serviceRequest map[string]interface{}, patient map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}

	var segments []string

	segments = append(segments, buildMSH("ORM", "O01"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildORC(serviceRequest))
	segments = append(segments, buildOBR(serviceRequest))

	return []byte(strings.Join(segments, "\r")), nil
}

// GenerateORU generates an ORU (Observation Result) HL7v2 message.
// diagnosticReport is a FHIR DiagnosticReport resource as a map.
// observations is a slice of FHIR Observation resources.
// patient is a FHIR Patient resource as a map.
func GenerateORU(diagnosticReport map[string]interface{}, observations []map[string]interface{}, patient map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}

	var segments []string

	segments = append(segments, buildMSH("ORU", "R01"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildOBRFromReport(diagnosticReport))

	for i, obs := range observations {
		segments = append(segments, buildOBX(i+1, obs))
	}

	return []byte(strings.Join(segments, "\r")), nil
}

// buildMSH constructs an MSH segment header for the given message type and trigger event.
func buildMSH(msgType, trigger string) string {
	now := time.Now().UTC()
	timestamp := now.Format("20060102150405")
	controlID := fmt.Sprintf("MSG%s", now.Format("20060102150405.000"))

	return fmt.Sprintf("MSH|^~\\&|EHR|EHRFac|Destination|DestFac|%s||%s^%s|%s|P|2.5.1",
		timestamp, msgType, trigger, controlID)
}

// buildEVN constructs an EVN (event type) segment.
func buildEVN(event string) string {
	now := time.Now().UTC()
	timestamp := now.Format("20060102150405")
	return fmt.Sprintf("EVN|%s|%s", event, timestamp)
}

// buildPID constructs a PID (patient identification) segment from a FHIR Patient resource.
func buildPID(patient map[string]interface{}) string {
	if patient == nil {
		return "PID|1"
	}

	// PID-3: Patient Identifier
	patientID := ""
	if ids, ok := getArray(patient, "identifier"); ok && len(ids) > 0 {
		if id, ok := ids[0].(map[string]interface{}); ok {
			if val, ok := getString(id, "value"); ok {
				patientID = escapeHL7(val)
			}
		}
	}

	// PID-5: Patient Name (family^given)
	patientName := ""
	if names, ok := getArray(patient, "name"); ok && len(names) > 0 {
		if name, ok := names[0].(map[string]interface{}); ok {
			family := ""
			given := ""
			if f, ok := getString(name, "family"); ok {
				family = escapeHL7(f)
			}
			if givens, ok := getArray(name, "given"); ok && len(givens) > 0 {
				if g, ok := givens[0].(string); ok {
					given = escapeHL7(g)
				}
			}
			patientName = family + "^" + given
		}
	}

	// PID-7: Date of Birth
	dob := ""
	if birthDate, ok := getString(patient, "birthDate"); ok {
		dob = strings.ReplaceAll(birthDate, "-", "")
	}

	// PID-8: Gender
	gender := ""
	if g, ok := getString(patient, "gender"); ok {
		gender = mapFHIRGender(g)
	}

	// PID-11: Address
	address := ""
	if addrs, ok := getArray(patient, "address"); ok && len(addrs) > 0 {
		if addr, ok := addrs[0].(map[string]interface{}); ok {
			address = buildHL7Address(addr)
		}
	}

	// PID-13: Phone
	phone := ""
	if telecoms, ok := getArray(patient, "telecom"); ok && len(telecoms) > 0 {
		if t, ok := telecoms[0].(map[string]interface{}); ok {
			if val, ok := getString(t, "value"); ok {
				phone = escapeHL7(val)
			}
		}
	}

	return fmt.Sprintf("PID|1||%s||%s||%s|%s|||%s||%s",
		patientID, patientName, dob, gender, address, phone)
}

// buildPV1 constructs a PV1 (patient visit) segment from a FHIR Encounter resource.
func buildPV1(encounter map[string]interface{}) string {
	if encounter == nil {
		return "PV1|1"
	}

	// PV1-2: Patient Class
	patientClass := ""
	if classObj, ok := getNestedMap(encounter, "class"); ok {
		if code, ok := getString(classObj, "code"); ok {
			patientClass = mapEncounterClass(code)
		}
	}

	// PV1-3: Assigned Patient Location
	location := ""
	if locs, ok := getArray(encounter, "location"); ok && len(locs) > 0 {
		if loc, ok := locs[0].(map[string]interface{}); ok {
			if locRef, ok := getNestedMap(loc, "location"); ok {
				if disp, ok := getString(locRef, "display"); ok {
					location = escapeHL7(disp)
				}
			}
		}
	}

	// PV1-7: Attending Doctor
	attending := ""
	if participants, ok := getArray(encounter, "participant"); ok && len(participants) > 0 {
		if p, ok := participants[0].(map[string]interface{}); ok {
			if ind, ok := getNestedMap(p, "individual"); ok {
				if disp, ok := getString(ind, "display"); ok {
					attending = escapeHL7(disp)
				}
			}
		}
	}

	return fmt.Sprintf("PV1|1|%s|%s||||%s",
		patientClass, location, attending)
}

// buildORC constructs an ORC (common order) segment from a FHIR ServiceRequest resource.
func buildORC(serviceRequest map[string]interface{}) string {
	orderID := ""
	if serviceRequest != nil {
		if id, ok := getString(serviceRequest, "id"); ok {
			orderID = escapeHL7(id)
		}
	}

	timestamp := ""
	if serviceRequest != nil {
		if authored, ok := getString(serviceRequest, "authoredOn"); ok {
			timestamp = convertFHIRDateTimeToHL7(authored)
		}
	}

	return fmt.Sprintf("ORC|NW|%s||||||||%s", orderID, timestamp)
}

// buildOBR constructs an OBR (observation request) segment from a FHIR ServiceRequest resource.
func buildOBR(serviceRequest map[string]interface{}) string {
	code := ""
	display := ""
	system := ""

	if serviceRequest != nil {
		if codeObj, ok := getNestedMap(serviceRequest, "code"); ok {
			if codings, ok := getArray(codeObj, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if cd, ok := getString(c, "code"); ok {
						code = cd
					}
					if d, ok := getString(c, "display"); ok {
						display = d
					}
					if s, ok := getString(c, "system"); ok {
						system = mapFHIRSystemToShort(s)
					}
				}
			}
		}
	}

	universalID := ""
	if code != "" {
		universalID = escapeHL7(code) + "^" + escapeHL7(display) + "^" + escapeHL7(system)
	}

	timestamp := ""
	if serviceRequest != nil {
		if authored, ok := getString(serviceRequest, "authoredOn"); ok {
			timestamp = convertFHIRDateTimeToHL7(authored)
		}
	}

	return fmt.Sprintf("OBR|1|%s||%s|||%s",
		"", universalID, timestamp)
}

// buildOBRFromReport constructs an OBR segment from a FHIR DiagnosticReport resource.
func buildOBRFromReport(report map[string]interface{}) string {
	code := ""
	display := ""
	system := ""

	if report != nil {
		if codeObj, ok := getNestedMap(report, "code"); ok {
			if codings, ok := getArray(codeObj, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if cd, ok := getString(c, "code"); ok {
						code = cd
					}
					if d, ok := getString(c, "display"); ok {
						display = d
					}
					if s, ok := getString(c, "system"); ok {
						system = mapFHIRSystemToShort(s)
					}
				}
			}
		}
	}

	universalID := ""
	if code != "" {
		universalID = escapeHL7(code) + "^" + escapeHL7(display) + "^" + escapeHL7(system)
	}

	timestamp := ""
	if report != nil {
		if dt, ok := getString(report, "effectiveDateTime"); ok {
			timestamp = convertFHIRDateTimeToHL7(dt)
		}
	}

	return fmt.Sprintf("OBR|1|||%s|||%s", universalID, timestamp)
}

// buildOBX constructs an OBX (observation result) segment from a FHIR Observation resource.
func buildOBX(setID int, obs map[string]interface{}) string {
	// OBX-2: Value Type
	valueType := "NM" // default to numeric

	// OBX-3: Observation Identifier
	code := ""
	display := ""
	system := ""
	if obs != nil {
		if codeObj, ok := getNestedMap(obs, "code"); ok {
			if codings, ok := getArray(codeObj, "coding"); ok && len(codings) > 0 {
				if c, ok := codings[0].(map[string]interface{}); ok {
					if cd, ok := getString(c, "code"); ok {
						code = cd
					}
					if d, ok := getString(c, "display"); ok {
						display = d
					}
					if s, ok := getString(c, "system"); ok {
						system = mapFHIRSystemToShort(s)
					}
				}
			}
		}
	}

	observationID := ""
	if code != "" {
		observationID = escapeHL7(code) + "^" + escapeHL7(display) + "^" + escapeHL7(system)
	}

	// OBX-5: Observation Value
	value := ""
	// OBX-6: Units
	unit := ""

	if obs != nil {
		if vq, ok := getNestedMap(obs, "valueQuantity"); ok {
			if v, exists := vq["value"]; exists {
				value = fmt.Sprintf("%v", v)
			}
			if u, ok := getString(vq, "unit"); ok {
				unit = u
			}
		} else if vs, ok := getString(obs, "valueString"); ok {
			valueType = "ST"
			value = escapeHL7(vs)
		}
	}

	// OBX-7: Reference Range
	refRange := ""
	if obs != nil {
		if ranges, ok := getArray(obs, "referenceRange"); ok && len(ranges) > 0 {
			if rr, ok := ranges[0].(map[string]interface{}); ok {
				low := ""
				high := ""
				if lowObj, ok := getNestedMap(rr, "low"); ok {
					if v, exists := lowObj["value"]; exists {
						low = fmt.Sprintf("%v", v)
					}
				}
				if highObj, ok := getNestedMap(rr, "high"); ok {
					if v, exists := highObj["value"]; exists {
						high = fmt.Sprintf("%v", v)
					}
				}
				if low != "" || high != "" {
					refRange = low + "-" + high
				}
			}
		}
	}

	// OBX-8: Abnormal Flags
	abnormalFlag := "N"

	// OBX-11: Observation Result Status
	status := "F"
	if obs != nil {
		if s, ok := getString(obs, "status"); ok {
			status = mapObservationStatus(s)
		}
	}

	return fmt.Sprintf("OBX|%d|%s|%s||%s|%s|%s|%s|||%s",
		setID, valueType, observationID, value, unit, refRange, abnormalFlag, status)
}

// escapeHL7 escapes HL7 special characters in a string.
// The HL7 escape sequences are:
//
//	\F\ = |  (field separator)
//	\S\ = ^  (component separator)
//	\R\ = ~  (repetition separator)
//	\E\ = \  (escape character)
//	\T\ = &  (subcomponent separator)
func escapeHL7(s string) string {
	// Escape backslash first to avoid double-escaping
	s = strings.ReplaceAll(s, "\\", "\\E\\")
	s = strings.ReplaceAll(s, "|", "\\F\\")
	s = strings.ReplaceAll(s, "^", "\\S\\")
	s = strings.ReplaceAll(s, "~", "\\R\\")
	s = strings.ReplaceAll(s, "&", "\\T\\")
	return s
}

// ---- FHIR Map Accessor Helpers ----

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

// ---- Mapping Helpers ----

// mapFHIRGender converts a FHIR gender string to HL7v2 administrative sex code.
func mapFHIRGender(gender string) string {
	switch strings.ToLower(gender) {
	case "male":
		return "M"
	case "female":
		return "F"
	case "other":
		return "O"
	case "unknown":
		return "U"
	default:
		return "U"
	}
}

// mapEncounterClass maps a FHIR Encounter class code to HL7v2 patient class.
func mapEncounterClass(code string) string {
	switch strings.ToUpper(code) {
	case "IMP":
		return "I"
	case "AMB":
		return "O"
	case "EMER":
		return "E"
	default:
		return code
	}
}

// mapFHIRSystemToShort converts a FHIR code system URL to a short identifier.
func mapFHIRSystemToShort(system string) string {
	switch system {
	case "http://loinc.org":
		return "LN"
	case "http://snomed.info/sct":
		return "SCT"
	case "http://www.nlm.nih.gov/research/umls/rxnorm":
		return "RXNORM"
	case "http://hl7.org/fhir/sid/icd-10-cm":
		return "I10"
	default:
		return system
	}
}

// mapObservationStatus converts a FHIR observation status to HL7v2 result status.
func mapObservationStatus(status string) string {
	switch status {
	case "final":
		return "F"
	case "preliminary":
		return "P"
	case "cancelled":
		return "X"
	case "corrected":
		return "C"
	default:
		return "F"
	}
}

// convertFHIRDateTimeToHL7 converts a FHIR datetime string to HL7v2 timestamp format.
func convertFHIRDateTimeToHL7(dt string) string {
	// Try full datetime first
	for _, layout := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, dt); err == nil {
			return t.Format("20060102150405")
		}
	}
	// Fallback: remove common separators
	result := strings.ReplaceAll(dt, "-", "")
	result = strings.ReplaceAll(result, "T", "")
	result = strings.ReplaceAll(result, ":", "")
	result = strings.ReplaceAll(result, "Z", "")
	return result
}

// buildHL7Address constructs an HL7v2 address string from a FHIR address map.
// Format: street^other^city^state^zip^country
func buildHL7Address(addr map[string]interface{}) string {
	street := ""
	if lines, ok := getArray(addr, "line"); ok && len(lines) > 0 {
		if line, ok := lines[0].(string); ok {
			street = escapeHL7(line)
		}
	}

	city := ""
	if c, ok := getString(addr, "city"); ok {
		city = escapeHL7(c)
	}

	state := ""
	if s, ok := getString(addr, "state"); ok {
		state = escapeHL7(s)
	}

	zip := ""
	if z, ok := getString(addr, "postalCode"); ok {
		zip = escapeHL7(z)
	}

	country := ""
	if c, ok := getString(addr, "country"); ok {
		country = escapeHL7(c)
	}

	return fmt.Sprintf("%s^^%s^%s^%s^%s", street, city, state, zip, country)
}
