package fhir

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ConditionalResult represents the outcome of a conditional search.
type ConditionalResult struct {
	Count      int
	ResourceID string
	FHIRID     string
}

// ResourceSearcher is called by conditional operations to find matching resources.
type ResourceSearcher func(c echo.Context, params map[string]string) (*ConditionalResult, error)

// ConditionalCreateMiddleware implements FHIR conditional create (If-None-Exist header).
// If the If-None-Exist header is present, it searches for existing resources matching the criteria.
// - 0 matches: proceed with create (call next)
// - 1 match: return 200 OK with existing resource (no create)
// - 2+ matches: return 412 Precondition Failed
func ConditionalCreateMiddleware(searcher ResourceSearcher) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ifNoneExist := c.Request().Header.Get("If-None-Exist")
			if ifNoneExist == "" {
				return next(c)
			}

			params := parseSearchString(ifNoneExist)
			result, err := searcher(c, params)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorOutcome("conditional search failed: "+err.Error()))
			}

			switch {
			case result.Count == 0:
				return next(c)
			case result.Count == 1:
				return c.JSON(http.StatusOK, map[string]interface{}{
					"resourceType": "OperationOutcome",
					"issue": []map[string]interface{}{{
						"severity":    "information",
						"code":        "duplicate",
						"diagnostics": "resource already exists matching If-None-Exist criteria",
					}},
				})
			default:
				return c.JSON(http.StatusPreconditionFailed, ErrorOutcome(
					"multiple resources match the If-None-Exist criteria"))
			}
		}
	}
}

// ConditionalUpdateHandler implements FHIR conditional update.
// PUT /fhir/ResourceType?search-params
// - 0 matches: create new resource
// - 1 match: update existing resource
// - 2+ matches: return 412 Precondition Failed
func ConditionalUpdateHandler(searcher ResourceSearcher, createHandler, updateHandler echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Only applies when there's no explicit ID in the path
		if c.Param("id") != "" {
			return updateHandler(c)
		}

		params := map[string]string{}
		for k, v := range c.QueryParams() {
			if len(v) > 0 && !strings.HasPrefix(k, "_") {
				params[k] = v[0]
			}
		}

		if len(params) == 0 {
			return createHandler(c)
		}

		result, err := searcher(c, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome("conditional search failed: "+err.Error()))
		}

		switch {
		case result.Count == 0:
			return createHandler(c)
		case result.Count == 1:
			c.SetParamNames("id")
			c.SetParamValues(result.FHIRID)
			return updateHandler(c)
		default:
			return c.JSON(http.StatusPreconditionFailed, ErrorOutcome(
				"multiple resources match the conditional update criteria"))
		}
	}
}

// ConditionalDeleteHandler implements FHIR conditional delete.
// DELETE /fhir/ResourceType?search-params
// - 0 matches: return 204 No Content (nothing to delete)
// - 1 match: delete the resource
// - 2+ matches: return 412 Precondition Failed (single mode) or delete all (multiple mode)
func ConditionalDeleteHandler(searcher ResourceSearcher, deleteHandler echo.HandlerFunc, allowMultiple bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Param("id") != "" {
			return deleteHandler(c)
		}

		params := map[string]string{}
		for k, v := range c.QueryParams() {
			if len(v) > 0 && !strings.HasPrefix(k, "_") {
				params[k] = v[0]
			}
		}

		if len(params) == 0 {
			return c.JSON(http.StatusBadRequest, ErrorOutcome("conditional delete requires search parameters"))
		}

		result, err := searcher(c, params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorOutcome("conditional search failed: "+err.Error()))
		}

		switch {
		case result.Count == 0:
			return c.NoContent(http.StatusNoContent)
		case result.Count == 1:
			c.SetParamNames("id")
			c.SetParamValues(result.FHIRID)
			return deleteHandler(c)
		default:
			if !allowMultiple {
				return c.JSON(http.StatusPreconditionFailed, ErrorOutcome(
					"multiple resources match the conditional delete criteria; use single delete"))
			}
			c.SetParamNames("id")
			c.SetParamValues(result.FHIRID)
			return deleteHandler(c)
		}
	}
}

// parseSearchString parses a search query string like "identifier=foo&name=bar" into a map.
func parseSearchString(query string) map[string]string {
	params := map[string]string{}
	query = strings.TrimPrefix(query, "?")
	for _, part := range strings.Split(query, "&") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return params
}
