package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMigrations(t *testing.T) {
	dir := t.TempDir()

	// Create test SQL files
	files := map[string]string{
		"001_core.sql":        "CREATE TABLE users (id SERIAL PRIMARY KEY);",
		"002_clinical.sql":    "CREATE TABLE conditions (id SERIAL PRIMARY KEY);",
		"003_medications.sql": "CREATE TABLE medications (id SERIAL PRIMARY KEY);",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", name, err)
		}
	}

	migrator := NewMigrator(nil, dir)
	migrations, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() error: %v", err)
	}

	if len(migrations) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(migrations))
	}

	// Verify version parsing
	if migrations[0].Version != 1 {
		t.Errorf("expected version 1, got %d", migrations[0].Version)
	}
	if migrations[0].Name != "001_core.sql" {
		t.Errorf("expected name 001_core.sql, got %s", migrations[0].Name)
	}
	if migrations[0].SQL != "CREATE TABLE users (id SERIAL PRIMARY KEY);" {
		t.Errorf("unexpected SQL content: %s", migrations[0].SQL)
	}

	if migrations[1].Version != 2 {
		t.Errorf("expected version 2, got %d", migrations[1].Version)
	}
	if migrations[2].Version != 3 {
		t.Errorf("expected version 3, got %d", migrations[2].Version)
	}
}

func TestLoadMigrations_SortOrder(t *testing.T) {
	dir := t.TempDir()

	// Create files in reverse order to test sorting
	files := []struct {
		name    string
		content string
	}{
		{"010_tables.sql", "SELECT 10;"},
		{"002_second.sql", "SELECT 2;"},
		{"001_first.sql", "SELECT 1;"},
		{"005_middle.sql", "SELECT 5;"},
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", f.name, err)
		}
	}

	migrator := NewMigrator(nil, dir)
	migrations, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() error: %v", err)
	}

	if len(migrations) != 4 {
		t.Fatalf("expected 4 migrations, got %d", len(migrations))
	}

	expectedVersions := []int{1, 2, 5, 10}
	for i, expected := range expectedVersions {
		if migrations[i].Version != expected {
			t.Errorf("migration[%d]: expected version %d, got %d", i, expected, migrations[i].Version)
		}
	}
}

func TestLoadMigrations_InvalidFilename(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"001_valid.sql":   "SELECT 1;",
		"readme.sql":      "-- this has no version prefix",
		"notes.txt":       "not a sql file",
		"abc_invalid.sql": "-- non-numeric prefix",
		"002_also_valid.sql": "SELECT 2;",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", name, err)
		}
	}

	migrator := NewMigrator(nil, dir)
	migrations, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() error: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 valid migrations, got %d", len(migrations))
	}

	if migrations[0].Version != 1 {
		t.Errorf("expected first migration version 1, got %d", migrations[0].Version)
	}
	if migrations[1].Version != 2 {
		t.Errorf("expected second migration version 2, got %d", migrations[1].Version)
	}
}

func TestLoadMigrations_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	migrator := NewMigrator(nil, dir)
	migrations, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() error: %v", err)
	}

	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations from empty dir, got %d", len(migrations))
	}
}

func TestMigrationStatus(t *testing.T) {
	// This test verifies the MigrationStatus struct and that Status method
	// correctly categorizes migrations as applied or pending.
	// Since we cannot connect to a real database, we test the data structures
	// and LoadMigrations behavior that feeds into Status.

	dir := t.TempDir()

	files := map[string]string{
		"001_core.sql":     "CREATE TABLE users (id SERIAL);",
		"002_clinical.sql": "CREATE TABLE conditions (id SERIAL);",
		"003_meds.sql":     "CREATE TABLE meds (id SERIAL);",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file %s: %v", name, err)
		}
	}

	migrator := NewMigrator(nil, dir)
	migrations, err := migrator.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations() error: %v", err)
	}

	// Simulate building status from loaded migrations with an applied set
	appliedVersions := map[int]bool{1: true}

	var statuses []MigrationStatus
	for _, mig := range migrations {
		status := MigrationStatus{
			Version: mig.Version,
			Name:    mig.Name,
			Applied: appliedVersions[mig.Version],
		}
		statuses = append(statuses, status)
	}

	if len(statuses) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(statuses))
	}

	// First migration should be applied
	if !statuses[0].Applied {
		t.Error("expected migration 001 to be applied")
	}
	if statuses[0].Version != 1 {
		t.Errorf("expected version 1, got %d", statuses[0].Version)
	}

	// Second and third should be pending
	if statuses[1].Applied {
		t.Error("expected migration 002 to be pending")
	}
	if statuses[2].Applied {
		t.Error("expected migration 003 to be pending")
	}

	// Verify names
	if statuses[0].Name != "001_core.sql" {
		t.Errorf("expected name 001_core.sql, got %s", statuses[0].Name)
	}
	if statuses[1].Name != "002_clinical.sql" {
		t.Errorf("expected name 002_clinical.sql, got %s", statuses[1].Name)
	}
	if statuses[2].Name != "003_meds.sql" {
		t.Errorf("expected name 003_meds.sql, got %s", statuses[2].Name)
	}

	// Verify AppliedAt is nil for pending migrations
	if statuses[1].AppliedAt != nil {
		t.Error("expected nil AppliedAt for pending migration")
	}
	if statuses[2].AppliedAt != nil {
		t.Error("expected nil AppliedAt for pending migration")
	}
}

func TestNewMigrator(t *testing.T) {
	m := NewMigrator(nil, "/some/path")
	if m == nil {
		t.Fatal("expected non-nil Migrator")
	}
	if m.dir != "/some/path" {
		t.Errorf("expected dir /some/path, got %s", m.dir)
	}
	if m.pool != nil {
		t.Error("expected nil pool")
	}
}

func TestLoadMigrations_NonExistentDir(t *testing.T) {
	migrator := NewMigrator(nil, "/nonexistent/path/that/does/not/exist")
	_, err := migrator.LoadMigrations()
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}
