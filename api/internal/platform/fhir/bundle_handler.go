package fhir

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ehr/ehr/internal/platform/db"
	"github.com/labstack/echo/v4"
)

// BundleProcessor defines the interface for processing individual bundle entries.
// Implementations should handle the actual FHIR resource operations (create, update, delete, read)
// by dispatching to the appropriate domain handlers.
type BundleProcessor interface {
	// ProcessEntry processes a single bundle entry and returns the response entry.
	// The method, resourceType, and resourceID are parsed from the entry's request.
	ProcessEntry(c echo.Context, method, resourceType, resourceID string, resource json.RawMessage) (BundleEntry, error)
}

// BundleHandler handles FHIR Bundle operations (transaction and batch).
type BundleHandler struct {
	processor BundleProcessor
	validator *Validator
}

// NewBundleHandler creates a new BundleHandler.
func NewBundleHandler(processor BundleProcessor) *BundleHandler {
	return &BundleHandler{
		processor: processor,
		validator: NewValidator(),
	}
}

// RegisterRoutes registers the bundle processing endpoint.
func (h *BundleHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.POST("", h.ProcessBundle)
}

// ProcessBundle handles POST /fhir with a Bundle of type "transaction" or "batch".
func (h *BundleHandler) ProcessBundle(c echo.Context) error {
	var bundle Bundle
	if err := json.NewDecoder(c.Request().Body).Decode(&bundle); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("invalid Bundle JSON: "+err.Error()))
	}

	if bundle.ResourceType != "Bundle" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("request body must be a Bundle resource"))
	}

	// Validate the bundle
	vResult := h.validator.ValidateBundle(&bundle)
	if !vResult.Valid {
		return c.JSON(http.StatusBadRequest, vResult.ToOperationOutcome())
	}

	switch bundle.Type {
	case "transaction":
		return h.processTransaction(c, &bundle)
	case "batch":
		return h.processBatch(c, &bundle)
	default:
		return c.JSON(http.StatusBadRequest, ErrorOutcome(
			fmt.Sprintf("unsupported bundle type '%s'; expected 'transaction' or 'batch'", bundle.Type)))
	}
}

// processTransaction processes a transaction bundle atomically.
// If any entry fails, all changes are rolled back.
func (h *BundleHandler) processTransaction(c echo.Context, bundle *Bundle) error {
	ctx := c.Request().Context()

	txCtx, tx, err := db.WithTx(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, NewOutcomeBuilder().
			AddIssue(IssueSeverityError, IssueTypeException,
				"failed to begin transaction: "+err.Error()).
			Build())
	}

	// Use the transaction context for all entry processing
	c.SetRequest(c.Request().WithContext(txCtx))

	responseEntries := make([]BundleEntry, len(bundle.Entry))

	for i, entry := range bundle.Entry {
		method, resourceType, resourceID := parseEntryRequest(entry)

		respEntry, err := h.processor.ProcessEntry(c, method, resourceType, resourceID, entry.Resource)
		if err != nil {
			// Transaction: any failure rolls back everything.
			_ = tx.Rollback(ctx)
			// Restore original context
			c.SetRequest(c.Request().WithContext(ctx))
			return c.JSON(http.StatusBadRequest, NewOutcomeBuilder().
				AddIssue(IssueSeverityError, IssueTypeProcessing,
					fmt.Sprintf("transaction failed at entry[%d]: %s", i, err.Error())).
				Build())
		}
		responseEntries[i] = respEntry
	}

	if err := tx.Commit(ctx); err != nil {
		// Restore original context
		c.SetRequest(c.Request().WithContext(ctx))
		return c.JSON(http.StatusInternalServerError, NewOutcomeBuilder().
			AddIssue(IssueSeverityError, IssueTypeException,
				"failed to commit transaction: "+err.Error()).
			Build())
	}

	// Restore original context
	c.SetRequest(c.Request().WithContext(ctx))
	return c.JSON(http.StatusOK, NewTransactionResponse(responseEntries))
}

// processBatch processes a batch bundle non-atomically.
// Each entry is processed independently; failures do not affect other entries.
func (h *BundleHandler) processBatch(c echo.Context, bundle *Bundle) error {
	responseEntries := make([]BundleEntry, len(bundle.Entry))

	for i, entry := range bundle.Entry {
		method, resourceType, resourceID := parseEntryRequest(entry)

		respEntry, err := h.processor.ProcessEntry(c, method, resourceType, resourceID, entry.Resource)
		if err != nil {
			// Batch: record the error for this entry but continue processing others.
			now := time.Now().UTC()
			outcome := ErrorOutcome(err.Error())
			outcomeData, _ := json.Marshal(outcome)
			responseEntries[i] = BundleEntry{
				Response: &BundleResponse{
					Status:       "400 Bad Request",
					LastModified: &now,
					Outcome:      outcome,
				},
				Resource: outcomeData,
			}
		} else {
			responseEntries[i] = respEntry
		}
	}

	return c.JSON(http.StatusOK, NewBatchResponse(responseEntries))
}

// parseEntryRequest extracts the HTTP method, resource type, and resource ID from a bundle entry.
func parseEntryRequest(entry BundleEntry) (method, resourceType, resourceID string) {
	if entry.Request == nil {
		return "", "", ""
	}

	method = strings.ToUpper(entry.Request.Method)

	// Parse the URL: expected format is "ResourceType" or "ResourceType/id"
	url := strings.TrimPrefix(entry.Request.URL, "/")
	// Remove any query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	parts := strings.SplitN(url, "/", 2)
	if len(parts) >= 1 {
		resourceType = parts[0]
	}
	if len(parts) >= 2 {
		resourceID = parts[1]
	}

	return method, resourceType, resourceID
}

// DefaultBundleProcessor is a no-op implementation of BundleProcessor.
// Real applications should provide their own implementation that dispatches
// to domain-specific handlers.
type DefaultBundleProcessor struct{}

// ProcessEntry is a stub implementation that returns a basic success response.
// Override this in production by providing a real BundleProcessor to NewBundleHandler.
func (p *DefaultBundleProcessor) ProcessEntry(c echo.Context, method, resourceType, resourceID string, resource json.RawMessage) (BundleEntry, error) {
	now := time.Now().UTC()

	var status string
	var location string

	switch method {
	case "POST":
		status = "201 Created"
		location = fmt.Sprintf("%s/%s", resourceType, resourceID)
	case "PUT":
		status = "200 OK"
		location = fmt.Sprintf("%s/%s", resourceType, resourceID)
	case "DELETE":
		status = "204 No Content"
	case "GET":
		status = "200 OK"
	default:
		return BundleEntry{}, fmt.Errorf("unsupported method: %s", method)
	}

	return BundleEntry{
		Response: &BundleResponse{
			Status:       status,
			Location:     location,
			LastModified: &now,
		},
		Resource: resource,
	}, nil
}
