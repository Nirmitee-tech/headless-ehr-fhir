package auth

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Shared test-suite that can run against ANY LaunchContextStorer implementation
// ---------------------------------------------------------------------------

func runStoreTests(t *testing.T, name string, newStore func(ttl time.Duration) LaunchContextStorer) {
	t.Run(name+"/SaveAndGet", func(t *testing.T) {
		store := newStore(5 * time.Minute)
		ctx := context.Background()

		lc := &LaunchContext{
			PatientID:   "patient-123",
			EncounterID: "encounter-456",
			FHIRUser:    "Practitioner/dr-smith",
			CreatedAt:   time.Now(),
		}

		if err := store.Save(ctx, "tok-1", lc); err != nil {
			t.Fatalf("Save: unexpected error: %v", err)
		}

		got, err := store.Get(ctx, "tok-1")
		if err != nil {
			t.Fatalf("Get: unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("Get: expected non-nil context")
		}
		if got.PatientID != "patient-123" {
			t.Errorf("PatientID = %q, want %q", got.PatientID, "patient-123")
		}
		if got.EncounterID != "encounter-456" {
			t.Errorf("EncounterID = %q, want %q", got.EncounterID, "encounter-456")
		}
		if got.FHIRUser != "Practitioner/dr-smith" {
			t.Errorf("FHIRUser = %q, want %q", got.FHIRUser, "Practitioner/dr-smith")
		}
	})

	t.Run(name+"/GetNonExistent", func(t *testing.T) {
		store := newStore(5 * time.Minute)
		ctx := context.Background()

		got, err := store.Get(ctx, "does-not-exist")
		if err != nil {
			t.Fatalf("Get: unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("Get: expected nil for non-existent key, got %+v", got)
		}
	})

	t.Run(name+"/Consume", func(t *testing.T) {
		store := newStore(5 * time.Minute)
		ctx := context.Background()

		lc := &LaunchContext{
			PatientID: "patient-789",
			CreatedAt: time.Now(),
		}
		if err := store.Save(ctx, "tok-consume", lc); err != nil {
			t.Fatalf("Save: %v", err)
		}

		consumed, err := store.Consume(ctx, "tok-consume")
		if err != nil {
			t.Fatalf("Consume: %v", err)
		}
		if consumed == nil {
			t.Fatal("Consume: expected non-nil")
		}
		if consumed.PatientID != "patient-789" {
			t.Errorf("PatientID = %q, want %q", consumed.PatientID, "patient-789")
		}
	})

	t.Run(name+"/ConsumeDeletesContext", func(t *testing.T) {
		store := newStore(5 * time.Minute)
		ctx := context.Background()

		lc := &LaunchContext{PatientID: "pat-del", CreatedAt: time.Now()}
		store.Save(ctx, "tok-del", lc)

		// First consume succeeds
		first, _ := store.Consume(ctx, "tok-del")
		if first == nil {
			t.Fatal("first Consume: expected non-nil")
		}

		// Second consume returns nil
		second, err := store.Consume(ctx, "tok-del")
		if err != nil {
			t.Fatalf("second Consume: unexpected error: %v", err)
		}
		if second != nil {
			t.Error("second Consume: expected nil")
		}

		// Get also returns nil
		got, _ := store.Get(ctx, "tok-del")
		if got != nil {
			t.Error("Get after Consume: expected nil")
		}
	})

	t.Run(name+"/ConsumeNonExistent", func(t *testing.T) {
		store := newStore(5 * time.Minute)
		ctx := context.Background()

		consumed, err := store.Consume(ctx, "no-such-token")
		if err != nil {
			t.Fatalf("Consume non-existent: unexpected error: %v", err)
		}
		if consumed != nil {
			t.Error("expected nil for non-existent token")
		}
	})

	t.Run(name+"/TTLExpiry", func(t *testing.T) {
		store := newStore(50 * time.Millisecond)
		ctx := context.Background()

		lc := &LaunchContext{PatientID: "pat-ttl", CreatedAt: time.Now()}
		store.Save(ctx, "tok-ttl", lc)

		// Should be available immediately
		got, _ := store.Get(ctx, "tok-ttl")
		if got == nil {
			t.Fatal("Get immediately after Save: expected non-nil")
		}

		// Wait for TTL expiry
		time.Sleep(100 * time.Millisecond)

		got, err := store.Get(ctx, "tok-ttl")
		if err != nil {
			t.Fatalf("Get after expiry: unexpected error: %v", err)
		}
		if got != nil {
			t.Error("Get after expiry: expected nil")
		}
	})

	t.Run(name+"/ConsumeExpired", func(t *testing.T) {
		store := newStore(50 * time.Millisecond)
		ctx := context.Background()

		lc := &LaunchContext{PatientID: "pat-exp", CreatedAt: time.Now()}
		store.Save(ctx, "tok-exp", lc)

		time.Sleep(100 * time.Millisecond)

		consumed, err := store.Consume(ctx, "tok-exp")
		if err != nil {
			t.Fatalf("Consume expired: unexpected error: %v", err)
		}
		if consumed != nil {
			t.Error("Consume expired: expected nil")
		}
	})

	t.Run(name+"/SaveOverwrites", func(t *testing.T) {
		store := newStore(5 * time.Minute)
		ctx := context.Background()

		lc1 := &LaunchContext{PatientID: "first-patient", CreatedAt: time.Now()}
		store.Save(ctx, "tok-overwrite", lc1)

		lc2 := &LaunchContext{PatientID: "second-patient", CreatedAt: time.Now()}
		store.Save(ctx, "tok-overwrite", lc2)

		got, err := store.Get(ctx, "tok-overwrite")
		if err != nil {
			t.Fatalf("Get after overwrite: %v", err)
		}
		if got == nil {
			t.Fatal("Get after overwrite: expected non-nil")
		}
		if got.PatientID != "second-patient" {
			t.Errorf("PatientID = %q, want %q (overwrite)", got.PatientID, "second-patient")
		}
	})
}

// ---------------------------------------------------------------------------
// InMemoryLaunchContextStore tests
// ---------------------------------------------------------------------------

func TestInMemoryLaunchContextStore(t *testing.T) {
	runStoreTests(t, "InMemory", func(ttl time.Duration) LaunchContextStorer {
		return NewInMemoryLaunchContextStore(ttl)
	})
}

func TestInMemoryLaunchContextStore_Cleanup(t *testing.T) {
	store := NewInMemoryLaunchContextStore(50 * time.Millisecond)
	ctx := context.Background()

	store.Save(ctx, "tok-1", &LaunchContext{PatientID: "p1", CreatedAt: time.Now()})
	store.Save(ctx, "tok-2", &LaunchContext{PatientID: "p2", CreatedAt: time.Now()})

	time.Sleep(100 * time.Millisecond)
	store.Cleanup()

	store.mu.RLock()
	count := len(store.contexts)
	store.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 contexts after cleanup, got %d", count)
	}
}

func TestInMemoryLaunchContextStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryLaunchContextStore(5 * time.Minute)
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent saves
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			lc := &LaunchContext{PatientID: "patient", CreatedAt: time.Now()}
			if err := store.Save(ctx, "concurrent-tok", lc); err != nil {
				t.Errorf("concurrent save %d: %v", idx, err)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Get(ctx, "nonexistent")
		}()
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// PGLaunchContextStore tests (unit tests with a mock DB layer)
// ---------------------------------------------------------------------------

// mockPGRow implements the pgRow interface for testing.
type mockPGRow struct {
	data    []byte
	scanErr error
	noRows  bool
}

func (r *mockPGRow) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	if r.noRows {
		return errors.New("no rows in result set")
	}
	if len(dest) > 0 {
		if b, ok := dest[0].(*[]byte); ok {
			*b = r.data
		}
	}
	return nil
}

// mockPGConn implements the pgConn interface for testing.
type mockPGConn struct {
	mu       sync.Mutex
	store    map[string]mockEntry
	queryErr error
	execErr  error
}

type mockEntry struct {
	data      []byte
	expiresAt time.Time
}

func newMockPGConn() *mockPGConn {
	return &mockPGConn{store: make(map[string]mockEntry)}
}

func (m *mockPGConn) QueryRow(ctx context.Context, sql string, args ...any) pgRow {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.queryErr != nil {
		return &mockPGRow{scanErr: m.queryErr}
	}

	if len(args) == 0 {
		return &mockPGRow{noRows: true}
	}

	id, ok := args[0].(string)
	if !ok {
		return &mockPGRow{noRows: true}
	}

	entry, exists := m.store[id]
	if !exists {
		return &mockPGRow{noRows: true}
	}

	// Check expiry
	if time.Now().After(entry.expiresAt) {
		delete(m.store, id)
		return &mockPGRow{noRows: true}
	}

	// If this is a consume (DELETE RETURNING) query, also remove the entry
	if len(sql) > 6 && sql[:6] == "DELETE" {
		delete(m.store, id)
	}

	return &mockPGRow{data: entry.data}
}

func (m *mockPGConn) Exec(ctx context.Context, sql string, args ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.execErr != nil {
		return m.execErr
	}

	// Handle INSERT (Save)
	if len(sql) > 6 && sql[:6] == "INSERT" {
		if len(args) >= 4 {
			id, _ := args[0].(string)
			data, _ := args[1].([]byte)
			expiresAt, _ := args[3].(time.Time)
			m.store[id] = mockEntry{data: data, expiresAt: expiresAt}
		}
		return nil
	}

	// Handle DELETE (cleanup)
	if len(sql) > 6 && sql[:6] == "DELETE" {
		now := time.Now()
		for k, v := range m.store {
			if now.After(v.expiresAt) {
				delete(m.store, k)
			}
		}
		return nil
	}

	return nil
}

func TestPGLaunchContextStore(t *testing.T) {
	runStoreTests(t, "PG", func(ttl time.Duration) LaunchContextStorer {
		mock := newMockPGConn()
		return NewPGLaunchContextStore(mock, ttl)
	})
}

func TestPGLaunchContextStore_Cleanup(t *testing.T) {
	mock := newMockPGConn()
	store := NewPGLaunchContextStore(mock, 50*time.Millisecond)
	ctx := context.Background()

	store.Save(ctx, "tok-1", &LaunchContext{PatientID: "p1", CreatedAt: time.Now()})
	store.Save(ctx, "tok-2", &LaunchContext{PatientID: "p2", CreatedAt: time.Now()})

	time.Sleep(100 * time.Millisecond)

	if err := store.Cleanup(ctx); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	// Verify entries are gone
	got, _ := store.Get(ctx, "tok-1")
	if got != nil {
		t.Error("expected nil after cleanup for tok-1")
	}
	got, _ = store.Get(ctx, "tok-2")
	if got != nil {
		t.Error("expected nil after cleanup for tok-2")
	}
}

func TestPGLaunchContextStore_SaveError(t *testing.T) {
	mock := newMockPGConn()
	mock.execErr = errors.New("db write failed")
	store := NewPGLaunchContextStore(mock, 5*time.Minute)

	err := store.Save(context.Background(), "tok-err", &LaunchContext{
		PatientID: "p1",
		CreatedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error from Save when DB fails")
	}
}

func TestPGLaunchContextStore_GetError(t *testing.T) {
	mock := newMockPGConn()
	mock.queryErr = errors.New("db read failed")
	store := NewPGLaunchContextStore(mock, 5*time.Minute)

	_, err := store.Get(context.Background(), "tok-err")
	if err == nil {
		t.Fatal("expected error from Get when DB fails")
	}
}

func TestPGLaunchContextStore_ConsumeError(t *testing.T) {
	mock := newMockPGConn()
	mock.queryErr = errors.New("db read failed")
	store := NewPGLaunchContextStore(mock, 5*time.Minute)

	_, err := store.Consume(context.Background(), "tok-err")
	if err == nil {
		t.Fatal("expected error from Consume when DB fails")
	}
}

func TestPGLaunchContextStore_JSONRoundTrip(t *testing.T) {
	mock := newMockPGConn()
	store := NewPGLaunchContextStore(mock, 5*time.Minute)
	ctx := context.Background()

	original := &LaunchContext{
		LaunchToken: "tok-json",
		PatientID:   "patient-abc",
		EncounterID: "encounter-xyz",
		FHIRUser:    "Practitioner/dr-house",
		CreatedAt:   time.Now(),
	}

	if err := store.Save(ctx, "tok-json", original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "tok-json")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("Get: expected non-nil")
	}

	// Verify all fields survived JSON round-trip
	if got.LaunchToken != original.LaunchToken {
		t.Errorf("LaunchToken = %q, want %q", got.LaunchToken, original.LaunchToken)
	}
	if got.PatientID != original.PatientID {
		t.Errorf("PatientID = %q, want %q", got.PatientID, original.PatientID)
	}
	if got.EncounterID != original.EncounterID {
		t.Errorf("EncounterID = %q, want %q", got.EncounterID, original.EncounterID)
	}
	if got.FHIRUser != original.FHIRUser {
		t.Errorf("FHIRUser = %q, want %q", got.FHIRUser, original.FHIRUser)
	}
}

func TestPGMigrationSQL(t *testing.T) {
	// Verify the migration SQL constant is non-empty and contains expected table name
	if MigrationSmartLaunchContexts == "" {
		t.Fatal("MigrationSmartLaunchContexts should not be empty")
	}
	if !contains(MigrationSmartLaunchContexts, "smart_launch_contexts") {
		t.Error("migration SQL should reference smart_launch_contexts table")
	}
	if !contains(MigrationSmartLaunchContexts, "CREATE TABLE") {
		t.Error("migration SQL should contain CREATE TABLE")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Legacy LaunchContextStore backward-compat tests
// ---------------------------------------------------------------------------

// TestLegacyLaunchContextStore_Create verifies the old Create method still works
// through the renamed InMemoryLaunchContextStore.
func TestLegacyLaunchContextStore_Create(t *testing.T) {
	store := NewLaunchContextStore(5 * time.Minute)

	lc, err := store.Create("patient-legacy", "enc-legacy", "Practitioner/legacy")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if lc.LaunchToken == "" {
		t.Fatal("expected non-empty launch token")
	}
	if lc.PatientID != "patient-legacy" {
		t.Errorf("PatientID = %q, want %q", lc.PatientID, "patient-legacy")
	}
}

// TestLegacyLaunchContextStore_JSONMarshal verifies the LaunchContext JSON
// serialization used by the PG store is consistent with the model.
func TestLaunchContext_JSONMarshalForStore(t *testing.T) {
	lc := &LaunchContext{
		LaunchToken: "tok-marshal",
		PatientID:   "p1",
		EncounterID: "e1",
		FHIRUser:    "Practitioner/doc",
		CreatedAt:   time.Now(),
	}

	// The store uses launchContextJSON which includes CreatedAt, unlike the
	// main LaunchContext JSON tags which have json:"-" on CreatedAt.
	data, err := json.Marshal(launchContextToJSON(lc))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var restored launchContextJSON
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if restored.PatientID != lc.PatientID {
		t.Errorf("PatientID = %q, want %q", restored.PatientID, lc.PatientID)
	}
	if restored.EncounterID != lc.EncounterID {
		t.Errorf("EncounterID = %q, want %q", restored.EncounterID, lc.EncounterID)
	}
	if restored.FHIRUser != lc.FHIRUser {
		t.Errorf("FHIRUser = %q, want %q", restored.FHIRUser, lc.FHIRUser)
	}
	if restored.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt after round-trip")
	}
}
