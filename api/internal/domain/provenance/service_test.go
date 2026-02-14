package provenance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// =========== Mock Repository ===========

type mockProvenanceRepo struct {
	store    map[uuid.UUID]*Provenance
	agents   map[uuid.UUID][]*ProvenanceAgent
	entities map[uuid.UUID][]*ProvenanceEntity
}

func newMockProvenanceRepo() *mockProvenanceRepo {
	return &mockProvenanceRepo{
		store:    make(map[uuid.UUID]*Provenance),
		agents:   make(map[uuid.UUID][]*ProvenanceAgent),
		entities: make(map[uuid.UUID][]*ProvenanceEntity),
	}
}

func (m *mockProvenanceRepo) Create(_ context.Context, p *Provenance) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	m.store[p.ID] = p
	return nil
}

func (m *mockProvenanceRepo) GetByID(_ context.Context, id uuid.UUID) (*Provenance, error) {
	p, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockProvenanceRepo) GetByFHIRID(_ context.Context, fhirID string) (*Provenance, error) {
	for _, p := range m.store {
		if p.FHIRID == fhirID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockProvenanceRepo) Update(_ context.Context, p *Provenance) error {
	if _, ok := m.store[p.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[p.ID] = p
	return nil
}

func (m *mockProvenanceRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockProvenanceRepo) List(_ context.Context, limit, offset int) ([]*Provenance, int, error) {
	var result []*Provenance
	for _, p := range m.store {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockProvenanceRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Provenance, int, error) {
	var result []*Provenance
	for _, p := range m.store {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockProvenanceRepo) AddAgent(_ context.Context, a *ProvenanceAgent) error {
	a.ID = uuid.New()
	m.agents[a.ProvenanceID] = append(m.agents[a.ProvenanceID], a)
	return nil
}

func (m *mockProvenanceRepo) GetAgents(_ context.Context, provenanceID uuid.UUID) ([]*ProvenanceAgent, error) {
	return m.agents[provenanceID], nil
}

func (m *mockProvenanceRepo) AddEntity(_ context.Context, e *ProvenanceEntity) error {
	e.ID = uuid.New()
	m.entities[e.ProvenanceID] = append(m.entities[e.ProvenanceID], e)
	return nil
}

func (m *mockProvenanceRepo) GetEntities(_ context.Context, provenanceID uuid.UUID) ([]*ProvenanceEntity, error) {
	return m.entities[provenanceID], nil
}

// =========== Helper ===========

func newTestService() *Service {
	return NewService(newMockProvenanceRepo())
}

// =========== Provenance Tests ===========

func TestCreateProvenance_Success(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	if err := svc.CreateProvenance(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateProvenance_MissingTargetType(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetID: "pat-123", Recorded: time.Now()}
	if err := svc.CreateProvenance(context.Background(), p); err == nil {
		t.Fatal("expected error for missing target_type")
	}
}

func TestCreateProvenance_MissingTargetID(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", Recorded: time.Now()}
	if err := svc.CreateProvenance(context.Background(), p); err == nil {
		t.Fatal("expected error for missing target_id")
	}
}

func TestGetProvenance(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	svc.CreateProvenance(context.Background(), p)

	got, err := svc.GetProvenance(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected ID %v, got %v", p.ID, got.ID)
	}
}

func TestGetProvenance_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetProvenance(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetProvenanceByFHIRID(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	svc.CreateProvenance(context.Background(), p)

	got, err := svc.GetProvenanceByFHIRID(context.Background(), p.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected ID %v, got %v", p.ID, got.ID)
	}
}

func TestGetProvenanceByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetProvenanceByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDeleteProvenance(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	svc.CreateProvenance(context.Background(), p)
	if err := svc.DeleteProvenance(context.Background(), p.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetProvenance(context.Background(), p.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestSearchProvenances(t *testing.T) {
	svc := newTestService()
	svc.CreateProvenance(context.Background(), &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()})
	items, total, err := svc.SearchProvenances(context.Background(), map[string]string{"target": "Patient/pat-123"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

// =========== Agent Tests ===========

func TestAddAgent_Success(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	svc.CreateProvenance(context.Background(), p)
	a := &ProvenanceAgent{ProvenanceID: p.ID, WhoType: "Practitioner", WhoID: "pract-456"}
	if err := svc.AddAgent(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	agents, _ := svc.GetAgents(context.Background(), p.ID)
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
}

func TestAddAgent_MissingProvenanceID(t *testing.T) {
	svc := newTestService()
	a := &ProvenanceAgent{WhoType: "Practitioner", WhoID: "pract-456"}
	if err := svc.AddAgent(context.Background(), a); err == nil {
		t.Fatal("expected error for missing provenance_id")
	}
}

func TestAddAgent_MissingWho(t *testing.T) {
	svc := newTestService()
	a := &ProvenanceAgent{ProvenanceID: uuid.New()}
	if err := svc.AddAgent(context.Background(), a); err == nil {
		t.Fatal("expected error for missing who_type/who_id")
	}
}

// =========== Entity Tests ===========

func TestAddEntity_Success(t *testing.T) {
	svc := newTestService()
	p := &Provenance{TargetType: "Patient", TargetID: "pat-123", Recorded: time.Now()}
	svc.CreateProvenance(context.Background(), p)
	e := &ProvenanceEntity{ProvenanceID: p.ID, Role: "source", WhatType: "DocumentReference", WhatID: "doc-789"}
	if err := svc.AddEntity(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entities, _ := svc.GetEntities(context.Background(), p.ID)
	if len(entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(entities))
	}
}

func TestAddEntity_MissingProvenanceID(t *testing.T) {
	svc := newTestService()
	e := &ProvenanceEntity{Role: "source", WhatType: "DocumentReference", WhatID: "doc-789"}
	if err := svc.AddEntity(context.Background(), e); err == nil {
		t.Fatal("expected error for missing provenance_id")
	}
}

func TestAddEntity_MissingRole(t *testing.T) {
	svc := newTestService()
	e := &ProvenanceEntity{ProvenanceID: uuid.New(), WhatType: "DocumentReference", WhatID: "doc-789"}
	if err := svc.AddEntity(context.Background(), e); err == nil {
		t.Fatal("expected error for missing role")
	}
}
