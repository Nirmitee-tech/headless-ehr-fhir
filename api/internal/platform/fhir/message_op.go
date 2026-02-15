package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MessageEventHandler processes a specific message event.
// It receives the MessageHeader and the focus resources, and returns response
// resources or an error.
type MessageEventHandler func(ctx context.Context, header map[string]interface{}, focus []map[string]interface{}) ([]map[string]interface{}, error)

// MessageProcessor handles FHIR message bundles by dispatching to registered
// event handlers. Handlers are registered once at startup and read concurrently
// at runtime, so the handlers map is effectively read-only after initialisation.
type MessageProcessor struct {
	handlers map[string]MessageEventHandler
}

// NewMessageProcessor creates a new processor with built-in event handlers.
func NewMessageProcessor() *MessageProcessor {
	p := &MessageProcessor{
		handlers: make(map[string]MessageEventHandler),
	}
	p.RegisterHandler("notification", notificationHandler)
	p.RegisterHandler("patient-link", patientLinkHandler)
	p.RegisterHandler("diagnostic-report", diagnosticReportHandler)
	return p
}

// RegisterHandler registers a handler for a message event URI.
func (p *MessageProcessor) RegisterHandler(eventURI string, handler MessageEventHandler) {
	p.handlers[eventURI] = handler
}

// ProcessMessage validates and processes a FHIR Message Bundle.
//
// Algorithm:
//  1. Validate bundle type is "message".
//  2. Extract the entries array.
//  3. First entry MUST be a MessageHeader resource.
//  4. Extract eventCoding.code or eventUri from the MessageHeader.
//  5. Resolve focus resources from the bundle entries.
//  6. Look up a registered handler by event.
//  7. Call handler with header + focus resources.
//  8. Build and return a response Message Bundle.
func (p *MessageProcessor) ProcessMessage(ctx context.Context, bundle map[string]interface{}) (map[string]interface{}, error) {
	// 1. Validate bundle type.
	bundleType, _ := bundle["type"].(string)
	if bundleType != "message" {
		return nil, fmt.Errorf("bundle type must be 'message', got '%s'", bundleType)
	}

	// 2. Extract entries.
	entriesRaw, ok := bundle["entry"]
	if !ok {
		return nil, fmt.Errorf("message bundle has no entries")
	}
	entries, ok := entriesRaw.([]interface{})
	if !ok || len(entries) == 0 {
		return nil, fmt.Errorf("message bundle has no entries")
	}

	// 3. First entry must be a MessageHeader.
	firstEntry, ok := entries[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("first entry is not an object")
	}
	header, ok := firstEntry["resource"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("first entry has no resource")
	}
	rt, _ := header["resourceType"].(string)
	if rt != "MessageHeader" {
		return nil, fmt.Errorf("first entry must be a MessageHeader, got '%s'", rt)
	}

	// 4. Extract event code.
	eventCode := extractEventCode(header)
	if eventCode == "" {
		return nil, fmt.Errorf("MessageHeader has no event code")
	}

	// 5. Resolve focus resources.
	focus := resolveFocusResources(header, entries)

	// 6. Look up handler.
	handler, found := p.handlers[eventCode]

	// 7 & 8. Call handler and build response.
	headerID, _ := header["id"].(string)

	if !found {
		return buildMessageResponse(header, headerID, "fatal-error", eventCode, nil,
			fmt.Sprintf("no handler registered for event '%s'", eventCode)), nil
	}

	responseResources, err := handler(ctx, header, focus)
	if err != nil {
		return buildMessageResponse(header, headerID, "fatal-error", eventCode, nil, err.Error()), nil
	}

	return buildMessageResponse(header, headerID, "ok", eventCode, responseResources, ""), nil
}

// extractEventCode obtains the event code from a MessageHeader. It checks
// eventCoding.code first, then falls back to eventUri.
func extractEventCode(header map[string]interface{}) string {
	if ec, ok := header["eventCoding"].(map[string]interface{}); ok {
		if code, ok := ec["code"].(string); ok && code != "" {
			return code
		}
	}
	if uri, ok := header["eventUri"].(string); ok && uri != "" {
		return uri
	}
	return ""
}

// resolveFocusResources matches MessageHeader.focus[].reference values against
// bundle entry fullUrl values and returns the corresponding resources.
func resolveFocusResources(header map[string]interface{}, entries []interface{}) []map[string]interface{} {
	focusRefs := make([]string, 0)
	if focusArr, ok := header["focus"].([]interface{}); ok {
		for _, f := range focusArr {
			if fMap, ok := f.(map[string]interface{}); ok {
				if ref, ok := fMap["reference"].(string); ok && ref != "" {
					focusRefs = append(focusRefs, ref)
				}
			}
		}
	}

	// Build a lookup from fullUrl to resource.
	urlToResource := make(map[string]map[string]interface{}, len(entries))
	for _, e := range entries {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		fullURL, _ := entry["fullUrl"].(string)
		resource, _ := entry["resource"].(map[string]interface{})
		if fullURL != "" && resource != nil {
			urlToResource[fullURL] = resource
		}
	}

	var result []map[string]interface{}
	for _, ref := range focusRefs {
		if res, ok := urlToResource[ref]; ok {
			result = append(result, res)
		}
	}
	return result
}

// buildMessageResponse constructs a FHIR response Message Bundle.
func buildMessageResponse(originalHeader map[string]interface{}, originalID, responseCode, eventCode string, additionalResources []map[string]interface{}, errDiagnostics string) map[string]interface{} {
	responseID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	// Build the response MessageHeader.
	responseHeader := map[string]interface{}{
		"resourceType": "MessageHeader",
		"id":           responseID,
		"response": map[string]interface{}{
			"identifier": originalID,
			"code":       responseCode,
		},
		"source": map[string]interface{}{
			"endpoint": "http://ehr.example.org",
		},
	}

	// Copy the event from the original header.
	if ec, ok := originalHeader["eventCoding"]; ok {
		responseHeader["eventCoding"] = ec
	} else if eu, ok := originalHeader["eventUri"]; ok {
		responseHeader["eventUri"] = eu
	}

	entries := []interface{}{
		map[string]interface{}{
			"fullUrl":  "MessageHeader/" + responseID,
			"resource": responseHeader,
		},
	}

	// If error, add an OperationOutcome entry.
	if responseCode == "fatal-error" && errDiagnostics != "" {
		outcome := map[string]interface{}{
			"resourceType": "OperationOutcome",
			"issue": []interface{}{
				map[string]interface{}{
					"severity":    "error",
					"code":        "processing",
					"diagnostics": errDiagnostics,
				},
			},
		}
		outcomeID := uuid.New().String()
		entries = append(entries, map[string]interface{}{
			"fullUrl":  "OperationOutcome/" + outcomeID,
			"resource": outcome,
		})
	}

	// Add any resources returned by the handler.
	for _, res := range additionalResources {
		resID, _ := res["id"].(string)
		resType, _ := res["resourceType"].(string)
		fullURL := resType + "/" + resID
		if resType == "" && resID == "" {
			fullURL = "Resource/" + uuid.New().String()
		}
		entries = append(entries, map[string]interface{}{
			"fullUrl":  fullURL,
			"resource": res,
		})
	}

	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "message",
		"timestamp":    now,
		"entry":        entries,
	}
}

// --- Built-in event handlers ---

// notificationHandler accepts patient notifications (admit, discharge, etc.)
// and simply acknowledges the message.
func notificationHandler(_ context.Context, _ map[string]interface{}, _ []map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

// patientLinkHandler handles patient link/merge notifications.
// It validates that the focus contains at least one Patient resource.
func patientLinkHandler(_ context.Context, _ map[string]interface{}, focus []map[string]interface{}) ([]map[string]interface{}, error) {
	for _, res := range focus {
		if rt, _ := res["resourceType"].(string); rt == "Patient" {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("patient-link message requires at least one Patient focus resource")
}

// diagnosticReportHandler handles lab result messages.
// It validates that the focus contains at least one DiagnosticReport resource.
func diagnosticReportHandler(_ context.Context, _ map[string]interface{}, focus []map[string]interface{}) ([]map[string]interface{}, error) {
	for _, res := range focus {
		if rt, _ := res["resourceType"].(string); rt == "DiagnosticReport" {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("diagnostic-report message requires at least one DiagnosticReport focus resource")
}

// --- HTTP Handler ---

// MessageHandler provides the HTTP endpoint for the $process-message operation.
type MessageHandler struct {
	processor *MessageProcessor
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(processor *MessageProcessor) *MessageHandler {
	return &MessageHandler{processor: processor}
}

// RegisterRoutes registers the $process-message route on the given FHIR group.
func (h *MessageHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/$process-message", h.ProcessMessage)
}

// ProcessMessage handles POST /fhir/$process-message.
// It reads a Message Bundle from the request body, processes it, and returns
// the response Bundle with 200, or an OperationOutcome with 400 on error.
func (h *MessageHandler) ProcessMessage(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("failed to read request body"))
	}

	if len(body) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("request body is empty"))
	}

	var bundle map[string]interface{}
	if err := json.Unmarshal(body, &bundle); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid JSON: "+err.Error()))
	}

	ctx := c.Request().Context()
	response, err := h.processor.ProcessMessage(ctx, bundle)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, response)
}
