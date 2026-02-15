package fhir

import (
	"context"
	"encoding/json"
	"fmt"
)

// VersionTracker wraps HistoryRepository with higher-level methods
// that domain services call during create/update/delete operations.
type VersionTracker struct {
	repo *HistoryRepository
}

// NewVersionTracker creates a new VersionTracker wrapping the given HistoryRepository.
func NewVersionTracker(repo *HistoryRepository) *VersionTracker {
	return &VersionTracker{repo: repo}
}

// RecordCreate saves version 1 of a resource after creation.
func (vt *VersionTracker) RecordCreate(ctx context.Context, resourceType, resourceID string, resource interface{}) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("version tracker: marshal resource: %w", err)
	}
	return vt.repo.SaveVersion(ctx, resourceType, resourceID, 1, json.RawMessage(data), "create")
}

// RecordUpdate increments the version and saves a snapshot.
// Returns the new version number.
func (vt *VersionTracker) RecordUpdate(ctx context.Context, resourceType, resourceID string, currentVersion int, resource interface{}) (int, error) {
	newVersion := currentVersion + 1
	data, err := json.Marshal(resource)
	if err != nil {
		return 0, fmt.Errorf("version tracker: marshal resource: %w", err)
	}
	if err := vt.repo.SaveVersion(ctx, resourceType, resourceID, newVersion, json.RawMessage(data), "update"); err != nil {
		return 0, err
	}
	return newVersion, nil
}

// RecordDelete saves a deletion marker at the next version.
func (vt *VersionTracker) RecordDelete(ctx context.Context, resourceType, resourceID string, currentVersion int) error {
	return vt.repo.SaveVersion(ctx, resourceType, resourceID, currentVersion+1, json.RawMessage("null"), "delete")
}

// GetVersion retrieves a specific version of a resource from history.
func (vt *VersionTracker) GetVersion(ctx context.Context, resourceType, resourceID string, versionID int) (*HistoryEntry, error) {
	return vt.repo.GetVersion(ctx, resourceType, resourceID, versionID)
}

// ListVersions retrieves all versions of a resource.
func (vt *VersionTracker) ListVersions(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*HistoryEntry, int, error) {
	return vt.repo.ListVersions(ctx, resourceType, resourceID, limit, offset)
}
