package hipaa

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
)

// RetentionHandler provides Echo HTTP handlers for data retention policy management.
type RetentionHandler struct {
	service *RetentionService
}

// NewRetentionHandler creates a new handler backed by the given retention service.
func NewRetentionHandler(service *RetentionService) *RetentionHandler {
	return &RetentionHandler{service: service}
}

// RegisterRetentionRoutes registers admin-only retention policy routes on the API group.
func RegisterRetentionRoutes(g *echo.Group, service *RetentionService) {
	h := NewRetentionHandler(service)

	admin := g.Group("/admin/retention-policies", auth.RequireRole("admin"))
	admin.GET("", h.HandleListPolicies)
	admin.GET("/:resourceType", h.HandleGetPolicy)

	g.GET("/admin/retention-status", h.HandleRetentionStatus, auth.RequireRole("admin"))
}

// HandleListPolicies handles GET /api/v1/admin/retention-policies.
// Returns all configured retention policies.
func (h *RetentionHandler) HandleListPolicies(c echo.Context) error {
	policies := h.service.GetAllPolicies()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"policies": policies,
		"total":    len(policies),
	})
}

// HandleGetPolicy handles GET /api/v1/admin/retention-policies/:resourceType.
// Returns the retention policy for a specific resource type.
func (h *RetentionHandler) HandleGetPolicy(c echo.Context) error {
	resourceType := c.Param("resourceType")
	policy := h.service.GetPolicy(resourceType)
	if policy == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "no retention policy found for resource type: " + resourceType,
		})
	}
	return c.JSON(http.StatusOK, policy)
}

// RetentionStatusSummary provides a summary of data grouped by retention state.
type RetentionStatusSummary struct {
	ResourceType     string `json:"resource_type"`
	ActiveCount      int    `json:"active_count"`
	ArchivableCount  int    `json:"archivable_count"`
	PurgeableCount   int    `json:"purgeable_count"`
	RetentionDays    int    `json:"retention_days"`
	ArchiveAfterDays int    `json:"archive_after_days"`
	PurgeAfterDays   int    `json:"purge_after_days"`
}

// HandleRetentionStatus handles GET /api/v1/admin/retention-status.
// Returns a summary of retention policies with counts (simulated for in-memory).
// In a production system, this would query the database for actual record counts.
func (h *RetentionHandler) HandleRetentionStatus(c echo.Context) error {
	policies := h.service.GetAllPolicies()

	summaries := make([]RetentionStatusSummary, 0, len(policies))
	for _, p := range policies {
		summary := RetentionStatusSummary{
			ResourceType:     p.ResourceType,
			ActiveCount:      0, // would be populated from DB in production
			ArchivableCount:  0,
			PurgeableCount:   0,
			RetentionDays:    p.RetentionDays,
			ArchiveAfterDays: p.ArchiveAfter,
			PurgeAfterDays:   p.PurgeAfter,
		}
		summaries = append(summaries, summary)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"summaries":  summaries,
		"as_of":      time.Now().UTC(),
		"total_types": len(summaries),
	})
}
