package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/db"
)

// HistoryEntry represents a single version of a resource stored in the resource_history table.
type HistoryEntry struct {
	ID           string          `json:"id"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	VersionID    int             `json:"version_id"`
	Resource     json.RawMessage `json:"resource"`
	Action       string          `json:"action"` // "create", "update", "delete"
	Timestamp    time.Time       `json:"timestamp"`
}

type historyQuerier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// HistoryRepository provides access to the shared resource_history table.
type HistoryRepository struct{}

// NewHistoryRepository creates a new HistoryRepository.
func NewHistoryRepository() *HistoryRepository {
	return &HistoryRepository{}
}

func (r *HistoryRepository) conn(ctx context.Context) historyQuerier {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return nil
}

// SaveVersion stores a snapshot of a resource version in the history table.
func (r *HistoryRepository) SaveVersion(ctx context.Context, resourceType, resourceID string, versionID int, resource interface{}, action string) error {
	q := r.conn(ctx)
	if q == nil {
		return fmt.Errorf("no database connection in context")
	}

	data, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("marshal resource for history: %w", err)
	}

	_, err = q.Exec(ctx, `
		INSERT INTO resource_history (resource_type, resource_id, version_id, resource, action, timestamp)
		VALUES ($1, $2, $3, $4, $5, NOW())`,
		resourceType, resourceID, versionID, data, action)
	if err != nil {
		return fmt.Errorf("save history version: %w", err)
	}
	return nil
}

// GetVersion retrieves a specific version of a resource.
func (r *HistoryRepository) GetVersion(ctx context.Context, resourceType, resourceID string, versionID int) (*HistoryEntry, error) {
	q := r.conn(ctx)
	if q == nil {
		return nil, fmt.Errorf("no database connection in context")
	}

	var h HistoryEntry
	err := q.QueryRow(ctx, `
		SELECT resource_type, resource_id, version_id, resource, action, timestamp
		FROM resource_history
		WHERE resource_type = $1 AND resource_id = $2 AND version_id = $3`,
		resourceType, resourceID, versionID).
		Scan(&h.ResourceType, &h.ResourceID, &h.VersionID, &h.Resource, &h.Action, &h.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("get history version: %w", err)
	}
	return &h, nil
}

// ListVersions retrieves all versions of a resource, ordered by version descending.
func (r *HistoryRepository) ListVersions(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*HistoryEntry, int, error) {
	q := r.conn(ctx)
	if q == nil {
		return nil, 0, fmt.Errorf("no database connection in context")
	}

	var total int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM resource_history
		WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count history versions: %w", err)
	}

	rows, err := q.Query(ctx, `
		SELECT resource_type, resource_id, version_id, resource, action, timestamp
		FROM resource_history
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY version_id DESC
		LIMIT $3 OFFSET $4`,
		resourceType, resourceID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list history versions: %w", err)
	}
	defer rows.Close()

	var entries []*HistoryEntry
	for rows.Next() {
		var h HistoryEntry
		if err := rows.Scan(&h.ResourceType, &h.ResourceID, &h.VersionID, &h.Resource, &h.Action, &h.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("scan history entry: %w", err)
		}
		entries = append(entries, &h)
	}
	return entries, total, nil
}

// ListTypeVersions retrieves all history entries for a given resource type,
// ordered by timestamp descending. It supports optional _since filtering and
// limit/offset pagination.
func (r *HistoryRepository) ListTypeVersions(ctx context.Context, resourceType string, since *time.Time, limit, offset int) ([]*HistoryEntry, int, error) {
	q := r.conn(ctx)
	if q == nil {
		return nil, 0, fmt.Errorf("no database connection in context")
	}

	var total int
	if since != nil {
		err := q.QueryRow(ctx, `
			SELECT COUNT(*) FROM resource_history
			WHERE resource_type = $1 AND timestamp >= $2`,
			resourceType, *since).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("count type history versions: %w", err)
		}
	} else {
		err := q.QueryRow(ctx, `
			SELECT COUNT(*) FROM resource_history
			WHERE resource_type = $1`,
			resourceType).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("count type history versions: %w", err)
		}
	}

	var rows pgx.Rows
	var err error
	if since != nil {
		rows, err = q.Query(ctx, `
			SELECT resource_type, resource_id, version_id, resource, action, timestamp
			FROM resource_history
			WHERE resource_type = $1 AND timestamp >= $2
			ORDER BY timestamp DESC
			LIMIT $3 OFFSET $4`,
			resourceType, *since, limit, offset)
	} else {
		rows, err = q.Query(ctx, `
			SELECT resource_type, resource_id, version_id, resource, action, timestamp
			FROM resource_history
			WHERE resource_type = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3`,
			resourceType, limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("list type history versions: %w", err)
	}
	defer rows.Close()

	var entries []*HistoryEntry
	for rows.Next() {
		var h HistoryEntry
		if err := rows.Scan(&h.ResourceType, &h.ResourceID, &h.VersionID, &h.Resource, &h.Action, &h.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("scan type history entry: %w", err)
		}
		entries = append(entries, &h)
	}
	return entries, total, nil
}

// ListAllVersions retrieves all history entries across all resource types,
// ordered by timestamp descending. It supports optional _since filtering and
// limit/offset pagination. This implements the system-level _history interaction.
func (r *HistoryRepository) ListAllVersions(ctx context.Context, since *time.Time, limit, offset int) ([]*HistoryEntry, int, error) {
	q := r.conn(ctx)
	if q == nil {
		return nil, 0, fmt.Errorf("no database connection in context")
	}

	var total int
	if since != nil {
		err := q.QueryRow(ctx, `
			SELECT COUNT(*) FROM resource_history
			WHERE timestamp >= $1`,
			*since).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("count all history versions: %w", err)
		}
	} else {
		err := q.QueryRow(ctx, `
			SELECT COUNT(*) FROM resource_history`).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("count all history versions: %w", err)
		}
	}

	var rows pgx.Rows
	var err error
	if since != nil {
		rows, err = q.Query(ctx, `
			SELECT resource_type, resource_id, version_id, resource, action, timestamp
			FROM resource_history
			WHERE timestamp >= $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3`,
			*since, limit, offset)
	} else {
		rows, err = q.Query(ctx, `
			SELECT resource_type, resource_id, version_id, resource, action, timestamp
			FROM resource_history
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2`,
			limit, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("list all history versions: %w", err)
	}
	defer rows.Close()

	var entries []*HistoryEntry
	for rows.Next() {
		var h HistoryEntry
		if err := rows.Scan(&h.ResourceType, &h.ResourceID, &h.VersionID, &h.Resource, &h.Action, &h.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("scan all history entry: %w", err)
		}
		entries = append(entries, &h)
	}
	return entries, total, nil
}

// HistoryHandler serves FHIR system-level and type-level _history endpoints.
type HistoryHandler struct {
	repo *HistoryRepository
}

// NewHistoryHandler creates a new HistoryHandler.
func NewHistoryHandler(repo *HistoryRepository) *HistoryHandler {
	return &HistoryHandler{repo: repo}
}

// RegisterRoutes registers the history routes on the given echo group.
func (h *HistoryHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/_history", h.SystemHistory)
	g.GET("/:resourceType/_history", h.TypeHistory)
}

// SystemHistory handles GET /fhir/_history.
// It returns a history bundle containing all resource changes across the system.
func (h *HistoryHandler) SystemHistory(c echo.Context) error {
	count := ParseCount(c, 20)
	offset := ParseOffset(c)
	since := parseSince(c)

	entries, total, err := h.repo.ListAllVersions(c.Request().Context(), since, count, offset)
	if err != nil {
		// Return an empty history bundle when the database is unavailable.
		bundle := NewHistoryBundle(nil, 0, "/fhir")
		return c.JSON(http.StatusOK, bundle)
	}

	bundle := NewHistoryBundle(entries, total, "/fhir")
	return c.JSON(http.StatusOK, bundle)
}

// TypeHistory handles GET /fhir/:resourceType/_history.
// It returns a history bundle containing all changes for the specified resource type.
func (h *HistoryHandler) TypeHistory(c echo.Context) error {
	resourceType := c.Param("resourceType")
	count := ParseCount(c, 20)
	offset := ParseOffset(c)
	since := parseSince(c)

	entries, total, err := h.repo.ListTypeVersions(c.Request().Context(), resourceType, since, count, offset)
	if err != nil {
		// Return an empty history bundle when the database is unavailable.
		bundle := NewHistoryBundle(nil, 0, "/fhir")
		return c.JSON(http.StatusOK, bundle)
	}

	bundle := NewHistoryBundle(entries, total, "/fhir")
	return c.JSON(http.StatusOK, bundle)
}

// parseSince parses the _since query parameter as an RFC3339 timestamp.
// Returns nil if the parameter is not present or cannot be parsed.
func parseSince(c echo.Context) *time.Time {
	sinceStr := c.QueryParam("_since")
	if sinceStr == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		return nil
	}
	return &t
}

// NewHistoryBundle creates a FHIR Bundle of type "history" from history entries.
func NewHistoryBundle(entries []*HistoryEntry, total int, baseURL string) *Bundle {
	now := time.Now().UTC()
	bundleEntries := make([]BundleEntry, len(entries))

	for i, entry := range entries {
		fullURL := fmt.Sprintf("%s/%s/%s/_history/%d", baseURL, entry.ResourceType, entry.ResourceID, entry.VersionID)

		method := "PUT"
		status := "200 OK"
		switch entry.Action {
		case "create":
			method = "POST"
			status = "201 Created"
		case "delete":
			method = "DELETE"
			status = "204 No Content"
		}

		bundleEntries[i] = BundleEntry{
			FullURL:  fullURL,
			Resource: entry.Resource,
			Request: &BundleRequest{
				Method: method,
				URL:    fmt.Sprintf("%s/%s", entry.ResourceType, entry.ResourceID),
			},
			Response: &BundleResponse{
				Status:       status,
				LastModified: &entry.Timestamp,
			},
		}
	}

	return &Bundle{
		ResourceType: "Bundle",
		Type:         "history",
		Total:        &total,
		Timestamp:    &now,
		Entry:        bundleEntries,
	}
}
