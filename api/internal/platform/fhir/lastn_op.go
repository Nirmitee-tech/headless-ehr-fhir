package fhir

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// LastNParams holds parameters for the $lastn operation.
type LastNParams struct {
	Patient  string // Patient reference
	Category string // Observation category (optional)
	Code     string // Observation code (optional, can be system|code)
	Max      int    // Maximum number per group (default 1)
}

// ParseLastNParams extracts $lastn parameters from the request.
func ParseLastNParams(c echo.Context) LastNParams {
	params := LastNParams{
		Patient:  c.QueryParam("patient"),
		Category: c.QueryParam("category"),
		Code:     c.QueryParam("code"),
		Max:      1,
	}

	if maxStr := c.QueryParam("max"); maxStr != "" {
		if v, err := strconv.Atoi(maxStr); err == nil && v > 0 {
			params.Max = v
		}
	}

	return params
}

// LastNExecutor is a function that executes the actual $lastn query.
type LastNExecutor func(ctx context.Context, params LastNParams) ([]map[string]interface{}, error)

// LastNHandler creates a handler for GET/POST /fhir/Observation/$lastn.
// It accepts a function that executes the actual query.
func LastNHandler(executor LastNExecutor) echo.HandlerFunc {
	return func(c echo.Context) error {
		params := ParseLastNParams(c)

		if params.Patient == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "patient parameter is required for $lastn",
			))
		}

		results, err := executor(c.Request().Context(), params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, "lastn operation failed: "+err.Error(),
			))
		}

		bundle := buildLastNBundle(results)

		return c.JSON(http.StatusOK, bundle)
	}
}

// buildLastNBundle creates a FHIR searchset Bundle from $lastn results.
func buildLastNBundle(results []map[string]interface{}) map[string]interface{} {
	now := time.Now().UTC().Format(time.RFC3339)
	total := len(results)

	entries := make([]interface{}, 0, total)
	for _, r := range results {
		raw, _ := json.Marshal(r)

		fullURL := ""
		rt, _ := r["resourceType"].(string)
		id, _ := r["id"].(string)
		if rt != "" && id != "" {
			fullURL = rt + "/" + id
		}

		entry := map[string]interface{}{
			"resource": json.RawMessage(raw),
			"search": map[string]interface{}{
				"mode": "match",
			},
		}
		if fullURL != "" {
			entry["fullUrl"] = fullURL
		}

		entries = append(entries, entry)
	}

	return map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        total,
		"timestamp":    now,
		"entry":        entries,
	}
}
