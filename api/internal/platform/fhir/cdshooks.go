package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// CDS Hooks 2.0 types (HL7 spec)
// ---------------------------------------------------------------------------

// CDSService describes a single CDS service returned in discovery.
type CDSService struct {
	Hook              string            `json:"hook"`
	Title             string            `json:"title,omitempty"`
	Description       string            `json:"description"`
	ID                string            `json:"id"`
	Prefetch          map[string]string `json:"prefetch,omitempty"`
	UsageRequirements string            `json:"usageRequirements,omitempty"`
}

// CDSHookRequest is the payload POSTed to invoke a hook.
type CDSHookRequest struct {
	Hook         string                 `json:"hook"`
	HookInstance string                 `json:"hookInstance"`
	FHIRServer   string                 `json:"fhirServer,omitempty"`
	FHIRAuth     *CDSFHIRAuth           `json:"fhirAuthorization,omitempty"`
	Context      map[string]interface{} `json:"context"`
	Prefetch     map[string]interface{} `json:"prefetch,omitempty"`
}

// CDSFHIRAuth carries FHIR authorization details from the EHR.
type CDSFHIRAuth struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	Subject     string `json:"subject"`
}

// CDSCard is a single card in the hook response.
type CDSCard struct {
	UUID              string          `json:"uuid,omitempty"`
	Summary           string          `json:"summary"`
	Detail            string          `json:"detail,omitempty"`
	Indicator         string          `json:"indicator"`
	Source            CDSSource       `json:"source"`
	Suggestions       []CDSSuggestion `json:"suggestions,omitempty"`
	Links             []CDSLink       `json:"links,omitempty"`
	OverrideReasons   []CDSCoding     `json:"overrideReasons,omitempty"`
	SelectionBehavior string          `json:"selectionBehavior,omitempty"`
}

// CDSSource identifies the source of a card.
type CDSSource struct {
	Label string     `json:"label"`
	URL   string     `json:"url,omitempty"`
	Icon  string     `json:"icon,omitempty"`
	Topic *CDSCoding `json:"topic,omitempty"`
}

// CDSSuggestion is a suggested action within a card.
type CDSSuggestion struct {
	Label         string      `json:"label"`
	UUID          string      `json:"uuid,omitempty"`
	IsRecommended bool        `json:"isRecommended,omitempty"`
	Actions       []CDSAction `json:"actions,omitempty"`
}

// CDSAction is an individual action within a suggestion.
type CDSAction struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Resource    interface{} `json:"resource,omitempty"`
}

// CDSLink is an external link within a card.
type CDSLink struct {
	Label      string `json:"label"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	AppContext string `json:"appContext,omitempty"`
}

// CDSCoding is a code/system/display triple used in CDS Hooks.
type CDSCoding struct {
	Code    string `json:"code"`
	System  string `json:"system,omitempty"`
	Display string `json:"display,omitempty"`
}

// CDSHookResponse is returned from hook invocation.
type CDSHookResponse struct {
	Cards         []CDSCard   `json:"cards"`
	SystemActions []CDSAction `json:"systemActions,omitempty"`
}

// CDSFeedbackRequest records what the user did with a card.
type CDSFeedbackRequest struct {
	Card             string      `json:"card"`
	Outcome          string      `json:"outcome"`
	OverrideReasons  []CDSCoding `json:"overrideReasons,omitempty"`
	OutcomeTimestamp string      `json:"outcomeTimestamp,omitempty"`
}

// ---------------------------------------------------------------------------
// Handler function types
// ---------------------------------------------------------------------------

// ServiceHandler processes a CDS hook request and returns cards.
type ServiceHandler func(ctx context.Context, req CDSHookRequest) (*CDSHookResponse, error)

// FeedbackHandler processes feedback for a service.
type FeedbackHandler func(ctx context.Context, serviceID string, fb CDSFeedbackRequest) error

// ---------------------------------------------------------------------------
// CDSHooksHandler
// ---------------------------------------------------------------------------

// CDSHooksHandler implements the HL7 CDS Hooks 2.0 REST API.
type CDSHooksHandler struct {
	services         map[string]CDSService
	handlers         map[string]ServiceHandler
	feedbackHandlers map[string]FeedbackHandler
	order            []string
}

// NewCDSHooksHandler creates a new CDSHooksHandler.
func NewCDSHooksHandler() *CDSHooksHandler {
	return &CDSHooksHandler{
		services:         make(map[string]CDSService),
		handlers:         make(map[string]ServiceHandler),
		feedbackHandlers: make(map[string]FeedbackHandler),
	}
}

// RegisterService registers a CDS service and its handler.
func (h *CDSHooksHandler) RegisterService(svc CDSService, handler ServiceHandler) {
	if _, exists := h.services[svc.ID]; !exists {
		h.order = append(h.order, svc.ID)
	}
	h.services[svc.ID] = svc
	h.handlers[svc.ID] = handler
}

// RegisterFeedbackHandler registers an optional feedback handler for a service.
func (h *CDSHooksHandler) RegisterFeedbackHandler(serviceID string, handler FeedbackHandler) {
	h.feedbackHandlers[serviceID] = handler
}

// RegisterRoutes registers CDS Hooks routes on the root Echo instance.
func (h *CDSHooksHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/cds-services", h.Discovery)
	e.POST("/cds-services/:id", h.HandleHook)
	e.POST("/cds-services/:id/feedback", h.HandleFeedback)
}

// Discovery handles GET /cds-services — returns all registered services.
func (h *CDSHooksHandler) Discovery(c echo.Context) error {
	services := make([]CDSService, 0, len(h.order))
	for _, id := range h.order {
		if svc, ok := h.services[id]; ok {
			services = append(services, svc)
		}
	}
	return c.JSON(http.StatusOK, map[string][]CDSService{
		"services": services,
	})
}

// HandleHook handles POST /cds-services/:id — invokes a CDS hook.
func (h *CDSHooksHandler) HandleHook(c echo.Context) error {
	serviceID := c.Param("id")

	svc, ok := h.services[serviceID]
	if !ok {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("CDS Service", serviceID))
	}

	var req CDSHookRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid request body: %v", err)))
	}

	if req.Hook != svc.Hook {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(
			fmt.Sprintf("hook mismatch: request hook %q does not match service hook %q", req.Hook, svc.Hook),
		))
	}

	if req.HookInstance == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("hookInstance is required"))
	}

	handler, ok := h.handlers[serviceID]
	if !ok {
		return c.JSON(http.StatusInternalServerError, InternalErrorOutcome("no handler registered for service"))
	}

	resp, err := handler(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, InternalErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, resp)
}

// HandleFeedback handles POST /cds-services/:id/feedback — records card feedback.
func (h *CDSHooksHandler) HandleFeedback(c echo.Context) error {
	serviceID := c.Param("id")

	if _, ok := h.services[serviceID]; !ok {
		return c.JSON(http.StatusNotFound, NotFoundOutcome("CDS Service", serviceID))
	}

	var fb CDSFeedbackRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&fb); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(fmt.Sprintf("invalid feedback body: %v", err)))
	}

	handler, ok := h.feedbackHandlers[serviceID]
	if !ok {
		// No feedback handler registered — return 200 as no-op
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	if err := handler(c.Request().Context(), serviceID, fb); err != nil {
		return c.JSON(http.StatusInternalServerError, InternalErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
