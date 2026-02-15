package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
)

// PatientSearcher abstracts patient search for matching (implemented by identity service adapter).
type PatientSearcher interface {
	SearchByDemographics(ctx context.Context, params map[string]string, limit int) ([]PatientRecord, error)
}

// PatientRecord is a simplified patient for matching (decoupled from domain model).
type PatientRecord struct {
	ID           string
	FHIRResource map[string]interface{} // Full FHIR Patient resource
	FirstName    string
	LastName     string
	BirthDate    string // YYYY-MM-DD
	Gender       string
	MRN          string
	Phone        string
	Email        string
	AddressLine  string
	City         string
	PostalCode   string
	SSNLast4     string
}

// MatchWeights configures the scoring weights for each field.
type MatchWeights struct {
	LastName  float64
	FirstName float64
	BirthDate float64
	Gender    float64
	MRN       float64
	Phone     float64
	Email     float64
	Address   float64
	SSNLast4  float64
}

// MatchResult represents a single match candidate.
type MatchResult struct {
	Patient PatientRecord
	Score   float64 // 0.0 to 1.0
	Grade   string  // "certain", "probable", "possible", "certainly-not"
}

// DefaultMatchWeights returns the default scoring weights.
func DefaultMatchWeights() MatchWeights {
	return MatchWeights{
		LastName:  0.15,
		FirstName: 0.15,
		BirthDate: 0.20,
		Gender:    0.05,
		MRN:       0.20,
		Phone:     0.10,
		Email:     0.05,
		Address:   0.05,
		SSNLast4:  0.05,
	}
}

// PatientMatcher performs probabilistic patient matching.
type PatientMatcher struct {
	searcher PatientSearcher
	weights  MatchWeights
}

// NewPatientMatcher creates a PatientMatcher with default weights.
func NewPatientMatcher(searcher PatientSearcher) *PatientMatcher {
	return &PatientMatcher{
		searcher: searcher,
		weights:  DefaultMatchWeights(),
	}
}

// NewPatientMatcherWithWeights creates a PatientMatcher with custom weights.
func NewPatientMatcherWithWeights(searcher PatientSearcher, weights MatchWeights) *PatientMatcher {
	return &PatientMatcher{
		searcher: searcher,
		weights:  weights,
	}
}

// matchInput holds extracted demographics from a FHIR Patient resource.
type matchInput struct {
	firstName   string
	lastName    string
	birthDate   string
	gender      string
	mrn         string
	ssnLast4    string
	phone       string
	email       string
	addressLine string
	city        string
	postalCode  string
}

// Match performs patient matching against a FHIR Patient input.
func (m *PatientMatcher) Match(ctx context.Context, inputPatient map[string]interface{}, count int, onlyCertainMatches bool) ([]MatchResult, error) {
	if inputPatient == nil {
		return nil, fmt.Errorf("input patient resource is required")
	}

	input := extractMatchInput(inputPatient)

	// Build search params from extracted demographics.
	searchParams := buildSearchParams(input)

	// Search for candidates (fetch more than count for scoring).
	searchLimit := count * 5
	if searchLimit < 20 {
		searchLimit = 20
	}

	candidates, err := m.searcher.SearchByDemographics(ctx, searchParams, searchLimit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(candidates) == 0 {
		return []MatchResult{}, nil
	}

	// Score each candidate.
	var results []MatchResult
	for _, candidate := range candidates {
		score := m.scoreCandidate(input, candidate)
		grade := assignGrade(score)

		// Filter out "certainly-not" matches.
		if grade == "certainly-not" {
			continue
		}

		// If only certain matches requested, filter.
		if onlyCertainMatches && grade != "certain" {
			continue
		}

		results = append(results, MatchResult{
			Patient: candidate,
			Score:   score,
			Grade:   grade,
		})
	}

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to count.
	if len(results) > count {
		results = results[:count]
	}

	return results, nil
}

// scoreCandidate computes a weighted score comparing input to a candidate.
func (m *PatientMatcher) scoreCandidate(input *matchInput, candidate PatientRecord) float64 {
	score := 0.0

	// Fuzzy match: LastName
	if input.lastName != "" && candidate.LastName != "" {
		score += m.weights.LastName * jaroWinklerSimilarity(input.lastName, candidate.LastName)
	}

	// Fuzzy match: FirstName
	if input.firstName != "" && candidate.FirstName != "" {
		score += m.weights.FirstName * jaroWinklerSimilarity(input.firstName, candidate.FirstName)
	}

	// Exact match: BirthDate
	if input.birthDate != "" && candidate.BirthDate != "" {
		if strings.EqualFold(input.birthDate, candidate.BirthDate) {
			score += m.weights.BirthDate
		}
	}

	// Exact match: Gender
	if input.gender != "" && candidate.Gender != "" {
		if strings.EqualFold(input.gender, candidate.Gender) {
			score += m.weights.Gender
		}
	}

	// Exact match: MRN
	if input.mrn != "" && candidate.MRN != "" {
		if strings.EqualFold(input.mrn, candidate.MRN) {
			score += m.weights.MRN
		}
	}

	// Partial match: Phone (compare last 4 digits)
	if input.phone != "" && candidate.Phone != "" {
		inputDigits := extractDigits(input.phone)
		candidateDigits := extractDigits(candidate.Phone)
		if len(inputDigits) >= 4 && len(candidateDigits) >= 4 {
			inputLast4 := inputDigits[len(inputDigits)-4:]
			candidateLast4 := candidateDigits[len(candidateDigits)-4:]
			if inputLast4 == candidateLast4 {
				score += m.weights.Phone
			}
		} else if inputDigits == candidateDigits {
			score += m.weights.Phone
		}
	}

	// Exact match: Email
	if input.email != "" && candidate.Email != "" {
		if strings.EqualFold(input.email, candidate.Email) {
			score += m.weights.Email
		}
	}

	// Partial match: Address (normalize and compare)
	if input.addressLine != "" && candidate.AddressLine != "" {
		addrScore := 0.0
		normalizedInput := normalizeAddress(input.addressLine)
		normalizedCandidate := normalizeAddress(candidate.AddressLine)
		if normalizedInput == normalizedCandidate {
			addrScore = 1.0
		} else {
			addrScore = jaroWinklerSimilarity(normalizedInput, normalizedCandidate)
		}
		// Also consider postal code match.
		if input.postalCode != "" && candidate.PostalCode != "" {
			if input.postalCode == candidate.PostalCode {
				addrScore = (addrScore + 1.0) / 2.0
			}
		}
		score += m.weights.Address * addrScore
	}

	// Exact match: SSNLast4
	if input.ssnLast4 != "" && candidate.SSNLast4 != "" {
		if input.ssnLast4 == candidate.SSNLast4 {
			score += m.weights.SSNLast4
		}
	}

	return math.Round(score*1000) / 1000
}

// assignGrade returns the FHIR match grade for a given score.
func assignGrade(score float64) string {
	switch {
	case score >= 0.95:
		return "certain"
	case score >= 0.80:
		return "probable"
	case score >= 0.60:
		return "possible"
	default:
		return "certainly-not"
	}
}

// buildSearchParams creates search parameters from match input.
func buildSearchParams(input *matchInput) map[string]string {
	params := make(map[string]string)
	if input.lastName != "" {
		params["family"] = input.lastName
	}
	if input.firstName != "" {
		params["given"] = input.firstName
	}
	if input.birthDate != "" {
		params["birthdate"] = input.birthDate
	}
	if input.mrn != "" {
		params["identifier"] = input.mrn
	}
	return params
}

// extractMatchInput extracts demographic fields from a FHIR Patient map.
func extractMatchInput(patient map[string]interface{}) *matchInput {
	input := &matchInput{}

	// Extract name[0].family and name[0].given[0].
	if nameVal, ok := patient["name"]; ok {
		if names, ok := nameVal.([]interface{}); ok && len(names) > 0 {
			if name, ok := names[0].(map[string]interface{}); ok {
				if family, ok := name["family"].(string); ok {
					input.lastName = family
				}
				if givenVal, ok := name["given"]; ok {
					if givens, ok := givenVal.([]interface{}); ok && len(givens) > 0 {
						if given, ok := givens[0].(string); ok {
							input.firstName = given
						}
					}
				}
			}
		}
	}

	// Extract birthDate.
	if bd, ok := patient["birthDate"].(string); ok {
		input.birthDate = bd
	}

	// Extract gender.
	if g, ok := patient["gender"].(string); ok {
		input.gender = g
	}

	// Extract identifiers (MRN and SSN).
	if identVal, ok := patient["identifier"]; ok {
		if idents, ok := identVal.([]interface{}); ok {
			for _, identRaw := range idents {
				ident, ok := identRaw.(map[string]interface{})
				if !ok {
					continue
				}
				system, _ := ident["system"].(string)
				value, _ := ident["value"].(string)

				systemLower := strings.ToLower(system)
				if strings.Contains(systemLower, "mrn") {
					input.mrn = value
				} else if strings.Contains(systemLower, "ssn") {
					// Extract last 4 digits.
					digits := extractDigits(value)
					if len(digits) >= 4 {
						input.ssnLast4 = digits[len(digits)-4:]
					}
				}
			}
		}
	}

	// Extract telecom (phone and email).
	if telecomVal, ok := patient["telecom"]; ok {
		if telecoms, ok := telecomVal.([]interface{}); ok {
			for _, tcRaw := range telecoms {
				tc, ok := tcRaw.(map[string]interface{})
				if !ok {
					continue
				}
				system, _ := tc["system"].(string)
				value, _ := tc["value"].(string)

				switch system {
				case "phone":
					if input.phone == "" {
						input.phone = value
					}
				case "email":
					if input.email == "" {
						input.email = value
					}
				}
			}
		}
	}

	// Extract address[0].
	if addrVal, ok := patient["address"]; ok {
		if addrs, ok := addrVal.([]interface{}); ok && len(addrs) > 0 {
			if addr, ok := addrs[0].(map[string]interface{}); ok {
				if lineVal, ok := addr["line"]; ok {
					if lines, ok := lineVal.([]interface{}); ok && len(lines) > 0 {
						if line, ok := lines[0].(string); ok {
							input.addressLine = line
						}
					}
				}
				if city, ok := addr["city"].(string); ok {
					input.city = city
				}
				if pc, ok := addr["postalCode"].(string); ok {
					input.postalCode = pc
				}
			}
		}
	}

	return input
}

// extractDigits returns only the digit characters from a string.
func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalizeAddress lowercases and removes extra whitespace and punctuation from an address.
func normalizeAddress(addr string) string {
	addr = strings.ToLower(addr)
	addr = strings.Map(func(r rune) rune {
		if r == '.' || r == ',' || r == '#' {
			return -1
		}
		return r
	}, addr)
	fields := strings.Fields(addr)
	return strings.Join(fields, " ")
}

// jaroWinklerSimilarity computes the Jaro-Winkler similarity between two strings (case-insensitive).
// Returns a value between 0.0 and 1.0.
func jaroWinklerSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	if s1 == s2 {
		return 1.0
	}

	// Jaro distance.
	s1Len := len(s1)
	s2Len := len(s2)

	maxDist := 0
	if s1Len > s2Len {
		maxDist = s1Len
	} else {
		maxDist = s2Len
	}
	maxDist = maxDist/2 - 1
	if maxDist < 0 {
		maxDist = 0
	}

	s1Matches := make([]bool, s1Len)
	s2Matches := make([]bool, s2Len)

	matches := 0
	transpositions := 0

	for i := 0; i < s1Len; i++ {
		start := i - maxDist
		if start < 0 {
			start = 0
		}
		end := i + maxDist + 1
		if end > s2Len {
			end = s2Len
		}

		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	k := 0
	for i := 0; i < s1Len; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	jaro := (float64(matches)/float64(s1Len) +
		float64(matches)/float64(s2Len) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	// Winkler adjustment: boost for common prefix (up to 4 chars).
	prefixLen := 0
	maxPrefix := 4
	if s1Len < maxPrefix {
		maxPrefix = s1Len
	}
	if s2Len < maxPrefix {
		maxPrefix = s2Len
	}
	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	return jaro + float64(prefixLen)*0.1*(1.0-jaro)
}

// MatchHandler provides the FHIR Patient/$match HTTP endpoint.
type MatchHandler struct {
	matcher *PatientMatcher
}

// NewMatchHandler creates a new MatchHandler.
func NewMatchHandler(matcher *PatientMatcher) *MatchHandler {
	return &MatchHandler{matcher: matcher}
}

// RegisterRoutes adds the $match route to the given FHIR group.
func (h *MatchHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/Patient/$match", h.HandleMatch)
}

// HandleMatch handles POST /fhir/Patient/$match.
// Expects a FHIR Parameters resource in the request body.
func (h *MatchHandler) HandleMatch(c echo.Context) error {
	// Read the request body.
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeInvalid, "Failed to read request body",
		))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeInvalid, "Request body is empty",
		))
	}

	// Parse the JSON.
	var params map[string]interface{}
	if err := json.Unmarshal(body, &params); err != nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeStructure, "Invalid JSON: "+err.Error(),
		))
	}

	// Validate resourceType is Parameters.
	rt, _ := params["resourceType"].(string)
	if rt != "Parameters" {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeStructure, "Expected resourceType 'Parameters'",
		))
	}

	// Extract parameters.
	paramList, ok := params["parameter"].([]interface{})
	if !ok {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeStructure, "Expected 'parameter' array in Parameters resource",
		))
	}

	var patientResource map[string]interface{}
	count := 5
	onlyCertainMatches := false

	for _, pRaw := range paramList {
		p, ok := pRaw.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := p["name"].(string)
		switch name {
		case "resource":
			if res, ok := p["resource"].(map[string]interface{}); ok {
				patientResource = res
			}
		case "count":
			if v, ok := p["valueInteger"].(float64); ok {
				count = int(v)
			}
		case "onlyCertainMatches":
			if v, ok := p["valueBoolean"].(bool); ok {
				onlyCertainMatches = v
			}
		}
	}

	if patientResource == nil {
		return c.JSON(http.StatusBadRequest, NewOperationOutcome(
			IssueSeverityError, IssueTypeRequired, "Missing 'resource' parameter with Patient resource",
		))
	}

	// Perform matching.
	results, err := h.matcher.Match(c.Request().Context(), patientResource, count, onlyCertainMatches)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
			IssueSeverityError, IssueTypeProcessing, "Match operation failed: "+err.Error(),
		))
	}

	// Build the FHIR Bundle response.
	bundle := buildMatchBundle(results)

	return c.JSON(http.StatusOK, bundle)
}

// buildMatchBundle creates a FHIR searchset Bundle from match results.
func buildMatchBundle(results []MatchResult) map[string]interface{} {
	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        len(results),
	}

	entries := make([]interface{}, 0, len(results))
	for _, r := range results {
		resource := r.Patient.FHIRResource
		if resource == nil {
			resource = map[string]interface{}{
				"resourceType": "Patient",
				"id":           r.Patient.ID,
			}
		}

		entry := map[string]interface{}{
			"resource": resource,
			"search": map[string]interface{}{
				"mode":  "match",
				"score": r.Score,
				"extension": []interface{}{
					map[string]interface{}{
						"url":       "http://hl7.org/fhir/StructureDefinition/match-grade",
						"valueCode": r.Grade,
					},
				},
			},
		}

		entries = append(entries, entry)
	}

	bundle["entry"] = entries

	return bundle
}
