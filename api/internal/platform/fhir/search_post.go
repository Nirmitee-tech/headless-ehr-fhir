package fhir

import (
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

// SearchPostMiddleware creates middleware that converts POST _search requests
// into the equivalent GET request format by merging form body params into
// the query string. This allows the same search handler to work for both
// GET /Resource?params and POST /Resource/_search with form body.
func SearchPostMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "POST" {
				contentType := c.Request().Header.Get("Content-Type")
				if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
					// Parse the form body and merge its params into the query.
					if err := c.Request().ParseForm(); err == nil {
						// Update both the raw URL (for anything reading it fresh)
						// and Echo's cached query map. echo.Context.QueryParams()
						// memoizes url.Query() on first access, so an earlier
						// middleware (e.g. content negotiation) may have already
						// cached the pre-merge query; mutating the returned map in
						// place keeps the cache and the URL consistent.
						existingQuery := c.Request().URL.Query()
						cachedQuery := c.QueryParams()
						for k, vals := range c.Request().PostForm {
							for _, val := range vals {
								existingQuery.Add(k, val)
								cachedQuery.Add(k, val)
							}
						}
						c.Request().URL.RawQuery = existingQuery.Encode()
					}
				}
			}
			return next(c)
		}
	}
}

// MergeSearchParams merges POST form body params with URL query params.
// URL query params take precedence.
func MergeSearchParams(queryStr string, formBody url.Values) map[string][]string {
	result := make(map[string][]string)

	// Parse query string
	if queryStr != "" {
		parsed, err := url.ParseQuery(queryStr)
		if err == nil {
			for k, v := range parsed {
				result[k] = v
			}
		}
	}

	// Add form body params (don't override query params)
	for k, v := range formBody {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}
