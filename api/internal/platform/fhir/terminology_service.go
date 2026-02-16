package fhir

import (
	"fmt"
	"strings"
	"sync"
)

// InMemoryTerminologyService provides in-memory terminology operations.
// It serves as both a ValueSetExpander and CodeSystemLookup.
type InMemoryTerminologyService struct {
	mu          sync.RWMutex
	codeSystems map[string]*inMemoryCodeSystem
	valueSets   map[string]*inMemoryValueSet
}

type inMemoryCodeSystem struct {
	URL     string
	Name    string
	Version string
	Codes   map[string]*inMemoryConcept
}

type inMemoryConcept struct {
	Code    string
	Display string
	Parent  string
}

type inMemoryValueSet struct {
	URL     string
	Name    string
	Title   string
	Version string
	Status  string
	Include []inMemoryVSInclude
}

type inMemoryVSInclude struct {
	System string
	Codes  []string
}

// NewInMemoryTerminologyService creates a new terminology service with common code systems.
func NewInMemoryTerminologyService() *InMemoryTerminologyService {
	svc := &InMemoryTerminologyService{
		codeSystems: make(map[string]*inMemoryCodeSystem),
		valueSets:   make(map[string]*inMemoryValueSet),
	}
	svc.registerBuiltins()
	return svc
}

func (s *InMemoryTerminologyService) registerBuiltins() {
	// Observation status codes
	s.registerCodeSystem("http://hl7.org/fhir/observation-status", "ObservationStatus", "4.0.1", map[string]string{
		"registered":        "Registered",
		"preliminary":       "Preliminary",
		"final":             "Final",
		"amended":           "Amended",
		"corrected":         "Corrected",
		"cancelled":         "Cancelled",
		"entered-in-error":  "Entered in Error",
		"unknown":           "Unknown",
	})

	// Administrative gender
	s.registerCodeSystem("http://hl7.org/fhir/administrative-gender", "AdministrativeGender", "4.0.1", map[string]string{
		"male":    "Male",
		"female":  "Female",
		"other":   "Other",
		"unknown": "Unknown",
	})

	// Encounter status
	s.registerCodeSystem("http://hl7.org/fhir/encounter-status", "EncounterStatus", "4.0.1", map[string]string{
		"planned":           "Planned",
		"arrived":           "Arrived",
		"triaged":           "Triaged",
		"in-progress":       "In Progress",
		"onleave":           "On Leave",
		"finished":          "Finished",
		"cancelled":         "Cancelled",
		"entered-in-error":  "Entered in Error",
		"unknown":           "Unknown",
	})

	// Condition clinical status
	s.registerCodeSystem("http://terminology.hl7.org/CodeSystem/condition-clinical", "ConditionClinicalStatusCodes", "4.0.1", map[string]string{
		"active":     "Active",
		"recurrence": "Recurrence",
		"relapse":    "Relapse",
		"inactive":   "Inactive",
		"remission":  "Remission",
		"resolved":   "Resolved",
	})

	// Request status
	s.registerCodeSystem("http://hl7.org/fhir/request-status", "RequestStatus", "4.0.1", map[string]string{
		"draft":            "Draft",
		"active":           "Active",
		"on-hold":          "On Hold",
		"revoked":          "Revoked",
		"completed":        "Completed",
		"entered-in-error": "Entered in Error",
		"unknown":          "Unknown",
	})

	// Publication status (used by many resources)
	s.registerCodeSystem("http://hl7.org/fhir/publication-status", "PublicationStatus", "4.0.1", map[string]string{
		"draft":   "Draft",
		"active":  "Active",
		"retired": "Retired",
		"unknown": "Unknown",
	})

	// Register corresponding value sets
	for url, cs := range s.codeSystems {
		codes := make([]string, 0, len(cs.Codes))
		for code := range cs.Codes {
			codes = append(codes, code)
		}
		s.valueSets[url] = &inMemoryValueSet{
			URL:     url,
			Name:    cs.Name,
			Title:   cs.Name,
			Version: cs.Version,
			Status:  "active",
			Include: []inMemoryVSInclude{{System: url, Codes: codes}},
		}
	}
}

func (s *InMemoryTerminologyService) registerCodeSystem(url, name, version string, codes map[string]string) {
	cs := &inMemoryCodeSystem{
		URL:     url,
		Name:    name,
		Version: version,
		Codes:   make(map[string]*inMemoryConcept),
	}
	for code, display := range codes {
		cs.Codes[code] = &inMemoryConcept{Code: code, Display: display}
	}
	s.codeSystems[url] = cs
}

// LookupCode implements CodeSystemLookup.
func (s *InMemoryTerminologyService) LookupCode(system, code, version string) (*LookupResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cs, ok := s.codeSystems[system]
	if !ok {
		return nil, fmt.Errorf("code system not found: %s", system)
	}

	concept, ok := cs.Codes[code]
	if !ok {
		return nil, fmt.Errorf("code not found: %s in system %s", code, system)
	}

	return &LookupResult{
		Name:    cs.Name,
		Version: cs.Version,
		Display: concept.Display,
	}, nil
}

// ExpandValueSet implements ValueSetExpander.
func (s *InMemoryTerminologyService) ExpandValueSet(urlOrID, filter string, offset, count int) (*ExpandedValueSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vs, ok := s.valueSets[urlOrID]
	if !ok {
		// Try by name
		for _, v := range s.valueSets {
			if v.Name == urlOrID {
				vs = v
				break
			}
		}
		if vs == nil {
			return nil, fmt.Errorf("value set not found: %s", urlOrID)
		}
	}

	// Collect all codes
	var allCodes []ValueSetContains
	for _, inc := range vs.Include {
		cs := s.codeSystems[inc.System]
		if cs == nil {
			continue
		}
		if len(inc.Codes) > 0 {
			for _, code := range inc.Codes {
				concept := cs.Codes[code]
				if concept == nil {
					continue
				}
				if filter != "" && !strings.Contains(strings.ToLower(concept.Display), strings.ToLower(filter)) &&
					!strings.Contains(strings.ToLower(concept.Code), strings.ToLower(filter)) {
					continue
				}
				allCodes = append(allCodes, ValueSetContains{
					System:  inc.System,
					Version: cs.Version,
					Code:    concept.Code,
					Display: concept.Display,
				})
			}
		}
	}

	total := len(allCodes)
	if offset > total {
		offset = total
	}
	end := offset + count
	if end > total {
		end = total
	}

	return &ExpandedValueSet{
		URL:      vs.URL,
		Version:  vs.Version,
		Name:     vs.Name,
		Title:    vs.Title,
		Status:   vs.Status,
		Total:    total,
		Offset:   offset,
		Contains: allCodes[offset:end],
	}, nil
}
