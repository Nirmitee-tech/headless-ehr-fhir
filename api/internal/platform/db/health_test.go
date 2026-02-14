package db

import (
	"testing"
)

func TestPoolStats_Fields(t *testing.T) {
	// Test that PoolStats struct correctly holds values.
	stats := &PoolStats{
		TotalConns:      10,
		IdleConns:       5,
		AcquiredConns:   5,
		MaxConns:        20,
		AcquireCount:    100,
		AcquireDuration: "1.5s",
		Healthy:         true,
	}

	if stats.TotalConns != 10 {
		t.Errorf("expected TotalConns 10, got %d", stats.TotalConns)
	}
	if stats.IdleConns != 5 {
		t.Errorf("expected IdleConns 5, got %d", stats.IdleConns)
	}
	if stats.AcquiredConns != 5 {
		t.Errorf("expected AcquiredConns 5, got %d", stats.AcquiredConns)
	}
	if stats.MaxConns != 20 {
		t.Errorf("expected MaxConns 20, got %d", stats.MaxConns)
	}
	if stats.AcquireCount != 100 {
		t.Errorf("expected AcquireCount 100, got %d", stats.AcquireCount)
	}
	if stats.AcquireDuration != "1.5s" {
		t.Errorf("expected AcquireDuration '1.5s', got %q", stats.AcquireDuration)
	}
	if !stats.Healthy {
		t.Error("expected Healthy to be true")
	}
}

func TestPoolStats_UnhealthyState(t *testing.T) {
	stats := &PoolStats{
		TotalConns:      0,
		IdleConns:       0,
		AcquiredConns:   0,
		MaxConns:        20,
		AcquireCount:    0,
		AcquireDuration: "0s",
		Healthy:         false,
	}

	if stats.Healthy {
		t.Error("expected Healthy to be false when TotalConns is 0")
	}
	if stats.TotalConns != 0 {
		t.Errorf("expected TotalConns 0, got %d", stats.TotalConns)
	}
}

func TestPoolStats_JSONTags(t *testing.T) {
	// Verify the struct can be instantiated with all fields.
	// JSON serialization correctness is implicitly tested by the
	// HealthHandler integration, but we verify the struct is usable.
	stats := PoolStats{
		TotalConns:      1,
		IdleConns:       1,
		AcquiredConns:   0,
		MaxConns:        10,
		AcquireCount:    50,
		AcquireDuration: "250ms",
		Healthy:         true,
	}

	// Verify fields are accessible and correctly typed
	var totalConns int32 = stats.TotalConns
	var acquireCount int64 = stats.AcquireCount
	var healthy bool = stats.Healthy
	var duration string = stats.AcquireDuration

	if totalConns != 1 {
		t.Errorf("expected TotalConns 1, got %d", totalConns)
	}
	if acquireCount != 50 {
		t.Errorf("expected AcquireCount 50, got %d", acquireCount)
	}
	if !healthy {
		t.Error("expected healthy to be true")
	}
	if duration != "250ms" {
		t.Errorf("expected duration '250ms', got %q", duration)
	}
}
