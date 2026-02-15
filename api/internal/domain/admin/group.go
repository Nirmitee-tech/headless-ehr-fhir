package admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GroupType enumerates the allowed FHIR Group types.
type GroupType string

const (
	GroupTypePerson       GroupType = "person"
	GroupTypeAnimal       GroupType = "animal"
	GroupTypePractitioner GroupType = "practitioner"
	GroupTypeDevice       GroupType = "device"
	GroupTypeMedication   GroupType = "medication"
	GroupTypeSubstance    GroupType = "substance"
)

// validGroupTypes is the set of allowed group types.
var validGroupTypes = map[GroupType]bool{
	GroupTypePerson:       true,
	GroupTypeAnimal:       true,
	GroupTypePractitioner: true,
	GroupTypeDevice:       true,
	GroupTypeMedication:   true,
	GroupTypeSubstance:    true,
}

// IsValidGroupType returns true if the given type string is a recognized group type.
func IsValidGroupType(t string) bool {
	return validGroupTypes[GroupType(t)]
}

// Group represents a FHIR Group resource in the domain model.
type Group struct {
	ID             uuid.UUID     `json:"id"`
	FHIRID         string        `json:"fhir_id"`
	Type           GroupType     `json:"type"`
	Actual         bool          `json:"actual"`
	Code           *string       `json:"code,omitempty"`
	Name           string        `json:"name"`
	Quantity       int           `json:"quantity"`
	ManagingEntity *string       `json:"managing_entity,omitempty"`
	Members        []GroupMember `json:"members,omitempty"`
	Active         bool          `json:"active"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// GroupMember represents a member entry within a Group.
type GroupMember struct {
	ID          uuid.UUID  `json:"id"`
	EntityID    string     `json:"entity_id"`
	EntityType  string     `json:"entity_type"`
	PeriodStart *time.Time `json:"period_start,omitempty"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`
	Inactive    bool       `json:"inactive"`
}

// ToFHIR converts the Group domain model to a FHIR R4 Group resource map.
func (g *Group) ToFHIR() map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Group",
		"id":           g.FHIRID,
		"type":         string(g.Type),
		"actual":       g.Actual,
		"active":       g.Active,
		"name":         g.Name,
		"quantity":     g.Quantity,
		"meta":         fhir.Meta{LastUpdated: g.UpdatedAt},
	}

	if g.Code != nil {
		result["code"] = fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: *g.Code}},
		}
	}

	if g.ManagingEntity != nil {
		result["managingEntity"] = fhir.Reference{
			Reference: *g.ManagingEntity,
		}
	}

	if len(g.Members) > 0 {
		members := make([]map[string]interface{}, 0, len(g.Members))
		for _, m := range g.Members {
			entry := map[string]interface{}{
				"entity": fhir.Reference{
					Reference: fhir.FormatReference(m.EntityType, m.EntityID),
				},
				"inactive": m.Inactive,
			}
			if m.PeriodStart != nil || m.PeriodEnd != nil {
				entry["period"] = fhir.Period{
					Start: m.PeriodStart,
					End:   m.PeriodEnd,
				}
			}
			members = append(members, entry)
		}
		result["member"] = members
	}

	return result
}

// GroupFromFHIR parses a FHIR R4 Group resource map into the domain model.
func GroupFromFHIR(data map[string]interface{}) (*Group, error) {
	g := &Group{}

	if v, ok := data["type"].(string); ok {
		if !IsValidGroupType(v) {
			return nil, fmt.Errorf("invalid group type: %s", v)
		}
		g.Type = GroupType(v)
	}

	if v, ok := data["actual"].(bool); ok {
		g.Actual = v
	}

	if v, ok := data["active"].(bool); ok {
		g.Active = v
	}

	if v, ok := data["name"].(string); ok {
		g.Name = v
	}

	if v, ok := data["quantity"].(float64); ok {
		g.Quantity = int(v)
	}

	if v, ok := data["code"].(map[string]interface{}); ok {
		if codings, ok := v["coding"].([]interface{}); ok && len(codings) > 0 {
			if coding, ok := codings[0].(map[string]interface{}); ok {
				if code, ok := coding["code"].(string); ok {
					g.Code = &code
				}
			}
		}
	}

	if v, ok := data["managingEntity"].(map[string]interface{}); ok {
		if ref, ok := v["reference"].(string); ok {
			g.ManagingEntity = &ref
		}
	}

	if members, ok := data["member"].([]interface{}); ok {
		for _, mi := range members {
			m, ok := mi.(map[string]interface{})
			if !ok {
				continue
			}
			member := GroupMember{}
			if entity, ok := m["entity"].(map[string]interface{}); ok {
				if ref, ok := entity["reference"].(string); ok {
					parts := strings.SplitN(ref, "/", 2)
					if len(parts) == 2 {
						member.EntityType = parts[0]
						member.EntityID = parts[1]
					}
				}
			}
			if inactive, ok := m["inactive"].(bool); ok {
				member.Inactive = inactive
			}
			if period, ok := m["period"].(map[string]interface{}); ok {
				if start, ok := period["start"].(string); ok {
					if t, err := time.Parse(time.RFC3339, start); err == nil {
						member.PeriodStart = &t
					}
				}
				if end, ok := period["end"].(string); ok {
					if t, err := time.Parse(time.RFC3339, end); err == nil {
						member.PeriodEnd = &t
					}
				}
			}
			member.ID = uuid.New()
			g.Members = append(g.Members, member)
		}
	}

	return g, nil
}

// --------------------------------------------------------------------------
// Repository interface
// --------------------------------------------------------------------------

// GroupRepository defines the persistence interface for Group resources.
type GroupRepository interface {
	Create(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filterType string, limit, offset int) ([]*Group, int, error)
	AddMember(ctx context.Context, groupID uuid.UUID, member *GroupMember) error
	RemoveMember(ctx context.Context, groupID uuid.UUID, memberID uuid.UUID) error
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMember, error)
}

// --------------------------------------------------------------------------
// In-memory repository
// --------------------------------------------------------------------------

// InMemoryGroupRepo is an in-memory implementation of GroupRepository for testing.
type InMemoryGroupRepo struct {
	mu     sync.RWMutex
	groups map[uuid.UUID]*Group
}

// NewInMemoryGroupRepo creates a new in-memory group repository.
func NewInMemoryGroupRepo() *InMemoryGroupRepo {
	return &InMemoryGroupRepo{groups: make(map[uuid.UUID]*Group)}
}

func (r *InMemoryGroupRepo) Create(_ context.Context, group *Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	group.ID = uuid.New()
	if group.FHIRID == "" {
		group.FHIRID = group.ID.String()
	}
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now
	// Deep copy members to avoid external slice mutation.
	stored := *group
	stored.Members = make([]GroupMember, len(group.Members))
	copy(stored.Members, group.Members)
	r.groups[group.ID] = &stored
	return nil
}

func (r *InMemoryGroupRepo) GetByID(_ context.Context, id uuid.UUID) (*Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.groups[id]
	if !ok {
		return nil, fmt.Errorf("group not found")
	}
	// Return a copy to avoid mutation.
	out := *g
	out.Members = make([]GroupMember, len(g.Members))
	copy(out.Members, g.Members)
	return &out, nil
}

func (r *InMemoryGroupRepo) Update(_ context.Context, group *Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[group.ID]; !ok {
		return fmt.Errorf("group not found")
	}
	group.UpdatedAt = time.Now()
	stored := *group
	stored.Members = make([]GroupMember, len(group.Members))
	copy(stored.Members, group.Members)
	r.groups[group.ID] = &stored
	return nil
}

func (r *InMemoryGroupRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[id]; !ok {
		return fmt.Errorf("group not found")
	}
	delete(r.groups, id)
	return nil
}

func (r *InMemoryGroupRepo) List(_ context.Context, filterType string, limit, offset int) ([]*Group, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Group
	for _, g := range r.groups {
		if filterType != "" && string(g.Type) != filterType {
			continue
		}
		out := *g
		out.Members = make([]GroupMember, len(g.Members))
		copy(out.Members, g.Members)
		result = append(result, &out)
	}
	total := len(result)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (r *InMemoryGroupRepo) AddMember(_ context.Context, groupID uuid.UUID, member *GroupMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[groupID]
	if !ok {
		return fmt.Errorf("group not found")
	}
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}
	g.Members = append(g.Members, *member)
	g.Quantity = len(g.Members)
	g.UpdatedAt = time.Now()
	return nil
}

func (r *InMemoryGroupRepo) RemoveMember(_ context.Context, groupID uuid.UUID, memberID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[groupID]
	if !ok {
		return fmt.Errorf("group not found")
	}
	for i, m := range g.Members {
		if m.ID == memberID {
			g.Members = append(g.Members[:i], g.Members[i+1:]...)
			g.Quantity = len(g.Members)
			g.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

func (r *InMemoryGroupRepo) ListMembers(_ context.Context, groupID uuid.UUID) ([]GroupMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.groups[groupID]
	if !ok {
		return nil, fmt.Errorf("group not found")
	}
	out := make([]GroupMember, len(g.Members))
	copy(out, g.Members)
	return out, nil
}

// --------------------------------------------------------------------------
// Service
// --------------------------------------------------------------------------

// GroupService provides business logic for Group resources.
type GroupService struct {
	repo GroupRepository
}

// NewGroupService creates a new GroupService.
func NewGroupService(repo GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) CreateGroup(ctx context.Context, group *Group) error {
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}
	if !IsValidGroupType(string(group.Type)) {
		return fmt.Errorf("invalid group type: %s", group.Type)
	}
	group.Active = true
	group.Quantity = len(group.Members)
	return s.repo.Create(ctx, group)
}

func (s *GroupService) GetGroup(ctx context.Context, id uuid.UUID) (*Group, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GroupService) UpdateGroup(ctx context.Context, group *Group) error {
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}
	if !IsValidGroupType(string(group.Type)) {
		return fmt.Errorf("invalid group type: %s", group.Type)
	}
	group.Quantity = len(group.Members)
	return s.repo.Update(ctx, group)
}

func (s *GroupService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *GroupService) ListGroups(ctx context.Context, filterType string, limit, offset int) ([]*Group, int, error) {
	if filterType != "" && !IsValidGroupType(filterType) {
		return nil, 0, fmt.Errorf("invalid group type filter: %s", filterType)
	}
	return s.repo.List(ctx, filterType, limit, offset)
}

func (s *GroupService) AddMember(ctx context.Context, groupID uuid.UUID, member *GroupMember) error {
	if member.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	if member.EntityType == "" {
		return fmt.Errorf("entity_type is required")
	}
	return s.repo.AddMember(ctx, groupID, member)
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID uuid.UUID, memberID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, groupID, memberID)
}

func (s *GroupService) ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMember, error) {
	return s.repo.ListMembers(ctx, groupID)
}

// --------------------------------------------------------------------------
// Handler
// --------------------------------------------------------------------------

// GroupHandler provides HTTP handlers for Group resources.
type GroupHandler struct {
	svc *GroupService
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(svc *GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

// RegisterGroupRoutes registers Group resource routes on the given Echo groups.
func (h *GroupHandler) RegisterGroupRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// Operational API routes
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	readGroup.GET("/groups", h.ListGroups)
	readGroup.GET("/groups/:id", h.GetGroup)
	readGroup.GET("/groups/:id/members", h.ListMembers)

	writeGroup := api.Group("", auth.RequireRole("admin"))
	writeGroup.POST("/groups", h.CreateGroup)
	writeGroup.PUT("/groups/:id", h.UpdateGroup)
	writeGroup.DELETE("/groups/:id", h.DeleteGroup)
	writeGroup.POST("/groups/:id/members", h.AddMember)
	writeGroup.DELETE("/groups/:id/members/:member_id", h.RemoveMember)

	// FHIR endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar"))
	fhirRead.GET("/Group", h.SearchGroupsFHIR)
	fhirRead.GET("/Group/:id", h.GetGroupFHIR)

	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin"))
	fhirWrite.POST("/Group", h.CreateGroupFHIR)
	fhirWrite.PUT("/Group/:id", h.UpdateGroupFHIR)
	fhirWrite.DELETE("/Group/:id", h.DeleteGroupFHIR)
}

// -- Operational Handlers --

func (h *GroupHandler) CreateGroup(c echo.Context) error {
	var group Group
	if err := c.Bind(&group); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateGroup(c.Request().Context(), &group); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, group)
}

func (h *GroupHandler) GetGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	group, err := h.svc.GetGroup(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	return c.JSON(http.StatusOK, group)
}

func (h *GroupHandler) UpdateGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var group Group
	if err := c.Bind(&group); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	group.ID = id
	// Preserve FHIR ID from existing.
	existing, err := h.svc.GetGroup(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	group.FHIRID = existing.FHIRID
	group.CreatedAt = existing.CreatedAt
	if err := h.svc.UpdateGroup(c.Request().Context(), &group); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, group)
}

func (h *GroupHandler) DeleteGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteGroup(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *GroupHandler) ListGroups(c echo.Context) error {
	p := pagination.FromContext(c)
	filterType := c.QueryParam("type")
	groups, total, err := h.svc.ListGroups(c.Request().Context(), filterType, p.Limit, p.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(groups, total, p.Limit, p.Offset))
}

func (h *GroupHandler) AddMember(c echo.Context) error {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group id")
	}
	var member GroupMember
	if err := c.Bind(&member); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.AddMember(c.Request().Context(), groupID, &member); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, member)
}

func (h *GroupHandler) RemoveMember(c echo.Context) error {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group id")
	}
	memberID, err := uuid.Parse(c.Param("member_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid member id")
	}
	if err := h.svc.RemoveMember(c.Request().Context(), groupID, memberID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *GroupHandler) ListMembers(c echo.Context) error {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group id")
	}
	members, err := h.svc.ListMembers(c.Request().Context(), groupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, members)
}

// -- FHIR Handlers --

func (h *GroupHandler) SearchGroupsFHIR(c echo.Context) error {
	p := pagination.FromContext(c)
	filterType := c.QueryParam("type")
	groups, total, err := h.svc.ListGroups(c.Request().Context(), filterType, p.Limit, p.Offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(groups))
	for i, g := range groups {
		resources[i] = g.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Group"))
}

func (h *GroupHandler) GetGroupFHIR(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Group", c.Param("id")))
	}
	group, err := h.svc.GetGroup(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Group", c.Param("id")))
	}
	return c.JSON(http.StatusOK, group.ToFHIR())
}

func (h *GroupHandler) CreateGroupFHIR(c echo.Context) error {
	var data map[string]interface{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	group, err := GroupFromFHIR(data)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateGroup(c.Request().Context(), group); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Group/"+group.FHIRID)
	return c.JSON(http.StatusCreated, group.ToFHIR())
}

func (h *GroupHandler) UpdateGroupFHIR(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Group", c.Param("id")))
	}
	existing, err := h.svc.GetGroup(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Group", c.Param("id")))
	}
	var data map[string]interface{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	group, err := GroupFromFHIR(data)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	group.ID = existing.ID
	group.FHIRID = existing.FHIRID
	group.CreatedAt = existing.CreatedAt
	if err := h.svc.UpdateGroup(c.Request().Context(), group); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, group.ToFHIR())
}

func (h *GroupHandler) DeleteGroupFHIR(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Group", c.Param("id")))
	}
	if err := h.svc.DeleteGroup(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Group", c.Param("id")))
	}
	return c.NoContent(http.StatusNoContent)
}
