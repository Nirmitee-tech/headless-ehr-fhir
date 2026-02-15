package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ResourceEvent describes a resource mutation that has been recorded.
type ResourceEvent struct {
	ResourceType string
	ResourceID   string
	VersionID    int
	Action       string // "create", "update", "delete"
	Resource     json.RawMessage
}

// ResourceEventListener is notified whenever a resource is created, updated, or deleted.
type ResourceEventListener interface {
	OnResourceEvent(ctx context.Context, event ResourceEvent)
}

// VersionTracker wraps HistoryRepository with higher-level methods
// that domain services call during create/update/delete operations.
type VersionTracker struct {
	repo      *HistoryRepository
	mu        sync.RWMutex
	listeners []ResourceEventListener
}

// NewVersionTracker creates a new VersionTracker wrapping the given HistoryRepository.
func NewVersionTracker(repo *HistoryRepository) *VersionTracker {
	return &VersionTracker{repo: repo}
}

// AddListener registers a listener that will be notified on resource events.
func (vt *VersionTracker) AddListener(l ResourceEventListener) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	vt.listeners = append(vt.listeners, l)
}

func (vt *VersionTracker) fireEvent(ctx context.Context, event ResourceEvent) {
	vt.mu.RLock()
	listeners := vt.listeners
	vt.mu.RUnlock()
	for _, l := range listeners {
		l.OnResourceEvent(ctx, event)
	}
}

// RecordCreate saves version 1 of a resource after creation.
func (vt *VersionTracker) RecordCreate(ctx context.Context, resourceType, resourceID string, resource interface{}) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("version tracker: marshal resource: %w", err)
	}
	if err := vt.repo.SaveVersion(ctx, resourceType, resourceID, 1, json.RawMessage(data), "create"); err != nil {
		return err
	}
	vt.fireEvent(ctx, ResourceEvent{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		VersionID:    1,
		Action:       "create",
		Resource:     data,
	})
	return nil
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
	vt.fireEvent(ctx, ResourceEvent{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		VersionID:    newVersion,
		Action:       "update",
		Resource:     data,
	})
	return newVersion, nil
}

// RecordDelete saves a deletion marker at the next version.
func (vt *VersionTracker) RecordDelete(ctx context.Context, resourceType, resourceID string, currentVersion int) error {
	if err := vt.repo.SaveVersion(ctx, resourceType, resourceID, currentVersion+1, json.RawMessage("null"), "delete"); err != nil {
		return err
	}
	vt.fireEvent(ctx, ResourceEvent{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		VersionID:    currentVersion + 1,
		Action:       "delete",
		Resource:     json.RawMessage("null"),
	})
	return nil
}

// GetVersion retrieves a specific version of a resource from history.
func (vt *VersionTracker) GetVersion(ctx context.Context, resourceType, resourceID string, versionID int) (*HistoryEntry, error) {
	return vt.repo.GetVersion(ctx, resourceType, resourceID, versionID)
}

// ListVersions retrieves all versions of a resource.
func (vt *VersionTracker) ListVersions(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*HistoryEntry, int, error) {
	return vt.repo.ListVersions(ctx, resourceType, resourceID, limit, offset)
}
