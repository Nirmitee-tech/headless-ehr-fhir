package fhir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// ResourceNarrativeFunc generates narrative XHTML for a specific resource type.
// It returns the inner XHTML content (without the wrapping <div>).
type ResourceNarrativeFunc func(resource map[string]interface{}) (string, error)

// NarrativeGenerator produces human-readable XHTML narratives for FHIR resources.
type NarrativeGenerator struct {
	generators map[string]ResourceNarrativeFunc
}

// NewNarrativeGenerator creates a NarrativeGenerator pre-loaded with built-in
// generators for common FHIR resource types.
func NewNarrativeGenerator() *NarrativeGenerator {
	g := &NarrativeGenerator{
		generators: make(map[string]ResourceNarrativeFunc),
	}
	g.registerBuiltins()
	return g
}

// RegisterGenerator registers (or replaces) a custom narrative generator for a
// given FHIR resource type.
func (g *NarrativeGenerator) RegisterGenerator(resourceType string, fn ResourceNarrativeFunc) {
	g.generators[resourceType] = fn
}

// Generate produces the text element for a FHIR resource.
// Returns nil for nil input.
// Returns {"status": "generated", "div": "<div ...>...</div>"} on success.
func (g *NarrativeGenerator) Generate(resource map[string]interface{}) map[string]interface{} {
	if resource == nil {
		return nil
	}

	resourceType, _ := resource["resourceType"].(string)

	var divContent string

	if fn, ok := g.generators[resourceType]; ok {
		content, err := fn(resource)
		if err == nil {
			divContent = content
		}
	}

	// Fallback if no generator or generator returned an error.
	if divContent == "" {
		divContent = g.fallbackNarrative(resource)
	}

	return map[string]interface{}{
		"status": "generated",
		"div":    divContent,
	}
}

// InjectNarrative adds or replaces the text element on a resource.
// It preserves text with status "additional" or "extensions" and re-generates
// text with status "generated" or "empty".
func (g *NarrativeGenerator) InjectNarrative(resource map[string]interface{}) map[string]interface{} {
	if resource == nil {
		return nil
	}

	// Check for existing text that should be preserved.
	if existing, ok := resource["text"].(map[string]interface{}); ok {
		status, _ := existing["status"].(string)
		if status == "additional" || status == "extensions" {
			return resource
		}
	}

	text := g.Generate(resource)
	if text != nil {
		resource["text"] = text
	}
	return resource
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// NarrativeMiddleware is Echo middleware that auto-injects narratives into FHIR
// responses. It processes single resources and Bundle entries.
func NarrativeMiddleware(generator *NarrativeGenerator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Fast path: skip if _narrative=none.
			if c.QueryParam("_narrative") == "none" {
				return next(c)
			}

			// Capture the response body.
			origWriter := c.Response().Writer
			rec := &narrativeResponseRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				c.Response().Writer = origWriter
				return err
			}

			// Only process JSON responses.
			ct := c.Response().Header().Get("Content-Type")
			if !isJSONContentType(ct) {
				return flushNarrativeOriginal(origWriter, rec)
			}

			// Try to parse as JSON map.
			var resource map[string]interface{}
			if err := json.Unmarshal(rec.body.Bytes(), &resource); err != nil {
				return flushNarrativeOriginal(origWriter, rec)
			}

			// If it's a Bundle, inject into each entry's resource.
			if resource["resourceType"] == "Bundle" {
				injectBundleNarratives(generator, resource)
			} else if _, ok := resource["resourceType"]; ok {
				// Single resource.
				resource = generator.InjectNarrative(resource)
			}

			result, err := json.Marshal(resource)
			if err != nil {
				return flushNarrativeOriginal(origWriter, rec)
			}

			c.Response().Writer = origWriter
			c.Response().Header().Set("Content-Type", "application/fhir+json")
			_, writeErr := origWriter.Write(result)
			return writeErr
		}
	}
}

// narrativeResponseRecorder captures the response body for post-processing.
type narrativeResponseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *narrativeResponseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *narrativeResponseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func flushNarrativeOriginal(w http.ResponseWriter, rec *narrativeResponseRecorder) error {
	_, err := w.Write(rec.body.Bytes())
	return err
}

func isJSONContentType(ct string) bool {
	return strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "application/fhir+json")
}

func injectBundleNarratives(generator *NarrativeGenerator, bundle map[string]interface{}) {
	entries, ok := bundle["entry"].([]interface{})
	if !ok {
		return
	}
	for i, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		res, ok := entryMap["resource"].(map[string]interface{})
		if !ok {
			continue
		}
		entryMap["resource"] = generator.InjectNarrative(res)
		entries[i] = entryMap
	}
	bundle["entry"] = entries
}

// ---------------------------------------------------------------------------
// Built-in generators
// ---------------------------------------------------------------------------

func (g *NarrativeGenerator) registerBuiltins() {
	g.generators["Patient"] = narrativePatient
	g.generators["Condition"] = narrativeCondition
	g.generators["Observation"] = narrativeObservation
	g.generators["AllergyIntolerance"] = narrativeAllergyIntolerance
	g.generators["MedicationRequest"] = narrativeMedicationRequest
	g.generators["Encounter"] = narrativeEncounter
	g.generators["Procedure"] = narrativeProcedure
	g.generators["Immunization"] = narrativeImmunization
	g.generators["DiagnosticReport"] = narrativeDiagnosticReport
	g.generators["DocumentReference"] = narrativeDocumentReference
}

// fallbackNarrative generates a generic narrative for unknown resource types.
func (g *NarrativeGenerator) fallbackNarrative(resource map[string]interface{}) string {
	rt := escapeHTML(strVal(resource, "resourceType"))
	id := escapeHTML(strVal(resource, "id"))

	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)
	b.WriteString(fmt.Sprintf("<p><b>%s</b> %s</p>", rt, id))
	b.WriteString("</div>")
	return b.String()
}

// ---------------------------------------------------------------------------
// Patient
// ---------------------------------------------------------------------------

func narrativePatient(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	// Name
	family, given := extractName(resource)
	if family != "" || given != "" {
		b.WriteString(fmt.Sprintf("<p><b>Patient:</b> %s, %s</p>", escapeHTML(family), escapeHTML(given)))
	} else {
		b.WriteString("<p><b>Patient:</b> (unknown)</p>")
	}

	// Gender & DOB
	gender := escapeHTML(strVal(resource, "gender"))
	dob := escapeHTML(strVal(resource, "birthDate"))
	if gender != "" || dob != "" {
		b.WriteString("<p>")
		if gender != "" {
			b.WriteString(fmt.Sprintf("<b>Gender:</b> %s", gender))
		}
		if gender != "" && dob != "" {
			b.WriteString(" | ")
		}
		if dob != "" {
			b.WriteString(fmt.Sprintf("<b>DOB:</b> %s", dob))
		}
		b.WriteString("</p>")
	}

	// Identifiers
	if ids := extractIdentifiers(resource); len(ids) > 0 {
		b.WriteString("<p><b>Identifiers:</b> ")
		b.WriteString(strings.Join(ids, ", "))
		b.WriteString("</p>")
	}

	// Telecom
	if telecoms := extractTelecoms(resource); len(telecoms) > 0 {
		b.WriteString("<p><b>Contact:</b> ")
		b.WriteString(strings.Join(telecoms, ", "))
		b.WriteString("</p>")
	}

	// Address
	if addr := extractAddress(resource); addr != "" {
		b.WriteString(fmt.Sprintf("<p><b>Address:</b> %s</p>", addr))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Condition
// ---------------------------------------------------------------------------

func narrativeCondition(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	display, code := extractCodeCoding(resource, "code")
	if display != "" || code != "" {
		b.WriteString(fmt.Sprintf("<p><b>Condition:</b> %s (%s)</p>", escapeHTML(display), escapeHTML(code)))
	}

	clinicalStatus := extractCodeableConceptCode(resource, "clinicalStatus")
	verificationStatus := extractCodeableConceptCode(resource, "verificationStatus")
	if clinicalStatus != "" || verificationStatus != "" {
		b.WriteString("<p>")
		if clinicalStatus != "" {
			b.WriteString(fmt.Sprintf("<b>Clinical Status:</b> %s", escapeHTML(clinicalStatus)))
		}
		if clinicalStatus != "" && verificationStatus != "" {
			b.WriteString(" | ")
		}
		if verificationStatus != "" {
			b.WriteString(fmt.Sprintf("<b>Verification:</b> %s", escapeHTML(verificationStatus)))
		}
		b.WriteString("</p>")
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	if onset := strVal(resource, "onsetDateTime"); onset != "" {
		b.WriteString(fmt.Sprintf("<p><b>Onset:</b> %s</p>", escapeHTML(onset)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Observation
// ---------------------------------------------------------------------------

func narrativeObservation(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	display, code := extractCodeCoding(resource, "code")
	if display != "" || code != "" {
		b.WriteString(fmt.Sprintf("<p><b>Observation:</b> %s (%s)</p>", escapeHTML(display), escapeHTML(code)))
	}

	if status := strVal(resource, "status"); status != "" {
		b.WriteString(fmt.Sprintf("<p><b>Status:</b> %s</p>", escapeHTML(status)))
	}

	// Value: try valueQuantity, then valueString, then valueCodeableConcept
	valueStr := extractObservationValue(resource)
	if valueStr != "" {
		b.WriteString(fmt.Sprintf("<p><b>Value:</b> %s</p>", escapeHTML(valueStr)))
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	if eff := strVal(resource, "effectiveDateTime"); eff != "" {
		b.WriteString(fmt.Sprintf("<p><b>Effective:</b> %s</p>", escapeHTML(eff)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// AllergyIntolerance
// ---------------------------------------------------------------------------

func narrativeAllergyIntolerance(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	display, _ := extractCodeCoding(resource, "code")
	if display != "" {
		b.WriteString(fmt.Sprintf("<p><b>Allergy:</b> %s</p>", escapeHTML(display)))
	}

	clinicalStatus := extractCodeableConceptCode(resource, "clinicalStatus")
	criticality := strVal(resource, "criticality")
	if clinicalStatus != "" || criticality != "" {
		b.WriteString("<p>")
		if clinicalStatus != "" {
			b.WriteString(fmt.Sprintf("<b>Clinical Status:</b> %s", escapeHTML(clinicalStatus)))
		}
		if clinicalStatus != "" && criticality != "" {
			b.WriteString(" | ")
		}
		if criticality != "" {
			b.WriteString(fmt.Sprintf("<b>Criticality:</b> %s", escapeHTML(criticality)))
		}
		b.WriteString("</p>")
	}

	if ref := extractReference(resource, "patient"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Patient:</b> %s</p>", escapeHTML(ref)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// MedicationRequest
// ---------------------------------------------------------------------------

func narrativeMedicationRequest(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	// Medication: try medicationCodeableConcept, then medicationReference
	medDisplay := ""
	if display, _ := extractCodeCoding(resource, "medicationCodeableConcept"); display != "" {
		medDisplay = display
	} else if medRef, ok := resource["medicationReference"].(map[string]interface{}); ok {
		if d, ok := medRef["display"].(string); ok && d != "" {
			medDisplay = d
		}
	}
	if medDisplay != "" {
		b.WriteString(fmt.Sprintf("<p><b>Medication:</b> %s</p>", escapeHTML(medDisplay)))
	}

	status := strVal(resource, "status")
	intent := strVal(resource, "intent")
	if status != "" || intent != "" {
		b.WriteString("<p>")
		if status != "" {
			b.WriteString(fmt.Sprintf("<b>Status:</b> %s", escapeHTML(status)))
		}
		if status != "" && intent != "" {
			b.WriteString(" | ")
		}
		if intent != "" {
			b.WriteString(fmt.Sprintf("<b>Intent:</b> %s", escapeHTML(intent)))
		}
		b.WriteString("</p>")
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	if authored := strVal(resource, "authoredOn"); authored != "" {
		b.WriteString(fmt.Sprintf("<p><b>Authored:</b> %s</p>", escapeHTML(authored)))
	}

	if dosage := extractDosageText(resource); dosage != "" {
		b.WriteString(fmt.Sprintf("<p><b>Dosage:</b> %s</p>", escapeHTML(dosage)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Encounter
// ---------------------------------------------------------------------------

func narrativeEncounter(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	typeDisplay := ""
	if types, ok := resource["type"].([]interface{}); ok && len(types) > 0 {
		if t, ok := types[0].(map[string]interface{}); ok {
			if codings, ok := t["coding"].([]interface{}); ok && len(codings) > 0 {
				if coding, ok := codings[0].(map[string]interface{}); ok {
					typeDisplay, _ = coding["display"].(string)
				}
			}
		}
	}
	classCode := ""
	if cls, ok := resource["class"].(map[string]interface{}); ok {
		classCode, _ = cls["code"].(string)
	}
	if typeDisplay != "" || classCode != "" {
		b.WriteString(fmt.Sprintf("<p><b>Encounter:</b> %s (%s)</p>", escapeHTML(typeDisplay), escapeHTML(classCode)))
	}

	if status := strVal(resource, "status"); status != "" {
		b.WriteString(fmt.Sprintf("<p><b>Status:</b> %s</p>", escapeHTML(status)))
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	if period, ok := resource["period"].(map[string]interface{}); ok {
		start, _ := period["start"].(string)
		end, _ := period["end"].(string)
		if start != "" || end != "" {
			b.WriteString(fmt.Sprintf("<p><b>Period:</b> %s to %s</p>", escapeHTML(start), escapeHTML(end)))
		}
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Procedure
// ---------------------------------------------------------------------------

func narrativeProcedure(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	display, code := extractCodeCoding(resource, "code")
	if display != "" || code != "" {
		b.WriteString(fmt.Sprintf("<p><b>Procedure:</b> %s (%s)</p>", escapeHTML(display), escapeHTML(code)))
	}

	if status := strVal(resource, "status"); status != "" {
		b.WriteString(fmt.Sprintf("<p><b>Status:</b> %s</p>", escapeHTML(status)))
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	performed := strVal(resource, "performedDateTime")
	if performed == "" {
		if pp, ok := resource["performedPeriod"].(map[string]interface{}); ok {
			performed, _ = pp["start"].(string)
		}
	}
	if performed != "" {
		b.WriteString(fmt.Sprintf("<p><b>Performed:</b> %s</p>", escapeHTML(performed)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Immunization
// ---------------------------------------------------------------------------

func narrativeImmunization(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	display, _ := extractCodeCoding(resource, "vaccineCode")
	if display != "" {
		b.WriteString(fmt.Sprintf("<p><b>Immunization:</b> %s</p>", escapeHTML(display)))
	}

	if status := strVal(resource, "status"); status != "" {
		b.WriteString(fmt.Sprintf("<p><b>Status:</b> %s</p>", escapeHTML(status)))
	}

	if ref := extractReference(resource, "patient"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Patient:</b> %s</p>", escapeHTML(ref)))
	}

	if dt := strVal(resource, "occurrenceDateTime"); dt != "" {
		b.WriteString(fmt.Sprintf("<p><b>Date:</b> %s</p>", escapeHTML(dt)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// DiagnosticReport
// ---------------------------------------------------------------------------

func narrativeDiagnosticReport(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	display, _ := extractCodeCoding(resource, "code")
	if display != "" {
		b.WriteString(fmt.Sprintf("<p><b>Diagnostic Report:</b> %s</p>", escapeHTML(display)))
	}

	if status := strVal(resource, "status"); status != "" {
		b.WriteString(fmt.Sprintf("<p><b>Status:</b> %s</p>", escapeHTML(status)))
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	if eff := strVal(resource, "effectiveDateTime"); eff != "" {
		b.WriteString(fmt.Sprintf("<p><b>Effective:</b> %s</p>", escapeHTML(eff)))
	}

	if conclusion := strVal(resource, "conclusion"); conclusion != "" {
		b.WriteString(fmt.Sprintf("<p><b>Conclusion:</b> %s</p>", escapeHTML(conclusion)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// DocumentReference
// ---------------------------------------------------------------------------

func narrativeDocumentReference(resource map[string]interface{}) (string, error) {
	var b strings.Builder
	b.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)

	// type is a CodeableConcept on DocumentReference
	display, _ := extractCodeCoding(resource, "type")
	if display != "" {
		b.WriteString(fmt.Sprintf("<p><b>Document:</b> %s</p>", escapeHTML(display)))
	}

	if status := strVal(resource, "status"); status != "" {
		b.WriteString(fmt.Sprintf("<p><b>Status:</b> %s</p>", escapeHTML(status)))
	}

	if ref := extractReference(resource, "subject"); ref != "" {
		b.WriteString(fmt.Sprintf("<p><b>Subject:</b> %s</p>", escapeHTML(ref)))
	}

	if dt := strVal(resource, "date"); dt != "" {
		b.WriteString(fmt.Sprintf("<p><b>Date:</b> %s</p>", escapeHTML(dt)))
	}

	b.WriteString("</div>")
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Helpers – safe data extraction
// ---------------------------------------------------------------------------

// escapeHTML escapes all HTML-significant characters in s.
// Uses html.EscapeString which handles <, >, &, ", and '.
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// strVal safely extracts a string value from a map. Returns "" for missing or
// non-string values.
func strVal(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

// extractName extracts the first name's family and given (joined) from a Patient.
func extractName(resource map[string]interface{}) (family, given string) {
	names, ok := resource["name"].([]interface{})
	if !ok || len(names) == 0 {
		return "", ""
	}
	name, ok := names[0].(map[string]interface{})
	if !ok {
		return "", ""
	}
	family, _ = name["family"].(string)

	if givens, ok := name["given"].([]interface{}); ok {
		parts := make([]string, 0, len(givens))
		for _, g := range givens {
			if s, ok := g.(string); ok {
				parts = append(parts, s)
			}
		}
		given = strings.Join(parts, " ")
	}
	return family, given
}

// extractIdentifiers extracts system:value pairs from a Patient's identifier array.
func extractIdentifiers(resource map[string]interface{}) []string {
	ids, ok := resource["identifier"].([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for _, id := range ids {
		idMap, ok := id.(map[string]interface{})
		if !ok {
			continue
		}
		sys, _ := idMap["system"].(string)
		val, _ := idMap["value"].(string)
		if sys != "" || val != "" {
			result = append(result, escapeHTML(sys)+": "+escapeHTML(val))
		}
	}
	return result
}

// extractTelecoms extracts system:value pairs from a Patient's telecom array.
func extractTelecoms(resource map[string]interface{}) []string {
	telecoms, ok := resource["telecom"].([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for _, tc := range telecoms {
		tcMap, ok := tc.(map[string]interface{})
		if !ok {
			continue
		}
		sys, _ := tcMap["system"].(string)
		val, _ := tcMap["value"].(string)
		if sys != "" || val != "" {
			result = append(result, escapeHTML(sys)+": "+escapeHTML(val))
		}
	}
	return result
}

// extractAddress extracts the first address as a formatted string.
func extractAddress(resource map[string]interface{}) string {
	addrs, ok := resource["address"].([]interface{})
	if !ok || len(addrs) == 0 {
		return ""
	}
	addr, ok := addrs[0].(map[string]interface{})
	if !ok {
		return ""
	}

	var parts []string

	if lines, ok := addr["line"].([]interface{}); ok {
		for _, l := range lines {
			if s, ok := l.(string); ok && s != "" {
				parts = append(parts, escapeHTML(s))
			}
		}
	}
	if city, ok := addr["city"].(string); ok && city != "" {
		parts = append(parts, escapeHTML(city))
	}

	statePostal := ""
	if state, ok := addr["state"].(string); ok && state != "" {
		statePostal = escapeHTML(state)
	}
	if postal, ok := addr["postalCode"].(string); ok && postal != "" {
		if statePostal != "" {
			statePostal += " " + escapeHTML(postal)
		} else {
			statePostal = escapeHTML(postal)
		}
	}
	if statePostal != "" {
		parts = append(parts, statePostal)
	}

	return strings.Join(parts, ", ")
}

// extractCodeCoding extracts display and code from a CodeableConcept field.
// Returns (display, code). If display is empty, falls back to code.
func extractCodeCoding(resource map[string]interface{}, field string) (display, code string) {
	cc, ok := resource[field].(map[string]interface{})
	if !ok {
		return "", ""
	}
	codings, ok := cc["coding"].([]interface{})
	if !ok || len(codings) == 0 {
		return "", ""
	}
	coding, ok := codings[0].(map[string]interface{})
	if !ok {
		return "", ""
	}
	display, _ = coding["display"].(string)
	code, _ = coding["code"].(string)
	return display, code
}

// extractCodeableConceptCode extracts the first coding code from a CodeableConcept field.
func extractCodeableConceptCode(resource map[string]interface{}, field string) string {
	cc, ok := resource[field].(map[string]interface{})
	if !ok {
		return ""
	}
	codings, ok := cc["coding"].([]interface{})
	if !ok || len(codings) == 0 {
		return ""
	}
	coding, ok := codings[0].(map[string]interface{})
	if !ok {
		return ""
	}
	code, _ := coding["code"].(string)
	return code
}

// extractReference extracts a Reference.reference value from a field.
func extractReference(resource map[string]interface{}, field string) string {
	ref, ok := resource[field].(map[string]interface{})
	if !ok {
		return ""
	}
	r, _ := ref["reference"].(string)
	return r
}

// extractObservationValue extracts the value from an Observation using the
// priority: valueQuantity > valueString > valueCodeableConcept.
func extractObservationValue(resource map[string]interface{}) string {
	// valueQuantity
	if vq, ok := resource["valueQuantity"].(map[string]interface{}); ok {
		val := ""
		if v, ok := vq["value"]; ok {
			val = fmt.Sprintf("%v", v)
		}
		unit, _ := vq["unit"].(string)
		if val != "" {
			if unit != "" {
				return val + " " + unit
			}
			return val
		}
	}

	// valueString
	if vs, ok := resource["valueString"].(string); ok && vs != "" {
		return vs
	}

	// valueCodeableConcept
	if vcc, ok := resource["valueCodeableConcept"].(map[string]interface{}); ok {
		if text, ok := vcc["text"].(string); ok && text != "" {
			return text
		}
	}

	return ""
}

// extractDosageText extracts the text from the first dosageInstruction.
func extractDosageText(resource map[string]interface{}) string {
	instructions, ok := resource["dosageInstruction"].([]interface{})
	if !ok || len(instructions) == 0 {
		return ""
	}
	first, ok := instructions[0].(map[string]interface{})
	if !ok {
		return ""
	}
	text, _ := first["text"].(string)
	return text
}
