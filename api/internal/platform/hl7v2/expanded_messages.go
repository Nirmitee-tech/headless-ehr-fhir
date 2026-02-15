package hl7v2

import (
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// ADT A40 — Merge Patient
// ---------------------------------------------------------------------------

// GenerateADT_A40 generates an ADT^A40 (Merge Patient) HL7v2 message.
// patient is the surviving patient record (FHIR Patient as a map).
// mergeParams must contain "priorPatientID" and optionally "priorAccountID".
func GenerateADT_A40(patient map[string]interface{}, mergeParams map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}
	if mergeParams == nil {
		return nil, fmt.Errorf("hl7v2: merge parameters are required")
	}

	var segments []string

	segments = append(segments, buildMSH("ADT", "A40"))
	segments = append(segments, buildEVN("A40"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildMRG(mergeParams))

	return []byte(strings.Join(segments, "\r")), nil
}

// ---------------------------------------------------------------------------
// ADT A41 — Merge Account
// ---------------------------------------------------------------------------

// GenerateADT_A41 generates an ADT^A41 (Merge Account) HL7v2 message.
// patient is the patient whose accounts are being merged.
// mergeParams must contain "priorAccountID".
func GenerateADT_A41(patient map[string]interface{}, mergeParams map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}
	if mergeParams == nil {
		return nil, fmt.Errorf("hl7v2: merge parameters are required")
	}

	var segments []string

	segments = append(segments, buildMSH("ADT", "A41"))
	segments = append(segments, buildEVN("A41"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildMRG(mergeParams))

	return []byte(strings.Join(segments, "\r")), nil
}

// ---------------------------------------------------------------------------
// RGV O15 — Pharmacy/Treatment Give
// ---------------------------------------------------------------------------

// GenerateRGV_O15 generates an RGV^O15 (Pharmacy/Treatment Give) HL7v2 message.
// patient is a FHIR Patient resource as a map.
// order contains pharmacy order fields: "orderControl", "orderID".
// give contains pharmacy give fields: "giveCode", "giveCodeText", "giveCodeSystem",
// "giveAmount", "giveUnits", "giveDosageForm".
func GenerateRGV_O15(patient map[string]interface{}, order map[string]interface{}, give map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}
	if order == nil {
		return nil, fmt.Errorf("hl7v2: pharmacy order is required")
	}
	if give == nil {
		return nil, fmt.Errorf("hl7v2: pharmacy give data is required")
	}

	var segments []string

	segments = append(segments, buildMSH("RGV", "O15"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildPharmacyORC(order))
	segments = append(segments, buildRXG(give))

	return []byte(strings.Join(segments, "\r")), nil
}

// ---------------------------------------------------------------------------
// BAR P01 — Add Patient Account
// ---------------------------------------------------------------------------

// GenerateBAR_P01 generates a BAR^P01 (Add Patient Account) HL7v2 message.
// patient is a FHIR Patient resource as a map.
// encounter is a FHIR Encounter resource as a map (used for PV1).
// diagnosis contains diagnosis fields: "code", "description", "type", "codeSystem".
func GenerateBAR_P01(patient map[string]interface{}, encounter map[string]interface{}, diagnosis map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}
	if diagnosis == nil {
		return nil, fmt.Errorf("hl7v2: diagnosis data is required")
	}

	var segments []string

	segments = append(segments, buildMSH("BAR", "P01"))
	segments = append(segments, buildEVN("P01"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildPV1(encounter))
	segments = append(segments, buildDG1(1, diagnosis))

	return []byte(strings.Join(segments, "\r")), nil
}

// ---------------------------------------------------------------------------
// BAR P05 — Update Account
// ---------------------------------------------------------------------------

// GenerateBAR_P05 generates a BAR^P05 (Update Account) HL7v2 message.
// patient is a FHIR Patient resource as a map.
// encounter is a FHIR Encounter resource as a map (used for PV1).
// diagnosis contains diagnosis fields: "code", "description", "type", "codeSystem".
func GenerateBAR_P05(patient map[string]interface{}, encounter map[string]interface{}, diagnosis map[string]interface{}) ([]byte, error) {
	if patient == nil {
		return nil, fmt.Errorf("hl7v2: patient resource is required")
	}
	if diagnosis == nil {
		return nil, fmt.Errorf("hl7v2: diagnosis data is required")
	}

	var segments []string

	segments = append(segments, buildMSH("BAR", "P05"))
	segments = append(segments, buildEVN("P05"))
	segments = append(segments, buildPID(patient))
	segments = append(segments, buildPV1(encounter))
	segments = append(segments, buildDG1(1, diagnosis))

	return []byte(strings.Join(segments, "\r")), nil
}

// ---------------------------------------------------------------------------
// New Segment Builders
// ---------------------------------------------------------------------------

// buildMRG constructs an MRG (merge) segment from merge parameters.
// MRG-1 = prior patient identifier (from "priorPatientID").
// MRG-3 = prior patient account number (from "priorAccountID").
func buildMRG(params map[string]interface{}) string {
	priorPatientID := ""
	if v, ok := params["priorPatientID"]; ok {
		if s, ok := v.(string); ok {
			priorPatientID = escapeHL7(s)
		}
	}

	priorAccountID := ""
	if v, ok := params["priorAccountID"]; ok {
		if s, ok := v.(string); ok {
			priorAccountID = escapeHL7(s)
		}
	}

	// MRG|priorPatientID||priorAccountID
	return fmt.Sprintf("MRG|%s||%s", priorPatientID, priorAccountID)
}

// buildPharmacyORC constructs an ORC (common order) segment for pharmacy messages.
// order must contain "orderControl" (e.g. "RE") and "orderID".
func buildPharmacyORC(order map[string]interface{}) string {
	orderControl := ""
	if v, ok := order["orderControl"]; ok {
		if s, ok := v.(string); ok {
			orderControl = escapeHL7(s)
		}
	}

	orderID := ""
	if v, ok := order["orderID"]; ok {
		if s, ok := v.(string); ok {
			orderID = escapeHL7(s)
		}
	}

	return fmt.Sprintf("ORC|%s|%s", orderControl, orderID)
}

// buildRXG constructs an RXG (pharmacy/treatment give) segment.
// give must contain "giveCode", "giveCodeText", "giveCodeSystem",
// "giveAmount", "giveUnits", "giveDosageForm".
//
// RXG field layout:
//
//	RXG-1  = Give Sub-ID Counter (always 1)
//	RXG-2  = Dispense Sub-ID Counter (empty)
//	RXG-3  = Quantity/Timing (CE: code^text^system)
//	RXG-4  = Give Amount - Minimum
//	RXG-5  = Give Amount - Maximum (empty)
//	RXG-6  = Give Units
//	RXG-7  = Give Dosage Form
func buildRXG(give map[string]interface{}) string {
	giveCode := ""
	if v, ok := give["giveCode"]; ok {
		if s, ok := v.(string); ok {
			giveCode = escapeHL7(s)
		}
	}

	giveCodeText := ""
	if v, ok := give["giveCodeText"]; ok {
		if s, ok := v.(string); ok {
			giveCodeText = escapeHL7(s)
		}
	}

	giveCodeSystem := ""
	if v, ok := give["giveCodeSystem"]; ok {
		if s, ok := v.(string); ok {
			giveCodeSystem = escapeHL7(s)
		}
	}

	giveAmount := ""
	if v, ok := give["giveAmount"]; ok {
		if s, ok := v.(string); ok {
			giveAmount = escapeHL7(s)
		}
	}

	giveUnits := ""
	if v, ok := give["giveUnits"]; ok {
		if s, ok := v.(string); ok {
			giveUnits = escapeHL7(s)
		}
	}

	giveDosageForm := ""
	if v, ok := give["giveDosageForm"]; ok {
		if s, ok := v.(string); ok {
			giveDosageForm = escapeHL7(s)
		}
	}

	// Build coded element for give code
	codedElement := giveCode + "^" + giveCodeText + "^" + giveCodeSystem

	// RXG|1||code^text^system|amount||units|dosageForm
	return fmt.Sprintf("RXG|1||%s|%s||%s|%s",
		codedElement, giveAmount, giveUnits, giveDosageForm)
}

// buildDG1 constructs a DG1 (diagnosis) segment.
// setID is the sequence number (typically 1 for a single diagnosis).
// diagnosis must contain "code", "description", "type", and optionally "codeSystem".
//
// DG1 field layout:
//
//	DG1-1  = Set ID
//	DG1-2  = Diagnosis Coding Method (e.g. "I10" for ICD-10)
//	DG1-3  = Diagnosis Code (CE: code^description^system)
//	DG1-4  = Diagnosis Description (deprecated, empty)
//	DG1-5  = Diagnosis Date/Time
//	DG1-6  = Diagnosis Type (A=admitting, W=working, F=final)
func buildDG1(setID int, diagnosis map[string]interface{}) string {
	code := ""
	if v, ok := diagnosis["code"]; ok {
		if s, ok := v.(string); ok {
			code = escapeHL7(s)
		}
	}

	description := ""
	if v, ok := diagnosis["description"]; ok {
		if s, ok := v.(string); ok {
			description = escapeHL7(s)
		}
	}

	codeSystem := ""
	if v, ok := diagnosis["codeSystem"]; ok {
		if s, ok := v.(string); ok {
			codeSystem = escapeHL7(s)
		}
	}

	diagType := ""
	if v, ok := diagnosis["type"]; ok {
		if s, ok := v.(string); ok {
			diagType = escapeHL7(s)
		}
	}

	now := time.Now().UTC()
	timestamp := now.Format("20060102")

	// Coded element for the diagnosis
	codedElement := code + "^" + description + "^" + codeSystem

	// DG1|setID|codingMethod|code^desc^system||date|type
	return fmt.Sprintf("DG1|%d|%s|%s||%s|%s",
		setID, codeSystem, codedElement, timestamp, diagType)
}
