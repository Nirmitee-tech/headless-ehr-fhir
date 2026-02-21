package hipaa

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Disclosure represents a record of PHI being shared with a third party.
// HIPAA Section 164.528 requires maintaining an accounting of disclosures for 6 years.
type Disclosure struct {
	ID              uuid.UUID `json:"id"`
	PatientID       uuid.UUID `json:"patient_id"`
	DisclosedTo     string    `json:"disclosed_to"`      // recipient name/org
	DisclosedToType string    `json:"disclosed_to_type"` // organization, individual, system
	Purpose         string    `json:"purpose"`           // public-health, research, law-enforcement, etc.
	ResourceTypes   []string  `json:"resource_types"`    // what was disclosed
	ResourceIDs     []string  `json:"resource_ids,omitempty"`
	DateDisclosed   time.Time `json:"date_disclosed"`
	DisclosedBy     string    `json:"disclosed_by"` // user who initiated
	Method          string    `json:"method"`        // api, export, fax, mail, portal
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// DisclosurePurpose constants define valid HIPAA disclosure purposes.
// These represent scenarios where PHI may be disclosed to third parties
// outside of treatment, payment, or healthcare operations (TPO).
const (
	PurposePublicHealth    = "public-health"
	PurposeResearch        = "research"
	PurposeLawEnforcement  = "law-enforcement"
	PurposeJudicial        = "judicial"
	PurposeWorkerComp      = "workers-comp"
	PurposeDecedent        = "decedent"
	PurposeOrganDonation   = "organ-donation"
	PurposeHealthOversight = "health-oversight"
	PurposeOther           = "other"
)

// ValidDisclosurePurposes returns the set of valid disclosure purpose values.
func ValidDisclosurePurposes() []string {
	return []string{
		PurposePublicHealth,
		PurposeResearch,
		PurposeLawEnforcement,
		PurposeJudicial,
		PurposeWorkerComp,
		PurposeDecedent,
		PurposeOrganDonation,
		PurposeHealthOversight,
		PurposeOther,
	}
}

// IsValidDisclosurePurpose checks whether a purpose string is a recognized value.
func IsValidDisclosurePurpose(purpose string) bool {
	for _, p := range ValidDisclosurePurposes() {
		if p == purpose {
			return true
		}
	}
	return false
}

// DisclosureStore manages disclosure records using an in-memory store.
// In production this would be backed by a database; this implementation
// is suitable for development, testing, and single-instance deployments.
type DisclosureStore struct {
	mu          sync.RWMutex
	disclosures []*Disclosure
}

// NewDisclosureStore creates a new empty DisclosureStore.
func NewDisclosureStore() *DisclosureStore {
	return &DisclosureStore{
		disclosures: make([]*Disclosure, 0),
	}
}

// Record adds a new disclosure entry. It assigns an ID and CreatedAt if not set.
func (s *DisclosureStore) Record(d *Disclosure) error {
	if d.PatientID == uuid.Nil {
		return fmt.Errorf("disclosure: patient_id is required")
	}
	if d.DisclosedTo == "" {
		return fmt.Errorf("disclosure: disclosed_to is required")
	}
	if d.Purpose == "" {
		return fmt.Errorf("disclosure: purpose is required")
	}

	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.DateDisclosed.IsZero() {
		d.DateDisclosed = time.Now().UTC()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.disclosures = append(s.disclosures, d)
	return nil
}

// ListByPatient returns all disclosures for a patient within the specified time range.
// HIPAA Section 164.528 requires that the accounting cover disclosures from the
// prior 6 years (from the date of the request).
func (s *DisclosureStore) ListByPatient(patientID uuid.UUID, from, to time.Time) ([]*Disclosure, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Disclosure
	for _, d := range s.disclosures {
		if d.PatientID != patientID {
			continue
		}
		if !from.IsZero() && d.DateDisclosed.Before(from) {
			continue
		}
		if !to.IsZero() && d.DateDisclosed.After(to) {
			continue
		}
		result = append(result, d)
	}

	// Sort by date disclosed descending (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].DateDisclosed.After(result[j].DateDisclosed)
	})

	return result, nil
}

// ListAll returns all disclosures with pagination. Returns the page of disclosures
// and the total count.
func (s *DisclosureStore) ListAll(limit, offset int) ([]*Disclosure, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.disclosures)

	// Sort by created_at descending
	sorted := make([]*Disclosure, total)
	copy(sorted, s.disclosures)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	if offset >= total {
		return []*Disclosure{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return sorted[offset:end], total, nil
}

// GetByID returns a single disclosure by ID, or nil if not found.
func (s *DisclosureStore) GetByID(id uuid.UUID) *Disclosure {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.disclosures {
		if d.ID == id {
			return d
		}
	}
	return nil
}
